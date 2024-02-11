HERE ?= $(shell pwd)
LOCALBIN ?= $(shell pwd)/bin
VERSION    :=$(shell cat .version)
YAML_FILES :=$(shell find . ! -path "./vendor/*" -type f -regex ".*y*ml" -print)
REG_URI    ?=ghcr.io/converged-computing
REPO_NAME  :=$(shell basename $(PWD))

all: help

.PHONY: $(LOCALBIN)
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: protoc
protoc: $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	GOBIN=$(LOCALBIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	
.PHONY: proto
proto: protoc ## Generates the API code and documentation
	mkdir -p pkg/api/v1
	PATH=$(LOCALBIN):${PATH} protoc --proto_path=api/v1 --go_out=pkg/api/v1 --go_opt=paths=source_relative --go-grpc_out=pkg/api/v1  --go-grpc_opt=paths=source_relative api.proto

.PHONY: version
version: ## Prints the current version
	@echo $(VERSION)

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependancies 
	go mod tidy
	go mod vendor

.PHONY: upgrade
upgrade: ## Upgrades all dependancies 
	go get -d -u ./...
	go mod tidy
	go mod vendor

.PHONY: test
test: tidy ## Runs unit tests
	go test -count=1 -race -covermode=atomic -coverprofile=cover.out ./...

.PHONY: server
server: ## Runs uncompiled version of the server
	go run cmd/server/server.go

.PHONY: stream
stream: ## Runs the interface client
	go run cmd/stream/stream.go

.PHONY: register
register: ## Run mock registration
	go run cmd/register/register.go

.PHONY: tag
tag: ## Creates release tag 
	git tag -s -m "version bump to $(VERSION)" $(VERSION)
	git push origin $(VERSION)

.PHONY: tagless
tagless: ## Delete the current release tag 
	git tag -d $(VERSION)
	git push --delete origin $(VERSION)

.PHONY: clean
clean: ## Cleans bin and temp directories
	go clean
	rm -fr ./vendor
	rm -fr ./bin

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
