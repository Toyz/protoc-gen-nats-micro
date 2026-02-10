# Quick Start

Build a working NATS microservice in 5 minutes.

## 1. Define your service

```protobuf
// protos/greeter/v1/service.proto
syntax = "proto3";

package greeter.v1;

import "natsmicro/options.proto";

option go_package = "example/gen/greeter/v1";

service GreeterService {
  option (natsmicro.service) = {
    subject_prefix: "api.v1.greeter"
  };

  rpc SayHello(HelloRequest) returns (HelloResponse) {}
}

message HelloRequest { string name = 1; }
message HelloResponse { string greeting = 1; }
```

## 2. Generate code

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen
    opt: [module=example/gen]
  - local: protoc-gen-nats-micro
    out: gen
    opt: [module=example/gen, language=go]
```

```bash
buf generate --path protos/greeter/v1/service.proto
```

## 3. Implement the server

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"

    "github.com/nats-io/nats.go"
    greeterv1 "example/gen/greeter/v1"
)

type greeterService struct{}

func (s *greeterService) SayHello(ctx context.Context, req *greeterv1.HelloRequest) (*greeterv1.HelloResponse, error) {
    return &greeterv1.HelloResponse{
        Greeting: "Hello, " + req.Name + "!",
    }, nil
}

func main() {
    nc, _ := nats.Connect(nats.DefaultURL)
    defer nc.Close()

    greeterv1.RegisterGreeterServiceHandlers(nc, &greeterService{})

    log.Println("✅ Server running")
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt)
    <-sig
}
```

## 4. Use the client

```go
package main

import (
    "context"
    "fmt"

    "github.com/nats-io/nats.go"
    greeterv1 "example/gen/greeter/v1"
)

func main() {
    nc, _ := nats.Connect(nats.DefaultURL)
    defer nc.Close()

    client := greeterv1.NewGreeterServiceNatsClient(nc)

    resp, err := client.SayHello(context.Background(), &greeterv1.HelloRequest{
        Name: "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Greeting) // "Hello, World!"
}
```

## What happened?

1. **Proto options** defined the NATS subject prefix (`api.v1.greeter`)
2. **`buf generate`** created a type-safe service interface and client
3. **Server** implements the interface and registered with NATS micro
4. **Client** called the RPC method — NATS handled discovery & routing

No service mesh, no load balancer config, no port management. NATS does it all.

## Next Steps

- [Streaming RPC →](/guide/streaming) — Stream responses, upload chunks, real-time chat
- [Interceptors →](/guide/interceptors) — Add logging, auth, and tracing middleware
- [Error Handling →](/guide/error-handling) — Typed errors with codes and messages
