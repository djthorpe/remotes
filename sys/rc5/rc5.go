/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package rc5

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

// RC5 Configuration
type Codec struct {
	LIRC gopi.LIRC
	Type remotes.CodecType
}

type codec struct {
	log        gopi.Logger
	lirc       gopi.LIRC
	codec_type remotes.CodecType
	bit_length uint

	// State
	state  state
	bits   []bool
	repeat bool

	event.Publisher
	event.Tasks
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
	LONG_TIMEOUT = remotes.NewMarkSpace(gopi.LIRC_TYPE_TIMEOUT, 1778, TOLERANCE)
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
	log.Debug("<remotes.codec.rc5>Open{ lirc=%v type=%v }", config.LIRC, config.Type)

	// Check for LIRC
	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)

	// Set log and lirc objects
	this.log = log
	this.lirc = config.LIRC

	// Set codec type
	this.codec_type = config.Type

	// Set bit length to 14 bits
	this.bit_length = 14

	// Reset
	this.Reset(false)

	// Backround tasks
	this.Tasks.Start(this.PulseTask)

	// Return success
	return this, nil
}

func (this *codec) Close() error {
	this.log.Debug("<remotes.codec.rc5>Close>{ type=%v }", this.codec_type)

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
	return fmt.Sprintf("<remotes.codec.rc5>{ type=%v }", this.codec_type)
}

////////////////////////////////////////////////////////////////////////////////
// CODEC INTERFACE

func (this *codec) Type() remotes.CodecType {
	return this.codec_type
}

func (this *codec) Reset(repeat bool) {
	this.log.Debug2("<remotes.codec.rc5>Reset{ repeat=%v }", repeat)
	this.state = STATE_EXPECT_FIRST_PULSE
	this.bits = make([]bool, 0, this.bit_length*2)
	this.repeat = repeat
}

////////////////////////////////////////////////////////////////////////////////
// SENDING

func (this *codec) Send(device uint32, scancode uint32, repeats uint) error {
	return gopi.ErrNotImplemented
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

	// Unsubscribe
	this.lirc.Unsubscribe(events)

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *codec) receive(evt gopi.LIRCEvent) {
	this.log.Debug("<remotes.codec.rc5>Receive{ type=%v state=%v type=%v value=%v }", this.codec_type, this.state, evt.Type(), evt.Value())
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
	case STATE_EXPECT_SPACE:
		if SHORT_SPACE.Matches(evt) {
			this.eject(false)
			this.state = STATE_EXPECT_PULSE
		} else if LONG_SPACE.Matches(evt) {
			this.eject(false, false)
			this.state = STATE_EXPECT_PULSE
		} else if LONG_TIMEOUT.GreaterThan(evt) {
			this.eject(false)
			this.state = STATE_EXPECT_FIRST_PULSE
		} else if REPEAT_SPACE.Matches(evt) {
			this.eject(false)
			this.Reset(true)
		} else {
			this.Reset(false)
		}
	case STATE_EXPECT_PULSE:
		if SHORT_PULSE.Matches(evt) {
			this.eject(true)
			this.state = STATE_EXPECT_SPACE
		} else if LONG_PULSE.Matches(evt) {
			this.eject(true, true)
			this.state = STATE_EXPECT_SPACE
		} else {
			this.Reset(false)
		}
	default:
		this.Reset(false)
	}
}

func (this *codec) eject(bits ...bool) {
	this.bits = append(this.bits, bits...)

	/*
		for _, b := range this.bits {
			if b {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		fmt.Println("")
	*/

	// Return if not the right number of bits
	if uint(len(this.bits)) != this.bit_length*2 {
		return
	}

	// Accumlate into value
	var value uint32
	for i, j := 0, 0; i < int(this.bit_length); i, j = i+1, j+2 {
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
	this.emit(value, this.repeat)
}

func (this *codec) emit(value uint32, repeat bool) {
	if scancode, device, err := codeForCodec(this.codec_type, value); err != nil {
		if err != gopi.ErrBadParameter {
			this.log.Warn("Emit: %v", err)
		}
	} else {
		this.Emit(remotes.NewRemoteEvent(this, time.Since(timestamp), scancode, device, repeat))
	}
}

func codeForCodec(codec remotes.CodecType, value uint32) (uint32, uint32, error) {
	switch codec {
	case remotes.CODEC_RC5:
		// scancode is lowest 6 bits (0x03F), device is next 5 bits (7C0)
		scancode := value & 0x003F
		device := value & 0x07C0 >> 6
		//header := value & 0x3800 >> 11
		//fmt.Printf("value=%v header=%v scancode=%X device=%X\n", strconv.FormatUint(uint64(value), 2), strconv.FormatUint(uint64(header), 2), scancode, device)
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
