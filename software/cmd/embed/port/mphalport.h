// Ferretboard embed port: extra MP HAL declarations.

// Define so there's no dependency on extmod/virtpin.h
#define mp_hal_pin_obj_t

// pyexec.c calls mp_hal_set_interrupt_char (Ctrl-C while a script runs) but
// only gets the declaration from shared/runtime/interrupt_char.h; the embed
// port previously never compiled pyexec.c. Pull it in through the port header
// so every translation unit sees the prototype.
#include "shared/runtime/interrupt_char.h"