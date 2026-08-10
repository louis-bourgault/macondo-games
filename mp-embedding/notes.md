## What this scaffold is doing

This is the simplest possible MicroPython plus Go embedding path:

1. Go defines one exported function, `say_hello`.
2. `bridge/libbridge.a` is the C archive produced by Go.
3. `host/main.c` starts MicroPython, registers a Python-visible function called `say_hello`, and runs a short Python script.
4. When Python calls `say_hello('world')`, the C trampoline converts the Python string into the Go `GoString` ABI and calls the Go function.

The important thing to notice is that MicroPython never calls Go directly. It calls a C function first, and that C function is the bridge into Go.

## How to run it

Run these steps from `mp-embedding`.

### 1. Build the Go archive

This creates `bridge/libbridge.a` and `bridge/libbridge.h`.

```bash
cd /Users/louisb/Documents/hardware/macondo-games/mp-embedding/bridge
go build -buildmode=c-archive -o libbridge.a main.go
```

Why this matters: `-buildmode=c-archive` tells Go to package the exported Go function into a C-compatible static library.

### 2. Generate the embedded MicroPython package

This step prepares the self-contained MicroPython source tree under `micropython/examples/embedding/micropython_embed`.

```bash
cd /Users/louisb/Documents/hardware/macondo-games/mp-embedding/micropython/examples/embedding
make -f micropython_embed.mk MICROPYTHON_TOP=../..
```

Why this matters: the embed package contains the MicroPython runtime sources that the host program compiles and links against.

### 3. Compile the host C file

The include path must reach two different places:

1. `micropython/examples/embedding/micropython_embed`, which contains the generated MicroPython headers.
2. `micropython/examples/embedding`, which contains `mpconfigport.h`.

```bash
cd /Users/louisb/Documents/hardware/macondo-games/mp-embedding
ED=micropython/examples/embedding/micropython_embed
gcc -I. -I$ED -I$ED/port -I$ED/.. -I$ED/../../.. -c host/main.c -o host/main.o
```

Why `-I$ED/..` matters: `py/mpconfig.h` includes `mpconfigport.h`, and that file lives one directory above the generated embed package.

### 4. Compile the MicroPython sources

The embed package contains the runtime `.c` files that need object files before linking.

```bash
cd /Users/louisb/Documents/hardware/macondo-games/mp-embedding
ED=micropython/examples/embedding/micropython_embed
gcc -I. -I$ED -I$ED/port -I$ED/.. -I$ED/../../.. -c $(find $ED -name '*.c')
```

### 5. Link everything into one executable

```bash
cd /Users/louisb/Documents/hardware/macondo-games/mp-embedding
gcc -o spike host/main.o $(find micropython/examples/embedding/micropython_embed -name '*.o') bridge/libbridge.a -lm -lpthread
```

The `-lm -lpthread` flags are needed on the desktop host build.

### 6. Run the program

```bash
./spike
```

If everything is wired correctly, MicroPython runs `say_hello('world')`, which calls into C, which calls into Go, and Go prints the greeting.

## What the C file is doing

`host/main.c` is the actual bridge layer.

The important parts are:

1. `mp_embed_init(...)` starts the embedded MicroPython runtime.
2. `register_globals()` inserts a Python-visible `say_hello` into MicroPython’s main globals dictionary.
3. `mp_hello(...)` is the trampoline function. It accepts a MicroPython string object, converts it into a C string, then converts that into the Go `GoString` struct expected by `bridge/libbridge.h`.
4. `mp_embed_exec_str("say_hello('world')\n")` runs the test script.
5. `mp_embed_deinit()` shuts the runtime down.

That means the Python call goes through three layers:

Python `say_hello('world')` -> C trampoline `mp_hello(...)` -> Go export `say_hello(...)`

## Why this is simpler than the other agent’s version

The other version was trying to register a full MicroPython module. That is useful later, but it adds extra moving parts:

1. module object creation,
2. qstr registration,
3. module import behavior,
4. more MicroPython object plumbing.

For a first scaffold, a direct global function is cleaner because it proves the Go/C/Python boundary without forcing you to understand module registration yet.

## If it fails

If the host compile fails with a missing `mpconfigport.h`, the include path is wrong.

If the link fails with missing `say_hello`, the Go archive was not rebuilt or the link order is wrong.

If Python raises an exception, the runtime started, but the function registration or trampoline has a bug.

## Current build command that works here

```bash
ED=micropython/examples/embedding/micropython_embed
gcc -I. -I$ED -I$ED/port -I$ED/.. -I$ED/../../.. -c host/main.c -o host/main.o
```

This is the important compile step for the current scaffold.

## Moving MicroPython out of the project tree

You do not need to keep the MicroPython repository inside `mp-embedding`.

The Makefile now accepts `MICROPYTHON_ROOT`, so you can keep the clone anywhere and point the build at it:

```bash
make MICROPYTHON_ROOT=/Users/louisb/src/micropython
```

What that changes:

1. `MICROPYTHON_ROOT` becomes the source of the generated embed tree.
2. The build still writes `build/` and `spike` inside this project.
3. The Go bridge still lives in `bridge/` and is built the same way.

That is the simplest local setup if you want the MicroPython repo elsewhere.

## Doing it in GitHub Actions

The workflow can clone MicroPython into the runner’s temp directory instead of storing it in the repo checkout.

The basic pattern is:

1. Check out this repository.
2. Clone MicroPython into `$RUNNER_TEMP/micropython`.
3. Run `make MICROPYTHON_ROOT="$RUNNER_TEMP/micropython"`.
4. Run `./spike`.

That is the cleanest CI version because it keeps the project repository small while still building against a real MicroPython checkout.
