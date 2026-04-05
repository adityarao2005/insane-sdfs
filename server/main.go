package main

import (
	"log"
	"os"
	"server/filesystem_service"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "8080"
	}

	basePath := os.Getenv("FILESYSTEM_BASE_PATH")

	if basePath == "" {
		log.Printf("FILESYSTEM_BASE_PATH not set, using current directory as base path")
		basePath = "./temp"
		os.MkdirAll(basePath, 0o755)
	}

	fileSystemService, err := filesystem_service.NewFileSystemService(basePath)
	if err != nil {
		log.Fatalf("failed to initialize file system service: %v", err)
	}

	grpcServer := filesystem_service.NewGrpcFileSystemServer()
	if err := grpcServer.AddService(fileSystemService); err != nil {
		log.Fatalf("failed to add grpc service: %v", err)
	}

	log.Printf("gRPC file system server is running on port %s with base path %q", port, basePath)
	if err := grpcServer.Start(port); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}