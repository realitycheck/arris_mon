# Arris Monitor Makefile
GO ?= go
VGO ?= $(GOPATH)/bin/vgo 
STATIK ?= $(GOPATH)/bin/statik

STATIK_DIR ?= $(shell pwd)/statik

build: generate vendor
	@echo ">> go build"
	@$(GO) build

test: generate vendor
	@echo ">> go test"	
	@$(GO) test -cover

install: build
	@echo ">> go install"
	@$(GO) install

clean: 
	@echo ">> go clean"	
	@$(GO) clean && rm -rf $(STATIK_DIR)
	
generate: $(STATIK)
	@echo ">> go generate statik"
	@$(STATIK) -f -src=./res -dest=.

vendor:	$(VGO)
	@echo ">> vgo vendor"
	@$(VGO) vendor

$(VGO): 
	@echo ">> go get vgo"
	@$(GO) get -u golang.org/x/vgo
	@rm -rf $(GOPATH)/src/golang.org/x/vgo

$(STATIK): 
	@echo ">> go get statik"
	@$(GO) get -u github.com/rakyll/statik
	@rm -rf $(GOPATH)/src/github.com/rakyll/statik

all: clean build test install

.PHONY: build install clean test generate all