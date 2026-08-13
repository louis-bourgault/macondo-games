// flash_storage.c — the Ferretboard internal filesystem, mounted at "/".
//
// The raw flash is owned by Go (flash_common.go / flash_device.go /
// flash_storage.go: a 4 KiB-block device over the reserved 8 MiB region),
// exposed to C through the //export funcs ferret_flash_read/write/erase/
// block_count/block_size. This file wraps those exports in a MicroPython
// "Flash" block device (readblocks/writeblocks/ioctl), and bootstraps the
// LittleFS2 filesystem that host.c mounts as "/":
//
//   - first boot: the region is blank 0xFF, so detect the absent "littlefs"
//     superblock and format the volume with lfs2_format before mounting;
//   - subsequent boots: mount directly (VfsLfs2 make_new mounts eagerly).
//
// Ferretboard-specific: no make_new on the Flash type — the device is a
// singleton whose size/layout is fixed by Go, so I/O errors from the exports
// are the only failure modes.

#include <stdint.h>
#include <string.h>

#include "py/runtime.h"
#include "py/obj.h"
#include "py/stream.h"
#include "py/mperrno.h"
#include "extmod/vfs.h"
#include "extmod/vfs_lfs.h"
#include "lib/littlefs/lfs2.h"
#include "ferret_abi.h"


#ifndef STATIC
#define STATIC static
#endif

// --- the Flash block device object (talks to the Go //export funcs) ---------

STATIC mp_obj_t t_flash_readblocks(size_t n_args, const mp_obj_t *args) {
    mp_buffer_info_t bufinfo;
    mp_get_buffer_raise(args[2], &bufinfo, MP_BUFFER_WRITE);
    uint32_t off = (n_args == 4) ? (uint32_t)mp_obj_get_int(args[3]) : 0;
    if (ferret_flash_read((uint32_t)mp_obj_get_int(args[1]), off,
                          (uint8_t *)bufinfo.buf, bufinfo.len) != 0) {
        mp_raise_OSError(MP_EIO);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR_BETWEEN(t_flash_readblocks_obj, 3, 4, t_flash_readblocks);

STATIC mp_obj_t t_flash_writeblocks(size_t n_args, const mp_obj_t *args) {
    mp_buffer_info_t bufinfo;
    mp_get_buffer_raise(args[2], &bufinfo, MP_BUFFER_READ);
    uint32_t off = (n_args == 4) ? (uint32_t)mp_obj_get_int(args[3]) : 0;
    if (ferret_flash_write((uint32_t)mp_obj_get_int(args[1]), off,
                           (uint8_t *)bufinfo.buf, bufinfo.len) != 0) {
        mp_raise_OSError(MP_EIO);
    }
    return mp_const_none;
}
STATIC MP_DEFINE_CONST_FUN_OBJ_VAR_BETWEEN(t_flash_writeblocks_obj, 3, 4, t_flash_writeblocks);

STATIC mp_obj_t t_flash_ioctl(mp_obj_t self_in, mp_obj_t cmd_in, mp_obj_t arg_in) {
    (void)self_in;
    switch (mp_obj_get_int(cmd_in)) {
        case MP_BLOCKDEV_IOCTL_INIT:
        case MP_BLOCKDEV_IOCTL_DEINIT:
        case MP_BLOCKDEV_IOCTL_SYNC:
            return MP_OBJ_NEW_SMALL_INT(0);
        case MP_BLOCKDEV_IOCTL_BLOCK_COUNT:
            return MP_OBJ_NEW_SMALL_INT(ferret_flash_block_count());
        case MP_BLOCKDEV_IOCTL_BLOCK_SIZE:
            return MP_OBJ_NEW_SMALL_INT(ferret_flash_block_size());
        case MP_BLOCKDEV_IOCTL_BLOCK_ERASE: {
            if (ferret_flash_erase((uint32_t)mp_obj_get_int(arg_in)) != 0) {
                mp_raise_OSError(MP_EIO);
            }
            return MP_OBJ_NEW_SMALL_INT(0);
        }
        default:
            return mp_const_none;
    }
}
STATIC MP_DEFINE_CONST_FUN_OBJ_3(t_flash_ioctl_obj, t_flash_ioctl);

STATIC const mp_rom_map_elem_t t_flash_locals_dict_table[] = {
    { MP_ROM_QSTR(MP_QSTR_readblocks), MP_ROM_PTR(&t_flash_readblocks_obj) },
    { MP_ROM_QSTR(MP_QSTR_writeblocks), MP_ROM_PTR(&t_flash_writeblocks_obj) },
    { MP_ROM_QSTR(MP_QSTR_ioctl), MP_ROM_PTR(&t_flash_ioctl_obj) },
};
STATIC MP_DEFINE_CONST_DICT(t_flash_locals_dict, t_flash_locals_dict_table);

MP_DEFINE_CONST_OBJ_TYPE(
    ferret_flash_type,
    MP_QSTR_Flash,
    MP_TYPE_FLAG_NONE,
    locals_dict, &t_flash_locals_dict
    );

// The singleton block device; mounted as "/" by ferret_fs_init().
STATIC const mp_obj_base_t ferret_flash_obj = { &ferret_flash_type };

// --- first-boot detection and formatting (no Python-level exceptions) -------

// LittleFS2 metadata-pair superblock layout in block 0 of the volume:
//   0..3   revision counter (0x000003eb after format)
//   4..7   magic 0xfffffff0 (bit-reversed)
//   8..15  the "littlefs" superblock name
STATIC bool ferret_flash_has_fs(void) {
    uint8_t sb[24];
    if (ferret_flash_read(0, 0, sb, sizeof(sb)) != 0) {
        return false;
    }
    return memcmp(sb + 8, "littlefs", 8) == 0;
}

STATIC int ferret_lfs_read(const struct lfs2_config *c, lfs2_block_t block,
                           lfs2_off_t off, void *buffer, lfs2_size_t size) {
    return ferret_flash_read(block, off, (uint8_t *)buffer, size) == 0 ? 0 : LFS2_ERR_IO;
}

STATIC int ferret_lfs_prog(const struct lfs2_config *c, lfs2_block_t block,
                           lfs2_off_t off, const void *buffer, lfs2_size_t size) {
    return ferret_flash_write(block, off, (uint8_t *)buffer, size) == 0 ? 0 : LFS2_ERR_IO;
}

STATIC int ferret_lfs_erase(const struct lfs2_config *c, lfs2_block_t block) {
    return ferret_flash_erase(block) == 0 ? 0 : LFS2_ERR_IO;
}

STATIC int ferret_lfs_sync(const struct lfs2_config *c) {
    (void)c;
    return 0;
}

// Format the region as LittleFS2. Only called before first mount; a blank or
// corrupt volume is rebuilt rather than surfacing a mount error.
STATIC void ferret_fs_format(void) {
    struct lfs2_config cfg;
    memset(&cfg, 0, sizeof(cfg));
    cfg.read = ferret_lfs_read;
    cfg.prog = ferret_lfs_prog;
    cfg.erase = ferret_lfs_erase;
    cfg.sync = ferret_lfs_sync;
    cfg.read_size = 32;
    cfg.prog_size = 32;
    cfg.block_size = ferret_flash_block_size();
    cfg.block_count = ferret_flash_block_count();
    cfg.cache_size = 256;
    cfg.lookahead_size = 256;
    cfg.block_cycles = 500;

    lfs2_t fs;
    if (lfs2_format(&fs, &cfg) != 0) {
        mp_raise_OSError(MP_EIO);
    }
}

// --- boot-time mount (called from host.c after every mp_embed_init) ---------

void ferret_fs_init(void) {
    if (!ferret_flash_has_fs()) {
        ferret_fs_format();
    }

    // VfsLfs2 make_new mounts eagerly, so a formatted volume is required.
    // Instantiate the class the way Python does (VfsLfs2(flash)), via its
    // make_new slot: mp_vfs_lfs2_make_new is static in vfs_lfs.c.
    mp_obj_t bdev = MP_OBJ_FROM_PTR(&ferret_flash_obj);
    mp_obj_t args[2];
    args[0] = mp_call_function_n_kw(MP_OBJ_FROM_PTR(&mp_type_vfs_lfs2), 1, 0, &bdev);
    args[1] = MP_OBJ_NEW_QSTR(qstr_from_str("/"));
    mp_vfs_mount(2, args, (mp_map_t *)&mp_const_empty_map);
}