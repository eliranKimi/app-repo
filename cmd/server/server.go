package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	xdsgrpc "google.golang.org/grpc/xds"

	pb "app-repo/helloworld"
)

const (
	port       = ":50051"
	healthPort = ":50052" // Separate port for health checks (plain gRPC, no xDS)
)

// server implements helloworld.GreeterServer
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	hostname, _ := os.Hostname()
	log.Printf("Received request from: %q, serving from pod: %s", in.GetName(), hostname)
	return &pb.HelloReply{
		Message: "Hello " + in.GetName() + " from " + hostname,
	}, nil
}

func main() {
	// Start a plain gRPC health check server on a separate port.
	// This is needed because xds.NewGRPCServer() starts in NOT_SERVING mode
	// until it receives configuration from Traffic Director, creating a
	// chicken-and-egg problem with health checks.
	// Traffic Director health checks hit this plain port instead.
	go func() {
		healthLis, err := net.Listen("tcp", healthPort)
		if err != nil {
			log.Fatalf("failed to listen on health port %s: %v", healthPort, err)
		}
		healthSrv := grpc.NewServer()
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(healthSrv, healthServer)
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		healthServer.SetServingStatus("helloworld.Greeter", grpc_health_v1.HealthCheckResponse_SERVING)
		log.Printf("Health check server listening on %s", healthPort)
		if err := healthSrv.Serve(healthLis); err != nil {
			log.Fatalf("health server failed: %v", err)
		}
	}()

	// Main xDS gRPC server for the Greeter service
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", port, err)
	}

	// Use xds.NewGRPCServer() to enable proxyless gRPC integration with
	// Cloud Service Mesh (Traffic Director).
	s, err := xdsgrpc.NewGRPCServer(grpc.ChainUnaryInterceptor())
	if err != nil {
		log.Fatalf("failed to create xDS gRPC server: %v", err)
	}

	// Register the Greeter service
	pb.RegisterGreeterServer(s, &server{})

	// Also register health on the xDS server (for in-mesh health checks)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("helloworld.Greeter", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("xDS gRPC server listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
