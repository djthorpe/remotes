/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package nec

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

// NEC Configuration - NEC32 is supported
type Codec struct {
	LIRC gopi.LIRC
	Type remotes.CodecType
}

type codec struct {
	log        gopi.Logger
	lirc       gopi.LIRC
	codec_type remotes.CodecType
	bit_length uint
	state      state
	value      uint32
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
	STATE_EXPECT_END_PULSE
	STATE_EXPECT_TRAIL_SPACE_17500
	STATE_EXPECT_TRAIL_SPACE_35000
	STATE_EXPECT_REPEAT_SPACE
	STATE_EXPECT_REPEAT_PULSE
)

const (
	TOLERANCE    = 35 // 35% tolerance on values
	APPLETV_CODE = 0x77E1
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE      = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 9000, TOLERANCE) // 9ms
	HEADER_SPACE      = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 4500, TOLERANCE) // 4.5ms
	BIT_PULSE         = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 562, TOLERANCE)  // 650ns
	ONE_SPACE         = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 1688, TOLERANCE) // 1.6ms
	ZERO_SPACE        = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 562, TOLERANCE)
	TRAIL_PULSE       = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 562, TOLERANCE)
	TRAIL_SPACE_17500 = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 17500, TOLERANCE) // 17.5ms
	TRAIL_SPACE_35000 = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 35000, TOLERANCE) // 35ms
	REPEAT_PULSE      = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 9000, TOLERANCE)  // 9ms
	REPEAT_SPACE      = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 2500, TOLERANCE)
	REPEAT_SPACE2     = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 96577, TOLERANCE)
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.codec.nec>Open{ lirc=%v type=%v }", config.LIRC, config.Type)

	// Check for LIRC
	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)
	this.log = log
	this.lirc = config.LIRC

	// Set codec and bit length
	if bit_length := bitLengthForCodec(config.Type); bit_length == 0 {
		return nil, gopi.ErrBadParameter
	} else {
		this.bit_length = bit_length
		this.codec_type = config.Type
	}

	// Reset state
	this.Reset()

	// Backround tasks
	this.Tasks.Start(this.PulseTask)

	// Return success
	return this, nil
}

func (this *codec) Close() error {
	this.log.Debug("<remotes.codec.nec>Close>{ type=%v }", this.codec_type)

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
	return fmt.Sprintf("<remotes.codec.nec>{ type=%v bit_length=%v }", this.codec_type, this.bit_length)
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
	this.log.Debug2("<remotes.Codec.NEC>Send{ codec_type=%v device=0x%08X scancode=0x%08X repeats=%v }", this.codec_type, device, scancode, repeats)

	// 9ms leading pulse burst and 4.5ms space
	pulses := make([]uint32, 0, 100)
	pulses = append(pulses, HEADER_PULSE.Value, HEADER_SPACE.Value)

	switch this.codec_type {
	case remotes.CODEC_NEC32:
		// Ensure the device is 16 bits and the scancode is 8 bits
		if uint32(uint16(device)) != device {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid device parameter")
			return gopi.ErrBadParameter
		}
		if uint32(uint8(scancode)) != scancode {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid scancode parameter")
			return gopi.ErrBadParameter
		}
		// Emit the device and scancode
		pulses = this.sendbyte(pulses, uint8((device&0xFF00)>>8))
		pulses = this.sendbyte(pulses, uint8(device&0x00FF))
		pulses = this.sendbyte(pulses, uint8(scancode&0x00FF))
		pulses = this.sendbyte(pulses, uint8(scancode^0xFF))
	case remotes.CODEC_NEC16:
		// Ensure the device is 8 bits and the scancode is 8 bits
		if uint32(uint8(device)) != device {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid device parameter")
			return gopi.ErrBadParameter
		}
		if uint32(uint8(scancode)) != scancode {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid scancode parameter")
			return gopi.ErrBadParameter
		}
		// Emit the device and scancode
		pulses = this.sendbyte(pulses, uint8(device&0x00FF))
		pulses = this.sendbyte(pulses, uint8(scancode&0x00FF))
	case remotes.CODEC_APPLETV:
		// Ensure device code is 8 bits and scancode is 8 bits
		if uint32(uint8(device)) != device {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid device parameter")
			return gopi.ErrBadParameter
		}
		if uint32(uint8(scancode)) != scancode {
			this.log.Error("<remotes.Codec.NEC> Send: Invalid scancode parameter")
			return gopi.ErrBadParameter
		}
		// Emit the AppleTV code, then the scancode and device
		pulses = this.sendbyte(pulses, uint8(APPLETV_CODE&0xFF00>>8))
		pulses = this.sendbyte(pulses, uint8(APPLETV_CODE&0x00FF))
		pulses = this.sendbyte(pulses, uint8(device&0x00FF))
		pulses = this.sendbyte(pulses, uint8(scancode&0x00FF))
	default:
		return gopi.ErrNotImplemented
	}

	// A final 562.5µs pulse
	pulses = append(pulses, TRAIL_PULSE.Value)

	// If there is one or more repeats, then do these
	if repeats > 0 {
		pulses = append(pulses, TRAIL_SPACE_35000.Value)
		for i := uint(0); i < repeats; i++ {
			pulses = append(pulses, REPEAT_PULSE.Value, REPEAT_SPACE.Value)
		}
		// A final 562.5µs pulse
		pulses = append(pulses, TRAIL_PULSE.Value)
	}

	// Perform the pulse send
	return this.lirc.PulseSend(pulses)
}

func (this *codec) sendbyte(pulses []uint32, value uint8) []uint32 {
	for i := 0; i < 8; i++ {
		pulses = append(pulses, BIT_PULSE.Value)
		if value&0x80 == 0 { // Send zero
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
	this.log.Debug2("<remotes.codec.nec>Receive{ type=%v state=%v evt=%v }", this.codec_type, this.state, evt)
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
		if ZERO_SPACE.Matches(evt) {
			this.value = (this.value << 1) | 0
			this.length = this.length + 1
		} else if ONE_SPACE.Matches(evt) {
			this.value = (this.value << 1) | 1
			this.length = this.length + 1
		} else {
			this.Reset()
		}

		// Advance state to expect the trailing pulse
		if this.length == this.bit_length {
			//this.Emit(this.value, false)
			this.state = STATE_EXPECT_END_PULSE
		} else if this.length > 0 {
			this.state = STATE_EXPECT_PULSE
		}
	case STATE_EXPECT_END_PULSE:
		// Mark the end of transmission
		if BIT_PULSE.Matches(evt) {
			if this.codec_type == remotes.CODEC_NEC16 {
				this.state = STATE_EXPECT_TRAIL_SPACE_17500
			} else {
				// Emit key press for NEC32
				this.emit(this.value, this.repeat)
				this.state = STATE_EXPECT_TRAIL_SPACE_35000
			}
		} else {
			this.Reset()
		}
	case STATE_EXPECT_TRAIL_SPACE_17500:
		if TRAIL_SPACE_17500.Matches(evt) {
			// End of NEC16 code
			this.emit(this.value, this.repeat)
			this.state = STATE_EXPECT_PULSE
			this.value = 0
			this.length = 0
			this.repeat = true
		} else {
			this.Reset()
		}
	case STATE_EXPECT_TRAIL_SPACE_35000:
		if TRAIL_SPACE_35000.Matches(evt) {
			// End of NEC32 code
			this.state = STATE_EXPECT_REPEAT_PULSE
		} else if REPEAT_SPACE2.Matches(evt) {
			// End of NEC32 repeat
			this.state = STATE_EXPECT_REPEAT_PULSE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT_PULSE:
		if REPEAT_PULSE.Matches(evt) {
			this.repeat = true
			this.state = STATE_EXPECT_REPEAT_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT_SPACE:
		if REPEAT_SPACE.Matches(evt) || REPEAT_SPACE2.Matches(evt) {
			this.emit(this.value, this.repeat)
			this.state = STATE_EXPECT_END_PULSE
		} else if HEADER_SPACE.Matches(evt) {
			this.state = STATE_EXPECT_PULSE
			this.value = 0
			this.length = 0
			this.repeat = true
		} else {
			this.Reset()
		}
	default:
		this.Reset()
	}
}

func (this *codec) emit(value uint32, repeat bool) {
	this.log.Debug2("<remotes.codec.nec>Emit{ type=%v value=0x%08X repeat=%v }", this.codec_type, value, repeat)
	if scancode, device, err := codeForCodec(this.codec_type, value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
}

func bitLengthForCodec(codec remotes.CodecType) uint {
	switch codec {
	case remotes.CODEC_NEC32:
		return 32
	case remotes.CODEC_NEC16:
		return 16
	case remotes.CODEC_APPLETV:
		return 32
	default:
		return 0
	}
}

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	switch codec {
	case remotes.CODEC_APPLETV:
		// Check for Apple TV device
		appletv := value & 0xFFFF0000 >> 16
		if appletv != APPLETV_CODE {
			return 0, 0, gopi.ErrBadParameter
		}
		device := value & 0x000000FF
		scancode := value & 0x0000FF00 >> 8
		return scancode, device, nil
	case remotes.CODEC_NEC32:
		// Ignore Apple TV
		device := value & 0xFFFF0000 >> 16
		if device == APPLETV_CODE {
			return 0, 0, gopi.ErrBadParameter
		}
		// Check to make sure scancode and reverse of scancode match
		scancode1 := value & 0x0000FF00 >> 8
		scancode2 := value & 0x000000FF
		if scancode1 != scancode2^0x00FF {
			return 0, 0, fmt.Errorf("Invalid scancode 0x%02X or device 0x%04X for codec %v (the code received was 0x%08X, the scancodes were %02X and %02X)", scancode1, device, codec, value, scancode1, scancode2^0xFF)
		}
		return scancode1, device, nil
	case remotes.CODEC_NEC16:
		// Check to make sure scancode and reverse of scancode match
		scancode := value & 0x000000FF
		device := value & 0x0000FF00 >> 8
		return scancode, device, nil
	default:
		return 0, 0, gopi.ErrBadParameter
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
	case STATE_EXPECT_END_PULSE:
		return "STATE_EXPECT_END_PULSE"
	case STATE_EXPECT_TRAIL_SPACE_17500:
		return "STATE_EXPECT_TRAIL_SPACE_17500"
	case STATE_EXPECT_TRAIL_SPACE_35000:
		return "STATE_EXPECT_TRAIL_SPACE_35000"
	case STATE_EXPECT_REPEAT_SPACE:
		return "STATE_EXPECT_REPEAT_SPACE"
	case STATE_EXPECT_REPEAT_PULSE:
		return "STATE_EXPECT_REPEAT_PULSE"
	default:
		return "[?? Invalid state]"
	}
}
