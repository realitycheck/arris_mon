# Arris Monitor Makefile
GO ?= go
VGO ?= $(GOPATH)/bin/vgo 
STATIK ?= $(GOPATH)/bin/statik
WGET ?= wget
GIT ?= git

STATIK_DIR ?= $(shell pwd)/statik

VENDOR_JS_DIR ?= $(shell pwd)/res/static/vendor
SMOOTHIE_JS ?= $(VENDOR_JS_DIR)/smoothie.js
SMOOTHIE_JS_URL ?= http://smoothiecharts.org/smoothie.js

.SHELLFLAGS = -c # Run commands in a -c flag 
.PHONY: build install clean test generate all help
.DEFAULT: help

help: ## Help	
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'	

build: generate vendor ## Build
	@echo ">> go build"
	@$(GO) build

test: generate vendor ## Test
	@echo ">> go test"	
	@$(GO) test -cover

install: build ## Install
	@echo ">> go install"
	@$(GO) install

clean: ## Clean
	@echo ">> go & git clean"	
	@$(GO) clean && $(GIT) clean -qdxf
	
generate: $(STATIK) $(SMOOTHIE_JS) ## Generate
	@echo ">> go generate statik"
	@$(STATIK) -f -src=./res -dest=.

vendor:	$(VGO) ## Vendor
	@echo ">> vgo vendor"
	@$(VGO) vendor

all: clean build test install

$(SMOOTHIE_JS):
	@echo ">> wget smoothie.js"
	@mkdir -p $(VENDOR_JS_DIR)
	@cd $(VENDOR_JS_DIR) && $(WGET) -q $(SMOOTHIE_JS_URL)

$(VGO): 
	@echo ">> go get vgo"
	@$(GO) get -u golang.org/x/vgo
	@rm -rf $(GOPATH)/src/golang.org/x/vgo

$(STATIK): 
	@echo ">> go get statik"
	@$(GO) get -u github.com/rakyll/statik
	@rm -rf $(GOPATH)/src/github.com/rakyll/statik
