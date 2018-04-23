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
	Name    string    `xml:"name"`
	Type    CodecType `xml:"codec"`
	Repeats uint      `xml:"repeats"`
	Map     []*KeyMap `xml:"keymap"`
}

type KeyMap struct {
	Scancode uint32     `xml:"scancode"`
	Keycode  RemoteCode `xml:"keycode"`
	Name     string     `xml:"name"`
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
