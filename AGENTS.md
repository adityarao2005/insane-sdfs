# AGENTS.md — Repository guide for coding agents

Purpose
- Provide concise, actionable instructions for AI coding agents to be immediately productive in this repo.

Quick commands
- Root: run common orchestration via Taskfiles

```sh
task proto    # regenerate protos for client & server
task test     # run project-level tests (client + server tasks)
```

- Client (uses Bun)

```sh
cd client
bun install
bun test
```

- Server (Go)

```sh
cd server
go test ./...
go run .
```

Notes & conventions
- Runtime: client uses Bun (not Node). Prefer `bun` for install/run/test in `client/`.
- Proto-first workflow: edit `proto/*.proto` and run `task proto` (root) or `npm run proto:create` from `client/` to regenerate stubs.
- Generated outputs:
  - client: `client/src/proto/` (JS/TS gRPC stubs)
  - server: `server/pb/` (Go pb files)
- Tests:
  - client tests: `client/src/tests/`
  - server tests: `server/*_test.go` and package tests under `server/`
- Taskfiles: Use the per-directory `Taskfile.yml` (root, `client/`, `server/`) as authoritative automation steps.

Where to look first (high-value files)
- [README.md](README.md) — repo overview
- [Taskfile.yml](Taskfile.yml) — root orchestration
- [client/README.md](client/README.md) — client-specific notes
- [client/Taskfile.yml](client/Taskfile.yml)
- [client/package.json](client/package.json)
- [client/src/index.ts](client/src/index.ts)
- [client/src/crypto/cert-manager.ts](client/src/crypto/cert-manager.ts)
- [client/src/cli/actions.ts](client/src/cli/actions.ts)
- [client/src/tests](client/src/tests)
- [proto/FileSystemService.proto](proto/FileSystemService.proto)
- [proto/HelloService.proto](proto/HelloService.proto)
- [server/Taskfile.yml](server/Taskfile.yml)
- [server/main.go](server/main.go)
- [server/server.go](server/server.go)
- [server/filesystem_service/filesystem_service.go](server/filesystem_service/filesystem_service.go)
- [server/pb](server/pb) — generated Go protobuf files
- [server/auth](server/auth) — token & certificate logic

Agent guidance
- Link, don't copy: reference docs using the links above rather than duplicating long sections.
- Re-run `task proto` after .proto edits and verify generated stubs are updated in `client/src/proto` and `server/pb`.
- For client changes, run `bun test` in `client/` to validate behavior.
- For server changes, run `go test ./...` in `server/`.
- When creating or modifying APIs defined by protos, update both `.proto` and callsites in client/server then regenerate stubs.

Next steps
- If helpful, create `.github/copilot-instructions.md` with a shortened mission statement and mention preferred commands (Bun/Go/Taskfile).

Feedback
- Tell me any extra conventions you want emphasized (e.g., commit hooks, lint rules), and I will update this file.
