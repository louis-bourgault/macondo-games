// Filesystem-facing stubs for the Ferretboard embed port.
//
// With MICROPY_VFS disabled, MicroPython's import machinery (builtinimport.c,
// behind MICROPY_ENABLE_EXTERNAL_IMPORT) delegates everything to two
// port-provided functions:
//
//   - mp_import_stat: does "foo" / "foo.py" exist? The import code iterates
//     sys.path ([""] by default), so `import foo` ends up here as "foo.py".
//   - mp_lexer_new_from_file: load a source file for compilation. This is
//     also what pyexec.c uses for EXEC_FLAG_SOURCE_IS_FILENAME.
//
// Both run on the cgo/device thread inside run_repl, so they can safely call
// the Go //export funcs, which own the LittleFS2 volume (flash_fs.go). The
// shared scratch buffer (ferret_fs_buf, also used by host.c for the boot
// script) holds one file at a time; the boot script's bytes are only needed
// while it is being compiled, so imports during its execution can reuse it.

#include "py/lexer.h"
#include "py/obj.h"
#include "py/runtime.h"
#include "py/mperrno.h"
#include "py/builtin.h"
#include "ferret_abi.h"

char ferret_fs_buf[FERRET_FS_BUF_MAX];

#if MICROPY_ENABLE_EXTERNAL_IMPORT

mp_import_stat_t mp_import_stat(const char *path) {
    int r = ferret_stat((char *)path);
    if (r < 0) {
        return MP_IMPORT_STAT_NO_EXIST;
    }
    return (mp_import_stat_t)r;
}

#endif

#if MICROPY_ENABLE_COMPILER

mp_lexer_t *mp_lexer_new_from_file(qstr filename) {
    const char *path = qstr_str(filename);
    int n = ferret_read_file((char *)path, ferret_fs_buf, FERRET_FS_BUF_MAX);
    if (n < 0) {
        if (n == -2) {
            mp_raise_OSError(MP_EFBIG);
        }
        mp_raise_OSError(MP_EIO);
    }
    if (n == 0) {
        mp_raise_OSError(MP_ENOENT);
    }
    // free_len = 0: the lexer borrows the static buffer (never frees it); the
    // source is only needed until compilation finishes, after which the
    // buffer can be reused by the next file load.
    return mp_lexer_new_from_str_len(qstr_from_str(path), ferret_fs_buf, (size_t)n, 0);
}

#endif
