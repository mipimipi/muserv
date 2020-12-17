// Package types implements basic types
package types

// BootID represents BOOTID.UPNP.ORG asdescribed in the of the UPnP Device
// Architecture 2.0
type BootID struct {
	id   uint32
	next uint32
}

// NewBootID creates a new boot id
func NewBootID() (id *BootID) {
	id = new(BootID)
	id.id = 0
	id.next = 1
	return
}

// Val returns the current value of boot id
func (me *BootID) Val() uint32 { return me.id }

// Next returns the next value that boot id would have
func (me *BootID) Next() uint32 { return me.next }

// Incr increases the current and the next value of boot id
func (me *BootID) Incr() {
	me.id = me.next
	me.next++
}

// Set sets the current value of boot id to i and set the next value
// accordingly
func (me *BootID) Set(i uint32) {
	me.id = i
	me.next = i + 1
}

// maxConfigID is the maximum value of CONFIGID.UPNP.ORG according to the UPnP
// Device Architecture 2.0
const maxConfigID uint32 = 16777215

// ConfigID represents CONFIGID.UPNP.ORG asdescribed in the of the UPnP Device
// Architecture 2.0
type ConfigID struct {
	id uint32
}

// Val returns the current value of config id
func (me *ConfigID) Val() uint32 { return me.id }

// Incr increases the current value of config id
func (me *ConfigID) Incr() {
	if me.id == maxConfigID {
		me.id = 0
		return
	}
	me.id++
}

// Set sets the current value of config id to i
func (me *ConfigID) Set(i uint32) {
	me.id = i
}
