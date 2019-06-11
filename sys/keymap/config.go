/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type keymap struct {
	Name_   string            `json:"name"`
	Codec_  remotes.CodecType `json:"codec",omitempty`
	Device_ uint32            `json:"device",omitempty`
	Keys_   []*key            `json:"keys",omitempty`
	map_    map[string]*key
}

type key struct {
	Name_     string            `json:"name"`
	Codec_    remotes.CodecType `json:"codec",omitempty`
	Device_   uint32            `json:"device",omitempty`
	Scancode_ uint32            `json:"scancode",omitempty`
	Keycode_  gopi.KeyCode      `json:"keycode",omitempty`
	Keystate_ gopi.KeyState     `json:"keystate",omitempty`
}

type db_ struct {
	Keymaps []*keymap `json:"keymaps"`
}

type config struct {
	// Private Members
	log      gopi.Logger
	path     string
	modified bool

	db_
	sync.Mutex
	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	FILENAME_DEFAULT = "keymap.json"
	WRITE_DELTA      = 5 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DESTROY

func (this *config) Init(config Keymap, logger gopi.Logger) error {
	logger.Debug("<keymap.config.Init>{ config=%+v }", config)

	this.log = logger
	this.Keymaps = make([]*keymap, 0)

	// Read or create file
	if config.Path != "" {
		if err := this.ReadPath(config.Path); err != nil {
			return fmt.Errorf("ReadPath: %v: %v", config.Path, err)
		}
	}

	// Start process to write occasionally to disk
	this.Tasks.Start(this.WriteConfigTask)

	// Success
	return nil
}

func (this *config) Destroy() error {
	this.log.Debug("<keymap.config.Destroy>{ path=%v }", strconv.Quote(this.path))

	// Stop all tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *config) String() string {
	return fmt.Sprintf("<keymap.config>{ path=%v }", strconv.Quote(this.path))
}

func (this *keymap) String() string {
	return fmt.Sprintf("<keymap>{ name=%v }", strconv.Quote(this.Name_))
}

func (this *key) String() string {
	return fmt.Sprintf("<key>{ name=%v id=%v }", strconv.Quote(this.Name_), this.Id())
}

////////////////////////////////////////////////////////////////////////////////
// READ AND WRITE CONFIG

// SetModified sets the modified flag to true
func (this *config) SetModified() {
	this.Lock()
	defer this.Unlock()
	this.modified = true
}

// ReadPath creates regular file if it doesn't exist, or else reads from the path
func (this *config) ReadPath(path string) error {
	this.log.Debug2("<keymap.config>ReadPath{ path=%v }", strconv.Quote(path))

	// Append home directory if relative path
	if filepath.IsAbs(path) == false {
		if homedir, err := os.UserHomeDir(); err != nil {
			return err
		} else {
			path = filepath.Join(homedir, path)
		}
	}

	// Set path
	this.path = path

	// Append filename
	if stat, err := os.Stat(this.path); err == nil && stat.IsDir() {
		// append default filename
		this.path = filepath.Join(this.path, FILENAME_DEFAULT)
	}

	// Read file
	if stat, err := os.Stat(this.path); err == nil && stat.Mode().IsRegular() {
		if err := this.ReadPath_(this.path); err != nil {
			return err
		} else {
			return nil
		}
	} else if os.IsNotExist(err) {
		// Create file
		if fh, err := os.Create(this.path); err != nil {
			return err
		} else if err := fh.Close(); err != nil {
			return err
		} else {
			this.SetModified()
			return nil
		}
	} else {
		return err
	}
}

// WritePath writes the configuration file to disk
func (this *config) WritePath(path string, indent bool) error {
	this.log.Debug2("<keymap.config>WritePath{ path=%v indent=%v }", strconv.Quote(path), indent)
	this.Lock()
	defer this.Unlock()
	if fh, err := os.Create(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if err := this.Writer(fh, indent); err != nil {
			return err
		} else {
			this.modified = false
		}
	}

	// Success
	return nil
}

func (this *config) ReadPath_(path string) error {
	this.Lock()
	defer this.Unlock()

	if fh, err := os.Open(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if err := this.Reader(fh); err != nil {
			return err
		} else {
			this.modified = false
		}
	}

	// Success
	return nil
}

// Reader reads the configuration from an io.Reader object
func (this *config) Reader(fh io.Reader) error {
	dec := json.NewDecoder(fh)
	if err := dec.Decode(&this.db_); err != nil {
		return err
	} else {
		// Regenerate keymaps
		for i, keymap := range this.db_.Keymaps {
			this.db_.Keymaps[i] = keymap.copy()
		}
	}

	// Success
	return nil
}

// Writer writes an array of service records to a io.Writer object
func (this *config) Writer(fh io.Writer, indent bool) error {
	enc := json.NewEncoder(fh)
	if indent {
		enc.SetIndent("", "  ")
	}
	this.log.Warn("Write", this.db_)
	if err := enc.Encode(this.db_); err != nil {
		return err
	}
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *config) WriteConfigTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	ticker := time.NewTimer(100 * time.Millisecond)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if this.modified {
				if this.path == "" {
					// Do nothing
				} else if err := this.WritePath(this.path, true); err != nil {
					this.log.Warn("Write: %v: %v", this.path, err)
				}
			}
			ticker.Reset(WRITE_DELTA)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Stop the ticker
	ticker.Stop()

	// Try and write
	if this.modified {
		if this.path == "" {
			// Do nothing
		} else if err := this.WritePath(this.path, true); err != nil {
			this.log.Warn("Write: %v: %v", this.path, err)
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *config) lookup(name string, codec remotes.CodecType, device uint32) *keymap {
	this.Lock()
	defer this.Unlock()

	// Lookup keymap
	for _, keymap := range this.Keymaps {
		if name != "" && keymap.Name() != name {
			continue
		}
		if codec != remotes.CODEC_NONE && keymap.Codec() != codec {
			continue
		}
		if device != remotes.DEVICE_UNKNOWN && keymap.Device() != device {
			continue
		}
		return keymap
	}

	// Not found, return nil
	return nil
}

func (this *config) create_keymap(name string, codec remotes.CodecType, device uint32) *keymap {
	if keymap := new_keymap(name, codec, device); keymap == nil {
		return nil
	} else {
		// Critical section
		this.Lock()
		this.Keymaps = append(this.Keymaps, keymap)
		this.Unlock()
		// Set modified
		this.SetModified()
		return keymap
	}
}

func (this *config) lookup_key_event(k *keymap, e remotes.RemoteEvent) *key {
	this.Lock()
	defer this.Unlock()
	if key := k.new_key_event(e, gopi.KEYCODE_NONE, gopi.KEYSTATE_NONE); key == nil {
		return nil
	} else if key, exists := k.map_[key.Id()]; exists {
		return key
	} else {
		return nil
	}
}

func (this *config) new_key_event(k *keymap, e remotes.RemoteEvent, code gopi.KeyCode, state gopi.KeyState) *key {
	if key := this.keycodes.new_key(e, code, state); key == nil {
		return nil
	} else {
		this.Lock()
		defer this.Unlock()
		if key = k.add_key(key); key != nil {
			this.modified = true
		}
		return key
	}
}

////////////////////////////////////////////////////////////////////////////////
// KEYMAP

func new_keymap(name string, codec remotes.CodecType, device uint32) *keymap {
	k := new(keymap)
	k.Name_ = name
	k.Codec_ = codec
	k.Device_ = device
	k.Keys_ = make([]*key, 0)
	k.map_ = make(map[string]*key)
	return k
}

func (this *keymap) Name() string {
	return this.Name_
}

func (this *keymap) Device() uint32 {
	return this.Device_
}

func (this *keymap) Codec() remotes.CodecType {
	return this.Codec_
}

func (this *keymap) Keys() []remotes.Key {
	keys := make([]remotes.Key, len(this.Keys_))
	for i, key := range this.Keys_ {
		keys[i] = key
	}
	return keys
}

func (this *keymap) copy() *keymap {
	if other := new_keymap(this.Name(), this.Codec(), this.Device()); other == nil {
		return nil
	} else {
		for _, key := range this.Keys_ {
			other.add_key(key)
		}
		return other
	}
}

func (this *keymap) add_key(k *key) *key {
	if k == nil {
		return nil
	} else if _, exists := this.map_[k.Id()]; exists {
		return nil
	} else if k.Keycode_ == gopi.KEYCODE_NONE {
		return nil
	} else {
		this.Keys_ = append(this.Keys_, k)
		this.map_[k.Id()] = k
		return k
	}
}

////////////////////////////////////////////////////////////////////////////////
// KEY

func (this *key) new_key() *key {

}

func (this *key) Id() string {
	return fmt.Sprintf("%08X-%08X-%08X", uint32(this.Codec_), this.Device_, this.Scancode_)
}

func (this *key) Name() string {
	return this.Name_
}

func (this *key) Codec() remotes.CodecType {
	return this.Codec_
}

func (this *key) Device() uint32 {
	return this.Device_
}

func (this *key) ScanCode() uint32 {
	return this.Scancode_
}

func (this *key) KeyCode() gopi.KeyCode {
	return this.Keycode_
}

func (this *key) KeyState() gopi.KeyState {
	return this.Keystate_
}
