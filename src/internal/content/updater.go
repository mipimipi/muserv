package content

import (
	"context"
	"sort"
	"sync"

	f "gitlab.com/mipimipi/go-utils/file"
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
var updaters = map[string](func(func(string) *fileInfos, func(context.Context, *fileInfos, *fileInfos) (uint32, error)) updater){
	updModeNotify: func(tracksByPath func(string) *fileInfos, update func(context.Context, *fileInfos, *fileInfos) (uint32, error)) updater {
		return newNotifier(tracksByPath, update)
	},
	updModeScan: func(tracksByPath func(string) *fileInfos, update func(context.Context, *fileInfos, *fileInfos) (uint32, error)) updater {
		return newScanner(tracksByPath, update)
	},
}

// newUpdater creates an updater instance based on cfg.UpdateMode
func newUpdater(updMode string, tracksByPath func(string) *fileInfos, update func(context.Context, *fileInfos, *fileInfos) (uint32, error)) updater {
	upd, ok := updaters[updMode]
	if ok {
		return upd(tracksByPath, update)
	}
	return nil
}

// filesFromDir recursively determines all valid files of the folder tree
// below dir. Valid in this context means that the files have a mime type that
// is supported by muserv
func filesFromDir(dir string) *fileInfos {
	var fis fileInfos

	log.Tracef("reading tracks from '%s' ...", dir)

	var fileInfos = make(chan fileInfo)
	defer close(fileInfos)

	// filter: only accepts files that have the supported mime types
	filter := func(srcFile f.Info, vp f.ValidPropagate) (bool, f.ValidPropagate) {
		if !srcFile.IsDir() && srcFile.Mode().IsRegular() {
			if config.IsValidPlaylistFile(srcFile.Path()) {
				fileInfos <- newPlaylistInfo(srcFile.Path(), 0)
				return true, f.NoneFromSuper
			}
			if config.IsValidTrackFile(srcFile.Path()) {
				fileInfos <- newTrackInfo(srcFile.Path(), 0)
				return true, f.NoneFromSuper
			}
		}
		return false, f.NoneFromSuper
	}

	// collect valid tracks
	go func() {
		// collect results into results array
		for fi := range fileInfos {
			fis = append(fis, fi)
		}
	}()

	// determine files according to filter
	root, err := f.Stat(dir)
	if err != nil {
		log.Error(err)
		return &fis
	}
	_ = f.Find([]f.Info{root}, filter, 1)

	sort.Sort(fis)

	log.Tracef("read tracks from '%s'", dir)
	return &fis
}

// diff takes an array of files from existing content (fiCnt) and a list of
// files from the music dir (fiDir), determines the differences and returns two
// arrays. One contains the files that must be removed from the muserv content
// (fiDel), the other one contains the files that must be added to the content
// to make it consistent to the music dir (fiAdd).
func diff(fiCnt fileInfos, fiDir fileInfos) (fiDel, fiAdd fileInfos) {
	if len(fiCnt) == 0 {
		fiAdd = append(fiAdd, fiDir...)
		return
	}
	if len(fiDir) == 0 {
		fiDel = append(fiDel, fiCnt...)
		return
	}

	for i, j := 0, 0; i < len(fiCnt) || j < len(fiDir); {
		if i >= len(fiCnt) || j < len(fiDir) && fiCnt[i].path() > fiDir[j].path() {
			fiAdd = append(fiAdd, fiDir[j])
			j++
			continue
		}
		if j >= len(fiDir) || i < len(fiCnt) && fiCnt[i].path() < fiDir[j].path() {
			fiDel = append(fiDel, fiCnt[i])
			i++
			continue
		}
		if fiCnt[i].path() == fiDir[j].path() {
			// check is files have changed though the name didn't
			if fiCnt[i].lastChange() < fiDir[j].lastChange() {
				fiDel = append(fiDel, fiCnt[i])
				fiAdd = append(fiAdd, fiDir[j])
			}
			i++
			j++
			continue
		}
	}

	return
}

// fullScan (a) reads all files from the muserv content and from the music dir
//      and (b) determines and returns the differences (i.e. which files must
// 	            be deleted from and added to the content hierarchies to make it
//              consistent with the music dir)
func fullScan(musicDir string, filesByPath func(string) *fileInfos) (*fileInfos, *fileInfos) {
	log.Trace("scanning ...")

	// get changes / differences between music directory and muserv content
	cntData := make(chan *fileInfos)
	dirData := make(chan *fileInfos)

	// retrieve files from content
	go func(ret chan<- *fileInfos) {
		ret <- filesByPath(musicDir)
	}(cntData)

	// retrieve files from music dir
	go func(musicDir string, ret chan<- *fileInfos) {
		ret <- filesFromDir(musicDir)
	}(musicDir, dirData)

	fiCnt := <-cntData
	fiDir := <-dirData

	fiDel, fiAdd := diff(*fiCnt, *fiDir)

	log.Trace("scanning done")

	return &fiDel, &fiAdd
}
