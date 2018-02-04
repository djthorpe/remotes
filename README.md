# IR Sending and Receiving

This repository contains software which provides the ability to send 
and receive infrared signals for a variety of devices usually controlled
using Infrared Remote Controllers (TV's, CD players, etc)
on a Raspberry Pi. In order to use it, you'll need some IR Sending and 
Receiving hardware, which you can build yourself. The schematics and bill of 
materials are listed below.

## Installation

This software has been tested on Raspbian Jessie. Any linux flavour which
provides the `lirc` driver should work fine. Firstly, you should add this 
to your `/boot/config.txt` file in order to load the LIRC (Linux Infrared Control) 
driver, and then reboot your Raspberry Pi:

```
dtoverlay=lirc-rpi,gpio_in_pin=22,gpio_out_pin=23
```

Your LIRC should then be able to see the device `/dev/lirc0`. The best reference
for this device is [here](https://www.kernel.org/doc/html/latest/media/uapi/rc/lirc-dev-intro.html).

## Schematic

Here is the schematic of the circuit with the bill of materials:

![IR Schematic](https://raw.githubusercontent.com/djthorpe/remotes/master/etc/ir_schematic.png)

| Part                  | Description |
| ---- | ---- |
|  Vishay TSOP38238     | 38kHz IR Receiver, 950nm, 45m Range, Through Hole, 5 x 4.8 x 6.95mm |
|  Vishay TSAL6200      | 940nm IR LED, 5mm (T-1 3/4) Through Hole package |
|  Fairchild KSP2222ABU | NPN Transistor, 600 mA, 40 V, 3-Pin TO-92 |
|  680Ω ±5% 0.25W       | Carbon Resistor, 0.25W ,5%, 680R |
|  36Ω ±1% 0.6W         | MRS25 Resistor A/P,0.6W,1%,36R |
|  HV100                | TE Connectivity AMPMODU HV100 Series 2.54mm Pitch 26 Way 2 Row Straight PCB Socket, Through Hole, Solder Termination |

If you want to make a PCB of this design you can [manufacturer one here from Aisler](https://aisler.net/djthorpe/djthorpe/raspberry-pi-ir-sender-receiver).

## Downloading The Software

You'll need a working Go Language environment to compile and install the software. Once installed,
you can use the following command to install the binaries:

```
bash% go install github.com/djthorpe/remotes
```

