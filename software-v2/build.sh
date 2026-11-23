#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PICO_SDK_PATH="${PICO_SDK_PATH:-"$HOME/src/pico-sdk"}"

if [[ ! -f "$PICO_SDK_PATH/external/pico_sdk_import.cmake" ]]; then
    echo "Pico SDK not found at: $PICO_SDK_PATH"
    echo "Set PICO_SDK_PATH or install it under ~/src/pico-sdk."
    exit 1
fi

export PICO_SDK_PATH

for tool in cmake ninja arm-none-eabi-gcc; do
    if ! command -v "$tool" >/dev/null; then
        echo "Missing tool: $tool"
        exit 1
    fi
done

build_target() {
    local name="$1"
    local board="$2"
    local platform="$3"
    local build_dir="$ROOT/build-$name"

    echo "== Building $name =="

    cmake -S "$ROOT" -B "$build_dir" -G Ninja \
        "-DPICO_SDK_PATH=$PICO_SDK_PATH" \
        "-DPICO_BOARD=$board" \
        "-DPICO_PLATFORM=$platform"

    cmake --build "$build_dir"
}

build_target rp2040 ferretboard rp2040
# build_target rp2350 pico2 rp2350

cd "$ROOT/build-rp2040"

picotool load -f -x ferret.uf2