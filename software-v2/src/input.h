#ifndef INPUT_H
#define INPUT_H

#include <stdbool.h>
typedef enum {
    BTN_A = 0,
    BTN_B,
    BTN_X,
    BTN_Y,
    BTN_UP,
    BTN_DOWN,
    BTN_LEFT,
    BTN_RIGHT,
    BTN_MENU,
    BTN_COUNT
} Button;
//having button_count at the end here lets us use it to define size of the array

void input_init(void);
void input_update(void);
bool button_pressed(Button button);
bool button_just_pressed(Button button);
bool button_just_released(Button button);

#endif