---
layout: '../../layouts/DocsLayout.astro'
title: 'Micropython SDK - Embedded Reference'
draft: false
category: "mp"
---

The firmware exposes a single ```ferret``` module with all drawing, input and
helper functions built in. No imports are needed - it is always available in
the REPL and in your game script.

All colours are RGB565 integers; use ```ferret.rgb_to_565(r, g, b)``` to build
one from 8-bit channels. The screen is 240x240 pixels.

## Drawing

Drawing calls update a frame buffer in RAM. Call ```ferret.present()``` to
send it to the screen.

### ```ferret.pixel```
Arguments: x, y, colour.

Draws a single pixel.

### ```ferret.fill_rect```
Arguments: x, y (top left corner), width, height, colour.

Draws a filled rectangle.

### ```ferret.fill```
Argument: colour.

Fills the whole screen with a single colour. Useful to clear the frame
between frames.

### ```ferret.draw_text```
Arguments: text, x, y, colour.

Draws text at the top-left coordinate x, y. A single character is 8x8.

### ```ferret.measure_text```
Argument: text. Returns (width, height).

Measures the space text will take up, so you can centre it.

### ```ferret.draw_image```
Arguments: image_name, x, y.

Draws an image stored on the device at (x, y). The size is read from the image
manifest, so sync your images from the web editor first.

### ```ferret.present```
No arguments.

Flushes the frame buffer to the display.

## Input

Call ```ferret.input_update()``` once per frame **before** checking buttons.
Key names are strings: "A", "B", "UP", "DOWN", "LEFT", "RIGHT", "START",
"EXIT".

### ```ferret.input_is_pressed```
Argument: key name. Returns a bool.

True if the key is currently held.

### ```ferret.input_was_just_pressed```
Argument: key name. Returns a bool.

True if the key is pressed this frame but was not last frame.

### ```ferret.input_was_just_released```
Argument: key name. Returns a bool.

True if the key was pressed last frame but is not this frame.

## Helpers

### ```ferret.rgb_to_565```
Arguments: r, g, b. Returns an RGB565 colour.

### ```ferret.random_int```
Arguments: min, max. Returns a random integer in the inclusive range [min, max].

## Example

```python
while True:
    ferret.input_update()
    if ferret.input_was_just_pressed("A"):
        ferret.fill(ferret.rgb_to_565(255, 0, 0))
        ferret.present()
```

The editor also uses a few internal ```ferret``` functions
(```write_file```, ```write_image```, ```image_manifest```, ...) to save code
and sync images - you do not need to call these yourself.