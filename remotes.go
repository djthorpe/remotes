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
	"math"
	"time"

	"github.com/djthorpe/gopi"
)

// A coder/decoder for remote
type Codec interface {
	gopi.Driver
	gopi.Publisher

	// Return short name for the codec
	Name() string
}

/////////////////////////////////////////////////////////////////////
// Mark Space implementation

// A mark or space value
type MarkSpace struct {
	Type     gopi.LIRCType
	Min, Max uint32
}

func NewMarkSpace(t gopi.LIRCType, value, tolerance uint32) *MarkSpace {
	delta := float64(value) * float64(tolerance) / 100.0
	min := math.Max(0, float64(value)-delta)
	max := float64(value) + delta
	return &MarkSpace{Type: t, Min: uint32(min), Max: uint32(max)}
}

func (m *MarkSpace) Matches(evt gopi.LIRCEvent) bool {
	if m.Type != evt.Type() {
		return false
	}
	if m.Min > evt.Value() {
		return false
	}
	if m.Max < evt.Value() {
		return false
	}
	return true
}

/////////////////////////////////////////////////////////////////////
// gopi.InputEvent implementation

type RemoteEvent struct {
	source   Codec
	ts       time.Duration
	scancode uint32
}

func NewRemoteEvent(source Codec, ts time.Duration, scancode uint32) *RemoteEvent {
	return &RemoteEvent{
		source:   source,
		ts:       ts,
		scancode: scancode,
	}
}

func (this *RemoteEvent) Source() gopi.Driver {
	return this.source
}

func (*RemoteEvent) Name() string {
	return "RemoteEvemt"
}

func (this *RemoteEvent) Timestamp() time.Duration {
	return this.ts
}

func (*RemoteEvent) DeviceType() gopi.InputDeviceType {
	return gopi.INPUT_TYPE_KEYBOARD
}

func (*RemoteEvent) EventType() gopi.InputEventType {
	return gopi.INPUT_EVENT_KEYPRESS
}

func (this *RemoteEvent) Scancode() uint32 {
	return this.scancode
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
	return fmt.Sprintf("remotes.RemoteEvent{ scancode=0x%X source=%v ts=%v }", this.scancode, this.source, this.ts)
}
