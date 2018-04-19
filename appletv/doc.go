/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package appletv

// Reference:
//   https://gist.github.com/darconeous/4437f79a34e3b6441628
//
// Apple Remote IR Code
//
// This document covers the old white apple remote.
// Carrier: ~38KHz
// Start: 9ms high, 4.5ms low
// Pulse width: ~0.58ms (~853Hz)
// Uses pulse-distance modulation.
//
// Bit encoding:
//   0: 1 pulse-width high, 1 pulse-width low
//   1: 1 pulse-width high, 3 pulse-widths low
//   4 octets are transmitted, LSB first.
//   First two octets in normal transmission are 0x77 0xE1. (Different for pair command, which is 0x07 0xE1) Third octet is command. Fourth octet is remote ID.
//   One of these bits is used for the low-battery indication. I haven't yet identified which one.
//
// Example codes (in transmission order):
//
// 01110111 11100001 01000000 11101011 MENU
// 01110111 11100001 10110000 11101011 VOL-
// 01110111 11100001 11010000 11101011 VOL+
// 01110111 11100001 00100000 11101011 PLAY
// 01110111 11100001 11100000 11101011 NEXT
// 01110111 11100001 00010000 11101011 PREV
// 01110111 11100001 01101000 11101011 ??? (MENU+VOLUP)
// 00000111 11100001 11000000 11101011 PAIR (MENU+NEXT)
//
// Normal Command format: 000XXXXP, where XXXX is the command and P is a parity bit(even parity).
//
// Commands (MSB First):
//
// 0x01: 000 0001 0 MENU
// 0x02: 000 0010 0 PLAY
// 0x03: 000 0011 1 NEXT
// 0x04: 000 0100 0 PREV
// 0x05: 000 0101 1 VOL+
// 0x06: 000 0110 1 VOL-
// 0x0B: 000 1011 0 ???? (MENU+VOLUP)
// 0x0C: 000 1100 1 ???? (MENU+VOLDOWN)
// Special Commands:
//
// MENU+NEXT : Pair
// MENU+PLAY : Increment Remote Code
// MENU+PREV : ???
//
// As far as I can tell, there are 8 normal signals that the remote can emit,
// and 3 special signals---for a total of 11 signals.
//
// Different remote
// 01110111 11100001 00100000 01010000 PLAY
// 00000111 11100001 11000000 01010000 PAIR
// 00000111 11100001 10100000 01010000 ??? (MENU+PREV)
// 00000111 11100001 01000000 11010000 CHANGE CODE (MENU+PLAY)
// 01110111 11100001 10011000 00110000 ??? (MENU+VOLDOWN)
