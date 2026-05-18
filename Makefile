.DEFAULT_GOAL := build
TARGETS ?= ./cmd/zrok2

.PHONY: clean build test

clean:
	rm -rf ui/node_modules ui/dist agent/agentUi/node_modules agent/agentUi/dist

build:
	npm --prefix ui install
	npm --prefix ui run build
	npm --prefix agent/agentUi install
	npm --prefix agent/agentUi run build
	go install $(TARGETS)

test:
	go test ./... -count=1
	go vet ./...
