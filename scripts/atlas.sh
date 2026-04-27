#!/usr/bin/env sh
set -eu

# Atlas 配置文件固定放在 db/atlas.hcl。
ATLAS_BIN="${ATLAS_BIN:-atlas}"
CONFIG_PATH="${CONFIG_PATH:-file://db/atlas.hcl}"

exec "$ATLAS_BIN" -c "$CONFIG_PATH" "$@"
