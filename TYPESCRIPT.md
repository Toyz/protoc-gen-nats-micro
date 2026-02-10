# TypeScript Guide

TypeScript-specific usage for `protoc-gen-nats-micro`.

## Setup

### Prerequisites

- Node.js 18+
- `nats` npm package
- `@nats-io/services` npm package
- `protoc-gen-ts` (for protobuf TypeScript generation)

### Installation

```bash
npm install nats @nats-io/services
```

### Code Generation

```yaml
# buf.gen.ts.yaml
version: v2
plugins:
  - local: protoc-gen-ts
    out: gen
    opt: output_javascript
  - local: protoc-gen-nats-micro
    out: gen
    opt: language=typescript
```

```bash
buf generate --template buf.gen.ts.yaml .
```

## Generated Files

For each `.proto` file with services, two files are generated:

| File                | Contents                                                                   |
| ------------------- | -------------------------------------------------------------------------- |
| `*_nats.pb.ts`      | Service interface, client class, error types, registration function        |
| `shared_nats.pb.ts` | Shared interceptor types, header utilities, error codes (once per package) |

## Server Usage

```typescript
import { connect } from "nats";
import {
  IProductServiceNats,
  registerProductService,
} from "./gen/product/v1/service_nats.pb";

class ProductServiceImpl implements IProductServiceNats {
  async createProduct(
    req: CreateProductRequest,
  ): Promise<CreateProductResponse> {
    // implementation
  }
}

const nc = await connect();
const impl = new ProductServiceImpl();
const service = await registerProductService(nc, impl, {
  interceptors: [loggingInterceptor],
});
```

## Client Usage

```typescript
import { connect } from "nats";
import { ProductServiceNatsClient } from "./gen/product/v1/service_nats.pb";

const nc = await connect();
const client = new ProductServiceNatsClient(nc, {
  interceptors: [clientLoggingInterceptor],
});

const response = await client.createProduct(request);
```

## Interceptors

### Server Interceptor

```typescript
import {
  UnaryServerInterceptor,
  UnaryServerInfo,
} from "./gen/product/v1/shared_nats.pb";

const loggingInterceptor: UnaryServerInterceptor = async (
  request,
  info,
  handler,
) => {
  console.log(`[${info.service}] ${info.method} called`);
  const start = Date.now();
  const response = await handler(request);
  console.log(`[${info.service}] ${info.method} took ${Date.now() - start}ms`);
  return response;
};
```

### Client Interceptor

```typescript
import { UnaryClientInterceptor } from "./gen/product/v1/shared_nats.pb";

const clientInterceptor: UnaryClientInterceptor = async (
  method,
  request,
  reply,
  invoker,
  headers,
) => {
  headers?.set("X-Request-Id", crypto.randomUUID());
  await invoker(method, request, reply, headers);
};
```

## Error Handling

```typescript
import {
  ProductServiceError,
  isProductServiceNotFound,
  newProductServiceNotFoundError,
} from "./gen/product/v1/service_nats.pb";

// Throw typed errors from server handlers
throw newProductServiceNotFoundError("GetProduct", "product not found");

// Check error types on the client
try {
  await client.getProduct(request);
} catch (err) {
  if (isProductServiceNotFound(err)) {
    // handle not found
  }
}
```

## See Also

- [API.md](API.md) — Proto extension options reference
- [PYTHON.md](PYTHON.md) — Python-specific guide
