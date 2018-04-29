/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"encoding/xml"
	"errors"
	"fmt"

	// Frameworks
	"github.com/djthorpe/gopi"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type (
	RemoteCode           gopi.KeyCode
	CodecType            uint
	LoadSaveCallbackFunc func(filename string, keymap *KeyMap)
)

// KeyMapEntry maps a (keycode,codec,device) onto a single scancode
type KeyMapEntry struct {
	Scancode uint32     `xml:"scancode"`
	Keycode  RemoteCode `xml:"keycode"`
	Name     string     `xml:"name"`
	Device   uint32     `xml:"id,attr,omitempty"` // Overrides device if non-zero
	Type     CodecType  `xml:"codec,omitempty"`   // Overrides codec if non-zero
	Repeats  uint       `xml:"repeats,omitempty"` // Overrides repeats if non-zero
}

// KeyMap maps one or more keys and scancodes
type KeyMap struct {
	XMLName xml.Name       `xml:"remote"`
	Type    CodecType      `xml:"codec"`
	Device  uint32         `xml:"id,attr,omitempty"`
	Name    string         `xml:"name"`
	Repeats uint           `xml:"repeats"`
	Map     []*KeyMapEntry `xml:"keymap"`
}

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	CODEC_NONE CodecType = iota
	CODEC_RC5
	CODEC_RC5X_20
	CODEC_RC5_SZ
	CODEC_JVC
	CODEC_SONY12
	CODEC_SONY15
	CODEC_SONY20
	CODEC_NEC16
	CODEC_NEC32
	CODEC_NECX
	CODEC_SANYO
	CODEC_RC6_0
	CODEC_RC6_6A_20
	CODEC_RC6_6A_24
	CODEC_RC6_6A_32
	CODEC_RC6_MCE
	CODEC_SHARP
	CODEC_APPLETV
	CODEC_PANASONIC
)

const (
	DEVICE_UNKNOWN   = 0xFFFFFFFF
	SCANCODE_UNKNOWN = 0xFFFFFFFF
)

/////////////////////////////////////////////////////////////////////
// INTERFACES

type Codec interface {
	gopi.Driver
	gopi.Publisher

	// Return type for the codec
	Type() CodecType

	// Send scancode
	Send(device uint32, scancode uint32, repeats uint) error
}

type KeyMaps interface {
	gopi.Driver

	// Return properties
	Modified() bool

	// Create a new KeyMap with unknown codec and device
	NewKeyMap(name string) *KeyMap

	// LoadKeyMaps database functions and load individual keymap
	LoadKeyMaps(callback LoadSaveCallbackFunc) error
	LoadKeyMap(path string) (*KeyMap, error)

	// Save modified KepMaps to files and save individual keymap
	SaveModifiedKeyMaps(callback LoadSaveCallbackFunc) error
	SaveKeyMap(keymap *KeyMap, path string) error

	// Get a keymap from the database. Use DEVICE_UNKNOWN and
	// CODEC_NONE to retrieve all keymaps
	KeyMaps(codec CodecType, device uint32, name string) []*KeyMap

	// Return keymapentry records matching a particular
	// set of search terms. Will return name and keycode in
	// each KeyMapEntry or nil. Returns all keycodes on empty
	// searchterm array
	LookupKeyCode(searchterm ...string) []*KeyMapEntry

	// Set, get and lookup KeyMapEntry mapping
	SetKeyMapEntry(keymap *KeyMap, codec CodecType, device uint32, keycode RemoteCode, scancode uint32) error
	GetKeyMapEntry(keymap *KeyMap, codec CodecType, device uint32, keycode RemoteCode, scancode uint32) []*KeyMapEntry
	LookupKeyMapEntry(codec CodecType, device uint32, scancode uint32) []*KeyMapEntry
}

/////////////////////////////////////////////////////////////////////
// ERROR CODES

var (
	ErrInvalidKey      = errors.New("Invalid Key")
	ErrDuplicateKeyMap = errors.New("Duplicate KeyMap")
	ErrNotFound        = errors.New("Not Found")
	ErrAmbiguous       = errors.New("Ambiguous Parameter")
)

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e *KeyMapEntry) String() string {
	params := fmt.Sprintf("name=\"%v\" keycode=%v scancode=0x%08X", e.Name, e.Keycode, e.Scancode)
	if e.Type != CODEC_NONE {
		params += fmt.Sprintf(" codec=%v", e.Type)
	}
	if e.Device != 0 && e.Device != DEVICE_UNKNOWN {
		params += fmt.Sprintf(" device=0x%08X", e.Device)
	}
	if e.Repeats != 0 {
		params += fmt.Sprintf(" repeats=%v", e.Repeats)
	}
	return "<remotes.KeyMapEntry>{ " + params + " }"
}

func (c CodecType) String() string {
	switch c {
	case CODEC_NONE:
		return "CODEC_NONE"
	case CODEC_RC5:
		return "CODEC_RC5"
	case CODEC_RC5X_20:
		return "CODEC_RC5X_20"
	case CODEC_RC5_SZ:
		return "CODEC_RC5_SZ"
	case CODEC_JVC:
		return "CODEC_JVC"
	case CODEC_SONY12:
		return "CODEC_SONY12"
	case CODEC_SONY15:
		return "CODEC_SONY15"
	case CODEC_SONY20:
		return "CODEC_SONY20"
	case CODEC_NEC16:
		return "CODEC_NEC16"
	case CODEC_NEC32:
		return "CODEC_NEC32"
	case CODEC_APPLETV:
		return "CODEC_APPLETV"
	case CODEC_NECX:
		return "CODEC_NECX"
	case CODEC_SANYO:
		return "CODEC_SANYO"
	case CODEC_RC6_0:
		return "CODEC_RC6_0"
	case CODEC_RC6_6A_20:
		return "CODEC_RC6_6A_20"
	case CODEC_RC6_6A_24:
		return "CODEC_RC6_6A_24"
	case CODEC_RC6_6A_32:
		return "CODEC_RC6_6A_32"
	case CODEC_RC6_MCE:
		return "CODEC_RC6_MCE"
	case CODEC_SHARP:
		return "CODEC_SHARP"
	case CODEC_PANASONIC:
		return "CODEC_PANASONIC"
	default:
		return "[?? Invalid CodecType value]"
	}
}
