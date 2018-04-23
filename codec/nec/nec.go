/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package nec

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	evt "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"
	appletv "github.com/djthorpe/remotes/codec/appletv"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// NEC Configuration - NEC32 is supported
type Codec struct {
	LIRC gopi.LIRC
	Type remotes.CodecType
}

type codec struct {
	log         gopi.Logger
	lirc        gopi.LIRC
	codec_type  remotes.CodecType
	bit_length  uint
	cancel      context.CancelFunc
	done        chan struct{}
	events      <-chan gopi.Event
	subscribers *evt.PubSub
	state       state
	value       uint32
	length      uint
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
	STATE_EXPECT_REPEAT_SPACE
	STATE_EXPECT_REPEAT_PULSE
	STATE_EXPECT_REPEAT_SPACE2
	STATE_EXPECT_END
)

const (
	TOLERANCE = 25 // 25% tolerance on values
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 9000, TOLERANCE) // 9ms
	HEADER_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 4500, TOLERANCE) // 4.5ms
	BIT_PULSE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 562, TOLERANCE)  // 650ns
	ONE_SPACE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 1688, TOLERANCE) // 1.6ms
	ZERO_SPACE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 562, TOLERANCE)  // 500ns
	TRAIL_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 562, TOLERANCE)  // 650ns

	REPEAT_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 35000, TOLERANCE)  // 35ms
	REPEAT_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 9000, TOLERANCE)   // 9ms
	REPEAT_SPACE2 = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 2250, TOLERANCE)   // 2.25ms
	REPEAT_SPACE3 = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 250000, TOLERANCE) // 250ms
	REPEAT_SPACE4 = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 90000, TOLERANCE)  // 90ms
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.Codec.NEC.Open>{ lirc=%v type=%v }", config.LIRC, config.Type)

	// Check for LIRC
	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)

	// Set log and lirc objects
	this.log = log
	this.lirc = config.LIRC

	// Set codec and bit length
	if bit_length := bitLengthForCodec(config.Type); bit_length == 0 {
		return nil, gopi.ErrBadParameter
	} else {
		this.bit_length = bit_length
		this.codec_type = config.Type
	}

	// Set up channels
	this.done = make(chan struct{})
	this.events = this.lirc.Subscribe()
	this.subscribers = evt.NewPubSub(0)

	// Reset
	this.Reset()

	// Create background routine
	if ctx, cancel := context.WithCancel(context.Background()); ctx != nil {
		this.cancel = cancel
		go this.acceptEvents(ctx)
	}

	// Return success
	return this, nil

}

func (this *codec) Close() error {
	this.log.Debug("<remotes.Codec.NEC.Close>{ type=%v }", this.codec_type)

	// Unsubscribe from LIRC signals
	this.lirc.Unsubscribe(this.events)

	// Cancel background thread, wait for done signal
	this.cancel()
	_ = <-this.done

	// Remove subscribers to this codec
	this.subscribers.Close()

	// Blank out member variables
	close(this.done)
	this.events = nil
	this.subscribers = nil
	this.lirc = nil
	this.done = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *codec) String() string {
	return fmt.Sprintf("<remotes.Codec.NEC>{ type=%v bit_length=%v }", this.codec_type, this.bit_length)
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
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISHER INTERFACE

func (this *codec) Subscribe() <-chan gopi.Event {
	return this.subscribers.Subscribe()
}

func (this *codec) Unsubscribe(subscriber <-chan gopi.Event) {
	this.subscribers.Unsubscribe(subscriber)
}

func (this *codec) Emit(value uint32, repeat bool) {
	this.log.Debug("<remotes.Codec.NEC.Receive>Emit{ value=0x%08X repeat=%v }", value, repeat)
	if scancode, device, err := codeForCodec(this.codec_type, value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.subscribers.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
}

////////////////////////////////////////////////////////////////////////////////
// STATE MACHINE

func (this *codec) acceptEvents(ctx context.Context) {
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-this.events:
			if evt != nil {
				this.receive(evt.(gopi.LIRCEvent))
			}
		}
	}
	this.done <- gopi.DONE
}

func (this *codec) receive(evt gopi.LIRCEvent) {
	this.log.Debug2("<remotes.Codec.NEC.Receive>{ evt=%v }", evt)
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

		// Advance state and emit scancode
		if this.length == this.bit_length {
			this.Emit(this.value, false)
			this.state = STATE_EXPECT_END
		} else if this.length > 0 {
			this.state = STATE_EXPECT_PULSE
		}
	case STATE_EXPECT_END:
		// Mark the end of transmission
		if BIT_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_REPEAT_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT_SPACE:
		if REPEAT_SPACE3.LessThan(evt) {
			// Not a repeat code, it's the start of a new cycle
			this.Reset()
		} else if REPEAT_SPACE.GreaterThan(evt) {
			// It's a repeat code
			this.state = STATE_EXPECT_REPEAT_PULSE
		} else if REPEAT_SPACE4.Matches(evt) {
			// It's a repeat code
			this.state = STATE_EXPECT_REPEAT_PULSE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT_PULSE:
		if REPEAT_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_REPEAT_SPACE2
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT_SPACE2:
		if REPEAT_SPACE2.Matches(evt) {
			// Emit a repeated code
			if this.length > 0 {
				this.Emit(this.value, true)
			}
			this.state = STATE_EXPECT_END
		} else {
			this.Reset()
		}
	default:
		this.Reset()
	}
}

////////////////////////////////////////////////////////////////////////////////
// SENDING

func (this *codec) Send(device uint32, scancode uint32, repeats uint) error {
	this.log.Debug("<remotes.Codec.NEC>Send{ codec_type=%v device=0x%08X scancode=0x%08X repeats=%v }", this.codec_type, device, scancode, repeats)

	if device&0xFFFFFF00 != 0 {
		this.log.Error("<remotes.Codec.NEC> Send: Invalid device parameter")
		return gopi.ErrBadParameter
	}
	if scancode&0xFFFFFF00 != 0 {
		this.log.Error("<remotes.Codec.NEC> Send: Invalid scancode parameter")
		return gopi.ErrBadParameter
	}

	pulses := make([]uint32, 0, 100)

	// 9ms leading pulse burst and 4.5ms space
	pulses = append(pulses, HEADER_PULSE.Value, HEADER_SPACE.Value)

	// device and scancode
	pulses = this.sendbyte(pulses, uint8(device))
	pulses = this.sendbyte(pulses, uint8(device^0xFF))
	pulses = this.sendbyte(pulses, uint8(scancode))
	pulses = this.sendbyte(pulses, uint8(scancode^0xFF))

	// A final 562.5µs pulse
	pulses = append(pulses, TRAIL_PULSE.Value)

	// If there is one or more repeats, then do these
	if repeats > 0 {
		pulses = append(pulses, REPEAT_SPACE.Value)
		for i := uint(0); i < repeats; i++ {
			pulses = append(pulses, REPEAT_PULSE.Value, REPEAT_SPACE2.Value)
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

func bitLengthForCodec(codec remotes.CodecType) uint {
	switch codec {
	case remotes.CODEC_NEC32:
		return 32
	default:
		return 0
	}
}

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	/* If we receive an AppleTV code, then return badparameter */
	if (value & 0xFFFF0000 >> 16) == appletv.APPLETV_CODE {
		return 0, 0, gopi.ErrBadParameter
	}

	/* Or else deal with non-AppleTV */
	switch codec {
	case remotes.CODEC_NEC32:
		value2 := value ^ 0x00FF00FF
		if (value2 & 0x00FF00FF) != (value & 0xFF00FF00 >> 8) {
			return 0, 0, fmt.Errorf("Invalid scancode or device 0x%08X for codec %v", value, codec)
		}
		scan := value & 0x0000FF00 >> 8
		device := value & 0xFF000000 >> 24
		return scan, device, nil
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
	case STATE_EXPECT_END:
		return "STATE_EXPECT_END"
	case STATE_EXPECT_REPEAT_SPACE:
		return "STATE_EXPECT_REPEAT_SPACE"
	case STATE_EXPECT_REPEAT_SPACE2:
		return "STATE_EXPECT_REPEAT_SPACE2"
	case STATE_EXPECT_REPEAT_PULSE:
		return "STATE_EXPECT_REPEAT_PULSE"
	default:
		return "[?? Invalid state]"
	}
}
