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
	"io"
)

type KeyMap struct {
	Scancode uint32
	Keycode  string
	Key      string
}

type DeviceMap struct {
	Device string
	KeyMap []*KeyMap
}

func NewDeviceMap(device string) *DeviceMap {
	this := new(DeviceMap)
	this.Device = device
	this.KeyMap = make([]*KeyMap, 0)
	return this
}

func (this *DeviceMap) Write(writer io.Writer) error {
	enc := xml.NewEncoder(writer)
	enc.Indent("  ", "    ")
	if err := enc.Encode(this); err != nil {
		return err
	} else {
		return nil
	}
}
