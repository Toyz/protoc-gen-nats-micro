# protoc-gen-nats-micro

A Protocol Buffers compiler plugin that generates type-safe NATS microservice code. Define services in protobuf, get production-ready NATS microservices with automatic service discovery, load balancing, and zero configuration.

## Overview

`protoc-gen-nats-micro` is a code generation tool that brings the developer experience of gRPC to NATS.io. Write standard `.proto` files and generate:

- **NATS microservices** using the official `nats.io/micro` framework
- **gRPC services** for comparison and compatibility  
- **REST gateways** via grpc-gateway
- **OpenAPI specifications** for documentation

All from a single protobuf definition.

## Motivation

**Why this exists:**

The NATS ecosystem lacked a modern, maintained code generation solution comparable to gRPC. Existing tools like nRPC were abandoned and didn't integrate with modern protobuf toolchains. This project fills that gap.

**Key advantages over gRPC for internal microservices:**

- No service mesh complexity (Istio, Linkerd)
- Built-in service discovery and load balancing
- Simpler operations and deployment
- Native pub/sub patterns when needed
- Lower latency for small messages

**When to use NATS vs gRPC:**

- **NATS**: Internal microservices, event-driven systems, high-throughput messaging
- **gRPC**: External APIs, language-heterogeneous systems, HTTP/2 requirements
- **REST**: Public APIs, browser clients, third-party integrations

This plugin lets you generate all three from the same proto files and choose based on your needs.

## Features

- **Zero configuration** - Service metadata lives in proto files
- **Type-safe generated code** - Compile-time safety for requests/responses
- **Multi-language ready** - Template system supports Go (implemented), Rust, TypeScript (planned)
- **Standard tooling** - Works with `buf`, `protoc`, and existing protobuf workflows
- **Service discovery** - Automatic via NATS, no Consul/etcd needed
- **Load balancing** - Built into NATS queue groups
- **API versioning** - Subject prefix isolation per version
- **Side-by-side comparison** - Generate NATS + gRPC + REST to evaluate approaches

## Quick Start

### Prerequisites

- Go 1.21 or later
- [Buf](https://buf.build/docs/installation) v2
- [Task](https://taskfile.dev) (optional, for convenience)
- NATS server (Docker or local)

### Installation

```bash
go install github.com/toyz/protoc-gen-nats-micro/cmd/protoc-gen-nats-micro@latest
```

### Generate Code

```bash
# Using Task
task generate

# Or manually with buf
buf generate
```

### Run Example

```bash
# Terminal 1: Start NATS
docker run -p 4222:4222 nats

# Terminal 2: Start services
go run ./examples/complex-server

# Terminal 3: Run client
go run ./examples/complex-client
```

## Usage

### 1. Define Service in Protobuf

```protobuf
syntax = "proto3";

package order.v1;

import "nats/options.proto";
import "google/api/annotations.proto";

service OrderService {
  option (nats.micro.service) = {
    subject_prefix: "api.v1"
    name: "order_service"
    version: "1.0.0"
    description: "Order management service"
  };

  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse) {
    option (google.api.http) = {
      post: "/v1/orders"
      body: "*"
    };
  }

  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse) {
    option (google.api.http) = {
      get: "/v1/orders/{id}"
    };
  }
}

message CreateOrderRequest {
  string customer_id = 1;
  repeated OrderItem items = 2;
}

message CreateOrderResponse {
  Order order = 1;
}

// ... additional messages
```

### 2. Implement Service Interface

```go
package main

import (
    "context"
    orderv1 "yourmodule/gen/order/v1"
)

type orderService struct {
    orders map[string]*orderv1.Order
}

func (s *orderService) CreateOrder(
    ctx context.Context,
    req *orderv1.CreateOrderRequest,
) (*orderv1.CreateOrderResponse, error) {
    order := &orderv1.Order{
        Id:         generateID(),
        CustomerId: req.CustomerId,
        Items:      req.Items,
        Status:     orderv1.OrderStatus_PENDING,
    }
    s.orders[order.Id] = order
    return &orderv1.CreateOrderResponse{Order: order}, nil
}

func (s *orderService) GetOrder(
    ctx context.Context,
    req *orderv1.GetOrderRequest,
) (*orderv1.GetOrderResponse, error) {
    order, exists := s.orders[req.Id]
    if !exists {
        return nil, errors.New("order not found")
    }
    return &orderv1.GetOrderResponse{Order: order}, nil
}
```

### 3. Register with NATS

```go
package main

import (
    "github.com/nats-io/nats.go"
    orderv1 "yourmodule/gen/order/v1"
)

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    svc := &orderService{
        orders: make(map[string]*orderv1.Order),
    }

    // Subject prefix read from proto automatically
    _, err = orderv1.RegisterOrderService(nc, svc)
    if err != nil {
        log.Fatal(err)
    }

    // Service is now discoverable at "api.v1.order_service"
    select {} // Keep running
}
```

### 4. Use Generated Client

```go
package main

import (
    "context"
    "github.com/nats-io/nats.go"
    orderv1 "yourmodule/gen/order/v1"
)

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    client := orderv1.NewOrderServiceNatsClient(nc)

    resp, err := client.CreateOrder(context.Background(),
        &orderv1.CreateOrderRequest{
            CustomerId: "user-123",
            Items: []*orderv1.OrderItem{
                {ProductId: "prod-456", Quantity: 2},
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created order: %s\n", resp.Order.Id)
}
```

## Generated Code

From a single `.proto` file, the plugin generates:

```
gen/order/v1/
├── service.pb.go           # Standard protobuf messages
├── service_nats.pb.go      # NATS service and client (this plugin)
├── service_grpc.pb.go      # gRPC service and client (protoc-gen-go-grpc)
├── service.pb.gw.go        # REST gateway handlers (grpc-gateway)
└── service.swagger.yaml    # OpenAPI specification
```

### NATS Service Interface

```go
type OrderServiceNats interface {
    CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
    GetOrder(context.Context, *GetOrderRequest) (*GetOrderResponse, error)
}

func RegisterOrderService(nc *nats.Conn, impl OrderServiceNats, opts ...RegisterOption) (micro.Service, error)
```

### NATS Client

```go
type OrderServiceNatsClient struct { /* ... */ }

func NewOrderServiceNatsClient(nc *nats.Conn, opts ...NatsClientOption) *OrderServiceNatsClient

func (c *OrderServiceNatsClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
```

## Configuration

Service configuration is defined in proto files using custom options:

```protobuf
import "nats/options.proto";

service OrderService {
  option (nats.micro.service) = {
    subject_prefix: "api.v1"        // NATS subject namespace
    name: "order_service"           // Service name for discovery
    version: "1.0.0"                // Semantic version
    description: "Order management" // Human-readable description
  };
}
```

These values are read at code generation time and embedded in the generated code. No runtime configuration files needed.

### Runtime Overrides (Optional)

While configuration lives in proto files, you can override at runtime:

```go
orderv1.RegisterOrderService(nc, svc,
    orderv1.WithSubjectPrefix("custom.prefix"),
    orderv1.WithVersion("2.0.0"),
)
```

## API Versioning

Run multiple service versions simultaneously using subject prefix isolation:

```go
// Version 1 service
import orderv1 "yourmodule/gen/order/v1"

orderv1.RegisterOrderService(nc, svcV1)
// Registered at: api.v1.order_service.*

// Version 2 service  
import orderv2 "yourmodule/gen/order/v2"

orderv2.RegisterOrderService(nc, svcV2)
// Registered at: api.v2.order_service.*

// Clients choose version by import
clientV1 := orderv1.NewOrderServiceNatsClient(nc)
clientV2 := orderv2.NewOrderServiceNatsClient(nc)
```

Clients automatically target the correct version based on the imported package.

## Architecture

### Code Generation Pipeline

```
proto files
    ↓
buf generate
    ↓
├─→ protoc-gen-go          → messages (service.pb.go)
├─→ protoc-gen-go-grpc     → gRPC (service_grpc.pb.go)
├─→ protoc-gen-grpc-gateway → REST (service.pb.gw.go)
├─→ protoc-gen-openapiv2   → OpenAPI (service.swagger.yaml)
└─→ protoc-gen-nats-micro  → NATS (service_nats.pb.go)
```

### Two-Phase Build

The plugin uses a two-phase build to read custom proto extensions:

1. **Phase 1**: Generate extension types from `nats/options.proto`
2. **Phase 2**: Build plugin that imports and reads those extensions
3. **Phase 3**: Generate service code with embedded configuration

This is orchestrated via `go:generate` or Task:

```bash
task generate:extensions  # Phase 1
task build:plugin        # Phase 2  
task generate           # Phase 3
```

## Comparison: NATS vs gRPC

| Aspect | NATS (this plugin) | gRPC (standard) |
|--------|-------------------|-----------------|
| **Transport** | NATS messaging | HTTP/2 |
| **Service Discovery** | Built-in via NATS | Requires infrastructure (Consul, etcd) |
| **Load Balancing** | Automatic (queue groups) | Client-side or proxy |
| **Configuration** | Zero (proto-driven) | Service mesh or manual |
| **Network Topology** | Pub/sub, request/reply | Point-to-point |
| **Deployment** | Start instances, auto-discover | DNS, load balancers, service mesh |
| **Protocol** | Binary protobuf over NATS | Binary protobuf over HTTP/2 |
| **Streaming** | Native (JetStream) | Requires bidirectional streams |
| **Best For** | Internal microservices | External APIs, cross-org |

Both are generated from identical proto files in this project.

## Extending to Other Languages

The plugin uses a template-based architecture. To add a new language:

1. Create `tools/protoc-gen-nats-micro/generator/templates/<language>/`
2. Add templates: `header.tmpl`, `service.tmpl`, `client.tmpl`
3. Register in `generator/generator.go`

See [tools/protoc-gen-nats-micro/README.md](tools/protoc-gen-nats-micro/README.md) for details.

Planned languages: Rust, TypeScript, Python

## Examples

### Multi-Service Architecture

See `examples/complex-server` for a complete example with:

- Product catalog service
- Order service (v1 and v2)
- Service-to-service communication
- Shared protobuf types

### Client Usage

See `examples/complex-client` for examples of:

- Creating resources across services
- Handling errors with type-safe error checking
- Working with complex types
- API versioning

#### Error Handling

The generated code includes structured error types with helper functions for type-safe error checking:

```go
client := productv1.NewProductServiceNatsClient(nc)
product, err := client.GetProduct(ctx, &productv1.GetProductRequest{Id: "123"})
if err != nil {
    // Check error type using generated helpers
    if productv1.IsNotFound(err) {
        log.Println("Product not found")
        return
    }
    if productv1.IsInvalidArgument(err) {
        log.Println("Invalid request:", err)
        return
    }
    // Handle other errors
    log.Fatal("Unexpected error:", err)
}
```

Service implementations can return semantic errors:

```go
func (s *productService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
    product, exists := s.products[req.Id]
    if !exists {
        // Return structured error that client can check
        return nil, productv1.NewNotFoundError("GetProduct", fmt.Sprintf("product %s not found", req.Id))
    }
    return &productv1.GetProductResponse{Product: product}, nil
}
```

Available error codes:
- `INVALID_ARGUMENT` - Bad request data
- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Resource already exists
- `PERMISSION_DENIED` - Permission denied
- `UNAUTHENTICATED` - Authentication required
- `INTERNAL` - Server error
- `UNAVAILABLE` - Service unavailable

### REST Gateway

See `examples/rest-gateway` for HTTP/JSON access:

- OpenAPI spec serving
- Swagger UI integration
- CORS configuration

## Development

### Project Structure

```
.
├── proto/                     # Protobuf definitions
│   ├── nats/                  # NATS extension definitions
│   ├── order/v1/              # Order service v1
│   ├── order/v2/              # Order service v2
│   ├── product/v1/            # Product service
│   └── common/                # Shared types
├── gen/                       # Generated code (gitignored)
├── examples/                  # Example applications
│   ├── complex-server/        # Multi-service server
│   ├── complex-client/        # Client example
│   ├── rest-gateway/          # HTTP/JSON gateway
│   └── openapi-merge/         # OpenAPI spec combiner
├── tools/
│   └── protoc-gen-nats-micro/ # Plugin source
│       ├── generator/         # Code generation logic
│       │   └── templates/     # Language templates
│       ├── main.go            # Plugin entry point
│       └── README.md          # Plugin documentation
├── buf.yaml                   # Buf configuration
├── buf.gen.yaml               # Code generation config
├── buf.gen.extensions.yaml    # Extension generation config
└── Taskfile.yml               # Build automation
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/toyz/protoc-gen-nats-micro
cd protoc-gen-nats-micro

# Generate code
task generate

# Build plugin
task build:plugin

# Run tests
task test

# Clean generated files
task clean
```

### Available Tasks

```bash
task --list

* build          Build all example applications
* clean          Remove all generated files  
* generate       Generate all protobuf code
* test           Run all tests
* nats           Start NATS server in Docker
* run:server     Run complex-server example
* run:client     Run complex-client example
* run:gateway    Run REST gateway
```

## Contributing

Contributions welcome. Areas of interest:

- Additional language templates (Rust, TypeScript, Python)
- Streaming support (bidirectional, server streaming)
- Enhanced error handling patterns
- Observability integrations (OpenTelemetry)
- Performance benchmarks vs gRPC

## Related Projects

- [nats.go](https://github.com/nats-io/nats.go) - Official NATS Go client
- [nats.go/micro](https://github.com/nats-io/nats.go/tree/main/micro) - Microservices framework
- [buf](https://buf.build) - Modern protobuf toolchain
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) - REST gateway for gRPC

## License

MIT License - See LICENSE file for details

## Author

Built as an R&D exploration of NATS-based microservice patterns with modern protobuf tooling.
