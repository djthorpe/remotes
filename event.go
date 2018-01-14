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
	source     Codec
	ts         time.Duration
	devicecode uint32
	scancode   uint32
}

func NewRemoteEvent(source Codec, ts time.Duration, scancode, devicecode uint32) *RemoteEvent {
	return &RemoteEvent{
		source:     source,
		ts:         ts,
		scancode:   scancode,
		devicecode: devicecode,
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

func (*RemoteEvent) EventType() gopi.InputEventType {
	return gopi.INPUT_EVENT_KEYPRESS
}

func (this *RemoteEvent) Codec() RemoteCodec {
	return this.source.Type()
}

func (this *RemoteEvent) Scancode() uint32 {
	return this.scancode
}

func (this *RemoteEvent) Devicecode() uint32 {
	return this.devicecode
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
	return fmt.Sprintf("remotes.RemoteEvent{ scancode=0x%X devicecode=0x%X source=%v ts=%v }", this.scancode, this.devicecode, this.source, this.ts)
}
