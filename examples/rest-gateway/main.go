package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderv1 "github.com/toyz/protoc-gen-nats-micro/gen/order/v1"
	productv1 "github.com/toyz/protoc-gen-nats-micro/gen/product/v1"
)

func main() {
	// Connect to NATS for the backend services
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	log.Println("✓ Connected to NATS")

	// Start gRPC server (that uses NATS backend)
	grpcAddr := "localhost:9090"
	go func() {
		// This would need a gRPC server implementation that calls NATS
		// For now, this is just a placeholder to show the architecture
		log.Printf("gRPC server would run on %s", grpcAddr)
	}()

	// Create gRPC-Gateway mux
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	// CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			h.ServeHTTP(w, r)
		})
	}

	// Register gRPC-Gateway handlers
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = orderv1.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		log.Printf("Warning: Failed to register OrderService gateway: %v", err)
	}

	err = productv1.RegisterProductServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		log.Printf("Warning: Failed to register ProductService gateway: %v", err)
	}

	// Start REST gateway
	restAddr := ":8080"
	
	// Serve OpenAPI spec
	mux.HandlePath("GET", "/openapi.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "gen/api.swagger.json")
	})
	
	// Swagger UI redirect
	mux.HandlePath("GET", "/docs", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.Redirect(w, r, "https://petstore.swagger.io/?url=http://localhost:8080/openapi.json", http.StatusTemporaryRedirect)
	})
	
	server := &http.Server{
		Addr:    restAddr,
		Handler: corsHandler(mux),
	}

	go func() {
		log.Printf("✓ REST Gateway listening on %s", restAddr)
		log.Println("\nAvailable REST endpoints:")
		log.Println("  Products (v1):")
		log.Println("    POST   /v1/products")
		log.Println("    GET    /v1/products/{id}")
		log.Println("    PATCH  /v1/products/{id}")
		log.Println("    DELETE /v1/products/{id}")
		log.Println("    GET    /v1/products (search)")
		log.Println("  Orders (v1):")
		log.Println("    POST   /v1/orders")
		log.Println("    GET    /v1/orders/{id}")
		log.Println("    GET    /v1/orders")
		log.Println("    PATCH  /v1/orders/{id}/status")
		log.Println("  Orders (v2):")
		log.Println("    POST   /v2/orders")
		log.Println("    GET    /v2/orders/{id}")
		log.Println("    GET    /v2/orders")
		log.Println("    PATCH  /v2/orders/{id}/status")
		log.Println("\nAPI Documentation:")
		log.Println("  OpenAPI spec:  http://localhost:8080/openapi.json")
		log.Println("  Swagger UI:    http://localhost:8080/docs")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("\n✓ Shutting down REST gateway...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
}
