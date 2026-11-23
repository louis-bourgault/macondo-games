#include "lfs.h"
#include "fs.h"
#include "hardware/flash.h"
#include "hardware/sync.h"
#include <string.h>
#include "pico/stdlib.h"
#include "pico/stdio_usb.h"
#include "pico/flash.h"

#include <stdio.h>

#define FS_OFFSET (1 * 1024 * 1024)
#define FS_SIZE (14 * 1024 * 1024) //14MB for images and scripts, 2mb for C firmware.

typedef struct {
    uint32_t offset;
    const uint8_t *data;
    size_t size;
} flash_op_t;

static void do_erase(void *p) {
    flash_op_t *op = p;
    flash_range_erase(op->offset, op->size);
}

static void do_prog(void *p) {
    flash_op_t *op = p;
    flash_range_program(op->offset, op->data, op->size);
}


int pico_lfs_read(const struct lfs_config *c, lfs_block_t block,
    lfs_off_t off, void *buffer, lfs_size_t size) {
uint32_t addr = XIP_BASE + FS_OFFSET + (block * c->block_size) + off;
memcpy(buffer, (const void *)addr, size);
return 0;
}

int pico_lfs_prog(const struct lfs_config *c, lfs_block_t block,
    lfs_off_t off, const void *buffer, lfs_size_t size) {
flash_op_t op = { FS_OFFSET + block * c->block_size + off, buffer, size };
return flash_safe_execute(do_prog, &op, UINT32_MAX) == PICO_OK ? 0 : LFS_ERR_IO;
}

int pico_lfs_erase(const struct lfs_config *c, lfs_block_t block) {
    printf("pico_lfs_erase: erasing block %u\n", block);
flash_op_t op = { FS_OFFSET + block * c->block_size, NULL, c->block_size };
int fserr =  flash_safe_execute(do_erase, &op, 1000) == PICO_OK ? 0 : LFS_ERR_IO;
printf("pico_lfs_erase: flash_safe_execute returned %d\n", fserr);
return fserr;
}

int pico_lfs_sync(const struct lfs_config *c) {
//dont need this for rp2040/rp2350
return 0;
}


lfs_t lfs;
lfs_file_t file;

uint8_t read_buffer[256];
uint8_t prog_buffer[256];
uint8_t lookahead_buffer[32];


//storage chip is the W25Q128JVS, connected over QSPI.
const struct lfs_config cfg = {
    .read = pico_lfs_read,
    .prog = pico_lfs_prog,
    .erase = pico_lfs_erase,
    .sync = pico_lfs_sync,

    .read_buffer = read_buffer,
    .prog_buffer = prog_buffer,
    .lookahead_buffer = lookahead_buffer,

    .read_size = 1,
    .prog_size = 256,
    .block_size = 4096,
    .block_count = FS_SIZE / 4096,
    .lookahead_size = 32,
    .block_cycles = 500,
    .cache_size = 256,
    .name_max = 255,
};

int fs_init() {
    printf("fs_init: initializing filesystem...\n");

    int err = lfs_mount(&lfs, &cfg);
    printf("fs_init: lfs_mount returned %d\n", err);
    if (err < 0) {
        //first boot, we format it. maybe later we should put a warning before doing it so the user can decide to format or not
        err = lfs_format(&lfs, &cfg);
        printf("fs_init: lfs_format returned %d\n", err);
        if (err < 0) {
            // prob shoudl handle this, format failed
            //TODO: handle error
        }
        err = lfs_mount(&lfs, &cfg);
        printf("fs_init: lfs_mount returned %d\n", err);
        if (err < 0) {
            // mount error
            //TODO: handle error
        } 
    }
    return err;
}