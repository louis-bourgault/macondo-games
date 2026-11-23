#include "mp.h"
#include "port/micropython_embed.h"
#include "display.h"
#include "input.h"
#include "py/runtime.h"
// #include "py/mprint.h"
#include "py/nlr.h"
#include <stdio.h>
#include "py/qstr.h"
#include "lfs.h"
#include "fs.h"
#include "py/lexer.h"
#include "py/objmodule.h"
#include "py/builtin.h"


static char heap[32 * 1024];

// //TODO: change to a filesystem script
// static const char user_script[] =
// "from ferret import *\n"
// "def init():\n"
// "    fill_display(0xF800)\n"
// "    \n"
// "def update():\n"
// "    print('micropython is alive weewoo')\n";

static const char user_script[] =
"from ferret import *\n"
"from test import *\n"
"def init():\n"
"    fill()\n"
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
static MP_DEFINE_CONST_FUN_OBJ_1(py_fill_display_obj, py_fill_display);
static mp_obj_t py_draw_pixel(mp_obj_t x, mp_obj_t y, mp_obj_t colour)
{
    int16_t x_val = (int16_t)mp_obj_get_int(x);
    int16_t y_val = (int16_t)mp_obj_get_int(y);
    uint16_t c = (uint16_t)mp_obj_get_int(colour);
    draw_pixel(x_val, y_val, c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_3(py_draw_pixel_obj, py_draw_pixel);
static mp_obj_t py_draw_rect(size_t n_args, const mp_obj_t *args)
{
    (void)n_args;
    int16_t x0_val = (int16_t)mp_obj_get_int(args[0]);
    int16_t y0_val = (int16_t)mp_obj_get_int(args[1]);
    int16_t x1_val = (int16_t)mp_obj_get_int(args[2]);
    int16_t y1_val = (int16_t)mp_obj_get_int(args[3]);
    uint16_t c = (uint16_t)mp_obj_get_int(args[4]);
    draw_rect(x0_val, y0_val, x1_val, y1_val, c);
    return mp_const_none;
}
static MP_DEFINE_CONST_FUN_OBJ_VAR_BETWEEN(py_draw_rect_obj, 5, 5, py_draw_rect); // 5 args, min 5, max 5
static mp_obj_t py_button_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError(MP_ERROR_TEXT("that button does not exist"));
    }
    return mp_obj_new_bool(button_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(py_button_pressed_obj, py_button_pressed);
static mp_obj_t py_button_just_pressed(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError(MP_ERROR_TEXT("that button does not exist"));
    }
    return mp_obj_new_bool(button_just_pressed((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(py_button_just_pressed_obj, py_button_just_pressed);
static mp_obj_t py_button_just_released(mp_obj_t button)
{
    int btn = mp_obj_get_int(button);
    if (btn < 0 || btn >= BTN_COUNT)
    {
        mp_raise_ValueError(MP_ERROR_TEXT("that button does not exist"));
    }
    return mp_obj_new_bool(button_just_released((Button)btn));
}
static MP_DEFINE_CONST_FUN_OBJ_1(py_button_just_released_obj, py_button_just_released);


typedef struct { const char *name; mp_obj_t fn; } api_fn_t;
typedef struct { const char *name; int value; } api_const_t;

static const api_fn_t api_fns[] = {
    { "fill_display",         MP_OBJ_FROM_PTR(&py_fill_display_obj) },
    { "draw_pixel",           MP_OBJ_FROM_PTR(&py_draw_pixel_obj) },
    { "draw_rect",            MP_OBJ_FROM_PTR(&py_draw_rect_obj) },
    { "button_pressed",       MP_OBJ_FROM_PTR(&py_button_pressed_obj) },
    { "button_just_pressed",  MP_OBJ_FROM_PTR(&py_button_just_pressed_obj) },
    { "button_just_released", MP_OBJ_FROM_PTR(&py_button_just_released_obj) },
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
                          api_fns[i].fn);
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
    printf("script (%u bytes):\n%s\n---\n", (unsigned)strlen(user_script), user_script);
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

typedef struct _mp_reader_lfs_t {
    lfs_file_t file;
    struct lfs_file_config file_cfg;
    uint8_t lfs_cache[LFS_CACHE_SIZE]; // Static cache for LittleFS (avoids malloc)
    uint8_t read_buf[LFS_CACHE_SIZE];  // Read-ahead buffer for MicroPython lexer
    lfs_ssize_t buf_len;
    lfs_ssize_t buf_pos;
    bool is_open;
} mp_reader_lfs_t;

static mp_uint_t lfs_reader_readbyte(void *data) {
    mp_reader_lfs_t *reader = (mp_reader_lfs_t *)data;

    if (reader->buf_pos >= reader->buf_len) {
        //if we add mutexes later, lock here
        reader->buf_len = lfs_file_read(&lfs, &reader->file, reader->read_buf, sizeof(reader->read_buf));
        //would unlock here
        if (reader->buf_len <= 0) {
            return MP_READER_EOF;
        }
        reader->buf_pos = 0;
    }

    return reader->read_buf[reader->buf_pos++];
}

static void lfs_reader_close(void *data) {
    mp_reader_lfs_t *reader = (mp_reader_lfs_t *)data;
    if (reader->is_open) {
        // mutex thing here
        lfs_file_close(&lfs, &reader->file);
        //unlock
        reader->is_open = false;
    }
}


mp_import_stat_t mp_import_stat(const char *path) {
    printf("mp_import_stat: checking path '%s'\n", path);
    struct lfs_info info;
    if (lfs_stat(&lfs, path, &info) < 0) return MP_IMPORT_STAT_NO_EXIST;
    return info.type == LFS_TYPE_DIR ? MP_IMPORT_STAT_DIR : MP_IMPORT_STAT_FILE;
}



mp_lexer_t *mp_lexer_new_from_file(qstr filename) {
    printf("mp_lexer_new_from_file: opening file '%s'\n", qstr_str(filename));
    const char *path = qstr_str(filename);

    // put it on the heap
    mp_reader_lfs_t *reader = m_new_obj(mp_reader_lfs_t);
    memset(&reader->file_cfg, 0, sizeof(struct lfs_file_config));
    reader->file_cfg.buffer = reader->lfs_cache;
    reader->buf_len = 0;
    reader->buf_pos = 0;
    reader->is_open = false;

    // mutex thing here if we add mutexes later
    int err = lfs_file_opencfg(&lfs, &reader->file, path, LFS_O_RDONLY, &reader->file_cfg);
    // release mutex

    if (err < 0) {
        mp_raise_OSError(-err);
    }

    reader->is_open = true;

    mp_reader_t mp_reader = {
        .data = reader,
        .readbyte = lfs_reader_readbyte,
        .close = lfs_reader_close,
    };

    return mp_lexer_new(filename, mp_reader);
}