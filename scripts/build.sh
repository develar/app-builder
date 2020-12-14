#!/usr/bin/env bash
set -ex

cd app-builder-bin

rm -rf win
rm -rf mac
rm -rf linux

mkdir mac
GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o mac/app-builder ..

mkdir -p linux/ia32
GOOS=linux GOARCH=386 go build -ldflags='-s -w' -o linux/ia32/app-builder ..

mkdir -p linux/x64
GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o linux/x64/app-builder ..

mkdir -p linux/arm
GOOS=linux GOARCH=arm go build -ldflags='-s -w' -o linux/arm/app-builder ..

mkdir -p linux/arm64
GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o linux/arm64/app-builder ..

mkdir -p win/ia32
# set GOARCH=386
GOOS=windows GOARCH=386 go build -o win/ia32/app-builder.exe ..

mkdir -p win/x64
# set GOARCH=amd64
GOOS=windows GOARCH=amd64 go build -o win/x64/app-builder.exe ..