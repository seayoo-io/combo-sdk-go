# AGENTS.md

## Project Overview

combo-sdk-go is a Go server-side SDK for the Seayoo Combo game platform (世游核心系统).
It provides identity token verification, server REST API client, notification handling, and GM command processing.

- Module: `github.com/seayoo-io/combo-sdk-go`
- Package: `combo`
- Go version: 1.21.5
- Current version: 1.1.0 (see `version.go`)
- License: MIT

## Build & Development Commands

```bash
make deps      # Install gotestsum and golangci-lint
make build     # Build the package
make test      # Run tests with race detection and coverage (requires CGO_ENABLED=1)
make lint      # Run golangci-lint
make fmt       # Format code with gofmt
make tidy      # Clean and tidy go.mod/go.sum
```

## Code Style & Conventions

- Go files use **tab** indentation (standard Go style)
- Other files use **spaces** (indent_size=4 default, 2 for xml/yml/yaml/json/sh)
- All files use UTF-8, LF line endings, trailing whitespace trimmed, final newline inserted
- Single package (`combo`) — all source files are in the repo root, no sub-packages
- Chinese comments are used throughout (this is a Chinese game platform SDK)
- Strong typing: custom types for `Endpoint`, `GameId`, `SecretKey`, `Platform`, `IdP` rather than raw strings/bytes
- Constants use `Type_Value` naming (e.g., `Platform_iOS`, `IdP_Google`, `Endpoint_China`)
- Error responses use `*ErrorResponse` with `errors.As()` pattern for typed error handling
- HTTP handlers implement `http.Handler` interface for standard library compatibility

## Architecture

| Area | Files | Description |
|------|-------|-------------|
| Types & Config | `model.go`, `config.go`, `version.go`, `user_agent.go` | Core types, enums, configuration validation |
| HTTP Signing | `signer.go` | SEAYOO-HMAC-SHA256 signing and verification (5-min time skew tolerance) |
| API Client | `api_client.go`, `api_response.go` | Base HTTP client with auto-signing |
| APIs | `api_create_order.go`, `api_enter_game.go`, `api_leave_game.go` | Individual API implementations |
| Notifications | `notifications.go` | HTTP handler for server push notifications (ship order, refund) |
| GM Commands | `gm.go`, `gm_idempotency.go` | GM command handler with optional idempotency (Redis or in-memory) |
| Token Verification | `verifier.go` | JWT (HMAC-SHA256) token verification for identity and ad tokens |
| Tests | `*_test.go` | Unit tests for all modules above (run with `make test`) |

## Key Dependencies

- `github.com/golang-jwt/jwt/v5` — JWT token parsing
- `github.com/google/uuid` — UUID generation for idempotency
- `github.com/redis/go-redis/v9` — Redis client for idempotency store (optional)

## Important Notes

- No CI/CD workflows configured (no `.github/workflows/`)
- API naming pattern: `api_<action>.go` for new API endpoint files
- Notification types follow `Handle<Action>` pattern in `NotificationListener` interface
- GM idempotency requires Redis >= 7.0 for the Redis-backed store
- `Config.validate()` mutates the Endpoint (trims trailing slash) — be aware of this side effect
