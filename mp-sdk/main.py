from machine import Pin
import time
import os
print("files included in this run: ", os.listdir())

from modules.display import Display
from modules.input import Input

print("initialising display")
display = Display()
input = Input()

while True:
    input.update()
    if input.is_pressed("UP"):
        display.fill(0x0000)
    else:
        display.fill(0xFFFF)
    time.sleep(0.1)




