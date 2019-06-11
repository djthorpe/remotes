/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	"fmt"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Keymap struct {
	Path string
}

type keymapper struct {
	log gopi.Logger

	config
	keycodes
	event.Merger
	event.Publisher
	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Keymap) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<remotes.keymap>Open{ path=%v }", strconv.Quote(config.Path))

	this := new(keymapper)
	this.log = logger

	// Load configuration
	if err := this.config.Init(config, logger); err != nil {
		logger.Debug2("config.Init returned nil")
		return nil, err
	}
	if err := this.keycodes.Init(config, logger); err != nil {
		logger.Debug2("keycodes.Init returned nil")
		return nil, err
	}

	// Backround tasks
	this.Tasks.Start(this.EventTask)

	// Return success
	return this, nil
}

func (this *keymapper) Close() error {
	this.log.Debug("<remotes.keymap>Close>{ config=%v keycodes=%v }", this.config.String(), this.keycodes.String())

	// Remove subscribers
	this.Publisher.Close()

	// End tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Stop merging
	this.Merger.Close()

	// Release resources, etc
	if err := this.config.Destroy(); err != nil {
		return err
	}
	if err := this.keycodes.Destroy(); err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *keymapper) String() string {
	return fmt.Sprintf("<remotes.keymap>{ config=%v keycodes=%v }", this.config.String(), this.keycodes.String())
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASK

func (this *keymapper) EventTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	events := this.Merger.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt != nil {
				this.Emit(evt)
			}
		case <-stop:
			break FOR_LOOP

		}
	}

	this.Merger.Unsubscribe(events)

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *keymapper) RegisterCodec(codec remotes.Codec) error {
	this.log.Debug2("<remotes.keymap>RegisterCodec{ codec=%v }", codec)

	// Check parameters
	if codec == nil {
		return gopi.ErrBadParameter
	}

	// Subscribe to the codec
	this.Merger.Merge(codec)

	// Success
	return nil
}

// Return keymaps. Use DEVICE_UNKNOWN and CODEC_NONE to retrieve all keymaps
func (this *keymapper) Keymaps(remotes.CodecType, uint32) []remotes.Keymap {
	return nil
}

// Create a new keymap with name, codec and device
func (this *keymapper) New(name string, codec remotes.CodecType, device uint32) (remotes.Keymap, error) {
	this.log.Debug2("<remotes.keymap>New{ name=%v type=%v device=0x%08X }", strconv.Quote(name), codec, device)

	// The name cannot be empty
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, gopi.ErrBadParameter
	}

	// See if keymap exists with same name, codec and device
	if keymap := this.config.lookup(name, remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN); keymap != nil {
		return keymap, remotes.ErrDuplicateKeyMap
	} else if keymap := this.config.lookup(name, codec, device); keymap != nil {
		return keymap, remotes.ErrDuplicateKeyMap
	} else if keymap := this.config.create_keymap(name, codec, device); keymap == nil {
		return nil, gopi.ErrAppError
	} else {
		return keymap, nil
	}
}

// Delete keymap
func (this *keymapper) Delete(remotes.Keymap) error {
	return gopi.ErrNotImplemented
}

// Lookup a key by an event
func (this *keymapper) LookupByEvent(remotes.RemoteEvent) remotes.Key {
	return nil
}

// Map an input event to a key
func (this *keymapper) Map(k remotes.Keymap, event remotes.RemoteEvent, code gopi.KeyCode, state gopi.KeyState) (remotes.Key, error) {
	this.log.Debug("<remotes.keymap>Map{ keymap=%v event=%v key_code=%v key_state=%v }", k, event, code, state)

	// Check incoming parameters
	if k == nil || event == nil || code == gopi.KEYCODE_NONE {
		return nil, gopi.ErrBadParameter
	}

	// There should be one key for any <codec>/<device>/<scancode>
	if keymap_, ok := k.(*keymap); keymap_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		key := this.config.lookup_key_event(keymap_, event)
		if key == nil {
			key = this.config.new_key_event(keymap_, event, code, state)
		}
		if key == nil {
			return nil, gopi.ErrBadParameter
		}
		// TODO: Set key code/state
		return key, nil
	}
}
