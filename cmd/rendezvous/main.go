package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"insane-sdfs/internal/connectivity/punch"
)

func main() {
	listen := os.Getenv("SDFS_RENDEZVOUS_ADDR")
	if listen == "" {
		listen = ":39000"
	}

	srv, err := punch.NewServer(listen)
	if err != nil {
		log.Fatalf("failed to start rendezvous: %v", err)
	}
	defer srv.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("rendezvous listening on %s", srv.LocalAddr().String())
	if err := srv.Serve(ctx); err != nil {
		log.Fatalf("rendezvous stopped with error: %v", err)
	}
}
