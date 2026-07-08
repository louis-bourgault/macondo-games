import machine
import modules.framebuf2 as framebuf
import time
from machine import Pin

class Display:
    def __init__(self):
        self.sck = Pin(18, Pin.OUT)
        self.sd0 = Pin(19, Pin.OUT)
        self.dc = Pin(21, Pin.OUT)
        self.rst = Pin(20, Pin.OUT)
        self.cs = Pin(17, Pin.OUT)
        self.baud = 20000000
        self.spi = machine.SPI(0, baudrate=self.baud, polarity=1, phase=1, sck=self.sck, mosi=self.sd0)

        self.buffer = bytearray(240 * 240 * 2)
        self.fb = framebuf.FrameBuffer(self.buffer, 240, 240, framebuf.RGB565)

        self.init_st7789()

    def write_cmd(self, cmd):
        self.dc.low()
        self.cs.low()
        self.spi.write(bytearray([cmd]))
        self.cs.high()

    def write_data(self, data):
        self.dc.high()
        self.cs.low()
        self.spi.write(data)
        self.cs.high()

    def init_st7789(self):
        self.rst.low()
        time.sleep_ms(50)
        self.rst.high()
        time.sleep_ms(50)
        
        self.write_cmd(0x11) 
        time.sleep_ms(120)
        
        self.write_cmd(0x3A) 
        self.write_data(b'\x55') #rgb565

        self.write_cmd(0x3A) # Interface Pixel Format
        self.write_data(b'\x55') 

        self.write_cmd(0x21) #inv on

        self.write_cmd(0x29) 

    def update(self):
        self.write_cmd(0x36)
        self.write_data(b'\x14')

        self.write_cmd(0x2A) #set col address
        self.write_data(b'\x00\x00\x00\xEF')
        self.write_cmd(0x2B) #set row address
        self.write_data(b'\x00\x00\x00\xEF')
        
        self.write_cmd(0x2C) #dump pixels
        
        buf = bytearray(len(self.buffer))
        for i in range(0, len(self.buffer), 2):
            buf[i] = self.buffer[i + 1]
            buf[i + 1] = self.buffer[i]

        self.dc.high()
        self.cs.low()
        self.spi.write(buf)
        self.cs.high()

    def fill(self, color):
        self.fb.fill(color)

    def pixel(self, x, y, color):
        self.fb.pixel(239-x, 239-y, color)
        self.update()

    def rect(self, x, y, w, h, color):
        flipped_x = 239 - x - (w - 1)
        flipped_y = 239 - y - (h - 1)
        self.fb.rect(flipped_x, flipped_y, w, h, color)
    
    def text(self, text_str, x, y, color):
        str_len = len(str(text_str))
        flipped_x = 239 - x - (str_len * 8 - 1)
        flipped_y = 239 - y - 7 
        
        self.fb.large_text(str(text_str), flipped_x, flipped_y, 1, color, r=180)
