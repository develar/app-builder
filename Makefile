GOPATH = $(CURDIR)/vendor

.PHONY: lint build publish

# vetshadow is disabled because of err (https://groups.google.com/forum/#!topic/golang-nuts/ObtoxsN7AWg)
# goconst doesn't make sense
lint:
	gometalinter --aggregate --sort=path --vendor --skip=node_modules ./...

build:
	goreleaser --rm-dist --snapshot

# ln -sf ~/go/src/github.com/develar/app-builder/dist/darwin_amd64/app-builder ~/Documents/electron-builder/node_modules/app-builder-bin/mac/app-builder
build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o dist/darwin_amd64/app-builder

publish: build
	./scripts/publish-npm.sh