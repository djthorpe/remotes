/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package remotes

/*
	This file implements the event which is emitted on remote key
	invocation
*/

import (
	"fmt"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

/////////////////////////////////////////////////////////////////////
// RemoteEvent Implementation

type remoteevent struct {
	source   Codec
	ts       time.Duration
	device   uint32
	scancode uint32
	repeat   bool
}

func NewRemoteEvent(source Codec, ts time.Duration, scancode, device uint32, repeat bool) RemoteEvent {
	return &remoteevent{
		source:   source,
		ts:       ts,
		scancode: scancode,
		device:   device,
		repeat:   repeat,
	}
}

func (this *remoteevent) Source() gopi.Driver {
	return this.source
}

func (this *remoteevent) Name() string {
	return "RemoteEvent"
}

func (this *remoteevent) Timestamp() time.Duration {
	return this.ts
}

func (*remoteevent) DeviceType() gopi.InputDeviceType {
	return gopi.INPUT_TYPE_REMOTE
}

func (this *remoteevent) EventType() gopi.InputEventType {
	if this.repeat {
		return gopi.INPUT_EVENT_KEYREPEAT
	} else {
		return gopi.INPUT_EVENT_KEYPRESS
	}
}

func (this *remoteevent) Codec() CodecType {
	return this.source.Type()
}

func (this *remoteevent) ScanCode() uint32 {
	return this.scancode
}

func (this *remoteevent) Device() uint32 {
	return this.device
}

func (*remoteevent) Position() gopi.Point {
	return gopi.ZeroPoint
}

func (*remoteevent) Relative() gopi.Point {
	return gopi.ZeroPoint
}

func (*remoteevent) KeyCode() gopi.KeyCode {
	return gopi.KEYCODE_NONE
}

func (*remoteevent) KeyState() gopi.KeyState {
	return gopi.KEYSTATE_NONE
}

func (*remoteevent) Slot() uint {
	return 0
}

func (this *remoteevent) String() string {
	return fmt.Sprintf("remotes.RemoteEvent{ scancode=0x%X device=0x%X repeat=%v codec=%v ts=%v source=%v }", this.scancode, this.device, this.repeat, this.Codec(), this.ts, this.source)
}
