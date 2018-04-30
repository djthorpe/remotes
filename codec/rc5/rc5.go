/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package rc5

import (
	"context"
	"fmt"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	evt "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Sony Configuration - for 12, 15 and 20
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
	bits        []bool
	length      uint
	repeat      bool
}

type state uint32

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// state
	STATE_EXPECT_FIRST_PULSE state = iota
	STATE_EXPECT_PULSE
	STATE_EXPECT_SPACE
)

const (
	TOLERANCE = 35 // 35% tolerance on values
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	LONG_PULSE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1778, TOLERANCE)
	LONG_SPACE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 1778, TOLERANCE)
	SHORT_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 889, TOLERANCE)
	SHORT_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 889, TOLERANCE)
	REPEAT_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 90000, TOLERANCE)
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.Codec.RC5.Open>{ lirc=%v type=%v }", config.LIRC, config.Type)

	// Check for LIRC
	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)

	// Set log and lirc objects
	this.log = log
	this.lirc = config.LIRC

	// Set up channels
	this.done = make(chan struct{})
	this.events = this.lirc.Subscribe()
	this.subscribers = evt.NewPubSub(0)

	// Set bit length to 14 bits
	this.bit_length = 14
	this.codec_type = remotes.CODEC_RC5

	// Reset
	this.Reset(false)

	// Create background routine
	if ctx, cancel := context.WithCancel(context.Background()); ctx != nil {
		this.cancel = cancel
		go this.acceptEvents(ctx)
	}

	// Return success
	return this, nil
}

func (this *codec) Close() error {
	this.log.Debug("<remotes.Codec.RC5.Close>{ type=%v }", this.codec_type)

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
	return fmt.Sprintf("<remotes.Codec.RC5>{ type=%v }", this.codec_type)
}

////////////////////////////////////////////////////////////////////////////////
// CODEC INTERFACE

func (this *codec) Type() remotes.CodecType {
	return this.codec_type
}

func (this *codec) Reset(repeat bool) {
	this.state = STATE_EXPECT_FIRST_PULSE
	this.bits = make([]bool, 0, this.bit_length*2)
	this.length = 0
	this.repeat = repeat
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
	if scancode, device, err := codeForCodec(this.codec_type, value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.subscribers.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
}

////////////////////////////////////////////////////////////////////////////////
// RECEIVING STATE MACHINE

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
	this.log.Debug2("<remotes.Codec.RC5.Receive>{ type=%v state=%v evt=%v }", this.codec_type, this.state, evt)
	fmt.Println(this.state, evt)
	switch this.state {
	case STATE_EXPECT_FIRST_PULSE:
		if LONG_PULSE.Matches(evt) {
			this.eject(false, true, true)
			this.state = STATE_EXPECT_SPACE
		} else if SHORT_PULSE.Matches(evt) {
			this.eject(false, true)
			this.state = STATE_EXPECT_SPACE
		} else {
			this.Reset(false)
		}
	case STATE_EXPECT_PULSE:
		if LONG_PULSE.Matches(evt) {
			this.eject(true, true)
			this.state = STATE_EXPECT_SPACE
		} else if SHORT_PULSE.Matches(evt) {
			this.eject(true)
			this.state = STATE_EXPECT_SPACE
		} else {
			this.Reset(false)
		}
	case STATE_EXPECT_SPACE:
		if LONG_SPACE.Matches(evt) {
			this.eject(false, false)
			this.state = STATE_EXPECT_PULSE
		} else if SHORT_SPACE.Matches(evt) {
			this.eject(false)
			this.state = STATE_EXPECT_PULSE
		} else if REPEAT_SPACE.Matches(evt) {
			this.Reset(true)
		} else {
			this.Reset(false)
		}
	default:
		this.Reset(false)
	}
}

func (this *codec) eject(bits ...bool) {
	this.bits = append(this.bits, bits...)
	this.length += uint(len(bits))

	if uint(len(this.bits)) == this.bit_length*2 {
		value := uint32(0)
		for i, j := uint(0), uint(0); i < this.bit_length; i, j = i+1, j+2 {
			value <<= 1
			if this.bits[j] == this.bits[j+1] {
				// Invalid Manchester Code
				return
			} else if this.bits[j] {
				// 10 => 0
				value |= 0
			} else {
				// 01 => 1
				value |= 1
			}
		}
		this.Emit(value, this.repeat)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SENDING

func (this *codec) Send(device uint32, scancode uint32, repeats uint) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	switch codec {
	case remotes.CODEC_RC5:
		// scancode is lowest 6 bits (0x03F), device is next 5 bits (7C0)
		scancode := value & 0x003F
		device := value & 0x07C0 >> 6
		header := value & 0x3800 >> 11
		fmt.Printf("value=%v header=%v scancode=%X device=%X\n", strconv.FormatUint(uint64(value), 2), strconv.FormatUint(uint64(header), 2), scancode, device)
		// TODO: Check header
		return scancode, device, nil
	default:
		return 0, 0, gopi.ErrNotImplemented
	}
}

func (s state) String() string {
	switch s {
	case STATE_EXPECT_FIRST_PULSE:
		return "STATE_EXPECT_FIRST_PULSE"
	case STATE_EXPECT_PULSE:
		return "STATE_EXPECT_PULSE"
	case STATE_EXPECT_SPACE:
		return "STATE_EXPECT_SPACE"
	default:
		return "[?? Invalid state]"
	}
}
