package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

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

	// Use a standard gRPC server. In proxyless gRPC with Traffic Director,
	// the xDS integration is on the CLIENT side (service discovery and load
	// balancing). The server is a regular gRPC server registered as a backend
	// in Traffic Director via a Network Endpoint Group (NEG).
	s := grpc.NewServer()

	// Register the Greeter service
	pb.RegisterGreeterServer(s, &server{})

	// Register the health check service (required by Traffic Director health checks)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("helloworld.Greeter", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("gRPC server listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
