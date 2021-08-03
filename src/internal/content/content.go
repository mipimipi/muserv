package content

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"sync"

	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/go-utils/file"
	"gitlab.com/mipimipi/muserv/src/internal/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var log *l.Entry = l.WithFields(l.Fields{"srv": "content"})

// values of the BrowseFlag attribute of the ContentDirectory service
const (
	ModeMetadata = "BrowseMetadata"
	ModeChildren = "BrowseDirectChildren"
)

// root folder for music and picture requests
// note: they must end with a slash!
const (
	MusicFolder   = "/music/"
	PictureFolder = "/pictures/"
)

// status implements the content status
type status struct {
	overall string
	update  struct {
		task        string
		total, done int
	}
}

// status value of content
const (
	statusWaiting  = "waiting"
	statusRunning  = "running"
	statusUpdating = "updating"
)

// idGenerator return a function that creates a new object ID when it's called
func idGenerator() func() ObjID {
	var id ObjID
	return func() ObjID { id++; return id }
}

// Content contains the different muserv content objects, such as tracks,
// albums, hierarchies and methods to management them
type Content struct {
	status         status           // content status
	updater        updater          // regular content updates
	root           container        // root object
	objects        objects          // all objects
	albums         albums           // all albums
	folders        folders          // all folders
	pictures       pictures         // all pictures
	playlists      playlists        // all playlists
	tracks         tracks           // all tracks
	newID          func() ObjID     // object ID generator
	cfg            *config.Cfg      // muserv configuration
	extMusicPath   string           // external, virtual music path
	extPicturePath string           // external, virtual picture path
	updCounts      map[ObjID]uint32 // update counter per container object
}

// New creats a new Content instance
func New(cfg *config.Cfg) (cnt *Content, err error) {
	log.Trace("creating content object ...")

	addr, err := utils.IPaddr()
	if err != nil {
		err = errors.Wrap(err, "cannot create content since IP address cannot be determined")
		log.Fatal(err)
		return
	}

	// assemble URLs for music and pictures
	musicURL := url.URL{
		Scheme: "http",
		Path:   MusicFolder,
	}
	pictureURL := url.URL{
		Scheme: "http",
		Path:   PictureFolder,
	}
	if cfg.UPnP.Port == 0 {
		musicURL.Host = addr.String()
		pictureURL.Host = addr.String()
	} else {
		musicURL.Host = fmt.Sprintf("%s:%d", addr.String(), cfg.UPnP.Port)
		pictureURL.Host = fmt.Sprintf("%s:%d", addr.String(), cfg.UPnP.Port)

	}

	cnt = &Content{
		objects:        make(objects),
		albums:         make(albums),
		folders:        make(folders),
		pictures:       pictures{data: make(map[uint64]*[]byte)},
		playlists:      make(playlists),
		tracks:         make(tracks),
		newID:          idGenerator(),
		cfg:            cfg,
		extMusicPath:   musicURL.String(),
		extPicturePath: pictureURL.String(),
		updCounts:      make(map[ObjID]uint32),
	}
	cnt.updater = newUpdater(cfg.Cnt.UpdateMode, cnt.filesByPaths, cnt.update)

	// create the root object and its direct children (the hierarchy containers)
	cnt.makeTree()

	cnt.status.overall = statusWaiting

	log.Trace("content object created ...")
	return
}

// Browse implements the Browse SOAP action of the ContentDirectory service
func (me *Content) Browse(id ObjID, mode string, start, wanted uint32) (result string, returned, total uint32, err error) {
	// requested object must exist
	obj, exists := me.objects[id]
	if !exists {
		err = fmt.Errorf("no object found for id %d", id)
		log.Error(err)
		return
	}

	// if children are requested, the requested object must be a container
	if mode == ModeChildren && !obj.isContainer() {
		err = fmt.Errorf("object %d is no container but browse mode is 'BrowseDirectChildren'", id)
		log.Error(err)
		return
	}

	// calculate the requested index range
	var first, last int
	if obj.isContainer() {
		first, last = indices(start, wanted, obj.(container).numChildren())
	}

	// marshal the result as DIDL-Lite
	didl := obj.marshal(mode, first, last)
	didl = append(
		append(
			[]byte(didlStartElem),
			didl...,
		),
		[]byte(didlEndElem)...,
	)
	result = string(didl)

	// set values for the output attributes NumberReturned and TotalMatches
	if mode == ModeMetadata {
		returned, total = 1, 1
	} else {
		returned, total = uint32(last-first), uint32(obj.(container).numChildren())
	}

	return
}

// ContainerUpdateIDs assembles the new value for the state variable
// ContainerUpdateIDs
func (me *Content) ContainerUpdateIDs() (updates string) {
	for id, count := range me.updCounts {
		updates += fmt.Sprintf(",%d,%d", id, count)
	}
	if len(updates) > 0 {
		updates = updates[1:]
	}
	return
}

// Errors returns a receive-only channel for errors that occur during the
// regular update
func (me *Content) Errors() <-chan error {
	return me.updater.errors()
}

// InitialUpdate executes a one-time content update after muserv has been started
func (me *Content) InitialUpdate(ctx context.Context) (err error) {
	// set status
	me.status.overall = statusUpdating
	me.status.update.task = ""
	me.status.update.total = 0
	me.status.update.done = 0

	// extract config from context
	cfg := ctx.Value(config.KeyCfg).(config.Cfg)

	// get changes that must be applied to content
	tDel, tAdd := fullScan(cfg.Cnt.MusicDirs, me.filesByPaths)

	// update content
	_, err = me.update(ctx, tDel, tAdd)
	return
}

// Picture returns the picture with the given ID. If it doesn't exist, nil is
// returned
func (me *Content) Picture(id uint64) *[]byte {
	return me.pictures.get(id)
}

// ResetCtrUpdCounts resets the ContainerUpdateIDValues for all container
// objects
func (me *Content) ResetCtrUpdCounts() {
	me.root.resetUpdCount()
}

// Run starts the regular content updates
func (me *Content) Run(ctx context.Context, wg *sync.WaitGroup) {
	me.updater.run(ctx, wg)
	me.status.overall = statusRunning
}

// Trackpath return the path of the music track with the object id id. An error
// is returned if the track cannot be found
func (me *Content) Trackpath(id uint64) (string, error) {
	obj, exists := me.objects[ObjID(id)]
	if !exists {
		return "", fmt.Errorf("an object with id %d could not be found", id)
	}
	return obj.(*track).path, nil
}

// UpdateNotification returns a receive-only channel to notify about updates
func (me *Content) UpdateNotification() <-chan UpdateNotification {
	return me.updater.updateNotification()
}

// WriteStatus writes the content status to w
func (me *Content) WriteStatus(w io.Writer) {
	switch me.status.overall {
	case statusWaiting:
		fmt.Fprint(w, "Waiting ...\n")

	case statusRunning:
		fmt.Fprint(w, "    Content:\n")
		fmt.Fprintf(w, "    %6d tracks\n", len(me.tracks))
		fmt.Fprintf(w, "    %6d albums\n", len(me.albums))
		fmt.Fprintf(w, "    %6d playlists\n\n", len(me.playlists))
		// memory consumption
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		message.NewPrinter(language.English).Fprintf(w, "    Memory consumption: %d Bytes\n", m.HeapAlloc)

	case statusUpdating:
		fmt.Fprint(w, "   Updating content ...\n")
		if me.status.update.total > 0 {
			fmt.Fprintf(w,
				"        %s %d tracks, %d done (%.2f%%)\n",
				me.status.update.task,
				me.status.update.total,
				me.status.update.done,
				float64(100*me.status.update.done)/float64(me.status.update.total))
		}
	}
}

// filesByPaths returns all files (i.e. tracks and playlists) whose filepath
// begins with a path from paths
func (me *Content) filesByPaths(paths []string) *fileInfos {
	var fis fileInfos
	for p, fi := range me.tracks {
	L0:
		for _, path := range paths {
			if isSub, _ := file.IsSub(p, path); isSub {
				fis = append(fis, newTrackInfo(p, fi.lastChange))
				break L0
			}
		}
	}
	for p, fi := range me.playlists {
	L1:
		for _, path := range paths {
			if isSub, _ := file.IsSub(path, p); isSub {
				fis = append(fis, newPlaylistInfo(p, fi.lastChange))
				break L1
			}
		}
	}
	return &fis
}

// update updates the muserv content. fiDel and fiAdd contain the files that
// must be deleted (fiDel) or added (fiAdd). count contains the number of object
// changes that happened during content update
func (me *Content) update(ctx context.Context, fiDel, fiAdd *fileInfos) (count uint32, err error) {
	log.Trace("updating content ...")

	// set status
	me.status.overall = statusUpdating
	me.status.update.task = ""
	me.status.update.total = 0
	me.status.update.done = 0

	// initialize container update counter
	me.updCounts = make(map[ObjID]uint32)

	// delete files
	if err = me.procUpdates(ctx, &count, fiDel,
		func(wg *sync.WaitGroup, count *uint32, pli playlistInfo) error { return me.delPlaylist(wg, count, pli) },
		func(wg *sync.WaitGroup, count *uint32, ti trackInfo) error { return me.delTrack(wg, count, ti) },
	); err != nil {
		return
	}

	// add files
	if err = me.procUpdates(ctx, &count, fiAdd,
		func(wg *sync.WaitGroup, count *uint32, pli playlistInfo) error { return me.addPlaylist(wg, count, pli) },
		func(wg *sync.WaitGroup, count *uint32, ti trackInfo) error { return me.addTrack(wg, count, ti) },
	); err != nil {
		return
	}

	// remove obsolete objects such as cover pictures that are no longer
	// required
	me.cleanup()

	// set status
	me.status.overall = statusRunning

	log.Trace("content updated")

	return
}

// makeTree creates the object tree. It creates a new root object and the level
// below (i.e. the hierarchy containers)
func (me *Content) makeTree() {
	log.Trace("making root object ...")

	me.root = newCtr(me, 0, "root")
	me.objects.add(me.root)

	// create one generic container object as direct children of the root object
	// - one for each configured hierarchy
	for i, h := range me.cfg.Cnt.Hiers {
		hier := newCtr(me, me.newID(), h.Name)
		hier.sf = []string{fmt.Sprintf("%02d", i)}
		me.root.addChild(hier)
		// set the comparison functions for the sorting of child objects
		hier.setComparison(h.Levels[0].Comparisons())
		me.objects.add(hier)
	}
	index := len(me.cfg.Cnt.Hiers)
	// - create playlist hierarchy
	if me.cfg.Cnt.ShowPlaylists {
		hier := newCtr(me, me.newID(), me.cfg.Cnt.PlaylistHierName)
		hier.sf = []string{fmt.Sprintf("%02d", index)}
		me.root.addChild(hier)
		me.objects.add(hier)
		index++
	}
	// - create folder hierarchy
	if me.cfg.Cnt.ShowFolders {
		hier := newCtr(me, me.newID(), me.cfg.Cnt.FolderHierName)
		hier.sf = []string{fmt.Sprintf("%02d", index)}
		me.root.addChild(hier)
		me.objects.add(hier)
	}

	log.Trace("made root object")
}

// cleanup removes obsolete onjects
func (me *Content) cleanup() {
	// remove obsolete pictures from picture map
	newPics := make(map[uint64]*[]byte)
	for _, t := range me.tracks {
		if !t.picID.valid {
			continue
		}
		if me.pictures.data[t.picID.id] != nil {
			newPics[t.picID.id] = me.pictures.data[t.picID.id]
		}
	}
	me.pictures.data = newPics
}

func (me *Content) procUpdates(ctx context.Context, count *uint32, fis *fileInfos,
	procPlaylistUpdate func(*sync.WaitGroup, *uint32, playlistInfo) error,
	procTrackUpdate func(*sync.WaitGroup, *uint32, trackInfo) error) (err error) {
	if len(*fis) == 0 {
		log.Trace("no updates to process")
		return
	}

	log.Tracef("processing %d updates ...", len(*fis))

	// set update status values
	me.status.update.task = "processing updates"
	me.status.update.total = len(*fis)

	fInfos := make(chan fileInfo)
	go func() {
		for _, fi := range *fis {
			fInfos <- fi
		}
		close(fInfos)
	}()

	var wg sync.WaitGroup

L:
	for {
		select {
		case fi, ok := <-fInfos:
			if !ok {
				log.Tracef("%d updates processed", len(*fis))
				break L
			}
			switch fi.kind() {
			case infoPlaylist:
				if me.cfg.Cnt.ShowPlaylists {
					if err = procPlaylistUpdate(&wg, count, fi.(playlistInfo)); err != nil {
						log.Fatal(err)
					}
				}
			case infoTrack:
				if err = procTrackUpdate(&wg, count, fi.(trackInfo)); err != nil {
					log.Fatal(err)
				}
			default:
				log.Errorf("unknown fileInfo type %d: cannot process update", fi.kind())
			}
			me.status.update.done++

		case <-ctx.Done():
			log.Trace("processing updates interrupted")
			break L
		}
	}

	wg.Wait()
	return
}

func (me *Content) addPlaylist(wg *sync.WaitGroup, count *uint32, pli playlistInfo) (err error) {
	// parse playlist file and create a playlist object
	var pl *playlist
	if pl, err = newPlaylist(me, wg, count, pli); err != nil {
		log.Fatal(err)
		return
	}

	// if the playlist container has items/children, add it to the playlist
	// hierarchy node
	if pl.numChildren() > 0 {
		me.root.childByIndex(len(me.cfg.Cnt.Hiers)).(container).addChild(pl)
	}

	return
}

func (me *Content) delPlaylist(wg *sync.WaitGroup, count *uint32, pli playlistInfo) (err error) {
	// get corresponding playlist object
	pl, exists := me.playlists[pli.path()]
	if !exists {
		return
	}

	// count deletion of playlist object
	*count++
	// remove from playlists
	delete(me.playlists, pli.path())
	// remove from objects
	delete(me.objects, pl.id())
	// remove from hierarchies
	pl.parent().delChild(pl)

	// remove playlist items (i.e. the corresponding track references)
	for i := 0; i < pl.numChildren(); i++ {
		tRef := pl.childByIndex(i).(*trackRef)
		// remove from objects
		delete(me.objects, tRef.id())
		// remove from the reference list of the corresponding track
		tRef.track.delTrackRef(tRef)
	}
	pl.delChildren()

	return
}

func (me *Content) addTrack(wg *sync.WaitGroup, count *uint32, ti trackInfo) (err error) {
	t, err := newTrack(me, wg, count, ti)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// add t to all configured hierarchies and (if configured) the folder
	// hierarchy, but don't add it to the playlists hierarchy
	for i := 0; i < len(me.cfg.Cnt.Hiers); i++ {
		if err := me.addTrackToHierarchy(count, &me.cfg.Cnt.Hiers[i], me.root.childByIndex(i).(container), t); err != nil {
			return err
		}
	}
	if me.cfg.Cnt.ShowFolders {
		// determine the right hierarchy index of the folder hierarchy and add
		// t to the hierarchy
		if me.cfg.Cnt.ShowPlaylists {
			me.addTrackToFolderHierarchy(count, me.root.childByIndex(len(me.cfg.Cnt.Hiers)+1).(container), t)
		} else {
			me.addTrackToFolderHierarchy(count, me.root.childByIndex(len(me.cfg.Cnt.Hiers)).(container), t)
		}
	}

	return
}

func (me *Content) delTrack(wg *sync.WaitGroup, count *uint32, ti trackInfo) (err error) {
	// get corresponding track object
	t, exists := me.tracks[ti.path()]
	if !exists {
		return
	}
	// count deletion of track object
	*count++
	// remove from tracks
	delete(me.tracks, ti.path())
	// remove from objects
	delete(me.objects, t.id())
	// remove from albums
	a, exists := me.albums[t.albumKey()]
	if exists {
		a.delChild(t)
		if a.numChildren() == 0 {
			delete(me.objects, a.id())
			delete(me.albums, a.key())
			// count deletion of album object
			*count++
		}
	}
	// remove from hierarchies
	for _, tRef := range t.refs {
		var obj object = tRef
		for parent := tRef.parent(); parent.parent() != nil; parent = parent.parent() {
			delete(me.objects, obj.id())
			// count object deletion
			*count++

			// delete obj from parent and stop propagating this deletion
			// upwards the hierarchy if there are still other children
			parent.delChild(obj)
			if parent.numChildren() > 0 {
				break
			}

			// prepare for next loop
			obj = parent
		}
	}
	return
}

// traceUpdate increases the update counter for the container object with the
// given id
func (me *Content) traceUpdate(id ObjID) {
	_, exists := me.updCounts[id]
	if !exists {
		me.updCounts[id] = 1
		return
	}
	me.updCounts[id]++
}

const space = "--------------------------------------------------------------------------------"

func strOfLength(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// AlbumsWithInconsistentTrackNumbers determines albums that have either
// overlapping track numbers or that have gaps in the track numbering
func (me *Content) AlbumsWithInconsistentTrackNumbers(w io.Writer) {
	fmt.Fprint(w, "Albums with inconsistent track numbers:\n\n")
	fmt.Fprintf(w, "%-18s %-30s %-30s\n", "Genre", "AlbumArtist", "Album")
	fmt.Fprintf(w, "%s\n", space)

	for _, a := range me.albums {
		if a.numChildren() == 0 {
			return
		}

		nums := make(map[int]struct{})
		t := a.childByIndex(0).(*track)
		consistent := true

		for i := 0; i < a.numChildren() && consistent; i++ {
			t := a.childByIndex(i).(*track)
			if _, exists := nums[t.tags.trackNo]; exists {
				fmt.Fprintf(w, "%-18s %-30s %-30s\n", strOfLength(t.tags.genres[0], 18), strOfLength(t.tags.albumArtists[0], 30), strOfLength(t.tags.album, 30))
				consistent = false
			} else {
				nums[t.tags.trackNo] = struct{}{}
			}
		}

		if !consistent {
			continue
		}

		for i := 0; i < len(nums) && consistent; i++ {
			if _, exists := nums[i+1]; !exists {
				fmt.Fprintf(w, "%-18s %-30s %-30s\n", strOfLength(t.tags.genres[0], 18), strOfLength(t.tags.albumArtists[0], 30), strOfLength(t.tags.album, 30))
				consistent = false
			}
		}
	}
}

// AlbumsWithMultipleCovers determines albums that contain tracks that have not the
// same cover picture
func (me *Content) AlbumsWithMultipleCovers(w io.Writer) {
	fmt.Fprint(w, "Albums with multiple covers:\n\n")
	fmt.Fprintf(w, "%-18s %-30s %-30s\n", "Genre", "AlbumArtist", "Album")
	fmt.Fprintf(w, "%s\n", space)

	for _, a := range me.albums {
		var picID nonePicID
	L:
		for i := 0; i < a.numChildren(); i++ {
			t := a.childByIndex(i).(*track)
			if i == 0 {
				picID = t.picID
				continue
			}
			if t.picID.valid != picID.valid || t.picID.id != picID.id {
				fmt.Fprintf(w, "%-18s %-30s %-30s\n", strOfLength(t.tags.genres[0], 18), strOfLength(t.tags.albumArtists[0], 30), strOfLength(t.tags.album, 30))
				break L
			}
		}
	}
}

// InconsistentAlbums checks if albums with the same title from the same album
// artists have the same year and compilation flag assigned. If that's not the
// case, that's an indicator for an inconsistency and the album data is
// printed to w
func (me *Content) InconsistentAlbums(w io.Writer) {
	albums := make(map[string]struct {
		albumArtists []string
		year         int
		compilation  bool
	})
	incons := make(map[string]bool)

	fmt.Fprint(w, "Potentially inconsistent albums:\n")

	for _, t := range me.tracks {
		key := fmt.Sprintf("%v|%s", t.tags.albumArtists, t.tags.album)
		album, exists := albums[key]
		if !exists {
			albums[key] = struct {
				albumArtists []string
				year         int
				compilation  bool
			}{
				albumArtists: t.tags.albumArtists,
				year:         t.tags.year,
				compilation:  t.tags.compilation,
			}
			continue
		}
		if album.year != t.tags.year || album.compilation != t.tags.compilation {
			_, exists := incons[key]
			if !exists {
				fmt.Fprintf(w, "Genre: '%v', albumArtist: '%v', Album: '%s',  track: '%s' - differences: ", t.tags.genres, t.tags.albumArtists, t.tags.album, t.name())
				if album.year != t.tags.year {
					fmt.Fprint(w, "years ")
				}
				if album.compilation != t.tags.compilation {
					fmt.Fprint(w, "compilation flag ")
				}
				fmt.Fprint(w, "\n")
				incons[key] = true
			}
			continue
		}
	}
}

// TracksWithoutAlbum determines tracks that do not have a album tag assigned
func (me *Content) TracksWithoutAlbum(w io.Writer) {
	fmt.Fprint(w, "Tracks without album:\n")
	for _, t := range me.tracks {
		if len(t.tags.album) == 0 {
			fmt.Fprintf(w, "Genre: '%v', albumArtists: '%v', album: '%s',  track: '%s'\n", t.tags.genres, t.tags.albumArtists, t.tags.album, t.name())
		}
	}
}

// TracksWithoutCover determines tracks that do not have a cover picture assigned
func (me *Content) TracksWithoutCover(w io.Writer) {
	fmt.Fprint(w, "Tracks without cover pictures:\n")
	for _, t := range me.tracks {
		if !t.picID.valid {
			fmt.Fprintf(w, "Genre: '%v', albumArtists: '%v', album: '%s',  track: '%s'\n", t.tags.genres, t.tags.albumArtists, t.tags.album, t.name())
		}
	}
}
