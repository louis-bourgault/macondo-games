
#include "input.h"
#include <stdbool.h>

#include "hardware/gpio.h"
#include "pico/stdlib.h"



static ButtonMask current;
static ButtonMask previous;
typedef uint32_t ButtonMask;
static_assert(BTN_COUNT <= (sizeof(ButtonMask) * 8), "there are too many buttons to fit");

void input_init(void) {
    current = 0;
    previous = 0;

    for (uint8_t i = 0; i < BTN_COUNT; i++) {
        uint8_t pin = BUTTON_PINS[i];
        gpio_init(pin); 
        gpio_set_dir(pin, GPIO_IN);
        gpio_pull_up(pin);
    }
}

//pins for the v2
static const uint8_t BUTTON_PINS[BTN_COUNT] = {
    [BTN_A]     = 4,
    [BTN_B]     = 5,
    [BTN_X]     = 13,
    [BTN_Y]     = 14,
    [BTN_UP]    = 9,
    [BTN_DOWN]  = 11,
    [BTN_LEFT]  = 10,
    [BTN_RIGHT] = 12,
    [BTN_MENU]  = 6
};


void input_update(void) {
    previous = current;
    current = 0;

    //this makes it faster to read at oce
    uint32_t raw_pins = gpio_get_all();

    for (uint8_t i = 0; i < BTN_COUNT; i++) {
        if (!(raw_pins & (1U << BUTTON_PINS[i]))) {
            current |= (1U << i);
        }
    }
}
static inline bool button_pressed(Button button) {
    return (current & (1U << button)) != 0;
}

static inline bool button_just_pressed(Button btn) {
    ButtonMask mask = (1U << btn);
    return (current & mask) && !(previous & mask);
}

static inline bool button_just_released(Button btn) {
    ButtonMask mask = (1U << btn);
    return !(current & mask) && (previous & mask);
}