// ferret_abi.h — C prototypes for the Go-exported engine entry points.
//
// Declared by hand (rather than relying on cgo's generated libbridge.h) so the
// C files can be compiled independently of the Go build:
//   - host.c lives in this package and is compiled by cgo.
//   - mphalport.c / repl_stubs.c are copied into the MicroPython embed package
//     and compiled as part of libmp_embed.a by the justfile.
// Keep the signatures in sync with bridge.go / cdc_*.go / flash_fs.go.

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

// Helpers (internal/helpers): colour conversion and text measurement.
uint16_t ferret_rgb_to_565(int r, int g, int b);
void ferret_measure_text(char *text, int *w, int *h);

// Go-owned filesystem (flash_fs.go). Return codes: 0 = ok / missing, negative
// = error, positive = payload length. Paths are "/main.py"-style.
int ferret_get_boot_script(char *buf, int max);   // 0 none, len, -1 too big, -2 io
int ferret_stat(char *path);                      // 0 missing, 1 dir, 2 file, -1 io
int ferret_read_file(char *path, char *buf, int max); // len, 0 missing, -1 io, -2 too big
int ferret_write_file(char *path, char *data, int n); // 0 ok, -1 io, -2 too big
int ferret_append_file(char *path, char *data, int n); // 0 ok, -1 io, -2 too big
int ferret_write_image(char *name, int w, int h, char *b64, int n); // 0 ok, -1 invalid
int ferret_append_image(char *name, char *b64, int n); // 0 ok, -1 invalid, -2 no upload, -3 too big
int ferret_write_image_end(char *name);           // 0 ok, -1 size mismatch, -2 no upload
int ferret_image_manifest(char *buf, int max);    // len, -1 buffer too small
int ferret_delete_image(char *name);              // 0 ok, -1 invalid/io
int ferret_draw_image(char *name, int x, int y);  // 0 ok, -1 unknown, -2 too big, -3 io

int ferret_cdc_read(void);
void ferret_cdc_write(char *s, int n);
// Drain USB CDC while MicroPython bytecode is running.  Returns non-zero when
// interrupt_char was received; ordinary input remains queued for the REPL.
int ferret_cdc_poll_interrupt(int interrupt_char);

// Shared scratch buffer for file content (boot script + module imports),
// defined in port/repl_stubs.c.
#define FERRET_FS_BUF_MAX (16 * 1024)
extern char ferret_fs_buf[FERRET_FS_BUF_MAX];

#ifdef __cplusplus
}
#endif

#endif // FERRET_ABI_H
