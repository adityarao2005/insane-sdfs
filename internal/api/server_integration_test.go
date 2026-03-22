package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"insane-sdfs/internal/enrollment"
	"insane-sdfs/internal/filestore"
	"insane-sdfs/internal/security"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	tokens := security.NewInviteTokenStore()
	svc := enrollment.NewService(tokens)
	store := filestore.NewLocalStore(t.TempDir())
	return NewServer(ServerDeps{
		AdminToken:       "admin-secret",
		Enrollment:       svc,
		FileStore:        store,
		DefaultInviteTTL: 10 * time.Minute,
	})
}

func TestIssueInviteAndEnrollmentLANOnly(t *testing.T) {
	h := newTestServer(t)

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
	h := newTestServer(t)
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
	h := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/enrollments/pending", nil)
	req.RemoteAddr = "192.168.1.5:52000"
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got=%d", rec.Code)
	}
}

func TestAdminDashboardServed(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got=%d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "SDFS Admin Dashboard") {
		t.Fatalf("dashboard content missing marker")
	}
}

func TestRevokeDeviceFlow(t *testing.T) {
	h := newTestServer(t)

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

	heartbeatAfterRevokeBody, _ := json.Marshal(map[string]string{"sessionId": startSessionResp.ID})
	heartbeatAfterRevokeReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sessions/heartbeat", bytes.NewReader(heartbeatAfterRevokeBody))
	heartbeatAfterRevokeReq.Header.Set("X-Admin-Token", "admin-secret")
	heartbeatAfterRevokeRec := httptest.NewRecorder()
	h.ServeHTTP(heartbeatAfterRevokeRec, heartbeatAfterRevokeReq)
	if heartbeatAfterRevokeRec.Code != http.StatusConflict {
		t.Fatalf("expected heartbeat conflict after revoke, got=%d body=%s", heartbeatAfterRevokeRec.Code, heartbeatAfterRevokeRec.Body.String())
	}
}

func TestClientFileUploadListDownload(t *testing.T) {
	h := newTestServer(t)

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
		"deviceName":      "client-a",
		"devicePublicKey": "pk-a",
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

	startBody, _ := json.Marshal(map[string]string{"deviceId": deviceResp.ID})
	startReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sessions/start", bytes.NewReader(startBody))
	startReq.Header.Set("X-Admin-Token", "admin-secret")
	startRec := httptest.NewRecorder()
	h.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start session status=%d body=%s", startRec.Code, startRec.Body.String())
	}
	var sess struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(startRec.Body.Bytes(), &sess); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}

	uploadReq := httptest.NewRequest(http.MethodPut, "/v1/client/files/object?path=docs/note.txt", strings.NewReader("hello-client"))
	uploadReq.Header.Set("X-Session-Id", sess.ID)
	uploadRec := httptest.NewRecorder()
	h.ServeHTTP(uploadRec, uploadReq)
	if uploadRec.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", uploadRec.Code, uploadRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/client/files/list?prefix=docs", nil)
	listReq.Header.Set("X-Session-Id", sess.ID)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Items []struct {
			Path string `json:"path"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listed.Items) != 1 || listed.Items[0].Path != "docs/note.txt" {
		t.Fatalf("unexpected list payload: %s", listRec.Body.String())
	}

	dlReq := httptest.NewRequest(http.MethodGet, "/v1/client/files/object?path=docs/note.txt", nil)
	dlReq.Header.Set("X-Session-Id", sess.ID)
	dlRec := httptest.NewRecorder()
	h.ServeHTTP(dlRec, dlReq)
	if dlRec.Code != http.StatusOK {
		t.Fatalf("download status=%d body=%s", dlRec.Code, dlRec.Body.String())
	}
	if got := dlRec.Body.String(); got != "hello-client" {
		t.Fatalf("unexpected download body: %q", got)
	}
}

func TestClientSessionGetByPublicKey(t *testing.T) {
	h := newTestServer(t)

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

	publicKey := "pk-client-session-get"
	enrollBody, _ := json.Marshal(map[string]string{
		"inviteToken":     inviteResp.Token,
		"deviceName":      "client-b",
		"devicePublicKey": publicKey,
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

	sessionGetBody, _ := json.Marshal(map[string]string{"devicePublicKey": publicKey})
	sessionGetReq := httptest.NewRequest(http.MethodPost, "/v1/client/sessions/get", bytes.NewReader(sessionGetBody))
	sessionGetRec := httptest.NewRecorder()
	h.ServeHTTP(sessionGetRec, sessionGetReq)
	if sessionGetRec.Code != http.StatusCreated {
		t.Fatalf("session-get status=%d body=%s", sessionGetRec.Code, sessionGetRec.Body.String())
	}
	var sess struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sessionGetRec.Body.Bytes(), &sess); err != nil {
		t.Fatalf("unmarshal session-get: %v", err)
	}
	if sess.ID == "" {
		t.Fatalf("expected non-empty session id")
	}

	uploadReq := httptest.NewRequest(http.MethodPut, "/v1/client/files/object?path=docs/from-session-get.txt", strings.NewReader("hello-session-get"))
	uploadReq.Header.Set("X-Session-Id", sess.ID)
	uploadRec := httptest.NewRecorder()
	h.ServeHTTP(uploadRec, uploadReq)
	if uploadRec.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", uploadRec.Code, uploadRec.Body.String())
	}
}
