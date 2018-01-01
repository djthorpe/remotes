/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"math"

	"github.com/djthorpe/gopi"
)

// A coder/decoder for remote
type Codec interface {
	gopi.Driver

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
