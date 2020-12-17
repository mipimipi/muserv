package desc

import (
	"fmt"
	"strings"
)

// Action represents an action from a service description
type Action struct {
	Name      string     `xml:"name"`
	Arguments []Argument `xml:"argumentList>argument"`
}

// Trim removes leading and trailing spaces for some attributes
func (me *Action) Trim() {
	me.Name = strings.TrimSpace(me.Name)
	for i := 0; i < len(me.Arguments); i++ {
		me.Arguments[i].Trim()
	}
}

// Validate checks whether the attribute values are OK. Problem messages are
// added to res.
func (me *Action) Validate(res *[]string) (ok bool) {
	me.Trim()

	ok = true
	if me.Name == "" {
		ok = false
		*res = append(*res, "action: name must not be empty")
	}
	for _, arg := range me.Arguments {
		b := arg.Validate(res)
		ok = ok && b
	}

	return
}

// Argument represents an argument from a service description
type Argument struct {
	Name                 string `xml:"name"`
	Direction            string `xml:"direction"`
	RelatedStateVariable string `xml:"relatedStateVariable"`
}

// Trim removes leading and trailing spaces for some attributes
func (me *Argument) Trim() {
	me.Name = strings.TrimSpace(me.Name)
	me.Direction = strings.ToLower(strings.TrimSpace(me.Direction))
	me.RelatedStateVariable = strings.TrimSpace(me.RelatedStateVariable)
}

// Validate checks whether the attribute values are OK. Problem messages are
// added to res.
func (me *Argument) Validate(res *[]string) (ok bool) {
	me.Trim()

	ok = true
	// name
	if me.Name == "" {
		ok = false
		*res = append(*res, "argument: name must not be empty")
	}
	// direction
	if me.Direction != "in" && me.Direction != "out" {
		ok = false
		*res = append(*res, fmt.Sprintf("argument: wrong direction: %s", me.Direction))
	}

	return
}

// StateVariable represents a state variable from a service description
type StateVariable struct {
	Name              string            `xml:"name"`
	SendEvents        string            `xml:"sendEvents,attr,omitempty"`
	Multicast         string            `xml:"multicast,attr,omitempty"`
	DataType          string            `xml:"dataType"`
	DefaultValue      string            `xml:"defaultValue"`
	AllowedValueList  []string          `xml:"allowedValueList>allowedValue,omitempty"`
	AllowedValueRange AllowedValueRange `xml:"allowedValueRange,omitempty"`
}

// Trim removes leading and trailing spaces for some attributes
func (me *StateVariable) Trim() {
	me.Name = strings.TrimSpace(me.Name)
	me.SendEvents = strings.ToLower(strings.TrimSpace(me.SendEvents))
	me.Multicast = strings.ToLower(strings.TrimSpace(me.Multicast))
	me.DataType = strings.TrimSpace(me.DataType)
	me.DefaultValue = strings.TrimSpace(me.DefaultValue)
	for i := 0; i < len(me.AllowedValueList); i++ {
		me.AllowedValueList[i] = strings.TrimSpace(me.AllowedValueList[i])
	}
}

// Validate checks whether the attribute values are OK. Problem messages are
// added to res.
func (me *StateVariable) Validate(res *[]string) (ok bool) {
	me.Trim()

	ok = true
	// name
	if me.Name == "" {
		ok = false
		*res = append(*res, "variable: name must not be empty")
	}

	return
}

// AllowedValueRange represents the allowed value range of aa state variable
type AllowedValueRange struct {
	Minimum string `xml:"minimum"`
	Maximum string `xml:"maximum"`
	Step    string `xml:"step"`
}

// IsZero returns true if the allowed value range is empty, otherwise false
func (me *AllowedValueRange) IsZero() bool {
	return me.Minimum == "" && me.Maximum == "" && me.Step == ""
}
