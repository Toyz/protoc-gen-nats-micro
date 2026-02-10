# Python

`protoc-gen-nats-micro` generates Python code using the [nats-py](https://github.com/nats-io/nats.py) async client.

## Code Generation

```yaml
# buf.gen.py.yaml
version: v2
plugins:
  - local: protoc-gen-nats-micro
    out: gen
    opt:
      - language=python
```

```bash
buf generate --template buf.gen.py.yaml
```

## Generated Service Interface

```python
class ProductServiceNats:
    async def create_product(self, req: CreateProductRequest) -> CreateProductResponse:
        ...
    async def get_product(self, req: GetProductRequest) -> GetProductResponse:
        ...
```

## Server Registration

```python
import nats

nc = await nats.connect()

class MyProductService:
    async def create_product(self, req, headers=None):
        return CreateProductResponse(product=Product(id="123", name=req.name))

    async def get_product(self, req, headers=None):
        return GetProductResponse(product=Product(id=req.id, name="Widget"))

await register_product_service_handlers(nc, MyProductService())
```

## Client Usage

```python
client = ProductServiceClient(nc)
response, headers = await client.create_product(
    CreateProductRequest(name="Widget")
)
print(response.product.id)
```

## Streaming (Server-Streaming)

Python supports server-streaming via an async iterator:

```python
# Service handler
async def count_up(req, sender):
    for i in range(req.count):
        await sender.send(CountUpResponse(number=req.start + i))

# Client
receiver = await client.count_up(CountUpRequest(start=1, count=5))
async for msg in receiver:
    print(msg.number)
await receiver.close()
```

## Options

```python
# Custom subject prefix
client = ProductServiceClient(nc, WithClientSubjectPrefix("staging.api.v1"))

# Interceptors
client = ProductServiceClient(nc, WithClientInterceptor(logging_interceptor))
```

::: info
See [PYTHON.md](https://github.com/Toyz/protoc-gen-nats-micro/blob/main/PYTHON.md) in the repo for the full Python reference.
:::
