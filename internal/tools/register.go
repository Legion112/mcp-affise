package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Legion112/mcp-affise/internal/tenant"
)

// Deps holds shared dependencies for tool handlers.
type Deps struct {
	Clients *tenant.Clients
}

// Register adds Affise MCP tools.
func Register(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tenants",
		Description: "List configured Affise tenants (id, host, description; no tokens).",
	}, deps.listTenants)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_offers",
		Description: "List Affise offers for a tenant (GET /3.0/offers). Includes logo URL when present.",
	}, deps.listOffers)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_offer",
		Description: "Get one Affise offer by numeric id (GET /3.0/offer/{ID}).",
	}, deps.getOffer)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_offers_by_ids",
		Description: "Fetch multiple Affise offers by int ids (uses int_id filter on GET /3.0/offers).",
	}, deps.getOffersByIDs)
}
