VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/deviantony/labctl/types.VERSION=$(VERSION)

.PHONY: build install clean

build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o dist/labctl cmd/labctl.go

install: build
	install -m 755 dist/labctl $$(go env GOPATH)/bin/labctl

clean:
	rm -rf dist/
