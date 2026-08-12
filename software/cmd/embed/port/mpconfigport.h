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
 *   - MICROPY_PY_SYS / MICROPY_PY_OS: the web editor talks to the device over
 *     the raw REPL and touches sys/os, and the on-device filesystem needs them.
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

// REPL / editor support.
#define MICROPY_HELPER_REPL                     (1)
#define MICROPY_REPL_INFO                       (1)
#define MICROPY_KBD_EXCEPTION                   (1)
#define MICROPY_PY_SYS                          (1)
#define MICROPY_PY_OS                           (1)
#define MICROPY_VFS                             (1)
#define MICROPY_READER_VFS                      (0)
#define MICROPY_READER_POSIX                    (0)
#define MICROPY_READLINE_HISTORY_SIZE           (8)

#define MICROPY_BANNER_MACHINE                  "Ferretboard"

// The `ferret` module is registered in Go (host.c); re-register it before any
// paste/exec so it survives a soft reset.
extern void register_ferret_module(void);
#define MICROPY_BOARD_BEFORE_PYTHON_EXEC(input_kind, exec_flags) register_ferret_module()