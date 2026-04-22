# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, test, lint

```bash
make          # build binary to bin/bm
make build    # same
make test     # go test ./...
make vet      # go vet ./...
make fmt      # gofmt -w .
make check    # fmt, vet, test combined
make run      # go run ./cmd/bm
```

Smoke test (config written to temp dir):

```bash
XDG_CONFIG_HOME=/tmp/bm-smoke bm search test --type movie
XDG_CONFIG_HOME=/tmp/bm-smoke bm stream tt1375666 --type movie --json | head
```

## Architecture

`bm` is a Stremio **addon client** (not an addon server). Core packages live under `internal/` and are shared by the Cobra CLI, Bubble Tea TUI, and MCP stdio server.


| Path                 | Role                                                               |
| -------------------- | ------------------------------------------------------------------ |
| `cmd/bm/main.go`     | Entry point → `internal/cli`                                       |
| `internal/stremio`   | HTTP client + types for manifest/catalog/meta/stream/subtitles     |
| `internal/config`    | XDG-backed TOML via Viper; seeds Cinemeta + Torrentio on first run |
| `internal/addons`    | Install/list/remove; `PickSearchCatalog` helper                    |
| `internal/search`    | TMDB path (if key set) else catalog search                         |
| `internal/streams`   | Parallel fan-out + dedupe; `PlayableURL()` = HTTP or magnet        |
| `internal/clipboard` | System clipboard                                                   |
| `internal/app`       | Wiring / `Meta` helper                                             |
| `internal/cli`       | Cobra commands + `--json`                                          |
| `internal/tui`       | Bubble Tea + lipgloss                                              |
| `internal/mcp`       | `mark3labs/mcp-go` stdio server                                    |


## Conventions

- **IMDB ids** are the primary id for streams (`tt…`, series episodes `tt…:S:E`).
- **CLI JSON** fields are stable for external launchers; avoid renaming JSON tags without a version bump.
- **Errors**: return wrapped errors from the `stremio` client; MCP tools use `mcp.NewToolResultErrorFromErr`.
- **MCP stdout** must stay clean (JSON-RPC only); debug goes to stderr.