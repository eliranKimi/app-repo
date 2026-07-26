package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds" // Register xDS resolvers and balancers

	pb "app-repo/helloworld"
)

const (
	defaultName    = "world"
	serverAddr     = "xds:///greeter-service:50051"
	requestTimeout = 5 * time.Second
	loopInterval   = 2 * time.Second
)

func main() {
	// Set up a connection to the server using the xDS resolver.
	// The xds:/// scheme tells gRPC to use the xDS control plane (Traffic Director)
	// for service discovery and load balancing — no sidecar proxy needed.
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Use the first CLI argument as the name, or fall back to default
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	log.Printf("Starting greeter client, sending requests to %s", serverAddr)

	for {
		// Create a new context per request (not deferred — cancel called explicitly)
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		cancel() // Always cancel immediately after the call, not deferred

		if err != nil {
			log.Printf("could not greet: %v", err)
		} else {
			log.Printf("Greeting: %s", r.GetMessage())
		}

		time.Sleep(loopInterval)
	}
}
