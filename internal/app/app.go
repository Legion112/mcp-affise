package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Legion112/mcp-affise/internal/config"
	"github.com/Legion112/mcp-affise/internal/tenant"
	"github.com/Legion112/mcp-affise/internal/tools"
)

const version = "v0.2.0"

// Run loads config and serves MCP over stdio.
func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	clients, err := tenant.NewClients(cfg)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("mcp-affise starting", "tenants", len(cfg.Tenants), "version", version)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-affise",
		Version: version,
	}, nil)

	tools.Register(server, &tools.Deps{Clients: clients})
	return server.Run(ctx, &mcp.StdioTransport{})
}
