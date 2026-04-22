# bm

`bm` is a Go CLI that speaks the [Stremio addon protocol](https://github.com/Stremio/stremio-addon-sdk): search catalogs, resolve streams, copy playable URLs (HTTP or `magnet:`) to the clipboard, and expose an **MCP server** for AI clients.

## Install

```bash
go install ./cmd/bm
# or
go build -o bm ./cmd/bm
```

## Quick start

Run the TUI (tabs **1–4** Browse / Streams / Addons / Settings; **Enter** on a stream copies to clipboard; **/** to search, **p**/**f** Cinemeta popular/featured):

```bash
bm
```

Search (Cinemeta catalog by default; set a TMDB key for richer search):

```bash
bm search "inception" --type movie
bm search "severance" --type series --json
```

Streams (aggregates every installed addon that exposes `stream`):

```bash
bm stream tt1375666 --type movie
bm stream tt0944947 --type series --season 1 --episode 1 --json
# shorthand for series:
bm stream tt0944947:1:1 --type series --json
```

Copy the first resolved URL:

```bash
bm stream tt1375666 --type movie --copy
```

Addons:

```bash
bm addons list
bm addons add 'https://example.com/manifest.json'
bm addons remove com.example.addon
```

Config (TOML under XDG config dir, e.g. `~/.config/bm/config.toml` on Linux):

```bash
bm config path
bm config set tmdb.api_key '<your TMDB v3 API key>'
bm config get tmdb.api_key
```

## Default addons

On first run, `bm` creates a config pre-seeded with:

- **Cinemeta** (`https://v3-cinemeta.strem.io/manifest.json`) — official catalog/meta.
- **Torrentio** (`https://torrentio.strem.fun/manifest.json`) — stream aggregation (torrents / debrid; **remove it** from `config.toml` if you do not want it).

## MCP (stdio)

Run the MCP server for Cursor / Claude / other MCP hosts:

```bash
bm mcp
```

**Tools:** `search_title`, `get_streams`, `list_addons`, `install_addon`, `remove_addon`, `get_meta`, `resolve_imdb_id`

**Resources:** `bm://config` (redacted), `bm://addons`

Example **Cursor** user MCP snippet:

```json
{
  "mcpServers": {
    "bm": {
      "command": "bm",
      "args": ["mcp"]
    }
  }
}
```

Use the absolute path to the `bm` binary if `bm` is not on `PATH`.

## JSON for launchers (future)

Subcommands support `--json` for machine-readable output (Raycast, PowerToys CmdPal, etc.).

## License

MIT (same spirit as upstream Stremio ecosystem tools you integrate with).
