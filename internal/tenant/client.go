package tenant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Legion112/mcp-affise/internal/affiseapi"
	"github.com/Legion112/mcp-affise/internal/config"
)

// Clients maps tenant id → generated Affise API client.
type Clients struct {
	byID map[string]*affiseapi.ClientWithResponses
	meta map[string]config.Tenant
}

// NewClients builds HTTP clients for each configured tenant.
func NewClients(cfg *config.Config) (*Clients, error) {
	c := &Clients{
		byID: make(map[string]*affiseapi.ClientWithResponses, len(cfg.Tenants)),
		meta: make(map[string]config.Tenant, len(cfg.Tenants)),
	}
	for _, t := range cfg.Tenants {
		server := "https://" + t.Host
		cli, err := affiseapi.NewClientWithResponses(server, affiseapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("API-Key", t.Token)
			return nil
		}))
		if err != nil {
			return nil, fmt.Errorf("tenant %s: %w", t.ID, err)
		}
		c.byID[t.ID] = cli
		c.meta[t.ID] = t
	}
	return c, nil
}

// Get returns the generated client for a tenant id.
func (c *Clients) Get(tenantID string) (*affiseapi.ClientWithResponses, error) {
	cli, ok := c.byID[tenantID]
	if !ok {
		return nil, fmt.Errorf("unknown tenant %q", tenantID)
	}
	return cli, nil
}

// List returns tenant metadata (no tokens).
func (c *Clients) List() []config.Tenant {
	out := make([]config.Tenant, 0, len(c.meta))
	for _, t := range c.meta {
		out = append(out, config.Tenant{
			ID:          t.ID,
			Host:        t.Host,
			Description: t.Description,
		})
	}
	return out
}
