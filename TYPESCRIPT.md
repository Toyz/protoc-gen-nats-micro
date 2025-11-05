# TypeScript Support for protoc-gen-nats-micro

This guide explains how to use the generated TypeScript code with NATS microservices.

## Prerequisites

Install the required npm packages:

```bash
npm install nats @nats-io/services
```

For protobuf support, you'll also need:

```bash
npm install protobufjs @protobuf-ts/runtime @protobuf-ts/runtime-rpc
```

## Generating TypeScript Code

Generate TypeScript code using buf:

```bash
# Generate only TypeScript
task generate:ts

# Or generate both Go and TypeScript
task generate:all
```

This will create `*_nats.pb.ts` files alongside the protobuf TypeScript files.

## Client Usage

```typescript
import { connect } from 'nats';
import { ProductServiceNatsClient } from './gen/product/v1/service_nats.pb';

async function main() {
  // Connect to NATS
  const nc = await connect({ servers: 'nats://localhost:4222' });

  // Create client
  const client = new ProductServiceNatsClient(nc, {
    subjectPrefix: 'product.v1',
    timeout: 5000, // 5 seconds
  });

  try {
    // Call service method
    const response = await client.getProduct({
      id: '123',
    });

    console.log('Product:', response);
  } catch (error) {
    if (isProductServiceError(error)) {
      console.error(`Error [${error.code}]:`, error.message);
    } else {
      console.error('Unexpected error:', error);
    }
  } finally {
    await nc.close();
  }
}

main();
```

## Server Usage

```typescript
import { connect } from 'nats';
import {
  IProductServiceNats,
  registerProductServiceHandlers,
} from './gen/product/v1/service_nats.pb';
import * as pb from './gen/product/v1/service.pb';

// Implement the service interface
class ProductServiceImpl implements IProductServiceNats {
  async getProduct(request: pb.GetProductRequest): Promise<pb.GetProductResponse> {
    // Your business logic here
    return {
      product: {
        id: request.id,
        name: 'Example Product',
        price: { amount: 1999, currency: 'USD' },
      },
    };
  }

  async listProducts(request: pb.ListProductsRequest): Promise<pb.ListProductsResponse> {
    // Your business logic here
    return {
      products: [],
      nextPageToken: '',
    };
  }

  // Implement other methods...
}

async function main() {
  // Connect to NATS
  const nc = await connect({ servers: 'nats://localhost:4222' });

  // Create service implementation
  const impl = new ProductServiceImpl();

  // Register service handlers
  const service = await registerProductServiceHandlers(nc, impl, {
    name: 'ProductService',
    version: '1.0.0',
    description: 'Product management service',
    subjectPrefix: 'product.v1',
    timeout: 30000, // 30 seconds
  });

  console.log('Service started:', service.info());
  console.log('Endpoints:', service.endpoints());

  // Keep the service running
  process.on('SIGINT', async () => {
    console.log('Shutting down...');
    await service.stop();
    await nc.close();
    process.exit(0);
  });
}

main();
```

## Error Handling

The generated code includes typed error handling:

```typescript
import {
  ProductServiceError,
  isProductServiceNotFound,
  isProductServiceInvalidArgument,
  newProductServiceNotFoundError,
} from './gen/product/v1/service_nats.pb';

// In service implementation
async getProduct(request: pb.GetProductRequest): Promise<pb.GetProductResponse> {
  if (!request.id) {
    throw newProductServiceInvalidArgumentError('GetProduct', 'Product ID is required');
  }

  const product = await this.findProduct(request.id);
  if (!product) {
    throw newProductServiceNotFoundError('GetProduct', `Product ${request.id} not found`);
  }

  return { product };
}

// In client code
try {
  const response = await client.getProduct({ id: '123' });
} catch (error) {
  if (isProductServiceNotFound(error)) {
    console.log('Product not found');
  } else if (isProductServiceInvalidArgument(error)) {
    console.log('Invalid request:', error.message);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## Configuration Options

### Client Options

```typescript
interface ProductServiceClientOptions {
  subjectPrefix?: string;  // Override subject prefix
  timeout?: number;        // Default timeout in milliseconds
}
```

### Service Registration Options

```typescript
interface ProductServiceRegisterOptions {
  name?: string;                    // Service name
  version?: string;                 // Service version
  description?: string;             // Service description
  subjectPrefix?: string;           // Subject prefix
  timeout?: number;                 // Request timeout in milliseconds
  metadata?: Record<string, string>; // Additional metadata
}
```

## Service Discovery

Get information about service endpoints:

```typescript
// From client
const endpoints = client.endpoints();
console.log('Available endpoints:', endpoints);
// [{ name: 'GetProduct', subject: 'product.v1.get_product' }, ...]

// From service
const endpoints = service.endpoints();
console.log('Registered endpoints:', endpoints);
```

## Integration with protobuf-ts

The generated code works with protobuf-ts generated types. Make sure your `buf.gen.ts.yaml` includes:

```yaml
plugins:
  - remote: buf.build/community/timostamm-protobuf-ts
    out: gen
    opt:
      - generate_dependencies
      - long_type_string
```

## Best Practices

1. **Type Safety**: Use TypeScript's strict mode for better type safety
2. **Error Handling**: Always handle errors appropriately using the typed error helpers
3. **Timeouts**: Set appropriate timeouts for both client and service operations
4. **Connection Management**: Reuse NATS connections when possible
5. **Graceful Shutdown**: Always stop services and close connections properly

## Example Project Structure

```
my-project/
├── src/
│   ├── client.ts        # Client implementation
│   ├── server.ts        # Server implementation
│   └── gen/             # Generated code
│       └── product/
│           └── v1/
│               ├── service.pb.ts       # protobuf-ts generated
│               └── service_nats.pb.ts  # NATS micro generated
├── proto/               # Protocol buffer definitions
├── buf.gen.ts.yaml     # Buf TypeScript configuration
├── package.json
└── tsconfig.json
```

## Troubleshooting

### Import Errors

Make sure the generated TypeScript files can find the protobuf definitions:

```typescript
// If you get import errors, adjust the import path in the generated files
// or configure your TypeScript paths in tsconfig.json
{
  "compilerOptions": {
    "paths": {
      "@gen/*": ["./gen/*"]
    }
  }
}
```

### NATS Connection Issues

```typescript
// Use connection options for debugging
const nc = await connect({
  servers: 'nats://localhost:4222',
  debug: true,
  verbose: true,
});
```
