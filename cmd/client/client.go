package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"

	pb "app-repo/helloworld"
)

const (
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	// The xds:/// scheme is used to connect to the gRPC service using the xDS API.
	// The service name is "greeter-service:50051", which will be resolved by the xDS server.
	conn, err := grpc.Dial("xds:///greeter-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("could not greet: %v", err)
		} else {
			log.Printf("Greeting: %s", r.GetMessage())
		}
		time.Sleep(2 * time.Second)
	}
}
