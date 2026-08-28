---
layout: '../../layouts/DocsLayout.astro'
title: 'Micropython SDK (legacy) - Quick Reference'
draft: false
category: "ogmp"
---

The micropython sdk contains two main modules - ```modules.display``` and ```modules.input```. Import them at the top of your file with:

```python
import modules.display as display
import modules.input as input
```

## Display

### General shape
To make changes to the display, call the following functions to update the frame buffer in RAM. Then call ```display.update()``` to send the data to the screen.

All colours are in RGB565 format.

### ```display.pixel```
Arguments: x, y, colour.

Draws a single pixel to the screen

### ```display.rect```
Arguments: x, y (top left corner), width, height, colour.

Draws a rectangle to the screen

### ```display.text```
Arguments: text, x, y, colour

Draws text to the screen at top left coords of x and y. A single character is size 8x8.

### ```display.draw_image```
Arguments: image_name, x, y, image_width, image_height

Draws an image stored in ```/img``` in the device's flash to the screen. As of now, the logic cannot infer the image width and height, so you must explicitly pass them. This my change in future.
When you upload an image from the web editor, it is automatically placed in this directory.

### ```display.update```
No arguments.

Flushes data to the display.


## Input
The input module deals with button presses. 

### General shape
Call ```input.update``` to check all the buttons, and update state (this includes moving the store of current button presses to last button presses). Call this **before** you do any input handling in your main game loop

Key names are strings, and are "A", "B", "UP", "DOWN", "LEFT" and "RIGHT"

### ```input.is_pressed```
Argument: key name

Returns a boolean on whether it is currently pressed

### ```input.was_just_pressed```
Argument: key name

Returns true if the key is pressed this frame but was not last frame.

### ```input.was_just_released```
Argument: key name

Returns true if the key was pressed last frame but was not this frame.