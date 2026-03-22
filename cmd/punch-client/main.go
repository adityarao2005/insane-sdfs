package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"insane-sdfs/internal/connectivity/punch"
)

func main() {
	var (
		peerID  = flag.String("peer-id", "", "peer identifier")
		token   = flag.String("token", "", "shared rendezvous token")
		server  = flag.String("server", "", "rendezvous UDP address host:port")
		timeout = flag.Duration("timeout", 20*time.Second, "attempt timeout")
	)
	flag.Parse()

	if *peerID == "" || *token == "" || *server == "" {
		log.Fatal("peer-id, token, and server are required")
	}

	addr, err := net.ResolveUDPAddr("udp", *server)
	if err != nil {
		log.Fatalf("resolve server: %v", err)
	}

	cli := punch.Client{
		PeerID:     *peerID,
		Token:      *token,
		ServerAddr: addr,
		Timeout:    *timeout,
	}
	remote, err := cli.Attempt(context.Background())
	if err != nil {
		log.Fatalf("hole punch failed: %v", err)
	}
	fmt.Fprintf(os.Stdout, "hole-punch-success peer=%s remote=%s\n", *peerID, remote.String())
}
