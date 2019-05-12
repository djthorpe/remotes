/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package sony

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

// Sony Configuration - for 12, 15 and 20
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
	duration   uint32
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
	STATE_EXPECT_BIT
	STATE_EXPECT_SPACE
	STATE_EXPECT_TRAIL
	STATE_EXPECT_REPEAT
)

const (
	TOLERANCE   = 35    // 35% tolerance on values
	TX_DURATION = 45000 // 45ms between each transmission
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 2400, TOLERANCE)
	ONEZERO_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 575, TOLERANCE)
	ONE_PULSE     = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1200, TOLERANCE)
	ZERO_PULSE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 575, TOLERANCE)
	TRAIL_PULSE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1200, TOLERANCE)
	REPEAT_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, TX_DURATION, TOLERANCE)
)

var (
	timestamp = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.codec.sony>Open{ lirc=%v type=%v }", config.LIRC, config.Type)

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
	this.log.Debug("<remotes.codec.sony>Close>{ type=%v }", this.codec_type)

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

func (this *codec) Emit(value uint32, repeat bool) {
	if scancode, device, err := codeForCodec(this.codec_type, value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.Publisher.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
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

func (this *codec) receive(evt gopi.LIRCEvent) {
	this.log.Debug2("<remotes.codec.sony>Receive{ type=%v evt=%v }", this.codec_type, evt)
	switch this.state {
	case STATE_EXPECT_HEADER_PULSE:
		if HEADER_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_SPACE
			this.duration += evt.Value()
		} else {
			this.Reset()
		}
	case STATE_EXPECT_SPACE:
		if ONEZERO_SPACE.Matches(evt) {
			this.value <<= 1
			this.state = STATE_EXPECT_BIT
			this.duration += evt.Value()
		} else {
			REPEAT_SPACE.Set(TX_DURATION-this.duration, TOLERANCE)
			if REPEAT_SPACE.Matches(evt) && this.length == this.bit_length {
				this.Emit(this.value, this.repeat)
				this.value = 0
				this.length = 0
				this.duration = 0
				this.repeat = true
				this.state = STATE_EXPECT_HEADER_PULSE
			} else {
				this.Reset()
			}
		}
	case STATE_EXPECT_BIT:
		if ONE_PULSE.Matches(evt) {
			this.value |= 1
			this.length += 1
			this.state = STATE_EXPECT_SPACE
			this.duration += evt.Value()
		} else if ZERO_PULSE.Matches(evt) {
			this.value |= 0
			this.length += 1
			this.state = STATE_EXPECT_SPACE
			this.duration += evt.Value()
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
	this.log.Debug2("<remotes.codec.sony>Send{ codec_type=%v device=0x%08X scancode=0x%08X repeats=%v }", this.codec_type, device, scancode, repeats)

	// Array of pulses
	pulses := make([]uint32, 0, 100)

	// Make bits and pulses
	if bits, err := bitsForCodec(this.codec_type, device, scancode); err != nil {
		return err
	} else {
		for j := uint(0); j < (repeats + 1); j++ {
			length := HEADER_PULSE.Value
			pulses = append(pulses, HEADER_PULSE.Value)

			// Send the bits
			for i := 0; i < len(bits); i++ {
				pulses = append(pulses, ONEZERO_SPACE.Value)
				length += ONEZERO_SPACE.Value
				if bits[i] {
					pulses = append(pulses, ONE_PULSE.Value)
					length += ONEZERO_SPACE.Value
				} else {
					pulses = append(pulses, ZERO_PULSE.Value)
					length += ZERO_PULSE.Value
				}
			}

			// If repeats then send trail
			if repeats > 0 && j < repeats {
				pulses = append(pulses, TX_DURATION-length)
			}
		}
	}

	// Perform the sending
	return this.lirc.PulseSend(pulses)
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
		return (value & 0x0FE0) >> 5, (value & 0x001F), nil
	case remotes.CODEC_SONY15:
		// 15 bit codes are similar, with 7 command bits and 8 device bits
		return (value & 0x7F00) >> 8, (value & 0xFF), nil
	case remotes.CODEC_SONY20:
		// 20 bit codes have 7 command bits and 13 device bits
		return (value & 0xFE000) >> 13, (value & 0x1FFF), nil
	default:
		return 0, 0, gopi.ErrBadParameter
	}
}

func bitsForCodec(codec remotes.CodecType, device uint32, scancode uint32) ([]bool, error) {
	bits := make([]bool, 0, bitLengthForCodec(codec))
	switch codec {
	case remotes.CODEC_SONY12:
		// 7 scancode bits and 5 device bits
		bits = bitsAppend(bits, scancode, 7)
		bits = bitsAppend(bits, device, 5)
	case remotes.CODEC_SONY15:
		// 7 scancode bits and 8 device bits
		bits = bitsAppend(bits, scancode, 7)
		bits = bitsAppend(bits, device, 8)
	case remotes.CODEC_SONY20:
		// 7 scancode bits and 13 device bits
		bits = bitsAppend(bits, scancode, 7)
		bits = bitsAppend(bits, device, 13)
	default:
		return nil, gopi.ErrBadParameter
	}
	return bits, nil
}

func bitsAppend(array []bool, value uint32, length uint) []bool {
	mask := uint32(1) << (length - 1)
	for i := uint(0); i < length; i++ {
		array = append(array, value&mask != 0)
		mask >>= 1
	}
	return array
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
