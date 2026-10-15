#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"
mkdir -p build

build_archive() {
    local os="$1" target="$2" binary="$3" archive="$4"
    shift 4

    mkdir -p "$(dirname "$binary")"
    env CGO_ENABLED=0 GOOS="$os" GOARCH=amd64 go build -o "$binary" "./cmd/$target"
    rm -f "$archive"
    zip -j "$archive" "$binary" "$@"
}

for platform in macos linux windows; do
    os="$platform"
    suffix=""
    case "$platform" in
        macos) os=darwin ;;
        windows) suffix=.exe ;;
    esac

    build_archive "$os" server "build/dnslog-$platform$suffix" "build/dnslog-$platform.zip" config.yaml
    build_archive "$os" dnslog-cli "build/cli-$platform/dnslog-cli$suffix" "build/dnslog-cli-$platform.zip"
done
