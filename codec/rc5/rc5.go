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
	value       uint64
	length      uint
}

type state uint32

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// state
	STATE_EXPECT_PULSE state = iota
	STATE_EXPECT_SPACE
)

const (
	TOLERANCE = 25 // 25% tolerance on values
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

	// Set bit length as 14 bits
	this.bit_length = 14

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

func (this *codec) Reset() {
	this.state = STATE_EXPECT_PULSE
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
	this.log.Debug("<remotes.Codec.RC5.Receive>{ type=%v state=%v evt=%v }", this.codec_type, this.state, evt)
	switch this.state {
	case STATE_EXPECT_PULSE:
		if LONG_PULSE.Matches(evt) {
			this.eject(1, 2)
			this.state = STATE_EXPECT_SPACE
		} else if SHORT_PULSE.Matches(evt) {
			this.value = (this.value | 0x1) << 1
			this.length = this.length + 1
			this.state = STATE_EXPECT_SPACE
			this.eject(false)
		} else {
			this.Reset()
		}
	case STATE_EXPECT_SPACE:
		if LONG_SPACE.Matches(evt) {
			this.value = this.value << 2
			this.length = this.length + 2
			this.state = STATE_EXPECT_PULSE
			this.eject(false)
		} else if SHORT_SPACE.Matches(evt) {
			this.value = this.value << 1
			this.length = this.length + 1
			this.state = STATE_EXPECT_PULSE
			this.eject(false)
		} else if REPEAT_SPACE.Matches(evt) {
			this.eject(true)
			this.Reset()
		} else {
			this.Reset()
		}
	default:
		this.Reset()
	}
}

func (this *codec) eject(repeat bool) {
	// Make even
	if this.length%2 != 0 {
		this.length += 1
	}
	fmt.Printf("binary=%v length=%v repeat=%v\n", strconv.FormatUint(this.value, 2), this.length, repeat)
	value := this.value
	for i := uint(0); i < this.length; i += 2 {
		fmt.Printf("value=%v\n", value&0x3)
		value >>= 2
	}
}

////////////////////////////////////////////////////////////////////////////////
// SENDING

func (this *codec) Send(value uint32, repeats uint) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	return 0, 0, gopi.ErrBadParameter
}
