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

	fileSystemService := filesystem_service.FileSystemService{}

	grpcServer := filesystem_service.NewGrpcFileSystemServer()
	grpcServer.AddService(&fileSystemService)
	grpcServer.Start(port)

	log.Printf("gRPC file system server is running on port %s", port)
}