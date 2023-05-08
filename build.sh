#!/bin/bash

go get -v -t -d ./...

mkdir -p build

# Build binary for macOS
env GOOS=darwin GOARCH=amd64 go build -o build/dnslog-macos

# Build binary for Linux
env GOOS=linux GOARCH=amd64 go build -o build/dnslog-linux

# Build binary for Windows
env GOOS=windows GOARCH=amd64 go build -o build/dnslog-windows.exe