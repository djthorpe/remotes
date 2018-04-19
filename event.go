/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"fmt"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

/////////////////////////////////////////////////////////////////////
// gopi.InputEvent implementation

type RemoteEvent struct {
	source   Codec
	ts       time.Duration
	device   uint32
	scancode uint32
	repeat   bool
}

func NewRemoteEvent(source Codec, ts time.Duration, scancode, device uint32, repeat bool) *RemoteEvent {
	return &RemoteEvent{
		source:   source,
		ts:       ts,
		scancode: scancode,
		device:   device,
		repeat:   repeat,
	}
}

func (this *RemoteEvent) Source() gopi.Driver {
	return this.source
}

func (this *RemoteEvent) Name() string {
	return "RemoteEvent"
}

func (this *RemoteEvent) Timestamp() time.Duration {
	return this.ts
}

func (*RemoteEvent) DeviceType() gopi.InputDeviceType {
	return gopi.INPUT_TYPE_REMOTE
}

func (this *RemoteEvent) EventType() gopi.InputEventType {
	if this.repeat {
		return gopi.INPUT_EVENT_KEYREPEAT
	} else {
		return gopi.INPUT_EVENT_KEYPRESS
	}
}

func (this *RemoteEvent) Codec() RemoteCodec {
	return this.source.Type()
}

func (this *RemoteEvent) Scancode() uint32 {
	return this.scancode
}

func (this *RemoteEvent) Device() uint32 {
	return this.device
}

func (*RemoteEvent) Position() gopi.Point {
	return gopi.ZeroPoint
}

func (*RemoteEvent) Relative() gopi.Point {
	return gopi.ZeroPoint
}

func (*RemoteEvent) Slot() uint {
	return 0
}

func (this *RemoteEvent) String() string {
	return fmt.Sprintf("remotes.RemoteEvent{ scancode=0x%X device=0x%X repeat=%v source=%v ts=%v }", this.scancode, this.device, this.repeat, this.source, this.ts)
}
