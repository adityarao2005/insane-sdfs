package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"insane-sdfs/internal/api"
	"insane-sdfs/internal/enrollment"
	"insane-sdfs/internal/filestore"
	"insane-sdfs/internal/security"
)

func main() {
	adminToken := os.Getenv("SDFS_ADMIN_TOKEN")
	if adminToken == "" {
		log.Fatal("SDFS_ADMIN_TOKEN must be set")
	}

	listenAddr := os.Getenv("SDFS_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8080"
	}

	tokenStore := security.NewInviteTokenStore()
	enrollmentSvc := enrollment.NewService(tokenStore)
	dataDir := os.Getenv("SDFS_DATA_DIR")
	fileStore := filestore.NewLocalStore(dataDir)

	handler := api.NewServer(api.ServerDeps{
		AdminToken:       adminToken,
		Enrollment:       enrollmentSvc,
		FileStore:        fileStore,
		DefaultInviteTTL: 10 * time.Minute,
	})

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("sdfs server listening on %s", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
