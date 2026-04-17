package main

import (
	"log"
	"os"
	"path/filepath"
	"server/auth"
	"server/filesystem_service"
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

	// create the necessary directories for certificates and files
	certDir := filepath.Join(basePath, "certs")
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		log.Fatalf("failed to create certificate directory: %v", err)
	}

	// create the file directory
	fileDir := filepath.Join(basePath, "root")
	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		log.Fatalf("failed to create file directory: %v", err)
	}

	// create the token service
	tokenService := auth.NewTokenService()

	// create the certificate service
	certificateService, err := auth.NewCertificateService(certDir)
	if err != nil {
		log.Fatalf("failed to initialize certificate service: %v", err)
	}

	// create the file system service
	fileSystemService, err := filesystem_service.NewFileSystemService(fileDir)
	if err != nil {
		log.Fatalf("failed to initialize file system service: %v", err)
	}

	// create the grpc server with the specified options
	grpcServer := filesystem_service.NewGrpcFileSystemServer(certificateService)
	if err := grpcServer.AddService(fileSystemService); err != nil {
		log.Fatalf("failed to add grpc service: %v", err)
	}

	// create the admin service
	adminService, err := NewAdminService(tokenService, certificateService)
	if err != nil {
		log.Fatalf("failed to initialize admin service: %v", err)
	}

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
