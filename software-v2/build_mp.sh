#!/bin/bash
set -euo pipefail
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd ~/src/micropython/examples/embedding

git reset --hard HEAD
git pull

cp "${PROJECT_DIR}/src/mpconfigport.h" ./mpconfigport.h

rm -rf build-embed micropython_embed
make -f micropython_embed.mk