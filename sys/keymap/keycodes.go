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
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type keycode struct {
	name   string
	code   remotes.RemoteCode
	tokens map[string]bool
}

type keycodes struct {
	log  gopi.Logger
	keys map[gopi.KeyCode]*keycode
}

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	KEYCODE_PREFIX = "KEYCODE_"
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DESTROY

func (this *keycodes) Init(config Keymap, logger gopi.Logger) error {
	logger.Debug("<keymap.keycodes>Init{ config=%+v }", config)

	this.log = logger
	this.keys = make(map[gopi.KeyCode]*keycode)

	for code := remotes.KEYCODE_NONE; code < remotes.KEYCODE_MAX; code++ {
		if name := defaultKeyName(code); name != "" {
			this.keys[gopi.KeyCode(code)] = new_keycode(code, name)
		}
	}

	fmt.Println(this.keys)

	// Success
	return nil
}

func (this *keycodes) Destroy() error {
	this.log.Debug("<keymap.keycodes>Destroy>{ }")
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *keycodes) String() string {
	return fmt.Sprintf("<keymap.keycodes>{ num_keys=%v }", len(this.keys))
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

// Lookup keys by name, or all keys
func (this *keycodes) KeyByName(tokens ...string) []remotes.Key {
	// Where there are no tokens
	if len(tokens) == 0 {
		keys := make([]remotes.Key, len(this.keys))
		for i, keycode := range this.keys {
			keys[i] = this.key_from_keycode(keycode)
		}
		return keys
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create Key from Keycode
func (this *keycodes) key_from_keycode(k *keycode) remotes.Key {
	// Check incoming parameters
	if k == nil {
		return nil
	}
	key := new(key)
	key.Name_ = k.name
	key.Codec_ = remotes.CODEC_NONE
	key.Device_ = remotes.DEVICE_UNKNOWN
	key.Scancode_ = 0
	key.Keycode_ = gopi.KeyCode(k.code)
	key.Keystate_ = 0
	return key
}

// Create Key from Event
func (this *keycodes) key_from_event(remotes.RemoteEvent, gopi.KeyCode, gopi.KeyState) remotes.Key {
	return nil
}

func new_keycode(code remotes.RemoteCode, name string) *keycode {
	// Check incoming parameters
	if code == remotes.KEYCODE_NONE || name == "" {
		return nil
	}
	// Create tokens
	tokens := make([]string, 0, 1)
	tokens = append(tokens, strings.Split(strings.ToUpper(name), " ")...)
	constant_name := strings.TrimPrefix(strings.ToUpper(fmt.Sprint(code)), KEYCODE_PREFIX)
	tokens = append(tokens, constant_name)
	tokens = append(tokens, strings.Split(constant_name, "_")...)
	// Create keycode
	this := new(keycode)
	this.code = code
	this.name = name
	this.tokens = make(map[string]bool, len(tokens))
	// Index tokens
	for _, token := range tokens {
		this.tokens[token] = true
	}
	// Return keycode
	return this
}

func defaultKeyName(keycode remotes.RemoteCode) string {
	if name := fmt.Sprint(keycode); strings.HasPrefix(name, KEYCODE_PREFIX) == false {
		// Invalid key
		return ""
	} else {
		name = strings.ToLower(strings.TrimPrefix(name, KEYCODE_PREFIX))
		name = strings.Title(strings.Replace(name, "_", " ", -1))
		name = strings.Replace(name, " Hdmi", " HDMI", -1)
		name = strings.Replace(name, " Pc", " PC", -1)
		name = strings.Replace(name, " Aux", " AUX", -1)
		name = strings.Replace(name, " Cd", " CD", -1)
		name = strings.Replace(name, " Dvd", " DVD", -1)
		name = strings.Replace(name, " Pip", " PIP", -1)
		name = strings.Replace(name, " 10plus", " 10+", -1)
		return name
	}
}
