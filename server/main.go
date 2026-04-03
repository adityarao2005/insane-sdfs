package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"server/pb"

	"google.golang.org/grpc"
)

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "world"
	}

	return &pb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", name),
	}, nil
}

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "8080"
	}

	listenAddr := ":" + port
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", listenAddr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &greeterServer{})

	log.Printf("gRPC server listening on %s", listenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}