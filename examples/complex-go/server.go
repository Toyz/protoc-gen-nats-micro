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

	metadatav1 "example/gen/common/metadata/v1"
	typesv1 "example/gen/common/types/v1"
	demov1 "example/gen/demo/v1"
	orderv1 "example/gen/order/v1"
	orderv2 "example/gen/order/v2"
	productv1 "example/gen/product/v1"
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

// JSON Demo service implementation
type jsonDemoService struct{}

func (s *jsonDemoService) Echo(ctx context.Context, req *demov1.EchoRequest) (*demov1.EchoResponse, error) {
	log.Printf("✓ [JSON] Echo: %s", req.Message)
	return &demov1.EchoResponse{
		Message:   "JSON Echo: " + req.Message,
		Timestamp: time.Now().Unix(),
		Encoding:  "json",
	}, nil
}

func (s *jsonDemoService) GetUser(ctx context.Context, req *demov1.GetUserRequest) (*demov1.GetUserResponse, error) {
	log.Printf("✓ [JSON] GetUser: %s", req.Id)
	return &demov1.GetUserResponse{
		User: &demov1.User{
			Id:    req.Id,
			Name:  "json_user_" + req.Id,
			Email: "json@example.com",
			Roles: []string{"user", "admin"},
			Metadata: map[string]string{
				"encoding": "json",
				"service":  "json_service",
			},
		},
	}, nil
}

// Binary Demo service implementation
type binaryDemoService struct{}

func (s *binaryDemoService) Echo(ctx context.Context, req *demov1.EchoRequest) (*demov1.EchoResponse, error) {
	log.Printf("✓ [BINARY] Echo: %s", req.Message)
	return &demov1.EchoResponse{
		Message:   "Binary Echo: " + req.Message,
		Timestamp: time.Now().Unix(),
		Encoding:  "binary",
	}, nil
}

func (s *binaryDemoService) GetUser(ctx context.Context, req *demov1.GetUserRequest) (*demov1.GetUserResponse, error) {
	log.Printf("✓ [BINARY] GetUser: %s", req.Id)
	return &demov1.GetUserResponse{
		User: &demov1.User{
			Id:    req.Id,
			Name:  "binary_user_" + req.Id,
			Email: "binary@example.com",
			Roles: []string{"user", "tester"},
			Metadata: map[string]string{
				"encoding": "binary",
				"service":  "binary_service",
			},
		},
	}, nil
}

// Example server interceptors

// Product service interceptors - demonstrates reading request headers and setting response headers
func productLoggingInterceptor(ctx context.Context, req interface{}, info *productv1.UnaryServerInfo, handler productv1.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("→ [%s.%s] Request started", info.Service, info.Method)

	// Read incoming headers from context
	if headers := productv1.IncomingHeaders(ctx); headers != nil {
		if traceID, ok := headers["X-Trace-Id"]; ok && len(traceID) > 0 {
			log.Printf("  [Headers] Trace-ID: %s", traceID[0])
		}
		if clientVer, ok := headers["X-Client-Version"]; ok && len(clientVer) > 0 {
			log.Printf("  [Headers] Client-Version: %s", clientVer[0])
		}
	}

	// Add response headers
	responseHeaders := nats.Header{}
	responseHeaders.Set("X-Server-Version", "1.0.0")
	responseHeaders.Set("X-Request-Id", fmt.Sprintf("req-%d", time.Now().UnixNano()))
	productv1.SetResponseHeaders(ctx, responseHeaders)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("✗ [%s.%s] Request failed after %v: %v", info.Service, info.Method, duration, err)
	} else {
		log.Printf("✓ [%s.%s] Request completed in %v", info.Service, info.Method, duration)
	}

	return resp, err
}

func productMetricsInterceptor(ctx context.Context, req interface{}, info *productv1.UnaryServerInfo, handler productv1.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	status := "success"
	if err != nil {
		status = "error"
	}

	log.Printf("📊 [METRICS] service=%s method=%s status=%s duration_ms=%d",
		info.Service, info.Method, status, duration.Milliseconds())

	return resp, err
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	log.Println("✓ Connected to NATS")

	// Register product service (subject prefix "api.v1" read from proto!)
	// with logging and metrics interceptors
	productSvc := &productService{products: make(map[string]*productv1.Product)}
	productService, err := productv1.RegisterProductServiceHandlers(nc, productSvc,
		productv1.WithServerInterceptor(productLoggingInterceptor),
		productv1.WithServerInterceptor(productMetricsInterceptor),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered ProductService (with logging + metrics interceptors)")

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

	// Register JSON demo service (subject prefix "demo.json" with JSON encoding!)
	jsonDemoSvc := &jsonDemoService{}
	jsonService, err := demov1.RegisterJSONServiceHandlers(nc, jsonDemoSvc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered JSONService (JSON encoding)")

	// Register Binary demo service (subject prefix "demo.binary" with binary encoding!)
	binaryDemoSvc := &binaryDemoService{}
	binaryService, err := demov1.RegisterBinaryServiceHandlers(nc, binaryDemoSvc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Registered BinaryService (Binary encoding)")

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

	log.Println("\n📡 JSONService Endpoints (JSON encoding):")
	for _, ep := range jsonService.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n📡 BinaryService Endpoints (Binary encoding):")
	for _, ep := range binaryService.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n✅ Server running. Press Ctrl+C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("\n✓ Shutting down...")
}
