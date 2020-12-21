package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"os/user"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.com/mipimipi/go-utils"
	"gitlab.com/mipimipi/go-utils/file"
)

// UserName is the name of the muserv system user
const UserName = "muserv"

// ValueKey represents value keys for contexts
type ValueKey string

const (
	// KeyCfg is the key for the muserv configuration
	KeyCfg ValueKey = "cfg"
	// KeyVersion is the key for the muserv version
	KeyVersion ValueKey = "version"
)

const (
	// CfgDir is the directory where the muserv configuration is stored
	CfgDir = "/etc/muserv"
	// IconDir is the directory where the muserv icons are stored
	IconDir = CfgDir + "/icons"
	// path of muserv configuration file
	cfgFilepath = CfgDir + "/config.json"
)

// audioMimeTypes contains the audio mime types that muserv supports
var audioMimeTypes = map[string]bool{
	"audio/aac":    true,
	"audio/flac":   true,
	"audio/mp4":    true,
	"audio/mpeg":   true,
	"audio/ogg":    true,
	"audio/x-flac": true,
}

// imageMimeTypes contains the image mime types that muserv supports
var imageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

const (
	// TagAlbum is the album tag
	TagAlbum = "album"
	// TagAlbumArtist is the album artist tag
	TagAlbumArtist = "albumartist"
	// TagArtist is the artist tag
	TagArtist = "artist"
	// TagGenre is the genre tag
	TagGenre = "genre"
	// TagTrack represents the track in the definition of content hierarchies
	TagTrack = "track"
)

// allowedHierarchies contains the allowed successors of tag in content
// hierarchies
var allowedHierarchies = map[string]([]string){
	TagGenre:       {TagAlbumArtist, TagArtist, TagAlbum, TagTrack},
	TagAlbumArtist: {TagAlbum},
	TagArtist:      {TagTrack},
	TagAlbum:       {TagTrack},
	TagTrack:       {},
}

// Cfg stores the data from the MuServ configuration file
type Cfg struct {
	Cnt      cnt    `json:"content"`
	UPnP     upnp   `json:"upnp"`
	LogDir   string `json:"log_dir"`
	LogLevel string `json:"log_level"`
}
type cnt struct {
	MusicDir       string        `json:"music_dir"`
	Separator      string        `json:"separator"`
	CacheDir       string        `json:"cache_dir"`
	UpdateMode     string        `json:"update_mode"`
	UpdateInterval time.Duration `json:"update_interval"`
}
type upnp struct {
	Interfaces []string    `json:"interfaces"`
	Port       int         `json:"port"`
	ServerName string      `json:"server_name"`
	UUID       string      `json:"udn"`
	MaxAge     int         `json:"max_age"`
	StatusFile string      `json:"status_file"`
	Device     device      `json:"device"`
	Hiers      []Hierarchy `json:"hierarchies"`
}
type device struct {
	Manufacturer     string `json:"manufacturer"`
	ManufacturerURL  string `json:"manufacturer_url"`
	ModelDescription string `json:"model_desc"`
	ModelName        string `json:"model_name"`
	ModelURL         string `json:"model_url"`
	ModelNumber      string `json:"model_no"`
	SerialNumber     string `json:"serial_no"`
	UPC              string `json:"upc"`
}

// Hierarchy contains the definition of one content hierarchy. Name must not be
// empty. Either ID or Levels must be filled, but not both. If ID is filled,
// it's a hierarchy that's defined muserv internally. Those hierarchies cannot
// be changed. They (can) occur in the config to configure where in the sequence
// of hierarchies they shall appear. Those hierarchies can be removed without
// problem. For the other hierarchies, Levels must be set.
type Hierarchy struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Levels []string `json:"levels"`
}

// IsValidAudioFile returns true if file has a mime type that is relevant for muserv as per
// the configuration, otherwise false is returned
func IsValidAudioFile(file string) bool {
	_, exists := audioMimeTypes[mime.TypeByExtension(path.Ext(file))]
	return exists
}

// SupportedMimeTypes assembles a string containing the audio and image mime
// types that muserv supports. The string is used to set the state variable
// SpurceProtocolInfo of the connection manager service
func SupportedMimeTypes() (s string) {
	for m := range audioMimeTypes {
		s += ",http-get:*:" + m + ":*"
	}
	for m := range imageMimeTypes {
		s += ",http-get:*:" + m + ":*"
	}
	// note: the leading comma must be removed
	return s[1:]
}

// Load reads the configuration file and returns the muserv config as structure
func Load() (cfg Cfg, err error) {
	cfgFile, err := ioutil.ReadFile(cfgFilepath)
	if err != nil {
		return Cfg{}, errors.Wrapf(err, "config file '%s' couldn't be read", cfgFilepath)
	}

	if err = json.Unmarshal(cfgFile, &cfg); err != nil {
		return Cfg{}, errors.Wrapf(err, "config file '%s' couldn't be marshalled", cfgFilepath)
	}

	return
}

// Validate check if the configuration is complete and correct. If it's not, an
// error is returned
func (me *Cfg) Validate() (err error) {
	// check if muserv system user exists
	if err = validateUser(); err != nil {
		return
	}

	// validate the existence of directories
	if err = validateDir(me.Cnt.MusicDir, "music_dir"); err != nil {
		return
	}
	if err = validateDir(me.Cnt.CacheDir, "cache_dir"); err != nil {
		return
	}
	if err = validateDir(me.LogDir, "log_dir"); err != nil {
		return
	}

	if me.Cnt.UpdateMode != "notify" && me.Cnt.UpdateMode != "scan" {
		err = fmt.Errorf("unknown update_mode '%s'", me.Cnt.UpdateMode)
		return
	}
	if me.Cnt.UpdateInterval <= 0 {
		err = fmt.Errorf("update_interval must be > 0")
		return
	}
	if me.UPnP.Port <= 0 {
		err = fmt.Errorf("port must be > 0")
		return
	}
	if len(me.UPnP.ServerName) == 0 {
		err = fmt.Errorf("the server must have a name, but server_name is empty")
		return
	}
	// if a UUID/UDN is set it must be a valid UUID. If it's empty, a new and
	// valid UUID will be generated later on
	if len(me.UPnP.UUID) > 0 {
		if _, err = uuid.Parse(me.UPnP.UUID); err != nil {
			err = errors.Wrapf(err, "the servers' UDN '%s' is not a valid UUID", me.UPnP.UUID)
			return
		}
	}
	if len(me.UPnP.StatusFile) == 0 {
		err = fmt.Errorf("status_file must not be empty")
		return
	}
	if me.UPnP.MaxAge <= 0 {
		err = fmt.Errorf("max_age must be > 0")
		return
	}

	// validate hierarchies
	if len(me.UPnP.Hiers) == 0 {
		err = fmt.Errorf("at least one hierarchy must be defined")
		return
	}
	for i := 0; i < len(me.UPnP.Hiers); i++ {
		if err = me.UPnP.Hiers[i].validate(); err != nil {
			return
		}
	}

	return
}

// Test reads the configuration file and checks the configuration for
// completeness and consistency
func Test() (err error) {
	var cfg Cfg

	if cfg, err = Load(); err != nil {
		err = errors.Wrap(err, "the muserv configuration file '/etc/muserv/config.json' couldn't be read")
		return
	}

	if err = cfg.Validate(); err != nil {
		return
	}

	fmt.Println("Congrats: The muserv configuration is complete and consistent :)")
	return
}

// validateDir checks if dir exists. name is the name that is used for that
// directory in the configuration
func validateDir(dir, name string) (err error) {
	if dir == "" {
		err = fmt.Errorf("no %s maintained", name)
		return
	}
	var exists bool
	if exists, err = file.Exists(dir); err != nil {
		err = errors.Wrapf(err, "cannot check if %s '%s' exists", name, dir)
		return
	}
	if !exists {
		err = fmt.Errorf("%s '%s' doesn't exist", name, dir)
		return
	}
	return
}

// validate checks if the hierarchy is OK. If it's not, an error is returned
func (me *Hierarchy) validate() (err error) {
	// either ID or Levels must be set but not both
	if len(me.ID) > 0 && len(me.Levels) > 0 {
		err = fmt.Errorf("hierarchy with id '%s' must not have levels", me.ID)
		return
	}
	// name must be set
	if len(me.Name) == 0 {
		if len(me.ID) > 0 {
			err = fmt.Errorf("hierarchy with id '%s' has no name", me.ID)
			return
		}
		err = fmt.Errorf("not all hierarchies have a name")
		return
	}

	if len(me.ID) > 0 {
		return
	}

	// levels must be set
	if len(me.Levels) == 0 {
		err = fmt.Errorf("hierarchy '%s' does not have levels", me.Name)
		return
	}

	// check levels (here, we know already that there is at least one level)
	for i, level := range me.Levels {
		allowedSuccs, exists := allowedHierarchies[level]
		if !exists {
			err = fmt.Errorf("hierarchy '%s' must not contain level '%s'", me.Name, level)
			return
		}
		// is successor allowed?
		if i < len(me.Levels)-1 {
			if !utils.Contains(allowedSuccs, me.Levels[i+1]) {
				err = fmt.Errorf("hierarchy '%s' must not contain '%s' as successor of '%s'", me.Name, me.Levels[i+1], level)
				return
			}
		}
	}

	return
}

func validateUser() (err error) {
	_, err = user.Lookup(UserName)
	if err != nil {
		err = errors.Wrap(err, "muserv system user does not exist")
		return
	}
	return
}
