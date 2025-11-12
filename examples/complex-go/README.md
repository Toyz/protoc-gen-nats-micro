# Complex Go NATS Micro Example

This example demonstrates advanced features of `protoc-gen-nats-micro` with Go, including interceptors, bidirectional headers, and error handling.

## Prerequisites

- Go 1.21 or higher
- NATS Server running on `localhost:4222`
- Buf CLI installed

## Setup

1. Generate the code:
```bash
# From the root of the repository
task generate:go
# or
buf generate --template examples/buf-configs/buf.gen.yaml
```

2. Install Go dependencies:
```bash
cd examples/complex-go
go mod download
```

## Running

### Start the Server

```bash
# From repository root
task run:server
# or
cd examples/complex-go
go run server.go
```

### Run the Client

In another terminal:

```bash
# From repository root
task run:client
# or
cd examples/complex-go
go run client.go
```

## Features Demonstrated

### Server Interceptors
- Logging interceptor that logs all requests
- Authentication interceptor that validates tokens
- Interceptor chaining with proper order

### Client Interceptors
- Request ID injection
- Retry logic
- Response time measurement

### Bidirectional Headers
- Client sends headers to server
- Server sends headers back to client
- Custom metadata in headers

### Type-Safe Error Handling
- Service-specific error types
- Error code constants
- Error checking helpers

### Service Configuration
- Custom subject prefixes
- Service metadata
- Endpoint-level configuration
- Timeout configuration

This example works with the same proto definitions as the TypeScript and Python examples, demonstrating complete feature parity across all three language implementations.
