package main

import (
	"context"
	"net"
	"testing"
	"time"

	pb "app-repo/helloworld"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// startTestServer starts a plain gRPC server (not xDS) for unit testing.
// xDS requires a live control plane, so we use grpc.NewServer() in tests.
func startTestServer(t *testing.T) (addr string, cleanup func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0") // random available port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	// Register health server
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()

	return lis.Addr().String(), func() { s.GracefulStop() }
}

func TestSayHello(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "test"})
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}

	if resp.GetMessage() == "" {
		t.Error("expected non-empty message in response")
	}

	t.Logf("Response: %s", resp.GetMessage())
}

func TestSayHelloEmptyName(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Empty name should still succeed (server should not panic or error)
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: ""})
	if err != nil {
		t.Fatalf("SayHello with empty name failed: %v", err)
	}

	if resp.GetMessage() == "" {
		t.Error("expected non-empty message even with empty name")
	}
}

func TestHealthCheck(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	healthClient := grpc_health_v1.NewHealthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: ""})
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("expected SERVING status, got %v", resp.GetStatus())
	}
}
