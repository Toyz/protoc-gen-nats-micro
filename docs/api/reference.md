# API Reference

Complete reference for `protoc-gen-nats-micro` proto extension options.

## Service Options

Service-level configuration using `option (natsmicro.service)`.

| Option           | Type       | Default                    | Description                                  |
| ---------------- | ---------- | -------------------------- | -------------------------------------------- |
| `subject_prefix` | `string`   | Snake-case of service name | NATS subject prefix for all endpoints        |
| `name`           | `string`   | Service name               | Service name for NATS micro registration     |
| `version`        | `string`   | `"1.0.0"`                  | Service version                              |
| `description`    | `string`   | —                          | Human-readable description                   |
| `timeout`        | `Duration` | No timeout                 | Default timeout for all endpoints            |
| `use_json`       | `bool`     | `false`                    | Use JSON encoding instead of binary protobuf |
| `skip`           | `bool`     | `false`                    | Skip NATS code generation for this service   |

```protobuf
service ProductService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1"
    name: "product_service"
    version: "2.0.0"
    description: "Product catalog API"
    timeout: {seconds: 30}
  };
}
```

## Endpoint Options

Per-method configuration using `option (natsmicro.endpoint)`.

| Option     | Type           | Default         | Description                          |
| ---------- | -------------- | --------------- | ------------------------------------ |
| `timeout`  | `Duration`     | Service timeout | Override timeout for this method     |
| `skip`     | `bool`         | `false`         | Skip NATS generation for this method |
| `metadata` | `repeated Map` | —               | Endpoint metadata for discovery      |

```protobuf
rpc CreateProduct(CreateReq) returns (CreateResp) {
  option (natsmicro.endpoint) = {
    timeout: {seconds: 10}
    metadata: { key: "category" value: "write" }
    metadata: { key: "requires_auth" value: "true" }
  };
}

// Skip this endpoint entirely
rpc AdminReset(ResetReq) returns (ResetResp) {
  option (natsmicro.endpoint).skip = true;
}
```

## KV Store Options

Per-method auto-persistence to NATS KV Store using `option (natsmicro.kv_store)`.

| Option         | Type       | Default      | Description                              |
| -------------- | ---------- | ------------ | ---------------------------------------- |
| `bucket`       | `string`   | **Required** | KV bucket name                           |
| `key_template` | `string`   | **Required** | Key template with `{field}` placeholders |
| `description`  | `string`   | —            | Bucket description                       |
| `max_history`  | `int32`    | —            | Max revisions per key                    |
| `ttl`          | `Duration` | —            | Time-to-live for entries                 |

```protobuf
rpc SaveProfile(SaveReq) returns (ProfileResp) {
  option (natsmicro.kv_store) = {
    bucket: "user_profiles"
    key_template: "user.{id}"
    max_history: 5
    ttl: {seconds: 3600}
  };
}
```

## Object Store Options

Per-method auto-persistence to NATS Object Store using `option (natsmicro.object_store)`.

| Option           | Type     | Default      | Description                              |
| ---------------- | -------- | ------------ | ---------------------------------------- |
| `bucket`         | `string` | **Required** | Object store bucket name                 |
| `key_template`   | `string` | **Required** | Key template with `{field}` placeholders |
| `description`    | `string` | —            | Bucket description                       |
| `max_chunk_size` | `int32`  | —            | Max chunk size for large objects         |

```protobuf
rpc GenerateReport(ReportReq) returns (ReportResp) {
  option (natsmicro.object_store) = {
    bucket: "reports"
    key_template: "report.{id}"
  };
}
```

## Key Template Syntax

Key templates extract values from the **request** message to build storage keys:

| Template                          | Request Fields                      | Result         |
| --------------------------------- | ----------------------------------- | -------------- |
| `user.{id}`                       | `id: "abc"`                         | `user.abc`     |
| `{region}.{id}`                   | `region: "us", id: "123"`           | `us.123`       |
| `orders.{customer_id}.{order_id}` | `customer_id: "c1", order_id: "o5"` | `orders.c1.o5` |

Static segments are kept as-is. `{field}` placeholders are replaced with the corresponding request field value.

## Runtime Options

### Server Registration Options

| Option                        | Description                        |
| ----------------------------- | ---------------------------------- |
| `WithName(name)`              | Override service name              |
| `WithVersion(version)`        | Override version                   |
| `WithDescription(desc)`       | Override description               |
| `WithSubjectPrefix(prefix)`   | Override subject prefix            |
| `WithTimeout(duration)`       | Override default timeout           |
| `WithMetadata(map)`           | Replace service metadata           |
| `WithAdditionalMetadata(map)` | Merge into service metadata        |
| `WithServerInterceptor(fn)`   | Add server-side interceptor        |
| `WithJetStream(js)`           | Enable KV/Object Store auto-create |
| `WithStatsHandler(fn)`        | Set stats handler                  |
| `WithDoneHandler(fn)`         | Set done handler                   |
| `WithErrorHandler(fn)`        | Set error handler                  |

### Client Options

| Option                            | Description                  |
| --------------------------------- | ---------------------------- |
| `WithClientSubjectPrefix(prefix)` | Override subject prefix      |
| `WithClientInterceptor(fn)`       | Add client-side interceptor  |
| `WithClientJetStream(js)`         | Enable KV/Object Store reads |

## Timeout Precedence

From highest to lowest priority:

1. **Runtime** — `WithTimeout(60s)` on registration
2. **Endpoint-level** — `option (natsmicro.endpoint) = { timeout: {seconds: 10} }`
3. **Service-level** — `option (natsmicro.service) = { timeout: {seconds: 30} }`
4. **Default** — No timeout (0)

## Proto Import

```protobuf
import "natsmicro/options.proto";
```

All options are defined in [extensions/proto/natsmicro/options.proto](https://github.com/Toyz/protoc-gen-nats-micro/blob/main/extensions/proto/natsmicro/options.proto).
