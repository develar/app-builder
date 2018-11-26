# go get -u github.com/go-bindata/go-bindata/go-bindata (pack not used because cannot properly select dir to generate and no way to specify explicitly)

.PHONY: lint build publish assets

# brew install goreleaser
build:
	goreleaser --rm-dist --snapshot

# brew install golangci/tap/golangci-lint && brew upgrade golangci/tap/golangci-lint
lint:
	golangci-lint run

test:
	go test -v ./pkg/...

# ln -sf ~/Documents/app-builder/dist/darwin_amd64/app-builder ~/Documents/electron-builder/node_modules/app-builder-bin/mac/app-builder
build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o dist/darwin_amd64/app-builder

assets:
	go-bindata -o ./pkg/package-format/bindata.go -pkg package_format -prefix ./pkg/package-format ./pkg/package-format/appimage/templates

publish: build
	./scripts/publish-npm.sh

update-deps:
	go get -u
	go mod tidy