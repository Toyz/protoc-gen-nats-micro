# Error Handling

`protoc-gen-nats-micro` generates structured, typed errors for every service — making it easy to return specific error codes and check for them on the client side.

## Error Codes

Standard error codes are generated for each service:

| Code  | Constant                  | Use Case            |
| ----- | ------------------------- | ------------------- |
| `400` | `ErrCodeInvalidArgument`  | Bad request data    |
| `404` | `ErrCodeNotFound`         | Resource not found  |
| `409` | `ErrCodeAlreadyExists`    | Duplicate resource  |
| `403` | `ErrCodePermissionDenied` | Not authorized      |
| `401` | `ErrCodeUnauthenticated`  | Missing credentials |
| `500` | `ErrCodeInternal`         | Server error        |
| `503` | `ErrCodeUnavailable`      | Service down        |

## Returning Errors (Server)

Use the generated error constructors in your handler:

```go
func (s *myService) GetProduct(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
    if req.Id == "" {
        return nil, NewProductServiceInvalidArgumentError("GetProduct", "id is required")
    }

    product, ok := s.products[req.Id]
    if !ok {
        return nil, NewProductServiceNotFoundError("GetProduct", "product not found: " + req.Id)
    }

    return &GetProductResponse{Product: product}, nil
}
```

## Checking Errors (Client)

Use the generated type-check helpers:

```go
resp, err := client.GetProduct(ctx, &GetProductRequest{Id: "abc"})
if err != nil {
    if IsProductServiceNotFound(err) {
        log.Println("Product doesn't exist")
    } else if IsProductServiceInvalidArgument(err) {
        log.Println("Bad request:", err)
    } else {
        log.Println("Unexpected error:", err)
    }
    return
}
```

## Extracting Error Codes

```go
code := GetProductServiceErrorCode(err)
switch code {
case ProductServiceErrCodeNotFound:
    // handle not found
case ProductServiceErrCodeInternal:
    // handle internal error
}
```

## Custom Error Data

Errors can carry binary payload data for rich error details:

```go
type myError struct {
    code    string
    message string
    details []byte
}

func (e *myError) NatsErrorCode() string    { return e.code }
func (e *myError) NatsErrorMessage() string { return e.message }
func (e *myError) NatsErrorData() []byte    { return e.details }
```

The generated handler code automatically checks for these interfaces, so any error implementing `NatsErrorCode()`, `NatsErrorMessage()`, and `NatsErrorData()` will have its values sent to the client.

## Wire Format

Errors are transmitted using standard NATS micro error headers:

| Header        | Value                      |
| ------------- | -------------------------- |
| `Status`      | Error code (e.g., `"404"`) |
| `Description` | Human-readable message     |
| Body          | Optional binary error data |
