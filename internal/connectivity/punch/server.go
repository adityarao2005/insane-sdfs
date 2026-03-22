package punch

import (
	"context"
	"errors"
	"net"
	"sync"
)

type Server struct {
	conn *net.UDPConn

	mu       sync.Mutex
	sessions map[string][]peerRegistration
}

type peerRegistration struct {
	PeerID string
	Addr   *net.UDPAddr
}

func NewServer(listenAddr string) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{conn: conn, sessions: make(map[string][]peerRegistration)}, nil
}

func (s *Server) LocalAddr() *net.UDPAddr {
	return s.conn.LocalAddr().(*net.UDPAddr)
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) Serve(ctx context.Context) error {
	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := s.conn.SetReadDeadline(deadlineFromContext(ctx)); err != nil {
			return err
		}
		n, remote, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return err
		}
		msg, err := decodeMessage(buf[:n])
		if err != nil {
			continue
		}
		if msg.Type != "register" || msg.Token == "" || msg.PeerID == "" {
			continue
		}

		s.handleRegister(msg.Token, msg.PeerID, remote)
	}
}

func (s *Server) handleRegister(token, peerID string, remote *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.sessions[token]
	for _, p := range list {
		if p.PeerID == peerID {
			return
		}
	}
	list = append(list, peerRegistration{PeerID: peerID, Addr: remote})
	s.sessions[token] = list
	if len(list) < 2 {
		return
	}

	first := list[0]
	second := list[1]

	msgToFirst, _ := encodeMessage(Message{Type: "peer", PeerID: second.PeerID, PeerAddr: second.Addr.String()})
	msgToSecond, _ := encodeMessage(Message{Type: "peer", PeerID: first.PeerID, PeerAddr: first.Addr.String()})
	_, _ = s.conn.WriteToUDP(msgToFirst, first.Addr)
	_, _ = s.conn.WriteToUDP(msgToSecond, second.Addr)

	delete(s.sessions, token)
}
