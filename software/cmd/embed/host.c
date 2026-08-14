// host.c — C glue between the Go `ferret` engine and libmicropython.
//
// Compiled by cgo as part of the Go build (both the device and host reference),
// so it has access to the MP embed headers. Responsibilities:
//   1. Trampolines that unpack a MicroPython call into the Go //export funcs.
//   2. Runtime registration of the `ferret` module (qstr_from_str, because the
//      Go engine is not scanned by MP's qstr generator, so MP_QSTR_ferret does
//      not exist at compile time).
//   3. run_repl(): a line-based REPL over the USB CDC, invoked from Go main().
//
// main() lives in Go (main.go); this file deliberately has no main().

#include <stdint.h>
#include <string.h>

#include "py/qstr.h"
#include "py/runtime.h"
#include "py/obj.h"
#include "py/objmodule.h"
#include "py/objstr.h"
#include "py/lexer.h"
#include "py/parse.h"
#include "py/compile.h"
#include "py/mperrno.h"
#include "py/nlr.h"
#include "py/gc.h"
#include "py/stackctrl.h"
#include "shared/runtime/pyexec.h"
#include "shared/runtime/interrupt_char.h"
#include "shared/readline/readline.h"
#include "port/micropython_embed.h"
#include "ferret_abi.h"

#ifndef STATIC
#define STATIC static
#endif

// The MicroPython GC heap (allocated by Go main) as captured by ferret_boot(),
// so run_repl can re-initialise it during a soft reset (Ctrl-D).
STATIC uint8_t *repl_heap;
STATIC size_t repl_heap_size;

// Called once from Go main() to capture the GC heap for the soft-reset path.
void ferret_boot(void *gc_heap, size_t gc_heap_size) {
    repl_heap = (uint8_t *)gc_heap;
    repl_heap_size = gc_heap_size;
}

// --- trampolines: unpack a MicroPython call into the Go export ---

STATIC mp_obj_t t_fill(mp_obj_t color) {
    ferret_fill((uint16_t)mp_obj_get_int(color));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_fill_obj, t_fill);

STATIC mp_obj_t t_fill_rect(size_t n_args, const mp_obj_t *args) {
    ferret_fill_rect(mp_obj_get_int(args[0]), mp_obj_get_int(args[1]),
                     mp_obj_get_int(args[2]), mp_obj_get_int(args[3]),
                     (uint16_t)mp_obj_get_int(args[4]));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR(t_fill_rect_obj, 5, t_fill_rect);

STATIC mp_obj_t t_pixel(mp_obj_t x, mp_obj_t y, mp_obj_t color) {
    ferret_pixel(mp_obj_get_int(x), mp_obj_get_int(y), (uint16_t)mp_obj_get_int(color));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_3(t_pixel_obj, t_pixel);

STATIC mp_obj_t t_present(void) {
    ferret_present();
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_0(t_present_obj, t_present);

STATIC mp_obj_t t_draw_text(size_t n_args, const mp_obj_t *args) {
    ferret_draw_text((char *)mp_obj_str_get_str(args[0]), mp_obj_get_int(args[1]),
                     mp_obj_get_int(args[2]), (uint16_t)mp_obj_get_int(args[3]));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR(t_draw_text_obj, 4, t_draw_text);

STATIC mp_obj_t t_random_int(mp_obj_t min, mp_obj_t max) {
    return mp_obj_new_int(ferret_random_int(mp_obj_get_int(min), mp_obj_get_int(max)));
}
STATIC MP_DEFINE_CONST_FUN_OBJ_2(t_random_int_obj, t_random_int);

// --- input ---

STATIC mp_obj_t t_input_update(void) {
    ferret_input_update();
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_0(t_input_update_obj, t_input_update);

STATIC mp_obj_t t_input_is_pressed(mp_obj_t key) {
    return mp_obj_new_bool(ferret_input_is_pressed((char *)mp_obj_str_get_str(key)) != 0);
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_input_is_pressed_obj, t_input_is_pressed);

STATIC mp_obj_t t_input_was_just_pressed(mp_obj_t key) {
    return mp_obj_new_bool(ferret_input_was_just_pressed((char *)mp_obj_str_get_str(key)) != 0);
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_input_was_just_pressed_obj, t_input_was_just_pressed);

STATIC mp_obj_t t_input_was_just_released(mp_obj_t key) {
    return mp_obj_new_bool(ferret_input_was_just_released((char *)mp_obj_str_get_str(key)) != 0);
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_input_was_just_released_obj, t_input_was_just_released);

// --- helpers ---

STATIC mp_obj_t t_rgb_to_565(mp_obj_t r, mp_obj_t g, mp_obj_t b) {
    return MP_OBJ_NEW_SMALL_INT(ferret_rgb_to_565(mp_obj_get_int(r), mp_obj_get_int(g),
                                                  mp_obj_get_int(b)));
}
STATIC MP_DEFINE_CONST_FUN_OBJ_3(t_rgb_to_565_obj, t_rgb_to_565);

STATIC mp_obj_t t_draw_image(size_t n_args, const mp_obj_t *args) {
    int r = ferret_draw_image((char *)mp_obj_str_get_str(args[0]),
                              mp_obj_get_int(args[1]), mp_obj_get_int(args[2]));
    if (r != 0) {
        mp_raise_OSError(r == -1 ? MP_ENOENT : (r == -2 ? MP_EFBIG : MP_EIO));
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR(t_draw_image_obj, 3, t_draw_image);

// --- filesystem (Go-owned volume, flash_fs.go) ---

// Raw payload helpers: strings from MicroPython are NUL-terminated, so the
// content length comes from the str object itself.
STATIC int str_len(mp_obj_t o) {
    size_t n;
    mp_obj_str_get_data(o, &n);
    return (int)n;
}

STATIC int fs_result(int r) {
    if (r < 0) {
        mp_raise_OSError(r == -2 ? MP_EFBIG : MP_EIO);
    }
    return r;
}

STATIC mp_obj_t t_write_file(mp_obj_t path, mp_obj_t content) {
    fs_result(ferret_write_file((char *)mp_obj_str_get_str(path),
                                (char *)mp_obj_str_get_str(content),
                                (int)str_len(content)));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_2(t_write_file_obj, t_write_file);

STATIC mp_obj_t t_append_file(mp_obj_t path, mp_obj_t content) {
    fs_result(ferret_append_file((char *)mp_obj_str_get_str(path),
                                 (char *)mp_obj_str_get_str(content),
                                 (int)str_len(content)));
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_2(t_append_file_obj, t_append_file);

STATIC mp_obj_t t_write_image(size_t n_args, const mp_obj_t *args) {
    int r = ferret_write_image((char *)mp_obj_str_get_str(args[0]),
                               mp_obj_get_int(args[1]), mp_obj_get_int(args[2]),
                               (char *)mp_obj_str_get_str(args[3]),
                               (int)str_len(args[3]));
    if (r != 0) {
        mp_raise_OSError(MP_EINVAL);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR(t_write_image_obj, 4, t_write_image);

STATIC mp_obj_t t_append_image(mp_obj_t name, mp_obj_t b64) {
    int r = ferret_append_image((char *)mp_obj_str_get_str(name),
                                (char *)mp_obj_str_get_str(b64),
                                (int)str_len(b64));
    if (r != 0) {
        mp_raise_OSError(r == -2 ? MP_EINVAL : MP_EFBIG);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_2(t_append_image_obj, t_append_image);

STATIC mp_obj_t t_write_image_end(mp_obj_t name) {
    int r = ferret_write_image_end((char *)mp_obj_str_get_str(name));
    if (r != 0) {
        mp_raise_OSError(MP_EINVAL);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_write_image_end_obj, t_write_image_end);

STATIC mp_obj_t t_image_manifest(void) {
    int n = ferret_image_manifest(ferret_fs_buf, FERRET_FS_BUF_MAX);
    if (n < 0) {
        mp_raise_OSError(MP_ENOMEM);
    }
    return mp_obj_new_str(ferret_fs_buf, (size_t)n);
}
STATIC MP_DEFINE_CONST_FUN_OBJ_0(t_image_manifest_obj, t_image_manifest);

STATIC mp_obj_t t_delete_image(mp_obj_t name) {
    int r = ferret_delete_image((char *)mp_obj_str_get_str(name));
    if (r != 0) {
        mp_raise_OSError(MP_EIO);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_delete_image_obj, t_delete_image);

STATIC mp_obj_t t_measure_text(mp_obj_t text) {
    int w, h;
    ferret_measure_text((char *)mp_obj_str_get_str(text), &w, &h);
    mp_obj_t elems[2] = { MP_OBJ_NEW_SMALL_INT(w), MP_OBJ_NEW_SMALL_INT(h) };
    return mp_obj_new_tuple(2, elems);
}
STATIC MP_DEFINE_CONST_FUN_OBJ_1(t_measure_text_obj, t_measure_text);

// Register the `ferret` module at runtime (survives the Go/.a qstr boundary).
void register_ferret_module(void) {
    mp_obj_t mod = mp_obj_new_module(qstr_from_str("ferret"));
    mp_obj_dict_t *g = mp_obj_module_get_globals(mod);
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("fill")),
                      MP_OBJ_FROM_PTR(&t_fill_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("fill_rect")),
                      MP_OBJ_FROM_PTR(&t_fill_rect_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("pixel")),
                      MP_OBJ_FROM_PTR(&t_pixel_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("present")),
                      MP_OBJ_FROM_PTR(&t_present_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("draw_text")),
                      MP_OBJ_FROM_PTR(&t_draw_text_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("random_int")),
                      MP_OBJ_FROM_PTR(&t_random_int_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("input_update")),
                      MP_OBJ_FROM_PTR(&t_input_update_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("input_is_pressed")),
                      MP_OBJ_FROM_PTR(&t_input_is_pressed_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("input_was_just_pressed")),
                      MP_OBJ_FROM_PTR(&t_input_was_just_pressed_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("input_was_just_released")),
                      MP_OBJ_FROM_PTR(&t_input_was_just_released_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("rgb_to_565")),
                      MP_OBJ_FROM_PTR(&t_rgb_to_565_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("draw_image")),
                      MP_OBJ_FROM_PTR(&t_draw_image_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("measure_text")),
                      MP_OBJ_FROM_PTR(&t_measure_text_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("write_file")),
                      MP_OBJ_FROM_PTR(&t_write_file_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("append_file")),
                      MP_OBJ_FROM_PTR(&t_append_file_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("write_image")),
                      MP_OBJ_FROM_PTR(&t_write_image_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("append_image")),
                      MP_OBJ_FROM_PTR(&t_append_image_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("write_image_end")),
                      MP_OBJ_FROM_PTR(&t_write_image_end_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("image_manifest")),
                      MP_OBJ_FROM_PTR(&t_image_manifest_obj));
    mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str("delete_image")),
                      MP_OBJ_FROM_PTR(&t_delete_image_obj));
}

// run_boot_script compiles and runs /main.py (fetched from the Go-owned
// filesystem) right after mp_embed_init, mirroring a stock firmware running
// its boot script: a Ctrl-D soft reset re-enters run_repl, which re-reads the
// file, so the editor can "run a project" by writing files and soft-resetting.
//
// The source lives in the shared static buffer (ferret_fs_buf); the str lexer
// borrows it (free_len = 0) and only needs it until compilation finishes, so
// module imports during the script's execution can reuse the same buffer.
STATIC void run_boot_script(void) {
    int n = ferret_get_boot_script(ferret_fs_buf, FERRET_FS_BUF_MAX);
    if (n < 0) {
        mp_printf(&mp_plat_print, "boot: cannot read /main.py (err %d)\r\n", n);
        return;
    }
    if (n == 0) {
        return; // no boot script: drop straight into the REPL
    }
    // The REPL only creates its globals dict on first input (pyexec.c), so
    // give the boot script a fresh module-level namespace of its own, like a
    // stock firmware running /main.py. It must be passed to
    // mp_parse_compile_execute explicitly: NULL globals would set the thread
    // globals to NULL for the compile (which captures them into the module
    // context) and crash on the first name lookup.
    mp_obj_dict_t *g = mp_obj_new_dict(32);
    // Allow Ctrl-C to interrupt a looping game, like a stock port.
    mp_hal_set_interrupt_char(CHAR_CTRL_C);
    mp_lexer_t *lex = mp_lexer_new_from_str_len(qstr_from_str("main.py"),
                                                ferret_fs_buf, (size_t)n, 0);
    // Catch uncaught exceptions (a crashing boot script must land in the
    // REPL, and an exception must never propagate past the cgo boundary).
    nlr_buf_t nlr;
    if (nlr_push(&nlr) == 0) {
        mp_parse_compile_execute(lex, MP_PARSE_FILE_INPUT, g, g);
        nlr_pop();
    } else {
        mp_obj_print_exception(&mp_plat_print, MP_OBJ_FROM_PTR(nlr.ret_val));
    }
    mp_hal_set_interrupt_char(-1);
}

// run_repl runs MicroPython's real REPL (shared/runtime/pyexec.c), which is
// what the web editor speaks to:
//   - friendly REPL: ">>> " line editing (readline), autocomplete, Ctrl-C
//   - Ctrl-A: raw REPL mode (used by mpremote/ampy-style tools)
//   - Ctrl-E: paste mode (the editor pastes whole programs here)
//   - Ctrl-D: soft reset
//
// Input/output flow through mp_hal_stdin_rx_chr / mp_hal_stdout_tx_strn,
// which mphalport.c routes to the Go CDC exports. pyexec_friendly_repl returns
// (with PYEXEC_FORCED_EXIT) on Ctrl-D soft reset or when switching to/from raw
// mode (Ctrl-A/Ctrl-B toggles pyexec_mode_kind), so dispatch on the current
// mode and mp_reset() on a forced exit, exactly like the unix port's main().
void run_repl(void) {
    // All MP activity for the lifetime of this function runs on the cgo/device
    // thread that called us, so anchor the MP stack scan at our own frame. This
    // matters on the PC reference build: cgo runs C on an OS thread separate
    // from Go's goroutine stack (which GC scan would otherwise walk too).
    char stack_marker;
    mp_stack_set_top(&stack_marker);
    mp_embed_init(repl_heap, repl_heap_size, &stack_marker);
    register_ferret_module();
    run_boot_script();

    for (;;) {
        int ret;
        if (pyexec_mode_kind == PYEXEC_MODE_RAW_REPL) {
            ret = pyexec_raw_repl();
        } else {
            ret = pyexec_friendly_repl();
        }
        if (ret & PYEXEC_FORCED_EXIT) {
            // soft reset (Ctrl-D from the REPL): tear down and re-bootstrap MP
            // with a fresh heap, like rp2's soft_setup. register_ferret_module()
            // runs again via MICROPY_BOARD_BEFORE_PYTHON_EXEC on each exec, and
            // the boot script is re-read, so the game restarts.
            mp_deinit();
            gc_sweep_all();
            mp_embed_init(repl_heap, repl_heap_size, &stack_marker);
            register_ferret_module();
            run_boot_script();
        }
    }
}
