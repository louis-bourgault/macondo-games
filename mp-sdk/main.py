from machine import Pin
import machine
import time
import os

import sys
print(sys.path)
print(os.listdir())

from display import Display

print("initialising display")
display = Display()

display.update()
display.fill(0x0000) 
display.text("Hello, World!", 10, 10, 0xFFFF)
display.update()

