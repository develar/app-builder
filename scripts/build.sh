#!/usr/bin/env bash
set -ex

rm -rf win
rm -rf mac
rm -rf linux

mkdir mac
GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o mac/app-builder_amd64
GOOS=darwin GOARCH=arm64 go build -ldflags='-s -w' -o mac/app-builder_arm64
ln -s app-builder_amd64 mac/app-builder

mkdir -p linux/ia32
GOOS=linux GOARCH=386 go build -ldflags='-s -w' -o linux/ia32/app-builder

mkdir -p linux/x64
GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o linux/x64/app-builder

mkdir -p linux/riscv64
GOOS=linux GOARCH=riscv64 go build -ldflags='-s -w' -o linux/riscv64/app-builder

mkdir -p linux/arm
GOOS=linux GOARCH=arm go build -ldflags='-s -w' -o linux/arm/app-builder

mkdir -p linux/arm64
GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o linux/arm64/app-builder

mkdir -p linux/loong64
GOOS=linux GOARCH=loong64 go build -ldflags='-s -w' -o linux/loong64/app-builder

mkdir -p win/ia32
# $env:GOARCH='386'; go build -o win/ia32/app-builder.exe
GOOS=windows GOARCH=386 go build -o win/ia32/app-builder.exe

mkdir -p win/x64
# $env:GOARCH='amd64'; go build -o win/x64/app-builder.exe
GOOS=windows GOARCH=amd64 go build -o win/x64/app-builder.exe

mkdir -p win/arm64
# $env:GOARCH='arm64'; go build -o win/arm64/app-builder.exe
GOOS=windows GOARCH=arm64 go build -o win/arm64/app-builder.exe

find mac/ win/ linux/ -type f -exec chmod +x {} \;
