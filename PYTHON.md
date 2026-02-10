# Python Guide

Python-specific usage for `protoc-gen-nats-micro`.

## Setup

### Prerequisites

- Python 3.10+
- `nats-py` package (with micro support)
- `protobuf` package
- `protoc` with Python plugin

### Installation

```bash
pip install nats-py protobuf
```

### Code Generation

```yaml
# buf.gen.py.yaml
version: v2
plugins:
  - protoc_builtin: python
    out: gen
  - local: protoc-gen-nats-micro
    out: gen
    opt: language=python
```

```bash
buf generate --template buf.gen.py.yaml .
```

## Generated Files

For each `.proto` file with services, two files are generated:

| File                 | Contents                                                                              |
| -------------------- | ------------------------------------------------------------------------------------- |
| `*_nats_pb2.py`      | Handler protocol, client class, error types, registration function, service wrapper   |
| `shared_nats_pb2.py` | Shared interceptor types, error codes, registration/client options (once per package) |

## Server Usage

```python
import asyncio
import nats
from gen.product.v1.service_nats_pb2 import (
    ProductServiceHandler,
    register_product_service,
)
from gen.product.v1.shared_nats_pb2 import with_server_interceptor, with_timeout

class ProductServiceImpl:
    async def create_product(self, req, info):
        # info.service, info.method, info.headers available
        # info.set_response_header("X-Custom", "value") for response headers
        return CreateProductResponse(id="123")

async def main():
    nc = await nats.connect()
    impl = ProductServiceImpl()

    wrapper = await register_product_service(
        nc, impl,
        with_server_interceptor(logging_interceptor),
        with_timeout(30.0),
    )

    print(wrapper.endpoints())  # List[EndpointInfo]
    # await wrapper.stop()  # Graceful shutdown

asyncio.run(main())
```

## Client Usage

```python
from gen.product.v1.service_nats_pb2 import ProductServiceClient
from gen.product.v1.shared_nats_pb2 import with_client_interceptor, with_client_subject_prefix

nc = await nats.connect()
client = ProductServiceClient(
    nc,
    with_client_subject_prefix("staging.api.v1"),
    with_client_interceptor(client_logging),
)

response, headers = await client.create_product(request)
```

## Interceptors

### Server Interceptor

```python
from gen.product.v1.shared_nats_pb2 import UnaryServerInterceptor, ServerInfo

async def logging_interceptor(request, info: ServerInfo, handler):
    print(f"[{info.service}] {info.method} called")
    response = await handler(request, info)
    print(f"[{info.service}] {info.method} completed")
    return response
```

### Client Interceptor

```python
async def client_logging(method, request, invoker, headers):
    headers["X-Request-Id"] = str(uuid.uuid4())
    response, resp_headers = await invoker(method, request, headers)
    print(f"{method} returned with headers: {resp_headers}")
    return response, resp_headers
```

## Error Handling

```python
from gen.product.v1.service_nats_pb2 import (
    ProductServiceError,
    is_product_service_not_found,
    new_product_service_not_found_error,
)

# Throw typed errors from server handlers
raise new_product_service_not_found_error("GetProduct", "product not found")

# Check error types on the client
try:
    response, headers = await client.get_product(request)
except ProductServiceError as e:
    if is_product_service_not_found(e):
        # handle not found
        pass
```

## Runtime Options

| Option                               | Description                    |
| ------------------------------------ | ------------------------------ |
| `with_subject_prefix(prefix)`        | Override NATS subject prefix   |
| `with_name(name)`                    | Override service name          |
| `with_version(version)`              | Override service version       |
| `with_description(desc)`             | Override service description   |
| `with_timeout(seconds)`              | Override default timeout       |
| `with_metadata(dict)`                | Override service metadata      |
| `with_additional_metadata(dict)`     | Merge additional metadata      |
| `with_server_interceptor(fn)`        | Add a server interceptor       |
| `with_client_subject_prefix(prefix)` | Override client subject prefix |
| `with_client_interceptor(fn)`        | Add a client interceptor       |

## See Also

- [API.md](API.md) — Proto extension options reference
- [TYPESCRIPT.md](TYPESCRIPT.md) — TypeScript-specific guide
