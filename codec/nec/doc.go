/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package nec

// Reference:
//   http://techdocs.altium.com/display/FPGA/NEC+Infrared+Transmission+Protocol
//
// The NEC IR transmission protocol uses pulse distance encoding
// of the message bits. Each pulse burst (mark – RC transmitter ON)
// is 562.5µs in length, at a carrier frequency of 38kHz (26.3µs).
// Logical bits are transmitted as follows:
//
//   * Logical '0' – a 562.5µs pulse burst followed by a 562.5µs space,
//     with a total transmit time of 1.125ms
//   * Logical '1' – a 562.5µs pulse burst followed by a 1.6875ms space,
//     with a total transmit time of 2.25ms
//
// When a key is pressed on the remote controller, the message transmitted
// consists of the following, in order:
//
//  * A 9ms leading pulse burst (16 times the pulse burst length used for
//    a logical data bit)
//  * A 4.5ms space
//  * The 8-bit address for the receiving device
//  * The 8-bit logical inverse of the address
//  * The 8-bit command
//  * The 8-bit logical inverse of the command
//  * A final 562.5µs pulse burst to signify the end of message transmission.
//
// REPEAT CODES
//
// If the key on the remote controller is kept depressed, a repeat code will
// be issued, typically around 40ms after the pulse burst that signified the
// end of the message. A repeat code will continue to be sent out at 108ms
// intervals, until the key is finally released. The repeat code consists of
// the following, in order:
//
//  * A 9ms leading pulse burst
//  * A 2.25ms space
//  * A 562.5µs pulse burst to mark the end of the space (and hence end
//    of the transmitted repeat code).
