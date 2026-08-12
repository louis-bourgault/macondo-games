# How the Ferretboard build system works

This document explains how MicroPython is embedded inside the Ferretboard Go
firmware and how the `cmd/embed` build turns that into two artefacts:

| Target | Command                     | Output                | Purpose                                   |
| ------ | --------------------------- | --------------------- | ----------------------------------------- |
| host   | `just host`                 | `spike`               | PC reference: full REPL over stdin/stdout |
| device | `just device`               | `ferret.uf2`          | RP2040/RP2350 firmware (flash with picotool) |

Both share the same sources and differ only in the compiler/board support
files. The authoritative entry point is `justfile`; read it top to bottom with
this document.

---

## 1. The core idea

MicroPython's C interpreter is precompiled into a **static archive**
(`build/libmp_embed.a` on device, `build/libmp_embed_host.a` on the host). Go's
`main()` is the real entry point. The two sides meet through **cgo**:

* MicroPython C code that must reach the outside world calls Go functions that
  are `//export`ed (USB CDC, display, random, …).
* Go calls into MicroPython through `host.c`, which is compiled by cgo as part
  of the Go build.

```
+-------------------------------------------------------------+
| Go runtime (main.go, bridge.go, cdc_*.go, display_*.go)     |
|   main() -> ferret_boot() -> run_repl()                     |
+----------------------------|--------------------------------+
                             | cgo  (host.c: C glue)          |
+----------------------------v--------------------------------+
| libmicropython  (libmp_embed*.a, precompiled C)             |
|   py/*  shared/runtime/pyexec.c  shared/readline  extmod/*  |
+-------------------------------------------------------------+
```

---

## 2. Layout

```
cmd/embed/
  justfile             recipes: embed, mp-arm, mp-host, host, device, clean
  gen_cgo.sh           writes z_cgo_link_{host,tinygo}.go (include/link flags)
  host.c               C glue: ferret module trampolines, run_repl(), soft reset
  bridge.go            //export funcs the `ferret` module calls (thin, engine-facing)
  cdc_stub.go          host CDC: reads stdin / writes stdout            [go:build !tinygo]
  cdc_tinygo.go        device CDC: TinyUSB ring buffer + pump goroutine  [go:build tinygo]
  display_stub.go      host display singleton (no-op)                   [go:build !tinygo]
  display_device.go    device display singleton (ST7789)                [go:build tinygo]
  main.go              Go entry point, MP GC heap, display singleton
  ferret_abi.h         hand-written C prototypes for the Go //export funcs
  port/
    mpconfigport.h     our MicroPython feature configuration
    mphalport.c        MP stdio/ticks hooks -> ferret_cdc_* (Go)
    mphalport.h        MP HAL extra declarations (interrupt_char prototype)
    repl_stubs.c       mp_lexer_new_from_file stub (no POSIX FS on device)
    embed-genhdr.mk    regenerates qstrs / root pointers / module defs
  z_cgo_link_*.go      generated (see §5)
  build/               per-target .o files + the static archives (gitignored)
```

`MICROPYTHON_ROOT` must point at a full MicroPython **checkout** (not a vendored
copy): the build generates an embed package inside it and runs MicroPython's own
make rules. Set it as
`just --set MICROPYTHON_ROOT /path/to/micropython <recipe>`
(it defaults to `~/src/micropython`).

---

## 3. Stage 1: the embed package (recipe `embed`)

MicroPython ships a generator at
`examples/embedding/micropython_embed.mk` that emits a self-contained package
into `$MICROPYTHON_ROOT/examples/embedding/micropython_embed/`:

* `py/*` — the whole core interpreter, plus `genhdr/` (generated headers),
* `port/` — the stock embed port (`mpconfigport_common.h`, `embed_util.c`, …).

The `embed` recipe then **overlays our files** onto that package:

1. Our `port/mpconfigport.h`, `mphalport.c`, `mphalport.h`, `repl_stubs.c`
   replace the stock ones. `mpconfigport.h` is the master switch for every MP
   feature (compiler, GC, `sys`, `os`, VFS, readline, REPL helpers, banner).
2. Extra upstream sources are copied in because the stock generator ships only a
   minimal subset:
   * `shared/runtime/pyexec.{c,h}` — the REPL (friendly / raw / paste modes),
   * `shared/runtime/interrupt_char.{c,h}` — `mp_hal_set_interrupt_char` (Ctrl-C),
   * `shared/readline/` — line editing used by the friendly REPL,
   * `extmod/vfs.c`, `vfs_blockdev.c`, `modos.c`, `misc.h` — the `os` module
     and the VFS layer (file ops; a filesystem mount is the next step).
3. **`embed-genhdr.mk` regenerates the package's `genhdr/`.** This is the
   subtle bit — see §4.
4. `gen_cgo.sh` writes the cgo include/link flags (see §5).

---

## 4. Why `genhdr/` must be regenerated

MicroPython scans its **compiled sources** to build three generated headers:

* `qstrdefs.generated.h` — every `MP_QSTR_*` used in C,
* `root_pointers.h` — every `MP_REGISTER_ROOT_POINTER` (state that the GC must
  trace, e.g. `vfs_mount_table`, readline history),
* `moduledefs.h` — every `MP_REGISTER_MODULE` (e.g. the `os` module).

The stock embed generator only scans `py/` sources, so the extra files copied in
step 2 would contribute nothing: you would get link errors like
`undefined reference to MP_QSTR_mount`, a missing `vfs_mount_table` root pointer,
and `import os` would fail. `embed-genhdr.mk` therefore re-runs MicroPython's
own generation machinery *inside the package directory*, adding our sources to
`SRC_QSTR`:

```make
include $(MICROPYTHON_TOP)/py/py.mk
SRC_QSTR += $(wildcard extmod/*.c) $(wildcard shared/runtime/*.c) shared/readline/readline.c
include $(MICROPYTHON_TOP)/py/mkrules.mk
```

Two gotchas, both hit during bring-up:

* **`SRC_QSTR +=` must precede `include mkrules.mk`.** GNU make expands rule
  prerequisites immediately, so the `qstr.i.last` rule captures `SRC_QSTR` at
  parse time — adding files after the include has no effect.
* **The make must run with the package directory as CWD** (the justfile does
  `cd $EMBED_DIR`), because `$(wildcard extmod/*.c)` and the `-I.` include paths
  are relative. Also add `vpath mpconfigport.h port` so the qstr dependency
  scan can find our config file.

The regenerated headers land in `build-local/genhdr/` and are copied back over
the package's `genhdr/`.

---

## 5. Stage 2: compiling the archive (recipes `mp-arm` / `mp-host`)

`mp-arm` and `mp-host` are the same recipe with a different compiler:

```sh
find $EMBED_DIR -name '*.c' | while read src; do
    <gcc | arm-none-eabi-gcc> -I$ROOT -I$EMBED_DIR ... -c "$src"
done
ar rcs build/libmp_embed*.a build/{arm,host}/*.o
```

* `-I$ROOT` lets sources find `ferret_abi.h`.
* The device uses `-mcpu=cortex-m0plus` (RP2040) or `-mcpu=cortex-m33`
  (RP2350), chosen by `TARGET` (`pico` / `pico2`).
* Sources are compiled with `-fno-common`; the archive is then handed to the Go
  link step through cgo.

---

## 6. Stage 3: cgo glue (`gen_cgo.sh`, `host.c`, `z_cgo_link_*.go`)

`host.c` is compiled by cgo (it is a normal file in the Go package) and needs
the MicroPython include paths. Those absolute paths depend on
`MICROPYTHON_ROOT`, so `gen_cgo.sh` bakes them into two generated files:

* `z_cgo_link_host.go` (`//go:build !tinygo`) —
  `#cgo LDFLAGS: -Lbuild -lmp_embed_host -lm`,
* `z_cgo_link_tinygo.go` (`//go:build tinygo`) —
  `#cgo LDFLAGS: -Lbuild -lmp_embed`.

These are generated (`DO NOT EDIT`) and gitignored.

`host.c` is the bridge in both directions:

* **Go -> MicroPython:** `ferret_boot()` captures the GC heap; `run_repl()`
  boots MP and runs its REPL forever. `register_ferret_module()` builds the
  `ferret` module at runtime (see §8).
* **MicroPython -> Go:** each `ferret.*` function is a small C trampoline
  (`t_fill`, `t_fill_rect`, …) that unpacks MP arguments and calls the matching
  `//export` Go function (`ferret_fill`, … in `bridge.go`).

The C prototypes for those Go functions are hand-written in `ferret_abi.h`
rather than pulled from cgo's generated header, so `mphalport.c` (compiled into
the archive *outside* cgo) can include them too. Keep the signatures in sync
with `bridge.go` and `cdc_*.go`.

---

## 7. Stage 4: the final link

* **Host:** `go build -o spike .` — the Go linker calls cgo, which links
  `host.c` + `libmp_embed_host.a`. `display_stub.go` and `cdc_stub.go` (build
  tag `!tinygo`) stand in for hardware, so the reference runs anywhere.
* **Device:** `tinygo build -target pico -o ferret.uf2 .` — TinyGo compiles the
  Go engine, links the same archive, and packages a UF2. `display_device.go`
  and `cdc_tinygo.go` (build tag `tinygo`) provide the real display and USB.

`tinygo` must be on `PATH` (e.g. `/tmp/opencode/tinygo/bin`); the justfile does
not bundle it.

---

## 8. The `ferret` module and the qstr boundary

MicroPython interns every identifier at build time into an enum
(`MP_QSTR_*`). Our Go engine is **not** scanned by that generator, so
`MP_QSTR_ferret` does not exist in C. Instead the module is assembled at
runtime:

```c
void register_ferret_module(void) {
    mp_obj_t mod = mp_obj_new_module(qstr_from_str("ferret"));
    mp_obj_dict_store(mp_obj_module_get_globals(mod),
        MP_OBJ_NEW_QSTR(qstr_from_str("fill")), MP_OBJ_FROM_PTR(&t_fill_obj));
    ...
}
```

`qstr_from_str` interns on the fly, so any string works. A soft reset (Ctrl-D)
clears MP's module registry, so `mpconfigport.h` also defines
`MICROPY_BOARD_BEFORE_PYTHON_EXEC(...) register_ferret_module()`, which pyexec
calls before *every* paste/exec. Re-registration is idempotent.

---

## 9. Runtime flow

1. `main()` creates the display singleton, then calls `ferret_boot(heap, size)`
   (stores the GC heap for soft-reset re-init) and `run_repl()`.
2. `run_repl()` anchors MP's stack scanner at its own frame
   (`mp_stack_set_top`), calls `mp_embed_init` (`gc_init` + `mp_init`) and
   `register_ferret_module()`, then loops:
   * `pyexec_friendly_repl()` — `>>> ` line editing, paste mode (Ctrl-E),
     raw mode (Ctrl-A), soft reset (Ctrl-D on an empty line),
   * `pyexec_raw_repl()` — the `OK`/`\x04\x04` framing the editor's raw-REPL
     tools use.
3. **stdio routing:** MP's `mp_hal_stdin_rx_chr` / `mp_hal_stdout_tx_strn`
   (in `mphalport.c`) call `ferret_cdc_read` / `ferret_cdc_write`.
   * Device: `cdc_tinygo.go` — `pumpCDC()` fills a ring buffer from TinyUSB;
     `ferret_cdc_read()` blocks (spin + `runtime.Gosched()`) because pyexec
     expects blocking reads; writes go straight to `machine.Serial`.
   * Host: `cdc_stub.go` bridges to `os.Stdin` / `os.Stdout`.
4. **soft reset (Ctrl-D):** pyexec returns `PYEXEC_FORCED_EXIT`; `run_repl`
   runs `mp_deinit()` → `gc_sweep_all()` → `mp_embed_init(...)` (fresh heap +
   VM state) and continues the loop. `ferret` is re-registered on the next exec
   via the board hook.

---

## 10. Pitfalls discovered during bring-up

1. **Stack top must be a C-frame address.** On the host, cgo runs C on a
   dedicated OS thread; if `mp_stack_set_top` used a Go goroutine address, GC's
   stack scan walked gigabytes of unmapped memory and segfaulted. Anchoring it
   at `run_repl()`'s own frame is correct because all MP activity stays within
   that one cgo call.
2. **`mp_init()` must happen before `register_ferret_module()`** — registering
   against an uninitialised VM hangs. Both now live inside `run_repl()`.
3. **No POSIX on bare metal.** `MICROPY_READER_POSIX` pulls in `open/read/close/
   errno`, which TinyGo's libc lacks. `pyexec.c` references
   `mp_lexer_new_from_file` for the never-used "run file by path" path, so
   `repl_stubs.c` provides a stub raising `OSError(ENOENT)` instead.
4. **`mp_hal_ticks_ms` / `mp_hal_stdout_tx_str` are port responsibilities.**
   pyexec needs them; `mphalport.c` implements both (hardware timer on device,
   `clock_gettime` on host).
5. **genhdr scan coverage** — see §4. If a new module/source adds qstrs or root
   pointers and isn't in `SRC_QSTR`, the *link* fails with `undefined reference`
   to `MP_QSTR_*`, not the compile.
6. **`MP_STATE_PORT`** must alias `MP_STATE_VM` or readline history doesn't
   compile.
7. **tinygo is not bundled** — `just device` needs it on `PATH`.

---

## 11. Common tasks

**Add a `ferret` SDK function.** 1) add a Go method to
`internal/platform.Screen` if needed; 2) implement the `//export` func in
`bridge.go`; 3) declare it in `ferret_abi.h`; 4) add a trampoline +
`MP_DEFINE_CONST_FUN_OBJ_*` in `host.c` and store it in
`register_ferret_module()`.

**Enable an MP feature/module.** Add the `MICROPY_*` define to
`port/mpconfigport.h`. If it compiles extra sources (e.g. `extmod/foo.c`), copy
them into the package in the `embed` recipe **and** make sure they reach
`SRC_QSTR` in `embed-genhdr.mk`, then rebuild `embed` so `genhdr/` regenerates.

**Upgrade the MicroPython checkout.** Point `MICROPYTHON_ROOT` at the new
checkout and run `just embed` — the package and generated files are rebuilt
from scratch. Re-check §4/§10 items if upstream moved a generated-header rule or
changed a HAL prototype.

**Clean everything:** `just clean` removes `build/`, `spike`, `ferret.uf2` and
the generated cgo link files.
