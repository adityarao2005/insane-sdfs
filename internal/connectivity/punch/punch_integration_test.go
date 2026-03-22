package punch

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestHolePunchAttemptIntegration(t *testing.T) {
	srv, err := NewServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Serve(ctx)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	token := "int-test-token"
	results := make(chan error, 2)
	startClient := func(peer string) {
		defer wg.Done()
		cli := Client{
			PeerID:     peer,
			Token:      token,
			ServerAddr: srv.LocalAddr(),
			Timeout:    5 * time.Second,
		}
		_, err := cli.Attempt(context.Background())
		results <- err
	}

	go startClient("a")
	go startClient("b")

	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatalf("attempt failed: %v", err)
		}
	}
}
