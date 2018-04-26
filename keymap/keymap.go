/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Configuration
type Database struct {
	// Root for the keymap files
	Root string

	// Extension for keymap files
	Ext string
}

// Driver
type db struct {
	// Logger
	log gopi.Logger

	// Root path and file extension
	root, ext string

	// Keymap database
	keymap map[remotes.CodecType]map[uint32]*tuple

	// New keymap
	empty *remotes.KeyMap
}

// (path,keymap tuple)
type tuple struct {
	path     string
	keymap   *remotes.KeyMap
	modified bool
}

// callback function for allKeyMaps
type allKeyMapsFunc func(*tuple) bool

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_EXT    = ".keymap"
	KEYCODE_PREFIX = "KEYCODE_"
	CODEC_PREFIX   = "CODEC_"
)

var (
	allKeyCodes   []*remotes.KeyMapEntry
	keyCodeTokens map[remotes.RemoteCode][]string
)

/////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Database) ext() string {
	if config.Ext == "" {
		return DEFAULT_EXT
	} else if strings.HasPrefix(".", config.Ext) {
		return config.Ext
	} else {
		return "." + config.Ext
	}
}

func (config Database) root() string {
	if config.Root == "" {
		if root, err := os.Getwd(); err != nil {
			return ""
		} else {
			return root
		}
	} else if stat, err := os.Stat(config.Root); os.IsNotExist(err) || stat.IsDir() == false {
		return ""
	} else {
		return config.Root
	}
}

func (config Database) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<keymap.db>Open{ root=\"%v\" ext=\"%v\"}", config.Root, config.ext())

	this := new(db)
	this.log = log
	this.root = config.root()
	this.ext = config.ext()
	this.keymap = make(map[remotes.CodecType]map[uint32]*tuple)

	if this.root == "" || this.ext == "" {
		log.Debug("keymap.db: Bad Parameter (root=%v ext=%v)", this.root, this.ext)
		return nil, gopi.ErrBadParameter
	}

	return this, nil
}

func (this *db) Close() error {
	this.log.Debug("<keymap.db>Close{ root=\"%v\"}", this.root)

	// Warn on modified keymaps
	if this.Modified() {
		this.log.Warn("<keymap.db>Close: There are modified keymaps, invoke SaveModifiedKeyMaps before Close")
	}

	// Release resources
	this.keymap = nil
	this.empty = nil

	// Return success
	return nil
}

func init() {
	allKeyCodes = getAllKeycodes()
	keyCodeTokens = make(map[remotes.RemoteCode][]string, len(allKeyCodes))
}

/////////////////////////////////////////////////////////////////////
// INTERFACE IMPLEMENTATION

// Create a new KeyMap and register it
func (this *db) NewKeyMap(name string) *remotes.KeyMap {

	// Return nil if the name is invalid
	if name = strings.TrimSpace(name); name == "" {
		return nil
	}

	this.log.Debug2("<keymap.db>NewKeyMap{ name=\"%v\"}", name)

	// Create a new empty KeyMap file
	keymap := new(remotes.KeyMap)
	keymap.Type = remotes.CODEC_NONE
	keymap.Device = remotes.DEVICE_UNKNOWN
	keymap.Name = name
	keymap.Repeats = 0
	keymap.Map = make([]*remotes.KeyMapEntry, 0)

	// Make this the 'empty' keymap
	if this.empty != nil {
		this.log.Warn("<keymap.db>NewKeyMap: Overwriting an existing empty keymap")
	}
	this.empty = keymap

	// Success
	return keymap
}

// Load keymaps from the root path, or another path
func (this *db) LoadKeyMaps(path string, callback remotes.LoadSaveCallbackFunc) error {

	// Override path with root if it's empty
	if path == "" {
		path = this.root
	}

	// Debug output
	this.log.Debug2("<keymap.db>LoadKeyMaps{ path=\"%v\"}", path)

	// Check path to make sure it's a directory
	if stat, err := os.Stat(path); os.IsNotExist(err) || stat.IsDir() == false {
		if err != nil {
			return err
		} else {
			return gopi.ErrBadParameter
		}
	}
	// Walk path loading in files with the correct extension
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsDir() {
			return nil
		}
		if info.Mode().IsRegular() && filepath.Ext(path) == this.ext {
			if keymap, err := this.LoadKeyMap(path); err != nil {
				return err
			} else if callback != nil {
				callback(path, keymap)
			}
		}
		return nil
	})
}

// Load a single keymap
func (this *db) LoadKeyMap(path string) (*remotes.KeyMap, error) {
	this.log.Debug2("<keymap.db>LoadKeyMap{ path=\"%v\"}", path)

	keymap := new(remotes.KeyMap)
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	dec := xml.NewDecoder(fh)
	if err := dec.Decode(keymap); err != nil {
		return nil, err
	} else if err := this.registerNewKeyMap(path, keymap, false); err != nil {
		return nil, err
	}

	// Success
	return keymap, nil
}

// Return true if any keymaps were modified
func (this *db) Modified() bool {
	// Iterate through all the existing keymaps
	for codec := range this.keymap {
		for device := range this.keymap[codec] {
			// Ignore codecs with invalid codec/device combination
			if codec == remotes.CODEC_NONE || device == remotes.DEVICE_UNKNOWN {
				continue
			}
			// Don't save unmodified keymaps
			if tuple := this.keymap[codec][device]; tuple.modified {
				return true
			}
		}
	}
	// Nothing modified at this point
	return false
}

// Save all modified keymaps
func (this *db) SaveModifiedKeyMaps(callback remotes.LoadSaveCallbackFunc) error {
	this.log.Debug2("<keymap.db>SaveModifiedKeyMaps{ modified=%v }", this.Modified())

	// Iterate through all the existing keymaps
	for codec := range this.keymap {
		for device := range this.keymap[codec] {
			// Ignore codecs with invalid codec/device combination
			if codec == remotes.CODEC_NONE || device == remotes.DEVICE_UNKNOWN {
				continue
			}
			// Don't save unmodified keymaps
			if t := this.keymap[codec][device]; t.modified == false {
				continue
			} else if err := this.SaveKeyMap(t.keymap, t.path); err != nil {
				return err
			} else {
				// Callback
				if callback != nil {
					callback(t.path, t.keymap)
				}
				// Set modified to no
				t.modified = false
			}
		}
	}

	// Success
	return nil
}

// Save keymap to file
func (this *db) SaveKeyMap(keymap *remotes.KeyMap, path string) error {
	this.log.Debug2("<keymap.db>SaveKeyMap{ path=%v keymap=%v }", path, keymap)

	// Sanity check codec and device
	if keymap.Type == remotes.CODEC_NONE || keymap.Device == remotes.DEVICE_UNKNOWN {
		this.log.Debug("<keymap.db>SaveKeyMap: Invalid codec and/or device")
		return gopi.ErrBadParameter
	}

	// Save
	if fh, err := os.Create(path); err != nil {
		return err
	} else {
		defer fh.Close()

		// Encode XML
		enc := xml.NewEncoder(fh)
		enc.Indent("", "  ")
		if err := enc.Encode(keymap); err != nil {
			return err
		}
	}

	// Success
	return nil
}

func (this *db) KeyMaps(codec remotes.CodecType, device uint32, name string) []*remotes.KeyMap {
	this.log.Debug2("<keymap.db>KeyMaps{ codec=%v device=0x%08X name=%v }", codec, device, name)
	keymaps := this.allKeyMaps(func(t *tuple) bool {
		this.log.Debug("<keymap.db>KeyMaps{ tuple=%v }", t)
		if codec != remotes.CODEC_NONE && codec != t.keymap.Type {
			return false
		}
		if device != remotes.DEVICE_UNKNOWN && device != t.keymap.Device {
			return false
		}
		if name != "" && name != t.keymap.Name {
			return false
		}
		return true
	})
	return keymaps
}

func (this *db) LookupKeyCode(searchterms ...string) []*remotes.KeyMapEntry {
	this.log.Debug2("<keymap.db>LookupKeyCode{ searchterm=%v }", searchterms)

	// Without any search terms, all keymap entries are returned
	if len(searchterms) == 0 {
		return allKeyCodes
	}

	// Or else we create a map of search entries keyed by Keycode
	keycodes := make(map[remotes.RemoteCode]*remotes.KeyMapEntry, len(allKeyCodes))

	// Iterate through the search terms
	for _, token := range searchterms {
		token = strings.ToUpper(token)
		for _, k := range allKeyCodes {
			// Exact match for the partial constant name
			if strings.HasPrefix(token, KEYCODE_PREFIX) && strings.HasPrefix(fmt.Sprint(k.Keycode), token) {
				keycodes[k.Keycode] = k
				continue
			}
			// Fuzzy match
			if fuzzyKeyCodeMatch(k, token) {
				keycodes[k.Keycode] = k
				continue
			}
		}
	}

	// De-duplicate keys
	keycodekeys := make([]*remotes.KeyMapEntry, 0, len(keycodes))
	for _, entry := range keycodes {
		keycodekeys = append(keycodekeys, entry)
	}
	return keycodekeys
}

func (this *db) SetKeyMapEntry(keymap *remotes.KeyMap, codec remotes.CodecType, device uint32, keycode remotes.RemoteCode, scancode uint32) error {
	this.log.Debug2("<keymap.db>SetKeyMapEntry{ keymap=\"%v\" codec=%v device=0x%08X keycode=%v scancode=0x%08X }", keymap.Name, codec, device, keycode, scancode)

	// Sanity check to make sure keymap codec and device are correct
	if codec == remotes.CODEC_NONE || device == remotes.DEVICE_UNKNOWN {
		this.log.Debug("SetKeyMapEntry: Cannot register keymap with CODEC_NONE or DEVICE_UNKNOWN")
		return gopi.ErrBadParameter
	}

	// Set the codec in the keymap if CODEC_NONE
	if keymap.Type == remotes.CODEC_NONE {
		keymap.Type = codec
	} else if keymap.Type != codec {
		return fmt.Errorf("Different codec (%v) than expected (%v)", codec, keymap.Type)
	}
	if keymap.Device == remotes.DEVICE_UNKNOWN {
		keymap.Device = device
	} else if keymap.Device != device {
		return fmt.Errorf("Different device (0x%08X) than expected (0x%08X)", device, keymap.Device)
	}

	// If this is the 'empty' keymap then register it, If there's a collision with an
	// existing keymap file, then return an error
	if keymap == this.empty {
		path := uniqueKeyMapPath(codec, device, this.root, this.ext)
		if err := this.registerNewKeyMap(path, keymap, true); err == remotes.ErrDuplicateKeyMap {
			return remotes.ErrDuplicateKeyMap
		} else if err != nil {
			return err
		} else {
			// Empty
			this.empty = nil
		}
	}

	// Obtain the tuple for this keymap - it needs to exist
	if tuple := this.getTuple(codec, device); tuple == nil || tuple.keymap != keymap {
		this.log.Debug("SetKeyMapEntry: Invalid keymap file")
		return gopi.ErrBadParameter
	} else {
		// Set modified
		tuple.modified = true
	}

	// Search through the entries looking for an existing one. There
	// is one keycode per map
	for _, entry := range keymap.Map {
		if entry.Keycode == keycode {
			// Set defaults
			entry.Scancode = scancode
			entry.Type = remotes.CODEC_NONE
			entry.Device = 0
			entry.Repeats = 0
			// Set name
			if entry.Name == "" {
				entry.Name = defaultKeyName(entry.Keycode)
			}
			// Return success
			return nil
		}
	}

	// Here we're adding a new entry
	keymap.Map = append(keymap.Map, &remotes.KeyMapEntry{
		Scancode: scancode,
		Name:     defaultKeyName(keycode),
	})

	// Return success
	return nil
}

func (this *db) LookupKeyMapEntry(codec remotes.CodecType, device uint32, scancode uint32) []*remotes.KeyMapEntry {
	this.log.Debug2("<keymap.db>LookupKeyMapEntry{ codec=%v device=0x%08X scancode=0x%08X }", codec, device, scancode)

	// Iterate through the keymaps to find the keymapentry
	if keymaps := this.KeyMaps(codec, device, ""); len(keymaps) == 0 {
		return nil
	} else {
		entries := make([]*remotes.KeyMapEntry, 0, 1)
		for _, keymap := range keymaps {
			for _, entry := range keymap.Map {
				// Scancode search
				if scancode != remotes.SCANCODE_UNKNOWN && scancode != entry.Scancode {
					continue
				}
				// Append a new entry with additional details populated
				entries = appendKeyMapEntry(entries, entry, keymap)
			}
		}
		if len(entries) == 0 {
			return nil
		} else {
			return entries
		}
	}
}

func appendKeyMapEntry(array []*remotes.KeyMapEntry, entry *remotes.KeyMapEntry, keymap *remotes.KeyMap) []*remotes.KeyMapEntry {
	new_entry := &remotes.KeyMapEntry{
		Scancode: entry.Scancode,
		Keycode:  entry.Keycode,
		Name:     entry.Name,
		Device:   entry.Device,
		Type:     entry.Type,
		Repeats:  entry.Repeats,
	}
	// Override some values if they are empty
	if new_entry.Name == "" {
		new_entry.Name = defaultKeyName(entry.Keycode)
	}
	if new_entry.Device == 0 || new_entry.Device == remotes.DEVICE_UNKNOWN {
		new_entry.Device = keymap.Device
	}
	if new_entry.Type == remotes.CODEC_NONE {
		new_entry.Type = keymap.Type
	}
	if new_entry.Repeats == 0 {
		new_entry.Repeats = keymap.Repeats
	}
	return append(array, new_entry)
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *db) registerNewKeyMap(path string, keymap *remotes.KeyMap, modified bool) error {

	// Sanity check to make sure keymap codec and device are correct
	if keymap.Type == remotes.CODEC_NONE || keymap.Device == remotes.DEVICE_UNKNOWN {
		this.log.Debug("registerNewKeyMap: Cannot register keymap with CODEC_NONE or DEVICE_UNKNOWN")
		return gopi.ErrBadParameter
	}

	// Create device mapping if it doesn't exist
	if _, exists := this.keymap[keymap.Type]; exists == false {
		this.keymap[keymap.Type] = make(map[uint32]*tuple, 1)
	}

	// Check for duplicate mapping
	mapping := this.keymap[keymap.Type]
	if _, exists := mapping[keymap.Device]; exists {
		return remotes.ErrDuplicateKeyMap
	} else {
		mapping[keymap.Device] = &tuple{
			path:     path,
			keymap:   keymap,
			modified: modified,
		}
	}

	// Success
	return nil
}

func (this *db) allKeyMaps(callback allKeyMapsFunc) []*remotes.KeyMap {
	keymaps := make([]*remotes.KeyMap, 0)
	for _, devices := range this.keymap {
		for _, tuple := range devices {
			if callback(tuple) {
				keymaps = append(keymaps, tuple.keymap)
			}
		}
	}
	return keymaps
}

func (this *db) getTuple(codec remotes.CodecType, device uint32) *tuple {
	if devices, exists := this.keymap[codec]; exists == false {
		return nil
	} else if keymap, exists := devices[device]; exists == false {
		return nil
	} else {
		return keymap
	}
}

func getAllKeycodes() []*remotes.KeyMapEntry {
	keycodes := make([]*remotes.KeyMapEntry, 0, 100)
	for c := remotes.KEYCODE_NONE; c < remotes.KEYCODE_MAX; c++ {
		if name := fmt.Sprint(c); strings.HasPrefix(name, KEYCODE_PREFIX) {
			// Convert name into words
			name = strings.ToLower(strings.TrimPrefix(name, KEYCODE_PREFIX))
			name = strings.Title(strings.Replace(name, "_", " ", -1))
			// Append keycode
			keycodes = append(keycodes, &remotes.KeyMapEntry{
				Scancode: remotes.SCANCODE_UNKNOWN,
				Keycode:  c,
				Name:     name,
				Device:   remotes.DEVICE_UNKNOWN,
				Type:     remotes.CODEC_NONE,
				Repeats:  0,
			})
		}
	}
	return keycodes
}

func fuzzyKeyCodeMatch(keycode *remotes.KeyMapEntry, token string) bool {
	if tokens, exists := keyCodeTokens[keycode.Keycode]; exists == false {
		tokens = make([]string, 0, 1)
		tokens = append(tokens, strings.Split(strings.ToUpper(keycode.Name), " ")...)
		constant_name := strings.TrimPrefix(strings.ToUpper(fmt.Sprint(keycode.Keycode)), KEYCODE_PREFIX)
		tokens = append(tokens, constant_name)
		tokens = append(tokens, strings.Split(constant_name, "_")...)
		keyCodeTokens[keycode.Keycode] = tokens
		return fuzzyKeyCodeMatch(keycode, token)
	} else {
		for _, t := range tokens {
			if token == t {
				return true
			}
		}
		return false
	}
}

func uniqueKeyMapPath(codec remotes.CodecType, device uint32, root, ext string) string {
	codec_name := strings.ToLower(strings.TrimPrefix(fmt.Sprint(codec), CODEC_PREFIX))
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("%v_%08X_%v%v", codec_name, device, i, ext)
		path := filepath.Join(root, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
	// Cannot find a unique name
	return ""
}

func defaultKeyName(keycode remotes.RemoteCode) string {
	if name := fmt.Sprint(keycode); strings.HasPrefix(name, KEYCODE_PREFIX) == false {
		// Invalid key
		return ""
	} else {
		name = strings.TrimLeft(name, KEYCODE_PREFIX)
		name = strings.Replace(name, "_", " ", -1)
		return strings.Title(strings.ToLower(name))
	}
}
