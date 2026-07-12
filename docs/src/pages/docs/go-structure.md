---
layout: '../../layouts/DocsLayout.astro'
title: 'Golang Project Structure'
draft: false
category: "go"
---

## Software structure
This is architected like a pretty standard Go project.
All within the /software directory:

- /cmd - entry points for wasm and the device version
- /internal - what it sounds like
  - /internal/game - all the games that I've made for the system. There's also a definition of the interfaces that must be filled in ```/internal/game/game.go``` and a template game under /template.
    - /internal/game/menu/menu.go - the menu page is just coded like a standard game! All the references to the game's constructors are in a go object at the top of this file
  - /helpers - miscellaneous things, mostly for rendering. For example ```text.go``` contains the font file, and a function to draw it to any display object just with Pixel() functions. Also, the image implementation things are here. Basically, for rendering stuff that is too high level to be bothered implementing seperately for wasm and the device. There's also a nice wrapper for a RNG system, mostly since I like the syntax of a randint function with a min and max, like in Python. Also, there's a RGB value to RGB565 converter, which is, for example, used to create the gradient for the snake game by varying values of green.
- /platform - different implementations of all the interfaces for the input system and display system. Pretty simple functions.

## Software Interfaces
The system is designed to be as interchangeable as possible. There are interfaces defined in ```/software/internal/platform/platform.go```, which describe the shape of the input and display handlers for both the wasm system and the embedded system. Games are also fully abstracted: there's an interface at ```/software/internal/game/game.go``` which describes the shape of the games. 

Basically, every game defines three functions: a New() function, which is referenced in ```/software/internal/game/menu/menu.go``` which takes no arguments and returns an initialised game, an update function that takes a reference to the input system and delta time and returns a game (if you want to quit the game, you return nil and the manager system will automatically send the user back to the menu, otherwise just return the current game), and a draw function which takes a reference to the screen object being used, which you can draw to.

Each of the screen interfaces provides basic, optimised functions for filling the entire screen, drawing a rectangle, and setting an individual pixel. As well as that, there are a bunch of functions in ```/software/internal/helpers/``` to do with rendering text (i will add more later) that in turn act by just calling the Pixel() function a ton of times.

If you really want to contribute, i wouldn't mind if you made a game using the interfaces, and just chucked it in its own folder in the /internal/game/ directory. As long as you don't change anything other than that directory and the menu entry to instantiate it, i'll probably merge it. Only if you want to, though. And make sure you use the hardware efficiently.