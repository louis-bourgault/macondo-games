---
layout: '../../layouts/DocsLayout.astro'
title: 'MicroPython SDK - Embedded Reference'
draft: false
category: "mp"
---

The firmware exposes a single ```ferret``` module with all drawing, input and
helper functions built in. Import it at the top of your game script with:

```python
import ferret
```

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
manifest, so sync your images from the web editor first. Transparent pixels use
the RGB565 chroma key colour ```0xF81F``` and are skipped. The current firmware
limits each raw RGB565 image to 64KiB, or 32,768 pixels.

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
import ferret

while True:
    ferret.input_update()
    if ferret.input_was_just_pressed("A"):
        ferret.fill(ferret.rgb_to_565(255, 0, 0))
        ferret.present()
```

## Files and editor functions

The editor also uses ```ferret.write_file```, ```ferret.append_file```,
```ferret.write_image```, ```ferret.append_image```,
```ferret.write_image_end```, ```ferret.image_manifest``` and
```ferret.delete_image``` to save code and sync images. Games do not normally
need to call these functions themselves. A single file is currently limited to
16KiB. Normal Python ```open()``` and ```os``` filesystem access is not included.
