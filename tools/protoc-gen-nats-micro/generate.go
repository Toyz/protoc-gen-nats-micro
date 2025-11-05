// Package main provides go:generate directives for building the plugin
package main

//go:generate sh -c "cd ../.. && buf generate --template buf.gen.extensions.yaml --path proto/nats/options.proto"
//go:generate go build -o protoc-gen-nats-micro.exe .

// This file is just for go:generate directives
// The actual plugin entry point is in main.go
