/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package panasonic

import (
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Codec struct {
	LIRC gopi.LIRC
	Type remotes.CodecType
}

type codec struct {
	log        gopi.Logger
	lirc       gopi.LIRC
	codec_type remotes.CodecType
	state      state
	value      uint64
	length     uint
	repeat     bool

	event.Publisher
	event.Tasks
}

type state uint32

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// state
	STATE_EXPECT_HEADER_PULSE state = iota
	STATE_EXPECT_HEADER_SPACE
	STATE_EXPECT_PULSE
	STATE_EXPECT_SPACE
	STATE_EXPECT_TRAIL
	STATE_EXPECT_REPEAT
)

const (
	TOLERANCE  = 35 // 35% tolerance on values
	BIT_LENGTH = 48
	PREAMBLE   = 0x4004
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 3500, TOLERANCE)
	HEADER_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 1700, TOLERANCE)
	BIT_PULSE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 450, TOLERANCE)
	ONE_SPACE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 1300, TOLERANCE)
	ZERO_SPACE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 450, TOLERANCE)
	TRAIL_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 450, TOLERANCE)
	REPEAT_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 75000, TOLERANCE)
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.codec.panasonic>Open{ lirc=%v type=%v }", config.LIRC, config.Type)

	// Check for LIRC
	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)
	this.log = log
	this.lirc = config.LIRC
	this.codec_type = config.Type

	// Reset state
	this.Reset()

	// Backround tasks
	this.Tasks.Start(this.PulseTask)

	// Return success
	return this, nil
}

func (this *codec) Close() error {
	this.log.Debug("<remotes.codec.panasonic>Close>{ type=%v }", this.codec_type)

	// Remove subscribers to this codec
	this.Publisher.Close()

	// End tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Release resources
	this.lirc = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *codec) String() string {
	return fmt.Sprintf("<remotes.codec.panasonic>{ type=%v }", this.codec_type)
}

////////////////////////////////////////////////////////////////////////////////
// CODEC INTERFACE

func (this *codec) Type() remotes.CodecType {
	return this.codec_type
}

func (this *codec) Reset() {
	this.state = STATE_EXPECT_HEADER_PULSE
	this.value = 0
	this.length = 0
	this.repeat = false
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASK

func (this *codec) PulseTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	events := this.lirc.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt != nil {
				this.receive(evt.(gopi.LIRCEvent))
			}
		case <-stop:
			break FOR_LOOP

		}
	}

	this.lirc.Unsubscribe(events)

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SENDING

func (this *codec) Send(device uint32, scancode uint32, repeats uint) error {
	this.log.Debug2("<remotes.codec.panasonic>Send{ device=0x%08X scancode=0x%08X repeats=%v }", device, scancode, repeats)

	// Header Pulse of 3.5ms, then space of 1.7ms
	pulses := make([]uint32, 0, 100)
	pulses = append(pulses, HEADER_PULSE.Value, HEADER_SPACE.Value)

	for i := uint(0); i < (repeats + 1); i++ {
		// Send preamble
		pulses = this.sendbyte(pulses, byte(PREAMBLE>>8))
		pulses = this.sendbyte(pulses, byte(PREAMBLE&0xFF))

		// Send device
		if device&0xFFFF0000 != 0 {
			this.log.Debug("remotes.codec.panasonic: Send: Invalid device parameter")
			return gopi.ErrBadParameter
		}
		pulses = this.sendbyte(pulses, byte(device>>8))
		pulses = this.sendbyte(pulses, byte(device&0xFF))

		// Send scancode
		if scancode&0xFFFFFF00 != 0 {
			this.log.Debug("remotes.codec.panasonic: Send: Invalid scancode parameter")
			return gopi.ErrBadParameter
		}
		pulses = this.sendbyte(pulses, byte(scancode))

		// Send checksum
		ck := byte(device>>8) ^ byte(device&0xFF) ^ byte(scancode)
		pulses = this.sendbyte(pulses, byte(ck))

		// Send trail pulse
		pulses = append(pulses, TRAIL_PULSE.Value)

		// If repeats > 0 then send repeats space
		if repeats > 0 && i < repeats {
			pulses = append(pulses, REPEAT_SPACE.Value)
		}
	}
	return this.lirc.PulseSend(pulses)
}

func (this *codec) sendbyte(pulses []uint32, value uint8) []uint32 {
	for i := 0; i < 8; i++ {
		pulses = append(pulses, BIT_PULSE.Value)
		if value&0x80 == 0 {
			// Send zero
			pulses = append(pulses, ZERO_SPACE.Value)
		} else {
			// Send one
			pulses = append(pulses, ONE_SPACE.Value)
		}
		value <<= 1
	}
	return pulses
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *codec) receive(evt gopi.LIRCEvent) {
	this.log.Debug2("<remotes.codec.panasonic>Receive>{ evt=%v }", evt)
	switch this.state {
	case STATE_EXPECT_HEADER_PULSE:
		if HEADER_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_HEADER_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_HEADER_SPACE:
		if HEADER_SPACE.Matches(evt) {
			this.state = STATE_EXPECT_PULSE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_PULSE:
		if BIT_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_SPACE:
		// Register a zero or one
		this.value <<= 1
		if ZERO_SPACE.Matches(evt) {
			this.value |= 0
			this.length = this.length + 1
		} else if ONE_SPACE.Matches(evt) {
			this.value |= 1
			this.length = this.length + 1
		} else {
			this.Reset()
		}

		// Advance state if the correct length
		if this.length == BIT_LENGTH {
			this.state = STATE_EXPECT_TRAIL
		} else if this.length > 0 {
			this.state = STATE_EXPECT_PULSE
		}
	case STATE_EXPECT_TRAIL:
		if TRAIL_PULSE.Matches(evt) {
			this.emit(this.value, this.repeat)
			this.state = STATE_EXPECT_REPEAT
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT:
		if REPEAT_SPACE.Matches(evt) {
			this.repeat = true
			this.state = STATE_EXPECT_HEADER_PULSE
			this.value = 0
			this.length = 0
		} else {
			this.Reset()
		}
	default:
		this.Reset()
	}
}

func (this *codec) emit(value uint64, repeat bool) {
	if scancode, device, err := codeForCodec(value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
}

func codeForCodec(value uint64) (uint32, uint32, error) {
	preamble := value & 0xFFFF00000000
	device := value & 0x0000FF000000 >> 24
	subdevice := value & 0x000000FF0000 >> 16
	scancode := value & 0x00000000FF00 >> 8
	ck1 := value & 0x0000000000FF
	ck2 := device ^ subdevice ^ scancode
	if (preamble >> 32) != PREAMBLE {
		// Bad preamble
		return 0, 0, gopi.ErrBadParameter
	} else if ck1 != ck2 {
		// Bad checksum
		return 0, 0, gopi.ErrBadParameter
	} else {
		// Merge device together with subdevice
		return uint32(scancode), uint32(device<<8 | subdevice), nil
	}
}

func (s state) String() string {
	switch s {
	case STATE_EXPECT_HEADER_PULSE:
		return "STATE_EXPECT_HEADER_PULSE"
	case STATE_EXPECT_HEADER_SPACE:
		return "STATE_EXPECT_HEADER_SPACE"
	case STATE_EXPECT_PULSE:
		return "STATE_EXPECT_PULSE"
	case STATE_EXPECT_SPACE:
		return "STATE_EXPECT_SPACE"
	case STATE_EXPECT_TRAIL:
		return "STATE_EXPECT_TRAIL"
	case STATE_EXPECT_REPEAT:
		return "STATE_EXPECT_REPEAT"
	default:
		return "[?? Invalid state]"
	}
}
