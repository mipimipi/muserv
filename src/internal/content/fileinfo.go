package content

import (
	"fmt"

	"mime"
	"os"
	"path"
	"strings"

	"github.com/mipimipi/tag"
	"github.com/pkg/errors"
)

// tags of a music file / track file
type tags struct {
	title        string
	album        string
	artists      []string
	albumArtists []string
	composers    []string
	genres       []string
	year         int
	trackNo      int
	tracksTotal  int
	discNo       int
	discsTotal   int
	compilation  bool
}

type infoKind int

const (
	infoNone infoKind = iota
	infoPlaylist
	infoTrack
)

type fileInfo interface {
	kind() infoKind
	path() string
	lastChange() int64
	mimeType() string // time of last change in UNIX format
	size() int64
}

type baseInfo struct {
	p    string // path
	info func() os.FileInfo
	lChg func() int64 // time of last change in UNIX format
}

// newBaseInfo creates an instance of baseInfo
func newBaseInfo(path string, lastChange int64) (bi baseInfo) {
	bi = baseInfo{p: path}

	var info os.FileInfo
	bi.info = func() os.FileInfo {
		if info == nil {
			var err error
			info, err = os.Stat(bi.p)
			if err != nil {
				err = errors.Wrapf(err, "cannot create baseInfo for '%s': %v", bi.path(), err)
				log.Fatal(err)
				return nil
			}
		}
		return info
	}

	bi.lChg = func() int64 {
		if lastChange != 0 {
			return lastChange
		}
		lastChange = bi.info().ModTime().Unix()
		return lastChange
	}

	return
}

func (me baseInfo) kind() infoKind    { return infoNone }
func (me baseInfo) path() string      { return me.p }
func (me baseInfo) lastChange() int64 { return me.lChg() }
func (me baseInfo) mimeType() string  { return mime.TypeByExtension(path.Ext(me.path())) }
func (me baseInfo) size() int64       { return me.info().Size() }

type playlistInfo struct {
	baseInfo
}

// newPlaylistInfo creates an instance of playlistInfo
func newPlaylistInfo(path string, lastChange int64) trackInfo {
	return trackInfo{newBaseInfo(path, lastChange)}
}

func (me playlistInfo) kind() infoKind { return infoPlaylist }

type trackInfo struct {
	baseInfo
}

// newTrackInfo creates an instance of trackInfo
func newTrackInfo(path string, lastChange int64) trackInfo {
	return trackInfo{newBaseInfo(path, lastChange)}
}

func (me trackInfo) kind() infoKind { return infoTrack }

// metadata reads the ID3 tags and the picture for a track
func (me trackInfo) metadata(sep string) (tgs *tags, pic *tag.Picture, err error) {

	f, err := os.Open(me.path())
	if err != nil {
		err = errors.Wrapf(err, "cannot retrieve meta data for '%s'", me.path())
		return
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		err = errors.Wrapf(err, "cannot retrieve meta data for '%s'", me.path())
		return
	}

	// process tags
	tgs = new(tags)
	tgs.title = m.Title()
	tgs.trackNo, tgs.tracksTotal = m.Track()
	tgs.discNo, tgs.discsTotal = m.Disc()
	tgs.album = m.Album()
	tgs.composers = splitMultipleEntries(m.Composer(), sep)
	tgs.genres = splitMultipleEntries(m.Genre(), sep)
	tgs.year = m.Year()
	// - compilation
	i, ok := m.Raw()["compilation"]
	var s string
	if !ok {
		i, ok = m.Raw()["Compilation"]
		if ok {
			s = fmt.Sprintf("%v", i)
		}
	} else {
		s = fmt.Sprintf("%v", i)
	}
	tgs.compilation = (s == "1")
	// - (album) artists
	tgs.artists = splitMultipleEntries(m.Artist(), sep)
	tgs.albumArtists = splitMultipleEntries(m.AlbumArtist(), sep)
	if !tgs.compilation && len(tgs.albumArtists) == 0 {
		tgs.albumArtists = tgs.artists
	}

	pic = m.Picture()

	return
}

type fileInfos []fileInfo

// implementation of sort interface for trackpaths
func (me fileInfos) Len() int           { return len(me) }
func (me fileInfos) Less(i, j int) bool { return me[i].path() < me[j].path() }
func (me fileInfos) Swap(i, j int)      { me[i], me[j] = me[j], me[i] }

// removeDuplicates remove double entries (wrt. path) from an array of
// fileInfos. The implementation is inspired by
// https://learngolang.net/tutorials/how-to-remove-duplicates-from-sorted-array-in-go/
func (me *fileInfos) removeDuplicates() {
	n := len(*me)
	if n <= 1 {
		return
	}
	j := 1
	for i := 1; i < n; i++ {
		if (*me)[i].path() != (*me)[i-1].path() {
			(*me)[j] = (*me)[i]
			j++
		}
	}

	*me = (*me)[0:j]
}

// splitMultipleEntries splits a tag that contains multiple entries which are
// separated by sep into these entries. Each entry is trimmed wrt. left and
// right spaces
func splitMultipleEntries(tag, sep string) (meta []string) {
	for _, s := range strings.Split(tag, sep) {
		meta = append(meta, strings.TrimSpace(s))
	}
	return
}
