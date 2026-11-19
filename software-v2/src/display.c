#include "display.h"
#include <stdbool.h>

#include "hardware/gpio.h"
#include "hardware/spi.h"
#include "pico/stdlib.h"
#include <limits.h>

#define DISPLAY_WIDTH 240
#define DISPLAY_HEIGHT 240

#define PIN_SCK 18
#define PIN_MOSI 19
#define PIN_CS 17
#define PIN_RST 20
#define PIN_DC 21
#define SPI_SPEED_HZ 62500000u
#define SAVINGS_THRESHOLD 80
#define MAX_DIRTY_RECTS 8

static uint16_t framebuffer[DISPLAY_WIDTH * DISPLAY_HEIGHT];
static bool initialized = false;

typedef struct
{
    int16_t x0;
    int16_t y0;
    int16_t x1;
    int16_t y1;
} rect_t;

rect_t dirty_rects[MAX_DIRTY_RECTS];
static uint8_t dirty_rect_count = 0;

int32_t rect_area(int16_t x0, int16_t y0, int16_t x1, int16_t y1)
{
    // since the multiplying unit is 32bit, theres actually no added cpu complexity in making it int32
    int32_t w = x1 - x0;
    int32_t h = y1 - y0;
    return w * h;
}

static void dirty_add(int16_t x0, int16_t y0, int16_t x1, int16_t y1)
{
    // check the most obvious fitting within ones that already exist.
    int8_t fastest_dirty = -1;
    int32_t least_added = INT32_MAX;
    int32_t individualArea = rect_area(x0, y0, x1, y1);
    if (individualArea <= 0) {
        return;
    }
    if (dirty_rect_count == 0)
    {
        dirty_rects[0].x0 = x0;
        dirty_rects[0].y0 = y0;
        dirty_rects[0].x1 = x1;
        dirty_rects[0].y1 = y1;
        dirty_rect_count++;
        return;
    }
    for (int8_t i = 0; i < dirty_rect_count; i++)
    {
        if (x0 >= dirty_rects[i].x0 && y0 >= dirty_rects[i].y0 && x1 <= dirty_rects[i].x1 && y1 <= dirty_rects[i].y1)
        {
            return;
        };
        int16_t lesserX;
        int16_t lesserY;
        int16_t greaterX;
        int16_t greaterY;
        if (x0 <= dirty_rects[i].x0)
        {
            lesserX = x0;
        }
        else
        {
            lesserX = dirty_rects[i].x0;
        }
        if (x1 >= dirty_rects[i].x1)
        {
            greaterX = x1;
        }
        else
        {
            greaterX = dirty_rects[i].x1;
        }
        if (y0 <= dirty_rects[i].y0)
        {
            lesserY = y0;
        }
        else
        {
            lesserY = dirty_rects[i].y0;
        }
        if (y1 >= dirty_rects[i].y1)
        {
            greaterY = y1;
        }
        else
        {
            greaterY = dirty_rects[i].y1;
        }
        int32_t totalArea = rect_area(lesserX, lesserY, greaterX, greaterY);
        int32_t dirty_rect_original_area = rect_area(dirty_rects[i].x0, dirty_rects[i].y0, dirty_rects[i].x1, dirty_rects[i].y1);
        int32_t added_pixels = totalArea - dirty_rect_original_area - individualArea;
        if (added_pixels < least_added)
        {
            least_added = added_pixels;
            fastest_dirty = i;
        }
    }
    if (least_added > SAVINGS_THRESHOLD && dirty_rect_count < MAX_DIRTY_RECTS)
    {
        // add a new dirty rect with the boundary of the new rect
        dirty_rects[dirty_rect_count].x0 = x0;
        dirty_rects[dirty_rect_count].y0 = y0;
        dirty_rects[dirty_rect_count].x1 = x1;
        dirty_rects[dirty_rect_count].y1 = y1;
        dirty_rect_count++;
    }
    else
    {
        // if this runs, EITHER the new rect is close enough to be added to another already
        // dirty rect, OR we're at max capacity, so we're forced to.
        if (fastest_dirty < 0)
        {
            fastest_dirty = 0;
        }
        if (x0 <= dirty_rects[fastest_dirty].x0)
        {
            dirty_rects[fastest_dirty].x0 = x0;
        }
        if (x1 >= dirty_rects[fastest_dirty].x1)
        {
            dirty_rects[fastest_dirty].x1 = x1;
        }
        if (y0 <= dirty_rects[fastest_dirty].y0)
        {
            dirty_rects[fastest_dirty].y0 = y0;
        }
        if (y1 >= dirty_rects[fastest_dirty].y1)
        {
            dirty_rects[fastest_dirty].y1 = y1;
        }
        // we might have expanded the dirty rects, so maybe they overlap now so we should merge them
        // under no circumstances shoudl the same pixel be drawn twice.
        for (int8_t i = 0; i < dirty_rect_count; i++)
        {
            if (i == fastest_dirty)
            {
                continue;
            }
            if (dirty_rects[i].x0 <= dirty_rects[fastest_dirty].x1 && dirty_rects[i].x1 >= dirty_rects[fastest_dirty].x0 && dirty_rects[i].y0 <= dirty_rects[fastest_dirty].y1 && dirty_rects[i].y1 >= dirty_rects[fastest_dirty].y0)
            {
                // merge them
                if (dirty_rects[i].x0 <= dirty_rects[fastest_dirty].x0)
                {
                    dirty_rects[fastest_dirty].x0 = dirty_rects[i].x0;
                }
                if (dirty_rects[i].x1 >= dirty_rects[fastest_dirty].x1)
                {
                    dirty_rects[fastest_dirty].x1 = dirty_rects[i].x1;
                }
                if (dirty_rects[i].y0 <= dirty_rects[fastest_dirty].y0)
                {
                    dirty_rects[fastest_dirty].y0 = dirty_rects[i].y0;
                }
                if (dirty_rects[i].y1 >= dirty_rects[fastest_dirty].y1)
                {
                    dirty_rects[fastest_dirty].y1 = dirty_rects[i].y1;
                }
                // remove the other rect from the list
                for (int8_t j = i; j + 1 < dirty_rect_count; j++)
                {
                    dirty_rects[j] = dirty_rects[j + 1];
                }
                dirty_rect_count--;
                if (i < fastest_dirty)
                {
                    fastest_dirty--;
                }
                i--;
            }
        }
        // TODO: examine this for efficiency, since it is o(n^2).
        // should be fine as long as MAX_DIRTY_RECTS stays small
    }
}

//stupid little helper functions
//basically here becuase i forget what all the little instructions do
static void begin_write(void)
{
    gpio_put(PIN_CS, 0);
    // sleep_us(1);
}

static void end_write(void)
{
    gpio_put(PIN_CS, 1);
    // sleep_us(1);
}

static void write_command(uint8_t command)
{
    gpio_put(PIN_DC, 0);
    // sleep_us(1);
    spi_write_blocking(spi0, &command, 1);
    // sleep_us(1);
}

static void write_data(const uint8_t *data, size_t length)
{
    gpio_put(PIN_DC, 1);
    // sleep_us(1);
    spi_write_blocking(spi0, data, length);
    // sleep_us(1);
}

static void write_command_with_data(
    uint8_t command,
    const uint8_t *data,
    size_t length)
{
    begin_write();
    write_command(command);
    if (length != 0)
    {
        write_data(data, length);
    }
    end_write();
}



void display_init(void)
{

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

    spi_init(spi0, SPI_SPEED_HZ);
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

    for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++)
    {
        framebuffer[i] = 0x0000;
    }
    dirty_add(0, 0, DISPLAY_WIDTH, DISPLAY_HEIGHT); //so that we actually init the screen
    

    initialized = true;
    display_present();
}

void display_fill(uint16_t color)
{
    if (!initialized)
    {
        return;
    }

    for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++)
    {
        framebuffer[i] = color;
    }
    dirty_add(0, 0, DISPLAY_WIDTH, DISPLAY_HEIGHT);
}

void draw_pixel(int x, int y, uint16_t color)
{
    if (!initialized ||
        x < 0 || x >= DISPLAY_WIDTH ||
        y < 0 || y >= DISPLAY_HEIGHT)
    {
        return;
    }

    framebuffer[y * DISPLAY_WIDTH + x] = color;
    dirty_add(x, y, x + 1, y + 1);
}

void draw_rect(int x0, int y0, int x1, int y1, uint16_t color)
{
    if (!initialized)
    {
        return;
    }

    if (x0 < 0)
        x0 = 0;
    if (y0 < 0)
        y0 = 0;
    if (x1 > DISPLAY_WIDTH)
        x1 = DISPLAY_WIDTH;
    if (y1 > DISPLAY_HEIGHT)
        y1 = DISPLAY_HEIGHT;

    for (int y = y0; y < y1; y++)
    {
        for (int x = x0; x < x1; x++)
        {
            framebuffer[y * DISPLAY_WIDTH + x] = color;
        }
    }
    dirty_add(x0, y0, x1, y1);
}

// static void set_window_and_start_memory_write(void)
// {
//     const uint8_t columns[] = {0x00, 0x00, 0x00, 0xEF};
//     const uint8_t rows[] = {0x00, 0x00, 0x00, 0xEF};

//     begin_write();

//     write_command(0x2A); // CASET
//     write_data(columns, sizeof(columns));

//     write_command(0x2B); // RASET
//     write_data(rows, sizeof(rows));

//     write_command(0x2C); // RAMWR
//     gpio_put(PIN_DC, 1);
//     sleep_us(1);
// }

static void set_dirty_rectangle_and_draw(rect_t rect) {
    uint16_t x_start = (uint16_t)rect.x0;
    uint16_t x_end   = (uint16_t)(rect.x1 - 1);
    uint16_t y_start = (uint16_t)rect.y0;
    uint16_t y_end   = (uint16_t)(rect.y1 - 1);

    const uint8_t columns[] = {
        (uint8_t)(x_start >> 8), (uint8_t)(x_start & 0xFF),
        (uint8_t)(x_end >> 8),   (uint8_t)(x_end & 0xFF),
    };
    const uint8_t rows[] = {
        (uint8_t)(y_start >> 8), (uint8_t)(y_start & 0xFF),
        (uint8_t)(y_end >> 8),   (uint8_t)(y_end & 0xFF),
    };

    begin_write();
    write_command(0x2A); // CASET
    write_data(columns, sizeof(columns));

    write_command(0x2B); // RASET
    write_data(rows, sizeof(rows));

    write_command(0x2C); // RAMWR
    gpio_put(PIN_DC, 1);

    for (int16_t y = rect.y0; y < rect.y1; y++) {
        for (int16_t x = rect.x0; x < rect.x1; x++) {
            uint16_t color = framebuffer[(uint32_t)y * DISPLAY_WIDTH + (uint32_t)x];
            uint8_t bytes[2] = {
                (uint8_t)(color >> 8),
                (uint8_t)(color & 0xFF),
            };
            spi_write_blocking(spi0, bytes, sizeof(bytes));
        }
    }

    end_write();
}

void display_present(void)
{
    if (!initialized)
    {
        return;
    }
    if (dirty_rect_count == 0)
    {
        return;
        // assume this is because nothing is changed
    }

    for (int i = 0; i < dirty_rect_count; i++)
    {
        set_dirty_rectangle_and_draw(dirty_rects[i]);
    }
    //set it for the next frame
    dirty_rect_count = 0;
    

    // // CS is already low, DC is already high.
    // for (int i = 0; i < DISPLAY_WIDTH * DISPLAY_HEIGHT; i++)
    // {
    //     uint16_t color = framebuffer[i];
    //     uint8_t bytes[] = {
    //         (uint8_t)(color >> 8),
    //         (uint8_t)(color & 0xFF),
    //     };

    //     spi_write_blocking(spi0, bytes, 2);
    // }

    // end_write();
}
