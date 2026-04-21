package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"bm/internal/app"
	"bm/internal/config"
)

const (
	uriConfig = "bm://config"
	uriAddons = "bm://addons"
)

// ServeStdio starts the MCP server on stdin/stdout.
func ServeStdio() error {
	srv := buildServer()
	return server.ServeStdio(srv)
}

func buildServer() *server.MCPServer {
	s := server.NewMCPServer(
		"bm",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
	)

	s.AddResource(mcp.NewResource(uriConfig, "bm configuration (redacted)"),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			a, err := app.New()
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(redactedConfigSnapshot(a.Config))
			if err != nil {
				return nil, err
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(b)},
			}, nil
		},
	)

	s.AddResource(mcp.NewResource(uriAddons, "installed Stremio addons"),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			a, err := app.New()
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(a.Addons.List(true))
			if err != nil {
				return nil, err
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(b)},
			}, nil
		},
	)

	s.AddTool(mcp.NewTool("search_title",
		mcp.WithDescription("Search movies/TV via TMDB (if api key set) or Stremio catalog addons"),
		mcp.WithString("query", mcp.Description("Title search query"), mcp.Required()),
		mcp.WithString("type", mcp.Description("movie or series (optional)")),
		mcp.WithNumber("year", mcp.Description("Release year filter (TMDB only)")),
	), toolSearchTitle)

	s.AddTool(mcp.NewTool("get_streams",
		mcp.WithDescription("Resolve playable URLs from stream addons"),
		mcp.WithString("imdb_id", mcp.Description("IMDB id e.g. tt1375666"), mcp.Required()),
		mcp.WithString("type", mcp.Description("movie or series"), mcp.Required()),
		mcp.WithNumber("season", mcp.Description("season number for series")),
		mcp.WithNumber("episode", mcp.Description("episode number for series")),
	), toolGetStreams)

	s.AddTool(mcp.NewTool("list_addons",
		mcp.WithDescription("List installed Stremio addons"),
	), toolListAddons)

	s.AddTool(mcp.NewTool("install_addon",
		mcp.WithDescription("Install addon by manifest.json URL"),
		mcp.WithString("manifest_url", mcp.Required()),
	), toolInstallAddon)

	s.AddTool(mcp.NewTool("remove_addon",
		mcp.WithDescription("Remove addon by manifest id"),
		mcp.WithString("id", mcp.Required()),
	), toolRemoveAddon)

	s.AddTool(mcp.NewTool("get_meta",
		mcp.WithDescription("Fetch Stremio meta document for a title"),
		mcp.WithString("imdb_id", mcp.Required()),
		mcp.WithString("type", mcp.Description("movie or series"), mcp.Required()),
	), toolGetMeta)

	s.AddTool(mcp.NewTool("resolve_imdb_id",
		mcp.WithDescription("Resolve IMDB id from natural language (requires TMDB api key)"),
		mcp.WithString("query", mcp.Required()),
		mcp.WithNumber("year", mcp.Description("optional year hint")),
	), toolResolveIMDB)

	return s
}

func redactedConfigSnapshot(cfg *config.Config) map[string]any {
	return map[string]any{
		"tmdb": map[string]any{
			"api_key_configured": strings.TrimSpace(cfg.TMDB.APIKey) != "",
		},
		"ui": map[string]any{
			"default_type": cfg.UI.DefaultType,
		},
		"addons_count": len(cfg.Addons),
	}
}

func toolSearchTitle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	q, _ := args["query"].(string)
	mediaType, _ := args["type"].(string)
	year := 0
	if y, ok := args["year"].(float64); ok {
		year = int(y)
	}
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	res, err := a.Search.Search(ctx, q, mediaType, year)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("search", err), nil
	}
	return mustJSON(res)
}

func toolGetStreams(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	imdb, _ := args["imdb_id"].(string)
	metaType, _ := args["type"].(string)
	season, episode := 0, 0
	if v, ok := args["season"].(float64); ok {
		season = int(v)
	}
	if v, ok := args["episode"].(float64); ok {
		episode = int(v)
	}
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	list, err := a.Streams.Resolve(ctx, imdb, metaType, season, episode)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("streams", err), nil
	}
	type row struct {
		AddonID     string `json:"addon_id"`
		AddonName   string `json:"addon_name"`
		Name        string `json:"name,omitempty"`
		Title       string `json:"title,omitempty"`
		URL         string `json:"url,omitempty"`
		InfoHash    string `json:"info_hash,omitempty"`
		PlayableURL string `json:"playable_url"`
	}
	out := make([]row, 0, len(list))
	for _, s := range list {
		out = append(out, row{
			AddonID:     s.AddonID,
			AddonName:   s.AddonName,
			Name:        s.Name,
			Title:       s.Title,
			URL:         s.URL,
			InfoHash:    s.InfoHash,
			PlayableURL: s.PlayableURL(),
		})
	}
	return mustJSON(out)
}

func toolListAddons(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	return mustJSON(a.Addons.List(true))
}

func toolInstallAddon(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	url, _ := args["manifest_url"].(string)
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	entry, err := a.Addons.Install(ctx, url)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("install", err), nil
	}
	return mustJSON(entry)
}

func toolRemoveAddon(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	if err := a.Addons.Remove(id); err != nil {
		return mcp.NewToolResultErrorFromErr("remove", err), nil
	}
	return mcp.NewToolResultJSON(map[string]string{"removed": id})
}

func toolGetMeta(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	imdb, _ := args["imdb_id"].(string)
	metaType, _ := args["type"].(string)
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	meta, err := a.Meta(ctx, imdb, metaType)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("meta", err), nil
	}
	return mustJSON(meta)
}

func toolResolveIMDB(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	q, _ := args["query"].(string)
	year := 0
	if y, ok := args["year"].(float64); ok {
		year = int(y)
	}
	a, err := app.New()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("app", err), nil
	}
	id, err := a.Search.ResolveIMDBID(ctx, q, year)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("resolve", err), nil
	}
	return mcp.NewToolResultJSON(map[string]string{"imdb_id": id})
}

func mustJSON(v any) (*mcp.CallToolResult, error) {
	r, err := mcp.NewToolResultJSON(v)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("json", err), nil
	}
	return r, nil
}
