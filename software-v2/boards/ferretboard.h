#ifndef _BOARDS_FERRETBOARD_H
#define _BOARDS_FERRETBOARD_H

// Must appear before including boards/pico.h.
pico_board_cmake_set(PICO_PLATFORM, rp2040)

#include "boards/pico.h"

// This is a 16 MiB W25Q128
pico_board_cmake_set_default(PICO_FLASH_SIZE_BYTES, (16 * 1024 * 1024))

#undef PICO_FLASH_SIZE_BYTES
#define PICO_FLASH_SIZE_BYTES (16 * 1024 * 1024)

// Keep the startup setting currently used by your firmware.
#ifndef PICO_XOSC_STARTUP_DELAY_MULTIPLIER
#define PICO_XOSC_STARTUP_DELAY_MULTIPLIER 64
#endif

#endif