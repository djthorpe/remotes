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
	events chan gopi.Event
}

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
// ACCEPT EVENTS

func (this *codec) acceptEvents(ctx context.Context) {
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-this.events:
			this.log.Info("EVT=%v", evt)
		}
	}
	this.done <- gopi.DONE
}
