// Ferretboard embed port: MicroPython stdio hooks.
//
// In the Go-hosts-MicroPython layout, Go owns the USB CDC (TinyGo's TinyUSB
// stack). MP's stdin/stdout must call back into Go, so these hooks delegate to
// the //export funcs in bridge.go / cdc_tinygo.go. On the device this routes
// the MP REPL over the Pico's USB CDC; the web editor talks to it exactly as
// it would to a stock MicroPython firmware.
//
// This file replaces the example embed port's mphalport.c (which uses printf).
// It is copied into the generated embed package by the justfile `embed` recipe
// and compiled into libmp_embed.a.

#include <stdio.h>
#include <string.h>
#include <time.h>
#include "py/mphal.h"
#include "py/runtime.h"
#include "shared/runtime/interrupt_char.h"
#include "ferret_abi.h"

// MicroPython calls this from its bytecode loop (configured through
// MICROPY_VM_HOOK_LOOP).  Poll at most once per millisecond: that keeps the
// Go/C boundary out of the hot path, while making Ctrl-C responsive even for
// a pure `while True: pass` loop that never reads from the REPL.
void ferret_vm_hook_loop(void) {
    static mp_uint_t last_poll_ms;
    mp_uint_t now = mp_hal_ticks_ms();
    if ((mp_uint_t)(now - last_poll_ms) == 0) {
        return;
    }
    last_poll_ms = now;

    if (mp_interrupt_char >= 0
        && ferret_cdc_poll_interrupt(mp_interrupt_char)) {
        mp_sched_keyboard_interrupt();
    }
}

// MP pulls one byte from stdin. Return -1 if none available (caller polls).
int mp_hal_stdin_rx_chr(void) {
    return ferret_cdc_read();
}

// MP pushes `len` bytes to stdout. Route them to the USB CDC via Go. This is
// the "raw" hook and is deliberately left unmodified: the raw REPL binary
// framing (\x01 window, \x04 EOF, paste payloads) must stay byte-exact.
mp_uint_t mp_hal_stdout_tx_strn(const char *str, size_t len) {
    ferret_cdc_write((char *)str, len);
    return len;
}

// Text stdout helpers. MP's exception/traceback printer emits bare '\n' (see
// py/obj.c mp_obj_print_exception) and some terminals run the tty with output
// post-processing off (notably `screen` on macOS), so a bare '\n' moves down a
// line without returning to column 0 and tracebacks "staircase" rightwards.
// Normalise bare '\n' to "\r\n" here on the text hooks; pyexec's own "\r\n"
// passes through unchanged because the '\r' is already present. Only the text
// hooks are translated so raw/binary REPL traffic is never corrupted.
#if defined(__arm__)
static uint8_t cdc_prev_cr = 0;
static void cdc_write_text(const char *str, size_t len) {
    size_t run_start = 0;
    for (size_t i = 0; i < len; i++) {
        if (str[i] != '\n') {
            continue;
        }
        uint8_t prev_cr = (i > 0) ? (str[i - 1] == '\r') : cdc_prev_cr;
        if (run_start < i) {
            ferret_cdc_write((char *)&str[run_start], (int)(i - run_start));
        }
        if (!prev_cr) {
            ferret_cdc_write((char *)"\r", 1);
        }
        ferret_cdc_write((char *)"\n", 1);
        run_start = i + 1;
    }
    if (run_start < len) {
        ferret_cdc_write((char *)&str[run_start], (int)(len - run_start));
    }
    cdc_prev_cr = (len > 0) ? (str[len - 1] == '\r') : cdc_prev_cr;
}
#else
// Host reference: the host terminal already does its own newline translation.
static void cdc_write_text(const char *str, size_t len) {
    ferret_cdc_write((char *)str, len);
}
#endif

// Short-string stdout helper; the REPL (pyexec.c) splices raw "OK", banners etc.
void mp_hal_stdout_tx_str(const char *str) {
    cdc_write_text(str, strlen(str));
}

// Cooked printf helper (used by MP internally); just forward to stdout hook.
void mp_hal_stdout_tx_strn_cooked(const char *str, size_t len) {
    cdc_write_text(str, len);
}

// Monotonic milliseconds, used by pyexec.c to time pastes/execs. On the device
// this reads the RP2040 hardware timer's raw low word (TIMERAWL, +0x28) in
// microseconds.  Do not use TIMELR/TIMEHR here: those are the latched read
// registers, and TIMEHR (+0x08) does not change until the low 32 bits wrap.
// The TinyGo runtime has already initialised this peripheral.
mp_uint_t mp_hal_ticks_ms(void) {
#if defined(__arm__)
    volatile uint32_t *timerawl = (uint32_t *)0x40054028;
    return *timerawl / 1000;
#else
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (mp_uint_t)(ts.tv_sec * 1000 + ts.tv_nsec / 1000000);
#endif
}

// 64-bit microsecond counter (TIMERAWH at +0x24, TIMERAWL at +0x28) on the
// RP2040. Read high before low and re-check high so a rollover between the two
// reads doesn't produce a wrapped 64-bit value.
#if defined(__arm__)
static uint64_t ferret_time_us_64(void) {
    volatile uint32_t *t = (uint32_t *)0x40054024;
    uint32_t hi = t[0];
    uint32_t lo = t[1];
    if (hi != t[0]) {
        hi = t[0];
        lo = t[1];
    }
    return ((uint64_t)hi << 32) | lo;
}
#endif

// Blocking sleeps backing the `time` module (time.sleep / sleep_ms / sleep_us).
// Device: busy-wait on the hardware timer. Host: nanosleep.
void mp_hal_delay_ms(mp_uint_t ms) {
#if defined(__arm__)
    uint64_t target = ferret_time_us_64() + (uint64_t)ms * 1000;
    while (ferret_time_us_64() < target) {
    }
#else
    struct timespec ts = { .tv_sec = ms / 1000, .tv_nsec = (long)(ms % 1000) * 1000000 };
    nanosleep(&ts, NULL);
#endif
}

void mp_hal_delay_us(mp_uint_t us) {
#if defined(__arm__)
    uint64_t target = ferret_time_us_64() + us;
    while (ferret_time_us_64() < target) {
    }
#else
    struct timespec ts = { .tv_sec = us / 1000000, .tv_nsec = (long)(us % 1000000) * 1000 };
    nanosleep(&ts, NULL);
#endif
}

// Microsecond ticks for time.ticks_us. The low 32 bits of the us counter is a
// sane ticks source (period wraps to MICROPY_PY_TIME_TICKS_PERIOD).
mp_uint_t mp_hal_ticks_us(void) {
#if defined(__arm__)
    return (mp_uint_t)(ferret_time_us_64() & 0xffffffffUL);
#else
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (mp_uint_t)(ts.tv_sec * 1000000 + ts.tv_nsec / 1000);
#endif
}

// time.ticks_cpu: no DWT setup on this board; the us counter is good enough.
mp_uint_t mp_hal_ticks_cpu(void) {
    return mp_hal_ticks_us();
}

// Nanosecond wall-clock, used by extmod/vfs_lfsx.c for file mtimes. ns
// resolution isn't needed; derive from the same clocks as mp_hal_ticks_ms.
uint64_t mp_hal_time_ns(void) {
    return (uint64_t)mp_hal_ticks_ms() * 1000000ULL;
}
