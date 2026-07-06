from machine import Pin
import time
import os
print("files included in this run: ", os.listdir())

from display import Display
from input import Input

print("initialising display")
display = Display()
input = Input()

input.update()
display.update()
display.fill(0x0000) 




