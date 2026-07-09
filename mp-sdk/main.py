from machine import Pin
import time
import os
print("files included in this run: ", os.listdir())

import modules.display as display
import modules.input as input

print("initialising display")

while True:
    input.update()
    if input.is_pressed("UP"):
        display.fill(0x0000)
    else:
        display.fill(0xFFFF)
    display.update()
    time.sleep(0.1)




