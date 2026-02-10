module example

go 1.25.3

require (
	github.com/nats-io/nats.go v1.37.0
	github.com/toyz/protoc-gen-nats-micro v0.0.0-20251111043830-f26d09cffc8f
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace github.com/toyz/protoc-gen-nats-micro => ../../
