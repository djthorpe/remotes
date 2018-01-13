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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Sony Configuration
type Codec struct {
	LIRC gopi.LIRC
}

type codec struct {
	log    gopi.Logger
	lirc   gopi.LIRC
	cancel context.CancelFunc
	done   chan struct{}
	events <-chan gopi.Event
	state  state
	value  uint32
	length uint
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
	TOLERANCE = 25 // 25% tolerance on values
	LENGTH    = 12 // 12 bits per scancode
)

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	HEADER_PULSE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 2500, TOLERANCE)
	HEADER_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 550, TOLERANCE)
	ONE_PULSE     = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 1200, TOLERANCE)
	ZERO_PULSE    = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 600, TOLERANCE)
	ONEZERO_SPACE = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 600, TOLERANCE)
	REPEAT_SPACE  = remotes.NewMarkSpace(gopi.LIRC_TYPE_SPACE, 24534, TOLERANCE)
	TRAIL_PULSE   = remotes.NewMarkSpace(gopi.LIRC_TYPE_PULSE, 600, TOLERANCE)
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Codec) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<remotes.Codec.Sony.Open>{ lirc=%v }", config.LIRC)

	if config.LIRC == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(codec)
	this.log = log
	this.lirc = config.LIRC
	this.done = make(chan struct{})
	this.events = this.lirc.Subscribe()
	this.state = STATE_EXPECT_HEADER_PULSE

	if ctx, cancel := context.WithCancel(context.Background()); ctx != nil {
		this.cancel = cancel
		go this.acceptEvents(ctx)
	}

	return this, nil
}

func (this *codec) Close() error {
	this.log.Debug2("<remotes.Codec.Sony.Close>{ }")

	// Unsubscribe
	this.lirc.Unsubscribe(this.events)

	// Cancel background thread, wait for done signal
	this.cancel()
	_ = <-this.done

	// Blank out member variables
	close(this.done)
	this.events = nil
	this.lirc = nil
	this.done = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *codec) String() string {
	return fmt.Sprintf("<remotes.Codec.Sony>{}")
}

////////////////////////////////////////////////////////////////////////////////
// CODEC INTERFACE

func (this *codec) Name() string {
	return "sony"
}

func (this *codec) Reset() {
	this.state = STATE_EXPECT_HEADER_PULSE
	this.value = 0
	this.length = 0
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
	this.log.Debug2("<remotes.Codec.Sony.Receive>{ evt=%v }", evt)
	switch this.state {
	case STATE_EXPECT_HEADER_PULSE:
		if HEADER_PULSE.Matches(evt) {
			this.state = STATE_EXPECT_HEADER_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_HEADER_SPACE:
		if HEADER_SPACE.Matches(evt) {
			this.state = STATE_EXPECT_BIT
		} else {
			this.Reset()
		}
	case STATE_EXPECT_BIT:
		if ONE_PULSE.Matches(evt) {
			this.value |= 1
			this.state = STATE_EXPECT_SPACE
		} else if ZERO_PULSE.Matches(evt) {
			this.value |= 0
			this.state = STATE_EXPECT_SPACE
		} else {
			this.Reset()
		}
	case STATE_EXPECT_SPACE:
		if ONEZERO_SPACE.Matches(evt) {
			this.value = this.value << 1
			this.length = this.length + 1
			if this.length == (LENGTH - 1) {
				this.state = STATE_EXPECT_TRAIL
			} else {
				this.state = STATE_EXPECT_BIT
			}
		} else {
			this.Reset()
		}
	case STATE_EXPECT_TRAIL:
		if TRAIL_PULSE.Matches(evt) {
			fmt.Printf("value=%X\n", this.value)
			fmt.Printf("TRAIL\n")
			this.state = STATE_EXPECT_REPEAT
		} else {
			this.Reset()
		}
	case STATE_EXPECT_REPEAT:
		if REPEAT_SPACE.Matches(evt) {
			fmt.Printf("REPEAT\n")
		}
		this.Reset()
	default:
		this.Reset()
	}
}

/*
func (this *codec) send(scancode uint32) error {
	buf := make([]*uint32,0)
	buf = append(buf,HEADER_PULSE)
	buf = append(buf,HEADER_SPACE)
	for bit := 0; bit < LENGTH; bit++ {
		if value & 1 != 0 {
			buf = append(buf,ONE_PULSE)
		} else {
			buf = append(buf,ZERO_PULSE)
		}
		buf = append(buf,ONEZERO_SPACE)
		value = value >> 1
	}
*/
