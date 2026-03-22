package enrollment

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"insane-sdfs/internal/security"
)

var ErrEnrollmentRequestNotFound = errors.New("enrollment request not found")

type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"publicKey"`
	CreatedAt time.Time `json:"createdAt"`
	Revoked   bool      `json:"revoked"`
}

type Request struct {
	ID              string    `json:"id"`
	DeviceName      string    `json:"deviceName"`
	DevicePublicKey string    `json:"devicePublicKey"`
	RemoteIP        string    `json:"remoteIp"`
	CreatedAt       time.Time `json:"createdAt"`
	Status          string    `json:"status"`
}

type Service struct {
	mu        sync.Mutex
	nextReqID int64
	nextDevID int64
	invites   *security.InviteTokenStore
	requests  map[string]Request
	devices   map[string]Device
}

func NewService(invites *security.InviteTokenStore) *Service {
	return &Service{
		invites:  invites,
		requests: make(map[string]Request),
		devices:  make(map[string]Device),
	}
}

func (s *Service) IssueInvite(ttl time.Duration, now time.Time) (string, time.Time, error) {
	return s.invites.Issue(ttl, now)
}

func (s *Service) CreateEnrollmentRequest(inviteToken, deviceName, publicKey, remoteIP string, now time.Time) (Request, error) {
	if err := s.invites.Consume(inviteToken, now); err != nil {
		return Request{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextReqID++
	rid := fmt.Sprintf("req-%06d", s.nextReqID)
	req := Request{
		ID:              rid,
		DeviceName:      deviceName,
		DevicePublicKey: publicKey,
		RemoteIP:        remoteIP,
		CreatedAt:       now,
		Status:          "pending_admin_approval",
	}
	s.requests[rid] = req
	return req, nil
}

func (s *Service) ListPendingRequests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Request, 0, len(s.requests))
	for _, r := range s.requests {
		if r.Status == "pending_admin_approval" {
			out = append(out, r)
		}
	}
	return out
}

func (s *Service) ApproveEnrollmentRequest(requestID string, now time.Time) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return Device{}, ErrEnrollmentRequestNotFound
	}

	s.nextDevID++
	devID := fmt.Sprintf("dev-%06d", s.nextDevID)
	dev := Device{
		ID:        devID,
		Name:      req.DeviceName,
		PublicKey: req.DevicePublicKey,
		CreatedAt: now,
		Revoked:   false,
	}
	s.devices[devID] = dev

	req.Status = "approved"
	s.requests[requestID] = req
	return dev, nil
}
