#include "display.h"
#include <stdbool.h>

#include "hardware/gpio.h"
#include "hardware/spi.h"
#include "pico/stdlib.h"

#define DISPLAY_WIDTH  240
#define DISPLAY_HEIGHT 240

#define PIN_SCK  18
#define PIN_MOSI 19
#define PIN_CS   17
#define PIN_RST  20
#define PIN_DC   21

static uint16_t framebuffer[DISPLAY_WIDTH * DISPLAY_HEIGHT];
static bool initialized = false;


static void begin_write(void) {
    gpio_put(PIN_CS, 0);
    sleep_us(1);
}

static void end_write(void) {
    gpio_put(PIN_CS, 1);
    sleep_us(1);
}

static void write_command(uint8_t command) {
    gpio_put(PIN_DC, 0);
    sleep_us(1);
    spi_write_blocking(spi0, &command, 1);
    sleep_us(1);
}

static void write_data(const uint8_t *data, size_t length) {
    gpio_put(PIN_DC, 1);
    sleep_us(1);
    spi_write_blocking(spi0, data, length);
    sleep_us(1);
}

static void write_command_with_data(
    uint8_t command,
    const uint8_t *data,
    size_t length
) {
    begin_write();
    write_command(command);
    if (length != 0) {
        write_data(data, length);
    }
    end_write();
}

static void set_window_and_start_memory_write(void) {
    const uint8_t columns[] = {0x00, 0x00, 0x00, 0xEF};
    const uint8_t rows[]    = {0x00, 0x00, 0x00, 0xEF};

   
    begin_write();

    write_command(0x2A); // CASET
    write_data(columns, sizeof(columns));

    write_command(0x2B); // RASET
    write_data(rows, sizeof(rows));

    write_command(0x2C); // RAMWR
    gpio_put(PIN_DC, 1);
    sleep_us(1);
}

void display_init(void) {
    
    gpio_init(PIN_CS);
    gpio_set_dir(PIN_CS, GPIO_OUT);
    gpio_put(PIN_CS, 1);

    gpio_init(PIN_DC);
    gpio_set_dir(PIN_DC, GPIO_OUT);
    gpio_put(PIN_DC, 0);

    gpio_init(PIN_RST);
    gpio_set_dir(PIN_RST, GPIO_OUT);
    gpio_put(PIN_RST, 0);

    
    sleep_ms(200);

    spi_init(spi0, 4 * 1000 * 1000);
    gpio_set_function(PIN_SCK, GPIO_FUNC_SPI);
    gpio_set_function(PIN_MOSI, GPIO_FUNC_SPI);

    gpio_put(PIN_RST, 1);
    sleep_ms(150);

    write_command_with_data(0x01, NULL, 0); // SWRESET
    sleep_ms(150);

    write_command_with_data(0x11, NULL, 0); // SLPOUT
    sleep_ms(120);

    const uint8_t color_mode[] = {0x55};
    write_command_with_data(0x3A, color_mode, sizeof(color_mode)); // RGB565

    const uint8_t madctl[] = {0x00};
    write_command_with_data(0x36, madctl, sizeof(madctl));

    const uint8_t porch[] = {0x0C, 0x0C, 0x00, 0x33, 0x33};
    write_command_with_data(0xB2, porch, sizeof(porch)); // PORCTRL

    const uint8_t frame_rate[] = {0x0F};
    write_command_with_data(0xC6, frame_rate, sizeof(frame_rate)); // FRCTRL2

    write_command_with_data(0x21, NULL, 0); // INVON
    sleep_ms(10);

    write_command_with_data(0x13, NULL, 0); // NORON
    sleep_ms(10);

    write_command_with_data(0x29, NULL, 0); // DISPON
    sleep_ms(10);

    for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++) {
        framebuffer[i] = 0x0000;
    }

    initialized = true;
    display_present();
}

void display_fill(uint16_t color) {
    if (!initialized) {
        return;
    }

    for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++) {
        framebuffer[i] = color;
    }
}

void display_pixel(int x, int y, uint16_t color) {
    if (!initialized ||
        x < 0 || x >= DISPLAY_WIDTH ||
        y < 0 || y >= DISPLAY_HEIGHT) {
        return;
    }

    framebuffer[y * DISPLAY_WIDTH + x] = color;
}

void display_present(void) {
    if (!initialized) {
        return;
    }

    set_window_and_start_memory_write();

    // CS is already low, DC is already high.
    for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++) {
        uint16_t color = framebuffer[i];
        uint8_t bytes[] = {
            (uint8_t)(color >> 8),
            (uint8_t)(color & 0xFF),
        };

        
        spi_write_blocking(spi0, bytes, 2);
    }

    end_write();
}
