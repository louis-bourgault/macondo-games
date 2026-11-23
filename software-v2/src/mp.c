#include "mp.h"
#include "port/micropython_embed.h"
#include "display.h"
#include "input.h"
#include "py/runtime.h"
// #include "py/mprint.h"
#include "py/nlr.h"
#include <stdio.h>
#include "py/qstr.h"

static char heap[32 * 1024];

static const char user_script[] =
"def init():\n"
"    fill_display(0xF800)\n"
"    \n"
"def update():\n"
"    print('micropython is alive weewoo')\n";



//FUNCTIONS FOR MICROPYTHON TO CALL
static mp_obj_t py_fill_display(mp_obj_t colour)
{
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    display_fill(c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_1(fill_display_obj, fill_display);
static mp_obj_t py_draw_pixel(mp_obj_t x, mp_obj_t y, mp_obj_t colour)
{
    int16_t x_val = (int16_t)mp_obj_get_int(x);
    int16_t y_val = (int16_t)mp_obj_get_int(y);
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    draw_pixel(x_val, y_val, c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_3(draw_pixel_obj, draw_pixel);
static mp_obj_t py_draw_rect(mp_obj_t x0, mp_obj_t y0, mp_obj_t x1, mp_obj_t y1, mp_obj_t colour)
{
    int16_t x0_val = (int16_t)mp_obj_get_int(x0);
    int16_t y0_val = (int16_t)mp_obj_get_int(y0);
    int16_t x1_val = (int16_t)mp_obj_get_int(x1);
    int16_t y1_val = (int16_t)mp_obj_get_int(y1);
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    draw_rect(x0_val, y0_val, x1_val, y1_val, c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_5(draw_rect_obj, draw_rect);

static mp_obj_t py_button_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_pressed_obj, button_pressed);
static mp_obj_t py_button_just_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_just_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_just_pressed_obj, button_just_pressed);
static mp_obj_t py_button_just_released(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_just_released((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_just_released_obj, button_just_released);


typedef struct { const char *name; const mp_obj_fun_builtin_var_t *fn; } api_fn_t;
typedef struct { const char *name; int value; } api_const_t;

static const api_fn_t api_fns[] = {
    { "fill_display",         &py_fill_display_obj },
    { "draw_pixel",           (const void *)&py_draw_pixel_obj },
    { "draw_rect",            (const void *)&py_draw_rect_obj },
    { "button_pressed",       (const void *)&py_button_pressed_obj },
    { "button_just_pressed",  (const void *)&py_button_just_pressed_obj },
    { "button_just_released", (const void *)&py_button_just_released_obj },
};

static const api_const_t api_consts[] = {
    { "BTN_A", 0 }, { "BTN_B", 1 }, { "BTN_X", 2 }, { "BTN_Y", 3 },
    { "BTN_UP", 4 }, { "BTN_DOWN", 5 }, { "BTN_LEFT", 6 }, { "BTN_RIGHT", 7 },
    { "BTN_MENU", 8 },
};



static void setup_api(void) {
    mp_obj_t mod = mp_obj_new_module(qstr_from_str("ferret"));
    mp_obj_t g = MP_OBJ_FROM_PTR(mp_obj_module_get_globals(mod));

    for (size_t i = 0; i < MP_ARRAY_SIZE(api_fns); i++) {
        mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str(api_fns[i].name)),
                          MP_OBJ_FROM_PTR(api_fns[i].fn));
    }
    for (size_t i = 0; i < MP_ARRAY_SIZE(api_consts); i++) {
        mp_obj_dict_store(g, MP_OBJ_NEW_QSTR(qstr_from_str(api_consts[i].name)),
                          MP_OBJ_NEW_SMALL_INT(api_consts[i].value));
    }
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

void script_init(void *stack_top)
{
    mp_embed_init(&heap[0], sizeof(heap), stack_top); //pass the pointer, not the value
    setup_api();
    mp_embed_exec_str(user_script);
    call_py(qstr_from_str("init"));
}

void script_update(void) {
    call_py(qstr_from_str("update"));
}

void script_deinit()
{
    mp_embed_deinit();
}

void mp_hal_stdout_tx_strn(const char *str, size_t len) {
    fwrite(str, 1, len, stdout);
    fflush(stdout);
}