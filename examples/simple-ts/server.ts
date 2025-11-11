#!/usr/bin/env bun
/**
 * Simple TypeScript NATS Microservice Server Example
 * 
 * This example demonstrates:
 * - Creating a simple product service
 * - Using server-side interceptors for logging and metrics
 * - Service registration with NATS
 * 
 * Run: bun run server.ts
 */

import { connect } from 'nats';
import { Status } from '../../gen/common/types/v1/status';
import * as pb from '../../gen/product/v1/service';
import type { UnaryServerInterceptor } from '../../gen/product/v1/service/shared_nats.pb';
import type { IProductServiceNats } from '../../gen/product/v1/service_nats.pb';
import { registerProductServiceHandlers } from '../../gen/product/v1/service_nats.pb';

// Simple in-memory product store
const products = new Map<string, pb.Product>();
let idCounter = 1;

// Product service implementation
const productService: IProductServiceNats = {
  async createProduct(request: pb.CreateProductRequest): Promise<pb.CreateProductResponse> {
    const id = `prod-${idCounter++}`;
    const now = Date.now();

    const product: pb.Product = {
      id,
      name: request.name,
      description: request.description,
      sku: request.sku,
      category: request.category,
      price: request.price,
      stockQuantity: request.stockQuantity,
      imageUrls: request.imageUrls,
      attributes: request.attributes,
      status: Status.ACTIVE,
      metadata: {
        createdAt: { seconds: String(Math.floor(now / 1000)), nanos: (now % 1000) * 1000000 },
        updatedAt: { seconds: String(Math.floor(now / 1000)), nanos: (now % 1000) * 1000000 },
        createdBy: 'system',
        updatedBy: 'system',
        tags: {},
      },
    };

    products.set(id, product);
    console.log(`✓ Created product: ${request.name} (${id})`);

    return { product };
  },

  async getProduct(request: pb.GetProductRequest): Promise<pb.GetProductResponse> {
    const product = products.get(request.id);
    if (!product) {
      throw new Error(`Product not found: ${request.id}`);
    }

    console.log(`✓ Retrieved product: ${product.name}`);
    return { product };
  },

  async updateProduct(request: pb.UpdateProductRequest): Promise<pb.UpdateProductResponse> {
    const product = products.get(request.id);
    if (!product) {
      throw new Error(`Product not found: ${request.id}`);
    }

    product.name = request.name;
    product.description = request.description;
    product.price = request.price;
    product.stockQuantity = request.stockQuantity;
    product.imageUrls = request.imageUrls;
    product.attributes = request.attributes;
    const updateTime = Date.now();
    product.metadata!.updatedAt = {
      seconds: String(Math.floor(updateTime / 1000)),
      nanos: (updateTime % 1000) * 1000000
    };

    console.log(`✓ Updated product: ${product.name}`);
    return { product };
  },

  async deleteProduct(request: pb.DeleteProductRequest): Promise<pb.DeleteProductResponse> {
    const deleted = products.delete(request.id);
    console.log(`✓ Deleted product: ${request.id}`);
    return { success: deleted };
  },

  async searchProducts(request: pb.SearchProductsRequest): Promise<pb.SearchProductsResponse> {
    const results: pb.Product[] = [];

    for (const product of products.values()) {
      if (request.category && product.category !== request.category) {
        continue;
      }
      if (request.query && !product.name.toLowerCase().includes(request.query.toLowerCase())) {
        continue;
      }
      results.push(product);
    }

    console.log(`✓ Search returned ${results.length} products`);
    return {
      products: results,
      nextPageToken: '',
      totalCount: results.length,
    };
  },
};

// Logging interceptor - also demonstrates reading headers
const loggingInterceptor: UnaryServerInterceptor = async (request, info, handler) => {
  const start = Date.now();
  console.log(`→ [${info.service}.${info.method}] Request started`);

  // Read incoming headers if present
  if (info.headers) {
    const traceId = info.headers.get('X-Trace-Id');
    const clientVersion = info.headers.get('X-Client-Version');
    if (traceId) {
      console.log(`  [Headers] Trace-ID: ${traceId}`);
    }
    if (clientVersion) {
      console.log(`  [Headers] Client-Version: ${clientVersion}`);
    }
  }

  try {
    const response = await handler(request);
    const duration = Date.now() - start;
    console.log(`✓ [${info.service}.${info.method}] Request completed in ${duration}ms`);
    return response;
  } catch (error) {
    const duration = Date.now() - start;
    console.error(`✗ [${info.service}.${info.method}] Request failed after ${duration}ms:`, error);
    throw error;
  }
};

// Metrics interceptor
const metricsInterceptor: UnaryServerInterceptor = async (request, info, handler) => {
  const start = Date.now();

  try {
    const response = await handler(request);
    const duration = Date.now() - start;
    console.log(`📊 [METRICS] service=${info.service} method=${info.method} status=success duration_ms=${duration}`);
    return response;
  } catch (error) {
    const duration = Date.now() - start;
    console.log(`📊 [METRICS] service=${info.service} method=${info.method} status=error duration_ms=${duration}`);
    throw error;
  }
};

// Main
async function main() {
  // Connect to NATS
  const nc = await connect({ servers: 'nats://localhost:4222' });
  console.log('✓ Connected to NATS');

  // Register service with interceptors
  const service = await registerProductServiceHandlers(nc, productService, {
    serverInterceptors: [loggingInterceptor, metricsInterceptor],
  });

  console.log('✓ Registered ProductService (with logging + metrics interceptors)');
  console.log('\n📡 Service Endpoints:');
  for (const ep of service.endpoints()) {
    console.log(`  • ${ep.name} → ${ep.subject}`);
  }

  console.log('\n🚀 Server ready! Waiting for requests...\n');

  // Handle shutdown
  process.on('SIGINT', async () => {
    console.log('\n\n✓ Shutting down...');
    await service.stop();
    await nc.close();
    process.exit(0);
  });
}

main().catch((err) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
