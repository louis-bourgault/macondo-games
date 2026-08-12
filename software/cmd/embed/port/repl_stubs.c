// REPL-support stubs for the Ferretboard embed port.
//
// pyexec.c references mp_lexer_new_from_file for EXEC_FLAG_SOURCE_IS_FILENAME
// (running a python file by path). The REPL never does that, but the symbol
// must exist. Stock MP provides it via py/lexer.c behind MICROPY_READER_VFS/
// MICROPY_READER_POSIX; both imply a POSIX filesystem (open/read/close/errno)
// which TinyGo's bare-metal libc does not provide. So implement it directly:
// on the device there is no host filesystem path to open, and the internal VFS
// is mounted later via the `os` module, not by path.
//
// Kept out of mphalport.c so that file stays limited to stdio/ticks hooks.

#include "py/lexer.h"
#include "py/obj.h"
#include "py/runtime.h"
#include "py/mperrno.h"

#if MICROPY_ENABLE_COMPILER

mp_lexer_t *mp_lexer_new_from_file(qstr filename) {
    (void)filename;
    mp_raise_OSError(MP_ENOENT);
    return NULL; // not reached
}

#endif