import machine
import micropython
import modules.framebuf2 as framebuf
import time
from machine import Pin


@micropython.viper
def _swap_range(buf: ptr8, start: int, n: int):
    i = start
    end = start + n
    while i < end:
        t = buf[i]
        buf[i] = buf[i + 1]
        buf[i + 1] = t
        i += 2

@micropython.viper
def _copy_clipped_reversed(fb: ptr8, fb_offset: int, src: ptr8, src_off: int, n: int):
    dst = fb_offset + (n - 1) * 2
    i = src_off
    end = src_off + n * 2
    while i < end:
        lo = src[i]
        hi = src[i + 1]
        if lo != 31 or hi != 248:
            fb[dst] = lo
            fb[dst + 1] = hi
        i += 2
        dst -= 2

class _Display:
    def __init__(self):
        self.sck = Pin(18, Pin.OUT)
        self.sd0 = Pin(19, Pin.OUT)
        self.dc = Pin(21, Pin.OUT)
        self.rst = Pin(20, Pin.OUT)
        self.cs = Pin(17, Pin.OUT)
        self.baud = 62500000
        self.spi = machine.SPI(0, baudrate=self.baud, polarity=1, phase=1, sck=self.sck, mosi=self.sd0)

        self.buffer = bytearray(240 * 240 * 2)
        self.fb = framebuf.FrameBuffer(self.buffer, 240, 240, framebuf.RGB565)
        self._mv = memoryview(self.buffer)
        self._cmd_buf = bytearray(1)
        self._col_addr = bytearray(4)
        self._row_addr = bytearray(4)
        self._row_buf = bytearray(480)

        self._dirty = False
        self._dx0 = 0
        self._dy0 = 0
        self._dx1 = 0
        self._dy1 = 0

        self.init_st7789()

    def write_cmd(self, cmd):
        b = self._cmd_buf
        b[0] = cmd
        self.dc.low()
        self.cs.low()
        self.spi.write(b)
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
        self.write_data(b'\x55')

        self.write_cmd(0x36)
        self.write_data(b'\x14')

        self.write_cmd(0x21)

        self.write_cmd(0x29)

    def _mark(self, x, y, w, h):
        if w <= 0 or h <= 0:
            return
        x0 = x
        y0 = y
        x1 = x + w - 1
        y1 = y + h - 1
        if x0 < 0: x0 = 0
        if y0 < 0: y0 = 0
        if x1 > 239: x1 = 239
        if y1 > 239: y1 = 239
        if not self._dirty:
            self._dx0 = x0
            self._dy0 = y0
            self._dx1 = x1
            self._dy1 = y1
            self._dirty = True
        else:
            if x0 < self._dx0: self._dx0 = x0
            if y0 < self._dy0: self._dy0 = y0
            if x1 > self._dx1: self._dx1 = x1
            if y1 > self._dy1: self._dy1 = y1

    def update(self):
        if not self._dirty:
            return
        x0 = self._dx0
        y0 = self._dy0
        x1 = self._dx1
        y1 = self._dy1
        self._dirty = False

        col = self._col_addr
        col[0] = 0
        col[1] = x0
        col[2] = 0
        col[3] = x1
        self.write_cmd(0x2A)
        self.write_data(col)

        row = self._row_addr
        row[0] = 0
        row[1] = y0
        row[2] = 0
        row[3] = y1
        self.write_cmd(0x2B)
        self.write_data(row)

        self.write_cmd(0x2C)

        buf = self.buffer
        if x0 == 0 and y0 == 0 and x1 == 239 and y1 == 239:
            _swap_range(buf, 0, len(buf))
            self.dc.high()
            self.cs.low()
            self.spi.write(buf)
            self.cs.high()
            _swap_range(buf, 0, len(buf))
        else:
            mv = self._mv
            wb = (x1 - x0 + 1) * 2
            self.dc.high()
            self.cs.low()
            for ry in range(y0, y1 + 1):
                s = ry * 480 + x0 * 2
                _swap_range(buf, s, wb)
                self.spi.write(mv[s:s + wb])
                _swap_range(buf, s, wb)
            self.cs.high()

    def fill(self, color):
        self.fb.fill(color)
        self._mark(0, 0, 240, 240)

    def pixel(self, x, y, color):
        fx = 239 - x
        fy = 239 - y
        self.fb.pixel(fx, fy, color)
        self._mark(fx, fy, 1, 1)
        self.update()

    def rect(self, x, y, w, h, color):
        flipped_x = 239 - x - (w - 1)
        flipped_y = 239 - y - (h - 1)
        self.fb.rect(flipped_x, flipped_y, w, h, color)
        self._mark(flipped_x, flipped_y, w, h)

    def text(self, text_str, x, y, color):
        str_len = len(str(text_str))
        flipped_x = 239 - x - (str_len * 8 - 1)
        flipped_y = 239 - y - 7

        self.fb.large_text(str(text_str), flipped_x, flipped_y, 1, color, r=180)
        self._mark(flipped_x, flipped_y, str_len * 8, 8)

    def draw_image(self, name, x, y, imageWidth, imageHeight):
        w = imageWidth
        h = imageHeight
        if w <= 0 or h <= 0 or w > 240:
            raise ValueError('draw_image: invalid dimensions')

        vis_x0 = x if x > 0 else 0
        vis_y0 = y if y > 0 else 0
        vis_x1 = x + w
        if vis_x1 > 240: vis_x1 = 240
        vis_y1 = y + h
        if vis_y1 > 240: vis_y1 = 240
        vis_w = vis_x1 - vis_x0
        vis_h = vis_y1 - vis_y0
        if vis_w <= 0 or vis_h <= 0:
            return

        sx0 = vis_x0 - x
        sy0 = vis_y0 - y
        src_off = sx0 * 2
        fb = self.buffer
        row_buf = self._row_buf
        read_mv = memoryview(row_buf)[:w * 2]
        col_start = 240 - vis_x1

        with open('/img/' + name, 'rb') as f:
            if sy0:
                f.seek(sy0 * w * 2)
            for viy in range(vis_h):
                f.readinto(read_mv)
                fb_row = 239 - vis_y0 - viy
                fb_offset = fb_row * 480 + col_start * 2
                _copy_clipped_reversed(fb, fb_offset, row_buf, src_off, vis_w)

        self._mark(col_start, 240 - vis_y1, vis_w, vis_h)


_display_manager = _Display()

text = _display_manager.text
fill = _display_manager.fill
pixel = _display_manager.pixel
rect = _display_manager.rect
update = _display_manager.update
draw_image = _display_manager.draw_image
