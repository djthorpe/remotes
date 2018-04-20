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

/////////////////////////////////////////////////////////////////////
// Mark Space implementation

// A mark or space value
type MarkSpace struct {
	Type            gopi.LIRCType
	Value, Min, Max uint32
}

func NewMarkSpace(t gopi.LIRCType, value, tolerance uint32) *MarkSpace {
	this := &MarkSpace{Type: t}
	this.Set(value, tolerance)
	return this
}

func (m *MarkSpace) Set(value, tolerance uint32) {
	delta := float64(value) * float64(tolerance) / 100.0
	m.Min = uint32(math.Max(0, float64(value)-delta))
	m.Max = uint32(float64(value) + delta)
	m.Value = value
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

func (m *MarkSpace) GreaterThan(evt gopi.LIRCEvent) bool {
	if m.Type != evt.Type() {
		return false
	}
	if evt.Value() > m.Max {
		return false
	}
	return true
}

func (m *MarkSpace) LessThan(evt gopi.LIRCEvent) bool {
	if m.Type != evt.Type() {
		return false
	}
	if evt.Value() < m.Min {
		return false
	}
	return true
}
