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
// CONSTANTS

const (
	KEYCODE_MIN RemoteCode = iota + RemoteCode(gopi.KEYCODE_MAX)
	KEYCODE_EJECT
	KEYCODE_POWER_OFF
	KEYCODE_POWER_ON
	KEYCODE_INPUT_SELECT
	KEYCODE_INPUT_PC
	KEYCODE_INPUT_VIDEO1
	KEYCODE_INPUT_VIDEO2
	KEYCODE_INPUT_VIDEO3
	KEYCODE_INPUT_VIDEO4
	KEYCODE_INPUT_VIDEO5
	KEYCODE_INPUT_HDMI1
	KEYCODE_INPUT_HDMI2
	KEYCODE_INPUT_HDMI3
	KEYCODE_INPUT_HDMI4
	KEYCODE_INPUT_HDMI5
	KEYCODE_INPUT_AUX1
	KEYCODE_INPUT_AUX2
	KEYCODE_INPUT_AUX3
	KEYCODE_INPUT_AUX4
	KEYCODE_INPUT_AUX5
	KEYCODE_INPUT_CD
	KEYCODE_INPUT_DVD
	KEYCODE_INPUT_PHONO
	KEYCODE_INPUT_TAPE1
	KEYCODE_INPUT_TAPE2
	KEYCODE_INPUT_TUNER
	KEYCODE_INPUT_ANALOG
	KEYCODE_INPUT_DIGITAL
	KEYCODE_INPUT_INTERNET
	KEYCODE_INPUT_TEXT
	KEYCODE_INPUT_NEXT
	KEYCODE_INPUT_PREV
	KEYCODE_VIDEO_ASPECT
	KEYCODE_VIDEO_PIP
	KEYCODE_AUDIO_MONITOR
	KEYCODE_CLEAR
	KEYCODE_TIMER
	KEYCODE_CHANNEL_PREV
	KEYCODE_CHANNEL_GUIDE
	KEYCODE_RECORD
	KEYCODE_RECORD_SPEED
	KEYCODE_PLAY_SPEED
	KEYCODE_PLAY_MODE
	KEYCODE_REPLAY
	KEYCODE_DISPLAY
	KEYCODE_MENU
	KEYCODE_INFO
	KEYCODE_HOME
	KEYCODE_THUMBS_UP
	KEYCODE_THUMBS_DOWN
	KEYCODE_FAVOURITE
	KEYCODE_BUTTON_RED
	KEYCODE_BUTTON_GREEN
	KEYCODE_BUTTON_YELLOW
	KEYCODE_BUTTON_BLUE
	KEYCODE_SEARCH_LEFT
	KEYCODE_SEARCH_RIGHT
	KEYCODE_CHAPTER_NEXT
	KEYCODE_CHAPTER_PREV
	KEYCODE_NAV_SELECT
	KEYCODE_SUBTITLE_TOGGLE
	KEYCODE_SUBTITLE_ON
	KEYCODE_SUBTITLE_OFF
	KEYCODE_STOP
	KEYCODE_PAUSE
	KEYCODE_SLEEP
	KEYCODE_BROWSE
	KEYCODE_SHUFFLE
	KEYCODE_REPEAT
	KEYCODE_KEYPAD_10PLUS
	KEYCODE_MAX
)

const (
	KEYCODE_NONE             = RemoteCode(gopi.KEYCODE_NONE)
	KEYCODE_POWER_TOGGLE     = RemoteCode(gopi.KEYCODE_POWER)
	KEYCODE_KEYPAD_1         = RemoteCode(gopi.KEYCODE_KP1)
	KEYCODE_KEYPAD_2         = RemoteCode(gopi.KEYCODE_KP2)
	KEYCODE_KEYPAD_3         = RemoteCode(gopi.KEYCODE_KP3)
	KEYCODE_KEYPAD_4         = RemoteCode(gopi.KEYCODE_KP4)
	KEYCODE_KEYPAD_5         = RemoteCode(gopi.KEYCODE_KP5)
	KEYCODE_KEYPAD_6         = RemoteCode(gopi.KEYCODE_KP6)
	KEYCODE_KEYPAD_7         = RemoteCode(gopi.KEYCODE_KP7)
	KEYCODE_KEYPAD_8         = RemoteCode(gopi.KEYCODE_KP8)
	KEYCODE_KEYPAD_9         = RemoteCode(gopi.KEYCODE_KP9)
	KEYCODE_KEYPAD_0         = RemoteCode(gopi.KEYCODE_KP0)
	KEYCODE_KEYPAD_SELECT    = RemoteCode(gopi.KEYCODE_KPENTER)
	KEYCODE_VOLUME_UP        = RemoteCode(gopi.KEYCODE_VOLUMEUP)
	KEYCODE_VOLUME_DOWN      = RemoteCode(gopi.KEYCODE_VOLUMEDOWN)
	KEYCODE_VOLUME_MUTE      = RemoteCode(gopi.KEYCODE_MUTE)
	KEYCODE_CHANNEL_UP       = RemoteCode(gopi.KEYCODE_PAGEUP)
	KEYCODE_CHANNEL_DOWN     = RemoteCode(gopi.KEYCODE_PAGEDOWN)
	KEYCODE_NAV_UP           = RemoteCode(gopi.KEYCODE_UP)
	KEYCODE_NAV_DOWN         = RemoteCode(gopi.KEYCODE_DOWN)
	KEYCODE_NAV_LEFT         = RemoteCode(gopi.KEYCODE_LEFT)
	KEYCODE_NAV_RIGHT        = RemoteCode(gopi.KEYCODE_RIGHT)
	KEYCODE_NAV_BACK         = RemoteCode(gopi.KEYCODE_CANCEL)
	KEYCODE_PLAY             = RemoteCode(gopi.KEYCODE_PLAY)
	KEYCODE_ADD              = RemoteCode(gopi.KEYCODE_KPPLUS)
	KEYCODE_SEARCH           = RemoteCode(gopi.KEYCODE_SEARCH)
	KEYCODE_BRIGHTNESS_CYCLE = RemoteCode(gopi.KEYCODE_BRIGHTNESS_CYCLE)
)

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (k RemoteCode) String() string {
	switch k {
	case KEYCODE_BRIGHTNESS_CYCLE:
		return "KEYCODE_BRIGHTNESS_CYCLE"
	case KEYCODE_SHUFFLE:
		return "KEYCODE_SHUFFLE"
	case KEYCODE_REPEAT:
		return "KEYCODE_REPEAT"
	case KEYCODE_SLEEP:
		return "KEYCODE_SLEEP"
	case KEYCODE_SEARCH:
		return "KEYCODE_SEARCH"
	case KEYCODE_BROWSE:
		return "KEYCODE_BROWSE"
	case KEYCODE_ADD:
		return "KEYCODE_ADD"
	case KEYCODE_PLAY:
		return "KEYCODE_PLAY"
	case KEYCODE_PAUSE:
		return "KEYCODE_PAUSE"
	case KEYCODE_STOP:
		return "KEYCODE_STOP"
	case KEYCODE_EJECT:
		return "KEYCODE_EJECT"
	case KEYCODE_POWER_OFF:
		return "KEYCODE_POWER_OFF"
	case KEYCODE_POWER_ON:
		return "KEYCODE_POWER_ON"
	case KEYCODE_POWER_TOGGLE:
		return "KEYCODE_POWER_TOGGLE"
	case KEYCODE_INPUT_SELECT:
		return "KEYCODE_INPUT_SELECT"
	case KEYCODE_INPUT_PC:
		return "KEYCODE_INPUT_PC"
	case KEYCODE_INPUT_VIDEO1:
		return "KEYCODE_INPUT_VIDEO1"
	case KEYCODE_INPUT_VIDEO2:
		return "KEYCODE_INPUT_VIDEO2"
	case KEYCODE_INPUT_VIDEO3:
		return "KEYCODE_INPUT_VIDEO3"
	case KEYCODE_INPUT_VIDEO4:
		return "KEYCODE_INPUT_VIDEO4"
	case KEYCODE_INPUT_VIDEO5:
		return "KEYCODE_INPUT_VIDEO5"
	case KEYCODE_INPUT_HDMI1:
		return "KEYCODE_INPUT_HDMI1"
	case KEYCODE_INPUT_HDMI2:
		return "KEYCODE_INPUT_HDMI2"
	case KEYCODE_INPUT_HDMI3:
		return "KEYCODE_INPUT_HDMI3"
	case KEYCODE_INPUT_HDMI4:
		return "KEYCODE_INPUT_HDMI4"
	case KEYCODE_INPUT_HDMI5:
		return "KEYCODE_INPUT_HDMI5"
	case KEYCODE_INPUT_AUX1:
		return "KEYCODE_INPUT_AUX1"
	case KEYCODE_INPUT_AUX2:
		return "KEYCODE_INPUT_AUX2"
	case KEYCODE_INPUT_AUX3:
		return "KEYCODE_INPUT_AUX3"
	case KEYCODE_INPUT_AUX4:
		return "KEYCODE_INPUT_AUX4"
	case KEYCODE_INPUT_AUX5:
		return "KEYCODE_INPUT_AUX5"
	case KEYCODE_INPUT_CD:
		return "KEYCODE_INPUT_CD"
	case KEYCODE_INPUT_DVD:
		return "KEYCODE_INPUT_DVD"
	case KEYCODE_INPUT_PHONO:
		return "KEYCODE_INPUT_PHONO"
	case KEYCODE_INPUT_TAPE1:
		return "KEYCODE_INPUT_TAPE1"
	case KEYCODE_INPUT_TAPE2:
		return "KEYCODE_INPUT_TAPE2"
	case KEYCODE_INPUT_TUNER:
		return "KEYCODE_INPUT_TUNER"
	case KEYCODE_INPUT_ANALOG:
		return "KEYCODE_INPUT_ANALOG"
	case KEYCODE_INPUT_DIGITAL:
		return "KEYCODE_INPUT_DIGITAL"
	case KEYCODE_INPUT_INTERNET:
		return "KEYCODE_INPUT_INTERNET"
	case KEYCODE_INPUT_TEXT:
		return "KEYCODE_INPUT_TEXT"
	case KEYCODE_INPUT_NEXT:
		return "KEYCODE_INPUT_NEXT"
	case KEYCODE_INPUT_PREV:
		return "KEYCODE_INPUT_PREV"
	case KEYCODE_VIDEO_ASPECT:
		return "KEYCODE_VIDEO_ASPECT"
	case KEYCODE_VIDEO_PIP:
		return "KEYCODE_VIDEO_PIP"
	case KEYCODE_AUDIO_MONITOR:
		return "KEYCODE_AUDIO_MONITOR"
	case KEYCODE_KEYPAD_1:
		return "KEYCODE_KEYPAD_1"
	case KEYCODE_KEYPAD_2:
		return "KEYCODE_KEYPAD_2"
	case KEYCODE_KEYPAD_3:
		return "KEYCODE_KEYPAD_3"
	case KEYCODE_KEYPAD_4:
		return "KEYCODE_KEYPAD_4"
	case KEYCODE_KEYPAD_5:
		return "KEYCODE_KEYPAD_5"
	case KEYCODE_KEYPAD_6:
		return "KEYCODE_KEYPAD_6"
	case KEYCODE_KEYPAD_7:
		return "KEYCODE_KEYPAD_7"
	case KEYCODE_KEYPAD_8:
		return "KEYCODE_KEYPAD_8"
	case KEYCODE_KEYPAD_9:
		return "KEYCODE_KEYPAD_9"
	case KEYCODE_KEYPAD_0:
		return "KEYCODE_KEYPAD_0"
	case KEYCODE_KEYPAD_10PLUS:
		return "KEYCODE_KEYPAD_10PLUS"
	case KEYCODE_KEYPAD_SELECT:
		return "KEYCODE_KEYPAD_SELECT"
	case KEYCODE_CLEAR:
		return "KEYCODE_CLEAR"
	case KEYCODE_TIMER:
		return "KEYCODE_TIMER"
	case KEYCODE_VOLUME_UP:
		return "KEYCODE_VOLUME_UP"
	case KEYCODE_VOLUME_DOWN:
		return "KEYCODE_VOLUME_DOWN"
	case KEYCODE_VOLUME_MUTE:
		return "KEYCODE_VOLUME_MUTE"
	case KEYCODE_CHANNEL_UP:
		return "KEYCODE_CHANNEL_UP"
	case KEYCODE_CHANNEL_DOWN:
		return "KEYCODE_CHANNEL_DOWN"
	case KEYCODE_CHANNEL_PREV:
		return "KEYCODE_CHANNEL_PREV"
	case KEYCODE_CHANNEL_GUIDE:
		return "KEYCODE_CHANNEL_GUIDE"
	case KEYCODE_RECORD:
		return "KEYCODE_RECORD"
	case KEYCODE_RECORD_SPEED:
		return "KEYCODE_RECORD_SPEED"
	case KEYCODE_PLAY_SPEED:
		return "KEYCODE_PLAY_SPEED"
	case KEYCODE_PLAY_MODE:
		return "KEYCODE_PLAY_MODE"
	case KEYCODE_REPLAY:
		return "KEYCODE_REPLAY"
	case KEYCODE_DISPLAY:
		return "KEYCODE_DISPLAY"
	case KEYCODE_MENU:
		return "KEYCODE_MENU"
	case KEYCODE_HOME:
		return "KEYCODE_HOME"
	case KEYCODE_INFO:
		return "KEYCODE_INFO"
	case KEYCODE_THUMBS_UP:
		return "KEYCODE_THUMBS_UP"
	case KEYCODE_THUMBS_DOWN:
		return "KEYCODE_THUMBS_DOWN"
	case KEYCODE_FAVOURITE:
		return "KEYCODE_FAVOURITE"
	case KEYCODE_BUTTON_RED:
		return "KEYCODE_BUTTON_RED"
	case KEYCODE_BUTTON_GREEN:
		return "KEYCODE_BUTTON_GREEN"
	case KEYCODE_BUTTON_YELLOW:
		return "KEYCODE_BUTTON_YELLOW"
	case KEYCODE_BUTTON_BLUE:
		return "KEYCODE_BUTTON_BLUE"
	case KEYCODE_SEARCH_LEFT:
		return "KEYCODE_SEARCH_LEFT"
	case KEYCODE_SEARCH_RIGHT:
		return "KEYCODE_SEARCH_RIGHT"
	case KEYCODE_CHAPTER_NEXT:
		return "KEYCODE_CHAPTER_NEXT"
	case KEYCODE_CHAPTER_PREV:
		return "KEYCODE_CHAPTER_PREV"
	case KEYCODE_NAV_UP:
		return "KEYCODE_NAV_UP"
	case KEYCODE_NAV_DOWN:
		return "KEYCODE_NAV_DOWN"
	case KEYCODE_NAV_LEFT:
		return "KEYCODE_NAV_LEFT"
	case KEYCODE_NAV_RIGHT:
		return "KEYCODE_NAV_RIGHT"
	case KEYCODE_NAV_SELECT:
		return "KEYCODE_NAV_SELECT"
	case KEYCODE_NAV_BACK:
		return "KEYCODE_NAV_BACK"
	case KEYCODE_SUBTITLE_TOGGLE:
		return "KEYCODE_SUBTITLE_TOGGLE"
	case KEYCODE_SUBTITLE_ON:
		return "KEYCODE_SUBTITLE_ON"
	case KEYCODE_SUBTITLE_OFF:
		return "KEYCODE_SUBTITLE_OFF"
	default:
		return "[?? Invalid RemoteCode]"
	}
}
