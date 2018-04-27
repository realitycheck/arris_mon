# Arris Monitor Makefile
GO ?= go
VGO ?= $(GOPATH)/bin/vgo 
STATIK ?= $(GOPATH)/bin/statik

STATIK_DIR ?= $(shell pwd)/statik

.SHELLFLAGS = -c # Run commands in a -c flag 
.PHONY: build install clean test generate all help
.DEFAULT: help

help: ## Help	
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'	

build: generate vendor  ## Build
	@echo ">> go build"
	@$(GO) build

test: generate vendor ## Test
	@echo ">> go test"	
	@$(GO) test -cover

install: build ## Install
	@echo ">> go install"
	@$(GO) install

clean: ## Clean
	@echo ">> go clean"	
	@$(GO) clean && rm -rf $(STATIK_DIR)
	
generate: $(STATIK)  ## Generate
	@echo ">> go generate statik"
	@$(STATIK) -f -src=./res -dest=.

vendor:	$(VGO)  ## Vendor
	@echo ">> vgo vendor"
	@$(VGO) vendor

all: clean build test install

$(VGO): 
	@echo ">> go get vgo"
	@$(GO) get -u golang.org/x/vgo
	@rm -rf $(GOPATH)/src/golang.org/x/vgo

$(STATIK): 
	@echo ">> go get statik"
	@$(GO) get -u github.com/rakyll/statik
	@rm -rf $(GOPATH)/src/github.com/rakyll/statik
