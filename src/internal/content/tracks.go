package content

import (
	"fmt"
	"mime"
	"path"
	"strings"
	"sync"

	"github.com/mipimipi/tag"
	"github.com/pkg/errors"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// track represents a track object. For each music track, exactly one track
// object exists
type track struct {
	*itm
	tags       *tags               // tags of the track
	picID      nonePicID           // ID of the cover picture (can be "null")
	mimeType   string              // mime type of track file
	size       int64               // size of track file in bytes
	lastChange int64               // UNIX time of last change of track file
	path       string              // path of track file
	refs       map[ObjID]*trackRef // corresponding track references
}

// newTrack creates a new track object from a trackinfo
func newTrack(cnt *Content, wg *sync.WaitGroup, count *uint32, ti trackInfo) (t *track, err error) {
	var (
		tgs        *tags
		picture    *tag.Picture
		lastChange int64
		size       int64
	)

	// get tags and picture
	if tgs, picture, err = ti.metadata(cnt.cfg.Cnt.Separator); err != nil {
		err = errors.Wrapf(err, "cannot create track from filepath '%s'", ti.path())
		log.Fatal(err)
		return
	}
	// get size of track
	size = ti.size()
	// get last changed time of track
	lastChange = ti.lastChange()

	t = &track{
		newItm(cnt, cnt.newID(), tgs.title),
		tgs,
		nonePicID{0, false},
		ti.mimeType(),
		size,
		lastChange,
		ti.path(),
		make(map[ObjID]*trackRef),
	}
	t.marshalFunc = newTrackMarshalFunc(t, cnt.cfg.Cnt.MusicDir, cnt.extMusicPath, cnt.extPicturePath)

	cnt.tracks.add(t)
	cnt.objects.add(t)

	// process picture
	wg.Add(1)
	go cnt.pictures.add(wg, picture, &t.picID)

	// count creation of track object
	*count++

	// add track to corresponding album. Create it if is doesn't exist
	if len(t.tags.album) > 0 {
		a, exists := cnt.albums[t.albumKey()]
		if !exists {
			a = newAlbum(cnt, t.albumKey())
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
	}

	return
}

// newExtTrack creates a new track object for an external track (i.e. a track
// that is not stored in the file system but somewhere in the WWW)
func newExtTrack(cnt *Content, count *uint32, url, title string) (t *track, err error) {
	t = &track{
		newItm(cnt, cnt.newID(), title),
		&tags{},
		nonePicID{0, false},
		mime.TypeByExtension(path.Ext(url)),
		0,
		0,
		url,
		make(map[ObjID]*trackRef),
	}
	t.marshalFunc = newTrackMarshalFunc(t, cnt.cfg.Cnt.MusicDir, cnt.extMusicPath, cnt.extPicturePath)

	cnt.tracks.add(t)
	cnt.objects.add(t)

	// count creation of track object
	*count++

	return
}

// albumKey calculates the key of an album as FNV hash from album artists, album
// title, year and whether it's a compilation or not
func (me *track) albumKey() uint64 {
	return utils.HashUint64("%v%s%d%t", me.tags.albumArtists, me.tags.album, me.tags.year, me.tags.compilation)
}

// delTrackRef removes a track references fro the reference map
func (me *track) delTrackRef(tRef *trackRef) {
	delete(me.refs, tRef.id())

	// external tracks without any reference are obsolete: delete it from the
	// objects and tracks maps
	if me.isExternal() && len(me.refs) == 0 {
		delete(me.cnt.objects, me.id())
		delete(me.cnt.tracks, me.path)
	}
}

// isExternal returns true if the track is not a local track (i.e. its path
// starts with "http://" or "https://")
func (me *track) isExternal() bool {
	return strings.HasPrefix(me.path, "http://") || strings.HasPrefix(me.path, "https://")
}

// newTrackRef creates a new track reference object from a track
func (me *track) newTrackRef(sfs []config.SortField) *trackRef {
	tRef := trackRef{
		newItm(me.cnt, me.cnt.newID(), me.n),
		me,
	}
	tRef.marshalFunc = newTrackRefMarshalFunc(tRef)

	me.refs[tRef.id()] = &tRef
	me.cnt.objects.add(tRef)

	// set sort fields of track reference
	if len(sfs) > 0 {
		tRef.sf = []string{}
		for _, sf := range sfs {
			var s string
			switch sf {
			case config.SortDiscNo:
				s = fmt.Sprintf("%03d", me.tags.discNo)
			case config.SortLastChange:
				s = fmt.Sprintf("%020d", me.lastChange)
			case config.SortTitle:
				s = me.tags.title
			case config.SortTrackNo:
				s = fmt.Sprintf("%04d", me.tags.trackNo)
			case config.SortYear:
				s = fmt.Sprintf("%d", me.tags.year)
			}
			if len(s) > 0 {
				tRef.sf = append(tRef.sf, s)
			}
		}
	}

	return &tRef
}

// tagsByLevelType returns the tag values that correspond to a certain hierarchy
// level (lvl). I.e. if the hierarchy level is "genre", the values of tag
// "genre" are returned
func (me *track) tagsByLevelType(lvl config.LevelType) []string {
	switch lvl {
	case config.LvlGenre:
		return me.tags.genres
	case config.LvlAlbumArtist:
		return me.tags.albumArtists
	case config.LvlArtist:
		return me.tags.artists
	}
	return []string{}
}

// tracks maps track paths to the corresponding track instance
type tracks map[string]*track

// add adds a track object to tracks
func (me tracks) add(t *track) { me[t.path] = t }

// trackRef represents a reference to a track object. One trackRef instance is
// created for each hierarchy a music track is part of. I.e. for each music
// track multiple trackRef instances can exist.
type trackRef struct {
	*itm
	track *track
}
