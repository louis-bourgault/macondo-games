# Ferretboard browser emulator

The editor runs the official MicroPython WebAssembly port in a Web Worker and
registers a browser implementation of the firmware's `ferret` module. It models
the RP2350 console: a 240x240 RGB565 display, d-pad, A/B/X/Y and MENU.

## Native parity

The native firmware remains the source of truth:

- `npm run generate:emulator` reads `software/cmd/embed/host.c` and generates
  the exact browser API type. Adding or removing a registered native function
  therefore makes the browser build fail until its implementation is updated.
- The same command generates the browser bitmap font directly from
  `software/internal/helpers/text.go`.
- Browser drawing follows the native framebuffer rules: RGB565, clipping,
  `0xF81F` chroma-key pixels, 8x8 text, and little-endian image data.
- The emulator deliberately uses `ProjectData` directly; serial paste mode and
  LittleFS transfer are device transport details, not student-facing behavior.

The current npm dependency pins MicroPython WASM. When firmware upgrades its
MicroPython checkout, update that dependency to the corresponding release and
run the same student-level compatibility programs on host, device and browser.
A custom WebAssembly build using the firmware's `mpconfigport.h` is the next
step if exact module/memory configuration becomes more important than the
convenience of the official prebuilt package.

## Hosting requirement

Live input uses a `SharedArrayBuffer`, allowing button changes to reach Python
even while it is executing a tight `while True` loop. The site must send:

```
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
```

The SvelteKit server hook and Vite development/preview servers set these
headers. A static production host must be configured to send them too.

## Controls

| Console | Keyboard |
| --- | --- |
| d-pad | arrows or WASD |
| A / B | J / K |
| X / Y | U / I |
| MENU | Enter |

The on-screen buttons support mouse, pen and touch. Stopping a program
terminates and recreates its worker, so even a non-cooperative infinite loop
can always be stopped without freezing the editor.
