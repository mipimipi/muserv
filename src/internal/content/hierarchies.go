package content

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// addToHierarchy adds track t to the hierarchy with the index index. count is
// increased by the number of object changes that happened during this activity
func (me *Content) addToHierarchy(count *uint32, index int, t *track) (err error) {
	switch me.cfg.UPnP.Hiers[index].ID {
	case hierFolder:
		me.addToFolderHierarchy(count, index, t)
	case hierLatestAlbums:
		err = me.addToLatestAlbumsHierarchy(count, index, t)
	case hierLatestTracks:
		me.addToLatestTracksHierarchy(count, index, t)
	default:
		err = me.addToAnyHierarchy(count, index, t)
	}
	return
}

// addToAnyHierarchy adds track t to any hierarchy that is neither the folder
// nor the the latest-album- nor the latest-tracks-hierarchy. index
// is the index of that hierarchy. addToHierarchy adds the "upper nodes" (i.e.
// everything - genre, album artist etc. - above the album and track level of
// the hierarchy). To add the "lower nodes" (track and - if required - album)
// addToSubHierarchy is called. count is increased by the number of object
// changes that happened during this activity
func (me *Content) addToAnyHierarchy(count *uint32, index int, t *track) (err error) {
	l := len(me.cfg.UPnP.Hiers[index].Levels)

	if l == 0 {
		return me.addToSubHierarchy(count, t, me.hierarchies[index], index, 0)
	}
	if l == 1 && me.cfg.UPnP.Hiers[index].Levels[0] == config.TagAlbum {
		return me.addToSubHierarchy(count, t, me.hierarchies[index], index, -1)
	}

	var (
		arr0, arr1 []string
		tag0, tag1 string
	)

	tags := func(tagName string) []string {
		switch tagName {
		case config.TagGenre:
			return t.tags.genres
		case config.TagAlbumArtist:
			return t.tags.albumArtists
		case config.TagArtist:
			return t.tags.artists
		}
		return []string{}
	}

	if me.cfg.UPnP.Hiers[index].Levels[0] != config.TagAlbum {
		tag0 = me.cfg.UPnP.Hiers[index].Levels[0]
		arr0 = tags(tag0)
		if l > 1 && me.cfg.UPnP.Hiers[index].Levels[1] != config.TagAlbum {
			tag1 = me.cfg.UPnP.Hiers[index].Levels[1]
			arr1 = tags(tag1)
		}
	}

	if len(tag0) > 0 {
		for i := 0; i < len(arr0); i++ {
			h := utils.HashUint64("%s", arr0[i])
			var ctr0 container
			obj, exists := me.hierarchies[index].childByKey(h)
			if exists {
				ctr0 = obj.(container)
			} else {
				ctr := newCtr(me, me.newID(), arr0[i])
				ctr.marshalFunc = marshalFuncMux(tag0, ctr)
				me.objects.add(ctr)
				me.hierarchies[index].addChild(ctr)
				ctr0 = ctr
				// count creation of new object
				*count++
			}

			if len(tag1) == 0 {
				if err = me.addToSubHierarchy(count, t, ctr0, index, 0); err != nil {
					return
				}
			}

			for j := 0; j < len(arr1); j++ {
				h := utils.HashUint64("%s", arr1[j])
				var ctr1 container
				obj, exists := ctr0.childByKey(h)
				if exists {
					ctr1 = obj.(container)
				} else {
					ctr := newCtr(me, me.newID(), arr1[j])
					ctr.marshalFunc = marshalFuncMux(tag1, ctr)
					me.objects.add(ctr)
					ctr0.addChild(ctr)
					ctr1 = ctr
					// count creation of new object
					*count++
				}

				if err = me.addToSubHierarchy(count, t, ctr1, index, 1); err != nil {
					return
				}
			}
		}
	}

	return
}

// addToSubHierarchy adds a trackRef object and (if required) above that an
// albumRef object as child object to container ctr as part of the hierarchy
// with the given index. level is the level of ctr in that hierarchy. I.e.
// after that, the hierarchy is is like: ... <- ctr [<- albumRef] <- trackRef.
// count is increased by the number of object changes that happened during this
// activity
func (me *Content) addToSubHierarchy(count *uint32, t *track, ctr container, index, level int) (err error) {
	// create track reference. Depending on the next upper level, the sort
	// field will be set later
	tRef := me.trackRefFromTrack(t)
	// count creation of trackRef object
	*count++

	if level != -1 && level == len(me.cfg.UPnP.Hiers[index].Levels)-1 {
		// add track reference to hierarchy
		ctr.addChild(tRef)
		return
	}
	// in this case the hierarchy level must be album: get album
	a, exists := me.albums[utils.HashUint64("%s%d%t", t.tags.album, t.tags.year, t.tags.compilation)]
	if !exists {
		a, err = me.albumFromTrack(t)
		if err != nil {
			err = errors.Wrapf(err, "cannot add album %s, %d, %t to sub hierarchy", t.tags.album, t.tags.year, t.tags.compilation)
			log.Error(err)
			return
		}
		// count creation of album object
		*count++
	}

	// get album reference. If the next upper level is albumartist, the sort
	// field is year, else the sort field is album name (which is the
	// default)
	var aRef *albumRef
	o, exists := ctr.childByKey(a.key())
	if exists {
		aRef = o.(*albumRef)
	} else {
		aRef = me.albumRefFromAlbum(a)
		ctr.addChild(aRef)
		// count creation of albumRef object
		*count++
	}
	if level > 0 && me.cfg.UPnP.Hiers[index].Levels[level] == config.TagAlbumArtist {
		aRef.sf = fmt.Sprint(a.year)
	}

	// set sort field for tracks to track_number and add track reference to
	// hierarchy. The sort field is formatted with leading zeros to ensure a
	// correct sorting, otherwise the sort sequence would be "1", "10", "2" ...
	tRef.sortFieldFromTrackNo(t.tags.trackNo)
	aRef.addChild(tRef)

	// add album reference to hierarchy
	ctr.addChild(aRef)

	return
}

// addToFolderHierarchy adds track t to the folder hierarchy. index is the
// index of that hierarchy. count is increased by the number of object changes
// that happened during this activity
func (me *Content) addToFolderHierarchy(count *uint32, index int, t *track) {
	// create track reference. In the folder hierarchy, tracks are ordered by
	// file name
	tRef := me.trackRefFromTrack(t)
	tRef.sf = filepath.Base(t.path)
	// count creation of trackRef object
	*count++

	var (
		child  object = tRef
		exists bool
	)
	for path := filepath.Dir(tRef.track.path); path != me.cfg.Cnt.MusicDir; path = filepath.Dir(path) {
		if len(path) == 0 {
			continue
		}

		// check if folder with path path already exists in hierarchy
		f, exists := me.folders[path]
		if !exists {
			fNew := folder{newCtr(me, me.newID(), filepath.Base(path)), path}
			fNew.marshalFunc = newFolderMarshalFunc(fNew)
			f = fNew
			me.folders.add(path, f)
			me.objects.add(f)
			// count folder creation
			*count++
		}
		f.addChild(child)
		child = f
	}
	if !exists {
		me.hierarchies[index].addChild(child)
	}
}

// addToLatestAlbumHierarchy adds track t to the latest-album-hierarchy. index
// is the index of that hierarchy. count is increased by the number of object
// changes that happened during this activity
func (me *Content) addToLatestAlbumsHierarchy(count *uint32, index int, t *track) (err error) {
	// nothing to do if album tag is not maintained in track
	if len(t.tags.album) == 0 {
		return
	}

	// create track reference. In the latest albums hierarchy, tracks are
	// ordered by track number. The sort field is formatted with leading zeros
	// to ensure a correct sorting, otherwise the sort sequence would be "1",
	// "10", "2" ...
	tRef := me.trackRefFromTrack(t)
	tRef.sortFieldFromTrackNo(t.tags.trackNo)
	// count creation of trackRef object
	*count++

	// get album
	a, exists := me.albums[utils.HashUint64("%s%d%t", t.tags.album, t.tags.year, t.tags.compilation)]
	if !exists {
		a, err = me.albumFromTrack(t)
		if err != nil {
			err = errors.Wrapf(err, "cannot add album %s, %d, %t to latest album hierarchy", t.tags.album, t.tags.year, t.tags.compilation)
			log.Error(err)
			return
		}
		// count creation of album object
		*count++
	}

	// get album reference. In the latest album hierarchy, albums are
	// ordered by last changed descending. Albums lattes change time is the
	// maximum of the last change times of its tracks
	var aRef *albumRef
	obj, exists := me.hierarchies[index].childByKey(a.key())
	if exists {
		aRef = obj.(*albumRef)
		if aRef.sf < fmt.Sprint(t.lastChanged) {
			aRef.sf = fmt.Sprint(t.lastChanged)
		}
	} else {
		aRef = me.albumRefFromAlbum(a)
		aRef.sf = fmt.Sprint(t.lastChanged)
		me.hierarchies[index].addChild(aRef)
		// count creation of albumRef object
		*count++
	}

	// add track references to hiearchy
	aRef.addChild(tRef)

	return
}

// addToLatestTracksHierarchy adds track t to the latest-tracks-hierarchy. index
// is the index of that hierarchy. count is increased by the number of object
// changes that happened during this activity
func (me *Content) addToLatestTracksHierarchy(count *uint32, index int, t *track) {
	// create track reference. In the latest tracks hierarchy, tracks are
	// ordered by last change date descending
	tRef := me.trackRefFromTrack(t)
	tRef.sf = fmt.Sprint(t.lastChanged)
	me.hierarchies[index].addChild(tRef)
	// count creation of trackRef object
	*count++
}
