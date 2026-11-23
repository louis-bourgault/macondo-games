#include "lfs.h"
#include "hardware/flash.h"
#include "hardware/sync.h"

#define FS_OFFSET (2 * 1024 * 1024)
#define FS_SIZE (14 * 1024 * 1024) //14MB for images and scripts, 2mb for C firmware.

int pico_lfs_read(const struct lfs_config *c, lfs_block_t block,
    lfs_off_t off, void *buffer, lfs_size_t size) {
uint32_t addr = XIP_BASE + FS_OFFSET + (block * c->block_size) + off;
memcpy(buffer, (const void *)addr, size);
return 0;
}

int pico_lfs_prog(const struct lfs_config *c, lfs_block_t block,
    lfs_off_t off, const void *buffer, lfs_size_t size) {
uint32_t offset = FS_OFFSET + (block * c->block_size) + off;
uint32_t ints = save_and_disable_interrupts();
flash_range_program(offset, (const uint8_t *)buffer, size);
restore_interrupts(ints);
return 0;
}

int pico_lfs_erase(const struct lfs_config *c, lfs_block_t block) {
uint32_t offset = FS_OFFSET + (block * c->block_size);
uint32_t ints = save_and_disable_interrupts();
flash_range_erase(offset, c->block_size);
restore_interrupts(ints);
return 0;
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
