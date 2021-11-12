SHELL := /bin/sh

# The name of the executable (default is current directory name)
TARGET := slb
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := v0.1.0
BUILD := `git rev-parse HEAD`

# Operating System Default (LINUX)
TARGETOS=linux

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -extldflags -static"
DOCKERTAG ?= $(VERSION)
REPOSITORY = plndr

.PHONY: all build clean install uninstall fmt simplify check run e2e-tests

all: check install

$(TARGET):
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)

install:
	@echo Building and Installing project
	@go install $(LDFLAGS)

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w ./...

release: releaseARM releaseX86

releaseARM:
	@GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build $(LDFLAGS) -o $(TARGET)
	@tar -cvzf slb-armv7-$(VERSION).tar.gz ./slb
	@rm ./slb

releaseX86:
	@GOOS=linux CGO_ENABLED=0 go build $(LDFLAGS) -o $(TARGET)
	@tar -cvzf slb-amd64-$(VERSION).tar.gz ./slb
	@rm ./slb