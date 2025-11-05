# protoc-gen-nats-micro
**Codename: Apex**

[![Go Version](https://img.shields.io/github/go-mod/go-version/toyz/protoc-gen-nats-micro)](https://github.com/Toyz/protoc-gen-nats-micro)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Built by Helba](https://img.shields.io/badge/built%20by-helba.ai-cyan)](https://helba.ai)

A Protocol Buffers compiler plugin that generates type-safe NATS microservice code. Define services in protobuf, get production-ready NATS microservices with automatic service discovery, load balancing, and zero configuration.

> *The apex predator of NATS code generation.*

## Overview

`protoc-gen-nats-micro` is a code generation tool that brings the developer experience of gRPC to NATS.io. Write standard `.proto` files and generate:

- **NATS microservices** using the official `nats.io/micro` framework
- **gRPC services** for comparison and compatibility  
- **REST gateways** via grpc-gateway
- **OpenAPI specifications** for documentation

All from a single protobuf definition.

## Motivation

**Why this exists:**

The NATS ecosystem lacked a modern, maintained code generation solution comparable to gRPC. Existing tools like [nRPC](https://github.com/nats-rpc/nrpc) were abandoned, used outdated patterns, and didn't integrate with the official `nats.io/micro` framework or modern protobuf toolchains. This project fills that gap.

**What makes this different:**

- **Modern micro.Service framework** - Built on NATS' official microservices API
- **Type-safe error handling** - Generated error constants
- **Context propagation** - Proper context handling with configurable timeouts
- **google.protobuf.Duration** - Industry-standard timeout configuration
- **Multi-level timeouts** - Service defaults, endpoint overrides, runtime options
- **Zero configuration** - Everything defined in proto files
- **Production-ready** - Clean generated code, elegant abstractions

**Compared to nRPC:**

| Feature | protoc-gen-nats-micro (Apex) | nRPC |
|---------|------------------------------|------|
| **NATS Framework** | micro.Service (official) | Manual MsgHandler |
| **Context Support** | With timeout handling | Basic |
| **Error Constants** | Generated type-safe | Magic strings |
| **Timeout Config** | Proto Duration (multi-level) | Not supported |
| **Service Discovery** | Automatic | Manual setup |
| **Maintenance** | Active | Abandoned (alpha) |
| **Generated Code** | Modern, idiomatic | Basic |

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
- **Context propagation** - Proper context handling with timeout support
- **Configurable timeouts** - Service-level, endpoint-level, and runtime options using `google.protobuf.Duration`
- **Generated error constants** - Type-safe error codes (no magic strings)
- **Elegant code generation** - Clean, idiomatic Go using modern patterns
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
import "google/protobuf/duration.proto";

service OrderService {
  option (nats.micro.service) = {
    subject_prefix: "api.v1"
    name: "order_service"
    version: "1.0.0"
    description: "Order management service"
    timeout: {seconds: 30}  // Default 30s timeout for all endpoints
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
  
  rpc SearchOrders(SearchOrdersRequest) returns (SearchOrdersResponse) {
    option (nats.micro.endpoint) = {
      timeout: {seconds: 60}  // Override: 60s for search operations
    };
    option (google.api.http) = {
      get: "/v1/orders/search"
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
    "time"
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

    // Register with configuration from proto (30s default timeout)
    // Service automatically registered at "api.v1.order_service"
    _, err = orderv1.RegisterOrderServiceHandlers(nc, svc)
    if err != nil {
        log.Fatal(err)
    }
    
    // Or override timeout at runtime
    _, err = orderv1.RegisterOrderServiceHandlers(nc, svc,
        orderv1.WithTimeout(45 * time.Second),
    )

    // Service is now discoverable with automatic load balancing
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

From a single `.proto` file, **this plugin** generates:

```
gen/order/v1/
├── service.pb.go           # Standard protobuf messages (protoc-gen-go)
└── service_nats.pb.go      # NATS service and client (protoc-gen-nats-micro)
```

**This example project** also uses additional plugins for demonstration:
- `protoc-gen-go-grpc` → gRPC services (`service_grpc.pb.go`)
- `protoc-gen-grpc-gateway` → REST gateway (`service.pb.gw.go`)
- `protoc-gen-openapiv2` → OpenAPI specs (`service.swagger.yaml`)

These are **optional** - you only need `protoc-gen-go` and `protoc-gen-nats-micro` for NATS microservices.

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

### Service Introspection

Both server and client code provide introspection to discover available endpoints:

```go
// Register service and get wrapped service with Endpoints() method
svc, err := productv1.RegisterProductServiceHandlers(nc, impl)
if err != nil {
    log.Fatal(err)
}

// Get all endpoints from the service
for _, ep := range svc.Endpoints() {
    fmt.Printf("%s -> %s\n", ep.Name, ep.Subject)
    // Output:
    // CreateProduct -> api.v1.create_product
    // GetProduct -> api.v1.get_product
    // UpdateProduct -> api.v1.update_product
    // DeleteProduct -> api.v1.delete_product
    // SearchProducts -> api.v1.search_products
}

// Client also has Endpoints() method
client := productv1.NewProductServiceNatsClient(nc)
endpoints := client.Endpoints()
// Returns same info with client's configured subject prefix

// The service wrapper embeds micro.Service, so you can call all micro.Service methods:
svc.Stop()
svc.Info()
svc.Stats()
```

This is useful for:
- **Service discovery** - List all available operations
- **Monitoring** - Track which subjects to monitor
- **Debugging** - Verify correct subject configuration
- **Documentation** - Generate API docs from live services

## Configuration

Service configuration is defined in proto files using custom options:

```protobuf
import "nats/options.proto";
import "google/protobuf/duration.proto";

service OrderService {
  option (nats.micro.service) = {
    subject_prefix: "api.v1"        // NATS subject namespace
    name: "order_service"           // Service name for discovery
    version: "1.0.0"                // Semantic version
    description: "Order management" // Human-readable description
    timeout: {seconds: 30}          // Default timeout for all endpoints
  };
  
  rpc SlowOperation(Request) returns (Response) {
    option (nats.micro.endpoint) = {
      timeout: {seconds: 120}       // Override timeout for this endpoint
    };
  }
}
```

These values are read at code generation time and embedded in the generated code. No runtime configuration files needed.

### Timeout Configuration

Timeouts can be configured at three levels (in order of precedence):

1. **Runtime override** (highest priority):
   ```go
   orderv1.RegisterOrderServiceHandlers(nc, svc,
       orderv1.WithTimeout(45 * time.Second),
   )
   ```

2. **Endpoint-level** (per-method in proto):
   ```protobuf
   rpc SearchProducts(...) returns (...) {
     option (nats.micro.endpoint) = {
       timeout: {seconds: 60}  // This method gets 60s
     };
   }
   ```

3. **Service-level** (default for all methods):
   ```protobuf
   service ProductService {
     option (nats.micro.service) = {
       timeout: {seconds: 30}  // All methods default to 30s
     };
   }
   ```

If no timeout is configured, handlers use `context.Background()` with no timeout.

### Runtime Overrides (Optional)

While configuration lives in proto files, you can override at runtime:

```go
orderv1.RegisterOrderServiceHandlers(nc, svc,
    orderv1.WithSubjectPrefix("custom.prefix"),
    orderv1.WithVersion("2.0.0"),
    orderv1.WithTimeout(45 * time.Second),
)
```

### Metadata Management

Service metadata can be defined in proto and managed at runtime:

**Proto definition** (embedded at code generation):
```protobuf
service ProductService {
  option (nats.micro.service) = {
    subject_prefix: "api.v1"
    metadata: {
      key: "environment"
      value: "production"
    }
    metadata: {
      key: "team"
      value: "platform"
    }
  };
}
```

**Runtime options** (two approaches):

1. **Replace all metadata** - Completely overrides proto-defined metadata:
   ```go
   productv1.RegisterProductServiceHandlers(nc, svc,
       productv1.WithMetadata(map[string]string{
           "custom_key": "custom_value",
           // Proto metadata is discarded
       }),
   )
   ```

2. **Merge with proto metadata** (recommended) - Adds or updates entries:
   ```go
   productv1.RegisterProductServiceHandlers(nc, svc,
       productv1.WithAdditionalMetadata(map[string]string{
           "instance_id": uuid.New().String(),
           "hostname":    "server-1",
           // Proto metadata is preserved and extended
       }),
   )
   ```

Use `WithAdditionalMetadata()` to add runtime context (instance IDs, hostnames, regions) while keeping compile-time metadata (team, environment, version).

## API Versioning

Run multiple service versions simultaneously using subject prefix isolation:

```go
// Version 1 service
import orderv1 "yourmodule/gen/order/v1"

orderv1.RegisterOrderServiceHandlers(nc, svcV1)
// Registered at: api.v1.order_service.*

// Version 2 service  
import orderv2 "yourmodule/gen/order/v2"

orderv2.RegisterOrderServiceHandlers(nc, svcV2)
// Registered at: api.v2.order_service.*

// Clients choose version by import
clientV1 := orderv1.NewOrderServiceNatsClient(nc)
clientV2 := orderv2.NewOrderServiceNatsClient(nc)
```

Clients automatically target the correct version based on the imported package.

## Architecture

### Code Generation Pipeline

This plugin integrates with the standard protobuf toolchain:

```
proto files
    ↓
buf generate (or protoc)
    ↓
├─→ protoc-gen-go          → messages (service.pb.go)
└─→ protoc-gen-nats-micro  → NATS (service_nats.pb.go) ⭐

Optional (used in this example project):
├─→ protoc-gen-go-grpc     → gRPC (service_grpc.pb.go)
├─→ protoc-gen-grpc-gateway → REST (service.pb.gw.go)
└─→ protoc-gen-openapiv2   → OpenAPI (service.swagger.yaml)
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

**Client-side error handling:**
```go
client := productv1.NewProductServiceNatsClient(nc)
product, err := client.GetProduct(ctx, &productv1.GetProductRequest{Id: "123"})
if err != nil {
    if productv1.IsProductServiceNotFound(err) {
        log.Println("Product not found")
        return
    }
    if productv1.IsProductServiceInvalidArgument(err) {
        log.Println("Invalid request:", err)
        return
    }
    log.Fatal("Unexpected error:", err)
}
```

**Server-side error responses:**

Service implementations can return errors in three ways:

1. **Return generated error types** (recommended):
```go
func (s *productService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
    product, exists := s.products[req.Id]
    if !exists {
        return nil, productv1.NewProductServiceNotFoundError("GetProduct", "product not found")
    }
    return &productv1.GetProductResponse{Product: product}, nil
}
```

2. **Implement custom error interfaces** (advanced):
```go
type OutOfStockError struct {
    ProductID string
    Requested int
    Available int
}

func (e *OutOfStockError) Error() string {
    return fmt.Sprintf("product %s: requested %d, only %d available", e.ProductID, e.Requested, e.Available)
}

// Implement these methods to control NATS error response:
func (e *OutOfStockError) NatsErrorCode() string {
    return productv1.ProductServiceErrCodeUnavailable
}

func (e *OutOfStockError) NatsErrorMessage() string {
    return e.Error()
}

func (e *OutOfStockError) NatsErrorData() []byte {
    // Optional: return custom error data (e.g., JSON, protobuf)
    return nil
}

// Now you can return your custom error:
func (s *productService) CreateOrder(ctx context.Context, req *productv1.CreateOrderRequest) (*productv1.CreateOrderResponse, error) {
    if stock < req.Quantity {
        return nil, &OutOfStockError{
            ProductID: req.ProductId,
            Requested: int(req.Quantity),
            Available: stock,
        }
    }
    // ...
}
```

3. **Return generic errors** (falls back to INTERNAL):
```go
func (s *productService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
    product, err := s.db.FindProduct(req.Id)
    if err != nil {
        return nil, err // Will be sent as INTERNAL error
    }
    return &productv1.GetProductResponse{Product: product}, nil
}
```

**Custom error interfaces:**
- Implement `NatsErrorCode() string` to set custom error codes
- Implement `NatsErrorMessage() string` to set custom error messages
- Implement `NatsErrorData() []byte` to attach custom data

The handler checks for these methods using inline interface assertions (no additional dependencies or interface types needed).

Available error codes:
- `INVALID_ARGUMENT` - Bad request data
- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Resource already exists
- `PERMISSION_DENIED` - Permission denied
- `UNAUTHENTICATED` - Authentication required
- `INTERNAL` - Server error (default for unhandled errors)
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

Built by [Helba](https://helba.ai) - Digital Architect specializing in high-performance backend systems.

An R&D exploration of NATS-based microservice patterns with modern protobuf tooling, pushing the boundaries of what's possible with code generation.

---

*"The apex predator of NATS code generation."*
