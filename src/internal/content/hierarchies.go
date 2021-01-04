package content

import (
	"path/filepath"

	"github.com/pkg/errors"
	utils "gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// addTrackToHierarchy adds track t to the hierarchy defined by hier below the
// hierarchy root ctr. count is increased by the number of object changes that
// happened during this activity
func (me *Content) addTrackToHierarchy(count *uint32, hier *config.Hierarchy, ctr container, t *track) (err error) {
	return me.addTrackToHierarchyLevel(count, hier, 0, ctr, t)
}

// addToTrackHierarchyLevel adds track t to the hierarchy defined by hier as
// level with the given index as children under ctr.
// addToHierarchyLevel itself adds the "upper nodes" (i.e. everything - genre,
// album artist etc. - above the album and track level of the hierarchy).
// The function calls itself recursively until a "lower node" (track and - if
// required - album) is reached. To add these nodes, addToSubHierarchy is
// called. count is increased by the number of object changes that happened
// during this activity
func (me *Content) addTrackToHierarchyLevel(count *uint32, hier *config.Hierarchy, index int, ctr container, t *track) (err error) {
	if hier.Levels[index].Type == config.LvlAlbum || hier.Levels[index].Type == config.LvlTrack {
		if err = me.addTrackToSubHierarchy(count, hier, index, ctr, t); err != nil {
			return
		}
		return
	}

	tags := t.tagsByLevelType(hier.Levels[index].Type)

	for i := 0; i < len(tags); i++ {
		var ctrNext container
		obj, exists := ctr.childByKey(utils.HashUint64("%s", tags[i]))
		if exists {
			ctrNext = obj.(container)
		} else {
			ctrNew := newCtr(me, me.newID(), tags[i])
			ctrNew.marshalFunc = marshalFuncMux(hier.Levels[index].Type, ctrNew)
			me.objects.add(ctrNew)
			ctr.addChild(ctrNew)
			// count creation of new object
			*count++
			// set comparison functions for sorting of the child objects. Here
			// we know that index is not the last level of the hierarchy. Thus
			// level index+1 exists as well.
			ctrNew.setComparison(hier.Levels[index+1].Comparisons())

			ctrNext = ctrNew
		}

		if err = me.addTrackToHierarchyLevel(count, hier, index+1, ctrNext, t); err != nil {
			return
		}
	}

	return
}

// addTrackToSubHierarchy adds track t to the hierarchy defined by hier as level with
// the given index as children under ctr. This function only takes care of the
// "lower node" (track and - if required - album). I.e. if required, also the
// album level is created - after that, the hierarchy is like: ... <- ctr
// [<- albumRef] <- trackRef. count is increased by the number of object changes
// that happened during this activity
func (me *Content) addTrackToSubHierarchy(count *uint32, hier *config.Hierarchy, index int, ctr container, t *track) (err error) {
	// create track reference
	tRef := me.trackRefFromTrack(t, hier.Levels[len(hier.Levels)-1].SortFields())
	// count creation of trackRef object
	*count++

	// check if album level must be created
	if hier.Levels[index].Type != config.LvlAlbum {
		// no album level must be created: add track reference to object tree
		// and return
		ctr.addChild(tRef)
		// count change of container
		*count++
		return
	}

	// album level must be created. Check if a corresponding album reference already exists
	var aRef *albumRef
	obj, exists := ctr.childByKey(t.albumKey())
	if exists {
		aRef = obj.(*albumRef)
	} else {
		// determine album and create a new album reference from it
		a, exists := me.albums[t.albumKey()]
		if !exists {
			err = errors.Wrapf(err, "cannot add album %s, %d, %t to sub hierarchy since it does not exist", t.tags.album, t.tags.year, t.tags.compilation)
			log.Fatal(err)
			return
		}
		aRef = me.albumRefFromAlbum(a, hier.Levels[index].SortFields())
		// set comparison functions for sosrting of child objects
		aRef.setComparison(hier.Levels[index+1].Comparisons())
		// add album reference to object tree
		ctr.addChild(aRef)
		// count change of container
		*count++
	}

	// add track reference to object tree
	aRef.addChild(tRef)
	// count change of album reference object
	*count++

	return
}

// addTrackToFolderHierarchy adds track t to the folder hierarchy. ctr is the
// corresponding hierarchy object (i.e. one level below root). count is
// increased by the number of object changes that happened during this activity
func (me *Content) addTrackToFolderHierarchy(count *uint32, ctr container, t *track) {
	// create track reference. In the folder hierarchy, tracks are ordered by
	// file name
	tRef := me.trackRefFromTrack(t, []config.SortField{})
	tRef.sf = []string{filepath.Base(t.path)}
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
		}
		f.addChild(child)
		// count change of folder object
		*count++
		child = f
	}
	if !exists {
		ctr.addChild(child)
	}
}
