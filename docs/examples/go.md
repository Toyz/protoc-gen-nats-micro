# Go Examples

## Complex Service (Interceptors, Headers, Multi-Version)

[Source code →](https://github.com/Toyz/protoc-gen-nats-micro/tree/main/examples/complex-go)

Demonstrates multiple services with interceptors, headers, and API versioning:

- **ProductService** — CRUD with logging + metrics interceptors
- **OrderService v1 & v2** — Side-by-side API versions
- **JSONService / BinaryService** — JSON vs protobuf encoding

```bash
cd examples/complex-go
go mod tidy
go run server.go    # start server
go run client.go    # run client demo
```

## KV Store & Object Store

[Source code →](https://github.com/Toyz/protoc-gen-nats-micro/tree/main/examples/kvstore-go)

Demonstrates auto-persistence of RPC responses to NATS KV and Object stores:

- **SaveProfile** — Persists to KV bucket `user_profiles` with key `user.{id}`
- **GenerateReport** — Persists to Object Store bucket `reports` with key `report.{id}`
- Client reads cached data directly from stores

```bash
cd examples/kvstore-go
go mod tidy
go run cmd/server/server.go    # start server
go run cmd/client/client.go    # run client demo
```

## Streaming RPC

[Source code →](https://github.com/Toyz/protoc-gen-nats-micro/tree/main/examples/streaming-go)

Demonstrates all four RPC patterns:

- **Ping** — Standard unary RPC
- **CountUp** — Server-streaming (server sends 5 numbers)
- **Sum** — Client-streaming (client sends values, server aggregates)
- **Chat** — Bidirectional streaming (echo service)

```bash
cd examples/streaming-go
go mod tidy
go run cmd/server/main.go    # start server (terminal 1)
go run cmd/client/main.go    # run client demo (terminal 2)
```

### Server Implementation

```go
// Server-streaming: emit numbers
func (s *streamService) CountUp(ctx context.Context, req *CountUpRequest, stream *StreamDemoService_CountUp_Stream) error {
    for i := int32(0); i < req.Count; i++ {
        stream.Send(&CountUpResponse{Number: req.Start + i})
        time.Sleep(200 * time.Millisecond)
    }
    return nil
}

// Client-streaming: aggregate
func (s *streamService) Sum(ctx context.Context, stream *StreamDemoService_Sum_Stream) (*SumResponse, error) {
    var total int64
    var count int32
    for {
        msg, err := stream.Recv(ctx)
        if err != nil { break }
        total += msg.Value
        count++
    }
    return &SumResponse{Total: total, Count: count}, nil
}

// Bidi: echo
func (s *streamService) Chat(ctx context.Context, stream *StreamDemoService_Chat_Stream) error {
    for {
        msg, err := stream.Recv(ctx)
        if err != nil { break }
        stream.Send(&ChatMessage{User: "server", Text: "echo: " + msg.Text})
    }
    return nil
}
```
