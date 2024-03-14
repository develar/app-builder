# go install -a -v github.com/go-bindata/go-bindata/...@latest (pack not used because cannot properly select dir to generate and no way to specify explicitly)

.PHONY: lint build publish assets

OS_ARCH = ""
ifeq ($(OS),Windows_NT)
	ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
		OS_ARCH := windows_amd64
	else ifeq ($(PROCESSOR_ARCHITEW6432),ARM64)
		OS_ARCH := windows_arm64
	else
		OS_ARCH := windows_386
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		ifeq ($(UNAME_M),riscv64)
			OS_ARCH := linux_riscv64
		else ifeq ($(UNAME_M),loongarch64)
			OS_ARCH := linux_loong64
		else
			OS_ARCH := linux_amd64
		endif
	endif
	ifeq ($(UNAME_S),Darwin)
		OS_ARCH := darwin_$(shell uname -m)
	endif
endif

# ln -sf ~/Documents/app-builder/dist/app-builder_darwin_amd64/app-builder ~/Documents/electron-builder/node_modules/app-builder-bin/mac/app-builder
# cp ~/Documents/app-builder/dist/app-builder_linux_amd64/app-builder ~/Documents/electron-builder/node_modules/app-builder-bin/linux/x64/app-builder
build: assets
	go build -ldflags='-s -w' -o dist/app-builder_$(OS_ARCH)/app-builder

build-all: assets
	./scripts/build.sh

# brew install golangci/tap/golangci-lint && brew upgrade golangci/tap/golangci-lint
lint:
	golangci-lint run

test:
	cd pkg/node-modules/pnpm-demo/ && pnpm install
	cd pkg/node-modules/npm-demo/ && npm install
	go test -v ./pkg/...

assets:
	~/go/bin/go-bindata -o ./pkg/package-format/bindata.go -pkg package_format -prefix ./pkg/package-format ./pkg/package-format/appimage/templates
	~/go/bin/go-bindata -o ./pkg/package-format/snap/snapScripts.go -pkg snap -prefix ./pkg/package-format/snap ./pkg/package-format/snap/desktop-scripts

publish:
	pnpm publish

update-deps:
	go get -u -d
	go mod tidy
