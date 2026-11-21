#include "mp.h"
#include "port/micropython_embed.h"
#include "display.h"
#include "py/runtime.h"
#include "py/mprint.h"
#include "py/nlr.h"

static char heap[8 * 1024];

static const char user_script[] =
    "def init():\n"
    "    fill_display(0xF800)\n";

static mp_obj_t fill_display(mp_obj_t colour)
{
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    display_fill(c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_1(fill_display_obj, fill_display);

static void setup_api(void)
{
    mp_store_global(MP_QSTR_fill_display, MP_OBJ_FROM_PTR(&fill_display_obj));
}

void call_py(qstr name)
{
    nlr_buf_t nlr;
    if (nlr_push(&nlr) == 0)
    {
        mp_obj_t fn = mp_load_global(name);
        mp_call_function_0(fn); // 0 args
        nlr_pop();
    }
    else
    {
        mp_obj_print_exception(&mp_plat_print, MP_OBJ_FROM_PTR(nlr.ret_val));
    }
}

void mp_init(void *stack_top)
{
    mp_embed_init(&heap[0], sizeof(heap), &stack_top);
    setup_api();
    mp_embed_exec_string(micropython_string);
    call_py(MP_QSTR_init);
}

void script_update(void) {
    call_py(MP_QSTR_update);
}

void mp_deinit()
{
    mp_embed_deinit();
}
