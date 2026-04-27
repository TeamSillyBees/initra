#!/usr/bin/env sh
set -eu

# Jet generator 会清空目标目录后重新生成，请不要把手写代码放进 internal/gen/jet。
JET_BIN="${JET_BIN:-jet}"
JET_DSN="${JET_DSN:-postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable}"
JET_SCHEMA="${JET_SCHEMA:-public}"
JET_PATH="${JET_PATH:-./internal/gen/jet}"

exec "$JET_BIN" -dsn="$JET_DSN" -schema="$JET_SCHEMA" -path="$JET_PATH"
