package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"os/user"
	"path"
	p "path"
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
var audioMimeTypes = map[string]struct{}{
	"audio/aac":    {},
	"audio/flac":   {},
	"audio/mp4":    {},
	"audio/mpeg":   {},
	"audio/ogg":    {},
	"audio/opus":   {},
	"audio/x-flac": {},
}

// imageMimeTypes contains the image mime types that muserv supports
var imageMimeTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
}

// playlistMimeTypes contains the playlist mime types that muserv supports
var playlistMimeTypes = map[string]struct{}{
	"application/pls+xml": {},
	"audio/x-mpegurl":     {},
	"audio/x-scpls":       {},
}

// LevelType represents the type of a music hierarchy level
type LevelType string

// possible types of hierarchy levels
const (
	LvlAlbum       LevelType = "album"
	LvlAlbumArtist LevelType = "albumartist"
	LvlArtist      LevelType = "artist"
	LvlGenre       LevelType = "genre"
	LvlTrack       LevelType = "track"
)

// IsValid checks if the level type has a valid value
func (me LevelType) IsValid() (err error) {
	if me != LvlAlbum && me != LvlAlbumArtist && me != LvlArtist && me != LvlGenre && me != LvlTrack {
		err = fmt.Errorf("%s is no valid hierarchy level", me)
	}
	return
}

// allowedHierarchies contains the allowed successors of a level type in
// content hierarchies
var allowedHierarchies = map[LevelType]([]LevelType){
	LvlGenre:       {LvlAlbumArtist, LvlArtist, LvlAlbum, LvlTrack},
	LvlAlbumArtist: {LvlAlbum},
	LvlArtist:      {LvlTrack},
	LvlAlbum:       {LvlTrack},
	LvlTrack:       {},
}

// SortOrd represents the sort order (ascending or descending)
type SortOrd string

// sort orders
const (
	OrdAsc  SortOrd = "+" // sort ascending
	OrdDesc SortOrd = "-" // sort descending
)

// SortField represents an attribute that is used for sorting objects such as
// albums or tracks
type SortField string

// sort field values
const (
	SortNone       SortField = ""
	SortTitle      SortField = "title"
	SortTrackNo    SortField = "trackNo"
	SortDiscNo     SortField = "discNo"
	SortYear       SortField = "year"
	SortLastChange SortField = "lastChange"
)

// allowedSortFields contains the allowed sort fields per hierarchy level type.
// The types that are not listed here correspond single value tags (e.g. genre).
// Those can only be sorted by that single value and thus do not support other
// sort fields
var allowedSortFields = map[LevelType]([]SortField){
	LvlAlbum: {SortTitle, SortYear, SortLastChange},
	LvlTrack: {SortTitle, SortYear, SortLastChange, SortTrackNo, SortDiscNo},
}

// Cfg stores the data from the muserv configuration file
type Cfg struct {
	Cnt      cnt    `json:"content"`
	UPnP     upnp   `json:"upnp"`
	CacheDir string `json:"cache_dir"`
	LogDir   string `json:"log_dir"`
	LogLevel string `json:"log_level"`
}
type cnt struct {
	MusicDirs        []string      `json:"music_dirs"`
	Separator        string        `json:"separator"`
	UpdateMode       string        `json:"update_mode"`
	UpdateInterval   time.Duration `json:"update_interval"`
	Hiers            []Hierarchy   `json:"hierarchies"`
	ShowPlaylists    bool          `json:"show_playlists"`
	PlaylistHierName string        `json:"playlist_hierarchy_name"`
	ShowFolders      bool          `json:"show_folders"`
	FolderHierName   string        `json:"folder_hierarchy_name"`
}
type upnp struct {
	Interfaces []string `json:"interfaces"`
	Port       int      `json:"port"`
	ServerName string   `json:"server_name"`
	UUID       string   `json:"udn"`
	MaxAge     int      `json:"max_age"`
	StatusFile string   `json:"status_file"`
	Device     device   `json:"device"`
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
	Name   string  `json:"name"`
	Levels []level `json:"levels"`
}

type level struct {
	Type       LevelType `json:"type"`
	Sort       []string  `json:"sort"`
	sortFields []SortField
	comps      []Comparison
}

func (me *level) SortFields() []SortField {
	if len(me.sortFields) == 0 {
		me.assembleSortAttr()
	}
	return me.sortFields
}

// Comparison represents a "less" function for strings
type Comparison func(string, string) bool

func (me *level) Comparisons() [](Comparison) {
	if len(me.comps) == 0 {
		me.assembleSortAttr()
	}
	return me.comps
}

func (me *level) assembleSortAttr() {
	for _, s := range me.Sort {
		ord, sf := splitSort(s)
		me.sortFields = append(me.sortFields, sf)
		switch ord {
		case OrdAsc:
			me.comps = append(me.comps, func(a, b string) bool { return a < b })
		case OrdDesc:
			me.comps = append(me.comps, func(a, b string) bool { return a > b })
		}
	}
}

// IsValidPlaylistFile returns true if path is a playlist that is relevant for
// muserv as per the configuration, otherwise false is returned
func IsValidPlaylistFile(path string) bool {
	_, exists := playlistMimeTypes[mime.TypeByExtension(p.Ext(path))]
	return exists
}

// IsValidTrackFile returns true if path is a music track that is relevant for
// muserv as per the configuration, otherwise false is returned
func IsValidTrackFile(path string) bool {
	_, exists := audioMimeTypes[mime.TypeByExtension(p.Ext(path))]
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
	if err = validateDir(me.CacheDir, "cache_dir"); err != nil {
		return
	}
	if err = validateDir(me.LogDir, "log_dir"); err != nil {
		return
	}

	// check if muserv system user exists
	if err = validateUser(); err != nil {
		return
	}

	// validate UPnP config
	if err = me.UPnP.validate(); err != nil {
		return
	}

	// validate content config
	if err = me.Cnt.validate(); err != nil {
		return
	}

	return
}

// MusicDir returns the music directory that path belongs to. If path does not
// belong to any of the music directories, an empty string is returned.
// If path is absolute, then the path of the music directory is returned that
// path is a sub path of. An empty string is returned if path is sub path of
// none of them.
// If path is relative, then the path of first music directory is returned
// where the concatenation of the music directory and path exists in the file
// system. An empty string is returned if that path doesn't exist for any of
// the music directories.
func (me *cnt) MusicDir(path string) string {
	if len(path) == 0 {
		return ""
	}

	for _, dir := range me.MusicDirs {
		if p.IsAbs(path) {
			if isSub, _ := file.IsSub(dir, path); isSub {
				return dir
			}
			continue
		}
		if exists, _ := file.Exists(p.Join(dir, path)); exists {
			return dir
		}
	}

	return ""
}

// validate checks if the content part of the configuration is complete and
// correct. If it's not, an error is returned
func (me *cnt) validate() (err error) {
	// check if music directories are compliant
	if err = me.validateMusicDirs(); err != nil {
		return
	}

	if me.UpdateMode != "notify" && me.UpdateMode != "scan" {
		err = fmt.Errorf("unknown update_mode '%s'", me.UpdateMode)
		return
	}
	if me.UpdateInterval <= 0 {
		err = fmt.Errorf("update_interval must be > 0")
		return
	}

	// validate hierarchies
	if len(me.Hiers) == 0 {
		err = fmt.Errorf("at least one hierarchy must be defined")
		return
	}
	for i := 0; i < len(me.Hiers); i++ {
		if err = me.Hiers[i].validate(); err != nil {
			return
		}
	}

	// if playlists shall be shown in hierarchy, a name must be configured
	if me.ShowPlaylists && len(me.PlaylistHierName) == 0 {
		err = errors.New("hierarchy node for playlists must have a name")
		return
	}

	// if folder shall be shown in hierarchy, a name must be configured
	if me.ShowFolders && len(me.FolderHierName) == 0 {
		err = errors.New("hierarchy node for folders must have a name")
		return
	}

	return
}

func (me *cnt) validateMusicDirs() (err error) {
	if len(me.MusicDirs) == 0 {
		err = fmt.Errorf("at least one music directory must be configured")
		return

	}
	for _, dir := range me.MusicDirs {
		if err = validateDir(dir, "music_dir"); err != nil {
			return
		}
	}

	// music dirs must not be sub dirs of each other
	for i := 0; i < len(me.MusicDirs); i++ {
		for j := i + 1; j < len(me.MusicDirs); j++ {
			if isSub, _ := file.IsSub(me.MusicDirs[i], me.MusicDirs[j]); isSub {
				err = fmt.Errorf("music dir '%s' if sub dir of '%s'", me.MusicDirs[j], me.MusicDirs[i])
				return
			}
			if isSub, _ := file.IsSub(me.MusicDirs[j], me.MusicDirs[i]); isSub {
				err = fmt.Errorf("music dir '%s' if sub dir of '%s'", me.MusicDirs[i], me.MusicDirs[i])
				return
			}
		}
	}
	return
}

// validate checks if the UPnP part of the configuration is complete and
// correct. If it's not, an error is returned
func (me *upnp) validate() (err error) {
	if me.Port <= 0 {
		err = fmt.Errorf("port must be > 0")
		return
	}
	if len(me.ServerName) == 0 {
		err = fmt.Errorf("the server must have a name, but server_name is empty")
		return
	}
	// if a UUID/UDN is set it must be a valid UUID. If it's empty, a new and
	// valid UUID will be generated later on
	if len(me.UUID) > 0 {
		if _, err = uuid.Parse(me.UUID); err != nil {
			err = errors.Wrapf(err, "the servers' UDN '%s' is not a valid UUID", me.UUID)
			return
		}
	}
	if len(me.StatusFile) == 0 {
		err = fmt.Errorf("status_file must not be empty")
		return
	}
	if me.MaxAge <= 0 {
		err = fmt.Errorf("max_age must be > 0")
		return
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

// splitSort splits s into the sort order (which is indicated by the character
// of the sort field, "+" or "-") and the sort field itself (i.e. the part after
// the order indicator). If there's no order indicator, "+" is assumed
func splitSort(s string) (ord SortOrd, sf SortField) {
	if SortOrd(s[0]) == OrdAsc || SortOrd(s[0]) == OrdDesc {
		ord = SortOrd(s[0])
		sf = SortField(s[1:])
	} else {
		ord = OrdAsc
		sf = SortField(s)
	}
	return
}

// validateSort checks if s is a valid sort string (i.e. if it's of the form
// (+|-)<sort field>)
func validateSort(s string) (err error) {
	if len(s) == 0 {
		return
	}
	_, sf := splitSort(s)
	if sf != SortNone && sf != SortTitle && sf != SortTrackNo && sf != SortDiscNo && sf != SortYear && sf != SortLastChange {
		err = fmt.Errorf("%s is no valid sort field", s)
	}
	return
}

// validateDir checks if dir exists. name is the name that is used for that
// directory in the configuration
func validateDir(dir, name string) (err error) {
	if dir == "" {
		err = fmt.Errorf("empty %s is not acceptable", name)
		return
	}
	if !path.IsAbs(dir) {
		err = fmt.Errorf("%s '%s' is not absolute", name, dir)
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
	// name must be set
	if len(me.Name) == 0 {
		err = fmt.Errorf("not all hierarchies have a name")
		return
	}
	// levels must be set
	if len(me.Levels) == 0 {
		err = fmt.Errorf("hierarchy '%s' does not have levels", me.Name)
		return
	}

	// check levels (here, we know already that there is at least one level)
	for i, level := range me.Levels {
		// last level must be track
		if i == len(me.Levels)-1 && level.Type != LvlTrack {
			err = fmt.Errorf("last level of hierarchy '%s' must be track", me.Name)
			return
		}
		// is successor allowed?
		allowedSuccs, exists := allowedHierarchies[level.Type]
		if !exists {
			err = fmt.Errorf("hierarchy '%s' must not contain level '%s'", me.Name, level.Type)
			return
		}
		if i < len(me.Levels)-1 {
			if !utils.Contains(allowedSuccs, me.Levels[i+1].Type) {
				err = fmt.Errorf("hierarchy '%s' must not contain '%s' as successor of '%s'", me.Name, me.Levels[i+1].Type, level.Type)
				return
			}
		}
		// check sort fields
		for _, s := range level.Sort {
			if err = validateSort(s); err != nil {
				return
			}
			_, sf := splitSort(s)
			if !utils.Contains(allowedSortFields[level.Type], sf) {
				err = fmt.Errorf("hierarchy level '%s' cannot be sorted by '%s'", level.Type, sf)
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
