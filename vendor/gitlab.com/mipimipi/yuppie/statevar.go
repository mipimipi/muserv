package yuppie

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gitlab.com/mipimipi/yuppie/desc"
	"gitlab.com/mipimipi/yuppie/internal/events"
)

// StateVar represents a SOAP variable (e.g. a SOAP state variable)
type StateVar interface {
	Type() string
	Init(interface{}) error
	Get() interface{}
	Set(interface{}) error
	SetFromString(string) error
	IsNumeric() bool
	IsString() bool
	IsZero() bool
	String() string
	Lock()
	Unlock()
}

// newStateVar creates a new state variable for the UPnP type typ with the value
// val
func newStateVar(typ, val string) (sv StateVar, err error) {
	f, exists := constructors[typ]
	if !exists {
		err = fmt.Errorf("could not create state variable of type '%s', value '%s'", typ, val)
		return
	}
	return f(val)
}

// StateVar represents a state variable
type stateVar struct {
	name            string
	service         *service
	toBeEvented     bool
	toBeMulticasted bool
	evented         bool
	def             StateVar
	list            map[string]bool
	rng             rng
	listener        func() chan events.StateVar
	StateVar
	*sync.Mutex
}

// rng represents a range
type rng struct{ Min, Max StateVar }

// isZero returns true if both, min and max of the range are zero
func (rng rng) isZero() bool {
	return rng.Min.IsZero() && rng.Max.IsZero()
}

// stateVarFromDesc creates a new state variable based on a description (i.e. the
// corresponding state variable part of the service description) and for the
// service svc. A listener for multicast eventing is added.
func stateVarFromDesc(sv desc.StateVariable, svc *service, listener func() chan events.StateVar) (*stateVar, error) {
	def, err := newStateVar(sv.DataType, sv.DefaultValue)
	if err != nil {
		err = errors.Wrap(err, "could no create state variable from description")
		return nil, err
	}

	stateVar := stateVar{
		name:            strings.TrimSpace(sv.Name),
		service:         svc,
		toBeEvented:     (sv.SendEvents == "yes"),
		toBeMulticasted: (sv.Multicast == "yes"),
		listener:        listener,
		StateVar:        def,
	}

	if sv.DefaultValue != "" {
		stateVar.def = def
	}

	if stateVar.IsNumeric() {
		var (
			min, max StateVar
			err      error
		)
		if sv.AllowedValueRange.IsZero() {
			min, err = newStateVar(sv.DataType, "0")
			if err != nil {
				err = errors.Wrap(err, "could no create state variable for zero minimum value")
				return nil, err
			}
			max, err = newStateVar(sv.DataType, "0")
			if err != nil {
				err = errors.Wrap(err, "could no create state variable for zero maximum value")
				return nil, err
			}
		} else {
			min, err = newStateVar(sv.DataType, sv.AllowedValueRange.Minimum)
			if err != nil {
				err = errors.Wrap(err, "could no create state variable for minimum value")
				return nil, err
			}
			max, err = newStateVar(sv.DataType, sv.AllowedValueRange.Maximum)
			if err != nil {
				err = errors.Wrap(err, "could no create state variable for maximum value")
				return nil, err
			}
		}
		stateVar.rng = rng{
			Min: min,
			Max: max,
		}
	}

	if stateVar.IsString() && len(sv.AllowedValueList) > 0 {
		stateVar.list = make(map[string]bool)
		for _, allowed := range sv.AllowedValueList {
			stateVar.list[allowed] = true
		}
	}

	return &stateVar, nil
}

// Name returns the name of the state variable
func (me *stateVar) Name() string {
	return me.name
}

// ServiceType returns the type of the service that the state variable belongs
// to
func (me *stateVar) ServiceType() string {
	return string(me.service.typ)
}

// ServiceVersion returns the version of the service the state variable belongs
// to
func (me *stateVar) ServiceVersion() string {
	return string(me.service.ver)
}

// DeviceUDN returns the UDN of the device that the service provides the state
// variable belongs to
func (me *stateVar) DeviceUDN() string {
	return me.service.device.UDN
}

// ServiceID returns the id of the service the state variable belongs to
func (me *stateVar) ServiceID() string {
	return string(me.service.id)
}

// SendEvent triggers the sending of a multicast event for the state variable
func (me *stateVar) SendEvent() {
	if me.toBeEvented && me.toBeMulticasted {
		log.Tracef("sent event for state variable '%s'", me.name)
		me.listener() <- me
	}
}

// Init initializes the state variable with v.
// Note: No multicast event is sent.
func (me *stateVar) Init(v interface{}) (err error) {
	if err = me.Set(v); err != nil {
		return errors.Wrapf(err, "could not initialize state variable '%s'", me.name)
	}
	return
}

// Set sets the value of the state variable to v
// Note: A multicast event is sent.
func (me *stateVar) Set(v interface{}) (err error) {
	// nothing to do if value would not change
	if reflect.ValueOf(me.Get()) != reflect.ValueOf(v) {
		return
	}

	// new eventing required
	me.evented = false

	// inform event listener about change
	if me.toBeEvented || me.toBeMulticasted {
		me.listener() <- me
	}

	return me.StateVar.Set(v)
}

// IsValid checks if the value of the state variable is valid. I.e. for numeric
// variables it's checked whether their value is in required range (if there is
// a range for that variable), for string variables it's checked whether the
// value is allowed (if there is a list of allowed values for that variable)
func (me *stateVar) IsValid(s string) (bool, UPnPErrorCode) {
	v, err := newStateVar(me.Type(), s)
	if err != nil {
		return false, UPnPErrorInvalidArgs
	}

	if me.IsNumeric() && !me.rng.isZero() {
		switch reflect.ValueOf(v.Get()).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if reflect.ValueOf(v.Get()).Int() < reflect.ValueOf(me.rng.Min.Get()).Int() || reflect.ValueOf(v.Get()).Int() > reflect.ValueOf(me.rng.Max.Get()).Int() {
				log.Errorf("state variable %s is out of range", me.name)
				return false, UPnPErrorArgValOutOfRange
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if reflect.ValueOf(v.Get()).Uint() < reflect.ValueOf(me.rng.Min.Get()).Uint() || reflect.ValueOf(v.Get()).Uint() > reflect.ValueOf(me.rng.Max.Get()).Uint() {
				log.Errorf("state variable %s is out of range", me.name)
				return false, UPnPErrorArgValOutOfRange
			}
		case reflect.Float32, reflect.Float64:
			if reflect.ValueOf(v.Get()).Float() < reflect.ValueOf(me.rng.Min.Get()).Float() || reflect.ValueOf(v.Get()).Float() > reflect.ValueOf(me.rng.Max.Get()).Float() {
				log.Errorf("state variable %s is out of range", me.name)
				return false, UPnPErrorArgValOutOfRange
			}
		default:
			log.Fatalf("state variable '%s' is not numeric", me.name)
			return false, UPnPErrorInvalidArgs
		}
	}

	if me.IsString() && me.list != nil && len(me.list) > 0 {
		_, allowed := me.list[s]
		if !allowed {
			log.Errorf("value '%s' of state variable %s is not allowed", me.Get().(string), me.name)
			return false, UPnPErrorArgValInvalid
		}
	}

	return true, 0
}

func (me *stateVar) SetEvented(evented bool) {
	me.evented = evented
}

func (me *stateVar) Evented() bool {
	return me.evented
}

func (me *stateVar) ToBeEvented() bool {
	return me.toBeEvented
}

func (me *stateVar) ToBeMulticasted() bool {
	return me.toBeMulticasted
}
