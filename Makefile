GOPATH = $(CURDIR)/vendor

.PHONY: router builder docker json

# vetshadow is disabled because of err (https://groups.google.com/forum/#!topic/golang-nuts/ObtoxsN7AWg)
# goconst doesn't make sense
lint:
	gometalinter --aggregate --sort=path --vendor --skip=node_modules ./...