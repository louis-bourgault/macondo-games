from machine import Pin
import time
import os
print("files included in this run: ", os.listdir())

from modules.display import Display
from modules.input import Input

print("initialising display")
display = Display()
input = Input()

input.update()
display.update()
display.fill(0x0000) 




