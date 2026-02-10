# TypeScript

`protoc-gen-nats-micro` generates TypeScript code using the [nats.ws](https://github.com/nats-io/nats.ws) or [nats](https://github.com/nats-io/nats.js) client.

## Code Generation

```yaml
# buf.gen.ts.yaml
version: v2
plugins:
  - local: protoc-gen-nats-micro
    out: gen
    opt:
      - language=ts
```

```bash
buf generate --template buf.gen.ts.yaml
```

## Generated Service Interface

```typescript
export interface ProductServiceNats {
  createProduct(
    req: CreateProductRequest,
    headers?: NatsHeaders,
  ): Promise<CreateProductResponse>;
  getProduct(
    req: GetProductRequest,
    headers?: NatsHeaders,
  ): Promise<GetProductResponse>;
}
```

## Server Registration

```typescript
import { connect } from "nats";

const nc = await connect();
await registerProductServiceHandlers(nc, {
  async createProduct(req, headers) {
    return { product: { id: "123", name: req.name } };
  },
  async getProduct(req, headers) {
    return { product: { id: req.id, name: "Widget" } };
  },
});
```

## Client Usage

```typescript
const client = new ProductServiceNatsClient(nc);
const response = await client.createProduct({ name: "Widget", price: 9.99 });
console.log(response.product.id);
```

## Streaming (Server-Streaming)

TypeScript supports server-streaming via `ClientStreamReceiver`:

```typescript
// Service handler
async function countUp(
  req: CountUpRequest,
  sender: ServerStreamSender,
): Promise<void> {
  for (let i = 0; i < req.count; i++) {
    await sender.send(CountUpResponse.encode({ number: req.start + i }));
  }
}

// Client
const receiver = await client.countUp({ start: 1, count: 5 });
for await (const msg of receiver) {
  console.log(msg.number);
}
```

## Options

```typescript
// Custom subject prefix
const client = new ProductServiceNatsClient(nc, {
  subjectPrefix: "staging.api.v1",
});

// Interceptors
const client = new ProductServiceNatsClient(nc, {
  interceptors: [loggingInterceptor],
});
```

::: info
See [TYPESCRIPT.md](https://github.com/Toyz/protoc-gen-nats-micro/blob/main/TYPESCRIPT.md) in the repo for the full TypeScript reference.
:::
