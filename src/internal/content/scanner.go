package content

import (
	"context"
	"sync"
	"time"

	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// scanner implements the updater interface to enable a content update via a
// scanning run, that regularly scans the music directory for changes that must
// be applied to the muserv content
type scanner struct {
	updNotif     chan UpdateNotification
	upd          chan struct{}
	errs         chan error
	filesByPaths func([]string) *fileInfos
	update       func(context.Context, *fileInfos, *fileInfos) (uint32, error)
}

// newScanner creates a new scanner instance
func newScanner(filesByPaths func([]string) *fileInfos, update func(context.Context, *fileInfos, *fileInfos) (uint32, error)) *scanner {
	sc := new(scanner)

	sc.errs = make(chan error)
	sc.updNotif = make(chan UpdateNotification)
	sc.upd = make(chan struct{})
	sc.filesByPaths = filesByPaths
	sc.update = update

	return sc
}

// run implements the regular scanning loop
func (me *scanner) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Trace("running scanner ...")

	// extract config from context
	cfg := ctx.Value(config.KeyCfg).(config.Cfg)

	var wg0 sync.WaitGroup
	ticker := time.NewTicker(cfg.Cnt.UpdateInterval * time.Second)

	// semaphore to ensure that only one content update run is done at any time
	sema := make(chan struct{}, 1)

	defer func() {
		ticker.Stop()
		close(me.errs)
		close(me.updNotif)
		close(me.upd)
		close(sema)
		log.Trace("scanner stopped")
	}()

	for {
		select {
		// periodic update trigger
		case <-ticker.C:
			wg.Add(1)
			go func(wg0 *sync.WaitGroup) {
				sema <- struct{}{}
				defer func() {
					<-sema
					wg.Done()
				}()

				fiDel, fiAdd := fullScan(cfg.Cnt.MusicDirs, me.filesByPaths)

				// channel to notify server about finalized update
				updated := make(chan uint32)
				// close channel after update is done (this implicitly notifies
				// the server that the update is done)
				defer close(updated)

				// notify server that an update is required and wait for
				// approval before update is executed
				me.updNotif <- UpdateNotification{
					Update:  func() { me.upd <- struct{}{} },
					Updated: updated,
				}
				<-me.upd

				// apply changes to content and report back the number of changed, deleted
				// or added objects
				var count uint32
				var err error
				if count, err = me.update(ctx, fiDel, fiAdd); err != nil {
					me.errs <- err
					return
				}
				updated <- count
			}(&wg0)

		// cancelation from server
		case <-ctx.Done():
			// wait until all changes are processed
			wg0.Wait()
			return
		}
	}
}

// errors returns a receive-only channel for errors from scanner
func (me *scanner) errors() <-chan error {
	return me.errs
}

// updateNotification returns a receive-only channel to notify about updates
func (me *scanner) updateNotification() <-chan UpdateNotification {
	return me.updNotif
}
