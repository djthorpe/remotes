/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package rc5

// Reference:
//   https://techdocs.altium.com//display/FPGA/Philips+RC5+Infrared+Transmission+Protocol
//
// The Philips RC5 IR transmission protocol uses Manchester encoding of the message bits. Each
// pulse burst (mark – RC transmitter ON) is 889us in length, at a carrier frequency of 36kHz (27.7us).
// Logical bits are transmitted as follows:
//
//   * Logical '0' – an 889us pulse burst followed by an 889us space, with a total
//     transmit time of 1.778ms
//   * Logical '1' – an 889us space followed by an 889us pulse burst, with a total
//     transmit time of 1.778ms
//
// When a key is pressed on the remote controller, the message frame transmitted consists of
// the following 14 bits, in order:
//
//   * Two Start bits (S1 and S2), both logical '1'.
//   * A Toggle bit (T). This bit is inverted each time a key is released and pressed again.
//   * The 5-bit address for the receiving device
//   * The 6-bit command.
//
// The address and command bits are each sent most significant bit first. The Toggle bit (T) is
// used by the receiver to distinguish weather the key has been pressed repeatedly, or weather
// it is being held depressed. As long as the key on the remote controller is kept depressed,
// the message frame will be repeated every 114ms. The Toggle bit will retain the same logic
// level during all of these repeated message frames. It is up to the receiver software to interpret
// this auto-repeat feature of the protocol.
