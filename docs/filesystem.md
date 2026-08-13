# The Ferretboard Filesystem

The Ferretboard runs a LittleFS2 volume at `/` on the last 8 MiB of its
16 MiB QSPI flash. Students see it as plain Python files:

```python
>>> f = open("hello.txt", "w")
>>> f.write("hello ferret")
>>> f.close()
>>> print(open("hello.txt").read())
```

This document explains how that works: the flash layout, the layers between
`open()` and the silicon, and the gotchas that had to be solved to get there.

## Architecture

The firmware is Go (TinyGo, RP2040/RP2350) hosting a MicroPython interpreter.
Go owns the hardware; MicroPython owns the student-facing runtime. The
filesystem spans that boundary:

```
Python  open() / os.listdir()
   │
   ▼
MicroPython VFS layer                 (extmod/vfs.c)
   ▼
VfsLfs2 driver                         (extmod/vfs_lfs.c — LittleFS2)
   │  readblocks / writeblocks / ioctl
   ▼
"Flash" block device object           (port/flash_storage.c)
   │  ferret_flash_read/write/erase/block_count/block_size  (//export)
   ▼
Go flashBlockDev                      (flash_common.go)
   │  ReadAt / WriteAt / EraseBlocks / Size
   ▼
raw flash backend
   ├─ device: machine.Flash           (flash_device.go — bootrom flash API)
   └─ host:   file-backed stub        (flash_stub.go  — NOR emulation)
```

## Flash layout

The chip is a 16 MiB W25Q128-class QSPI flash, memory-mapped at
`0x1000_0000` (XIP). The image is laid out as:

| Region               | XIP address        | Notes                                |
|----------------------|--------------------|--------------------------------------|
| boot2                | `0x1000_0100`      | 256 bytes, second-stage bootloader   |
| firmware (code+data) | `0x1000_0100` +    | ~330 KB                              |
| **filesystem**       | last 8 MiB of chip | 2048 blocks × 4 KiB                  |

The boundary between firmware and filesystem is computed at runtime, not
hard-coded:

- TinyGo's linker (`targets/arm.ld`) defines `__flash_data_start` (end of the
  program image) and `__flash_data_end` (end of the whole flash window) from
  `__flash_size`.
- `setupFlashRegion()` (`flash_device.go`) reads those symbols via
  `machine.FlashDataStart()/FlashDataEnd()` and places the FS region at
  `regionBase = rawTotalSize - 8 MiB`, i.e. ~7.6 MiB into the chip.

`__flash_size` comes from the target JSON. This is important:

```json
// ferretboard.json
"ldflags": ["--defsym=__flash_size=16M"]
```

`just device` builds with `-target ferretboard.json`. The stock `pico` target
uses `__flash_size=2048K`, which makes `regionBase` negative — the firmware
panics with "firmware larger than first 8 MiB of flash" before the REPL ever
appears (symptom: no serial output at all; the boot delay + stage traces we
added temporarily made it visible).

## Block device (Go side)

`flashBlockDev` (`flash_common.go`) presents the FS region as a zero-based
block device with 4 KiB blocks, hiding `regionBase` from its consumers. The
4 KiB size matches the flash erase sector, so LittleFS2 operations align with
hardware erases.

Key constants:

```go
flashBlockSize   = 4096   // erase sector == LFS2 block
flashStorageBytes = 8 MiB // reserved FS region
writePage        = 256    // NOR page-program granularity
```

`WriteAt` pads partial writes to whole 256-byte pages with `0xFF` before
calling the raw backend (`flash_common.go:76`). This is safe because NOR
programming only clears bits and LittleFS2 only writes into erased sectors,
so padding bytes are always `0xFF` no-ops.

### Device backend — `machine.Flash`

TinyGo's RP2040 flash driver (`machine_rp2_flash.go`):

- `ReadAt`: reads the memory-mapped XIP window directly (`unsafe.Slice` over
  `0x1000_0000 + off`).
- `WriteAt` / `EraseBlocks`: call the bootrom's `flash_range_program` /
  `flash_range_erase` tables (`machine_rp2040_rom.go`). The bootrom functions
  must run from RAM, which the driver handles; they also flush the XIP cache.

### Host backend — `stubbedFlash`

The PC reference build (`just host`) models the chip as a plain file
(`build/ferret.img`, overridable with `FERRET_FS_IMAGE`), initialised to
`0xFF` so erased reads behave like real flash. It is a debugging tool, so it
emulates NOR semantics faithfully:

- `WriteAt` does a read-modify-**AND**: programming a byte yields
  `existing & new`. A plain file overwrite would corrupt data, because
  LittleFS2 programmes a 4 KiB block in several 256-byte page-prog calls and
  the later ones would clobber the earlier ones. This exact bug corrupted the
  volume during development (formatted fine, then `open()` returned garbage
  or the mount failed after writes).
- `EraseBlocks` fills with `0xFF`.

## MicroPython side (C)

`port/flash_storage.c` wraps the `//export` funcs in an MP block device
object (`Flash` with `readblocks`/`writeblocks`/`ioctl`) and bootstraps the
volume:

```c
void ferret_fs_init(void) {
    if (!ferret_flash_has_fs()) {
        ferret_fs_format();       // first boot only
    }
    mp_obj_t bdev = ...&ferret_flash_obj;
    args[0] = mp_call_function_n_kw(&mp_type_vfs_lfs2, 1, 0, &bdev);
    args[1] = MP_OBJ_NEW_QSTR(qstr_from_str("/"));
    mp_vfs_mount(2, args, &mp_const_empty_map);
}
```

- **First-boot detection**: `ferret_flash_has_fs()` reads block 0 and checks
  for the `"littlefs"` superblock name at offset 8. A blank chip (all `0xFF`)
  has none, so it is formatted with `lfs2_format` first.
- **Mounting**: `VfsLfs2`'s `make_new` mounts eagerly, so instantiating
  `VfsLfs2(flash)` and passing it to `mp_vfs_mount("/", ...)` does the whole
  job. `mp_vfs_lfs2_make_new` is `static` in `vfs_lfs.c`, hence the
  roundabout `mp_call_function_n_kw` on the type object.
- **Soft reset**: `run_repl()` (host.c) re-runs `mp_embed_init` → registers
  the `ferret` module → `ferret_fs_init()` on every Ctrl-D, re-mounting `/`
  on the same volume.

The `ioctl` handler returns `BLOCK_COUNT`/`BLOCK_SIZE` from the Go exports
and `0` for INIT/DEINIT/SYNC. `readblocks`/`writeblocks` accept the optional
offset argument (block-level + sub-block offset) — both are summed correctly
in `ferret_flash_read/write` (`block*4096 + off`), no double counting.

## Boot flow (device)

```
bootrom → boot2 → Reset_Handler
  → runtime.init()          USB CDC enumerates here (before main!)
  → main()                  display init
  → setupFlashRegion()      FS region math (needs __flash_size=16M)
  → ferret_boot()           point MP's GC at mpHeap
  → run_repl()
      mp_embed_init
      register_ferret_module
      ferret_fs_init()      detect/format/mount "/"
      pyexec_friendly_repl  ">>> " banner
```

The REPL prompt only appears after the volume is mounted. If the mount
fails, the board looks dead even though USB is fine.

## Building and testing

```sh
# PC reference build — interactive MP REPL over a file-backed flash image
just host   # -> ./spike

# Device build — 16 MiB target, LittleFS2 region at ~7.6 MiB
just device # -> ferret.uf2  (flash with: picotool load ferret.uf2 -f)
```

`MICROPYTHON_ROOT=/path/to/micropython` must be set (the justfile's
`mp-arm`/`mp-host` recipes regenerate the embed package, genhdr and the
`libmp_embed*.a` archives; `gen_cgo.sh` bakes the archive fingerprints into
`z_cgo_link_*.go` — it re-runs after `ar rcs` so Go's build cache never
links a stale archive).

## Gotchas discovered along the way

1. **Wrong `__flash_size` (the "no serial at all" bug).** Building with the
   stock `pico` target (2 MiB) while the region math assumes 16 MiB made
   `regionBase` negative → immediate panic, before any USB output appeared.
   Fix: build with `ferretboard.json` (`__flash_size=16M`).
2. **TinyGo interp panic.** `machine.FlashDataStart()/End()` inside an
   `init()` function made the compile-time interpreter fold the
   `__flash_data_end` pointer constant and panic ("cannot convert pointer
   value to byte"). Fix: call them from `main()` via `setupFlashRegion()`
   — the interp only folds `*.init` functions.
3. **`mp_call_method_n_kw` hang.** Instantiating `VfsLfs2` must use
   `mp_call_function_n_kw(&mp_type_vfs_lfs2, 1, 0, &bdev)`, not
   `mp_call_method_n_kw` (which treated `&mp_type_vfs_lfs2` as a `self` and
   hung at boot).
4. **`ioctl` arity.** `t_flash_ioctl` must be `FUN_OBJ_3` (it takes
   self/cmd/arg); `FUN_OBJ_2` left the arg slot unset and `BLOCK_COUNT`
   returned garbage → mount failed.
5. **Probe offset.** The LFS2 superblock magic/name sits at offset 8 of
   block 0, not 16 — an off-by-one made every boot think the volume was
   unformatted and re-format it (data loss on every reboot).
6. **NOR write semantics (host).** The stub's plain-file overwrite did not
   emulate "program only clears bits", so padded page-progs clobbered
   earlier progs in the same 4 KiB block. Fixed with read-modify-AND.
7. **Fingerprint ordering.** `gen_cgo.sh` ran before the archives were
   rebuilt, so its md5 was one build behind; Go's cache then linked stale
   C objects. Fixed by re-running it after `ar rcs` in `mp-arm`/`mp-host`.