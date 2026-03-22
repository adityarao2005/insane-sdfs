# insane-sdfs

Initial implementation scaffold for a secure file sharing system with:

- LAN-only enrollment (admin invite + manual approval)
- Single-use, expiring invite tokens
- No implicit WAN enrollment path
- Foundation for NAT traversal + relay fallback in later phases
- Hole-punching module with integration + Docker e2e test harness

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
- `GET /v1/admin/devices` (header `X-Admin-Token`)
- `GET /v1/admin/sessions/active` (header `X-Admin-Token`)
- `POST /v1/admin/sessions/start` (header `X-Admin-Token`, testing scaffold)
- `POST /v1/admin/sessions/heartbeat` (header `X-Admin-Token`)
- `POST /v1/admin/devices/revoke` (header `X-Admin-Token`)
- `POST /v1/enroll/request` (LAN-only)

Revocation behavior:

- Revoking a device immediately terminates all active sessions for that device.
- Revoke response includes `terminatedSessions` count.

Session TTL behavior:

- Sessions have a server-side TTL and automatically expire when heartbeats stop.
- Heartbeat refresh extends session expiry.

This is an implementation starting point focused on trust onboarding and enrollment security.

## Test Strategy

This repository now includes a test pyramid:

- Unit tests: token security logic, LAN classification, punch protocol encoding.
- Integration tests: API behavior and in-process punch flow with rendezvous.
- E2E tests: Docker-container workflow in an isolated Docker network for hole-punch path validation.

Run default tests (unit + integration):

```bash
go test ./...
```

Run e2e tests (Docker required):

```bash
SDFS_RUN_E2E=1 go test -tags=e2e ./test/e2e -v
```

E2E timeout behavior:

- The suite has an overall timeout and per-command fail-fast timeouts (build/setup/client-run) so hangs fail quickly.
- If Docker is unavailable, the e2e test is skipped.

Notes:

- E2E is opt-in to keep regular CI/dev loops fast.
- Docker network e2e validates the protocol path and containerized connectivity behavior.
- Full real-world NAT variability still needs additional staging tests on heterogeneous networks.

## CI

GitHub Actions workflow runs on every push and pull request with two jobs:

- Unit + integration tests (`go test ./...`)
- E2E hole-punching test (`SDFS_RUN_E2E=1 go test -count=1 -tags=e2e ./test/e2e -v`)
