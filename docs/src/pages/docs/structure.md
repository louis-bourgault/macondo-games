---
layout: '../../layouts/DocsLayout.astro'
title: 'Project Structure'
draft: false
---

## File structure

## Project Structure
- /software - all software to run both on the machine and for development
- /mini-games - The kicad project and files for the hardware
- /mini-games/jlcpcb/production_files - The files for sending to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/bouni/kicad-jlcpcb-tools)
  - These are also duplicated in /production, for ease of review.
- /case - files for the making of the case. This includes f3d and stl files.
- /img - the images used in this readme
- /docs - this website

## Software structure
This is architected like a pretty standard Go project.
All within the /software directory:

- /cmd - entry points for wasm and the device version
- /internal - what it sounds like
  - /internal/game - all the games that I've made for the system. There's also a definition of the interfaces that must be filled in ```/internal/game/game.go``` and a template game under /template.
    - /internal/game/menu/menu.go - the menu page is just coded like a standard game! All the references to the game's constructors are in a go object at the top of this file
  - /helpers - miscellaneous things, mostly for rendering. For example ```text.go``` contains the font file, and a function to draw it to any display object just with Pixel() functions. Also, the image implementation things are here. Basically, for rendering stuff that is too high level to be bothered implementing seperately for wasm and the device. There's also a nice wrapper for a RNG system, mostly since I like the syntax of a randint function with a min and max, like in Python. Also, there's a RGB value to RGB565 converter, which is, for example, used to create the gradient for the snake game by varying values of green.
- /platform - different implementations of all the interfaces for the input system and display system. Pretty simple functions.

