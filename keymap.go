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
	"os"
	"strings"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Remote struct {
	XMLName xml.Name  `xml:"remote"`
	Device  uint32    `xml:"id,attr,omitempty"`
	Type    CodecType `xml:"codec"`
	Name    string    `xml:"name"`
	Repeats uint      `xml:"repeats"`
	Map     []*KeyMap `xml:"keymap"`
}

type KeyMap struct {
	Scancode uint32     `xml:"scancode"`
	Keycode  RemoteCode `xml:"keycode"`
	Name     string     `xml:"name"`
	Device   uint32     `xml:"id,attr,omitempty"` // Overrides device if non-zero
	Type     CodecType  `xml:"codec,omitempty"`   // Overrides codec if non-zero
	Repeats  uint       `xml:"repeats,omitempty"` // Overrides repeats if non-zero
}

/////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES AND CONSTANTS

var (
	ErrInvalidKey = errors.New("Invalid Key")
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// NewRemote returns a new empty remote
func NewRemote(codec CodecType, device uint32) *Remote {
	this := new(Remote)
	this.Type = codec
	this.Device = device
	this.Name = this.defaultName()
	this.Map = make([]*KeyMap, 0)
	return this
}

// RemoteFromFile returns a loaded remote
func RemoteFromFile(filename string) (*Remote, error) {
	var this Remote
	if fh, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		defer fh.Close()
		dec := xml.NewDecoder(fh)
		if err := dec.Decode(&this); err != nil {
			return nil, err
		} else {
			return &this, nil
		}
	}
}

// SetName names the device
func (this *Remote) SetName(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		this.Name = this.defaultName()
	} else {
		this.Name = name
	}
}

// SetKey assigns a keycode for a scancode
func (this *Remote) SetKey(keycode RemoteCode, scancode uint32, name string) error {
	name = strings.TrimSpace(name)
	if keyname := this.defaultKeyName(keycode); keyname == "" {
		return ErrInvalidKey
	} else if name == "" {
		name = keyname
	}
	this.Map = append(this.Map, &KeyMap{
		Scancode: scancode,
		Keycode:  keycode,
		Name:     name,
	})
	return nil
}

// Save the information to a file
func (this *Remote) SaveToFile(filepath string) error {
	var fh *os.File

	/* Open file */
	if filepath == "" || filepath == "-" {
		fh = os.Stdout
	} else {
		var err error
		if fh, err = os.Create(filepath); err != nil {
			return err
		}
	}
	defer fh.Close()

	/* Encode XML */
	enc := xml.NewEncoder(fh)
	enc.Indent("", "  ")
	if err := enc.Encode(this); err != nil {
		return err
	} else {
		return nil
	}
}

// Returns all codes
func Keycodes() []*KeyMap {
	keycodes := make([]*KeyMap, 0)
	for c := KEYCODE_NONE; c < KEYCODE_MAX; c++ {
		if name := fmt.Sprint(c); strings.HasPrefix(name, "KEYCODE_") {
			// Convert name into more English-language
			name = strings.ToLower(strings.TrimPrefix(name, "KEYCODE_"))
			name = strings.Title(strings.Replace(name, "_", " ", -1))
			// Append keycode
			keycodes = append(keycodes, &KeyMap{Keycode: c, Name: name})
		}
	}
	return keycodes
}

// Returns all codes which match a string or nil if nothing matches
func KeycodesForString(token string) []*KeyMap {
	selected := make(map[RemoteCode]*KeyMap)

	// Firstly, uppercase the token
	token = strings.ToUpper(token)

	// Iterate through all the keycodes
	for _, k := range Keycodes() {
		if strings.HasPrefix(token, "KEYCODE_") {
			if strings.HasPrefix(fmt.Sprint(k.Keycode), token) {
				selected[k.Keycode] = k
			}
		} else {
			search_tokens := strings.Split(strings.ToUpper(k.Name), " ")
			other_tokens := strings.TrimPrefix(strings.ToUpper(fmt.Sprint(k.Keycode)), "KEYCODE_")
			search_tokens = append(search_tokens, other_tokens)
			search_tokens = append(search_tokens, strings.Split(other_tokens, "_")...)
			for _, t := range search_tokens {
				if token == t {
					selected[k.Keycode] = k
				}
			}
		}
	}

	// Return if nothing selected
	if len(selected) == 0 {
		return nil
	}

	// De-duplicate keys
	keycodes := make([]*KeyMap, 0, len(selected))
	for _, v := range selected {
		keycodes = append(keycodes, v)
	}
	return keycodes
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *KeyMap) String() string {
	params := fmt.Sprintf("keycode=%v ", this.Keycode)
	params += fmt.Sprintf("scancode=0x%X ", this.Scancode)
	if this.Name != "" {
		params += fmt.Sprintf("name=\"%v\" ", this.Name)
	}
	return fmt.Sprintf("<remotes.KeyMap>{ %v}", params)
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *Remote) defaultName() string {
	if this.Device == 0 {
		return fmt.Sprintf("%v", this.Type)
	} else {
		return fmt.Sprintf("%v[%X]", this.Type, this.Device)
	}
}

func (this *Remote) defaultKeyName(keycode RemoteCode) string {
	if name := fmt.Sprint(keycode); strings.HasPrefix(name, "KEYCODE_") == false {
		// Invalid key
		return ""
	} else {
		name = strings.TrimLeft(name, "KEYCODE_")
		name = strings.Replace(name, "_", " ", -1)
		return strings.Title(strings.ToLower(name))
	}
}
