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

// file info of a track
type trackpath struct {
	path       string
	info       func() os.FileInfo
	lastChange func() int64 // time of last change in UNIX format
}

// newTrackPath creates an instance of trackpath
func newTrackpath(path string, lastChange int64) (tp trackpath) {
	tp = trackpath{path: path}

	var info os.FileInfo
	tp.info = func() os.FileInfo {
		if info == nil {
			var err error
			info, err = os.Stat(tp.path)
			if err != nil {
				err = errors.Wrapf(err, "cannot create trackpath for '%s': %v", tp.path, err)
				log.Fatal(err)
				return nil
			}
		}
		return info
	}

	tp.lastChange = func() int64 {
		if lastChange != 0 {
			return lastChange
		}
		lastChange = tp.info().ModTime().Unix()
		return lastChange
	}

	return
}

// mimeType returns the mime type of a track
func (me trackpath) mimeType() string {
	return mime.TypeByExtension(path.Ext(me.path))
}

// size return the size of a track file in
func (me trackpath) size() (size int64, err error) {
	return me.info().Size(), nil
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

// metadata reads the ID3 tags and the picture for a track
func (me trackpath) metadata(sep string) (tgs *tags, pic *tag.Picture, err error) {

	f, err := os.Open(me.path)
	if err != nil {
		err = errors.Wrapf(err, "cannot retrieve meta data for '%s'", me.path)
		return
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		err = errors.Wrapf(err, "cannot retrieve meta data for '%s'", me.path)
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

type trackpaths []trackpath

// implementation of sort interface for trackpaths
func (me trackpaths) Len() int           { return len(me) }
func (me trackpaths) Less(i, j int) bool { return me[i].path < me[j].path }
func (me trackpaths) Swap(i, j int)      { me[i], me[j] = me[j], me[i] }

// removeDuplicates remove double entries (wrt. path) from an array of
// trackpaths. The implementation is inspired by
// https://learngolang.net/tutorials/how-to-remove-duplicates-from-sorted-array-in-go/
func (me *trackpaths) removeDuplicates() {
	n := len(*me)
	if n <= 1 {
		return
	}
	j := 1
	for i := 1; i < n; i++ {
		if (*me)[i].path != (*me)[i-1].path {
			(*me)[j] = (*me)[i]
			j++
		}
	}

	*me = (*me)[0:j]
}
