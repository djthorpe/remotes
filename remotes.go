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

/////////////////////////////////////////////////////////////////////
// TYPES

type (
	RemoteCode gopi.KeyCode
	CodecType  uint
)

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

/////////////////////////////////////////////////////////////////////
// INTERFACE

type Codec interface {
	gopi.Driver
	gopi.Publisher

	// Return type for the codec
	Type() CodecType

	// Send scancode
	Send(device uint32, scancode uint32, repeats uint) error
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

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
	case CODEC_APPLETV:
		return "CODEC_APPLETV"
	case CODEC_PANASONIC:
		return "CODEC_PANASONIC"
	default:
		return "[?? Invalid CodecType value]"
	}
}
