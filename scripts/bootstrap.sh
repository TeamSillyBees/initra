#!/usr/bin/env sh
set -eu

# 一键拉起本地依赖、执行迁移、生成 Jet 代码并启动服务。
make deps
make up
make migrate-apply ENV="${ENV:-local}"
make jet
make dev
