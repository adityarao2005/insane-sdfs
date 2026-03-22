package enrollment

import (
	"testing"
	"time"

	"insane-sdfs/internal/security"
)

func TestCreateAndApproveEnrollment(t *testing.T) {
	tokens := security.NewInviteTokenStore()
	svc := NewService(tokens)
	now := time.Unix(1000, 0).UTC()

	invite, _, err := svc.IssueInvite(10*time.Minute, now)
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	req, err := svc.CreateEnrollmentRequest(invite, "alice-phone", "pubkey-1", "192.168.1.15", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if req.Status != "pending_admin_approval" {
		t.Fatalf("unexpected request status: %s", req.Status)
	}

	dev, err := svc.ApproveEnrollmentRequest(req.ID, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}
	if dev.Name != "alice-phone" {
		t.Fatalf("unexpected device name: %s", dev.Name)
	}
}

func TestRevokeDevice(t *testing.T) {
	tokens := security.NewInviteTokenStore()
	svc := NewService(tokens)
	now := time.Unix(1000, 0).UTC()

	invite, _, err := svc.IssueInvite(10*time.Minute, now)
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}
	req, err := svc.CreateEnrollmentRequest(invite, "alice-laptop", "pubkey-2", "192.168.1.20", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	dev, err := svc.ApproveEnrollmentRequest(req.ID, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	if _, err := svc.StartSession(dev.ID, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("start session 1: %v", err)
	}
	if _, err := svc.StartSession(dev.ID, now.Add(4*time.Minute)); err != nil {
		t.Fatalf("start session 2: %v", err)
	}

	revoked, terminated, err := svc.RevokeDevice(dev.ID, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("revoke device: %v", err)
	}
	if !revoked.Revoked {
		t.Fatal("expected revoked device state")
	}
	if terminated != 2 {
		t.Fatalf("expected 2 terminated sessions, got %d", terminated)
	}
	if got := len(svc.ListActiveSessions()); got != 0 {
		t.Fatalf("expected 0 active sessions, got %d", got)
	}

	_, _, err = svc.RevokeDevice(dev.ID, now.Add(6*time.Minute))
	if err != ErrDeviceAlreadyRevoked {
		t.Fatalf("expected ErrDeviceAlreadyRevoked, got: %v", err)
	}

	_, err = svc.StartSession(dev.ID, now.Add(7*time.Minute))
	if err != ErrDeviceRevoked {
		t.Fatalf("expected ErrDeviceRevoked, got: %v", err)
	}
}

func TestSessionHeartbeatAndTTLExpiry(t *testing.T) {
	tokens := security.NewInviteTokenStore()
	svc := NewServiceWithSessionTTL(tokens, 2*time.Minute)
	now := time.Unix(1000, 0).UTC()

	invite, _, err := svc.IssueInvite(10*time.Minute, now)
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}
	req, err := svc.CreateEnrollmentRequest(invite, "phone", "pk", "192.168.1.11", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	dev, err := svc.ApproveEnrollmentRequest(req.ID, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	sess, err := svc.StartSession(dev.ID, now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	hb, err := svc.HeartbeatSession(sess.ID, now.Add(4*time.Minute))
	if err != nil {
		t.Fatalf("heartbeat session: %v", err)
	}
	if !hb.ExpiresAt.Equal(now.Add(6 * time.Minute)) {
		t.Fatalf("unexpected expiresAt after heartbeat: %s", hb.ExpiresAt)
	}

	if active := svc.ListActiveSessionsAt(now.Add(7 * time.Minute)); len(active) != 0 {
		t.Fatalf("expected 0 active sessions after ttl expiry, got %d", len(active))
	}

	_, err = svc.HeartbeatSession(sess.ID, now.Add(7*time.Minute))
	if err != ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired after ttl, got %v", err)
	}
}

func TestStartSessionByPublicKey(t *testing.T) {
	tokens := security.NewInviteTokenStore()
	svc := NewService(tokens)
	now := time.Unix(2000, 0).UTC()

	invite, _, err := svc.IssueInvite(10*time.Minute, now)
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}
	_, err = svc.CreateEnrollmentRequest(invite, "phone", "pk-session", "192.168.1.21", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if _, err := svc.StartSessionByPublicKey("pk-session", now.Add(2*time.Minute)); err != ErrDevicePublicKeyNotFound {
		t.Fatalf("expected ErrDevicePublicKeyNotFound before approval, got %v", err)
	}

	req := svc.ListPendingRequests()
	if len(req) != 1 {
		t.Fatalf("expected 1 pending request, got %d", len(req))
	}
	dev, err := svc.ApproveEnrollmentRequest(req[0].ID, now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("approve request: %v", err)
	}

	sess, err := svc.StartSessionByPublicKey("pk-session", now.Add(4*time.Minute))
	if err != nil {
		t.Fatalf("start session by public key: %v", err)
	}
	if sess.DeviceID != dev.ID {
		t.Fatalf("unexpected device id in session: got=%s want=%s", sess.DeviceID, dev.ID)
	}

	_, _, err = svc.RevokeDevice(dev.ID, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("revoke device: %v", err)
	}
	if _, err := svc.StartSessionByPublicKey("pk-session", now.Add(6*time.Minute)); err != ErrDeviceRevoked {
		t.Fatalf("expected ErrDeviceRevoked after revocation, got %v", err)
	}
}
