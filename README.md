# IR Sending and Receiving

This repository contains software which provides the ability to send 
and receive infrared signals for a variety of devices usually controlled
using Infrared Remote Controllers (TV's, CD players, etc)
on a Raspberry Pi. In order to use it, you'll need some IR Sending and 
Receiving hardware, which you can build yourself. The schematics and bill of 
materials are listed below.

The types of remote controls supported may have one of the following encoding
schemes:

  * Sony 12-bit (CODEC_SONY12)
  * Sony 15-bit (CODEC_SONY15)
  * Sony 20-bit (CODEC_SONY20)
  * NEC 32 bit (CODEC_NEC32)
  * Panasonic (CODEC_PANASONIC)
  * Legacy Apple TV (CODEC_APPLETV)

It's fairly easy to add other encoding schemes. There is some software available
to interact with your remotes:

  * `ir_learn` can be used to learn a new remote or update an existing remote
    in the database of "key mappings"
  * `ir_rcv` can be used for debugging remotes
  * `remotes-service` is a microservice allowing remotes to be interacted with
    remotely through the gRPC protocol
  * `remotes-client` is a client which can interact with the microservice through
    the gRPC protocol

Ultimately the mircoservice should form a larger service for home automation; this
is being developed elsewhere.

## Hardware Installation

This hardware has been tested on Raspbian Jessie. Any linux which is compiled
with the `lirc` module should work fine. Firstly, you should add this 
to your `/boot/config.txt` file in order to load the LIRC (Linux Infrared Control) 
driver, and then reboot your Raspberry Pi:

```
dtoverlay=lirc-rpi,gpio_in_pin=22,gpio_out_pin=23
```

Your LIRC should then be able to see the device `/dev/lirc0`. The best reference
for how to interact with the device is [here](https://www.kernel.org/doc/html/latest/media/uapi/rc/lirc-dev-intro.html).

## Software Installation

You'll need a working Go Language environment to compile and install the software. There are
a few dependencies, for example gRPC and Protocol Buffers.

```
bash% go get github.com/djthorpe/remotes
bash% cd ${GOPATH}/src/github.com/djthorpe/remotes
bash% cmd/build-cmd.sh # To install the command-line utilities
bash% cmd/build-rpc.sh # To install the RPC binaries
```

This will result in a number of binaries: `ir_learn`, `ir_rcv`, `remotes-service` and `remotes-client`.
Clearly at least one more binary `ir_send` will be developed for sending IR Codes to devices. More
information about these binaries is below.

## Schematic

Here is the schematic of the circuit with the bill of materials:

![IR Schematic](https://raw.githubusercontent.com/djthorpe/remotes/master/etc/ir_schematic.png)

| Name | Part                  | Description |
| ---- | ---- | ---- |
| D1   |  Vishay TSAL6200      | 940nm IR LED, 5mm (T-1 3/4) Through Hole package |
| D2   |  Vishay TSOP38238     | 38kHz IR Receiver, 950nm, 45m Range, Through Hole, 5 x 4.8 x 6.95mm |
| Q1   |  Fairchild KSP2222ABU | NPN Transistor, 600 mA, 40 V, 3-Pin TO-92 |
| R1   |  680Ω ±5% 0.25W       | Carbon Resistor, 0.25W, 5%, 680R |
| R2   |  10K  ±5% 0.25W       | Carbon Resistor, 0.25W, 5%, 10K |
| R3   |  36Ω ±1% 0.6W         | Carbon Resistor, 0.6W, 1%, 36R |
| J1   |  26 Way PCB Header    | 2.54mm Pitch 13x2 Rows Straight PCB Socket |

If you want to make a PCB of this design you can [manufacturer one here from Aisler](https://aisler.net/djthorpe/djthorpe/raspberry-pi-ir-sender-receiver). Total cost
including components would cost about £5 / €5 / $5 per unit.

## Running Command-Line Tools

You can use the following optional command-line flags with all the binaries:

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
      	LIRC device
  -keymap.db string
    	Key mapping database path (default "/var/local/remotes")
  -keymap.ext string
    	Key mapping file extension          
```

In particular the software expects the LIRC device to be `/dev/lirc0` by default. You will
need to create a database at `/var/local/remotes` which will be used for storing the
key mapping database for each remote, which is stored in XML format with the file
extension `.keymap`.

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
remote device has been learnt, you can test it as follows:

```
bash% ir_rcv
```

This will display a table of learnt keys and also display scan codes which have yet to be matched.
You can re-invoke the learning to modify the database.

## Running Microservices

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







