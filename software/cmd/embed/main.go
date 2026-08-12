// Command embed hosts the MicroPython interpreter inside the Ferretboard Go
// firmware. Students write MicroPython; the display/input engine is Go.
//
// Unlike the other cmd targets this one is not a c-archive: Go's main() IS the
// entry point. It initialises the display singleton (init), bootstraps
// libmicropython (libmp_embed.a, compiled by the justfile), registers the
// `ferret` module, then hands control to the line-based REPL in host.c. The
// MP-facing functions live in bridge.go as //export funcs; host.c's trampolines
// map MP calls onto them.
//
// Reference build (host, no hardware): `just host`.
// Device build (RP2040): `just device`.

package main

/*
#include <stddef.h>
#include <stdint.h>

void register_ferret_module(void);
void run_repl(void);
void ferret_boot(void *gc_heap, size_t gc_heap_size);
*/
import "C"

import (
	"unsafe"

	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

// display is the singleton engine instance. Created once at startup (init)
// before any MicroPython code runs. All //export funcs act on it.
//
// The concrete type is provided by the build: device.NewDisplay() on a Pico,
// or a host stub (display_stub.go) when building on a PC for the reference.
var display platform.Screen

func init() {
	display = newDisplay() // defined per-build: device or stub
	_ = helpers.RandomInt  // ensure helpers is linked (used by games)
}

// mpHeap is the MicroPython GC heap. Enough for typical student programs;
// bump it if they grow larger (RAM is plentiful on the RP2040's 264KB).
var mpHeap [16 * 1024]byte

func main() {
	// Bootstrap libmicropython: point its GC at mpHeap and set the stack top
	// from a local so MP's stack-based GC scanning has a sane upper bound.
	C.ferret_boot(unsafe.Pointer(&mpHeap[0]), C.size_t(len(mpHeap)))
	C.run_repl()
}
