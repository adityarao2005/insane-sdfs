# insane-sdfs

Initial implementation scaffold for a secure file sharing system with:

- LAN-only enrollment (admin invite + manual approval)
- Single-use, expiring invite tokens
- No implicit WAN enrollment path
- Foundation for NAT traversal + relay fallback in later phases

## Run

1. Set admin token:

```bash
export SDFS_ADMIN_TOKEN="change-me"
```

2. Start server (binds to localhost by default):

```bash
go run ./cmd/server
```

3. Health check:

```bash
curl -s http://127.0.0.1:8080/healthz
```

## Current API

- `POST /v1/admin/invites` (header `X-Admin-Token`)
- `GET /v1/admin/enrollments/pending` (header `X-Admin-Token`)
- `POST /v1/admin/enrollments/approve` (header `X-Admin-Token`)
- `POST /v1/enroll/request` (LAN-only)

This is an implementation starting point focused on trust onboarding and enrollment security.
