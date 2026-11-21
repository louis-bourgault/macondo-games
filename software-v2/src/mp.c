#include "mp.h"
#include "port/micropython_embed.h"
#include "display.h"
#include "input.h"
#include "py/runtime.h"
// #include "py/mprint.h"
#include "py/nlr.h"
#include <stdio.h>
#include "py/qstr.h"

static char heap[8 * 1024];

static const char user_script[] =
"def init():\n"
"    fill_display(0xF800)\n"
"    \n"
"def update():\n"
"    print('micropython is alive weewoo')\n";

//define the constants for all the micropython buttons
static const mp_rom_map_elem_t button_module_globals_table[] = {
    {MP_ROM_QSTR(MP_QSTR_BTN_A), MP_ROM_INT(0)},
    {MP_ROM_QSTR(MP_QSTR_BTN_B), MP_ROM_INT(1)},
    {MP_ROM_QSTR(MP_QSTR_BTN_X), MP_ROM_INT(2)},
    {MP_ROM_QSTR(MP_QSTR_BTN_Y), MP_ROM_INT(3)},
    {MP_ROM_QSTR(MP_QSTR_BTN_UP), MP_ROM_INT(4)},
    {MP_ROM_QSTR(MP_QSTR_BTN_DOWN), MP_ROM_INT(5)},
    {MP_ROM_QSTR(MP_QSTR_BTN_LEFT), MP_ROM_INT(6)},
    {MP_ROM_QSTR(MP_QSTR_BTN_RIGHT), MP_ROM_INT(7)},
    {MP_ROM_QSTR(MP_QSTR_BTN_MENU), MP_ROM_INT(8)},
};
static MP_DEFINE_CONST_DICT(button_module_globals, button_module_globals_table);
//can access these buttons from micropython with button.BTN_A, button.BTN_B, etc

//FUNCTIONS FOR MICROPYTHON TO CALL
static mp_obj_t fill_display(mp_obj_t colour)
{
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    display_fill(c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_1(fill_display_obj, fill_display);
static mp_obj_t draw_pixel(mp_obj_t x, mp_obj_t y, mp_obj_t colour)
{
    int16_t x_val = (int16_t)mp_obj_get_int(x);
    int16_t y_val = (int16_t)mp_obj_get_int(y);
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    draw_pixel(x_val, y_val, c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_3(draw_pixel_obj, draw_pixel);
static mp_obj_t draw_rect(mp_obj_t x0, mp_obj_t y0, mp_obj_t x1, mp_obj_t y1, mp_obj_t colour)
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

static mp_obj_t button_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_pressed_obj, button_pressed);
static mp_obj_t button_just_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_just_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_just_pressed_obj, button_just_pressed);
static mp_obj_t button_just_released(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError("that button does not exist");
    }
    return mp_obj_new_bool(button_just_released((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(button_just_released_obj, button_just_released);



static void setup_api(void)
{
    mp_store_global(qstr_from_str("fill_display"), MP_OBJ_FROM_PTR(&fill_display_obj));
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
    mp_embed_init(&heap[0], sizeof(heap), &stack_top);
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