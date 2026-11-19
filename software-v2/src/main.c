#include "pico/stdlib.h"
#include "display.h"
#include "input.h"

int main(void) {
    stdio_init_all();
    sleep_ms(500);
    display_init();
    

   
    input_init();
    display_fill(0xF800); // red

    

    while (true) {
        input_update();
        if (button_just_pressed(BTN_A)) {
            display_fill(0x001F);// blue
            display_present();
        }
        else if (buttton_just_pressed(BTN_DOWN))
        {
            display_fill(0xF800);
            display_present();
        }
        delay_ms(100);
    }
}
