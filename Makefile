HERE ?= $(shell pwd)
LOCALBIN ?= $(shell pwd)/bin
VERSION    :=$(shell cat .version)
YAML_FILES :=$(shell find . ! -path "./vendor/*" -type f -regex ".*y*ml" -print)
REGISTRY  ?= ghcr.io/converged-computing
REPO_NAME  :=$(shell basename $(PWD))

all: help

.PHONY: $(LOCALBIN)
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: protoc
protoc: $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	GOBIN=$(LOCALBIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: build
build: build-cli build-rainbow

.PHONY: build-cli
build-cli: $(LOCALBIN)
	GO111MODULE="on" go build -o $(LOCALBIN)/rainbow cmd/rainbow/rainbow.go

.PHONY: build-rainbow
build-rainbow: $(LOCALBIN)
	GO111MODULE="on" go build -o $(LOCALBIN)/rainbow-scheduler cmd/server/server.go

.PHONY: docker
docker: docker-flux docker-ubuntu

.PHONY: docker-flux
docker-flux:
	docker build --build-arg base=fluxrm/flux-sched:jammy -t $(REGISTRY)/rainbow-flux:latest .

.PHONY: docker-ubuntu
docker-ubuntu: 
	docker build -t $(REGISTRY)/rainbow-scheduler:latest .

.PHONY: proto
proto: protoc ## Generates the API code and documentation
	mkdir -p pkg/api/v1
	PATH=$(LOCALBIN):${PATH} protoc --proto_path=api/v1 --go_out=pkg/api/v1 --go_opt=paths=source_relative --go-grpc_out=pkg/api/v1 --go-grpc_opt=paths=source_relative rainbow.proto

.PHONY: python
python: python ## Generate python proto files in python
	# pip freeze | grep grpcio-tools
	mkdir -p python/v1/rainbow/protos
	cd python/v1/rainbow/protos
	python -m grpc_tools.protoc -I./api/v1 --python_out=./python/v1/rainbow/protos --pyi_out=./python/v1/rainbow/protos --grpc_python_out=./python/v1/rainbow/protos ./api/v1/rainbow.proto
	# Not great, but gets the job done
	sed -i 's/import rainbow_pb2 as rainbow__pb2/from . import rainbow_pb2 as rainbow__pb2/' ./python/v1/rainbow/protos/rainbow_pb2_grpc.py

.PHONY: version
version: ## Prints the current version
	@echo $(VERSION)

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies
	go mod tidy
	go mod vendor

.PHONY: upgrade
upgrade: ## Upgrades all dependencies
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
	go run cmd/rainbow/rainbow.go register

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
	rm -rf ./rainbow.db
	go clean
	rm -fr ./vendor
	rm -fr ./bin

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
