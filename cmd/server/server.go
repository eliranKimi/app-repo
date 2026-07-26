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
	port = ":50051"
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
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", port, err)
	}

	// Use xds.NewGRPCServer() instead of grpc.NewServer() to enable
	// proxyless gRPC integration with Cloud Service Mesh (Traffic Director).
	// This registers the server with the xDS control plane so it can receive
	// traffic management configuration (load balancing, routing, etc.)
	// without a sidecar proxy.
	s, err := xdsgrpc.NewGRPCServer(grpc.ChainUnaryInterceptor())
	if err != nil {
		log.Fatalf("failed to create xDS gRPC server: %v", err)
	}

	// Register the Greeter service
	pb.RegisterGreeterServer(s, &server{})

	// Register the health check service (required by Traffic Director health checks)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("helloworld.Greeter", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("xDS gRPC server listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
