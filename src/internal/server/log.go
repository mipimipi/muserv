package server

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	l "github.com/sirupsen/logrus"
	"gitlab.com/go-utilities/file"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

const logFilename = "muserv.log"

// setupLogging sets up logging into file logDir with the level logLevel. If
// the log file does not exist yet, it is created. Its owner will be user
// userName (see constants).
func setupLogging(logDir, logLevel string) (err error) {
	// set up logging: no log entries possible before this statement!
	level, err := l.ParseLevel(logLevel)
	if err != nil {
		return
	}

	path := filepath.Join(logDir, logFilename)

	exists, err := file.Exists(path)
	if err != nil {
		return
	}

	// create or open file for write & append
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return
	}

	// if the log file did not exist and thus was just created, make sure that
	// user muserv is the owner
	if !exists {
		var u *user.User
		u, err = user.Lookup(config.UserName)
		if err != nil {
			return
		}
		var uid, gid int
		if uid, err = strconv.Atoi(u.Uid); err != nil {
			return
		}
		if gid, err = strconv.Atoi(u.Gid); err != nil {
			return
		}
		// get owner of the log file
		var info os.FileInfo
		info, err = os.Stat(path)
		if err != nil {
			return
		}
		stat := info.Sys().(*syscall.Stat_t)
		if uid != int(stat.Uid) || gid != int(stat.Gid) {
			if err = f.Chown(uid, gid); err != nil {
				return
			}
		}
	}

	l.SetOutput(f)
	l.SetLevel(level)
	return
}
