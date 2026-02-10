# Interceptors & Headers

Interceptors provide middleware for cross-cutting concerns like logging, authentication, metrics, and tracing. Headers enable bidirectional metadata propagation.

## Server Interceptors

Interceptors wrap every handler call. They see the full request context and can modify the response.

```go
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *productv1.UnaryServerInfo,
    handler productv1.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    log.Printf("→ %s.%s", info.Service, info.Method)

    resp, err := handler(ctx, req)

    log.Printf("← %s.%s (%v)", info.Service, info.Method, time.Since(start))
    return resp, err
}
```

### Registration

```go
svc, err := RegisterProductServiceHandlers(nc, impl,
    WithServerInterceptor(loggingInterceptor),
    WithServerInterceptor(metricsInterceptor),
    WithServerInterceptor(authInterceptor),
)
```

Interceptors execute in order: `logging → metrics → auth → handler → auth → metrics → logging`

## Client Interceptors

Same pattern on the client side:

```go
func clientTraceInterceptor(
    ctx context.Context,
    method string,
    req, reply interface{},
    invoker UnaryInvoker,
) error {
    // Add trace header before sending
    ctx = WithOutgoingHeaders(ctx, nats.Header{
        "X-Trace-Id": []string{uuid.New().String()},
    })
    return invoker(ctx, method, req, reply)
}
```

### Registration

```go
client := NewProductServiceNatsClient(nc,
    WithClientInterceptor(clientTraceInterceptor),
)
```

## Headers

### Reading Request Headers (Server)

```go
func myHandler(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    headers := IncomingHeaders(ctx)
    traceID := headers.Get("X-Trace-Id")
    // ...
}
```

### Setting Response Headers (Server)

```go
func myInterceptor(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
    responseHeaders := nats.Header{}
    responseHeaders.Set("X-Server-Version", "1.0.0")
    SetResponseHeaders(ctx, responseHeaders)

    return handler(ctx, req)
}
```

### Setting Request Headers (Client)

```go
ctx := WithOutgoingHeaders(context.Background(), nats.Header{
    "X-Client-Version": []string{"2.0.0"},
    "Authorization":    []string{"Bearer " + token},
})
resp, err := client.GetProduct(ctx, req)
```

### Reading Response Headers (Client)

```go
ctx := context.Background()
resp, err := client.GetProduct(ctx, req)
// Response headers are available via the context after the call
headers := ResponseHeaders(ctx)
serverVersion := headers.Get("X-Server-Version")
```

## Common Patterns

### Authentication

```go
func authInterceptor(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
    headers := IncomingHeaders(ctx)
    token := headers.Get("Authorization")
    if token == "" {
        return nil, NewMyServiceUnauthenticatedError(info.Method, "missing auth token")
    }
    // Verify token...
    return handler(ctx, req)
}
```

### Request Timing

```go
func timingInterceptor(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
    start := time.Now()
    resp, err := handler(ctx, req)
    duration := time.Since(start)

    responseHeaders := nats.Header{}
    responseHeaders.Set("X-Duration-Ms", fmt.Sprintf("%d", duration.Milliseconds()))
    SetResponseHeaders(ctx, responseHeaders)

    return resp, err
}
```
