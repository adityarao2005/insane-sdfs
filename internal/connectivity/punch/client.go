package punch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

type Client struct {
	PeerID     string
	Token      string
	ServerAddr *net.UDPAddr
	Timeout    time.Duration
}

// Attempt performs rendezvous registration and then UDP hole-punch style simultaneous dialing.
func (c Client) Attempt(ctx context.Context) (*net.UDPAddr, error) {
	if c.PeerID == "" || c.Token == "" || c.ServerAddr == nil {
		return nil, fmt.Errorf("peerID, token, and serverAddr are required")
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(deadlineFromContext(ctx)); err != nil {
		return nil, err
	}

	reg, _ := encodeMessage(Message{Type: "register", Token: c.Token, PeerID: c.PeerID})
	if _, err := conn.WriteToUDP(reg, c.ServerAddr); err != nil {
		return nil, err
	}

	peerAddr, err := c.waitForPeerInfo(ctx, conn, reg)
	if err != nil {
		return nil, err
	}

	return c.punch(ctx, conn, peerAddr)
}

func (c Client) waitForPeerInfo(ctx context.Context, conn *net.UDPConn, reg []byte) (*net.UDPAddr, error) {
	buf := make([]byte, 2048)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for peer info: %w", ctx.Err())
		case <-ticker.C:
			_, _ = conn.WriteToUDP(reg, c.ServerAddr)
		default:
		}

		if err := conn.SetReadDeadline(deadlineFromContext(ctx)); err != nil {
			return nil, err
		}
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				return nil, err
			}
			continue
		}

		msg, err := decodeMessage(buf[:n])
		if err != nil {
			continue
		}
		if msg.Type != "peer" || msg.PeerAddr == "" {
			continue
		}
		peerAddr, err := net.ResolveUDPAddr("udp", msg.PeerAddr)
		if err != nil {
			continue
		}
		return peerAddr, nil
	}
}

func (c Client) punch(ctx context.Context, conn *net.UDPConn, peer *net.UDPAddr) (*net.UDPAddr, error) {
	buf := make([]byte, 2048)
	msg, _ := encodeMessage(Message{Type: "punch", PeerID: c.PeerID})
	ack, _ := encodeMessage(Message{Type: "punch-ack", PeerID: c.PeerID})
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out while punching: %w", ctx.Err())
		case <-ticker.C:
			_, _ = conn.WriteToUDP(msg, peer)
		default:
		}

		if err := conn.SetReadDeadline(deadlineFromContext(ctx)); err != nil {
			return nil, err
		}
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				return nil, err
			}
			continue
		}
		in, err := decodeMessage(buf[:n])
		if err != nil {
			continue
		}
		if in.Type == "punch" {
			// Reply once so the peer that initiated this packet can complete as well.
			_, _ = conn.WriteToUDP(ack, remote)
			return remote, nil
		}
		if in.Type == "punch-ack" {
			return remote, nil
		}
	}
}

func deadlineFromContext(ctx context.Context) time.Time {
	const pollInterval = 200 * time.Millisecond
	next := time.Now().Add(pollInterval)
	if d, ok := ctx.Deadline(); ok {
		if d.Before(next) {
			return d
		}
	}
	return next
}
