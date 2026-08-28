---
layout: '../../layouts/DocsLayout.astro'
title: 'Project Overview'
draft: false
category: "general"
---

# Ferretboard!!!

Ferretboard is a small games console based around a custom pcb and a 1.54" SPI display. Version one uses the RP2040 chip. Version two is currently in development, and upgrades this to the RP2350 with a significantly different board and case. The overall dimensions of both versions are around 50x100x20mm, so a relatively typical size for a handheld.

This device is being made as part of Hack Club Macondo (which is awesome). You can find it at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137).

Another major element of this project is that the Go game system is also compatible with being built for web assembly. A demo of this can be found at the root of this website.

## Components
The finished version one system is made up of a custom pcb, including the RP2040 chip, 16MB of flash memory, and a battery management circuit. It also uses:
- 3.7v 1000mAh LiPo
- 1.54" LCD Screen of pixel dimensions 240x240 and with an SPI interface
- 6 basic 6x6x5 push to make switches for the d-pad, A and B
- 2 smaller buttons on the assembled pcb for Start and Exit
- 3d printed case, the stl file for which is in ```/case/v1/case.stl``` in the repository

The version two KiCad project can be found in ```/hardware-v2```, and its case files are in ```/case/v2```. It is still a work in progress, so the assembly instructions are for version one unless they specifically say otherwise.

## Controls
The system has a basic d-pad, as well as A and B buttons as the primary controls. There are also two buttons (called Start and Exit in software) for other purposes, but it is better to not use them as primary game controls since they are smaller and more annoying to access.

For the WASM version, these are mapped to:
| Button | Key |
| ------ | --- |
| D-Pad  | WASD |
| A      | J |
| B | K |
| Start | Enter |
| Exit | Shift |

## Constraints
The RP2040 has its hardware constraints, although most games will fit very comfortably within them. It is a dual core cpu rated for 133MHz, with 264KB of SRAM. A considerable amount of this ram is used for the display buffer: 115,200 bytes, or 112.5KiB, since we use full 16 bit colour. This leaves around 149KiB before the rest of the firmware and runtime allocations are taken into account.

Another constraint that I discovered through working on this is that the RP2040 does not have a FPU (floating point unit). This means that it takes a ton of cpu cycles to do any float division and multiplication, so we need to make sure to stay with integers for as much as possible.
This is a problem for maths visualisations like the Mandelbrot and Julia set, which probably will not run very well, if at all, on the actual device unless optimised further.

The flash storage is 16MB. The current embedded MicroPython firmware reserves the last 8MB for its LittleFS filesystem, with the first 8MB left for the firmware. There is plenty of overall flash for this kind of project, but individual MicroPython source files and images have smaller limits that are explained in the MicroPython reference.

## This website
This website is made with Astro, and no external libraries, and can be found in the ```/docs``` folder of the repository.

## Acknowledgements
Font file used in /software/internal/helpers/text.go: [github.com/dhepper/font8x8](https://github.com/dhepper/font8x8/blob/master/font8x8_basic.h). (Public domain)
