# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build          # Build binary to bin/wabee
make install        # Install to GOBIN
make test           # Run all tests
make lint           # Run golangci-lint (requires installation)
make fmt            # Format code with go fmt and goimports
make tidy           # Tidy go modules
make build-all      # Build for linux, darwin, windows (amd64/arm64)
```

Run a single test:
```bash
go test -v -run TestName ./path/to/package
```

## Architecture

This is a Go CLI application using Cobra for command structure and Viper for configuration management.

### Package Structure

- `cmd/` - Cobra command definitions (root, chat, sessions, trace, config, agent)
- `internal/client/` - HTTP client for Wabee API with SSE streaming support
- `internal/config/` - Configuration management with profile support
- `internal/output/` - Output formatting (JSON, table, tree, text) with lipgloss styling
- `pkg/models/` - Shared data models for API requests/responses

### Key Patterns

**Command Flow**: `main.go` → `cmd.Execute()` → `rootCmd.PersistentPreRunE` (config init) → subcommand

**API Client**: `internal/client/Client` wraps HTTP calls. Supports both request/response (`Chat`) and SSE streaming (`ChatStream`). Authentication via `X-API-Key` header or `Authorization: Bearer` token.

**Configuration**: Stored in `~/.wabee/config.yaml`. Supports multiple profiles with environment variable overrides (`WABEE_API_KEY`, `WABEE_ENDPOINT`, `WABEE_PROFILE`).

**Output Formatting**: `internal/output/Formatter` handles multiple output formats. Commands use `getFormatter()` from root.go. Color support detected automatically with `NO_COLOR` env var respected.

### API Endpoints

The client communicates with these Wabee agent endpoints (all prefixed with `/core/v1`):
- `POST /core/v1/chain` - Non-streaming chat
- `POST /core/v1/chain_streaming` - SSE streaming chat
- `GET /core/v1/sessions` - List sessions
- `GET /core/v1/sessions/{id}` - Get session details
- `GET /core/v1/sessions/{id}/trace` - Get execution trace
- `DELETE /core/v1/sessions/{id}` - Delete session
- `GET /core/v1/tools` - List available tools
- `GET /core/v1/metadata` - Agent info
- `GET /core/v1/health` - Health check
