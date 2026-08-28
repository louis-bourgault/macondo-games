---
layout: '../../layouts/DocsLayout.astro'
title: 'Assembly (v1)'
draft: false
category: "general"
---

## Assembly instructions
These instructions are for version one of the hardware. Version two has a different pcb and a two-part case, and is still in development.

To assemble the device, start by 3d printing the shell, which can be found at ```/case/v1/case.stl``` in this repository. I would reccomend doing this with some kind of high precision setting on your printer - my first print used the Draft quality with layer height of 0.2mm, but my second used 0.08mm and it looked a lot better. I just printed this on my school's Bambu Lab P1S printer. Here's how it looks, printed in white on this 0.08mm setting.
![The case, printed in white](https://cdn.hackclub.com/019f53dc-d891-75ac-8839-15e7a4a7df70/whitecase.jpg.webp)

Next, you'll need to solder the buttons to the PCB that you've got printed and assembled. Get 6 generic 6x6x5mm push-to-make switches, and solder them into the holes on the pcb. Get a 3.7v LiPo battery (the finished build uses 1000mAh), and solder the positive and negative lines to the copper pads of the board as indicated below. Be very careful about polarity and do not work on, cut, or solder battery wires in a way that can short them together:
![Where to solder things](https://cdn.hackclub.com/019f53dc-c312-7b13-8c0a-8aa6cdc4ade4/instructions.jpg.webp)

Key: Battery as indicated, positive to J4, negative to J3. Make sure these lines do not touch.
Switches: solder legs to the through-hole pads for SW2-SW7, as highlighted in yellow.

Next, get your screen and solder it to the screen connector at the top of the PCB. The leftmost pin, from the front of the board, is the BL pin, with GND on the far right. This means that I had to mount the specific display I bought from aliexpress upside down, and flip all the coordinates in software.
![Connecting the screen](https://cdn.hackclub.com/019f53dc-cdfa-7ea5-9303-2d4c1be6da54/screensolder.jpg.webp)

Once you've done that, the device is pretty much ready to go. Slot the whole contraption into the 3d printed case - the case is designed so that the battery goes behind the screen and then the wires for the battery come down and snake through the little notch to the left of the battery compartment. There are four posts on the lower section of the case, which you can use to hold the PCB in place through the circular holes on each corner of the PCB.

If you leave it like that, it will work, but depending on the sizes of all your components, everything will probably rattle and tilt and slide around a ton. I solved this problem by liberally applying small bits of furniture skates, double sided tape, and foam to hold everything in place. This is something that I can probably improve on in version two of the system.

From there, all you'll need to do is to load the code onto the thing. For the current MicroPython SDK and web editor, build the firmware using the [current MicroPython building guide](/docs/micropython-building). Hold the board's flash/BOOTSEL button while plugging it into your computer over USB C, then copy the compiled ```ferret.uf2``` file onto the drive that appears. You can also flash it from a terminal with ```picotool load ferret.uf2 -f```. The board will reboot when it has finished.
