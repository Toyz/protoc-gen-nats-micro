#!/usr/bin/env bun
/**
 * Simple TypeScript NATS Microservice Client Example
 * 
 * This example demonstrates:
 * - Creating a NATS client for the product service
 * - Using client-side interceptors for request logging
 * - Making service calls
 * 
 * Run: bun run client.ts
 */

import { connect, headers } from 'nats';
import * as pb from './gen/product/v1/service';
import type { UnaryClientInterceptor } from './gen/product/v1/service/shared_nats.pb';
import { ProductServiceNatsClient } from './gen/product/v1/service_nats.pb';

// Client logging interceptor - demonstrates adding headers and reading response headers
const clientLoggingInterceptor: UnaryClientInterceptor = async (method, request, reply, invoker, hdrs, responseHeaders) => {
  console.log(`→ [Client] Calling ${method}`);
  const start = Date.now();

  // Add custom headers for tracing and metadata
  const customHeaders = hdrs || headers();
  customHeaders.set('X-Trace-Id', `trace-${Date.now()}`);
  customHeaders.set('X-Client-Version', '1.0.0');

  try {
    await invoker(method, request, reply, customHeaders, responseHeaders);

    // Read response headers from server
    if (responseHeaders?.value) {
      const serverVer = responseHeaders.value.get('X-Server-Version');
      if (serverVer) {
        console.log(`  [Response Headers] Server-Version: ${serverVer}`);
      }
    }

    const duration = Date.now() - start;
    console.log(`✓ [Client] ${method} completed in ${duration}ms`);
  } catch (error) {
    const duration = Date.now() - start;
    console.error(`✗ [Client] ${method} failed after ${duration}ms:`, error);
    throw error;
  }
}; async function main() {
  // Connect to NATS
  const nc = await connect({ servers: 'nats://localhost:4222' });
  console.log('✓ Connected to NATS\n');

  // Create client with interceptor
  const client = new ProductServiceNatsClient(nc, {
    clientInterceptors: [clientLoggingInterceptor],
  });

  console.log('📡 ProductService Client Endpoints:');
  for (const ep of client.endpoints()) {
    console.log(`  • ${ep.name} → ${ep.subject}`);
  }
  console.log('');

  try {
    // Create a product
    console.log('→ Creating product...');
    const createResp = await client.createProduct({
      name: 'Wireless Headphones',
      description: 'Premium noise-cancelling wireless headphones',
      sku: 'HEADPHONES-001',
      category: pb.ProductCategory.CATEGORY_ELECTRONICS,
      price: {
        currencyCode: 'USD',
        units: '299',
        nanos: 990000000,
      },
      stockQuantity: 50,
      imageUrls: ['https://example.com/headphones.jpg'],
      attributes: {
        color: 'black',
        bluetooth: '5.0',
      },
    });

    const product = createResp.product!;
    console.log(`✓ Created product:`);
    console.log(`  ID:       ${product.id}`);
    console.log(`  Name:     ${product.name}`);
    console.log(`  Price:    $${product.price!.units}.${String(product.price!.nanos).padStart(9, '0').slice(0, 2)} ${product.price!.currencyCode}`);
    console.log(`  Category: ${pb.ProductCategory[product.category]}`);
    console.log(`  Stock:    ${product.stockQuantity} units\n`);

    // Get the product
    console.log('→ Fetching product...');
    const getResp = await client.getProduct({ id: product.id });
    console.log(`✓ Retrieved product: ${getResp.product!.name}\n`);

    // Search products
    console.log('→ Searching products...');
    const searchResp = await client.searchProducts({
      category: pb.ProductCategory.CATEGORY_ELECTRONICS,
      query: '',
      minPrice: undefined,
      maxPrice: undefined,
      pageSize: 10,
      pageToken: '',
    });
    console.log(`✓ Found ${searchResp.totalCount} products`);
    for (const p of searchResp.products) {
      console.log(`  - ${p.name}: $${p.price!.units}.${String(p.price!.nanos).padStart(9, '0').slice(0, 2)}`);
    }
    console.log('');

    // Update product
    console.log('→ Updating product...');
    const updateResp = await client.updateProduct({
      id: product.id,
      name: 'Wireless Headphones Pro',
      description: 'Premium noise-cancelling wireless headphones with extended battery',
      price: {
        currencyCode: 'USD',
        units: '349',
        nanos: 990000000,
      },
      stockQuantity: 45,
      imageUrls: ['https://example.com/headphones-pro.jpg'],
      attributes: {
        color: 'black',
        bluetooth: '5.3',
        battery: '40hrs',
      },
    });
    console.log(`✓ Updated product: ${updateResp.product!.name}\n`);

    // Delete product
    console.log('→ Deleting product...');
    const deleteResp = await client.deleteProduct({ id: product.id });
    console.log(`✓ Deleted product: ${deleteResp.success}\n`);

    console.log('✅ All operations completed successfully!');
  } catch (error) {
    console.error('\n❌ Error:', error);
  } finally {
    await nc.close();
    console.log('\n✓ Disconnected from NATS');
  }
}

main().catch((err) => {
  console.error('Failed to run client:', err);
  process.exit(1);
});
