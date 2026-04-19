package main

import (
	"log"
	"os"
)

func GetPort() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Printf("GRPC_PORT not set, defaulting to 8080")
		port = "8080"
	}
	return port
}

func GetBasePath() string {
	basePath := os.Getenv("FILESYSTEM_BASE_PATH")
	if basePath == "" {
		log.Printf("FILESYSTEM_BASE_PATH not set, using current directory as base path")
		basePath = "./temp"
		os.MkdirAll(basePath, 0o755)
	}
	return basePath
}

func main() {
	port := GetPort()
	basePath := GetBasePath()

	_, _, _, grpcServer, adminService := CreateServices(basePath)

	// start the admin service
	if err := adminService.Start(); err != nil {
		log.Fatalf("failed to start admin service: %v", err)
	}

	// start the gRPC server
	log.Printf("gRPC file system server is running on port %s with base path %q", port, basePath)
	if err := grpcServer.Start(port); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}
