/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package sony

import (
	"context"
	"fmt"
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
	value       uint32
	duration    uint32
	length      uint
	repeat      bool
}

type state uint32

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// state
	STATE_EXPECT_HEADER_PULSE state = iota
	STATE_EXPECT_HEADER_SPACE
	STATE_EXPECT_BIT
	STATE_EXPECT_SPACE
	STATE_EXPECT_TRAIL
	STATE_EXPECT_REPEAT
)

const (
	TOLERANCE   = 25    // 25% tolerance on values
	TX_DURATION = 45000 // 45ms between each transmission
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 2400, TOLERANCE)
	HEADER_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 575, TOLERANCE)
	ONE_PULSE     = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1200, TOLERANCE)
	ZERO_PULSE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 575, TOLERANCE)
	ONEZERO_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 575, TOLERANCE)
	TRAIL_PULSE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1200, TOLERANCE)
	REPEAT_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, TX_DURATION, TOLERANCE)
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.Codec.Sony.Open>{ lirc=%v type=%v }", config.LIRC, config.Type)

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
	this.log.Debug("<remotes.Codec.Sony.Close>{ type=%v }", this.codec_type)

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
	return fmt.Sprintf("<remotes.Codec.Sony>{ type=%v bit_length=%v }", this.codec_type, this.bit_length)
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
	this.duration = 0
	this.repeat = false
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
	this.log.Debug2("<remotes.Codec.Sony.Receive>{ type=%v evt=%v }", this.codec_type, evt)
	if this.codec_type == remotes.CODEC_SONY12 {
		this.log.Debug("<remotes.Codec.Sony.Receive>{ type=%v evt=%v state=%v }", this.codec_type, evt, this.state)
	}
	switch this.state {
	case STATE_EXPECT_HEADER_PULSE:
		if HEADER_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_HEADER_SPACE
			this.duration += evt.Value()
		} else {
			this.Reset()
		}
	case STATE_EXPECT_HEADER_SPACE:
		if HEADER_SPACE.Matches(evt) {
			this.state = STATE_EXPECT_BIT
			this.duration += evt.Value()
		} else {
			this.Reset()
		}
	case STATE_EXPECT_BIT:
		if ONE_PULSE.Matches(evt) {
			this.value |= 1
			this.state = STATE_EXPECT_SPACE
			this.duration += evt.Value()
		} else if ZERO_PULSE.Matches(evt) {
			this.value |= 0
			this.state = STATE_EXPECT_SPACE
			this.duration += evt.Value()
		} else {
			this.Reset()
		}
	case STATE_EXPECT_SPACE:
		if ONEZERO_SPACE.Matches(evt) {
			this.value = this.value << 1
			this.length = this.length + 1
			if this.length == this.bit_length {
				this.state = STATE_EXPECT_REPEAT
				this.duration += evt.Value()
			} else {
				this.state = STATE_EXPECT_BIT
				this.duration += evt.Value()
			}
		} else {
			this.Reset()
		}
	case STATE_EXPECT_TRAIL:
		if TRAIL_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_REPEAT
			this.duration += evt.Value()
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT:
		// Time for repeat varies
		REPEAT_SPACE.Set(TX_DURATION-this.duration, TOLERANCE)
		if REPEAT_SPACE.Matches(evt) {
			this.Emit(this.value, this.repeat)
			this.repeat = true
			this.state = STATE_EXPECT_HEADER_PULSE
			this.value = 0
			this.length = 0
			this.duration = 0
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
	return gopi.ErrNotImplemented
	/*
		this.log.Debug("<remotes.Codec.Sony.Send>{ type=%v value=%X repeats=%v }", this.codec_type, value, repeats)

		// Check to make sure the scancode value is less than
		// or equal to the length and repeats is at least one
		mask := uint32(1<<this.bit_length) - 1
		if value&mask != value || repeats == 0 {
			return gopi.ErrBadParameter
		}

		// Create a container for mark/space
		pulses := make([]uint32, 0)
		pulses = append(pulses, HEADER_PULSE.Value)

		for r := uint(0); r < repeats; r++ {
			mask := uint32(1) << (this.bit_length - 1)
			pulses = append(pulses, HEADER_SPACE.Value)
			for i := uint(0); i < this.bit_length; i++ {
				if value&mask != 0 {
					pulses = append(pulses, ONE_PULSE.Value)
				} else {
					pulses = append(pulses, ZERO_PULSE.Value)
				}
				pulses = append(pulses, ONEZERO_SPACE.Value)
				mask = mask >> 1
			}
			pulses = append(pulses, TRAIL_PULSE.Value)
			if r+1 < repeats {
				pulses = append(pulses, REPEAT_SPACE.Value, HEADER_PULSE.Value)
			}
		}

		// Debug
		if this.log.IsDebug() {
			for i, value := range pulses {
				if i%2 == 0 {
					fmt.Println(" mark", value)
				} else {
					fmt.Println("space", value)
				}
			}
		}

		// Perform the sending
		return this.lirc.PulseSend(pulses)*/
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func bitLengthForCodec(codec remotes.CodecType) uint {
	switch codec {
	case remotes.CODEC_SONY12:
		return 12
	case remotes.CODEC_SONY15:
		return 15
	case remotes.CODEC_SONY20:
		return 20
	default:
		return 0
	}
}

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	switch codec {
	case remotes.CODEC_SONY12:
		// 7 scancode bits and 5 device bits
		return (value & 0xFE0) >> 5, (value & 0x01F), nil
	case remotes.CODEC_SONY15:
		// 15 bit codes are similar, with 7 command bits and 8 device bits
		return (value & 0x7F00) >> 8, (value & 0xFF), nil
	case remotes.CODEC_SONY20:
		// TODO
		// 20 bit codes have 7 command bits, 5 device bits, and 8 extended device bits.
		return value, 0, nil
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
	case STATE_EXPECT_BIT:
		return "STATE_EXPECT_BIT"
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
