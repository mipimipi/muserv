package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	"gitlab.com/mipimipi/muserv/src/internal/config"
	"gitlab.com/mipimipi/muserv/src/internal/content"
	"gitlab.com/mipimipi/muserv/src/internal/upnp"
)

var log *l.Entry = l.WithFields(l.Fields{"srv": "server"})

// Run implements the main control loop of the server and starts the database
// and the UPnP service. version is the muserv version which is used to build
// the server string
func Run(version string) (err error) {
	// read and validate muserv configuration
	var cfg config.Cfg
	if cfg, err = config.Load(); err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		return
	}
	if err = cfg.Validate(); err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		return
	}

	// set up logging: no log entries possible before this statement!
	if err = setupLogging(cfg.LogDir, cfg.LogLevel); err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		return
	}

	log.Trace("running ...")

	// create root context
	ctx := context.WithValue(context.Background(), config.KeyCfg, cfg)
	ctx = context.WithValue(ctx, config.KeyVersion, version)

	// initialize server attributes (create content objects and UPnP server
	// objects). This must be done before the main control loop is started
	cnt, err := content.New(&cfg)
	if err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		return
	}
	upnp, err := upnp.New(ctx, cnt)
	if err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		return
	}

	// create context with cancel
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup

	// start UPnP server
	wg.Add(1)
	go upnp.Run(ctx, &wg)

	// update content initially
	if err = cnt.InitialUpdate(ctx); err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		cancel()
		return
	}

	// preparation to receive OS signals (e.g. from 'systemctl stop ...'). This
	// must be done before the main control loop is started
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// start regular content update
	wg.Add(1)
	go cnt.Run(ctx, &wg)

	// connect UPnP server
	if err = upnp.Connect(ctx); err != nil {
		err = errors.Wrap(err, "cannot run muserv")
		cancel()
		return
	}

	// main control loop
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			select {
			case sig := <-interrupt:
				// termination signal from OS received: stop processing
				log.Tracef("signal received: %v", sig)
				log.Trace("stopping ...")
				cancel()
				log.Trace("stopped")
				return

			case update := <-cnt.UpdateNotification():
				log.Trace("received update notification: executing update ...")
				// execute update
				update.Update()
				// receive number of updated objects, update ContainerUpdateIDs,
				// increase SystemUpdateID and - if the value range of
				// SystemUpdaetID  exceeded - trigger the service reset
				// procedure according to UPnP device architecture 2.0
				count := <-update.Updated
				upnp.SetContainerUpdateIDs(cnt.ContainerUpdateIDs())
				if upnp.IncrSystemUpdateID(count) {
					upnp.ServiceResetProcedure(ctx)
				}

			case err := <-upnp.Errors():
				// error received from UPNP: stop processing
				log.Tracef("UPNP error received: %v", err)
				log.Trace("stopping ...")
				cancel()
				log.Trace("stopped")
				return

			case err := <-cnt.Errors():
				// error received from updater: stop processing
				log.Tracef("updater error received: %v", err)
				log.Trace("stopping ...")
				cancel()
				log.Trace("stopped")
				return
			}
		}
	}(&wg)

	wg.Wait()

	return
}
