# remotes
Transcoding, Sending and Receiving Infrared Remote codes

For the Raspberry Pi you should add this to your `/boot/config.txt` file
in order to load the LIRC driver:

```
dtoverlay=lirc-rpi,gpio_in_pin=23,gpio_out_pin=22
```


