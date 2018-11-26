#!/usr/bin/env bash
set -ex

mkdir -p app-builder-bin/mac
cp dist/darwin_amd64/app-builder app-builder-bin/mac/app-builder

mkdir -p app-builder-bin/linux
mkdir -p app-builder-bin/linux/ia32
mkdir -p app-builder-bin/linux/x64
mkdir -p app-builder-bin/linux/arm
mkdir -p app-builder-bin/linux/arm64
cp dist/linux_386/app-builder app-builder-bin/linux/ia32/app-builder
cp dist/linux_amd64/app-builder app-builder-bin/linux/x64/app-builder
cp dist/linux_arm_7/app-builder app-builder-bin/linux/arm/app-builder
cp dist/linux_arm64/app-builder app-builder-bin/linux/arm64/app-builder

mkdir -p app-builder-bin/win
mkdir -p app-builder-bin/win/ia32
mkdir -p app-builder-bin/win/x64
cp dist/windows_386/app-builder.exe app-builder-bin/win/ia32/app-builder.exe
cp dist/windows_amd64/app-builder.exe app-builder-bin/win/x64/app-builder.exe

ln -f readme.md app-builder-bin/readme.md

npm publish app-builder-bin