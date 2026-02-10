# Getting Started

## What is protoc-gen-nats-micro?

A Protocol Buffers compiler plugin that generates type-safe [NATS](https://nats.io) microservice code using the official [`nats.io/micro`](https://github.com/nats-io/nats.go/tree/main/micro) framework.

Write standard `.proto` files, run `buf generate`, and get production-ready NATS microservices with:

- Automatic service discovery and load balancing
- Type-safe request/response handling
- Configurable timeouts, interceptors, and headers
- Streaming RPC (server, client, bidirectional)
- KV Store and Object Store auto-persistence

## Prerequisites

- Go 1.21 or later
- [Buf](https://buf.build/docs/installation) v2
- NATS server ([Docker](https://hub.docker.com/_/nats) or [local install](https://docs.nats.io/running-a-nats-service/introduction/installation))

## Installation

```bash
go install github.com/toyz/protoc-gen-nats-micro/cmd/protoc-gen-nats-micro@latest
```

## Proto Dependencies

Add the natsmicro proto options to your `buf.yaml`:

```yaml
deps:
  - buf.build/helba/natsmicro
```

This lets you use `import "natsmicro/options.proto"` in your proto files without copying anything locally.

## Why not gRPC / nRPC?

|                        | protoc-gen-nats-micro    | gRPC                  | nRPC      |
| ---------------------- | ------------------------ | --------------------- | --------- |
| Service discovery      | Built-in via NATS        | Requires service mesh | Manual    |
| Load balancing         | NATS queue groups        | External LB           | Manual    |
| Streaming              | ✅ Server/Client/Bidi    | ✅ All patterns       | ❌ None   |
| KV/Object auto-persist | ✅                       | ❌                    | ❌        |
| Multi-language         | Go, TS, Python           | Many                  | Go only   |
| Maintenance            | Active                   | Active                | Abandoned |
| Framework              | Official `nats.io/micro` | gRPC                  | Custom    |

## Next Steps

- [Quick Start →](/guide/quick-start) — Build your first service in 5 minutes
- [Streaming RPC →](/guide/streaming) — Server, client, and bidi streaming
- [KV & Object Store →](/guide/kv-object-store) — Auto-persist responses
- [API Reference →](/api/reference) — All proto options
