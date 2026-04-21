# AGENTS.md — bm

## Purpose

`bm` is a Stremio **addon client** (not an addon server). Core packages live under `internal/` and are shared by the Cobra CLI, Bubble Tea TUI, and MCP stdio server.

## Layout

| Path | Role |
|------|------|
| `cmd/bm` | `main` → `internal/cli` |
| `internal/stremio` | HTTP client + types for manifest/catalog/meta/stream/subtitles |
| `internal/config` | XDG-backed TOML via Viper; seeds Cinemeta + Torrentio on first run |
| `internal/addons` | Install/list/remove; `PickSearchCatalog` helper |
| `internal/search` | TMDB path (if key set) else catalog search |
| `internal/streams` | Parallel fan-out + dedupe; `PlayableURL()` = HTTP or magnet |
| `internal/clipboard` | System clipboard |
| `internal/app` | Wiring / `Meta` helper |
| `internal/cli` | Cobra commands + `--json` |
| `internal/tui` | Bubble Tea + lipgloss |
| `internal/mcp` | `mark3labs/mcp-go` stdio server |

## Conventions

- **IMDB ids** are the primary id for streams (`tt…`, series episodes `tt…:S:E`).
- **CLI JSON** fields are stable for external launchers; avoid renaming JSON tags without a version bump note in the changelog.
- **Errors**: return wrapped errors from the `stremio` client; MCP tools should use `mcp.NewToolResultErrorFromErr` for tool-visible failures.

## Verification

```bash
gofmt -w .
go vet ./...
go build -o bm ./cmd/bm
```

Optional smoke:

```bash
XDG_CONFIG_HOME=/tmp/bm-smoke bm search test --type movie
XDG_CONFIG_HOME=/tmp/bm-smoke bm stream tt1375666 --type movie --json | head
```

## MCP notes

- `bm mcp` must keep **stdout clean** (only JSON-RPC). Use stderr for debug if ever needed.
- `resolve_imdb_id` requires `tmdb.api_key`.
