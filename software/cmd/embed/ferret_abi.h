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

// Input (platform/device/input.go). Keys are SDK strings ("A", "B", "UP",
// "DOWN", "LEFT", "RIGHT", "START", "EXIT"); unknown names read as None.
void ferret_input_update(void);
int ferret_input_is_pressed(char *key);
int ferret_input_was_just_pressed(char *key);
int ferret_input_was_just_released(char *key);

// Helpers (internal/helpers): colour conversion, image blitting (raw RGB565
// bytes, chroma-keyed like the Go games), and text measurement.
uint16_t ferret_rgb_to_565(int r, int g, int b);
void ferret_draw_image(int x, int y, int w, int h, uint8_t *data, int n);
void ferret_measure_text(char *text, int *w, int *h);

int ferret_cdc_read(void);
void ferret_cdc_write(char *s, int n);

// Internal flash block device (flash_common.go): 4 KiB blocks over the
// reserved filesystem region. Used by port/flash_storage.c for the on-device
// LittleFS2 volume at "/".
int ferret_flash_read(uint32_t block, uint32_t off, uint8_t *buf, uint32_t n);
int ferret_flash_write(uint32_t block, uint32_t off, uint8_t *buf, uint32_t n);
int ferret_flash_erase(uint32_t block);
int ferret_flash_block_count(void);
int ferret_flash_block_size(void);

#ifdef __cplusplus
}
#endif

#endif // FERRET_ABI_H
