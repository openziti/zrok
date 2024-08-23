#!/bin/sh

protoc --go_out=agent/grpc \
	--go_opt=paths=source_relative \
	--go-grpc_out=agent/grpc \
	--go-grpc_opt=paths=source_relative \
	agent/grpc/agent.proto
