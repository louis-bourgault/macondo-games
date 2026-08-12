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
#include "py/gc.h"
#include "py/stackctrl.h"
#include "shared/runtime/pyexec.h"
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
            // runs again via MICROPY_BOARD_BEFORE_PYTHON_EXEC on each exec.
            mp_deinit();
            gc_sweep_all();
            mp_embed_init(repl_heap, repl_heap_size, &stack_marker);
        }
    }
}
