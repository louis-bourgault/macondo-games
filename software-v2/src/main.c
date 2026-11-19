#include "pico/stdlib.h"
#include "display.h"
#include "input.h"

int main(void) {
    stdio_init_all();
    sleep_ms(1000);
    display_init();
    

   

    display_fill(0xF800); // red

    // for (int i = 0; i < 240; i++) {
    //     display_pixel(i, i, 0xFFFF); // white diagonal
    // }

    display_present();

    sleep_ms(1000);

    display_fill(0x07E0); // green
    display_present();

    while (true) {
        sleep_ms(1000);
    }
}
