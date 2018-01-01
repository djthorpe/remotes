/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"github.com/djthorpe/gopi"
)

// A coder/decoder for remote
type Codec interface {
	gopi.Driver

	// Return short name for the codec
	Name() string

	// Reset codec to initial state
	Reset()
}
