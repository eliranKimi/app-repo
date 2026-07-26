package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/xds"

	pb "app-repo/helloworld"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	hostname, _ := os.Hostname()
	log.Printf("Received: %v on host %s", in.GetName(), hostname)
	return &pb.HelloReply{Message: "Hello " + in.GetName() + " from " + hostname}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server with xDS credentials.
	s := grpc.NewServer()

	// Register the Greeter server.
	pb.RegisterGreeterServer(s, &server{})

	// Register the health server.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
