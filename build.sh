#!/bin/bash

go get -v -t -d ./...

mkdir -p build

# Build binary for macOS
env GOOS=darwin GOARCH=amd64 go build -o build/dnslog-macos
zip -j build/dnslog-macos.zip build/dnslog-macos config.yaml


# Build binary for Linux
env GOOS=linux GOARCH=amd64 go build -o build/dnslog-linux
zip -j build/dnslog-linux.zip build/dnslog-linux config.yaml

# Build binary for Windows
env GOOS=windows GOARCH=amd64 go build -o build/dnslog-windows.exe
zip -j build/dnslog-windows.zip build/dnslog-windows.exe config.yaml