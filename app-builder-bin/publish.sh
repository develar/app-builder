#!/usr/bin/env bash

cp dist/darwin_amd64/app-builder app-builder-bin/mac/app-builder

cp dist/linux_386/app-builder app-builder-bin/linux/ia32/app-builder
cp dist/linux_amd64/app-builder app-builder-bin/linux/x64/app-builder
cp dist/linux_arm_7/app-builder app-builder-bin/linux/arm/app-builder
cp dist/linux_arm64/app-builder app-builder-bin/linux/arm64/app-builder

cp dist/windows_386/app-builder.exe app-builder-bin/win/ia32/app-builder.exe
cp dist/windows_amd64/app-builder.exe app-builder-bin/win/x64/app-builder.exe

npm publish app-builder-bin/linux
npm publish app-builder-bin/mac
npm publish app-builder-bin/win
npm publish app-builder-bin