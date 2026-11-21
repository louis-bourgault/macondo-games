cmake_minimum_required(VERSION 3.13)

include($ENV{PICO_SDK_PATH}/external/pico_sdk_import.cmake)
project(ferretboard_v2 C CXX ASM)
set(CMAKE_C_STANDARD 11)
set(CMAKE_CXX_STANDARD 17)

pico_sdk_init()
add_compile_definitions(PICO_XOSC_STARTUP_DELAY_MULTIPLIER=64)

set(MP_EMBED_DIR "$ENV{HOME}/src/micropython/examples/embedding/micropython_embed")
file(GLOB_RECURSE MICROPYTHON_SOURCES "${MP_EMBED_DIR}/*.c")

add_library(micropython STATIC ${MICROPYTHON_SOURCES})
target_include_directories(micropython PUBLIC
    "${MP_EMBED_DIR}"
    "${MP_EMBED_DIR}/port"
    "${CMAKE_CURRENT_SOURCE_DIR}/src"
)

add_executable(ferret
    src/main.c
    src/display.c
    src/input.c
)


target_link_libraries(ferret
    pico_stdlib
    hardware_spi
    hardware_dma
    hardware_flash
    micropython
)

pico_enable_stdio_usb(ferret 1)
pico_enable_stdio_uart(ferret 0)
pico_add_extra_outputs(ferret)
