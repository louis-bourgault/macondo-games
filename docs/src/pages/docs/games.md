---
layout: '../../layouts/DocsLayout.astro'
title: 'Go Games system'
draft: false
category: "go"
---

The code contains a system of interfaces for coding games, defined in ```software/internal/game/game.go```. This interface is:

```go
type Game interface {
	Update(dt float64, input platform.InputSystem, log platform.LogSystem) Game
	Draw(platform.Screen)
}
```

In addition to this, each game should implement a function that returns a new Game object, for instantiation.
# Functions
## Update function
The update function takes dt, the delta time since the last frame, a reference to the input system (which has methods such as ```input.WasKeyJustPressed```), and a logging system, which contains a single function (```log.Log```) to log information to either the browser console or to serial, depending on whether the wasm or device version is running.

At the end of the function, you must return a Game object. This normally is just a reference to the current game; the only times that this is not the case is in the Menu game where other games can be launched, and when a game is being exited, in which ```nil``` can be returned in order to exit the game and return to the menu

## Draw function
The draw function is pretty simple. It gives a reference to the Screen object (with functions such as ```Pixel``` and ```FillRect```), which can be drawn to. At the end of this function, you must call ```screen.Present()```. Also it should be noted that the buffer is not reset in between frames, so if you don't want to be drawing on top of the previous frame, you should call ```screen.Fill``` to set the entire buffer to a certain colour.

There are also some helpers functions in the helpers module for drawing images and text, which you can pass the screen object to. These functions work by just calling Pixel() a bunch of times to draw to the screen.

### Images
To include an image in the program, you'll need to process it before you can use it with ```helpers.DrawImage```. First, draw your image in whatever program you use, like krita or ms paint. Then open the converter at ```/software/web/convert/index.html``` and give it a transparent PNG. It will give you a little-endian RGB565 ```.bin``` file that you can put into your code with ```//go:embed```. Transparent pixels are stored as the chroma key colour ```0xF81F```, and will not be drawn.

For example, the flappy bird game does this:

```go
//go:embed bird.bin
var rawBirdData string

var BirdSprite = helpers.Image{
	Data: rawBirdData,
	W:    12,
	H:    12,
}
```
This lets the binary data be embedded with the program instead of constructing a large pixel array at runtime.


# Template
You can find a simple template game in ```software/internal/game/template/template.go```, which provides a minimal implementation of all these methods to build off.
