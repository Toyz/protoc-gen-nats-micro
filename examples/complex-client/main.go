package main

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	locationv1 "github.com/toyz/protoc-gen-nats-micro/gen/common/location/v1"
	typesv1 "github.com/toyz/protoc-gen-nats-micro/gen/common/types/v1"
	orderv1 "github.com/toyz/protoc-gen-nats-micro/gen/order/v1"
	orderv2 "github.com/toyz/protoc-gen-nats-micro/gen/order/v2"
	productv1 "github.com/toyz/protoc-gen-nats-micro/gen/product/v1"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	log.Println("✓ Connected to NATS")

	// Create clients (subject prefixes read from proto!)
	productClient := productv1.NewProductServiceNatsClient(nc)
	orderClient := orderv1.NewOrderServiceNatsClient(nc)

	// Print client endpoints
	log.Println("\n📡 ProductService Client Endpoints:")
	for _, ep := range productClient.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n📡 OrderService Client Endpoints:")
	for _, ep := range orderClient.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a product
	log.Println("\n→ Creating product...")
	createProdResp, err := productClient.CreateProduct(ctx, &productv1.CreateProductRequest{
		Name:        "Wireless Headphones",
		Description: "Premium noise-cancelling wireless headphones",
		Sku:         "HEADPHONES-001",
		Category:    productv1.ProductCategory_CATEGORY_ELECTRONICS,
		Price: &typesv1.Money{
			CurrencyCode: "USD",
			Units:        299,
			Nanos:        99 * 10000000,
		},
		StockQuantity: 50,
		ImageUrls:     []string{"https://example.com/headphones.jpg"},
		Attributes: map[string]string{
			"color":     "black",
			"bluetooth": "5.0",
		},
	})
	if err != nil {
		log.Fatalf("CreateProduct failed: %v", err)
	}

	product := createProdResp.Product
	log.Printf("✓ Created product:")
	log.Printf("  ID:       %s", product.Id)
	log.Printf("  Name:     %s", product.Name)
	log.Printf("  Price:    $%d.%02d %s", product.Price.Units, product.Price.Nanos/10000000, product.Price.CurrencyCode)
	log.Printf("  Category: %v", product.Category)
	log.Printf("  Stock:    %d units", product.StockQuantity)

	// Create an order
	log.Println("\n→ Creating order...")
	createOrderResp, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		CustomerId:   "customer-123",
		CustomerName: "Alice Johnson",
		Items: []*orderv1.OrderItem{
			{
				ProductId:   product.Id,
				ProductName: product.Name,
				Quantity:    2,
				UnitPrice:   product.Price,
				TotalPrice: &typesv1.Money{
					CurrencyCode: "USD",
					Units:        product.Price.Units * 2,
				},
			},
		},
		ShippingAddress: &locationv1.Address{
			Street:  "123 Main St",
			City:    "San Francisco",
			State:   "CA",
			ZipCode: "94102",
			Country: "USA",
		},
	})
	if err != nil {
		log.Fatalf("CreateOrder failed: %v", err)
	}

	order := createOrderResp.Order
	log.Printf("✓ Created order:")
	log.Printf("  ID:       %s", order.Id)
	log.Printf("  Customer: %s", order.CustomerName)
	log.Printf("  Items:    %d", len(order.Items))
	log.Printf("  Subtotal: $%d.00", order.Subtotal.Units)
	log.Printf("  Tax:      $%d.00", order.Tax.Units)
	log.Printf("  Total:    $%d.00", order.Total.Units)
	log.Printf("  Status:   %v", order.Status)
	log.Printf("  Address:  %s, %s %s", order.ShippingAddress.City, order.ShippingAddress.State, order.ShippingAddress.ZipCode)

	// Update order status
	log.Println("\n→ Updating order status...")
	updateResp, err := orderClient.UpdateOrderStatus(ctx, &orderv1.UpdateOrderStatusRequest{
		Id:     order.Id,
		Status: typesv1.Status_STATUS_ACTIVE,
		Reason: "Payment confirmed",
	})
	if err != nil {
		log.Fatalf("UpdateOrderStatus failed: %v", err)
	}
	log.Printf("✓ Order status updated to: %v", updateResp.Order.Status)

	// List orders
	log.Println("\n→ Listing orders...")
	listResp, err := orderClient.ListOrders(ctx, &orderv1.ListOrdersRequest{
		CustomerId: "customer-123",
	})
	if err != nil {
		log.Fatalf("ListOrders failed: %v", err)
	}
	log.Printf("✓ Found %d orders for customer", listResp.TotalCount)
	for _, o := range listResp.Orders {
		log.Printf("  - Order %s: $%d.00 (%v)", o.Id, o.Total.Units, o.Status)
	}

	// Search products
	log.Println("\n→ Searching products...")
	searchResp, err := productClient.SearchProducts(ctx, &productv1.SearchProductsRequest{
		Category: productv1.ProductCategory_CATEGORY_ELECTRONICS,
	})
	if err != nil {
		log.Fatalf("SearchProducts failed: %v", err)
	}
	log.Printf("✓ Found %d products", searchResp.TotalCount)
	for _, p := range searchResp.Products {
		log.Printf("  - %s: $%d.%02d", p.Name, p.Price.Units, p.Price.Nanos/10000000)
	}

	// ========== Test Order Service V2 ==========
	log.Println("\n\n========== Testing Order Service V2 ==========")

	// Subject prefix "api.v2" read from proto!
	orderClientV2 := orderv2.NewOrderServiceNatsClient(nc)

	// Print v2 client endpoints
	log.Println("\n📡 OrderService V2 Client Endpoints:")
	for _, ep := range orderClientV2.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	// Create order via v2
	log.Println("\n→ Creating order via v2...")
	createOrderV2Resp, err := orderClientV2.CreateOrder(ctx, &orderv2.CreateOrderRequest{
		CustomerId:   "customer-456",
		CustomerName: "Bob Smith",
		Items: []*orderv2.OrderItem{
			{
				ProductId:   product.Id,
				ProductName: product.Name,
				Quantity:    1,
				UnitPrice:   product.Price,
				TotalPrice: &typesv1.Money{
					CurrencyCode: "USD",
					Units:        product.Price.Units,
				},
			},
		},
		ShippingAddress: &locationv1.Address{
			Street:  "456 Oak Ave",
			City:    "New York",
			State:   "NY",
			ZipCode: "10001",
			Country: "USA",
		},
	})
	if err != nil {
		log.Fatalf("CreateOrder v2 failed: %v", err)
	}

	orderV2 := createOrderV2Resp.Order
	log.Printf("✓ [V2] Created order:")
	log.Printf("  ID:       %s", orderV2.Id)
	log.Printf("  Customer: %s", orderV2.CustomerName)
	log.Printf("  Total:    $%d.00", orderV2.Total.Units)
	log.Printf("  Status:   %v", orderV2.Status)

	// List v2 orders
	log.Println("\n→ Listing v2 orders...")
	listV2Resp, err := orderClientV2.ListOrders(ctx, &orderv2.ListOrdersRequest{
		CustomerId: "customer-456",
	})
	if err != nil {
		log.Fatalf("ListOrders v2 failed: %v", err)
	}
	log.Printf("✓ [V2] Found %d orders for customer", listV2Resp.TotalCount)
	for _, o := range listV2Resp.Orders {
		log.Printf("  - Order %s: $%d.00 (%v)", o.Id, o.Total.Units, o.Status)
	}

	log.Println("\n✅ All tests passed! Both v1 and v2 APIs working!")
}
