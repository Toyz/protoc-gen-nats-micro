package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	metadatav1 "github.com/toyz/protoc-gen-nats-micro/gen/common/metadata/v1"
	typesv1 "github.com/toyz/protoc-gen-nats-micro/gen/common/types/v1"
	orderv1 "github.com/toyz/protoc-gen-nats-micro/gen/order/v1"
	orderv2 "github.com/toyz/protoc-gen-nats-micro/gen/order/v2"
	productv1 "github.com/toyz/protoc-gen-nats-micro/gen/product/v1"
)

// Product service implementation
type productService struct {
	products map[string]*productv1.Product
}

func (s *productService) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	id := uuid.New().String()
	now := time.Now()

	product := &productv1.Product{
		Id:            id,
		Name:          req.Name,
		Description:   req.Description,
		Sku:           req.Sku,
		Category:      req.Category,
		Price:         req.Price,
		StockQuantity: req.StockQuantity,
		ImageUrls:     req.ImageUrls,
		Attributes:    req.Attributes,
		Status:        typesv1.Status_STATUS_ACTIVE,
		Metadata: &metadatav1.Metadata{
			CreatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			UpdatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			CreatedBy: "system",
			UpdatedBy: "system",
			Tags:      make(map[string]string),
		},
	}
	s.products[id] = product

	log.Printf("✓ Created product: %s (%s) - $%d.%02d", req.Name, id, req.Price.Units, req.Price.Nanos/10000000)
	return &productv1.CreateProductResponse{Product: product}, nil
}

func (s *productService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	product, ok := s.products[req.Id]
	if !ok {
		return nil, fmt.Errorf("product not found: %s", req.Id)
	}
	log.Printf("✓ Retrieved product: %s", product.Name)
	return &productv1.GetProductResponse{Product: product}, nil
}

func (s *productService) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	product, ok := s.products[req.Id]
	if !ok {
		return nil, fmt.Errorf("product not found: %s", req.Id)
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.StockQuantity = req.StockQuantity
	product.ImageUrls = req.ImageUrls
	product.Attributes = req.Attributes
	product.Metadata.UpdatedAt = &typesv1.Timestamp{Seconds: time.Now().Unix()}
	product.Metadata.UpdatedBy = "system"

	log.Printf("✓ Updated product: %s", product.Name)
	return &productv1.UpdateProductResponse{Product: product}, nil
}

func (s *productService) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	delete(s.products, req.Id)
	log.Printf("✓ Deleted product: %s", req.Id)
	return &productv1.DeleteProductResponse{Success: true}, nil
}

func (s *productService) SearchProducts(ctx context.Context, req *productv1.SearchProductsRequest) (*productv1.SearchProductsResponse, error) {
	var results []*productv1.Product
	for _, p := range s.products {
		results = append(results, p)
	}
	log.Printf("✓ Search returned %d products", len(results))
	return &productv1.SearchProductsResponse{
		Products:   results,
		TotalCount: int32(len(results)),
	}, nil
}

// Order service implementation
type orderService struct {
	orders   map[string]*orderv1.Order
	products *productService
}

func (s *orderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	id := uuid.New().String()
	now := time.Now()

	var subtotal int64
	for _, item := range req.Items {
		subtotal += item.TotalPrice.Units
	}

	tax := subtotal / 10 // 10% tax
	total := subtotal + tax

	order := &orderv1.Order{
		Id:              id,
		CustomerId:      req.CustomerId,
		CustomerName:    req.CustomerName,
		Items:           req.Items,
		Subtotal:        &typesv1.Money{CurrencyCode: "USD", Units: subtotal},
		Tax:             &typesv1.Money{CurrencyCode: "USD", Units: tax},
		Total:           &typesv1.Money{CurrencyCode: "USD", Units: total},
		ShippingAddress: req.ShippingAddress,
		Status:          typesv1.Status_STATUS_PENDING,
		Metadata: &metadatav1.Metadata{
			CreatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			UpdatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			CreatedBy: req.CustomerId,
			UpdatedBy: req.CustomerId,
			Tags:      make(map[string]string),
		},
	}
	s.orders[id] = order

	log.Printf("✓ Created order: %s for %s - $%d.00 (%d items)", id, req.CustomerName, total, len(req.Items))
	return &orderv1.CreateOrderResponse{Order: order}, nil
}

func (s *orderService) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	order, ok := s.orders[req.Id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", req.Id)
	}
	log.Printf("✓ Retrieved order: %s", order.Id)
	return &orderv1.GetOrderResponse{Order: order}, nil
}

func (s *orderService) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	var results []*orderv1.Order
	for _, o := range s.orders {
		if req.CustomerId == "" || o.CustomerId == req.CustomerId {
			results = append(results, o)
		}
	}
	log.Printf("✓ Listed %d orders", len(results))
	return &orderv1.ListOrdersResponse{
		Orders:     results,
		TotalCount: int32(len(results)),
	}, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.UpdateOrderStatusResponse, error) {
	order, ok := s.orders[req.Id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", req.Id)
	}

	order.Status = req.Status
	order.Metadata.UpdatedAt = &typesv1.Timestamp{Seconds: time.Now().Unix()}
	order.Metadata.UpdatedBy = "system"

	log.Printf("✓ Updated order %s status to: %v", order.Id, req.Status)
	return &orderv1.UpdateOrderStatusResponse{Order: order}, nil
}

// Order service v2 implementation (same logic, different types)
type orderServiceV2 struct {
	orders   map[string]*orderv2.Order
	products *productService
}

func (s *orderServiceV2) CreateOrder(ctx context.Context, req *orderv2.CreateOrderRequest) (*orderv2.CreateOrderResponse, error) {
	id := uuid.New().String()
	now := time.Now()

	var subtotal int64
	for _, item := range req.Items {
		subtotal += item.TotalPrice.Units
	}

	tax := subtotal / 10 // 10% tax
	total := subtotal + tax

	order := &orderv2.Order{
		Id:              id,
		CustomerId:      req.CustomerId,
		CustomerName:    req.CustomerName,
		Items:           req.Items,
		Subtotal:        &typesv1.Money{CurrencyCode: "USD", Units: subtotal},
		Tax:             &typesv1.Money{CurrencyCode: "USD", Units: tax},
		Total:           &typesv1.Money{CurrencyCode: "USD", Units: total},
		ShippingAddress: req.ShippingAddress,
		Status:          typesv1.Status_STATUS_PENDING,
		Metadata: &metadatav1.Metadata{
			CreatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			UpdatedAt: &typesv1.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
			CreatedBy: req.CustomerId,
			UpdatedBy: req.CustomerId,
			Tags:      make(map[string]string),
		},
	}
	s.orders[id] = order

	log.Printf("✓ [V2] Created order: %s for %s - $%d.00 (%d items)", id, req.CustomerName, total, len(req.Items))
	return &orderv2.CreateOrderResponse{Order: order}, nil
}

func (s *orderServiceV2) GetOrder(ctx context.Context, req *orderv2.GetOrderRequest) (*orderv2.GetOrderResponse, error) {
	order, ok := s.orders[req.Id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", req.Id)
	}
	log.Printf("✓ [V2] Retrieved order: %s", order.Id)
	return &orderv2.GetOrderResponse{Order: order}, nil
}

func (s *orderServiceV2) ListOrders(ctx context.Context, req *orderv2.ListOrdersRequest) (*orderv2.ListOrdersResponse, error) {
	var results []*orderv2.Order
	for _, o := range s.orders {
		if req.CustomerId == "" || o.CustomerId == req.CustomerId {
			results = append(results, o)
		}
	}
	log.Printf("✓ [V2] Listed %d orders", len(results))
	return &orderv2.ListOrdersResponse{
		Orders:     results,
		TotalCount: int32(len(results)),
	}, nil
}

func (s *orderServiceV2) UpdateOrderStatus(ctx context.Context, req *orderv2.UpdateOrderStatusRequest) (*orderv2.UpdateOrderStatusResponse, error) {
	order, ok := s.orders[req.Id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", req.Id)
	}

	order.Status = req.Status
	order.Metadata.UpdatedAt = &typesv1.Timestamp{Seconds: time.Now().Unix()}
	order.Metadata.UpdatedBy = "system"

	log.Printf("✓ [V2] Updated order %s status to: %v", order.Id, req.Status)
	return &orderv2.UpdateOrderStatusResponse{Order: order}, nil
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	log.Println("✓ Connected to NATS")

	// Register product service (subject prefix "api.v1" read from proto!)
	productSvc := &productService{products: make(map[string]*productv1.Product)}
	productService, err := productv1.RegisterProductServiceHandlers(nc, productSvc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered ProductService")

	// Register order service v1 (subject prefix "api.v1" read from proto!)
	orderSvc := &orderService{
		orders:   make(map[string]*orderv1.Order),
		products: productSvc,
	}
	orderServiceV1, err := orderv1.RegisterOrderServiceHandlers(nc, orderSvc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered OrderService v1")

	// Register order service v2 (subject prefix "api.v2" read from proto!)
	orderSvcV2 := &orderServiceV2{
		orders:   make(map[string]*orderv2.Order),
		products: productSvc,
	}
	orderServiceV2, err := orderv2.RegisterOrderServiceHandlers(nc, orderSvcV2)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered OrderService v2")

	// Print all service endpoints
	log.Println("\n📡 ProductService Endpoints:")
	for _, ep := range productService.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n📡 OrderService V1 Endpoints:")
	for _, ep := range orderServiceV1.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n📡 OrderService V2 Endpoints:")
	for _, ep := range orderServiceV2.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n✅ Server running. Press Ctrl+C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("\n✓ Shutting down...")
}

