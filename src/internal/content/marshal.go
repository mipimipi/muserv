package content

// this files contains the logic to marshal content objects such as tracks
// albums etc. to DIDL-Lite format

import (
	"bytes"
	"fmt"
	"html"

	"gitlab.com/mipimipi/muserv/src/internal/config"
)

const (
	didlStartElem = "<DIDL-Lite xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:upnp=\"urn:schemas-upnp-org:metadata-1-0/upnp/\" xmlns=\"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/\" xmlns:dlna=\"urn:schemas-dlna-org:metadata-1-0/\">"
	didlEndElem   = "</DIDL-Lite>"
)

// indices takes the input attributes StartIndex (represented as start) and
// RequestedCount (represented as wanted) of the Browse action of the
// ContentDirectory service and calculates the first and the last index of the
// child objects of a container object. len is the total number of cildren.
func indices(start, wanted uint32, len int) (first, last int) {
	first = int(start)
	if wanted == 0 {
		last = len
	} else {
		last = int(start) + int(wanted)
		if last > len {
			last = len
		}
	}
	return
}

// marshalFuncMux returns a marshal function generator for container object ctr
// that represents a certain hierarchy level tag. I.e. if tag lvl "genre", ctr
// represents a genre container
func marshalFuncMux(lvl config.LevelType, ctr container) objMarshalFunc {
	switch lvl {
	case config.LvlAlbumArtist:
		return newAlbumArtistMarshalFunc(ctr)
	case config.LvlArtist:
		return newArtistMarshalFunc(ctr)
	case config.LvlGenre:
		return newGenreMarshalFunc(ctr)
	default:
		return newContainerMarshalFunc(ctr)
	}
}

// newAlbumMarshalFunc creates a new marshal function for an album. ctr is the
// album container object, intMusicPath is the file system path of the music
// library, extMusicPath is the external music URL (i.e. the virtual path where
// music tracks can be requestd via HTTP) and extPicturePath is the external
// picture URL (i.e. the virtual path where pictures can be requestd via HTTP).
func newAlbumMarshalFunc(ctr container, intMusicPath, extMusicPath, extPicturePath string) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		a := ctr.(*album)
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(a.name()))
		fmt.Fprint(buf, "<upnp:class>object.container.album.musicAlbum</upnp:class>")

		// add meta data
		var t *track
		for _, obj := range a.children.byID {
			t = obj.(*track)
			break
		}
		if t.picID.valid {
			fmt.Fprintf(buf, "<upnp:albumArtURI>%s</upnp:albumArtURI>", extPicturePath+fmt.Sprint(t.picID.id)+".jpg")
		}
		if a.year > 0 {
			fmt.Fprintf(buf, "<dc:date>%d-06-30</dc:date>", a.year)
		}
		for i := 0; i < len(a.artists); i++ {
			if len(a.artists[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:albumArtist>%s</upnp:albumArtist>", html.EscapeString(a.artists[i]))
			fmt.Fprintf(buf, "<upnp:artist role=\"albumArtist\">%s</upnp:artist>", html.EscapeString(a.artists[i]))
		}
		for i := 0; i < len(a.composers); i++ {
			if len(a.composers[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:artist role=\"Composer\">%s</upnp:artist>", html.EscapeString(a.composers[i]))
		}

		return buf.Bytes()
	}
}

// newAlbumRefMarshalFunc creates a new marshal function for the album
// reference container aRef
func newAlbumRefMarshalFunc(aRef container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)
		switch mode {
		case ModeMetadata:
			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", aRef.id(), aRef.parent().id(), aRef.numChildren())
			_, err := buf.Write(aRef.(albumRef).album.marshal(mode, 0, 0))
			if err != nil {
				log.Errorf("error marshalling album ref %d", aRef.id())
				return []byte{}
			}
			fmt.Fprint(buf, "</container>")

		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(aRef.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling album ref %d", aRef.id())
					return []byte{}
				}
			}
		}
		return buf.Bytes()
	}
}

// newAlbumArtistMarshalFunc creates a new marshal function for the album artist
// container albumArtist
func newAlbumArtistMarshalFunc(albumArtist container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)

		switch mode {
		case ModeMetadata:
			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", albumArtist.id(), albumArtist.parent().id(), albumArtist.numChildren())
			fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(albumArtist.name()))
			fmt.Fprintf(buf, "<upnp:class>object.container.person.musicArtist</upnp:class>")
			fmt.Fprintf(buf, "<upnp:artist role=\"albumArtist\">%s</upnp:artist>", html.EscapeString(albumArtist.name()))
			fmt.Fprintf(buf, "</container>")
		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(albumArtist.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling folder %d", albumArtist.id())
					return []byte{}
				}
			}
		}

		return buf.Bytes()
	}
}

// newArtistMarshalFunc creates a new marshal function for the artist
// container artist
func newArtistMarshalFunc(artist container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)

		switch mode {
		case ModeMetadata:
			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", artist.id(), artist.parent().id(), artist.numChildren())
			fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(artist.name()))
			fmt.Fprintf(buf, "<upnp:class>object.container.person.musicArtist</upnp:class>")
			fmt.Fprintf(buf, "<upnp:artist>%s</upnp:artist>", html.EscapeString(artist.name()))
			fmt.Fprintf(buf, "</container>")
		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(artist.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling folder %d", artist.id())
					return []byte{}
				}
			}
		}

		return buf.Bytes()
	}
}

// newFolderMarshalFunc creates a new marshal function for the folder
// container folder
func newFolderMarshalFunc(folder container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)

		switch mode {
		case ModeMetadata:
			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", folder.id(), folder.parent().id(), folder.numChildren())
			fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(folder.name()))
			fmt.Fprintf(buf, "<upnp:class>object.container.storageFolder</upnp:class>")
			fmt.Fprintf(buf, "</container>")
		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(folder.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling folder %d", folder.id())
					return []byte{}
				}
			}
		}

		return buf.Bytes()
	}
}

// newArtistMarshalFunc creates a new marshal function for the genre container
// genre
func newGenreMarshalFunc(genre container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)

		switch mode {
		case ModeMetadata:
			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", genre.id(), genre.parent().id(), genre.numChildren())
			fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(genre.name()))
			fmt.Fprintf(buf, "<upnp:class>object.container.genre.musicGenre</upnp:class>")
			fmt.Fprintf(buf, "<upnp:genre>%s</upnp:genre>", html.EscapeString(genre.name()))
			fmt.Fprintf(buf, "</container>")
		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(genre.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling folder %d", genre.id())
					return []byte{}
				}
			}
		}

		return buf.Bytes()
	}
}

// newContainerMarshalFunc creates a new marshal function for generic container
// ctr
func newContainerMarshalFunc(ctr container) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)

		switch mode {
		case ModeMetadata:
			var parentID ObjID
			if ctr.parent() == nil {
				parentID = ObjID(-1)
			} else {
				parentID = ctr.parent().id()
			}

			fmt.Fprintf(buf, "<container id=\"%d\" parentID=\"%d\" restricted=\"1\" searchable=\"0\" childCount=\"%d\">", ctr.id(), parentID, ctr.numChildren())
			fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(ctr.name()))
			fmt.Fprintf(buf, "<upnp:class>object.container</upnp:class>")
			fmt.Fprintf(buf, "</container>")
		case ModeChildren:
			for i := first; i < last; i++ {
				_, err := buf.Write(ctr.childByIndex(i).marshal(ModeMetadata, 0, 0))
				if err != nil {
					log.Errorf("error marshalling object %d", ctr.id())
					return []byte{}
				}
			}
		}

		return buf.Bytes()
	}
}

// newTrackMarshalFunc creates a new marshal function for a track. itm is the
// track item object, intMusicPath is the file system path of the music
// library, extMusicPath is the external music URL (i.e. the virtual path where
// music tracks can be requestd via HTTP) and extPicturePath is the external
// picture URL (i.e. the virtual path where pictures can be requestd via HTTP).
func newTrackMarshalFunc(itm item, intMusicPath, extMusicPath, extPicturePath string) objMarshalFunc {
	t := itm.(*track)
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)
		tags := t.tags
		fmt.Fprintf(buf, "<dc:title>%s</dc:title>", html.EscapeString(tags.title))
		fmt.Fprint(buf, "<upnp:class>object.item.audioItem.musicTrack</upnp:class>")

		// add meta data
		if tags.year > 0 {
			fmt.Fprintf(buf, "<dc:date>%d-06-30</dc:date>", tags.year)
		}
		for i := 0; i < len(tags.artists); i++ {
			if len(tags.artists[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:artist>%s</upnp:artist>", html.EscapeString(tags.artists[i]))
		}
		for i := 0; i < len(tags.albumArtists); i++ {
			if len(tags.albumArtists[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:artist role=\"albumArtist\">%s</upnp:artist>", html.EscapeString(tags.albumArtists[i]))
			fmt.Fprintf(buf, "<upnp:albumArtist>%s</upnp:albumArtist>", html.EscapeString(tags.albumArtists[i]))
		}
		for i := 0; i < len(tags.composers); i++ {
			if len(tags.composers[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:artist role=\"Composer\">%s</upnp:artist>", html.EscapeString(tags.composers[i]))
		}
		for i := 0; i < len(tags.genres); i++ {
			if len(tags.genres[i]) == 0 {
				continue
			}
			fmt.Fprintf(buf, "<upnp:genre>%s</upnp:genre>", html.EscapeString(tags.genres[i]))
		}
		if len(tags.album) > 0 {
			fmt.Fprintf(buf, "<upnp:album>%s</upnp:album>", html.EscapeString(tags.album))
		}
		if tags.trackNo > 0 {
			fmt.Fprintf(buf, "<upnp:originalTrackNumber>%d</upnp:originalTrackNumber>", tags.trackNo)
		}
		if t.picID.valid {
			fmt.Fprintf(buf, "<upnp:albumArtURI>%s</upnp:albumArtURI>", extPicturePath+fmt.Sprint(t.picID.id)+".jpg")
		}
		fmt.Fprintf(buf, "<res protocolInfo=\"http-get:*:%s:*\" size=\"%d\">", html.EscapeString(t.mimeType), t.size)
		fmt.Fprint(buf, html.EscapeString(extMusicPath+t.path[len(intMusicPath)+1:]))
		fmt.Fprint(buf, "</res>")

		return buf.Bytes()
	}
}

// newTrackRefMarshalFunc creates a new marshal function for track reference
// container tRef
func newTrackRefMarshalFunc(tRef item) objMarshalFunc {
	return func(mode string, first, last int) []byte {
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "<item id=\"%d\" refID=\"%d\" parentID=\"%d\" restricted=\"1\">", tRef.id(), tRef.(trackRef).track.id(), tRef.parent().id())
		_, err := buf.Write(tRef.(trackRef).track.marshalFunc(ModeMetadata, 0, 0))
		if err != nil {
			log.Errorf("error marshalling track ref %d", tRef.id())
			return []byte{}
		}
		fmt.Fprint(buf, "</item>")

		return buf.Bytes()
	}
}
