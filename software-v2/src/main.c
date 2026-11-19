#include "pico/stdlib.h"
#include "display.h"
#include "input.h"

int main(void) {
    stdio_init_all();
    sleep_ms(50); //wait for the display to be ready, can probably lower this though
    display_init();
    
    input_init();
    display_fill(0xF800); // red
    display_present();

    

    while (true) {
        input_update();
        if (button_just_pressed(BTN_A)) {
            draw_rect(40, 40, 140, 140, 0x07E0); // green
            display_present();
        }
        else if (button_just_pressed(BTN_DOWN))
        {
            draw_rect(100, 100, 200, 200, 0x001F); // blue
            display_present();
        }
        sleep_ms(10);
    }


}
