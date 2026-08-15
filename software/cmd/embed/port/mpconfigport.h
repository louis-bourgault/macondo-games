/* This file is part of the MicroPython project, http://micropython.org/
 * The MIT License (MIT)
 * Copyright (c) 2022-2023 Damien P. George
 *
 * Ferretboard MicroPython configuration.
 *
 * Copied into the generated embed package by the justfile `embed` recipe, so
 * MP's C sources are built against these settings.
 *
 * Deviating from the stock embed example:
 *   - No VFS / os / io / binascii: the flash filesystem is owned by Go
 *     (flash_fs.go via tinyfs); module imports are served by mp_import_stat /
 *     mp_lexer_new_from_file in port/repl_stubs.c (MICROPY_ENABLE_EXTERNAL_IMPORT),
 *     which call back into Go. The web editor talks to the device through the
 *     `ferret` module instead of file objects.
 *   - MICROPY_BOARD_BEFORE_PYTHON_EXEC: the `ferret` module is registered at
 *     runtime (the Go engine is not scanned by MP's qstr generator); MP's soft
 *     reset (Ctrl-D) clears the module registry, so re-register before every
 *     paste/exec. register_ferret_module() is defined by host.c (compiled by
 *     cgo) and is a no-op-ish idempotent re-store.
 */

// Include common MicroPython embed configuration.
#include <port/mpconfigport_common.h>

#define MICROPY_CONFIG_ROM_LEVEL                (MICROPY_CONFIG_ROM_LEVEL_MINIMUM)

// readline.c stores history via MP_STATE_PORT; root it in the VM state like
// the unix/rp2 ports do.
#define MP_STATE_PORT                           MP_STATE_VM

#define MICROPY_ENABLE_COMPILER                 (1)
#define MICROPY_ENABLE_GC                       (1)
#define MICROPY_PY_GC                           (1)
#define MICROPY_ENABLE_FINALISER                (1)

// The `time` module (extmod/modtime.c, copied into the embed package by the
// justfile): sleep / sleep_ms / sleep_us and ticks_* for pacing game loops.
// Backed by mp_hal_delay_ms/us + mp_hal_ticks_us/cpu in mphalport.c (RP2040
// timer on device, nanosleep on host). GMTIME/LOCALTIME/MKTIME and time() stay
// off: there is no RTC/epoch source (mp_hal_time_ns is ticks-based).
#define MICROPY_PY_TIME                         (1)

// Floating point: needed for `math` (its whole module is gated on float) and
// for float literals/`float()`. Soft-float on the device: arithmetic comes from
// compiler-rt's AEABI helpers, transcendentals (sqrtf/sinf/...) from picolibc.
#define MICROPY_FLOAT_IMPL                      (MICROPY_FLOAT_IMPL_FLOAT)

// Builtin modules kept minimal but useful for games: math (rotation, easing),
// struct (binary save data), array (typed buffers) and collections (namedtuple
// / deque). All four live in py/ so no extra sources need copying; these flags
// are off at MICROPY_CONFIG_ROM_LEVEL_MINIMUM so they must be set explicitly.
#define MICROPY_PY_MATH                         (1)
#define MICROPY_PY_MATH_CONSTANTS               (1)  // pi, e
#define MICROPY_PY_MATH_FACTORIAL               (1)
#define MICROPY_PY_MATH_ISCLOSE                 (1)
#define MICROPY_PY_STRUCT                       (1)
#define MICROPY_PY_ARRAY                        (1)
#define MICROPY_PY_ARRAY_SLICE_ASSIGN           (1)
// Slice syntax (a[1:3]) is gated behind the slice type at MINIMUM; enable it
// (plus slice.start/stop/step attrs and .indices()) so array slicing works.
#define MICROPY_PY_BUILTINS_SLICE               (1)
#define MICROPY_PY_BUILTINS_SLICE_ATTRS         (1)
#define MICROPY_PY_BUILTINS_SLICE_INDICES       (1)
#define MICROPY_PY_COLLECTIONS                  (1)
#define MICROPY_PY_COLLECTIONS_DEQUE            (1)
#define MICROPY_PY_COLLECTIONS_DEQUE_ITER       (1)
#define MICROPY_PY_COLLECTIONS_DEQUE_SUBSCR     (1)
#define MICROPY_PY_COLLECTIONS_ORDEREDDICT      (1)

// REPL / editor support.
#define MICROPY_HELPER_REPL                     (1)
#define MICROPY_REPL_INFO                       (1)
#define MICROPY_KBD_EXCEPTION                   (1)
#define MICROPY_PY_SYS                          (1)
#define MICROPY_READLINE_HISTORY_SIZE           (8)

// Service USB CDC at MicroPython VM safe points so Ctrl-C can schedule a
// KeyboardInterrupt even while a program is running rather than waiting for
// the REPL to read stdin.  The hook rate-limits itself in mphalport.c.
extern void ferret_vm_hook_loop(void);
#define MICROPY_VM_HOOK_LOOP                     ferret_vm_hook_loop();

// `import foo` compiles foo.py from the Go-owned filesystem (repl_stubs.c
// provides mp_import_stat / mp_lexer_new_from_file).
#define MICROPY_ENABLE_EXTERNAL_IMPORT          (1)
#define MICROPY_READER_VFS                      (0)
#define MICROPY_READER_POSIX                    (0)

#define MICROPY_BANNER_MACHINE                  "Ferretboard"

// The `ferret` module is registered in Go (host.c); re-register it before any
// paste/exec so it survives a soft reset.
extern void register_ferret_module(void);
#define MICROPY_BOARD_BEFORE_PYTHON_EXEC(input_kind, exec_flags) register_ferret_module()
