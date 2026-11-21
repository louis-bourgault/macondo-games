#!/bin/bash
#we get micropython, make sure there have been no changes to the repo, then build it
set -euo pipefail
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd ~/src/micropython/examples/embedding
git reset --hard HEAD
git pull

cp "${PROJECT_DIR}/src/mpconfigport.h" ~/src/micropython/examples/embedding/micropython_embed/mpconfigport.h



rm -rf build-embed micropython_embed
make -f micropython_embed.mk