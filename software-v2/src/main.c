#include "pico/stdlib.h"
#include "display.h"
#include "input.h"
#include "mp.h"
#include "pico/stdio_usb.h"
#include <stdio.h>
#include "fs.h"
#include "lfs.h"
#include "pico/flash.h"
#include "hardware/flash.h"   // flash_range_erase
#include "hardware/sync.h" 

int main(void)
{
    int stack_top;
    stdio_init_all();
    absolute_time_t deadline = make_timeout_time_ms(3000);
    while (!stdio_usb_connected() && !time_reached(deadline))
    {
        sleep_ms(10);
    }

    printf("Booting...\n");

    printf("raw erase...\n");
uint32_t ints = save_and_disable_interrupts();
flash_range_erase(2 * 1024 * 1024, 4096);
restore_interrupts(ints);
printf("raw erase done\n");

    sleep_ms(100); // wait for the display to be ready, can probably lower this though
    display_init();
    input_init();
    display_fill(0x001F); // blue
    display_present();
    sleep_ms(500);
    printf("Initializing filesystem...\n");
    int fsiniterr = fs_init();
    if (fsiniterr != 0)
    {
        printf("prob with fs: %d\n", fsiniterr);
        return 1;
    }
    //test script to see if importing works
    static const char str[] =
        "from ferret import *\n"
        "def fill():\n"
        "    fill_display(0xF800)\n"
        "    \n";
    lfs_file_t file;
    int err;
    printf("writing test script to filesystem...\n");
    err = lfs_file_open(&lfs, &file, "/test.py", LFS_O_WRONLY | LFS_O_CREAT);
    
    if (err != 0)
    {
        printf("prob with file: %d\n", err);
        return 1;
    }
    lfs_ssize_t written = lfs_file_write(&lfs, &file, str, strlen(str));
    if (written < 0)
    {
        printf("prob writing to file: %d\n", written);
        return 1;
    }
    printf("wrote %d bytes to /test.py\n", (int)written);

    lfs_file_close(&lfs, &file);
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


    while (true)
    {
        sleep_ms(100);
        input_update();
        script_update();
        display_present();
    }
    script_deinit();
    return 0;
}
