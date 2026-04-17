# Web And Network Notes

- Use this directory guidance for changes under `internal/web/` and nearby network entrypoints.
- Keep it scoped to durable rules; use `ai-context/network-summary.md` only for optional orientation, not as authority over the current code.

## Boundaries

- `internal/web/` owns HTTP, HTTPS, WebSocket upgrade handling, template serving, and `/admin/` routes.
- Telnet protocol behavior lives in `internal/term/`, connection lifecycle and heartbeat logic live in `internal/connections/`, and login/input flow lives in `internal/inputhandlers/`.
- Listener startup is split: telnet ports are started from [main.go](/home/user/sync/git/tmp/kc-GoMud/main.go), while HTTP/HTTPS listeners are started from [web.go](/home/user/sync/git/tmp/kc-GoMud/internal/web/web.go).

## Source Of Truth

- Read `_datafiles/config.yaml` and `internal/configs/config.network.go` before changing documented or default ports.
- Current defaults are telnet `33333` and `44444`, local admin telnet `9999`, HTTP `80`, HTTPS disabled with `0`; verify them from code/config before claiming behavior.
- Preserve the split between public web access and localhost-only admin/telnet behavior unless the task explicitly changes it.

## Change Rules

- Keep WebSocket endpoint behavior aligned with `/ws` handling and related reconnect/copyover code paths.
- For HTTPS changes, validate both direct HTTPS startup and the `HttpsRedirect` path; avoid changes that silently enable redirect behavior without a working HTTPS server.
- If a change crosses `internal/web/`, `internal/connections/`, `internal/inputhandlers/`, or `main.go`, document the boundary and test the full flow instead of patching one layer in isolation.

## Verification

- Prefer the smallest relevant check first: `make validate` for Go-only changes, targeted `go test ./internal/...` for touched packages, then `make test` for broader network/runtime changes.
- For config or port changes, also verify any affected docs against `make help`, `_datafiles/config.yaml`, and the actual startup path in code.
