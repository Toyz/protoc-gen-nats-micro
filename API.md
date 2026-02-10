# API Reference

Complete reference for `protoc-gen-nats-micro` proto extension options.

## Table of Contents

- [Service Options](#service-options)
- [Endpoint Options](#endpoint-options)
- [Complete Examples](#complete-examples)
- [Generated Code Reference](#generated-code-reference)

## Service Options

Service-level configuration applied to the entire service. Defined using `option (natsmicro.service)`.

### subject_prefix

**Type:** `string`  
**Default:** Snake-case of service name (e.g., `product_service`)  
**Required:** No

NATS subject prefix for all endpoints in the service. Each endpoint becomes `{subject_prefix}.{method_name}`.

```protobuf
option (natsmicro.service) = {
  subject_prefix: "api.v1"
};
// Results in: api.v1.create_product, api.v1.get_product, etc.
```

**Best practices:**

- Use versioned prefixes: `api.v1`, `api.v2`
- Include environment for multi-tenant: `prod.api.v1`, `staging.api.v1`
- Keep it simple for discoverability

### name

**Type:** `string`  
**Default:** Snake-case of service name  
**Required:** No

Service name for NATS micro service registration and discovery.

```protobuf
option (natsmicro.service) = {
  name: "product_service"
};
```

**Best practices:**

- Use snake_case for consistency
- Include domain context: `catalog_product_service`, `order_fulfillment_service`
- Keep it descriptive but concise

### version

**Type:** `string`  
**Default:** `"1.0.0"`  
**Required:** No

Semantic version for service discovery and monitoring.

```protobuf
option (natsmicro.service) = {
  version: "2.1.0"
};
```

**Best practices:**

- Follow [semver](https://semver.org/): `MAJOR.MINOR.PATCH`
- Increment MAJOR for breaking changes
- Use runtime override for build-time versions

### description

**Type:** `string`  
**Default:** Empty  
**Required:** No

Human-readable service description for documentation and discovery.

```protobuf
option (natsmicro.service) = {
  description: "Product catalog management with inventory tracking"
};
```

**Best practices:**

- Keep it concise (1-2 sentences)
- Describe the service's primary purpose
- Avoid implementation details

### metadata

**Type:** `map<string, string>`  
**Default:** Empty  
**Required:** No

Service-level key-value metadata for discovery, monitoring, and routing.

```protobuf
option (natsmicro.service) = {
  metadata: {key: "team" value: "platform"}
  metadata: {key: "environment" value: "production"}
  metadata: {key: "region" value: "us-west-2"}
};
```

**Common patterns:**

- **Organizational**: `team`, `owner`, `cost_center`
- **Environmental**: `environment`, `region`, `datacenter`
- **Operational**: `sla`, `criticality`, `on_call`

**Best practices:**

- Use lowercase keys with underscores
- Keep values simple (avoid JSON/complex data)
- Use endpoint metadata for operation-specific data
- Merge runtime metadata with `WithAdditionalMetadata()`

### timeout

**Type:** `google.protobuf.Duration`  
**Default:** No timeout (context.Background)  
**Required:** No

Default timeout for all endpoints in the service. Can be overridden per-endpoint or at runtime.

```protobuf
import "google/protobuf/duration.proto";

option (natsmicro.service) = {
  timeout: {seconds: 30}  // 30 second default
};
```

**Timeout precedence** (highest to lowest):

1. Runtime override: `WithTimeout(45 * time.Second)`
2. Endpoint-level: `option (natsmicro.endpoint) = {timeout: {seconds: 60}}`
3. Service-level: `option (natsmicro.service) = {timeout: {seconds: 30}}`
4. No timeout: `context.Background()`

**Best practices:**

- Set reasonable service defaults (10-30s for typical APIs)
- Override for expensive operations (search, reports)
- Consider downstream dependencies
- Monitor timeout rates in production

### skip

**Type:** `bool`  
**Default:** `false`  
**Required:** No

Skip code generation for the entire service.

```protobuf
service AdminService {
  option (natsmicro.service) = {
    skip: true
  };
  // Methods not generated
}
```

**Use cases:**

- Internal-only services not exposed via NATS
- Services under development
- Deprecated services being phased out
- Test/mock services

### json

**Type:** `bool`  
**Default:** `false`  
**Required:** No

Use JSON encoding instead of binary protobuf for messages.

```protobuf
option (natsmicro.service) = {
  json: true
};
```

**When to use:**

- Debugging (human-readable messages)
- Interop with non-protobuf systems
- Browser-based clients

**Trade-offs:**

- **Pros**: Human-readable, browser-friendly
- **Cons**: Larger message size, slower serialization, no runtime schema validation

**Best practices:**

- Use binary protobuf (default) for production
- Enable JSON for debugging environments only
- Consider performance impact for high-throughput services

## Endpoint Options

Method-level configuration for individual RPC endpoints. Defined using `option (natsmicro.endpoint)`.

### timeout

**Type:** `google.protobuf.Duration`  
**Default:** Service-level timeout  
**Required:** No

Override service-level timeout for this specific endpoint.

```protobuf
rpc SearchProducts(SearchRequest) returns (SearchResponse) {
  option (natsmicro.endpoint) = {
    timeout: {seconds: 60}  // Override: 60s for expensive search
  };
}
```

**Best practices:**

- Override for expensive operations (search, aggregations, reports)
- Set longer timeouts for batch operations
- Keep default for simple CRUD operations
- Document why timeouts differ from service default

### skip

**Type:** `bool`  
**Default:** `false`  
**Required:** No

Skip code generation for this specific endpoint.

```protobuf
rpc InternalDebugMethod(Request) returns (Response) {
  option (natsmicro.endpoint) = {
    skip: true  // Not exposed via NATS
  };
}
```

**Use cases:**

- Internal-only methods
- Deprecated endpoints
- Methods only for gRPC/REST, not NATS
- Test/debug endpoints excluded from production

### metadata

**Type:** `map<string, string>`  
**Default:** Empty  
**Required:** No

Endpoint-specific metadata for operation characteristics.

```protobuf
rpc GetProduct(GetProductRequest) returns (GetProductResponse) {
  option (natsmicro.endpoint) = {
    metadata: {key: "operation" value: "read"}
    metadata: {key: "cacheable" value: "true"}
    metadata: {key: "cache_ttl" value: "300"}
    metadata: {key: "idempotent" value: "true"}
  };
}
```

**Common patterns:**

- **Operation type**: `operation: "read|write|delete"`
- **Caching**: `cacheable: "true|false"`, `cache_ttl: "300"`
- **Idempotency**: `idempotent: "true|false"`
- **Performance**: `expensive: "true"`, `rate_limit: "100"`
- **Authorization**: `requires_auth: "true"`, `permission: "admin"`
- **Versioning**: `deprecated: "true"`, `since_version: "2.0"`

**Best practices:**

- Use endpoint metadata for operation-specific characteristics
- Use service metadata for organizational context
- Keep keys consistent across services
- Document metadata conventions in your team

## Complete Examples

### Basic Service

Minimal configuration with defaults:

```protobuf
syntax = "proto3";
package hello.v1;
import "natsmicro/options.proto";

service GreeterService {
  option (natsmicro.service) = {
    subject_prefix: "hello.v1"
  };

  rpc SayHello(HelloRequest) returns (HelloResponse) {}
}
```

Results in:

- Subject: `hello.v1.say_hello`
- Name: `greeter_service`
- Version: `1.0.0`
- No timeout

### Production Service

Full configuration with timeouts and metadata:

```protobuf
syntax = "proto3";
package product.v1;
import "natsmicro/options.proto";
import "google/protobuf/duration.proto";

service ProductService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1"
    name: "product_service"
    version: "2.0.0"
    description: "Product catalog management"
    timeout: {seconds: 30}
    metadata: {key: "team" value: "catalog"}
    metadata: {key: "environment" value: "production"}
  };

  rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse) {
    option (natsmicro.endpoint) = {
      metadata: {key: "operation" value: "write"}
      metadata: {key: "idempotent" value: "false"}
    };
  }

  rpc GetProduct(GetProductRequest) returns (GetProductResponse) {
    option (natsmicro.endpoint) = {
      metadata: {key: "operation" value: "read"}
      metadata: {key: "cacheable" value: "true"}
      metadata: {key: "cache_ttl" value: "300"}
    };
  }

  rpc SearchProducts(SearchRequest) returns (SearchResponse) {
    option (natsmicro.endpoint) = {
      timeout: {seconds: 60}  // Override for expensive operation
      metadata: {key: "operation" value: "read"}
      metadata: {key: "expensive" value: "true"}
    };
  }
}
```

### Multi-Version Service

Running v1 and v2 simultaneously:

```protobuf
// proto/order/v1/service.proto
package order.v1;
service OrderService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1"
    name: "order_service"
    version: "1.0.0"
  };
  rpc CreateOrder(CreateOrderRequestV1) returns (CreateOrderResponseV1) {}
}

// proto/order/v2/service.proto
package order.v2;
service OrderService {
  option (natsmicro.service) = {
    subject_prefix: "api.v2"
    name: "order_service"
    version: "2.0.0"
  };
  rpc CreateOrder(CreateOrderRequestV2) returns (CreateOrderResponseV2) {}
}
```

Subjects:

- v1: `api.v1.create_order`
- v2: `api.v2.create_order`

Both versions run simultaneously, clients choose by import.

### Selective Generation

Skip certain services or endpoints:

```protobuf
service PublicAPI {
  option (natsmicro.service) = {
    subject_prefix: "public.v1"
  };

  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}

  rpc AdminDeleteUser(DeleteUserRequest) returns (Empty) {
    option (natsmicro.endpoint) = {
      skip: true  // Admin method not exposed via NATS
    };
  }
}

service InternalDebugService {
  option (natsmicro.service) = {
    skip: true  // Entire service excluded
  };
  rpc DebugDump(Empty) returns (DebugInfo) {}
}
```

### JSON Encoding

For debugging or browser clients:

```protobuf
service DebugService {
  option (natsmicro.service) = {
    subject_prefix: "debug.v1"
    json: true  // Use JSON instead of binary protobuf
  };

  rpc InspectState(InspectRequest) returns (InspectResponse) {}
}
```

Messages sent as:

```json
{ "userId": "123", "includeMetadata": true }
```

Instead of binary protobuf.

## Generated Code Reference

### Go

#### Service Registration

```go
// Generated function signature
func RegisterProductServiceHandlers(
    nc *nats.Conn,
    impl ProductServiceNats,
    opts ...RegisterOption,
) (ProductServiceWrapper, error)

// RegisterOption types
func WithSubjectPrefix(prefix string) RegisterOption
func WithName(name string) RegisterOption
func WithVersion(version string) RegisterOption
func WithDescription(desc string) RegisterOption
func WithTimeout(timeout time.Duration) RegisterOption
func WithMetadata(metadata map[string]string) RegisterOption
func WithAdditionalMetadata(metadata map[string]string) RegisterOption
func WithServerInterceptor(interceptor UnaryServerInterceptor) RegisterOption
```

#### Client Creation

```go
// Generated client constructor
func NewProductServiceNatsClient(
    nc *nats.Conn,
    opts ...NatsClientOption,
) *ProductServiceNatsClient

// NatsClientOption types
func WithNatsClientSubjectPrefix(prefix string) NatsClientOption
func WithClientInterceptor(interceptor UnaryClientInterceptor) NatsClientOption
```

#### Error Handling

```go
// Generated error types
const (
    ProductServiceErrCodeInvalidArgument = "INVALID_ARGUMENT"
    ProductServiceErrCodeNotFound        = "NOT_FOUND"
    ProductServiceErrCodeAlreadyExists   = "ALREADY_EXISTS"
    ProductServiceErrCodePermissionDenied = "PERMISSION_DENIED"
    ProductServiceErrCodeUnauthenticated = "UNAUTHENTICATED"
    ProductServiceErrCodeInternal        = "INTERNAL"
    ProductServiceErrCodeUnavailable     = "UNAVAILABLE"
)

// Generated error constructors
func NewProductServiceInvalidArgumentError(method, message string) error
func NewProductServiceNotFoundError(method, message string) error
// ... etc

// Generated error checkers
func IsProductServiceInvalidArgument(err error) bool
func IsProductServiceNotFound(err error) bool
// ... etc
```

#### Headers

```go
// Server-side
func IncomingHeaders(ctx context.Context) nats.Header  // Read request headers
func SetResponseHeaders(ctx context.Context, headers nats.Header)  // Set response headers

// Client-side
func WithOutgoingHeaders(ctx context.Context, headers nats.Header) context.Context  // Set request headers
func ResponseHeaders(ctx context.Context) nats.Header  // Read response headers
```

#### Interceptors

```go
// Server interceptor signature
type UnaryServerInterceptor func(
    ctx context.Context,
    req interface{},
    info *UnaryServerInfo,
    handler UnaryHandler,
) (interface{}, error)

type UnaryServerInfo struct {
    Method string
}

type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// Client interceptor signature
type UnaryClientInterceptor func(
    ctx context.Context,
    method string,
    req, reply interface{},
    invoker UnaryInvoker,
) error

type UnaryInvoker func(
    ctx context.Context,
    method string,
    req, reply interface{},
) error
```

### TypeScript

#### Service Registration

```typescript
// Generated class
class ProductServiceNatsServer {
  constructor(
    nc: NatsConnection,
    impl: ProductServiceNats,
    opts?: ServerOptions,
  );
}

// ServerOptions interface
interface ServerOptions {
  subjectPrefix?: string;
  interceptors?: UnaryServerInterceptor[];
}
```

#### Client Creation

```typescript
// Generated class
class ProductServiceNatsClient {
  constructor(nc: NatsConnection, opts?: ClientOptions);

  async createProduct(
    req: CreateProductRequest,
  ): Promise<CreateProductResponse>;
  // ... other methods
}

// ClientOptions interface
interface ClientOptions {
  subjectPrefix?: string;
  headers?: MsgHdrs;
  interceptors?: UnaryClientInterceptor[];
}
```

#### Headers

```typescript
// Client
const client = new ProductServiceNatsClient(nc, {
  headers: headers(), // Set request headers
});

const responseHeaders = { value: null };
const response = await client.getProduct(req, { responseHeaders });
// Access response headers: responseHeaders.value

// Server
class MyService implements ProductServiceNats {
  async getProduct(
    req: GetProductRequest,
    info: ServerInfo,
  ): Promise<GetProductResponse> {
    // Read request headers
    const traceId = info.headers.get("X-Trace-Id");

    // Set response headers
    info.responseHeaders.set("X-Server-Version", "1.0.0");

    return response;
  }
}
```

## Runtime Configuration Priority

Options can be specified at multiple levels with the following precedence:

1. **Runtime** (highest priority)
   - `WithTimeout()`, `WithSubjectPrefix()`, etc.
   - Overrides all proto configuration

2. **Endpoint-level** (proto)
   - `option (natsmicro.endpoint) = {...}`
   - Overrides service-level defaults

3. **Service-level** (proto)
   - `option (natsmicro.service) = {...}`
   - Provides defaults for all endpoints

4. **Global defaults** (lowest priority)
   - Hardcoded defaults in generated code
   - Used when nothing is specified

### Example

```protobuf
service MyService {
  option (natsmicro.service) = {
    timeout: {seconds: 30}  // Default: 30s
  };

  rpc FastOp(Req) returns (Resp) {}  // Uses 30s

  rpc SlowOp(Req) returns (Resp) {
    option (natsmicro.endpoint) = {
      timeout: {seconds: 120}  // Override: 120s
    };
  }
}
```

```go
// Runtime override: all methods get 60s
RegisterMyServiceHandlers(nc, impl,
    WithTimeout(60 * time.Second),
)
```

Final timeouts:

- `FastOp`: 60s (runtime override)
- `SlowOp`: 60s (runtime override beats endpoint-level)

## See Also

- [README.md](README.md) - Project overview and quick start
- [TYPESCRIPT.md](TYPESCRIPT.md) - TypeScript-specific documentation
- [extensions/proto/natsmicro/options.proto](extensions/proto/natsmicro/options.proto) - Proto extension definitions
- [examples/](examples/) - Working code examples
