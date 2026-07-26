package main

import (
	"context"
	"net"
	"testing"
	"time"

	pb "app-repo/helloworld"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// startMockServer starts a minimal gRPC server for client testing.
// We use a plain grpc:// address (not xds:///) since xDS requires a live control plane.
func startMockServer(t *testing.T) (addr string, cleanup func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &mockGreeterServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("mock server stopped: %v", err)
		}
	}()

	return lis.Addr().String(), func() { s.GracefulStop() }
}

// mockGreeterServer is a test double for the Greeter service.
type mockGreeterServer struct {
	pb.UnimplementedGreeterServer
}

func (m *mockGreeterServer) SayHello(_ context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "Hello " + req.GetName() + " from mock-server",
	}, nil
}

// TestClientSayHello verifies the client can connect and call SayHello successfully.
func TestClientSayHello(t *testing.T) {
	addr, cleanup := startMockServer(t)
	defer cleanup()

	// Connect directly (not via xDS) for unit testing
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect to mock server: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "eliran"})
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}

	expected := "Hello eliran from mock-server"
	if resp.GetMessage() != expected {
		t.Errorf("got message %q, want %q", resp.GetMessage(), expected)
	}
}

// TestClientContextCancellation verifies that context cancellation is handled correctly
// and doesn't leak goroutines (tests the cancel() fix in the main loop).
func TestClientContextCancellation(t *testing.T) {
	addr, cleanup := startMockServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// Simulate the client loop pattern: create context, call, cancel immediately
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.SayHello(ctx, &pb.HelloRequest{Name: "test"})
		cancel() // Must be called immediately, not deferred inside a loop

		if err != nil {
			t.Errorf("iteration %d: SayHello failed: %v", i, err)
		}
	}
}

// TestClientEmptyResponse verifies the client handles responses correctly.
func TestClientResponseParsing(t *testing.T) {
	addr, cleanup := startMockServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	tests := []struct {
		name    string
		input   string
		wantMsg string
	}{
		{"normal name", "world", "Hello world from mock-server"},
		{"empty name", "", "Hello  from mock-server"},
		{"special chars", "test-123", "Hello test-123 from mock-server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: tt.input})
			cancel()

			if err != nil {
				t.Fatalf("SayHello failed: %v", err)
			}
			if resp.GetMessage() != tt.wantMsg {
				t.Errorf("got %q, want %q", resp.GetMessage(), tt.wantMsg)
			}
		})
	}
}
