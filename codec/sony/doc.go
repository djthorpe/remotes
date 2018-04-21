/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package sony

// Reference:
//   http://users.telenet.be/davshomepage/sony.htm
//   http://picprojects.org.uk/projects/sirc/sonysirc.pdf
//
// The Sony remote control is based on the Pulse-Width signal coding scheme. The code exists of 12 bits
// sent on a 40kHz carrier wave. The code starts with a header of 2.4ms or
// 4 times T where T is 600µS. The header is followed by 7 command bits and 5 address bits.
//
// The address and commands exists of logical ones and zeros. A logical one is formed by
// a space of 600µS or 1T and a pulse of 1200 µS or 2T. A logical zero is formed by a space of 600 µS
// and pulse of 600µS.
//
// The space between 2 transmitted codes when a button is being pressed is 40mS
//
// The bits are transmitted least significant bits first. The total length of a bitstream is always 45ms.
