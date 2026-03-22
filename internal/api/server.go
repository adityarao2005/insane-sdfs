package api

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"insane-sdfs/internal/enrollment"
	"insane-sdfs/internal/network"
	"insane-sdfs/internal/security"
)

type ServerDeps struct {
	AdminToken       string
	Enrollment       *enrollment.Service
	DefaultInviteTTL time.Duration
}

type server struct {
	adminToken       string
	enrollment       *enrollment.Service
	defaultInviteTTL time.Duration
}

func NewServer(deps ServerDeps) http.Handler {
	s := &server{
		adminToken:       deps.AdminToken,
		enrollment:       deps.Enrollment,
		defaultInviteTTL: deps.DefaultInviteTTL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /v1/admin/invites", s.withAdminAuth(s.handleIssueInvite))
	mux.HandleFunc("GET /v1/admin/enrollments/pending", s.withAdminAuth(s.handleListPendingEnrollments))
	mux.HandleFunc("POST /v1/admin/enrollments/approve", s.withAdminAuth(s.handleApproveEnrollment))
	mux.HandleFunc("GET /v1/admin/devices", s.withAdminAuth(s.handleListDevices))
	mux.HandleFunc("GET /v1/admin/sessions/active", s.withAdminAuth(s.handleListActiveSessions))
	mux.HandleFunc("POST /v1/admin/sessions/start", s.withAdminAuth(s.handleStartSession))
	mux.HandleFunc("POST /v1/admin/sessions/heartbeat", s.withAdminAuth(s.handleHeartbeatSession))
	mux.HandleFunc("POST /v1/admin/devices/revoke", s.withAdminAuth(s.handleRevokeDevice))
	mux.HandleFunc("POST /v1/enroll/request", s.handleCreateEnrollmentRequest)
	return mux
}

func (s *server) withAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get("X-Admin-Token")
		if tok == "" || tok != s.adminToken {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin auth failed"})
			return
		}
		next(w, r)
	}
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type issueInviteRequest struct {
	TTLSeconds int `json:"ttlSeconds"`
}

type issueInviteResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s *server) handleIssueInvite(w http.ResponseWriter, r *http.Request) {
	var req issueInviteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	ttl := s.defaultInviteTTL
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	token, expiresAt, err := s.enrollment.IssueInvite(ttl, time.Now().UTC())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue invite"})
		return
	}

	writeJSON(w, http.StatusCreated, issueInviteResponse{Token: token, ExpiresAt: expiresAt})
}

type createEnrollmentRequest struct {
	InviteToken     string `json:"inviteToken"`
	DeviceName      string `json:"deviceName"`
	DevicePublicKey string `json:"devicePublicKey"`
}

func (s *server) handleCreateEnrollmentRequest(w http.ResponseWriter, r *http.Request) {
	remoteIP, err := clientIP(r.RemoteAddr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid remote address"})
		return
	}
	if !network.IsPrivateLANIP(remoteIP) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "enrollment only allowed on LAN"})
		return
	}

	var req createEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.InviteToken) == "" || strings.TrimSpace(req.DeviceName) == "" || strings.TrimSpace(req.DevicePublicKey) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "inviteToken, deviceName, and devicePublicKey are required"})
		return
	}

	enrollReq, err := s.enrollment.CreateEnrollmentRequest(req.InviteToken, req.DeviceName, req.DevicePublicKey, remoteIP.String(), time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, security.ErrInviteNotFound):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid invite"})
		case errors.Is(err, security.ErrInviteExpired):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invite expired"})
		case errors.Is(err, security.ErrInviteUsed):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invite already used"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create enrollment request"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, enrollReq)
}

func (s *server) handleListPendingEnrollments(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.enrollment.ListPendingRequests(),
	})
}

type approveEnrollmentRequest struct {
	RequestID string `json:"requestId"`
}

func (s *server) handleApproveEnrollment(w http.ResponseWriter, r *http.Request) {
	var req approveEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RequestID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "requestId is required"})
		return
	}

	dev, err := s.enrollment.ApproveEnrollmentRequest(req.RequestID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, enrollment.ErrEnrollmentRequestNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "request not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not approve enrollment"})
		return
	}
	writeJSON(w, http.StatusCreated, dev)
}

func (s *server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.enrollment.ListDevices(),
	})
}

type revokeDeviceRequest struct {
	DeviceID string `json:"deviceId"`
}

type startSessionRequest struct {
	DeviceID string `json:"deviceId"`
}

type heartbeatSessionRequest struct {
	SessionID string `json:"sessionId"`
}

func (s *server) handleListActiveSessions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.enrollment.ListActiveSessions(),
	})
}

func (s *server) handleStartSession(w http.ResponseWriter, r *http.Request) {
	var req startSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DeviceID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "deviceId is required"})
		return
	}

	sess, err := s.enrollment.StartSession(req.DeviceID, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, enrollment.ErrDeviceNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		case errors.Is(err, enrollment.ErrDeviceRevoked):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "device is revoked"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not start session"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, sess)
}

func (s *server) handleHeartbeatSession(w http.ResponseWriter, r *http.Request) {
	var req heartbeatSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.SessionID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sessionId is required"})
		return
	}

	sess, err := s.enrollment.HeartbeatSession(req.SessionID, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, enrollment.ErrSessionNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		case errors.Is(err, enrollment.ErrSessionExpired):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session expired"})
		case errors.Is(err, enrollment.ErrSessionInactive):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session inactive"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not heartbeat session"})
		}
		return
	}

	writeJSON(w, http.StatusOK, sess)
}

func (s *server) handleRevokeDevice(w http.ResponseWriter, r *http.Request) {
	var req revokeDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DeviceID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "deviceId is required"})
		return
	}

	dev, terminatedCount, err := s.enrollment.RevokeDevice(req.DeviceID, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, enrollment.ErrDeviceNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		case errors.Is(err, enrollment.ErrDeviceAlreadyRevoked):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "device already revoked"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not revoke device"})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device":             dev,
		"terminatedSessions": terminatedCount,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func clientIP(remoteAddr string) (net.IP, error) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("invalid ip")
	}
	return ip, nil
}
