---
layout: '../../layouts/DocsLayout.astro'
title: 'Project Overview'
draft: false
---

# Ferretboard!!!

Ferretboard is a small games console based around a custom pcb using the RP2040 chip, and a 1.54" SPI display. The overall dimensions of the device are around 50x100x15mm, so a relatively typical size for a handheld. 

This device is being made as part of Hack Club Macondo (which is awesome). You can find it at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137).

Another major element of this project is that everything is also compatible with being built for web assembly. A demo of this can be found at the root of this website.

## Components
As previously mentioned, the system is made up of a custom pcb, including the RP2040 chip, 16mb of flash memory, and a battery management circuit. It also uses:
- 3.7v 1000mAh LiPo
- 1.54" LCD Screen of pixel dimensions 240x240 and with an SPI interface
- 6 basic 6x6x5 push to make switches that I had lying around

## Controls
The system has a basic d-pad, as well as A and B buttons as the primary controls. There are also two buttons (currently called start and select, but somewhat confusingly named since that aren't what they do) for other purposes, but it is better to not use them in games since they are smaller, more annoying to access, buttons soldered to the board.

For the WASM version, these are mapped to:
| Button | Key |
| ------ | --- |
| D-Pad  | WASD |
| A      | J |
| B | K |
| Start | Enter | 
| Select | Shift |

## Constraints
The RP2040 has its hardware constraints, although most games will fit very comfortably within them. It runs on a 135MHz dual core cpu, with 265KB of SRAM. A considerable amount of this ram is used for the display buffer (115kb, since we use full 16 bit colour), leaving 150 remaining.

Another constraint that I discovered through working on this is that the RP2040 does not have a FPU (floating point unit). This means that it takes a ton of cpu cycles to do any float division and multiplication, so we need to make sure to stay with integers for as much as possible.
This is a problem for maths visualisations like the Mandelbrot and Julia set, which probably will not run very well, if at all, on the actual device unless optimised further.

The flash storage is constrained at 16MB, but realistically, there is no world in which i hit this. A full screen background takes up 115KB, and software really isn't that much. As of 12 June, the complete uf2 file comes to about 650KB, including quite a few images, so flash is pretty much unlimited for what we're using this for.

## This website
This website is made with Astro, and no external libraries, and can be found in the ```/docs``` folder of the repository. 

## Acknowledgements
Font file used in /software/helpers/text.go: [github.com/dhepper/font8x8](https://github.com/dhepper/font8x8/blob/master/font8x8_basic.h). (Public domain)