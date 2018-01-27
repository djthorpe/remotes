# remotes
Transcoding, Sending and Receiving Infrared Remote codes

For the Raspberry Pi you should add this to your `/boot/config.txt` file
in order to load the LIRC driver:

```
dtoverlay=lirc-rpi,gpio_in_pin=23,gpio_out_pin=22
```


## Bill Of Materials

| Part                  | Description |
| ---- | ---- |
|  Vishay TSOP38238     | 38kHz IR Receiver, 950nm, 45m Range, Through Hole, 5 x 4.8 x 6.95mm |
|  Vishay TSAL6200      | 940nm IR LED, 5mm (T-1 3/4) Through Hole package |
|  Fairchild KSP2222ABU | NPN Transistor, 600 mA, 40 V, 3-Pin TO-92 |
|  680Ω ±5% 0.25W       | Carbon Resistor, 0.25W ,5%, 680R |
|  36Ω ±1% 0.6W         | MRS25 Resistor A/P,0.6W,1%,36R |
|  HV100                | TE Connectivity AMPMODU HV100 Series 2.54mm Pitch 26 Way 2 Row Straight PCB Socket, Through Hole, Solder Termination |







