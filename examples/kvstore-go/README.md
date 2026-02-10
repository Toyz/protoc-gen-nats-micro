# KV Store & Object Store Example

This example demonstrates the **KV Store** and **Object Store** features of `protoc-gen-nats-micro`.

## What It Shows

### Server-Side Auto-Persistence

When a handler responds to an RPC with `kv_store` or `object_store` options, the generated code **automatically persists the response** to the NATS JetStream bucket — no extra code needed.

### Client-Side Direct Reads

Clients get **convenience methods** (`GetSaveProfileFromKV`, `GetGenerateReportFromObjectStore`) that read cached data directly from the store without making an RPC call.

### Compile-Time Validation

If a `key_template` references a field that doesn't exist on the input message (e.g., `{nonexistent}`), the code generator **fails with a clear error** at generation time, not at runtime.

## Proto Definition

```protobuf
rpc SaveProfile(SaveProfileRequest) returns (ProfileResponse) {
  option (natsmicro.kv_store) = {
    bucket: "user_profiles"
    key_template: "user.{id}"    // ← {id} must exist on SaveProfileRequest
  };
}

rpc GenerateReport(GenerateReportRequest) returns (ReportResponse) {
  option (natsmicro.object_store) = {
    bucket: "reports"
    key_template: "report.{id}"
  };
}
```

## Prerequisites

- Go 1.21+
- NATS Server with JetStream enabled: `nats-server -js`
- Generated code from the `kvstore_demo/v1` proto

## Running

```bash
# Terminal 1: Start server
go run server.go

# Terminal 2: Run client
go run client.go
```

## Flow

1. **Server registers** handlers with `WithJetStream(js)`
2. Client calls `SaveProfile` → server responds + auto-persists to KV bucket `user_profiles` with key `user.{id}`
3. Client calls `GetSaveProfileFromKV("user.123")` → reads directly from KV, no RPC needed
4. Client calls `GenerateReport` → server responds + auto-persists to Object Store bucket `reports`
5. Client calls `GetGenerateReportFromObjectStore("report.456")` → reads from Object Store

## Key Template Validation

If you try a bad template:

```protobuf
key_template: "user.{bad_field}"  // ← bad_field doesn't exist on SaveProfileRequest
```

The generator will fail with:

```
key_template "user.{bad_field}" references field {bad_field} which does not
exist on input message SaveProfileRequest (available fields: [id, name, email, bio])
```
