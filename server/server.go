package main

import (
	"log"
	"os"
	"path/filepath"
	"server/admin"
	"server/auth"
	"server/filesystem_service"
)

func CreateServices(basePath string) (auth.ITokenService, auth.ICertificateService, *filesystem_service.FileSystemService, *filesystem_service.GrpcFileSystemServer, *admin.AdminService) {

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
	adminService, err := admin.NewAdminService(tokenService, certificateService)
	if err != nil {
		log.Fatalf("failed to initialize admin service: %v", err)
	}

	return tokenService, certificateService, fileSystemService, grpcServer, adminService
}
