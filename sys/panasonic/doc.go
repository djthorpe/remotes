/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package panasonic

/*

Not sure there's a source for Panasonic but seems to be:
http://www.remotecentral.com/cgi-bin/mboard/rc-pronto/thread.cgi?26152

  * Header Pulse of 3.5ms, then space of 1.7ms
  * A one bit:
  *   450ns pulse, 1.3ms space
  * A zero bit:
  *   450ns pulse, 450ns space
  * Trail pulse of 450ns, a repeat space is 75ms then repeat the code
  *
  * It's 48 bits long (6 bytes)
  * 02 20 <D1> <D2> <SC> <XOR D1|D2|SC>
  * D1 = Device
  * D2 = Subdevice
  * SC = Scancode
  * XOR = Checksum (XOR of D1,D2 and SC)

*/
