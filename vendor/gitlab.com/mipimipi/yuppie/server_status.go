package yuppie

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	file "gitlab.com/mipimipi/go-utils/file"
)

// status represents the status of the UPnP server for persistence. The status
// is persisted in a JSON file
type status struct {
	BootID   uint32                         `json:"boot_id"`
	ConfigID uint32                         `json:"config_id"`
	Hashes   map[string]uint64              `json:"file_hashes"`
	StatVars map[string](map[string]string) `json:"state_vars"`
	Locals   map[string]string              `json:"local_vars,omitempty"`
}

// read reads the server status from a JSON file
func (me *status) read(filepath string) (err error) {
	exists, err := file.Exists(filepath)
	if err != nil {
		err := errors.Wrapf(err, "cannot check if status file '%s' exists", filepath)
		log.Fatal(err)
		return err
	}

	if !exists {
		log.Info("status file does not (yet) exist")
		return
	}

	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		err := errors.Wrapf(err, "couldn't read status file '%s'", filepath)
		log.Fatal(err)
		return err
	}
	if err = json.Unmarshal(f, me); err != nil {
		err := errors.Wrapf(err, "status file '%s' couldn't be unmarshalled: %v", filepath, err)
		log.Fatal(err)
		return err
	}

	return
}

// write writes the server status to a JSON file
func (me *status) write(filepath string) (err error) {
	var out []byte
	if out, err = json.MarshalIndent(me, "", "\t"); err != nil {
		log.Errorf("status couldn't be marshalled: %v", err)
		return
	}

	if err = ioutil.WriteFile(filepath, out, 0666); err != nil {
		log.Errorf("couldn't write status to file: %v", err)
		return
	}
	return
}

// setStatus reads the status value from file and sets the corresponding
// attributes of the server accordingly
func (me *Server) setStatus() (err error) {
	// set IDs from last status
	var st status
	if err = st.read(me.cfg.StatusFile); err != nil {
		return err
	}
	me.bootID.Set(st.BootID)
	me.configID.Set(st.ConfigID)

	// check if device or service config changed and adjust configID if
	// necessary
	if len(st.Hashes)-1 != len(me.services) {
		me.configID.Incr()
		return
	}
	for id, hash := range st.Hashes {
		var h uint64
		if id == "device::root" {
			me.Device.Desc.ConfigID = 0
			if h, err = me.Device.Desc.Hash(); err != nil {
				continue
			}
			if hash != h {
				log.Trace("config of root device changed: increase ConfigID")
				me.configID.Incr()
				continue
			}
		}
		if id[:9] == "service::" {
			svc, ok := me.services[id[9:]]
			if !ok {
				log.Tracef("service '%s' was deleted: increase ConfiID", id[9:])
				me.configID.Incr()
				continue
			}
			svc.desc.ConfigID = 0
			if h, err = svc.desc.Hash(); err != nil {
				continue
			}
			if hash != h {
				log.Tracef("config of service '%s' changed: increase ConfigID", id[9:])
				me.configID.Incr()
				continue
			}
		}
	}

	// set state variables
	for svcID, vars := range st.StatVars {
		svc, exists := me.services[svcID]
		if !exists {
			continue
		}
		for name, value := range vars {
			sv, exists := svc.stateVars[name]
			if !exists {
				continue
			}
			if err = sv.SetFromString(value); err != nil {
				return
			}
		}
	}

	// set locals
	me.Locals = st.Locals

	return
}

// writeStatus derives the to be persisted data and writes it to a file
func (me *Server) writeStatus() error {
	// collect status:
	// - IDs
	st := status{
		BootID:   me.bootID.Val(),
		ConfigID: me.configID.Val(),
		Locals:   me.Locals,
	}
	// - hashes
	//   initialize ConfigID to have a proper hash (ConfigID is not part of the
	//   configuration and thus would distort the hash)
	me.Device.Desc.ConfigID = 0
	st.Hashes = make(map[string]uint64)
	h, err := me.Device.Desc.Hash()
	if err != nil {
		return err
	}
	st.Hashes["device::root"] = h
	for id, svc := range me.services {
		// initialize ConfigID to have a proper hash (ConfigID is not part of the
		// configuration and thus would distort the hash)
		svc.desc.ConfigID = 0
		if h, err = svc.desc.Hash(); err != nil {
			return err
		}
		st.Hashes["service::"+id] = h
	}

	// state variables
	st.StatVars = make(map[string](map[string]string))
	for id, svc := range me.services {
		vars := make(map[string]string)
		for name, statVar := range svc.stateVars {
			if strings.ToUpper(name)[:11] == "A_ARG_TYPE_" {
				// according to the UPnP Device Architecture 2.0 spec, state
				// variables whose name starts wirh "A_ARG_TYPE_" do only exist
				// to assign a type to an argument. It makes no sense to store
				// these variables in the status file
				continue
			}
			vars[name] = statVar.String()
		}
		st.StatVars[id] = vars
	}

	// write status to file
	return st.write(me.cfg.StatusFile)
}
