# KV Store & Object Store

Automatically persist RPC responses to NATS KV Store or Object Store after every successful call. Clients can then read cached data directly — bypassing the RPC entirely.

## How It Works

1. Client calls an RPC method normally
2. Server processes the request and responds
3. **After responding**, the generated handler automatically writes the response to the configured store
4. Clients can later read from the store directly via generated helper methods

No extra code needed — just add the proto annotations.

## KV Store

### Proto Definition

```protobuf
rpc SaveProfile(SaveProfileRequest) returns (ProfileResponse) {
  option (natsmicro.kv_store) = {
    bucket: "user_profiles"
    key_template: "user.{id}"
  };
}
```

### Key Templates

Key templates extract values from the **request** message to build the storage key:

| Template                          | Request Field                       | Result         |
| --------------------------------- | ----------------------------------- | -------------- |
| `user.{id}`                       | `id: "abc"`                         | `user.abc`     |
| `{region}.{id}`                   | `region: "us", id: "123"`           | `us.123`       |
| `orders.{customer_id}.{order_id}` | `customer_id: "c1", order_id: "o5"` | `orders.c1.o5` |

### KV Store Options

| Option         | Type       | Description                                             |
| -------------- | ---------- | ------------------------------------------------------- |
| `bucket`       | `string`   | KV bucket name (auto-created if JetStream is available) |
| `key_template` | `string`   | Template for building the key from request fields       |
| `description`  | `string`   | Bucket description                                      |
| `max_history`  | `int32`    | Max revisions per key                                   |
| `ttl`          | `Duration` | Time-to-live for entries                                |

### Generated Methods

```go
// Server-side: auto-persists after responding (generated handler code)
// No extra implementation needed

// Client-side: read directly from KV store
profile, err := client.GetSaveProfileFromKV("user.abc")

// Client-side: write directly to KV store
err := client.PutSaveProfileToKV("user.abc", profileResponse)
```

## Object Store

For larger payloads (reports, files, binary data), use Object Store:

```protobuf
rpc GenerateReport(GenerateReportRequest) returns (ReportResponse) {
  option (natsmicro.object_store) = {
    bucket: "reports"
    key_template: "report.{id}"
  };
}
```

### Object Store Options

| Option           | Type     | Description                             |
| ---------------- | -------- | --------------------------------------- |
| `bucket`         | `string` | Object store bucket name (auto-created) |
| `key_template`   | `string` | Template for building the key           |
| `description`    | `string` | Bucket description                      |
| `max_chunk_size` | `int32`  | Max chunk size for large objects        |

### Generated Methods

```go
// Read from Object Store
report, err := client.GetGenerateReportFromObjectStore("report.monthly")

// Write to Object Store
err := client.PutGenerateReportToObjectStore("report.monthly", reportResponse)
```

## JetStream Configuration

KV and Object Store require JetStream. Pass a JetStream context during registration:

### Server

```go
js, _ := jetstream.New(nc)
svc, err := RegisterMyServiceHandlers(nc, impl, WithJetStream(js))
```

### Client

```go
js, _ := jetstream.New(nc)
client := NewMyServiceNatsClient(nc, WithClientJetStream(js))

// Now you can use the KV/Object Store read methods
profile, err := client.GetSaveProfileFromKV("user.abc")
```

::: warning
Without JetStream, KV/Object Store methods will return a runtime error. The RPC methods themselves still work fine — only the auto-persistence and direct store reads require JetStream.
:::

::: tip
Check out the [kvstore-go example](https://github.com/Toyz/protoc-gen-nats-micro/tree/main/examples/kvstore-go) for a complete working demo.
:::
