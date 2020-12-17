package content

import (
	"context"
	"sort"
	"sync"

	"gitlab.com/mipimipi/go-utils/file"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// UpdateNotification is used to inform the caller about updates. It contains a
// function to execute the update and a channel to inform the caller that the
// update was done
type UpdateNotification struct {
	// Update triggers update
	Update func()
	// Updated provides the info update was done and how many objects were
	// changed, deleted or added
	Updated chan uint32
}

// updater is the interface that must be implemented by content updaters
type updater interface {
	// run updater
	run(context.Context, *sync.WaitGroup)

	// communicate errors
	errors() <-chan error

	// inform caller about updates and provide a channel to communicate that
	// the update was done
	updateNotification() <-chan UpdateNotification
}

// content update modes
const (
	updModeNotify = "notify" // update via fsnotify
	updModeScan   = "scan"   // update via regular scans
)

// updaters maps the update mode to its implementations
var updaters = map[string](func(func(string) *trackpaths, func(context.Context, *trackpaths, *trackpaths) (uint32, error)) updater){
	updModeNotify: func(tracksByPath func(string) *trackpaths, update func(context.Context, *trackpaths, *trackpaths) (uint32, error)) updater {
		return newNotifier(tracksByPath, update)
	},
	updModeScan: func(tracksByPath func(string) *trackpaths, update func(context.Context, *trackpaths, *trackpaths) (uint32, error)) updater {
		return newScanner(tracksByPath, update)
	},
}

// newUpdater creates an updater instance based on cfg.UpdateMode
func newUpdater(updMode string, tracksByPath func(string) *trackpaths, update func(context.Context, *trackpaths, *trackpaths) (uint32, error)) updater {
	upd, ok := updaters[updMode]
	if ok {
		return upd(tracksByPath, update)
	}
	return nil
}

// tracksFromDir recursively determines all valid tracks of the folder tree
// below dir. Valid in this context means that the tracks have a mime type that
// is supported by muserv
func tracksFromDir(dir string) *trackpaths {
	var tps trackpaths

	log.Tracef("reading tracks from '%s' ...", dir)

	var trackpaths = make(chan trackpath)
	defer close(trackpaths)

	// filter: only accepts music files that have the supported mime types
	filter := func(srcFile file.Info, vp file.ValidPropagate) (bool, file.ValidPropagate) {
		if !srcFile.IsDir() && srcFile.Mode().IsRegular() {
			if config.IsValidAudioFile(srcFile.Path()) {
				trackpaths <- newTrackpath(srcFile.Path(), 0)
				return true, file.NoneFromSuper
			}
		}
		return false, file.NoneFromSuper
	}

	// collect valid tracks
	go func() {
		// collect results into results array
		for tp := range trackpaths {
			tps = append(tps, tp)
		}
	}()

	// determine files according to filter
	root, err := file.Stat(dir)
	if err != nil {
		log.Error(err)
		return &tps
	}
	_ = file.Find([]file.Info{root}, filter, 1)

	sort.Sort(tps)

	log.Tracef("read tracks from '%s'", dir)
	return &tps
}

// diff takes an array of tracks from existing content (tCnt) and a list of
// tracks from the music dir (tDir), determines the differences and returns two
// arrays. One contains the tracks that must be removed from the muserv content
// (tDel), the other one contains the tracks that must be added to the content
// to make it consistent to the music dir (tAdd).
func diff(tCnt trackpaths, tDir trackpaths) (tDel, tAdd trackpaths) {
	if len(tCnt) == 0 {
		tAdd = append(tAdd, tDir...)
		return
	}
	if len(tDir) == 0 {
		tDel = append(tDel, tCnt...)
		return
	}

	for i, j := 0, 0; i < len(tCnt) || j < len(tDir); {
		if i >= len(tCnt) || j < len(tDir) && tCnt[i].path > tDir[j].path {
			tAdd = append(tAdd, tDir[j])
			j++
			continue
		}
		if j >= len(tDir) || i < len(tCnt) && tCnt[i].path < tDir[j].path {
			tDel = append(tDel, tCnt[i])
			i++
			continue
		}
		if tCnt[i].path == tDir[j].path {
			// check is files have changed though the name didn't
			if tCnt[i].lastChanged() < tDir[j].lastChanged() {
				tDel = append(tDel, tCnt[i])
				tAdd = append(tAdd, tDir[j])
			}
			i++
			j++
			continue
		}
	}

	return
}

// fullScan (a) reads all tracks from the muserv content and from the music dir
//      and (b) determines and returns the differences (i.e. which tracks must
// 	            be deleted from and added to the content hierarchies to make it
//              consistent with the music dir)
func fullScan(musicDir string, tracksByPath func(string) *trackpaths) (*trackpaths, *trackpaths) {
	log.Trace("scanning ...")

	// get changes / differences between music directory and muserv content
	cntData := make(chan *trackpaths)
	dirData := make(chan *trackpaths)

	// retrieve tracks from content
	go func(ret chan<- *trackpaths) {
		ret <- tracksByPath("")
	}(cntData)

	// retrieve tracks from music dir
	go func(musicDir string, ret chan<- *trackpaths) {
		ret <- tracksFromDir(musicDir)
	}(musicDir, dirData)

	tCnt := <-cntData
	tDir := <-dirData

	tDel, tAdd := diff(*tCnt, *tDir)

	log.Trace("scanning done")

	return &tDel, &tAdd
}
