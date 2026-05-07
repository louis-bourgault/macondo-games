# Macondo Games
I would like to make a simple games console, based around an RP2040 chip on a custom PCB, and including a 240x240 LCD screen. I am designing this roughly to fit in the footprint of an iPod nano shape and size - hopefully acheiving final dimensions of about 550x1000x100mm.

## Components
The project will be made up of a custom pcb, including the chip, 16MB flash, battery management circuits etc. I will also use:
- 3.7v 1000mAh LiPo battery
- 1.54" LCD Screen (Already purchased) of pixel dimensions 240x240 and through an SPI interface

The exterior will be 3d printed

## Controls
D-pad controls (Up, down, left, right), as well as A and B buttons and 2 buttons for other functions, such as going to menus, pausing games, etc

## Games
I initially considered making this a gameboy emulator, but decided that that was against the spirit of the project, since I would like to code most of the things on there and don't particularly feel like playing gameboy games without sound. Initially, I will code games like Pong and Snake in Go, and then who knows what I will go to past then.
The one exception to trying to code all my own games is that, as a rite of passage into hardware engineering, I want to run Doom. 