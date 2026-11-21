#include "pico/stdlib.h"
#include "display.h"
#include "input.h"
#include "mp.h"
#include "pico/stdio_usb.h"
#include <stdio.h>


int main(void) {
    int stack_top;
    stdio_init_all();
    absolute_time_t deadline = make_timeout_time_ms(3000);
    while (!stdio_usb_connected() && !time_reached(deadline)) {
        sleep_ms(10);
    }

    printf("Booting...\n");
    sleep_ms(100); //wait for the display to be ready, can probably lower this though
    display_init();
    input_init();
    display_fill(0x001F); // blue
    display_present();
    sleep_ms(500);
   
    script_init(&stack_top);
    display_present();
    
    // while (true) {
    //     input_update();
    //     if (button_just_pressed(BTN_A)) {
    //         draw_rect(40, 40, 140, 140, 0x07E0); // green
    //         display_present();
    //     }
    //     else if (button_just_pressed(BTN_DOWN))
    //     {
    //         draw_rect(100, 100, 200, 200, 0x001F); // blue
    //         display_present();
    //     }
    //     sleep_ms(10);
    // }
    // mp_deinit();
    while (true) { sleep_ms(100); script_update();}
    script_deinit();
    return 0;
}
