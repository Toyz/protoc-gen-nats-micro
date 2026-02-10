# API Reference

Complete reference for `protoc-gen-nats-micro` proto extension options.

## Table of Contents

- [Service Options](#service-options)
- [Endpoint Options](#endpoint-options)
- [KV Store & Object Store](#kv-store--object-store)
- [Key Templates](#key-templates)
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

## KV Store & Object Store

Auto-persist RPC responses to NATS JetStream KV Store or Object Store. Clients get convenience methods to read cached data directly — no RPC needed.

### kv_store

**Extension:** `natsmicro.kv_store`

Automatically persists the serialized RPC response to a NATS KV bucket after the handler responds.

```protobuf
rpc SaveProfile(SaveProfileRequest) returns (ProfileResponse) {
  option (natsmicro.kv_store) = {
    bucket: "user_profiles"       // KV bucket name
    key_template: "user.{id}"     // Key with {field} placeholders from request
    ttl: {seconds: 3600}          // Auto-expire after 1 hour (optional)
    description: "User profile cache"  // Bucket description (optional)
    max_history: 5                // Keep 5 revisions per key (optional)
  };
}
```

**Fields:**

| Field          | Type       | Description                                                                              |
| -------------- | ---------- | ---------------------------------------------------------------------------------------- |
| `bucket`       | `string`   | **Required.** Name of the KV bucket to persist to.                                       |
| `key_template` | `string`   | **Required.** Key pattern with `{field}` placeholders resolved from the request message. |
| `ttl`          | `Duration` | Optional. Auto-expire entries after this duration.                                       |
| `description`  | `string`   | Optional. Human-readable bucket description.                                             |
| `max_history`  | `int32`    | Optional. Revisions to keep per key (default 1, max 64).                                 |
| `client_only`  | `bool`     | Optional. Skip server-side auto-persist; only generate client read/write methods.        |

**What happens at runtime:**

1. Handler processes the request and returns a response
2. Response is serialized (protobuf or JSON)
3. Generated code resolves the key template using request fields
4. Serialized response is written to the KV bucket (unless `client_only: true`)
5. Response is sent to the client normally

**Generated client methods:**

- `Get<Method>FromKV(key)` — read a value directly from the KV bucket
- `Put<Method>ToKV(key, value)` — write a value directly to the KV bucket

**Graceful degradation:** If no JetStream context is provided via `WithJetStream()`, KV writes and bucket creation are silently skipped. If the write fails, a warning is logged but the RPC still succeeds.

**Auto-creation:** When `WithJetStream(js)` is provided, buckets are automatically created (or updated) during service registration using `CreateOrUpdateKeyValue` with the configured options. No manual bucket setup required.

### object_store

**Extension:** `natsmicro.object_store`

Same pattern as `kv_store` but uses NATS Object Store (for larger payloads).

```protobuf
rpc GenerateReport(GenerateReportRequest) returns (ReportResponse) {
  option (natsmicro.object_store) = {
    bucket: "reports"
    key_template: "report.{id}"
    ttl: {seconds: 86400}         // Auto-expire after 24 hours (optional)
    description: "Generated reports cache"  // Bucket description (optional)
  };
}
```

**Fields:**

| Field          | Type       | Description                                                                       |
| -------------- | ---------- | --------------------------------------------------------------------------------- |
| `bucket`       | `string`   | **Required.** Name of the Object Store bucket.                                    |
| `key_template` | `string`   | **Required.** Key pattern with `{field}` placeholders.                            |
| `ttl`          | `Duration` | Optional. Auto-expire objects after this duration.                                |
| `description`  | `string`   | Optional. Human-readable bucket description.                                      |
| `client_only`  | `bool`     | Optional. Skip server-side auto-persist; only generate client read/write methods. |

**Generated client methods:**

- `Get<Method>FromObjectStore(key)` — read a value directly from the Object Store bucket
- `Put<Method>ToObjectStore(key, value)` — write a value directly to the Object Store bucket

**When to use KV vs Object Store:**

| Feature           | KV Store              | Object Store                |
| ----------------- | --------------------- | --------------------------- |
| Max value size    | ~1MB (configurable)   | Unlimited (chunked)         |
| Use case          | Small structured data | Large blobs, files, reports |
| History/revisions | Yes                   | Yes                         |
| Watch/notify      | Yes                   | No                          |

## Key Templates

Key templates use `{field}` placeholders that resolve to fields on the RPC request message.

### Syntax

```
user.{id}              → user.123
profile.{user_id}.{region} → profile.abc.us-west
```

Placeholders must reference top-level fields on the input message. Nested fields are not supported.

### Compile-Time Validation

Key templates are **validated at code generation time**. If a placeholder references a field that doesn't exist on the input message, the generator fails with a clear error:

```
key_template "user.{bad_field}" references field {bad_field}
which does not exist on input message SaveProfileRequest
(available fields: [id, name, email, bio])
```

This prevents runtime errors from typos or schema drift.

### Examples

```protobuf
message SaveProfileRequest {
  string id = 1;
  string name = 2;
  string email = 3;
}

// ✅ Valid — {id} exists on SaveProfileRequest
option (natsmicro.kv_store) = {
  bucket: "profiles"
  key_template: "user.{id}"
};

// ✅ Valid — multiple placeholders
option (natsmicro.kv_store) = {
  bucket: "profiles"
  key_template: "{name}.{id}"
};

// ❌ Invalid — {user_id} does not exist
option (natsmicro.kv_store) = {
  bucket: "profiles"
  key_template: "user.{user_id}"
};
```

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

### KV Store & Object Store Service

Auto-persist RPC responses to NATS JetStream:

```protobuf
syntax = "proto3";
package kvstore_demo.v1;
import "natsmicro/options.proto";

service KVStoreDemoService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1.kvdemo"
    name: "kvstore_demo_service"
    version: "1.0.0"
  };

  // Response auto-persisted to KV bucket "user_profiles" with key "user.{id}"
  rpc SaveProfile(SaveProfileRequest) returns (ProfileResponse) {
    option (natsmicro.endpoint) = { timeout: {seconds: 5} };
    option (natsmicro.kv_store) = {
      bucket: "user_profiles"
      key_template: "user.{id}"
    };
  }

  // Standard RPC — no auto-persistence
  rpc GetProfile(GetProfileRequest) returns (ProfileResponse) {
    option (natsmicro.endpoint) = { timeout: {seconds: 5} };
  }

  // Response auto-persisted to Object Store bucket "reports"
  rpc GenerateReport(GenerateReportRequest) returns (ReportResponse) {
    option (natsmicro.endpoint) = { timeout: {seconds: 30} };
    option (natsmicro.object_store) = {
      bucket: "reports"
      key_template: "report.{id}"
    };
  }
}
```

**Server (Go):**

```go
// Enable auto-persistence by providing a JetStream context
svc, err := RegisterKVStoreDemoServiceHandlers(nc, impl,
    WithJetStream(js),  // ← enables KV/ObjectStore auto-persist
)
```

**Client (Go):**

```go
client := NewKVStoreDemoServiceNatsClient(nc,
    WithNatsClientJetStream(js),  // ← enables direct KV/ObjectStore reads
)

// Normal RPC — server auto-persists the response
profile, err := client.SaveProfile(ctx, req)

// Direct KV read — no RPC, reads from NATS KV bucket
cached, err := client.GetSaveProfileFromKV(ctx, "user.123")

// Direct Object Store read — no RPC
report, err := client.GetGenerateReportFromObjectStore(ctx, "report.456")
```

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
func WithJetStream(js jetstream.JetStream) RegisterOption // Enable KV/ObjectStore auto-persist
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
func WithNatsClientJetStream(js jetstream.JetStream) NatsClientOption // Enable KV/ObjectStore reads
```

#### KV Store & Object Store Convenience Methods

```go
// Generated on the client for each method with kv_store option
func (c *Client) Get<MethodName>FromKV(ctx context.Context, key string) (*ResponseType, error)

// Generated on the client for each method with object_store option
func (c *Client) Get<MethodName>FromObjectStore(ctx context.Context, key string) (*ResponseType, error)
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
- [examples/complex-go/](examples/complex-go/) - Interceptors, headers, error handling
- [examples/kvstore-go/](examples/kvstore-go/) - KV Store & Object Store auto-persistence
