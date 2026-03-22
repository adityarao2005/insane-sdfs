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
