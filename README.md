# protoc-gen-nats-micro

[![Go Version](https://img.shields.io/github/go-mod/go-version/toyz/protoc-gen-nats-micro)](https://github.com/Toyz/protoc-gen-nats-micro)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Protocol Buffers compiler plugin that generates type-safe NATS microservice code using the official `nats.io/micro` framework.

## Overview

Write standard `.proto` files, run `buf generate`, get production-ready NATS microservices with automatic service discovery, load balancing, and zero configuration.

**This plugin generates:** NATS microservice code (server interfaces, clients, error handling)

**Demo project also includes:** gRPC, REST gateway, and OpenAPI generation to demonstrate interoperability - these are optional and not required for NATS services.

## Motivation

Existing NATS code generation tools like [nRPC](https://github.com/nats-rpc/nrpc) were abandoned and didn't integrate with the official `nats.io/micro` framework.

**Key features:**

- Official micro.Service framework integration
- Type-safe error handling and context propagation
- Multi-level timeout configuration via `google.protobuf.Duration`
- Service/endpoint metadata and interceptors
- Multi-language support (Go, TypeScript, Rust planned)

**vs nRPC**: Active maintenance, official micro.Service API, modern idioms, configurable timeouts

**vs gRPC**: Better for internal microservices - built-in service discovery/load balancing, no service mesh needed

## Features

- **Zero configuration** - Service metadata defined in proto files
- **Type-safe code** - Compile-time safety for requests/responses/errors
- **Configurable timeouts** - Service, endpoint, and runtime levels via `google.protobuf.Duration`
- **Metadata** - Service and endpoint-level for discovery and operations
- **Interceptors** - Middleware for logging, auth, tracing (client and server)
- **Headers** - Bidirectional header propagation (request and response)
- **Package-level shared types** - One shared file per package eliminates duplication
- **Skip support** - Exclude services or endpoints from generation
- **Multi-language** - Go, TypeScript (Rust planned)
- **Standard tooling** - Works with `buf`, `protoc`, existing workflows
- **Automatic service discovery** - Via NATS, no external dependencies
- **Built-in load balancing** - NATS queue groups
- **API versioning** - Subject prefix isolation

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
# Generate NATS code (Go + TypeScript)
task generate

# Or use buf directly
buf generate
```

**Note:** This project's buf config also generates gRPC, REST gateway, and OpenAPI for demonstration purposes. For production, you only need `protoc-gen-go` and `protoc-gen-nats-micro`.

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

import "natsmicro/options.proto";
import "google/api/annotations.proto";
import "google/protobuf/duration.proto";

service OrderService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1"
    name: "order_service"
    version: "1.0.0"
    description: "Order management service"
    timeout: {seconds: 30}  // Default 30s timeout for all endpoints
    metadata: {
      key: "team"
      value: "orders"
    }
  };

  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse) {
    option (natsmicro.endpoint) = {
      metadata: {
        key: "operation"
        value: "write"
      }
      metadata: {
        key: "idempotent"
        value: "false"
      }
    };
    option (google.api.http) = {
      post: "/v1/orders"
      body: "*"
    };
  }

  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse) {
    option (natsmicro.endpoint) = {
      metadata: {
        key: "operation"
        value: "read"
      }
      metadata: {
        key: "cacheable"
        value: "true"
      }
    };
    option (google.api.http) = {
      get: "/v1/orders/{id}"
    };
  }

  rpc SearchOrders(SearchOrdersRequest) returns (SearchOrdersResponse) {
    option (natsmicro.endpoint) = {
      timeout: {seconds: 60}  // Override: 60s for search operations
      metadata: {
        key: "operation"
        value: "read"
      }
      metadata: {
        key: "expensive"
        value: "true"
      }
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

**Required for NATS services:**

```
gen/order/v1/
├── service.pb.go           # Protobuf messages (protoc-gen-go)
├── service_nats.pb.go      # NATS service/client (protoc-gen-nats-micro)
└── shared_nats.pb.go       # Shared types (protoc-gen-nats-micro)
```

**Demo project also generates** (optional, for comparison):

- `service_grpc.pb.go` - gRPC services
- `service.pb.gw.go` - REST gateway
- `service.swagger.yaml` - OpenAPI specs

### Package-Level Shared File

Multiple services in the same package share one `shared_nats.pb.go` containing error constants, `RegisterOption`, and `NatsClientOption`. This eliminates duplication across services.

```
gen/order/v1/
├── service_nats.pb.go         # OrderService
├── fulfillment_nats.pb.go     # OrderFulfillmentService
└── shared_nats.pb.go          # Shared by all services in order/v1
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

### Service Introspection

Services expose an `Endpoints()` method for discovery:

```go
svc, _ := productv1.RegisterProductServiceHandlers(nc, impl)
for _, ep := range svc.Endpoints() {
    fmt.Printf("%s -> %s\n", ep.Name, ep.Subject)
}

// Client also has Endpoints()
client := productv1.NewProductServiceNatsClient(nc)
endpoints := client.Endpoints()

// Embeds micro.Service
svc.Stop()
svc.Info()
svc.Stats()
```

## TypeScript Support

Full TypeScript support with same features as Go. See [TYPESCRIPT.md](TYPESCRIPT.md) for details.

```typescript
import { connect } from "nats";
import { ProductServiceNatsClient } from "./gen/product/v1/service_nats.pb";

const nc = await connect({ servers: "nats://localhost:4222" });
const client = new ProductServiceNatsClient(nc);
const response = await client.getProduct({ id: "123" });
```

## Configuration

Service configuration is defined in proto files using custom options:

```protobuf
import "nats/options.proto";
import "google/protobuf/duration.proto";

service OrderService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1"
    name: "order_service"
    version: "1.0.0"
    description: "Order management"
    timeout: {seconds: 30}
  };

  rpc SlowOperation(Request) returns (Response) {
    option (natsmicro.endpoint) = {
      timeout: {seconds: 120}
    };
  }
}
```

See [API.md](API.md) for complete reference of all options.

### Timeout Configuration

Three levels (runtime > endpoint > service):

```go
// 1. Runtime override (highest priority)
orderv1.RegisterOrderServiceHandlers(nc, svc,
    orderv1.WithTimeout(45 * time.Second),
)
```

```protobuf
// 2. Endpoint-level (per method)
rpc SearchProducts(...) returns (...) {
  option (natsmicro.endpoint) = {timeout: {seconds: 60}};
}

// 3. Service-level (default)
service ProductService {
  option (natsmicro.service) = {timeout: {seconds: 30}};
}
```

### Runtime Overrides

```go
orderv1.RegisterOrderServiceHandlers(nc, svc,
    orderv1.WithSubjectPrefix("custom.prefix"),
    orderv1.WithVersion("2.0.0"),
    orderv1.WithTimeout(45 * time.Second),
)
```

### Metadata

Metadata is configured at **service-level** (organizational info) and **endpoint-level** (operation characteristics).

**Service metadata in proto:**

```protobuf
service ProductService {
  option (natsmicro.service) = {
    metadata: {key: "team" value: "platform"}
    metadata: {key: "environment" value: "production"}
  };
}
```

**Runtime options:**

```go
// Replace all metadata
productv1.RegisterProductServiceHandlers(nc, svc,
    productv1.WithMetadata(map[string]string{"custom": "value"}),
)

// Merge with proto metadata (recommended)
productv1.RegisterProductServiceHandlers(nc, svc,
    productv1.WithAdditionalMetadata(map[string]string{
        "instance_id": uuid.New().String(),
        "hostname":    os.Hostname(),
    }),
)
```

**Endpoint metadata in proto:**

```protobuf
rpc GetProduct(...) returns (...) {
  option (natsmicro.endpoint) = {
    metadata: {key: "operation" value: "read"}
    metadata: {key: "cacheable" value: "true"}
    metadata: {key: "cache_ttl" value: "300"}
  };
}
```

Common patterns: operation type (`read|write|delete`), caching (`cacheable`, `cache_ttl`), performance (`expensive`), auth (`requires_auth`), versioning (`deprecated`)

### Skip Support

Exclude services or endpoints from generation:

```protobuf
service AdminService {
  option (natsmicro.service) = {skip: true};  // Skip entire service
}

rpc AdminReset(...) returns (...) {
  option (natsmicro.endpoint) = {skip: true};  // Skip specific method
}
```

## API Versioning

Run multiple versions simultaneously via subject prefix isolation:

```go
import orderv1 "yourmodule/gen/order/v1"
import orderv2 "yourmodule/gen/order/v2"

orderv1.RegisterOrderServiceHandlers(nc, svcV1)  // api.v1.order_service.*
orderv2.RegisterOrderServiceHandlers(nc, svcV2)  // api.v2.order_service.*

clientV1 := orderv1.NewOrderServiceNatsClient(nc)
clientV2 := orderv2.NewOrderServiceNatsClient(nc)
```

## Interceptors and Headers

Interceptors provide middleware for logging, auth, metrics, and tracing.

### Server Interceptors

```go
func loggingInterceptor(ctx context.Context, req interface{}, info *productv1.UnaryServerInfo, handler productv1.UnaryHandler) (interface{}, error) {
    start := time.Now()

    // Read incoming request headers
    if headers := productv1.IncomingHeaders(ctx); headers != nil {
        if traceID, ok := headers["X-Trace-Id"]; ok && len(traceID) > 0 {
            log.Printf("[%s] Trace-ID: %s", info.Method, traceID[0])
        }
    }

    // Set response headers that will be sent back to client
    responseHeaders := nats.Header{}
    responseHeaders.Set("X-Server-Version", "1.0.0")
    responseHeaders.Set("X-Request-Id", generateRequestID())
    productv1.SetResponseHeaders(ctx, responseHeaders)

    // Call the actual handler
    resp, err := handler(ctx, req)

    duration := time.Since(start)
    log.Printf("[%s] completed in %v", info.Method, duration)

    return resp, err
}

// Register with interceptor
productv1.RegisterProductServiceHandlers(nc, impl,
    productv1.WithServerInterceptor(loggingInterceptor),
)
```

**Chain multiple:**

```go
productv1.RegisterProductServiceHandlers(nc, impl,
    productv1.WithServerInterceptor(authInterceptor),
    productv1.WithServerInterceptor(metricsInterceptor),
    productv1.WithServerInterceptor(loggingInterceptor),
)  // Execution: auth -> metrics -> logging -> handler
```

### Client Interceptors

```go
func clientLoggingInterceptor(ctx context.Context, method string, req, reply interface{}, invoker productv1.UnaryInvoker) error {
    // Add request headers
    headers := nats.Header{}
    headers.Set("X-Trace-Id", generateTraceID())
    headers.Set("X-Client-Version", "1.0.0")
    ctx = productv1.WithOutgoingHeaders(ctx, headers)

    // Make the call
    err := invoker(ctx, method, req, reply)

    // Read response headers
    if respHeaders := productv1.ResponseHeaders(ctx); respHeaders != nil {
        if serverVer, ok := respHeaders["X-Server-Version"]; ok && len(serverVer) > 0 {
            log.Printf("Server version: %s", serverVer[0])
        }
    }

    return err
}

client := productv1.NewProductServiceNatsClient(nc,
    productv1.WithClientInterceptor(clientLoggingInterceptor),
)
```

### Bidirectional Headers

**Request headers** (client → server):

```go
// Client
ctx = productv1.WithOutgoingHeaders(ctx, headers)
// Server
headers := productv1.IncomingHeaders(ctx)
```

**Response headers** (server → client):

```go
// Server
productv1.SetResponseHeaders(ctx, headers)
// Client
headers := productv1.ResponseHeaders(ctx)
```

Use cases: distributed tracing, authentication tokens, correlation IDs, versioning

## Architecture

### Code Generation Pipeline

This plugin integrates with the standard protobuf toolchain:

```
proto files
    ↓
buf generate (or protoc)
    ↓
├── protoc-gen-go          -> messages (service.pb.go)
└── protoc-gen-nats-micro  -> NATS (service_nats.pb.go)

Optional (used in this example project):
├── protoc-gen-go-grpc     -> gRPC (service_grpc.pb.go)
├── protoc-gen-grpc-gateway -> REST (service.pb.gw.go)
└── protoc-gen-openapiv2   -> OpenAPI (service.swagger.yaml)
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

## Extending to Other Languages

Template-based architecture. Add `<language>/` folder with templates, register in `generator/generator.go`. See [tools/protoc-gen-nats-micro/README.md](tools/protoc-gen-nats-micro/README.md).

Planned: Rust, Python

## Examples

- `examples/complex-server` - Multi-service setup (Product, Order v1/v2)
- `examples/complex-client` - Client usage with error handling
- `examples/rest-gateway` - HTTP/JSON gateway (optional)
- `examples/simple-ts` - TypeScript client/server

### Error Handling

**Client-side:**

```go
product, err := client.GetProduct(ctx, &productv1.GetProductRequest{Id: "123"})
if productv1.IsProductServiceNotFound(err) {
    // Handle not found
}
```

**Server-side (3 options):**

1. Generated error types (recommended):

```go
return nil, productv1.NewProductServiceNotFoundError("GetProduct", "not found")
```

2. Custom errors (implement `NatsErrorCode()`, `NatsErrorMessage()`, `NatsErrorData()`):

```go
type OutOfStockError struct { ProductID string }
func (e *OutOfStockError) Error() string { return "out of stock" }
func (e *OutOfStockError) NatsErrorCode() string { return productv1.ProductServiceErrCodeUnavailable }
```

3. Generic errors (become INTERNAL):

```go
return nil, fmt.Errorf("database error")
```

Built-in codes: `INVALID_ARGUMENT`, `NOT_FOUND`, `ALREADY_EXISTS`, `PERMISSION_DENIED`, `UNAUTHENTICATED`, `INTERNAL`, `UNAVAILABLE`

**Custom error codes** — define application-specific codes in your proto:

```protobuf
service OrderService {
  option (natsmicro.service) = {
    error_codes: ["ORDER_EXPIRED", "PAYMENT_FAILED", "STOCK_UNAVAILABLE"]
  };
}
```

This generates typed constants, constructors, and checkers for each code:

```go
// Server
return nil, orderv1.NewOrderServiceOrderExpiredError("CreateOrder", "expired after 30m")

// Client
if orderv1.IsOrderServiceOrderExpired(err) {
    log.Println("Order expired, please resubmit")
}
```

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

## Streaming

**Not supported.** NATS micro uses request-reply (1:1), not gRPC-style streaming.

**Alternatives:**

- **Pagination**: Use page tokens for large results
- **JetStream**: Use NATS JetStream directly for real-time event streams
- **Multiple calls**: Make sequential unary requests

For streaming, use JetStream - it's more powerful than trying to replicate gRPC streaming in micro.

## Contributing

Contributions welcome: language templates, observability integrations, benchmarks, interceptor examples.

## Related Projects

- [nats.go](https://github.com/nats-io/nats.go) - Official NATS Go client
- [nats.go/micro](https://github.com/nats-io/nats.go/tree/main/micro) - Microservices framework
- [buf](https://buf.build) - Modern protobuf toolchain
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) - REST gateway for gRPC

## License

MIT License - See LICENSE file for details

## Author

Created by [Helba](https://helba.ai)

A Protocol Buffers code generator for NATS microservices, integrating modern protobuf tooling with the official nats.go/micro framework.
