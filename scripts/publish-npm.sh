#!/usr/bin/env bash
set -ex

mkdir -p app-builder-bin/mac
cp dist/app-builder_darwin_amd64/app-builder app-builder-bin/mac/app-builder

mkdir -p app-builder-bin/linux
mkdir -p app-builder-bin/linux/ia32
mkdir -p app-builder-bin/linux/x64
mkdir -p app-builder-bin/linux/arm
mkdir -p app-builder-bin/linux/arm64
cp dist/app-builder_linux_386/app-builder app-builder-bin/linux/ia32/app-builder
cp dist/app-builder_linux_amd64/app-builder app-builder-bin/linux/x64/app-builder
cp dist/app-builder_linux_arm_7/app-builder app-builder-bin/linux/arm/app-builder
cp dist/app-builder_linux_arm64/app-builder app-builder-bin/linux/arm64/app-builder

mkdir -p app-builder-bin/win
mkdir -p app-builder-bin/win/ia32
mkdir -p app-builder-bin/win/x64
cp dist/app-builder_windows_386/app-builder.exe app-builder-bin/win/ia32/app-builder.exe
cp dist/app-builder_windows_amd64/app-builder.exe app-builder-bin/win/x64/app-builder.exe

ln -f readme.md app-builder-bin/readme.md

npm publish app-builder-bin