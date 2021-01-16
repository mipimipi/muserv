package content

import (
	"fmt"
	"reflect"

	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// album represents an album object. For each music album, exactly one album
// object exists
type album struct {
	*ctr
	year        int
	compilation bool
	artists     []string    // album artists
	composers   []string    // album composers
	lastChange  int64       // UNIX time of last change of track file
	refs        []*albumRef // corresponding album references
}

// newAlbum creates a new album object
func newAlbum(cnt *Content, key uint64) (a *album) {
	a = &album{
		newCtr(cnt, cnt.newID(), ""),
		0,
		false,
		[]string{},
		[]string{},
		0,
		[]*albumRef{},
	}
	a.k = key
	a.marshalFunc = newAlbumMarshalFunc(a, cnt.extPicturePath)

	cnt.objects.add(a)
	cnt.albums.add(a)

	return
}

// addChild adds a track as child and adjusts lastChange. If necessary, the
// sorting of corresponding albumRefs is invalidated
func (me *album) addChild(obj object) {
	// only tracks can be added as children to album
	if reflect.TypeOf(obj) != reflect.TypeOf((*track)(nil)) {
		log.Warnf("tried of add an object of type '%s' to album", reflect.TypeOf(obj).String())
		return
	}

	me.children.add(obj)
	obj.setParent(me)
	me.cnt.traceUpdate(me.i)

	// if lastChange was adjusted, propagate the change to all albumRefs
	t := obj.(*track)
	if t.lastChange > me.lastChange {
		me.lastChange = t.lastChange
		for _, aRef := range me.refs {
			aRef.parent().invalidateOrder()
		}
	}
}

// delChild removes a track (only tracks can be children of albums) and adjusts
// lastChange. If necessary, the sorting of corresponding albumRefs is
// invalidated
func (me *album) delChild(obj object) {
	me.children.del(obj)
	obj.setParent(nil)
	me.cnt.traceUpdate(me.i)

	// adjust lastChange, propagate the change to all albumRefs if necessary
	t := obj.(*track)
	if t.lastChange == me.lastChange {
		me.lastChange = 0
		for i := 0; i < me.numChildren(); i++ {
			t := me.childByIndex(i).(*track)
			if t.lastChange > me.lastChange {
				me.lastChange = t.lastChange
			}
		}
		for _, aRef := range me.refs {
			aRef.parent().invalidateOrder()
		}
	}
}

// newAlbumRef creates a new album reference object from an album
func (me *album) newAlbumRef(sfs []config.SortField) *albumRef {
	aRef := albumRef{
		newCtr(me.cnt, me.cnt.newID(), me.n),
		me,
	}
	aRef.marshalFunc = newAlbumRefMarshalFunc(aRef)
	aRef.k = me.k

	me.refs = append(me.refs, &aRef)

	// set sort fields of album reference
	if len(sfs) > 0 {
		aRef.sf = []string{}
		for _, sf := range sfs {
			var s string
			switch sf {
			case config.SortLastChange:
				s = fmt.Sprintf("%020d", me.lastChange)
			case config.SortTitle:
				s = me.n
			case config.SortYear:
				s = fmt.Sprintf("%d", me.year)
			}
			if len(s) > 0 {
				aRef.sf = append(aRef.sf, s)
			}
		}
	}

	me.cnt.objects.add(aRef)

	return &aRef
}

// albums maps album keys to the corresponding album instance. An album key is
// the uint64 FNV hash of the string concatenation of album attributes name,
// year and compilation.
type albums map[uint64]*album

// add adds an album to albums
func (me albums) add(a *album) { me[a.key()] = a }

// albumRef represents a reference to an album object. One albumRef instance is
// created for each hierarchy a music album is part of. I.e. for each music
// album multiple albumRef instances can exist.
type albumRef struct {
	*ctr
	album *album
}
