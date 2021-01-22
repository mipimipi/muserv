package content

import (
	"context"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
	f "gitlab.com/mipimipi/go-utils/file"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// notifier implements the updater interface to enable content updates based on
// file system changes detected by inotify
type notifier struct {
	changes      []notify.EventInfo
	mutChanges   sync.Mutex
	errs         chan error
	updNotif     chan UpdateNotification
	upd          chan struct{}
	filesByPaths func([]string) *fileInfos
	update       func(context.Context, *fileInfos, *fileInfos) (uint32, error)
}

// newNotifier creates a new instance of notifier
func newNotifier(filesByPaths func([]string) *fileInfos, update func(context.Context, *fileInfos, *fileInfos) (uint32, error)) *notifier {
	nf := new(notifier)

	nf.errs = make(chan error)
	nf.updNotif = make(chan UpdateNotification)
	nf.upd = make(chan struct{})
	nf.filesByPaths = filesByPaths
	nf.update = update

	return nf
}

// run implements the main control loop that listens to events from inotify
// and that regularly triggers a corresponding content update
func (me *notifier) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Trace("running notifier ...")

	// extract config from context
	cfg := ctx.Value(config.KeyCfg).(config.Cfg)

	// add watcher for inotify events for music dir. Changes can be received via
	// channel chgs
	chgs := make(chan notify.EventInfo, 1)
	for _, dir := range cfg.Cnt.MusicDirs {
		if err := notify.Watch(filepath.Join(dir, "..."), chgs, notify.All); err != nil {
			err = errors.Wrapf(err, "cannot add inotify watcher for '%s'", dir)
			me.errs <- err
		}
	}

	// main control loop
	var wg0 sync.WaitGroup
	ticker := time.NewTicker(cfg.Cnt.UpdateInterval * time.Second)

	// semaphore to ensure that only one content update run is done at any time
	sema := make(chan struct{}, 1)

	defer func() {
		notify.Stop(chgs)
		close(chgs)
		ticker.Stop()
		close(me.errs)
		close(me.updNotif)
		close(me.upd)
		close(sema)
		log.Trace("notifier stopped")
	}()

	for {
		select {
		case chg := <-chgs:
			// receive inotify events
			me.mutChanges.Lock()
			me.changes = append(me.changes, chg)
			me.mutChanges.Unlock()

		case <-ticker.C:
			// periodic update trigger
			wg0.Add(1)
			go func() {
				sema <- struct{}{}
				defer func() {
					<-sema
					wg0.Done()
				}()

				me.processChanges(ctx, cfg)
			}()

		case <-ctx.Done():
			// stop main control loop after last changes are processed
			wg0.Wait()
			return
		}
	}
}

// errors returns a receive-only channel for errors from notifier
func (me *notifier) errors() <-chan error {
	return me.errs
}

// updateNotification returns a receive-only channel to notify about updates
func (me *notifier) updateNotification() <-chan UpdateNotification {
	return me.updNotif
}

// processChanges detects which files need to either be deleted from or added
// to the muserv content based on the file system changes that have been
// observed by inotify. The DB is adjusted accordingly.
func (me *notifier) processChanges(ctx context.Context, cfg config.Cfg) {
	log.Trace("processing file system notifications ...")

	// check if there are changes at all. If yes copy changes to local table
	// protected by a mutex to avoid inconsistencies
	noChanges := false
	var changes []notify.EventInfo
	me.mutChanges.Lock()
	if len(me.changes) > 0 {
		changes = make([]notify.EventInfo, len(me.changes))
		copy(changes, me.changes)
		me.changes = nil
	} else {
		noChanges = true
	}
	me.mutChanges.Unlock()
	if noChanges {
		log.Trace("no changes to process")
		return
	}
	log.Trace("changes occurred: processing ...")

	// map for storing changed paths that were already processed (for some
	// changes notify delivers the same path multiple times)
	processed := make(map[string]struct{})

	// determine the files that were changed (according to inotify) and that
	// are either contained in the muserv content (which is an indicator that
	// they might have to be deleted from the content) or in the music dir
	// (which is an indicator that they might have to be added to the content)
	var fiCnt, fiDir fileInfos
	for _, chg := range changes {
		// don't process a changed path twice
		if _, processed := processed[chg.Path()]; processed {
			continue
		}
		processed[chg.Path()] = struct{}{}

		log.Tracef("%s :: %s", chg.Event().String(), chg.Path())

		// collect all changed files that are contained in music dir
		exists, err := f.Exists(chg.Path())
		if err != nil {
			err = errors.Wrapf(err, "cannot process changed path '%s'", chg.Path())
			log.Error(err)
			continue
		}
		if exists {
			// if it's a directory: Recursively expand it to the (supported)
			// files that are contained in that directory. Otherwise, go
			// forward with the single file
			isDir, err := f.IsDir(chg.Path())
			if err != nil {
				err = errors.Wrapf(err, "cannot process changed path '%s'", chg.Path())
				log.Error(err)
				continue
			}
			if isDir {
				fiDir = append(fiDir, *filesFromDirs([]string{chg.Path()})...)
			} else {
				if !isDir {
					if config.IsValidTrackFile(chg.Path()) {
						fiDir = append(fiDir, newTrackInfo(chg.Path(), 0))
					}
					if config.IsValidPlaylistFile(chg.Path()) {
						fiDir = append(fiDir, newPlaylistInfo(chg.Path(), 0))
					}
				}
			}
		}

		// collect all changed tracks that are contained in the content
		fiCnt = append(fiCnt, *me.filesByPaths([]string{chg.Path()})...)
	}

	// determine files to be deleted from or added to the content. fiCnt and
	// fiDir can contain duplicates. Thus, these duplicates must be removed
	// before diff is executed.
	sort.Sort(fiCnt)
	fiCnt.removeDuplicates()
	sort.Sort(fiDir)
	fiDir.removeDuplicates()
	fiDel, fiAdd := diff(fiCnt, fiDir)

	// create channel to notify server about finalized update
	updated := make(chan uint32)
	// close channel after update is done (this implicitly notifies the server
	// that the updated is done)
	defer close(updated)

	// notify server before update is executed
	me.updNotif <- UpdateNotification{
		Update:  func() { me.upd <- struct{}{} },
		Updated: updated,
	}
	<-me.upd

	// apply changes to content and report back the number of changed, deleted
	// or added objects
	var count uint32
	var err error
	if count, err = me.update(ctx, &fiDel, &fiAdd); err != nil {
		me.errs <- err
		return
	}
	updated <- count

	log.Trace("file system notifications processed")
}
