# ln -fs ~/go/src/github.com/develar/go-pkcs12 ~/go/src/github.com/develar/app-builder/vendor/github.com/develar/go-pkcs12

.PHONY: lint build publish

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

publish: build
	./scripts/publish-npm.sh

update-deps:
	go get -u