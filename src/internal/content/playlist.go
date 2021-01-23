package content

import (
	"fmt"
	"net/url"
	"os"
	p "path"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/ushis/m3u"
	"gitlab.com/mipimipi/go-utils/file"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// playlist represents a playlist container
type playlist struct {
	*ctr
	lastChange int64 // UNIX time of last change of track file
}

// newPlaylist creates a new playlist container
func newPlaylist(cnt *Content, wg *sync.WaitGroup, count *uint32, pli playlistInfo) (pl *playlist, err error) {
	pl = &playlist{
		newCtr(cnt, cnt.newID(), p.Base(file.PathTrunk(pli.path()))),
		pli.lastChange(),
	}
	pl.marshalFunc = newPlaylistMarshalFunc(pl)

	cnt.playlists[pli.path()] = pl
	cnt.objects.add(pl)

	var f *os.File
	if f, err = os.Open(pli.path()); err != nil {
		err = errors.Wrapf(err, "cannot open playlist file '%s'", pli.path())
		log.Error(err)
		return
	}
	defer f.Close()

	var playlist m3u.Playlist
	if playlist, err = m3u.Parse(f); err != nil {
		err = errors.Wrapf(err, "cannot parse playlist '%s'", pli.path())
		log.Error(err)
		return
	}

	for i, item := range playlist {
		var (
			t    *track
			path string
			uri  *url.URL
		)

		// check and normalize path of playlist item (either it's an external
		// path with the scheme "http" or "https" or it must be a sub path of
		// the music directory - if both is not the case, the item is ignored.
		// If the path is local and relative, it's turned into an absolute path
		path = strings.TrimSpace(item.Path)
		if len(path) == 0 {
			continue
		}
		if !p.IsAbs(path) {
			uri, err = url.ParseRequestURI(path)
			if err != nil {
				dir, _ := p.Split(pli.path())
				path = p.Join(dir, path)
			} else {
				if uri.Scheme != "" && uri.Scheme != "http" && uri.Scheme != "https" {
					log.Errorf("playlist item '%s' has invalid scheme '%s': ignore it", path, uri.Scheme)
					continue
				}
				if uri.Scheme == "" && uri.Host != "" {
					log.Errorf("playlist item '%s' has empty scheme but host ist not empty: ignore it", path)
					continue
				}
			}
		}

		if t, err = trackFromPlaylistItem(cnt, wg, count, path, item.Title); err != nil {
			continue
		}

		// add reference to track as children to playlist container
		tRef := t.newTrackRef([]config.SortField{})
		tRef.sf = []string{fmt.Sprintf("%06d", i)}
		pl.addChild(tRef)
	}

	return
}

// trackFromPlaylistItem create a track object from a playlist item
func trackFromPlaylistItem(cnt *Content, wg *sync.WaitGroup, count *uint32, path, title string) (t *track, err error) {
	var exists bool

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// get corresponding track for playlist item. Create it if it doesn't
		// exist
		t, exists = cnt.tracks[path]
		if !exists {
			if t, err = newExtTrack(cnt, count, path, title); err != nil {
				err = errors.Wrapf(err, "cannot create a track for playlist item '%s': ignore it", path)
				log.Error(err)
				return
			}
			if len(title) == 0 {
				title = p.Base(file.PathTrunk(path))
			}
			t.n = title
			t.sf = []string{title}
		}

	} else {
		path = p.Clean(path)
		if dir := cnt.cfg.Cnt.MusicDir(path); dir == "" {
			err = fmt.Errorf("playlist item '%s' is not in music directory: ignore it", path)
			log.Error(err)
			return
		}
		if exists, err = file.Exists(path); err != nil {
			err = errors.Wrapf(err, "cannot check existence of playlist item '%s': ignore it", path)
			log.Error(err)
			return
		}
		if !exists {
			err = fmt.Errorf("playlist item '%s' doesn't exist: ignore it", path)
			log.Error(err)
			return
		}

		// get corresponding track for playlist item. Create it if it doesn't
		// exist
		t, exists = cnt.tracks[path]
		if !exists {
			if t, err = newTrack(cnt, wg, count, newTrackInfo(path, 0)); err != nil {
				err = errors.Wrapf(err, "cannot create a track for playlist item '%s': ignore it", path)
				log.Error(err)
				return
			}
		}
	}
	return
}

// playlists maps track paths to the corresponding track instance
type playlists map[string]*playlist
