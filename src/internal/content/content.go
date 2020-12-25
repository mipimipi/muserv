package content

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"

	"github.com/dhowden/tag"
	"github.com/pkg/errors"
	l "github.com/sirupsen/logrus"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/muserv/src/internal/config"
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
		tracks:         make(tracks),
		newID:          idGenerator(),
		cfg:            cfg,
		extMusicPath:   musicURL.String(),
		extPicturePath: pictureURL.String(),
		updCounts:      make(map[ObjID]uint32),
	}
	cnt.updater = newUpdater(cfg.Cnt.UpdateMode, cnt.tracksByPath, cnt.update)

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

	// set values for the outbput attributes NumberReturned and TotalMatches
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

// InitialUpdate executes a one-time content update after muserv has been started
func (me *Content) InitialUpdate(ctx context.Context) (err error) {
	// set status
	me.status.overall = statusUpdating
	me.status.update.task = ""
	me.status.update.total = 0
	me.status.update.done = 0

	// extract config from context
	cfg := ctx.Value(config.KeyCfg).(config.Cfg)

	// get changes that must be applied to DB
	tDel, tAdd := fullScan(cfg.Cnt.MusicDir, me.tracksByPath)

	// update content
	_, err = me.update(ctx, tDel, tAdd)
	return
}

// Run starts the regular content updates
func (me *Content) Run(ctx context.Context, wg *sync.WaitGroup) {
	me.updater.run(ctx, wg)
	me.status.overall = statusRunning
}

// UpdateNotification returns a receive-only channel to notify about updates
func (me *Content) UpdateNotification() <-chan UpdateNotification {
	return me.updater.updateNotification()
}

// Errors returns a receive-only channel for errors from the regular update
func (me *Content) Errors() <-chan error {
	return me.updater.errors()
}

// WriteStatus writes the content status to w
func (me *Content) WriteStatus(w io.Writer) {
	switch me.status.overall {
	case statusWaiting:
		fmt.Fprint(w, "waiting\n")

	case statusRunning:
		fmt.Fprint(w, "running\n")
		fmt.Fprintf(w, "%6d tracks\n", len(me.tracks))
		fmt.Fprintf(w, "%6d albums\n", len(me.albums))
		fmt.Fprintf(w, "%6d cover pictures\n", len(me.pictures.data))

	case statusUpdating:
		fmt.Fprint(w, "updating\n")
		if me.status.update.total > 0 {
			fmt.Fprintf(w,
				"    %s %d tracks, %d done (%.2f%%)\n",
				me.status.update.task,
				me.status.update.total,
				me.status.update.done,
				float64(100*me.status.update.done)/float64(me.status.update.total))
		}
	}
}

// tracksByPath returns all tracks whose filepath begins with path
func (me *Content) tracksByPath(path string) *trackpaths {
	var tps trackpaths
	for p, t := range me.tracks {
		if len(path) == 0 || len(path) <= len(p) && path == p[:len(path)] {
			tps = append(tps, newTrackpath(p, t.lastChange))
		}
	}
	return &tps
}

// update updates the muserv content. tDel and tAdd contain the track files
// that must be deleted (tDel) or added (tAdd). count contains the number of
// object changes that happened during content update
func (me *Content) update(ctx context.Context, tDel, tAdd *trackpaths) (count uint32, err error) {
	log.Trace("updating content ...")

	// set status
	me.status.overall = statusUpdating
	me.status.update.task = ""
	me.status.update.total = 0
	me.status.update.done = 0

	// initialize container update counter
	me.updCounts = make(map[ObjID]uint32)

	// delete obsolete tracks
	me.delTracks(ctx, &count, tDel)

	// add new tracks
	if err = me.addTracks(ctx, &count, tAdd); err != nil {
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

	root := newCtr(me, 0, "root")
	me.objects.add(root)
	me.root = root

	// create one generic container object as direct children of the root object
	// - one for each configured hierarchy
	for i, h := range me.cfg.UPnP.Hiers {
		hier := newCtr(me, me.newID(), h.Name)
		hier.sf = []string{fmt.Sprintf("%02d", i)}
		me.objects.add(hier)
		me.root.addChild(hier)
		// set the comparison functions for the sorting of child objects
		hier.setComparison(h.Levels[0].Comparisons())
	}
	// create folder hierarchy
	if me.cfg.UPnP.ShowFolders {
		hier := newCtr(me, me.newID(), me.cfg.UPnP.FolderHierName)
		hier.sf = []string{fmt.Sprintf("%02d", len(me.cfg.UPnP.Hiers))}
		me.objects.add(hier)
		me.root.addChild(hier)
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

// addTracks adds tracks to muserv content. count is set to the number of object
// changes that happened during that activity
func (me *Content) addTracks(ctx context.Context, count *uint32, tps *trackpaths) (err error) {
	if len(*tps) == 0 {
		log.Trace("no tracks to add")
		return
	}

	log.Tracef("adding %d tracks ...", len(*tps))

	// set update status values
	me.status.update.task = "adding"
	me.status.update.total = len(*tps)

	tpaths := make(chan trackpath)
	go func() {
		for _, tp := range *tps {
			tpaths <- tp
		}
		close(tpaths)
	}()

	var wg sync.WaitGroup

L:
	for {
		select {
		case tp, ok := <-tpaths:
			if !ok {
				log.Tracef("%d tracks added", len(*tps))
				break L
			}
			t, err := me.trackFromPath(&wg, count, tp)
			if err != nil {
				log.Fatal(err)
				return err
			}
			for i := 0; i < me.root.numChildren(); i++ {
				if me.cfg.UPnP.ShowFolders && i == len(me.cfg.UPnP.Hiers) {
					me.addToFolderHierarchy(count, me.root.childByIndex(i).(container), t)
					continue
				}
				if err := me.addToHierarchy(count, &me.cfg.UPnP.Hiers[i], me.root.childByIndex(i).(container), t); err != nil {
					return err
				}
			}
			me.status.update.done++

		case <-ctx.Done():
			log.Trace("adding tracks interrupted")
			break L
		}
	}

	wg.Wait()
	return
}

// delTracks removes tracks to muserv content. count is set to the number of
// object changes that happened during that activity
func (me *Content) delTracks(ctx context.Context, count *uint32, tps *trackpaths) {
	if len(*tps) == 0 {
		log.Trace("no tracks to delete")
		return
	}

	log.Tracef("deleting %d tracks ...", len(*tps))

	// set update status values
	me.status.update.task = "deleting"
	me.status.update.total = len(*tps)

	tpaths := make(chan trackpath)
	go func() {
		for _, tp := range *tps {
			tpaths <- tp
		}
		close(tpaths)
	}()

L:
	for {
		select {
		case tp, ok := <-tpaths:
			if !ok {
				log.Tracef("%d tracks deleted", len(*tps))
				break L
			}

			// get corresponding track object
			t, exists := me.tracks[tp.path]
			if !exists {
				continue
			}
			// count deletion of track object
			*count++
			// remove from tracks
			delete(me.tracks, tp.path)
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
			me.status.update.done++

		case <-ctx.Done():
			log.Trace("deleting tracks interrupted")
			break L
		}

	}
}

// albumRefFromAlbum creates a new album reference object from an album
func (me *Content) albumRefFromAlbum(a *album, sfs []config.SortField) *albumRef {
	aRef := albumRef{
		newCtr(me, me.newID(), a.n),
		a,
	}
	aRef.marshalFunc = newAlbumRefMarshalFunc(aRef)
	aRef.k = a.k

	me.objects.add(&aRef)
	a.refs = append(a.refs, &aRef)

	// set sort fields of album reference
	if len(sfs) > 0 {
		aRef.sf = []string{}
		for _, sf := range sfs {
			var s string
			switch sf {
			case config.SortLastChange:
				s = fmt.Sprintf("%020d", a.lastChange)
			case config.SortTitle:
				s = a.n
			case config.SortYear:
				s = fmt.Sprintf("%d", a.year)
			}
			if len(s) > 0 {
				aRef.sf = append(aRef.sf, s)
			}
		}
	}

	return &aRef
}

// newAlbum creates a new album object
func (me *Content) newAlbum(key uint64) (a *album) {
	a = &album{
		newCtr(me, me.newID(), ""),
		0,
		false,
		[]string{},
		[]string{},
		0,
		[]*albumRef{},
	}
	a.k = key
	a.marshalFunc = newAlbumMarshalFunc(a, me.cfg.Cnt.MusicDir, me.extMusicPath, me.extPicturePath)

	me.objects.add(a)
	me.albums.add(a)

	return
}

// trackFromPath creates a new track object from a track filepath
func (me *Content) trackFromPath(wg *sync.WaitGroup, count *uint32, tp trackpath) (t *track, err error) {
	var (
		size    int64
		tags    *tags
		picture *tag.Picture
	)

	// get tags and picture
	if tags, picture, err = tp.metadata(me.cfg.Cnt.Separator); err != nil {
		err = errors.Wrapf(err, "cannot create track from filepath '%s'", tp.path)
		log.Fatal(err)
		return
	}

	// get size of music track
	size, err = tp.size()
	if err != nil {
		err = errors.Wrapf(err, "cannot create track from filepath '%s'", tp.path)
		log.Fatal(err)
		return nil, err
	}

	t = &track{
		newItm(me, me.newID(), tags.title),
		tags,
		nonePicID{0, false},
		tp.mimeType(),
		size,
		tp.lastChange(),
		tp.path,
		[]*trackRef{},
	}
	t.marshalFunc = newTrackMarshalFunc(t, me.cfg.Cnt.MusicDir, me.extMusicPath, me.extPicturePath)

	me.objects.add(t)
	me.tracks.add(t)

	// process picture
	wg.Add(1)
	go me.pictures.add(wg, picture, &t.picID)

	// count creation of track object
	*count++

	// add track to corresponding album. Create it if is doesn't exist.
	a, exists := me.albums[t.albumKey()]
	if !exists {
		a = me.newAlbum(t.albumKey())
		a.n = t.tags.album
		a.year = t.tags.year
		a.compilation = t.tags.compilation
		a.artists = t.tags.albumArtists
		a.composers = t.tags.composers
		a.lastChange = t.lastChange
	}
	a.addChild(t)
	// count change of album container
	*count++

	return
}

// trackRefFromTrack creates a new track reference object from a track
func (me *Content) trackRefFromTrack(t *track, sfs []config.SortField) *trackRef {
	tRef := trackRef{
		newItm(me, me.newID(), t.tags.title),
		t,
	}
	tRef.marshalFunc = newTrackRefMarshalFunc(tRef)

	me.objects.add(&tRef)
	t.refs = append(t.refs, &tRef)

	// set sort fields of track reference
	if len(sfs) > 0 {
		tRef.sf = []string{}
		for _, sf := range sfs {
			var s string
			switch sf {
			case config.SortDiscNo:
				s = fmt.Sprintf("%03d", t.tags.discNo)
			case config.SortLastChange:
				s = fmt.Sprintf("%020d", t.lastChange)
			case config.SortTitle:
				s = t.tags.title
			case config.SortTrackNo:
				s = fmt.Sprintf("%04d", t.tags.trackNo)
			case config.SortYear:
				s = fmt.Sprintf("%d", t.tags.year)
			}
			if len(s) > 0 {
				tRef.sf = append(tRef.sf, s)
			}
		}
	}

	return &tRef
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
// case, that's an indicator for an inconsistentcy and the album data is
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
