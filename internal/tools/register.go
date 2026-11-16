package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Legion112/mcp-affise/internal/tenant"
)

// Deps holds shared dependencies for tool handlers.
type Deps struct {
	Clients *tenant.Clients
}

// Register adds Affise MCP tools (read-only).
func Register(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tenants",
		Description: "List configured Affise tenants (id, host, description; no tokens).",
	}, deps.listTenants)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_read_endpoints",
		Description: "List all read-only Affise GET endpoints from the generated OpenAPI catalog (optional q filter).",
	}, deps.listReadEndpoints)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "affise_get",
		Description: "Call any Affise GET endpoint (read-only). Path may be concrete (/3.0/offer/123) or a template with path_params. Rejects unknown/non-GET paths.",
	}, deps.affiseGet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_offers",
		Description: "Convenience: list Affise offers (GET /3.0/offers). Includes logo URL when present.",
	}, deps.listOffers)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_offer",
		Description: "Convenience: get one Affise offer by numeric id (GET /3.0/offer/{ID}).",
	}, deps.getOffer)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_offers_by_ids",
		Description: "Convenience: fetch multiple Affise offers by int ids (int_id on GET /3.0/offers).",
	}, deps.getOffersByIDs)
}
