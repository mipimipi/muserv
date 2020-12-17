package content

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dhowden/tag"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	utils "gitlab.com/mipimipi/go-utils"
)

// size of images in pixel (i.e. each image is not larger than 300px x 300px)
const imgSize = 300

// ObjID is the unique identified of an object
type ObjID int64

// ObjIDFromString create an object ID from a string
func ObjIDFromString(s string) (ObjID, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		err = errors.Wrapf(err, "could not create object ID from '%s'", s)
		log.Error(err)
		return 0, err
	}
	return ObjID(id), nil
}

// objMarshalFunc is the type of the marshal function type of an object
type objMarshalFunc func(string, int, int) []byte

// object is an abstraction of a content object according to the
// ContentDirectory service specification
type object interface {
	id() ObjID
	key() uint64
	name() string
	setParent(container)
	parent() container
	marshal(string, int, int) []byte
	sortField() string
	isContainer() bool
	isItem() bool
}

// item is an abstraction of an item object according to the ContentDirectory
// service specification
type item interface {
	object
}

// container is an abstraction of a container object according to the
// ContentDirectory service specification
type container interface {
	object
	addChild(object)
	delChild(object)
	numChildren() int
	childByIndex(int) object
	childByKey(uint64) (object, bool)
	resetUpdCount()
}

// objects maps object IDs to the corresponding object instance
type objects map[ObjID]object

// add adds an object instance to objects
func (me objects) add(obj object) {
	me[obj.id()] = obj
}

type sortOrder int

const (
	sortAsc sortOrder = iota
	sortDesc
)

// refs implements the references to child objects of a container object. Child
// objects can be accessed via different ways:
// - by object ID
// - by object key (a key is a uint64 hash that typically encodes some object
//	 attributes)
// - by an index in sort order
// After a child object was added or removed, the inOrder array is generated
// from byID upon the access to it
type refs struct {
	byID    objects
	byKey   map[uint64]object
	inOrder []object
	ord     sortOrder
}

// newRefs create a new refs instance
func newRefs(ord sortOrder) (r refs) {
	return refs{
		byID:  make(objects),
		byKey: make(map[uint64]object),
		ord:   ord,
	}
}

// add adds a child object. The inOrder array is cleared.
func (me *refs) add(obj object) {
	me.byID[obj.id()] = obj
	me.byKey[obj.key()] = obj
	if len(me.inOrder) > 0 {
		me.inOrder = []object{}
	}
}

// del removes a child object. The inOrder array is cleared.
func (me *refs) del(obj object) {
	delete(me.byID, obj.id())
	delete(me.byKey, obj.key())
	me.inOrder = []object{}
}

// item returns child object number index according to the sort order. If the
// inOrder array is empty, it is first generated
func (me *refs) item(index int) object {
	// if sorted array does not exist: create it
	if len(me.inOrder) == 0 {
		for _, obj := range me.byID {
			me.inOrder = append(me.inOrder, obj)
		}
		var less func(i, j int) bool
		if me.ord == sortAsc {
			less = func(i, j int) bool { return me.inOrder[i].sortField() < me.inOrder[j].sortField() }
		} else {
			less = func(i, j int) bool { return me.inOrder[i].sortField() > me.inOrder[j].sortField() }
		}
		sort.Slice(me.inOrder, less)
	}

	return me.inOrder[index]
}

// len returns the number of child objects
func (me *refs) len() int { return len(me.byID) }

// obj represents a generic content object according to the ContentDirectory
// service specification
type obj struct {
	cnt         *Content
	i           ObjID          // object ID
	n           string         // object name
	k           uint64         // object key (tyically a hash)
	sf          string         // sort field (the field that is used for sorting an object array)
	p           container      // parent object (nil if object has no parent)
	marshalFunc objMarshalFunc // function to marshal object (i.e. create a representation in DIDL)
}

// newObj creates a new object instance
func newObj(cnt *Content, id ObjID, name string) *obj {
	obj := obj{
		cnt:         cnt,
		i:           id,
		k:           utils.HashUint64(name),
		n:           name,
		sf:          strings.ToLower(name),
		marshalFunc: func(mode string, first int, last int) []byte { return []byte{} },
	}
	return &obj
}

func (me *obj) id() ObjID               { return me.i }
func (me *obj) key() uint64             { return me.k }
func (me *obj) name() string            { return me.n }
func (me *obj) setParent(ctr container) { me.p = ctr }
func (me *obj) parent() container       { return me.p }
func (me *obj) sortField() string       { return me.sf }
func (me *obj) marshal(mode string, first, last int) []byte {
	return me.marshalFunc(mode, first, last)
}
func (me *obj) isContainer() bool {
	return false
}
func (me *obj) isItem() bool {
	return false
}

// ctr represents a generic container object according to the ContentDirectory
// service specification
type ctr struct {
	*obj
	updCount uint32 // ContainerUpdateIDValue
	children refs   // child objects
}

// newCtr creates a new instance of ctr
func newCtr(cnt *Content, id ObjID, name string) *ctr {
	ctr := ctr{
		newObj(cnt, id, name),
		0,
		newRefs(sortAsc),
	}
	ctr.marshalFunc = newContainerMarshalFunc(&ctr)
	return &ctr
}

// addChild adds an object as children, sets the parent of that object and
// registers the change to be able to set the state variable ContainerUpdateIDs
// accordingly
func (me *ctr) addChild(obj object) {
	me.children.add(obj)
	obj.setParent(me)
	me.cnt.traceUpdate(me.i)
}

// delChild removes an object as children, clears the parent of that object and
// registers the change to be able to set the state variable ContainerUpdateIDs
// accordingly
func (me *ctr) delChild(obj object) {
	me.children.del(obj)
	obj.setParent(nil)
	me.cnt.traceUpdate(me.i)
}

func (me *ctr) numChildren() int              { return me.children.len() }
func (me *ctr) childByIndex(index int) object { return me.children.item(index) }
func (me *ctr) childByKey(key uint64) (object, bool) {
	obj, exists := me.children.byKey[key]
	return obj, exists
}
func (me *ctr) isContainer() bool {
	return true
}

// resetUpdCount recursively resets the ContainerUpdateIDValue
func (me *ctr) resetUpdCount() {
	me.updCount = 0
	for i := 0; i < me.numChildren(); i++ {
		if me.childByIndex(i).isContainer() {
			me.childByIndex(i).(container).resetUpdCount()
		}
	}
}

// itm represents a generic item object according to the ContentDirectory
// service specification
type itm struct {
	*obj
}

// newItm creates a new instance of itm
func newItm(cnt *Content, id ObjID, name string) *itm {
	itm := itm{
		newObj(cnt, id, name),
	}
	return &itm
}

func (me *itm) isItem() bool {
	return true
}

// album represents an album object. For each music album, exactly one album
// object exists
type album struct {
	*ctr
	year        int
	compilation bool
	artists     []string // album artists
	composers   []string // album composers
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

// folder represents a folder object
type folder struct {
	*ctr
	path string // folder path
}

// folders maps folder paths to the corresponding folder instance
type folders map[string]folder

// add adds a folder to folders
func (me folders) add(path string, folder folder) { me[path] = folder }

// pictures maps a picture id (that's an uint64 FNV hash of the picture raw
// data) to the picture raw data
type pictures struct {
	mu   sync.Mutex           // required for concurrent-safe write access
	data map[uint64](*[]byte) // the actual map (id->raw data)
}

// get picture raw data by id
func (me *pictures) get(id uint64) *[]byte {
	return me.data[id]
}

// add adds pictures to the pictures map. It take a picture from the tags of a
// music file, resizes is and converts it to JPEG. It creates a picture id as
// uint64 FNV hash of the raw data and adds it to the pictures map.
// This function is designed to be executed concurrently.
func (me *pictures) add(wg *sync.WaitGroup, pic *tag.Picture, picID *nonePicID) {
	defer wg.Done()

	if pic == nil {
		return
	}

	//  resize picture
	img, err := imaging.Decode(bytes.NewReader(pic.Data))
	if err != nil {
		err = errors.New("could not decode picture")
		log.Fatal(err)
		return
	}
	img = imaging.Resize(img, imgSize, 0, imaging.Box)
	buf := new(bytes.Buffer)
	if err = imaging.Encode(
		buf,
		imaging.Resize(img, imgSize, 0, imaging.Box),
		imaging.JPEG,
	); err != nil {
		err = errors.New("could not encode resized picture")
		log.Fatal(err)
		return
	}
	picture := buf.Bytes()

	*picID = nonePicID{utils.HashUint64("%x", picture), true}

	me.mu.Lock()
	_, exists := me.data[picID.id]
	if !exists {
		me.data[picID.id] = &picture
	}
	me.mu.Unlock()
}

// nonePicID represents a picture ID incl. a "null" value
type nonePicID struct {
	id    uint64
	valid bool
}

// track represents a track object. For each music track, exactly one track
// object exists
type track struct {
	*itm
	tags        *tags       // tags of the track
	picID       nonePicID   // ID of the cover picture (can be "null")
	mimeType    string      // mime type of track file
	size        int64       // size of track file in bytes
	lastChanged int64       // UNIX time of last change of track file
	path        string      // path of track file
	refs        []*trackRef // corresponding track references
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

// sortFieldFromTrackNo creates a string sort field from the track number. To
// avoid a sorting such as "1", "10", "2", 20", ... leading zeros are added.
// Thus the sequence will be "0001", "0002", ..., "0010", ..., "0020", ...
func (me *trackRef) sortFieldFromTrackNo(trackNo int) {
	me.sf = fmt.Sprintf("%04d", trackNo)
}
