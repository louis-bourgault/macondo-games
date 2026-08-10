#include <stdint.h>
#include <string.h>

#include "py/qstr.h"
#include "py/runtime.h"
#include "py/obj.h"
#include "port/micropython_embed.h"

#include "bridge/libbridge.h"

#ifndef STATIC
#define STATIC static
#endif

static char heap[8 * 1024];

STATIC mp_obj_t mp_hello(mp_obj_t name_obj) {
    const char *name_cstr = mp_obj_str_get_str(name_obj);
    GoString name = {
        .p = name_cstr,
        .n = (ptrdiff_t)strlen(name_cstr),
    };
    say_hello(name);
    return mp_const_none;
}

STATIC MP_DEFINE_CONST_FUN_OBJ_1(mp_hello_obj, mp_hello);

STATIC void register_globals(void) {
    mp_obj_dict_store(
        MP_OBJ_FROM_PTR(&MP_STATE_VM(dict_main)),
        MP_OBJ_NEW_QSTR(qstr_from_str("say_hello")),
        MP_OBJ_FROM_PTR(&mp_hello_obj)
    );
}

int main(void) {
    int stack_top;

    mp_embed_init(&heap[0], sizeof(heap), &stack_top);
    register_globals();

    mp_embed_exec_str("say_hello('world')\n");

    mp_embed_deinit();
    return 0;
}