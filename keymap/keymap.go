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

	// Keyentry indexes
	bycodec    map[remotes.CodecType][]*etuple
	bydevice   map[uint32][]*etuple
	byscancode map[uint32][]*etuple

	// New keymap
	empty *remotes.KeyMap
}

// (path,keymap tuple)
type tuple struct {
	path     string
	keymap   *remotes.KeyMap
	modified bool
}

// (entry,keymap)
type etuple struct {
	entry  *remotes.KeyMapEntry
	keymap *remotes.KeyMap
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

	if this.root == "" {
		log.Debug("keymap: Bad Parameter (root=%v ext=%v)", this.root, this.ext)
		return nil, fmt.Errorf("Path not found: %v", config.Root)
	}
	if this.ext == "" {
		log.Debug("keymap: Bad Parameter (root=%v ext=%v)", this.root, this.ext)
		return nil, gopi.ErrBadParameter
	}

	this.keymap = make(map[remotes.CodecType]map[uint32]*tuple)
	this.bycodec = make(map[remotes.CodecType][]*etuple)
	this.bydevice = make(map[uint32][]*etuple)
	this.byscancode = make(map[uint32][]*etuple)

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
	this.bycodec = nil
	this.bydevice = nil
	this.byscancode = nil
	this.empty = nil

	// Return success
	return nil
}

func (this *db) String() string {
	return fmt.Sprintf("<keymap.db>{ root=%v empty=%v keymaps=%v }", this.root, this.empty, this.keymap)
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

// Load keymaps from the root path
func (this *db) LoadKeyMaps(callback remotes.LoadSaveCallbackFunc) error {
	this.log.Debug2("<keymap.db>LoadKeyMaps{ path=\"%v\"}", this.root)

	// Check path to make sure it's a directory
	if stat, err := os.Stat(this.root); os.IsNotExist(err) || stat.IsDir() == false {
		if err != nil {
			return err
		} else {
			return gopi.ErrBadParameter
		}
	}
	// Walk path loading in files with the correct extension
	return filepath.Walk(this.root, func(path string, info os.FileInfo, err error) error {
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
	// We generally check the codec and device to make sure they are
	// the same as the keymap values, but if the keymap has MultiCodec
	// set then we allow a variety of codecs and devices for a single
	// keymap
	if keymap.Type == remotes.CODEC_NONE {
		keymap.Type = codec
	} else if keymap.Type != codec && keymap.MultiCodec == false {
		return fmt.Errorf("Different codec (%v) than expected (%v)", codec, keymap.Type)
	}
	if keymap.Device == remotes.DEVICE_UNKNOWN {
		keymap.Device = device
	} else if keymap.Device != device && keymap.MultiCodec == false {
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
	if tuple := this.getTuple(keymap.Type, keymap.Device); tuple == nil || tuple.keymap != keymap {
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
			// Set codec & device overrides
			if keymap.Type != codec {
				entry.Type = codec
			}
			if keymap.Device != device {
				entry.Device = device
			}
			// Return success
			return nil
		}
	}

	// Here we're adding a new entry
	keymap.Map = appendKeyMapEntry(nil, keymap, &remotes.KeyMapEntry{
		Keycode:  keycode,
		Scancode: scancode,
		Name:     "",
	})

	// Reindex the keymap
	if err := this.indexKeyMap(keymap); err != nil {
		return err
	}

	// Return success
	return nil
}

func (this *db) GetKeyMapEntry(keymap *remotes.KeyMap, codec remotes.CodecType, device uint32, keycode remotes.RemoteCode, scancode uint32) []*remotes.KeyMapEntry {
	this.log.Debug2("<keymap.db>GetKeyMapEntry{ keymap=\"%v\" codec=%v device=0x%08X keycode=%v scancode=0x%08X }", keymap.Name, codec, device, keycode, scancode)

	// Search through the keymap and return the entries which match
	entries := make([]*remotes.KeyMapEntry, 0, 1)
	for _, entry := range keymap.Map {
		// Continue if the codec doesn't match
		if codec != remotes.CODEC_NONE {
			if entry.Type != remotes.CODEC_NONE && codec != entry.Type {
				continue
			} else if codec != keymap.Type {
				continue
			}
		}
		// Continue if device doesn't match
		if device != remotes.DEVICE_UNKNOWN {
			if entry.Device != remotes.DEVICE_UNKNOWN && entry.Device != 0 && entry.Device != device {
				continue
			} else if keymap.Device != device {
				continue
			}
		}
		// Continue if scancode doesn't match
		if scancode != remotes.SCANCODE_UNKNOWN {
			if entry.Scancode != remotes.SCANCODE_UNKNOWN && scancode != entry.Scancode {
				continue
			}
		}
		// Continue if keycode doesn't match
		if keycode != remotes.KEYCODE_NONE {
			if entry.Keycode != remotes.KEYCODE_NONE && keycode != entry.Keycode {
				continue
			}
		}
		// Or else we want to populate the entry
		entries = appendKeyMapEntry(entries, keymap, entry)
	}
	if len(entries) == 0 {
		return nil
	} else {
		return entries
	}
}

func (this *db) LookupKeyMapEntry(codec remotes.CodecType, device uint32, scancode uint32) map[*remotes.KeyMapEntry]*remotes.KeyMap {
	this.log.Debug2("<keymap.db>LookupKeyMapEntry{ codec=%v device=0x%08X scancode=0x%08X }", codec, device, scancode)

	if tuples := this.lookupEntryTuples(codec, device, scancode); len(tuples) == 0 {
		return nil
	} else {
		// Create entries from tuples
		entries := make(map[*remotes.KeyMapEntry]*remotes.KeyMap, len(tuples))
		for _, tuple := range tuples {
			entry := newKeyMapEntry(tuple.keymap, tuple.entry, false)
			entries[entry] = tuple.keymap
		}
		return entries
	}
}

func (this *db) DeleteKeyMapEntry(keymap *remotes.KeyMap, entry *remotes.KeyMapEntry) error {
	this.log.Info("TODO: Delete %v from %v", entry, keymap.Name)
	return nil
}

/////////////////////////////////////////////////////////////////////
// SET PARAMETERS

func (this *db) SetName(keymap *remotes.KeyMap, name string) error {
	// TODO: Not yet implemented
	return gopi.ErrNotImplemented
}

func (this *db) SetRepeats(keymap *remotes.KeyMap, repeats uint) error {
	// Check parameters
	if keymap == nil {
		return gopi.ErrBadParameter
	}

	// The 'new' keymap case
	if keymap == this.empty {
		keymap.Repeats = repeats
		return nil
	}

	// Get the tuple for the keymap and modify the repeats value
	if tuple := this.getTuple(keymap.Type, keymap.Device); tuple == nil {
		return gopi.ErrBadParameter
	} else if tuple.keymap != keymap {
		return gopi.ErrBadParameter
	} else if tuple.keymap.Repeats == repeats {
		return nil
	} else {
		tuple.keymap.Repeats = repeats
		tuple.modified = true
		return nil
	}
}

func (this *db) SetMultiCodec(keymap *remotes.KeyMap, flag bool) error {
	// Check parameters
	if keymap == nil {
		return gopi.ErrBadParameter
	}

	// Get the tuple for the keymap and modify the multicodec value
	if tuple := this.getTuple(keymap.Type, keymap.Device); tuple == nil {
		return gopi.ErrBadParameter
	} else if tuple.keymap != keymap {
		return gopi.ErrBadParameter
	} else if tuple.keymap.MultiCodec == flag {
		return nil
	} else {
		tuple.keymap.MultiCodec = flag
		tuple.modified = true
		return nil
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newKeyMapEntry(keymap *remotes.KeyMap, entry *remotes.KeyMapEntry, minimized bool) *remotes.KeyMapEntry {
	new_entry := &remotes.KeyMapEntry{
		Scancode: entry.Scancode,
		Keycode:  entry.Keycode,
		Name:     entry.Name,
		Device:   entry.Device,
		Type:     entry.Type,
		Repeats:  entry.Repeats,
	}

	// Set the name field if empty
	if new_entry.Name == "" {
		new_entry.Name = defaultKeyName(entry.Keycode)
	}

	// Minimized means that it removes details from the entry that are
	// already mentioned in the keymap
	if minimized {
		if keymap.Type == entry.Type {
			new_entry.Type = remotes.CODEC_NONE
		}
		if keymap.Device == entry.Device {
			new_entry.Device = 0
		}
		if keymap.Repeats == entry.Repeats {
			new_entry.Repeats = 0
		}
	} else {
		if new_entry.Type == remotes.CODEC_NONE {
			new_entry.Type = keymap.Type
		}
		if new_entry.Device == 0 || new_entry.Device == remotes.DEVICE_UNKNOWN {
			new_entry.Device = keymap.Device
		}
		if new_entry.Repeats == 0 {
			new_entry.Repeats = keymap.Repeats
		}
	}
	return new_entry
}

func appendKeyMapEntry(array []*remotes.KeyMapEntry, keymap *remotes.KeyMap, entry *remotes.KeyMapEntry) []*remotes.KeyMapEntry {
	if array == nil {
		array = keymap.Map
	}
	return append(array, newKeyMapEntry(keymap, entry, array == nil))
}

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
	}

	// Index keymap
	if err := this.indexKeyMap(keymap); err != nil {
		return err
	}

	// Create the mapping
	mapping[keymap.Device] = &tuple{
		path:     path,
		keymap:   keymap,
		modified: modified,
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

func (this *db) indexKeyMap(keymap *remotes.KeyMap) error {
	// Check parameters
	if keymap == nil || len(keymap.Map) == 0 {
		return gopi.ErrBadParameter
	}

	// Iterate through the keymap entries
	for _, entry := range keymap.Map {
		if err := this.indexKeyMapEntry(keymap, entry); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

func (this *db) indexKeyMapEntry(keymap *remotes.KeyMap, entry *remotes.KeyMapEntry) error {

	// Remove any existing entries from all indexes
	for codec, entries := range this.bycodec {
		for _, etuple := range entries {
			if entry == etuple.entry {
				this.log.Warn("TODO: removeindexKeyMapEntry codec %v => %v", codec, entry)
			}
		}
	}
	for device, entries := range this.bydevice {
		for _, etuple := range entries {
			if entry == etuple.entry {
				this.log.Warn("TODO: removeindexKeyMapEntry device %08X => %v", device, entry)
			}
		}
	}
	for scancode, entries := range this.byscancode {
		for _, etuple := range entries {
			if entry == etuple.entry {
				this.log.Warn("TODO: removeindexKeyMapEntry scancode %08X => %v", scancode, entry)
			}
		}
	}

	// Set index parameters
	codec := entry.Type
	device := entry.Device
	scancode := entry.Scancode
	etuple_ptr := &etuple{entry: entry, keymap: keymap}
	if codec == remotes.CODEC_NONE {
		codec = keymap.Type
	}
	if device == 0 || device == remotes.DEVICE_UNKNOWN {
		device = keymap.Device
	}

	// Make the arrays
	if _, exists := this.bycodec[codec]; exists == false {
		this.bycodec[codec] = make([]*etuple, 0, 1)
	}
	if _, exists := this.bydevice[device]; exists == false {
		this.bydevice[device] = make([]*etuple, 0, 1)
	}
	if _, exists := this.byscancode[scancode]; exists == false {
		this.byscancode[scancode] = make([]*etuple, 0, 1)
	}

	// Index
	this.bycodec[codec] = append(this.bycodec[codec], etuple_ptr)
	this.bydevice[device] = append(this.bydevice[device], etuple_ptr)
	this.byscancode[scancode] = append(this.byscancode[scancode], etuple_ptr)

	// Return success
	return nil
}

func (this *db) lookupEntryTuples(codec remotes.CodecType, device uint32, scancode uint32) []*etuple {
	// Create an etuple hash to count the number of occurences
	counter := make(map[*etuple]uint, 100)
	term := uint(0)
	increment := func(tuples []*etuple) {
		for _, tuple := range tuples {
			if _, exists := counter[tuple]; exists {
				counter[tuple] += 1
			} else {
				counter[tuple] = 1
			}
		}
		term += 1
	}
	// Scancode search first
	if scancode != remotes.SCANCODE_UNKNOWN {
		if tuples, exists := this.byscancode[scancode]; exists == false {
			return nil
		} else {
			increment(tuples)
		}
	}
	// Device second
	if device != remotes.DEVICE_UNKNOWN {
		if tuples, exists := this.bydevice[device]; exists == false {
			return nil
		} else {
			increment(tuples)
		}
	}
	// Codec third
	if codec != remotes.CODEC_NONE {
		if tuples, exists := this.bycodec[codec]; exists == false {
			return nil
		} else {
			increment(tuples)
		}
	}
	// Return nil if no terms
	if term == 0 {
		return nil
	}
	// Check for matching all search criteria
	tuples := make([]*etuple, 0, 1)
	for tuple, t := range counter {
		if t == term {
			tuples = append(tuples, tuple)
		}
	}
	// Return tuples
	return tuples
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
		if name := defaultKeyName(c); name != "" {
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
		filename := fmt.Sprintf("%v_%08X", codec_name, device)
		if i == 0 {
			filename = filename + ext
		} else {
			filename = filename + "_" + fmt.Sprint(i) + ext
		}
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
