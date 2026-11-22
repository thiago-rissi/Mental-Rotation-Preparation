#!/usr/bin/env bash
# Copyright (2026) Thiago Lizier Rissi
#
# build.sh: build the Mental-Rotation-Preparation experiment.
#
#   ./build.sh          desktop executable for this machine:
#                       ./mental-rotation-preparation (.exe on Windows)
#   ./build.sh dist     release binaries in dist/ + SHA256SUMS:
#                       mental-rotation-preparation-linux        (Linux x86-64)
#                       mental-rotation-preparation-macos-arm64  (Apple silicon)
#                       mental-rotation-preparation-macos-amd64  (Intel Mac)
#                       mental-rotation-preparation-windows.exe  (Windows x86-64)
#   ./build.sh web      browser bundle in _site/experiment/ (see web/index.html)
#   ./build.sh web DIR  browser bundle in DIR instead
#   ./build.sh serve    browser bundle served on http://localhost:8080/
#
# Everything is embedded (SDL3, SDL3_ttf, the font): no C compiler, no SDL
# installation. Works on Linux, macOS and Windows (Git Bash).
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# This folder may sit inside a goxpyriment checkout whose go.work does not
# list it; build it as the standalone module it is.
export GOWORK=off
export CGO_ENABLED=0

NAME=mental-rotation-preparation
LDFLAGS="-s -w"
# The browser bundler ships inside the go-sdl3 fork pinned by the replace line
# in go.mod, so `go run` uses exactly that version.
WASMSDL="go run github.com/Zyko0/go-sdl3/cmd/wasmsdl"

case "${1:-}" in
"")
    go build -trimpath -ldflags="$LDFLAGS" -o "$NAME$(go env GOEXE)" .
    echo "built ./$NAME$(go env GOEXE)"
    ;;
dist)
    rm -rf dist && mkdir -p dist
    for target in linux/amd64:linux darwin/arm64:macos-arm64 darwin/amd64:macos-amd64 windows/amd64:windows.exe; do
        platform=${target%%:*}; suffix=${target#*:}
        echo "building $platform -> dist/$NAME-$suffix"
        GOOS=${platform%/*} GOARCH=${platform#*/} \
            go build -trimpath -ldflags="$LDFLAGS" -o "dist/$NAME-$suffix" .
    done
    (cd dist && (sha256sum "$NAME"-* 2>/dev/null || shasum -a 256 "$NAME"-*) > SHA256SUMS)
    ls -l dist
    ;;
web)
    out=${2:-_site/experiment}
    rm -rf "$out"
    $WASMSDL build -html web/index.html -out "$out" .
    ls -l "$out"
    ;;
serve)
    # Sends the COOP/COEP headers that give the page the ~5 us clock.
    $WASMSDL serve -html "$PWD/web/index.html" .
    ;;
*)
    echo "usage: $0 [dist|web [DIR]|serve]" >&2
    exit 2
    ;;
esac
