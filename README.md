# IR Sending and Receiving

This repository contains software which provides the ability to send 
and receive infrared signals for a variety of devices usually controlled
using Infrared Remote Controllers (TV's, CD players, etc)
on a Raspberry Pi. In order to use it, you'll need some IR Sending and 
Receiving hardware, which you can build yourself. The schematics and bill of 
materials are listed below.

__Please note this project is in development__

## Installation

This software has been tested on Raspbian Jessie. Any linux which is compiled
with the `lirc` module should work fine. Firstly, you should add this 
to your `/boot/config.txt` file in order to load the LIRC (Linux Infrared Control) 
driver, and then reboot your Raspberry Pi:

```
dtoverlay=lirc-rpi,gpio_in_pin=22,gpio_out_pin=23
```

Your LIRC should then be able to see the device `/dev/lirc0`. The best reference
for how to interact with the device is [here](https://www.kernel.org/doc/html/latest/media/uapi/rc/lirc-dev-intro.html).

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

## Downloading The Software

You'll need a working Go Language environment to compile and install the software. Once installed,
you can use the following command to install the binaries:

```
bash% go get github.com/djthorpe/remotes/./...
```

Currently this installs two binaries, `ir_learn` and `ir_recv`.


