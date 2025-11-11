# Simple TypeScript NATS Microservice Example

A minimal example demonstrating NATS microservices with TypeScript and interceptors.

## Features

- Simple product service implementation
- Server-side interceptors (logging, metrics)
- Client-side interceptors (request timing)
- Full TypeScript type safety
- Zero build step with Bun

## Prerequisites

- [Bun](https://bun.sh) v1.0+
- NATS server running on `localhost:4222`

## Quick Start

### 1. Start NATS Server

```bash
# Using Docker
docker run -p 4222:4222 nats

# Or install and run locally
# See: https://docs.nats.io/running-a-nats-service/introduction/installation
```

### 2. Install Dependencies

```bash
bun install
```

### 3. Run Server

```bash
bun run server
```

You should see:
```
Connected to NATS
Registered ProductService (with logging + metrics interceptors)

Service Endpoints:
  - create_product -> api.v1.create_product
  - get_product -> api.v1.get_product
  ...

Server ready! Waiting for requests...
```

### 4. Run Client (in another terminal)

```bash
bun run client
```

The client will:
- Create a product
- Fetch the product
- Search products
- Update the product
- Delete the product

## Code Structure

```
simple-ts/
├── server.ts        # Service implementation with interceptors
├── client.ts        # Client with request logging interceptor
├── package.json     # Dependencies and scripts
└── README.md        # This file
```

## Interceptor Examples

### Server Interceptors

**Logging Interceptor** - Logs request start/completion:
```typescript
const loggingInterceptor: UnaryServerInterceptor = async (request, info, handler) => {
  console.log(`[${info.service}.${info.method}] Request started`);
  const response = await handler(request);
  console.log(`[${info.service}.${info.method}] Request completed`);
  return response;
};
```

**Metrics Interceptor** - Tracks timing and status:
```typescript
const metricsInterceptor: UnaryServerInterceptor = async (request, info, handler) => {
  const start = Date.now();
  const response = await handler(request);
  const duration = Date.now() - start;
  console.log(`[METRICS] method=${info.method} status=success duration_ms=${duration}`);
  return response;
};
```

### Client Interceptor

**Request Timing** - Logs client-side request timing:
```typescript
const clientLoggingInterceptor: UnaryClientInterceptor = async (method, request, reply, invoker) => {
  const start = Date.now();
  await invoker(method, request, reply);
  console.log(`[Client] ${method} completed in ${Date.now() - start}ms`);
};
```

## Using the Generated Code

The service and client use generated TypeScript code from protocol buffers:

```typescript
// Import generated types (protobuf messages)
import * as pb from '../../gen/product/v1/service';
import { Status } from '../../gen/common/types/v1/status';

// Import NATS service registration and client
import { registerProductServiceHandlers, ProductServiceNatsClient } from '../../gen/product/v1/service_nats';

// Import interceptor types from shared file
import type { UnaryServerInterceptor, UnaryClientInterceptor, IProductServiceNats } from '../../gen/product/v1/service/shared_nats';
```

## Regenerating Code

If you modify the `.proto` files, regenerate the TypeScript code:

```bash
# From project root
task generate:ts
# or
buf generate --template examples/buf-configs/buf.gen.ts.yaml
```

## Next Steps

- Add authentication interceptor
- Add retry logic in client interceptor
- Add distributed tracing
- Add circuit breaker pattern
- Add rate limiting

## Learn More

- [NATS Docs](https://docs.nats.io)
- [Bun Docs](https://bun.sh/docs)
- [Main Project README](../../README.md)
