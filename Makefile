.DEFAULT_GOAL := build
GOBIN ?= $(shell go env GOPATH)/bin

ifeq ($(filter-out /,$(abspath $(GOBIN))),)
$(error GOBIN is '$(GOBIN)'; it must name a real directory)
endif

.PHONY: build test clean frontend frontend-test

build: frontend
	go install ./...

test: frontend-test
	go test ./... -count=1
	go vet ./...

frontend:
	npm --prefix ui install
	npm --prefix ui run build
	npm --prefix agent/agentUi install
	npm --prefix agent/agentUi run build

frontend-test: frontend
	npm --prefix ui run lint
	npm --prefix agent/agentUi run lint

clean:
	go clean ./...
	rm -f "$(GOBIN)"/*
	rm -rf ui/node_modules ui/dist agent/agentUi/node_modules agent/agentUi/dist
