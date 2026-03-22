package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"insane-sdfs/internal/enrollment"
	"insane-sdfs/internal/security"
)

func newTestServer() http.Handler {
	tokens := security.NewInviteTokenStore()
	svc := enrollment.NewService(tokens)
	return NewServer(ServerDeps{
		AdminToken:       "admin-secret",
		Enrollment:       svc,
		DefaultInviteTTL: 10 * time.Minute,
	})
}

func TestIssueInviteAndEnrollmentLANOnly(t *testing.T) {
	h := newTestServer()

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/admin/invites", bytes.NewBufferString(`{"ttlSeconds":600}`))
	inviteReq.Header.Set("X-Admin-Token", "admin-secret")
	inviteRec := httptest.NewRecorder()
	h.ServeHTTP(inviteRec, inviteReq)
	if inviteRec.Code != http.StatusCreated {
		t.Fatalf("issue invite status=%d body=%s", inviteRec.Code, inviteRec.Body.String())
	}

	var inviteResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(inviteRec.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("unmarshal invite response: %v", err)
	}
	if inviteResp.Token == "" {
		t.Fatal("invite token is empty")
	}

	payload := map[string]string{
		"inviteToken":     inviteResp.Token,
		"deviceName":      "int-client",
		"devicePublicKey": "pk-int-client",
	}
	b, _ := json.Marshal(payload)

	enrollReq := httptest.NewRequest(http.MethodPost, "/v1/enroll/request", bytes.NewReader(b))
	enrollReq.RemoteAddr = "192.168.1.42:50000"
	enrollRec := httptest.NewRecorder()
	h.ServeHTTP(enrollRec, enrollReq)
	if enrollRec.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollRec.Code, enrollRec.Body.String())
	}
}

func TestEnrollmentRejectedFromPublicIP(t *testing.T) {
	h := newTestServer()
	payload := map[string]string{
		"inviteToken":     "unused",
		"deviceName":      "int-client",
		"devicePublicKey": "pk-int-client",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/enroll/request", bytes.NewReader(b))
	req.RemoteAddr = "8.8.8.8:51000"
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminAuthRequired(t *testing.T) {
	h := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/enrollments/pending", nil)
	req.RemoteAddr = "192.168.1.5:52000"
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got=%d", rec.Code)
	}
}

func TestRevokeDeviceFlow(t *testing.T) {
	h := newTestServer()

	inviteReq := httptest.NewRequest(http.MethodPost, "/v1/admin/invites", bytes.NewBufferString(`{"ttlSeconds":600}`))
	inviteReq.Header.Set("X-Admin-Token", "admin-secret")
	inviteRec := httptest.NewRecorder()
	h.ServeHTTP(inviteRec, inviteReq)
	if inviteRec.Code != http.StatusCreated {
		t.Fatalf("issue invite status=%d body=%s", inviteRec.Code, inviteRec.Body.String())
	}
	var inviteResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(inviteRec.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("unmarshal invite: %v", err)
	}

	enrollBody, _ := json.Marshal(map[string]string{
		"inviteToken":     inviteResp.Token,
		"deviceName":      "revocation-client",
		"devicePublicKey": "pk-revocation-client",
	})
	enrollReq := httptest.NewRequest(http.MethodPost, "/v1/enroll/request", bytes.NewReader(enrollBody))
	enrollReq.RemoteAddr = "192.168.1.42:50000"
	enrollRec := httptest.NewRecorder()
	h.ServeHTTP(enrollRec, enrollReq)
	if enrollRec.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollRec.Code, enrollRec.Body.String())
	}
	var enrollResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(enrollRec.Body.Bytes(), &enrollResp); err != nil {
		t.Fatalf("unmarshal enrollment: %v", err)
	}

	approveBody, _ := json.Marshal(map[string]string{"requestId": enrollResp.ID})
	approveReq := httptest.NewRequest(http.MethodPost, "/v1/admin/enrollments/approve", bytes.NewReader(approveBody))
	approveReq.Header.Set("X-Admin-Token", "admin-secret")
	approveRec := httptest.NewRecorder()
	h.ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusCreated {
		t.Fatalf("approve status=%d body=%s", approveRec.Code, approveRec.Body.String())
	}
	var deviceResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(approveRec.Body.Bytes(), &deviceResp); err != nil {
		t.Fatalf("unmarshal approved device: %v", err)
	}

	startSessionBody, _ := json.Marshal(map[string]string{"deviceId": deviceResp.ID})
	startSessionReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sessions/start", bytes.NewReader(startSessionBody))
	startSessionReq.Header.Set("X-Admin-Token", "admin-secret")
	startSessionRec := httptest.NewRecorder()
	h.ServeHTTP(startSessionRec, startSessionReq)
	if startSessionRec.Code != http.StatusCreated {
		t.Fatalf("start session status=%d body=%s", startSessionRec.Code, startSessionRec.Body.String())
	}
	var startSessionResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(startSessionRec.Body.Bytes(), &startSessionResp); err != nil {
		t.Fatalf("unmarshal start session: %v", err)
	}

	heartbeatBody, _ := json.Marshal(map[string]string{"sessionId": startSessionResp.ID})
	heartbeatReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sessions/heartbeat", bytes.NewReader(heartbeatBody))
	heartbeatReq.Header.Set("X-Admin-Token", "admin-secret")
	heartbeatRec := httptest.NewRecorder()
	h.ServeHTTP(heartbeatRec, heartbeatReq)
	if heartbeatRec.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", heartbeatRec.Code, heartbeatRec.Body.String())
	}

	listActiveReq := httptest.NewRequest(http.MethodGet, "/v1/admin/sessions/active", nil)
	listActiveReq.Header.Set("X-Admin-Token", "admin-secret")
	listActiveRec := httptest.NewRecorder()
	h.ServeHTTP(listActiveRec, listActiveReq)
	if listActiveRec.Code != http.StatusOK {
		t.Fatalf("list active status=%d body=%s", listActiveRec.Code, listActiveRec.Body.String())
	}
	var activeBefore struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listActiveRec.Body.Bytes(), &activeBefore); err != nil {
		t.Fatalf("unmarshal active sessions: %v", err)
	}
	if len(activeBefore.Items) != 1 {
		t.Fatalf("expected 1 active session before revoke, got %d", len(activeBefore.Items))
	}

	revokeBody, _ := json.Marshal(map[string]string{"deviceId": deviceResp.ID})
	revokeReq := httptest.NewRequest(http.MethodPost, "/v1/admin/devices/revoke", bytes.NewReader(revokeBody))
	revokeReq.Header.Set("X-Admin-Token", "admin-secret")
	revokeRec := httptest.NewRecorder()
	h.ServeHTTP(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusOK {
		t.Fatalf("revoke status=%d body=%s", revokeRec.Code, revokeRec.Body.String())
	}
	var revoked struct {
		Device struct {
			Revoked bool `json:"revoked"`
		} `json:"device"`
		TerminatedSessions int `json:"terminatedSessions"`
	}
	if err := json.Unmarshal(revokeRec.Body.Bytes(), &revoked); err != nil {
		t.Fatalf("unmarshal revoked: %v", err)
	}
	if !revoked.Device.Revoked {
		t.Fatal("expected revoked device in response")
	}
	if revoked.TerminatedSessions != 1 {
		t.Fatalf("expected terminatedSessions=1, got %d", revoked.TerminatedSessions)
	}

	listActiveAfterReq := httptest.NewRequest(http.MethodGet, "/v1/admin/sessions/active", nil)
	listActiveAfterReq.Header.Set("X-Admin-Token", "admin-secret")
	listActiveAfterRec := httptest.NewRecorder()
	h.ServeHTTP(listActiveAfterRec, listActiveAfterReq)
	if listActiveAfterRec.Code != http.StatusOK {
		t.Fatalf("list active after revoke status=%d body=%s", listActiveAfterRec.Code, listActiveAfterRec.Body.String())
	}
	var activeAfter struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal(listActiveAfterRec.Body.Bytes(), &activeAfter); err != nil {
		t.Fatalf("unmarshal active sessions after revoke: %v", err)
	}
	if len(activeAfter.Items) != 0 {
		t.Fatalf("expected 0 active sessions after revoke, got %d", len(activeAfter.Items))
	}
}
