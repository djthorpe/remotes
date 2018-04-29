# IR Sending and Receiving

This repository contains software which provides the ability to send 
and receive infrared signals for a variety of devices usually controlled
using Infrared Remote Controllers (TV's, CD players, etc)
on a Raspberry Pi. In order to use it, you'll need some IR Sending and 
Receiving hardware, which you can build yourself (more details below) or
purchase.

The types of remote controls supported may have one of the following encoding
schemes:

  * Sony 12- 15- and 20- bit (CODEC_SONY12,CODEC_SONY15,CODEC_SONY20)
  * Panasonic (CODEC_PANASONIC)
  * NEC 32- and 16- bit (CODEC_NEC32, CODEC_NEC16 and CODEC_APPLETV)
  * Philips RC5 (CODEC_RC5) __In development__

It's fairly easy to add other encoding schemes, please see the Appendix below.
There is some software available to interact with your remotes:

  * `ir_learn` can be used to learn a new remote or update an existing remote
    in the database of "key mappings"
  * `ir_rcv` can be used for debugging remotes
  * `ir_send` can be used for sending commands

In addition there are a couple of microservice binaries which allow remote
services and clients to interact remotely through [gRPC](https://grpc.io/):

  * `remotes-service` is a microservice allowing remotes to be interacted with
    remotely through the gRPC protocol
  * `remotes-client` is an example command-line client which can interact with
    the microservice through the gRPC protocol

Ultimately the mircoservice should form a larger service for home automation or 
could be used to develop a web-accessible application; this is being developed elsewhere.

## License

Please see the [LICENSE](https://github.com/djthorpe/remotes/blob/master/LICENSE)
file for how to redistribute in source or binary form. Ultimately you should
credit the authors on redistribution as per paragraph four of that license.

## Feedback

All feedback gratefully received, either as bugs, features or questions. For the
former, please use the GitHub issue tracker. For the latter, send me an email
which is listed in the GitHub [repository](https://github.com/djthorpe/remotes).

# Installation

## Hardware Installation

You'll need an IR sender and/or receiver plugged into your GPIO port. There
is a schematic and bill of materials below, plus a link to manufacture
a PCB, if you want to make one up, or you can easily purchase one or breadboard
it.

The hardware has been tested on a modern Raspbian installation but ultimately 
any Linux which is compiled with the `lirc` module should work fine. For a 
Raspberry Pi, you should add this to your `/boot/config.txt` file modifying
the GPIO pins in order to load the LIRC (Linux Infrared Control) driver, 
and then reboot your Raspberry Pi:

```
dtoverlay=lirc-rpi,gpio_in_pin=22,gpio_out_pin=23
```

Your LIRC should then be able to see the device `/dev/lirc0`. If not check output 
of the `lsmod` command.  The best reference for how to interact with the device at the low 
level is [here](https://www.kernel.org/doc/html/latest/media/uapi/rc/lirc-dev-intro.html).

## Software Installation

You'll need a working Go Language environment to compile and install the software. There 
are a few dependencies, for example gRPC and Protocol Buffers, if you're wanting to use the
microservice:

```
bash% go get github.com/djthorpe/remotes
bash% cd ${GOPATH}/src/github.com/djthorpe/remotes
bash% cmd/build-tool.sh # To install the command-line utilities
```

This will result in a number of binaries installed in `${GOBIN}`: `ir_learn`, 
`ir_rcv` and `ir_send` which are described below. For microservices installation
on Raspberry Pi you can download and install Protocol Buffers, gRPC and then 
install the binaries:

```
bash% sudo apt install protobuf-compiler
bash% sudo apt install libprotobuf-dev
bash% go get -u github.com/golang/protobuf/protoc-gen-go
bash% cd ${GOPATH}/src/github.com/djthorpe/remotes
bash% cmd/build-rpc.sh # To install the RPC binaries
```

This will result in the installation of the service `remotes-service`and and example
command-line client `remotes-client`. More information on these binaries below.

# Usage

## Running Command-Line Tools

Before you use the command-line tools, you will need to create a folder at
`/var/local/remotes` which will be used for storing the key mapping 
database files for the remotes, which is stored in XML format with the file
extension `.keymap`. The command-line tools are invoked as follows:

```
  ir_rcv <common flags>
  ir_learn <command flags> -device <device_name> -key <key_list>
  ir_send <command flags> -device <device_name> <key_list>  
```

You can use the following optional common flags with all the binaries:

```
  -debug
    	Set debugging mode
  -verbose
    	Verbose logging
  -log.append
    	When writing log to file, append output to end of file
  -log.file string
    	File for logging (default: log to stderr)
  -lirc.device string
      	LIRC device (default "/dev/lirc0")
  -keymap.db string
    	Key mapping database path (default "/var/local/remotes")
  -keymap.ext string
    	Key mapping file extension  (default ".keymap")         
```

Here is a detailed description of how to use each tool. You can check to see if the
software is working with the following command:

```
bash% ir_rcv
```

This will display a table of learnt keys and also display scan codes which have yet to be 
matched (marked as `<unmapped>`.) This is what some example output might look like:

```
Name                 Key                       Scancode   Device     Codec           Event                  Timestamp
-------------------- ------------------------- ---------- ---------- --------------- ---------------------- ----------
Menu                 KEYCODE_MENU              0x00000040 0x0000009F CODEC_APPLETV   INPUT_EVENT_KEYPRESS   2.117s
Eject                KEYCODE_EJECT             0x0000808D 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYPRESS   11.643s
Eject                KEYCODE_EJECT             0x0000808D 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYREPEAT  11.773s
Pad 3                KEYCODE_KEYPAD_3          0x00004845 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYPRESS   12.425s
Pad 3                KEYCODE_KEYPAD_3          0x00004845 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYREPEAT  12.555s
Nav Right            KEYCODE_NAV_RIGHT         0x0000111C 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYPRESS   13.925s
Nav Right            KEYCODE_NAV_RIGHT         0x0000111C 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYREPEAT  14.056s
<unmapped>           <unmapped>                0x0000CCC1 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYPRESS   14.97s
<unmapped>           <unmapped>                0x0000CCC1 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYREPEAT  15.102s
<unmapped>           <unmapped>                0x0000CCC1 0x40040D00 CODEC_PANASONIC INPUT_EVENT_KEYREPEAT  15.929s
Pad 2                KEYCODE_KEYPAD_2          0x00000008 0x00000076 CODEC_NEC32     INPUT_EVENT_KEYPRESS   18.98s
Pad 2                KEYCODE_KEYPAD_2          0x00000008 0x00000076 CODEC_NEC32     INPUT_EVENT_KEYPRESS   19.231s
Pause                KEYCODE_PAUSE             0x00000020 0x00000076 CODEC_NEC32     INPUT_EVENT_KEYPRESS   21.628s
<unmapped>           <unmapped>                0x000000C0 0x00000076 CODEC_NEC32     INPUT_EVENT_KEYPRESS   22.52s
```

Once you've finished using `ir_rcv` press CTRL+C to quit the software.

To learn a new or existing remote, use the `ir_learn` command-line tool as follows:

```
  bash% ir_learn -device "DVD Player"
```

This will cycle through all keys and will prompt you to press a key on your remote to map it.
If you don't press the key within a few seconds, no mapping is created and the next key is
learnt. Press the CTRL+C combination in order to abort the learning without saving.

You can learn specific keys by using the `-key` flag. For example, to learn the keypad digits
and the navigation buttons use:

```
  bash% ir_learn -device "DVD Player" -key keypad,nav,play,pause,stop
```

You can re-invoke the tool with the same device name to modify existing key mappings. Once the
remote device has been learnt, you can use the `ir_send` utility to check the keymapping database 
and send codes to your device:

```
bash% ir_send
KEY                  CODE                      CODEC             DEVICE     SCANCODE   REPEATS
-------------------- ------------------------- ----------------- ---------- ---------- -------
appletv              CODEC_APPLETV        0x0000009F       6       0
squeezebox           CODEC_NEC32          0x00000076      27       0
dvd                  CODEC_PANASONIC      0x40040D00      26       0

bash% ir_send -device dvd
KEY                  CODE                      CODEC             DEVICE     SCANCODE   REPEATS
-------------------- ------------------------- ----------------- ---------- ---------- -------
Nav Back             KEYCODE_NAV_BACK          CODEC_PANASONIC   0x40040D00 0x0000818C       0
Nav Down             KEYCODE_NAV_DOWN          CODEC_PANASONIC   0x40040D00 0x0000616C       0
Nav Left             KEYCODE_NAV_LEFT          CODEC_PANASONIC   0x40040D00 0x0000E1EC       0
Pad 0                KEYCODE_KEYPAD_0          CODEC_PANASONIC   0x40040D00 0x00009895       0
Pad 2                KEYCODE_KEYPAD_2          CODEC_PANASONIC   0x40040D00 0x00008885       0
Pad 7                KEYCODE_KEYPAD_7          CODEC_PANASONIC   0x40040D00 0x00006865       0
Search Left          KEYCODE_SEARCH_LEFT       CODEC_PANASONIC   0x40040D00 0x0000202D       0
Power Toggle         KEYCODE_POWER_TOGGLE      CODEC_PANASONIC   0x40040D00 0x0000BCB1       0
Sleep                KEYCODE_SLEEP             CODEC_PANASONIC   0x40040D00 0x0000D7DA       0
```

If you invoke `ir_send` without any arguments, it will display a list of learnt devices. If you invoke
it with a `-device` flag it should display a list of learnt keys. Finally you can invoke it with one
or more arguments to send an IR command. If you have some ambigious key names, then you'll need to
modify what you use on the command line to specify a key more exactly. For example,

```
bash% ir_send -device "appletv" volume
Ambiguous key: volume (It could mean one of 'Volume Down','Volume Up')

bash% ir_send -device "appletv" volume_down
KEY                  CODE                      CODEC             DEVICE     SCANCODE   REPEATS
-------------------- ------------------------- ----------------- ---------- ---------- -------
Volume Down          KEYCODE_VOLUME_DOWN       CODEC_APPLETV     0x0000009F 0x000000B0       0
```

There are some enhancements for the command-line utilities such as being able to send a command
more than once (the "Repeats") and cleaning up the database, all of which can be done by editing
the XML files under `/var/local/remotes` if you want to do it manually.

## Running Microservices

__The microservices are still under development, so this is preliminary documentation__

There are some extra flags for the remotes service:

```
  -rpc.port uint
    	Server Port
  -rpc.sslcert string
    	SSL Certificate Path
  -rpc.sslkey string
    	SSL Key Path
```

If you don't specify a port, a random unused one is chosen. You should run the command with the `-verbose`
flag to determine which port is being used. If you don't specify the SSL certificate or key files, the 
service runs insecurely.

You can examine the Protocol Buffer definition [here](https://raw.githubusercontent.com/djthorpe/remotes/master/protobuf/remotes)
and develop your own client software for interacting with it, or you can use the `remotes-client` software. The flags for the
client connecting to the server are as follows:

```
  -rpc.addr string
    	Server Address
  -rpc.insecure
    	Disable SSL Connection
  -rpc.skipverify
    	Skip SSL Verification (default true)
  -rpc.timeout duration
    	Connection timeout
```

There are three commands that the client allows:

  * `codecs` Lists the encodings registered with the service
  * `receive` Starts receiving keycodes from the service. Press CTRL+C to stop
  * `send` Transmits an IR code

More information on the client soon, since this is mostly still in devlopment.


# Appendix 

## Schematic

Here is the schematic of the circuit with the bill of materials:

![IR Schematic](https://raw.githubusercontent.com/djthorpe/remotes/master/etc/ir_schematic.png)


If you want to make a PCB of this design you can [manufacturer one here from Aisler](https://aisler.net/djthorpe/djthorpe/raspberry-pi-ir-sender-receiver). Total cost
including components would cost about £5 / €5 / $5 per unit. Here's a picture of what the populated PCB looks like:

![PCB](https://raw.githubusercontent.com/djthorpe/remotes/master/etc/aisler-pcb-201802.jpg)

This is the bill of materials:

| Name | Part                  | Description |
| ---- | ---- | ---- |
| D1   |  Vishay TSAL6200      | 940nm IR LED, 5mm (T-1 3/4) Through Hole package |
| D2   |  Vishay TSOP38238     | 38kHz IR Receiver, 950nm, 45m Range, Through Hole, 5 x 4.8 x 6.95mm |
| Q1   |  Fairchild KSP2222ABU | NPN Transistor, 600 mA, 40 V, 3-Pin TO-92 |
| R1   |  680Ω ±5% 0.25W       | Carbon Resistor, 0.25W, 5%, 680R |
| R2   |  10K  ±5% 0.25W       | Carbon Resistor, 0.25W, 5%, 10K |
| R3   |  36Ω ±1% 0.6W         | Carbon Resistor, 0.6W, 1%, 36R |
| J1   |  26 Way PCB Header    | 2.54mm Pitch 13x2 Rows Straight PCB Socket |

## Adding in more codecs

You can add in any number of Codecs which respond to LIRC pulse/space messages.
In order to do so, I would firstly understand how to develop a 'module' for the 
`gopi` framework. When registering the module in your `init` function, use
module type `gopi.MODULE_TYPE_OTHER` and the module should have the name 
prefix `remotes/`. You will want a dependency on the `lirc` module. Here is what
a typical `init()` function may contain:

```
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/panasonic",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
			}, app.Logger)
		},
	}) 
```

Secondly, your module should subscribe to events from the LIRC module
instance, and process them in the background. You'll receive a stream
of LIRC events:

```
type LIRCEvent interface {
	Event

	// The type of message
	Type() LIRCType

	// The value
	Value() uint32
}
```

The LIRC type will be one of `LIRC_TYPE_SPACE`,`LIRC_TYPE_PULSE`,
`LIRC_TYPE_FREQUENCY`or `LIRC_TYPE_TIMEOUT`. Practically the Raspberry
Pi module only supports pulses and spaces and the value is measured
in nanoseconds. For receiving and recognizing commands, it's perhaps
easiest to develop a state machine which responds to and recognizes
commands, or resets back to the initial state if not recognized. Your
codec should then emit a `RemoteEvent` structure on recognition:

```
type RemoteEvent struct {
	source   Codec           // Codec that emitted the event
	ts       time.Duration   // Rolling timecode
	device   uint32          // Unique device identifier, or 0
	scancode uint32          // Scancode identifying command
	repeat   bool            // Whether this is a repeat code
}
```

To send a command, you need to implement the following function. The
`repeats` value should be zero to send a single command, or greater than
zero to send one or more repeats for the command.

```
func (this *codec) Send(device uint32, scancode uint32, repeats uint) error {
	var pulses[]uint32

	// Send a pulse and a space
	pulses = append(pulses, pulse_nanoseconds,space_nanoseconds)

  // ...Continue to send pulses and spaces and end on a pulse...

	// Perform the pulse send
	return this.lirc.PulseSend(pulses)
}
```

You can of course see the examples in the repository. The `codec` folder contains
all of the implemented codecs so far. If you want to include your codec here, send
me a pull request!
