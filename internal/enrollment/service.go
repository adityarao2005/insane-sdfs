package enrollment

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"insane-sdfs/internal/security"
)

var ErrEnrollmentRequestNotFound = errors.New("enrollment request not found")
var ErrDeviceNotFound = errors.New("device not found")
var ErrDevicePublicKeyNotFound = errors.New("device public key not found")
var ErrDeviceAlreadyRevoked = errors.New("device already revoked")
var ErrDeviceRevoked = errors.New("device is revoked")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session expired")
var ErrSessionInactive = errors.New("session inactive")

const defaultSessionTTL = 10 * time.Minute

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

type Session struct {
	ID              string     `json:"id"`
	DeviceID        string     `json:"deviceId"`
	CreatedAt       time.Time  `json:"createdAt"`
	LastHeartbeatAt time.Time  `json:"lastHeartbeatAt"`
	ExpiresAt       time.Time  `json:"expiresAt"`
	Active          bool       `json:"active"`
	TerminatedAt    *time.Time `json:"terminatedAt,omitempty"`
}

type Service struct {
	mu            sync.Mutex
	nextReqID     int64
	nextDevID     int64
	nextSessionID int64
	invites       *security.InviteTokenStore
	requests      map[string]Request
	devices       map[string]Device
	sessions      map[string]Session
	sessionTTL    time.Duration
}

func NewService(invites *security.InviteTokenStore) *Service {
	return NewServiceWithSessionTTL(invites, defaultSessionTTL)
}

func NewServiceWithSessionTTL(invites *security.InviteTokenStore, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}

	return &Service{
		invites:    invites,
		requests:   make(map[string]Request),
		devices:    make(map[string]Device),
		sessions:   make(map[string]Session),
		sessionTTL: sessionTTL,
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

func (s *Service) ListDevices() []Device {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, d)
	}
	return out
}

func (s *Service) RevokeDevice(deviceID string, now time.Time) (Device, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dev, ok := s.devices[deviceID]
	if !ok {
		return Device{}, 0, ErrDeviceNotFound
	}
	if dev.Revoked {
		return Device{}, 0, ErrDeviceAlreadyRevoked
	}

	dev.Revoked = true
	s.devices[deviceID] = dev
	terminated := s.terminateDeviceSessionsLocked(deviceID, now)
	return dev, terminated, nil
}

func (s *Service) StartSession(deviceID string, now time.Time) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpiredSessionsLocked(now)

	dev, ok := s.devices[deviceID]
	if !ok {
		return Session{}, ErrDeviceNotFound
	}
	if dev.Revoked {
		return Session{}, ErrDeviceRevoked
	}

	return s.startSessionLocked(deviceID, now), nil
}

func (s *Service) StartSessionByPublicKey(publicKey string, now time.Time) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpiredSessionsLocked(now)

	for _, dev := range s.devices {
		if dev.PublicKey != publicKey {
			continue
		}
		if dev.Revoked {
			return Session{}, ErrDeviceRevoked
		}
		return s.startSessionLocked(dev.ID, now), nil
	}

	return Session{}, ErrDevicePublicKeyNotFound
}

func (s *Service) startSessionLocked(deviceID string, now time.Time) Session {

	s.nextSessionID++
	sid := fmt.Sprintf("sess-%06d", s.nextSessionID)
	session := Session{
		ID:              sid,
		DeviceID:        deviceID,
		CreatedAt:       now,
		LastHeartbeatAt: now,
		ExpiresAt:       now.Add(s.sessionTTL),
		Active:          true,
	}
	s.sessions[sid] = session
	return session
}

func (s *Service) ListActiveSessions() []Session {
	return s.ListActiveSessionsAt(time.Now().UTC())
}

func (s *Service) ListActiveSessionsAt(now time.Time) []Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpiredSessionsLocked(now)

	out := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		if sess.Active {
			out = append(out, sess)
		}
	}
	return out
}

func (s *Service) HeartbeatSession(sessionID string, now time.Time) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpiredSessionsLocked(now)

	sess, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	if !sess.Active {
		if !sess.ExpiresAt.IsZero() && !now.Before(sess.ExpiresAt) {
			return Session{}, ErrSessionExpired
		}
		return Session{}, ErrSessionInactive
	}

	if !sess.ExpiresAt.IsZero() && !now.Before(sess.ExpiresAt) {
		sess.Active = false
		term := now
		sess.TerminatedAt = &term
		s.sessions[sessionID] = sess
		return Session{}, ErrSessionExpired
	}

	sess.LastHeartbeatAt = now
	sess.ExpiresAt = now.Add(s.sessionTTL)
	s.sessions[sessionID] = sess
	return sess, nil
}

func (s *Service) GetActiveSession(sessionID string, now time.Time) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sweepExpiredSessionsLocked(now)

	sess, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	if !sess.Active {
		if !sess.ExpiresAt.IsZero() && !now.Before(sess.ExpiresAt) {
			return Session{}, ErrSessionExpired
		}
		return Session{}, ErrSessionInactive
	}
	return sess, nil
}

func (s *Service) TerminateSession(sessionID string, now time.Time) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	if !sess.Active {
		return sess, nil
	}

	sess.Active = false
	term := now
	sess.TerminatedAt = &term
	s.sessions[sessionID] = sess
	return sess, nil
}

func (s *Service) terminateDeviceSessionsLocked(deviceID string, now time.Time) int {
	terminated := 0
	for id, sess := range s.sessions {
		if sess.DeviceID != deviceID || !sess.Active {
			continue
		}
		sess.Active = false
		term := now
		sess.TerminatedAt = &term
		s.sessions[id] = sess
		terminated++
	}
	return terminated
}

func (s *Service) sweepExpiredSessionsLocked(now time.Time) int {
	terminated := 0
	for id, sess := range s.sessions {
		if !sess.Active {
			continue
		}
		if sess.ExpiresAt.IsZero() || now.Before(sess.ExpiresAt) {
			continue
		}
		sess.Active = false
		term := now
		sess.TerminatedAt = &term
		s.sessions[id] = sess
		terminated++
	}
	return terminated
}
