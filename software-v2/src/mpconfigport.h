/* This file is part of the MicroPython project, http://micropython.org/
 * The MIT License (MIT)
 * Copyright (c) 2022-2023 Damien P. George
 */

// Include common MicroPython embed configuration.
#include <port/mpconfigport_common.h>

// Use the minimal starting configuration (disables all optional features).
//we could up this to BASIC_FEATURES and then we get math and randomness which might be useful
#define MICROPY_CONFIG_ROM_LEVEL                (MICROPY_CONFIG_ROM_LEVEL_MINIMUM)

// MicroPython configuration.
#define MICROPY_ENABLE_COMPILER                 (1)
#define MICROPY_ENABLE_GC                       (1)
#define MICROPY_PY_GC                           (1)
#define MICROPY_PY_SYS                          (0)
//we doing filesystem ourselves
#define MICROPY_VFS                 (0)
#define MICROPY_READER_VFS          (0)
#define MICROPY_READER_POSIX        (0)

//our change
#define MICROPY_GCREGS_SETJMP (1)
    
