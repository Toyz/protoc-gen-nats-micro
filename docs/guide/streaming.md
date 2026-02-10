# Streaming RPC

`protoc-gen-nats-micro` supports all three protobuf streaming patterns over NATS. Streaming methods are automatically detected from your proto definitions — no extra annotations needed.

## Streaming Patterns

| Pattern              | Proto Syntax                                | Description                                           |
| -------------------- | ------------------------------------------- | ----------------------------------------------------- |
| **Server-streaming** | `rpc Foo(Req) returns (stream Resp)`        | Client sends one request, server sends many responses |
| **Client-streaming** | `rpc Foo(stream Req) returns (Resp)`        | Client sends many requests, server responds once      |
| **Bidirectional**    | `rpc Foo(stream Req) returns (stream Resp)` | Both sides send and receive concurrently              |

## Proto Definition

```protobuf
service StreamDemoService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1.stream"
  };

  // Standard unary — business as usual
  rpc Ping(PingRequest) returns (PingResponse) {}

  // Server-streaming — server sends N responses
  rpc CountUp(CountUpRequest) returns (stream CountUpResponse) {}

  // Client-streaming — client sends N requests
  rpc Sum(stream SumRequest) returns (SumResponse) {}

  // Bidi — both sides stream concurrently
  rpc Chat(stream ChatMessage) returns (stream ChatMessage) {}
}
```

## Wire Protocol

Streaming uses NATS pub/sub with custom headers for flow control. No JetStream required.

| Header                   | Direction       | Purpose                                        |
| ------------------------ | --------------- | ---------------------------------------------- |
| `Reply-To`               | Client → Server | Client's inbox for receiving streamed messages |
| `Nats-Stream-Inbox`      | Server → Client | Server's inbox (for client-streaming and bidi) |
| `Nats-Stream-Seq`        | Server → Client | Sequence number for ordered delivery           |
| `Nats-Stream-End`        | Server → Client | `"true"` signals end-of-stream                 |
| `Status` / `Description` | Server → Client | Error info on the end-of-stream message        |

## Server Implementation

### Server-Streaming

The handler receives a typed stream with a `Send()` method:

```go
func (s *myService) CountUp(ctx context.Context, req *CountUpRequest, stream *StreamDemoService_CountUp_Stream) error {
    for i := int32(0); i < req.Count; i++ {
        if err := stream.Send(&CountUpResponse{
            Number: req.Start + i,
        }); err != nil {
            return err
        }
    }
    return nil // stream automatically closed on return
}
```

### Client-Streaming

The handler receives a stream with `Recv()` and returns a final response:

```go
func (s *myService) Sum(ctx context.Context, stream *StreamDemoService_Sum_Stream) (*SumResponse, error) {
    var total int64
    var count int32
    for {
        msg, err := stream.Recv(ctx)
        if err != nil {
            break // EOF or stream ended
        }
        total += msg.Value
        count++
    }
    return &SumResponse{Total: total, Count: count}, nil
}
```

### Bidirectional

The handler gets a stream with both `Send()` and `Recv()`:

```go
func (s *myService) Chat(ctx context.Context, stream *StreamDemoService_Chat_Stream) error {
    for {
        msg, err := stream.Recv(ctx)
        if err != nil {
            break
        }
        stream.Send(&ChatMessage{
            User: "server",
            Text: "echo: " + msg.Text,
        })
    }
    return nil
}
```

## Client Usage

### Server-Streaming

```go
stream, err := client.CountUp(ctx, &CountUpRequest{Start: 1, Count: 5})
if err != nil { /* handle */ }

for {
    resp, err := stream.Recv(ctx)
    if err != nil { break } // EOF
    fmt.Println(resp.Number)
}
stream.Close()
```

### Client-Streaming

```go
stream, err := client.Sum(ctx)
if err != nil { /* handle */ }

stream.Send(&SumRequest{Value: 10})
stream.Send(&SumRequest{Value: 20})
stream.Send(&SumRequest{Value: 30})

result, err := stream.CloseAndRecv(ctx)
fmt.Println(result.Total) // 60
```

### Bidirectional

```go
stream, err := client.Chat(ctx)
if err != nil { /* handle */ }

stream.Send(&ChatMessage{User: "me", Text: "hello"})
reply, _ := stream.Recv(ctx) // "echo: hello"

stream.Send(&ChatMessage{User: "me", Text: "bye"})
reply, _ = stream.Recv(ctx)  // "echo: bye"

stream.CloseSend()
```

## Stream Types Reference

### Server-Side Stream (Send-only)

| Method                            | Description                        |
| --------------------------------- | ---------------------------------- |
| `Send(msg) error`                 | Send a typed message to the client |
| `Close() error`                   | Send end-of-stream marker          |
| `CloseWithError(code, msg) error` | Send error + end-of-stream         |

### Client-Side Stream (Recv-only)

| Method                  | Description                     |
| ----------------------- | ------------------------------- |
| `Recv(ctx) (*T, error)` | Block until next message or EOF |
| `Close() error`         | Unsubscribe from stream         |

### Bidi Stream

| Method                  | Description                 |
| ----------------------- | --------------------------- |
| `Send(msg) error`       | Send to the other side      |
| `Recv(ctx) (*T, error)` | Receive from the other side |
| `CloseSend() error`     | Signal end of sending       |
| `CloseRecv() error`     | Unsubscribe from receiving  |

## Language Support

| Feature                    | Go  | TypeScript | Python |
| -------------------------- | :-: | :--------: | :----: |
| Server-streaming (service) | ✅  |     ✅     |   ✅   |
| Server-streaming (client)  | ✅  |     ✅     |   ✅   |
| Client-streaming           | ✅  |     —      |   —    |
| Bidi-streaming             | ✅  |     —      |   —    |

::: tip
Check out the [streaming-go example](https://github.com/Toyz/protoc-gen-nats-micro/tree/main/examples/streaming-go) for a complete working demo of all four RPC patterns.
:::
