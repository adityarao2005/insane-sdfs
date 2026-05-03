# Copilot Instructions — short mission

Goal
- Help maintainers and contributors work effectively across the TS/Bun client and Go server.

Quick start (what I do automatically)
- Run `task proto` after `.proto` edits to regenerate client and server stubs.
- Use `task test` at the repo root to run client + server tests.

Preferred commands
- Client: use `bun` in `client/` for install/run/test (not Node/npm).
- Server: use `go test ./...` and `go run .` from `server/`.
- Orchestrate: use `Taskfile.yml` tasks at the repo root and per-component Taskfile.yml files.

Where to look first
- AGENTS.md — repository guide for coding agents (top-level).
- Taskfile.yml and client/Taskfile.yml for canonical build/test tasks.
- proto/*.proto for API surface; client/src/proto and server/pb for generated code.

Behavior guidance for the assistant
- Link to existing docs rather than copying them. Prefer to add short, actionable edits.
- When editing APIs defined by protos, update both the `.proto` and client/server callsites and re-run `task proto`.
- For client changes run `bun test` locally; for server changes run `go test ./...`.

If unsure, ask
- If a change touches both client and server or the proto API, ask whether to open a PR that updates both sides and regenerates stubs.
