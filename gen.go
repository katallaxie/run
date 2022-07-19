package main

//go:generate echo "Generating..."
//go:generate protoc -I. --go_out=paths=source_relative:. --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/plugin.proto
