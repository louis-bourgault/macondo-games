// ferret_abi.h — C prototypes for the Go-exported engine entry points.
//
// Declared by hand (rather than relying on cgo's generated libbridge.h) so the
// C files can be compiled independently of the Go build:
//   - host.c lives in this package and is compiled by cgo.
//   - mphalport.c is copied into the MicroPython embed package and compiled as
//     part of libmp_embed.a by the justfile.
// Keep the signatures in sync with bridge.go / cdc_*.go.

#ifndef FERRET_ABI_H
#define FERRET_ABI_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

void ferret_fill(uint16_t color);
void ferret_fill_rect(int x, int y, int w, int h, uint16_t color);
void ferret_pixel(int x, int y, uint16_t color);
void ferret_present(void);
void ferret_draw_text(char *text, int x, int y, uint16_t color);
int ferret_random_int(int min, int max);
int ferret_cdc_read(void);
void ferret_cdc_write(char *s, int n);

#ifdef __cplusplus
}
#endif

#endif // FERRET_ABI_H
