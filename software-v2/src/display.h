#ifndef DISPLAY_H
#define DISPLAY_H

#include <stdint.h>

void display_init(void);
void display_fill(uint16_t color);
void draw_pixel(int x, int y, uint16_t color);
void draw_rect(int x0, int y0, int x1, int y1, uint16_t color);
void display_present(void);

#endif