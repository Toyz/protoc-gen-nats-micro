package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	kvdemov1 "example/gen/kvstore_demo/v1"
)

// ============================================================================
// Service Implementation
// ============================================================================

type kvDemoService struct {
	profiles map[string]*kvdemov1.ProfileResponse
}

func (s *kvDemoService) SaveProfile(ctx context.Context, req *kvdemov1.SaveProfileRequest) (*kvdemov1.ProfileResponse, error) {
	resp := &kvdemov1.ProfileResponse{
		Id:        req.Id,
		Name:      req.Name,
		Email:     req.Email,
		Bio:       req.Bio,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Store in-memory (the KV auto-persistence happens in generated handler code)
	s.profiles[req.Id] = resp

	log.Printf("✓ SaveProfile: saved profile for %s (%s)", req.Name, req.Id)
	return resp, nil
}

func (s *kvDemoService) GetProfile(ctx context.Context, req *kvdemov1.GetProfileRequest) (*kvdemov1.ProfileResponse, error) {
	profile, ok := s.profiles[req.Id]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", req.Id)
	}

	log.Printf("✓ GetProfile: returning profile for %s", req.Id)
	return profile, nil
}

func (s *kvDemoService) GenerateReport(ctx context.Context, req *kvdemov1.GenerateReportRequest) (*kvdemov1.ReportResponse, error) {
	// Simulate report generation
	content := []byte(fmt.Sprintf("Report: %s (format: %s)\nGenerated at: %s",
		req.Title, req.Format, time.Now().Format(time.RFC3339)))

	resp := &kvdemov1.ReportResponse{
		Id:          req.Id,
		Title:       req.Title,
		Content:     content,
		ContentType: fmt.Sprintf("application/%s", req.Format),
		SizeBytes:   int64(len(content)),
	}

	log.Printf("✓ GenerateReport: created %s report '%s' (%d bytes)", req.Format, req.Title, len(content))
	return resp, nil
}

// ============================================================================
// Main
// ============================================================================

func main() {
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("✓ Connected to NATS")

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}
	log.Println("✓ JetStream context created")

	// Pre-create KV and Object Store buckets
	ctx := context.Background()

	_, err = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      "user_profiles",
		Description: "Auto-persisted user profiles from SaveProfile RPC",
	})
	if err != nil {
		log.Fatalf("Failed to create KV bucket: %v", err)
	}
	log.Println("✓ KV bucket 'user_profiles' ready")

	_, err = js.CreateOrUpdateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      "reports",
		Description: "Auto-persisted reports from GenerateReport RPC",
	})
	if err != nil {
		log.Fatalf("Failed to create Object Store bucket: %v", err)
	}
	log.Println("✓ Object Store bucket 'reports' ready")

	// Register service with JetStream enabled
	// The WithJetStream option enables automatic KV/ObjectStore persistence
	impl := &kvDemoService{profiles: make(map[string]*kvdemov1.ProfileResponse)}
	svc, err := kvdemov1.RegisterKVStoreDemoServiceHandlers(nc, impl,
		kvdemov1.WithJetStream(js), // ← This enables auto-persistence!
	)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// Print registered endpoints
	log.Println("\n📡 KVStoreDemoService Endpoints:")
	for _, ep := range svc.Endpoints() {
		log.Printf("  • %s → %s", ep.Name, ep.Subject)
	}

	log.Println("\n🔥 Server ready — waiting for requests...")
	log.Println("   SaveProfile  → auto-persists to KV 'user_profiles' (key: user.{id})")
	log.Println("   GenerateReport → auto-persists to Object Store 'reports' (key: report.{id})")

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("\n⏹ Shutting down...")
	svc.Stop()
}
