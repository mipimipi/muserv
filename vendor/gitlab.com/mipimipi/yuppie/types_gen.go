// ***********************************************************
// GENERATED FILE - DO NOT EDIT BY HAND.
// ***********************************************************

package yuppie

import (
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"

	utils "gitlab.com/mipimipi/go-utils"
)

// constructors maps UPnP types to constructor functions
var constructors = map[string]func(string) (StateVar, error){
	"ui1":         func(v string) (StateVar, error) { return newUpnpUI1(v) },
	"ui2":         func(v string) (StateVar, error) { return newUpnpUI2(v) },
	"ui4":         func(v string) (StateVar, error) { return newUpnpUI4(v) },
	"ui8":         func(v string) (StateVar, error) { return newUpnpUI8(v) },
	"i1":          func(v string) (StateVar, error) { return newUpnpI1(v) },
	"i2":          func(v string) (StateVar, error) { return newUpnpI2(v) },
	"i4":          func(v string) (StateVar, error) { return newUpnpI4(v) },
	"int":         func(v string) (StateVar, error) { return newUpnpInt(v) },
	"r4":          func(v string) (StateVar, error) { return newUpnpR4(v) },
	"r8":          func(v string) (StateVar, error) { return newUpnpR8(v) },
	"number":      func(v string) (StateVar, error) { return newUpnpNumber(v) },
	"fixed.14.4":  func(v string) (StateVar, error) { return newUpnpFixed14_4(v) },
	"float":       func(v string) (StateVar, error) { return newUpnpFloat(v) },
	"char":        func(v string) (StateVar, error) { return newUpnpChar(v) },
	"string":      func(v string) (StateVar, error) { return newUpnpString(v) },
	"date":        func(v string) (StateVar, error) { return newUpnpDate(v) },
	"dateTime":    func(v string) (StateVar, error) { return newUpnpDateTime(v) },
	"dateTime.tz": func(v string) (StateVar, error) { return newUpnpDateTimeTz(v) },
	"time":        func(v string) (StateVar, error) { return newUpnpTimeOfDay(v) },
	"time.tz":     func(v string) (StateVar, error) { return newUpnpTimeOfDayTz(v) },
	"boolean":     func(v string) (StateVar, error) { return newUpnpBoolean(v) },
	"bin.base64":  func(v string) (StateVar, error) { return newUpnpBinBase64(v) },
	"bin.hex":     func(v string) (StateVar, error) { return newUpnpBinHex(v) },
	"uri":         func(v string) (StateVar, error) { return newUpnpURI(v) },
}

// upnpUI1 is the representation of the UPnP type ui1 as Golang type
type upnpUI1 struct {
	val uint8
	*sync.Mutex
}

// newUpnpUI1 creates a new state variable that represents the UPnP type ui1
func newUpnpUI1(s string) (t *upnpUI1, err error) {
	v, err := unmarshalUpnpUI1(s)
	if err != nil {
		return nil, err
	}

	t = &upnpUI1{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpUI1) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpUI1) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "uint8" {
		err = fmt.Errorf("expected type uint8, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(uint8)
	return
}

// Init initializes t with a new value v
func (t *upnpUI1) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpUI1) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpUI1(s)
	return
}

// String returns the string representation of the value of t
func (t upnpUI1) String() string {
	s, _ := marshalUpnpUI1(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpUI1) Type() string { return "ui1" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpUI1) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpUI1) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpUI1) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpUI2 is the representation of the UPnP type ui2 as Golang type
type upnpUI2 struct {
	val uint16
	*sync.Mutex
}

// newUpnpUI2 creates a new state variable that represents the UPnP type ui2
func newUpnpUI2(s string) (t *upnpUI2, err error) {
	v, err := unmarshalUpnpUI2(s)
	if err != nil {
		return nil, err
	}

	t = &upnpUI2{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpUI2) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpUI2) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "uint16" {
		err = fmt.Errorf("expected type uint16, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(uint16)
	return
}

// Init initializes t with a new value v
func (t *upnpUI2) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpUI2) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpUI2(s)
	return
}

// String returns the string representation of the value of t
func (t upnpUI2) String() string {
	s, _ := marshalUpnpUI2(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpUI2) Type() string { return "ui2" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpUI2) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpUI2) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpUI2) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpUI4 is the representation of the UPnP type ui4 as Golang type
type upnpUI4 struct {
	val uint32
	*sync.Mutex
}

// newUpnpUI4 creates a new state variable that represents the UPnP type ui4
func newUpnpUI4(s string) (t *upnpUI4, err error) {
	v, err := unmarshalUpnpUI4(s)
	if err != nil {
		return nil, err
	}

	t = &upnpUI4{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpUI4) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpUI4) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "uint32" {
		err = fmt.Errorf("expected type uint32, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(uint32)
	return
}

// Init initializes t with a new value v
func (t *upnpUI4) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpUI4) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpUI4(s)
	return
}

// String returns the string representation of the value of t
func (t upnpUI4) String() string {
	s, _ := marshalUpnpUI4(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpUI4) Type() string { return "ui4" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpUI4) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpUI4) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpUI4) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpUI8 is the representation of the UPnP type ui8 as Golang type
type upnpUI8 struct {
	val uint64
	*sync.Mutex
}

// newUpnpUI8 creates a new state variable that represents the UPnP type ui8
func newUpnpUI8(s string) (t *upnpUI8, err error) {
	v, err := unmarshalUpnpUI8(s)
	if err != nil {
		return nil, err
	}

	t = &upnpUI8{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpUI8) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpUI8) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "uint64" {
		err = fmt.Errorf("expected type uint64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(uint64)
	return
}

// Init initializes t with a new value v
func (t *upnpUI8) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpUI8) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpUI8(s)
	return
}

// String returns the string representation of the value of t
func (t upnpUI8) String() string {
	s, _ := marshalUpnpUI8(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpUI8) Type() string { return "ui8" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpUI8) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpUI8) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpUI8) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpI1 is the representation of the UPnP type i1 as Golang type
type upnpI1 struct {
	val int8
	*sync.Mutex
}

// newUpnpI1 creates a new state variable that represents the UPnP type i1
func newUpnpI1(s string) (t *upnpI1, err error) {
	v, err := unmarshalUpnpI1(s)
	if err != nil {
		return nil, err
	}

	t = &upnpI1{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpI1) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpI1) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "int8" {
		err = fmt.Errorf("expected type int8, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(int8)
	return
}

// Init initializes t with a new value v
func (t *upnpI1) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpI1) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpI1(s)
	return
}

// String returns the string representation of the value of t
func (t upnpI1) String() string {
	s, _ := marshalUpnpI1(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpI1) Type() string { return "i1" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpI1) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpI1) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpI1) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpI2 is the representation of the UPnP type i2 as Golang type
type upnpI2 struct {
	val int16
	*sync.Mutex
}

// newUpnpI2 creates a new state variable that represents the UPnP type i2
func newUpnpI2(s string) (t *upnpI2, err error) {
	v, err := unmarshalUpnpI2(s)
	if err != nil {
		return nil, err
	}

	t = &upnpI2{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpI2) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpI2) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "int16" {
		err = fmt.Errorf("expected type int16, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(int16)
	return
}

// Init initializes t with a new value v
func (t *upnpI2) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpI2) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpI2(s)
	return
}

// String returns the string representation of the value of t
func (t upnpI2) String() string {
	s, _ := marshalUpnpI2(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpI2) Type() string { return "i2" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpI2) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpI2) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpI2) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpI4 is the representation of the UPnP type i4 as Golang type
type upnpI4 struct {
	val int32
	*sync.Mutex
}

// newUpnpI4 creates a new state variable that represents the UPnP type i4
func newUpnpI4(s string) (t *upnpI4, err error) {
	v, err := unmarshalUpnpI4(s)
	if err != nil {
		return nil, err
	}

	t = &upnpI4{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpI4) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpI4) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "int32" {
		err = fmt.Errorf("expected type int32, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(int32)
	return
}

// Init initializes t with a new value v
func (t *upnpI4) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpI4) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpI4(s)
	return
}

// String returns the string representation of the value of t
func (t upnpI4) String() string {
	s, _ := marshalUpnpI4(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpI4) Type() string { return "i4" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpI4) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpI4) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpI4) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpInt is the representation of the UPnP type int as Golang type
type upnpInt struct {
	val int64
	*sync.Mutex
}

// newUpnpInt creates a new state variable that represents the UPnP type int
func newUpnpInt(s string) (t *upnpInt, err error) {
	v, err := unmarshalUpnpInt(s)
	if err != nil {
		return nil, err
	}

	t = &upnpInt{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpInt) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpInt) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "int64" {
		err = fmt.Errorf("expected type int64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(int64)
	return
}

// Init initializes t with a new value v
func (t *upnpInt) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpInt) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpInt(s)
	return
}

// String returns the string representation of the value of t
func (t upnpInt) String() string {
	s, _ := marshalUpnpInt(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpInt) Type() string { return "int" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpInt) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpInt) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpInt) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpR4 is the representation of the UPnP type r4 as Golang type
type upnpR4 struct {
	val float32
	*sync.Mutex
}

// newUpnpR4 creates a new state variable that represents the UPnP type r4
func newUpnpR4(s string) (t *upnpR4, err error) {
	v, err := unmarshalUpnpR4(s)
	if err != nil {
		return nil, err
	}

	t = &upnpR4{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpR4) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpR4) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "float32" {
		err = fmt.Errorf("expected type float32, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(float32)
	return
}

// Init initializes t with a new value v
func (t *upnpR4) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpR4) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpR4(s)
	return
}

// String returns the string representation of the value of t
func (t upnpR4) String() string {
	s, _ := marshalUpnpR4(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpR4) Type() string { return "r4" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpR4) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpR4) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpR4) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpR8 is the representation of the UPnP type r8 as Golang type
type upnpR8 struct {
	val float64
	*sync.Mutex
}

// newUpnpR8 creates a new state variable that represents the UPnP type r8
func newUpnpR8(s string) (t *upnpR8, err error) {
	v, err := unmarshalUpnpR8(s)
	if err != nil {
		return nil, err
	}

	t = &upnpR8{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpR8) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpR8) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "float64" {
		err = fmt.Errorf("expected type float64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(float64)
	return
}

// Init initializes t with a new value v
func (t *upnpR8) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpR8) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpR8(s)
	return
}

// String returns the string representation of the value of t
func (t upnpR8) String() string {
	s, _ := marshalUpnpR8(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpR8) Type() string { return "r8" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpR8) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpR8) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpR8) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpNumber is the representation of the UPnP type number as Golang type
type upnpNumber struct {
	val float64
	*sync.Mutex
}

// newUpnpNumber creates a new state variable that represents the UPnP type number
func newUpnpNumber(s string) (t *upnpNumber, err error) {
	v, err := unmarshalUpnpNumber(s)
	if err != nil {
		return nil, err
	}

	t = &upnpNumber{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpNumber) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpNumber) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "float64" {
		err = fmt.Errorf("expected type float64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(float64)
	return
}

// Init initializes t with a new value v
func (t *upnpNumber) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpNumber) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpNumber(s)
	return
}

// String returns the string representation of the value of t
func (t upnpNumber) String() string {
	s, _ := marshalUpnpNumber(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpNumber) Type() string { return "number" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpNumber) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpNumber) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpNumber) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpFixed14_4 is the representation of the UPnP type fixed.14.4 as Golang type
type upnpFixed14_4 struct {
	val float64
	*sync.Mutex
}

// newUpnpFixed14_4 creates a new state variable that represents the UPnP type fixed.14.4
func newUpnpFixed14_4(s string) (t *upnpFixed14_4, err error) {
	v, err := unmarshalUpnpFixed14_4(s)
	if err != nil {
		return nil, err
	}

	t = &upnpFixed14_4{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpFixed14_4) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpFixed14_4) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "float64" {
		err = fmt.Errorf("expected type float64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(float64)
	return
}

// Init initializes t with a new value v
func (t *upnpFixed14_4) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpFixed14_4) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpFixed14_4(s)
	return
}

// String returns the string representation of the value of t
func (t upnpFixed14_4) String() string {
	s, _ := marshalUpnpFixed14_4(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpFixed14_4) Type() string { return "fixed.14.4" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpFixed14_4) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpFixed14_4) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpFixed14_4) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpFloat is the representation of the UPnP type float as Golang type
type upnpFloat struct {
	val float64
	*sync.Mutex
}

// newUpnpFloat creates a new state variable that represents the UPnP type float
func newUpnpFloat(s string) (t *upnpFloat, err error) {
	v, err := unmarshalUpnpFloat(s)
	if err != nil {
		return nil, err
	}

	t = &upnpFloat{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpFloat) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpFloat) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "float64" {
		err = fmt.Errorf("expected type float64, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(float64)
	return
}

// Init initializes t with a new value v
func (t *upnpFloat) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpFloat) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpFloat(s)
	return
}

// String returns the string representation of the value of t
func (t upnpFloat) String() string {
	s, _ := marshalUpnpFloat(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpFloat) Type() string { return "float" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpFloat) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpFloat) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpFloat) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpChar is the representation of the UPnP type char as Golang type
type upnpChar struct {
	val rune
	*sync.Mutex
}

// newUpnpChar creates a new state variable that represents the UPnP type char
func newUpnpChar(s string) (t *upnpChar, err error) {
	v, err := unmarshalUpnpChar(s)
	if err != nil {
		return nil, err
	}

	t = &upnpChar{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpChar) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpChar) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "rune" {
		err = fmt.Errorf("expected type rune, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(rune)
	return
}

// Init initializes t with a new value v
func (t *upnpChar) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpChar) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpChar(s)
	return
}

// String returns the string representation of the value of t
func (t upnpChar) String() string {
	s, _ := marshalUpnpChar(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpChar) Type() string { return "char" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpChar) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpChar) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpChar) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpString is the representation of the UPnP type string as Golang type
type upnpString struct {
	val string
	*sync.Mutex
}

// newUpnpString creates a new state variable that represents the UPnP type string
func newUpnpString(s string) (t *upnpString, err error) {
	v, err := unmarshalUpnpString(s)
	if err != nil {
		return nil, err
	}

	t = &upnpString{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpString) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpString) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "string" {
		err = fmt.Errorf("expected type string, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(string)
	return
}

// Init initializes t with a new value v
func (t *upnpString) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpString) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpString(s)
	return
}

// String returns the string representation of the value of t
func (t upnpString) String() string {
	s, _ := marshalUpnpString(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpString) Type() string { return "string" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpString) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpString) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpString) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpDate is the representation of the UPnP type date as Golang type
type upnpDate struct {
	val time.Time
	*sync.Mutex
}

// newUpnpDate creates a new state variable that represents the UPnP type date
func newUpnpDate(s string) (t *upnpDate, err error) {
	v, err := unmarshalUpnpDate(s)
	if err != nil {
		return nil, err
	}

	t = &upnpDate{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpDate) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpDate) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "time.Time" {
		err = fmt.Errorf("expected type time.Time, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(time.Time)
	return
}

// Init initializes t with a new value v
func (t *upnpDate) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpDate) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpDate(s)
	return
}

// String returns the string representation of the value of t
func (t upnpDate) String() string {
	s, _ := marshalUpnpDate(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpDate) Type() string { return "date" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpDate) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpDate) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpDate) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpDateTime is the representation of the UPnP type dateTime as Golang type
type upnpDateTime struct {
	val time.Time
	*sync.Mutex
}

// newUpnpDateTime creates a new state variable that represents the UPnP type dateTime
func newUpnpDateTime(s string) (t *upnpDateTime, err error) {
	v, err := unmarshalUpnpDateTime(s)
	if err != nil {
		return nil, err
	}

	t = &upnpDateTime{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpDateTime) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpDateTime) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "time.Time" {
		err = fmt.Errorf("expected type time.Time, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(time.Time)
	return
}

// Init initializes t with a new value v
func (t *upnpDateTime) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpDateTime) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpDateTime(s)
	return
}

// String returns the string representation of the value of t
func (t upnpDateTime) String() string {
	s, _ := marshalUpnpDateTime(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpDateTime) Type() string { return "dateTime" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpDateTime) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpDateTime) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpDateTime) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpDateTimeTz is the representation of the UPnP type dateTime.tz as Golang type
type upnpDateTimeTz struct {
	val time.Time
	*sync.Mutex
}

// newUpnpDateTimeTz creates a new state variable that represents the UPnP type dateTime.tz
func newUpnpDateTimeTz(s string) (t *upnpDateTimeTz, err error) {
	v, err := unmarshalUpnpDateTimeTz(s)
	if err != nil {
		return nil, err
	}

	t = &upnpDateTimeTz{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpDateTimeTz) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpDateTimeTz) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "time.Time" {
		err = fmt.Errorf("expected type time.Time, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(time.Time)
	return
}

// Init initializes t with a new value v
func (t *upnpDateTimeTz) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpDateTimeTz) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpDateTimeTz(s)
	return
}

// String returns the string representation of the value of t
func (t upnpDateTimeTz) String() string {
	s, _ := marshalUpnpDateTimeTz(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpDateTimeTz) Type() string { return "dateTime.tz" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpDateTimeTz) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpDateTimeTz) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpDateTimeTz) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpTimeOfDay is the representation of the UPnP type time as Golang type
type upnpTimeOfDay struct {
	val timeOfDay
	*sync.Mutex
}

// newUpnpTimeOfDay creates a new state variable that represents the UPnP type time
func newUpnpTimeOfDay(s string) (t *upnpTimeOfDay, err error) {
	v, err := unmarshalUpnpTimeOfDay(s)
	if err != nil {
		return nil, err
	}

	t = &upnpTimeOfDay{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpTimeOfDay) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpTimeOfDay) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "timeOfDay" {
		err = fmt.Errorf("expected type timeOfDay, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(timeOfDay)
	return
}

// Init initializes t with a new value v
func (t *upnpTimeOfDay) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpTimeOfDay) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpTimeOfDay(s)
	return
}

// String returns the string representation of the value of t
func (t upnpTimeOfDay) String() string {
	s, _ := marshalUpnpTimeOfDay(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpTimeOfDay) Type() string { return "time" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpTimeOfDay) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpTimeOfDay) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpTimeOfDay) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpTimeOfDayTz is the representation of the UPnP type time.tz as Golang type
type upnpTimeOfDayTz struct {
	val timeOfDay
	*sync.Mutex
}

// newUpnpTimeOfDayTz creates a new state variable that represents the UPnP type time.tz
func newUpnpTimeOfDayTz(s string) (t *upnpTimeOfDayTz, err error) {
	v, err := unmarshalUpnpTimeOfDayTz(s)
	if err != nil {
		return nil, err
	}

	t = &upnpTimeOfDayTz{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpTimeOfDayTz) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpTimeOfDayTz) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "timeOfDay" {
		err = fmt.Errorf("expected type timeOfDay, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(timeOfDay)
	return
}

// Init initializes t with a new value v
func (t *upnpTimeOfDayTz) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpTimeOfDayTz) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpTimeOfDayTz(s)
	return
}

// String returns the string representation of the value of t
func (t upnpTimeOfDayTz) String() string {
	s, _ := marshalUpnpTimeOfDayTz(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpTimeOfDayTz) Type() string { return "time.tz" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpTimeOfDayTz) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpTimeOfDayTz) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpTimeOfDayTz) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpBoolean is the representation of the UPnP type boolean as Golang type
type upnpBoolean struct {
	val bool
	*sync.Mutex
}

// newUpnpBoolean creates a new state variable that represents the UPnP type boolean
func newUpnpBoolean(s string) (t *upnpBoolean, err error) {
	v, err := unmarshalUpnpBoolean(s)
	if err != nil {
		return nil, err
	}

	t = &upnpBoolean{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpBoolean) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpBoolean) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "bool" {
		err = fmt.Errorf("expected type bool, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(bool)
	return
}

// Init initializes t with a new value v
func (t *upnpBoolean) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpBoolean) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpBoolean(s)
	return
}

// String returns the string representation of the value of t
func (t upnpBoolean) String() string {
	s, _ := marshalUpnpBoolean(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpBoolean) Type() string { return "boolean" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpBoolean) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpBoolean) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpBoolean) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpBinBase64 is the representation of the UPnP type bin.base64 as Golang type
type upnpBinBase64 struct {
	val []byte
	*sync.Mutex
}

// newUpnpBinBase64 creates a new state variable that represents the UPnP type bin.base64
func newUpnpBinBase64(s string) (t *upnpBinBase64, err error) {
	v, err := unmarshalUpnpBinBase64(s)
	if err != nil {
		return nil, err
	}

	t = &upnpBinBase64{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpBinBase64) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpBinBase64) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "[]byte" {
		err = fmt.Errorf("expected type []byte, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.([]byte)
	return
}

// Init initializes t with a new value v
func (t *upnpBinBase64) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpBinBase64) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpBinBase64(s)
	return
}

// String returns the string representation of the value of t
func (t upnpBinBase64) String() string {
	s, _ := marshalUpnpBinBase64(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpBinBase64) Type() string { return "bin.base64" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpBinBase64) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpBinBase64) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpBinBase64) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpBinHex is the representation of the UPnP type bin.hex as Golang type
type upnpBinHex struct {
	val []byte
	*sync.Mutex
}

// newUpnpBinHex creates a new state variable that represents the UPnP type bin.hex
func newUpnpBinHex(s string) (t *upnpBinHex, err error) {
	v, err := unmarshalUpnpBinHex(s)
	if err != nil {
		return nil, err
	}

	t = &upnpBinHex{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpBinHex) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpBinHex) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "[]byte" {
		err = fmt.Errorf("expected type []byte, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.([]byte)
	return
}

// Init initializes t with a new value v
func (t *upnpBinHex) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpBinHex) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpBinHex(s)
	return
}

// String returns the string representation of the value of t
func (t upnpBinHex) String() string {
	s, _ := marshalUpnpBinHex(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpBinHex) Type() string { return "bin.hex" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpBinHex) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpBinHex) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpBinHex) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// upnpURI is the representation of the UPnP type uri as Golang type
type upnpURI struct {
	val *url.URL
	*sync.Mutex
}

// newUpnpURI creates a new state variable that represents the UPnP type uri
func newUpnpURI(s string) (t *upnpURI, err error) {
	v, err := unmarshalUpnpURI(s)
	if err != nil {
		return nil, err
	}

	t = &upnpURI{v, &sync.Mutex{}}
	return
}

// Get returns the value of t
func (t upnpURI) Get() interface{} {
	return t.val
}

// Set sets t to the new value v
func (t *upnpURI) Set(v interface{}) (err error) {
	if reflect.TypeOf(v).String() != "*url.URL" {
		err = fmt.Errorf("expected type *url.URL, received: %s", reflect.TypeOf(v))
		return
	}
	t.val = v.(*url.URL)
	return
}

// Init initializes t with a new value v
func (t *upnpURI) Init(v interface{}) (err error) {
	return t.Set(v)
}

// SetFromString sets the value of t from s
func (t *upnpURI) SetFromString(s string) (err error) {
	t.val, err = unmarshalUpnpURI(s)
	return
}

// String returns the string representation of the value of t
func (t upnpURI) String() string {
	s, _ := marshalUpnpURI(t.val)

	return s
}

// Type returns the UPnP type of t
func (t upnpURI) Type() string { return "uri" }

// IsNumeric returns true is the value of t is numeric, otherwise false is
// returned
func (t *upnpURI) IsNumeric() bool { return isNumeric(t.val) }

// IsString returns true is the value of t is a string, otherwise false is
// returned
func (t *upnpURI) IsString() bool { return isString(t.val) }

// IsZero returns true is the value of t is the zero value of that type,
// otherwise false is returned
func (t upnpURI) IsZero() bool { return reflect.ValueOf(t.val).IsZero() }

// isNumeric returns true is val is a numeric value, otherwise false is returned
func isNumeric(val interface{}) bool {
	return utils.Contains(
		[]reflect.Kind{reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		},
		reflect.ValueOf(val).Kind())
}

// isString returns true is val is a string, otherwise false is returned
func isString(val interface{}) bool { return reflect.ValueOf(val).Kind() == reflect.String }
