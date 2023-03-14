# go-workbench Makefile template v1.0.1
# For a list of valid GOOS and GOARCH values, see: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
# Note: these can be overriden on the command line e.g. `make PLATFORM=<platform> ARCH=<arch>`
PLATFORM="$(shell go env GOOS)"
ARCH="$(shell go env GOARCH)"

.PHONY: pre build release image clean

dist := dist
bin := $(shell basename $(CURDIR))
image := deviantony/labctl

pre:
	mkdir -pv $(dist) 

build: pre
	GOOS=$(PLATFORM) GOARCH=$(ARCH) CGO_ENABLED=0 go build --installsuffix cgo --ldflags '-s' -o $(bin)
	mv $(bin) $(dist)/

release: pre
	GOOS=$(PLATFORM) GOARCH=$(ARCH) CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags '-s' -o $(bin)
	mv $(bin) $(dist)/

image: release
	docker buildx build --platform=$(PLATFORM)/$(ARCH) -t $(image) .

clean:
	rm -rf $(dist)/*