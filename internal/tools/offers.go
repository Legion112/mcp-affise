package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Legion112/mcp-affise/internal/affiseapi"
)

type listTenantsIn struct{}

type listOffersIn struct {
	Tenant string   `json:"tenant" jsonschema:"Tenant id from list_tenants"`
	Q      string   `json:"q,omitempty" jsonschema:"Search by title and id"`
	Status []string `json:"status,omitempty" jsonschema:"Offer statuses (default active)"`
	Page   int      `json:"page,omitempty" jsonschema:"Page number"`
	Limit  int      `json:"limit,omitempty" jsonschema:"Page size (max 500)"`
}

type getOfferIn struct {
	Tenant  string `json:"tenant" jsonschema:"Tenant id from list_tenants"`
	OfferID int64  `json:"offer_id" jsonschema:"Numeric Affise offer id"`
}

type getOffersByIDsIn struct {
	Tenant   string  `json:"tenant" jsonschema:"Tenant id from list_tenants"`
	OfferIDs []int64 `json:"offer_ids" jsonschema:"Numeric Affise offer ids"`
	Limit    int     `json:"limit,omitempty" jsonschema:"Page size (default len(ids) or 100)"`
}

func (d *Deps) listTenants(ctx context.Context, _ *mcp.CallToolRequest, _ listTenantsIn) (*mcp.CallToolResult, any, error) {
	_ = ctx
	text, err := jsonText(d.Clients.List())
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	return textResult(text), nil, nil
}

func (d *Deps) listOffers(ctx context.Context, _ *mcp.CallToolRequest, in listOffersIn) (*mcp.CallToolResult, any, error) {
	cli, err := d.Clients.Get(in.Tenant)
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	params := &affiseapi.Get30OffersParams{}
	if in.Q != "" {
		params.Q = &in.Q
	}
	if len(in.Status) > 0 {
		params.Status = &in.Status
	}
	if in.Page > 0 {
		p := int64(in.Page)
		params.Page = &p
	}
	if in.Limit > 0 {
		l := int64(in.Limit)
		params.Limit = &l
	}
	resp, err := cli.Get30OffersWithResponse(ctx, params)
	if err != nil {
		return textResult(fmt.Sprintf("list offers: %v", err)), nil, nil
	}
	if resp.StatusCode() >= 400 {
		return textResult(fmt.Sprintf("list offers: HTTP %d: %s", resp.StatusCode(), string(resp.Body))), nil, nil
	}
	if resp.JSON200 != nil {
		text, err := jsonText(resp.JSON200)
		if err != nil {
			return textResult(err.Error()), nil, nil
		}
		return textResult(text), nil, nil
	}
	return textResult(string(resp.Body)), nil, nil
}

func (d *Deps) getOffer(ctx context.Context, _ *mcp.CallToolRequest, in getOfferIn) (*mcp.CallToolResult, any, error) {
	cli, err := d.Clients.Get(in.Tenant)
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	if in.OfferID == 0 {
		return textResult("offer_id is required"), nil, nil
	}
	resp, err := cli.Get30OfferIDWithResponse(ctx, in.OfferID)
	if err != nil {
		return textResult(fmt.Sprintf("get offer: %v", err)), nil, nil
	}
	if resp.StatusCode() >= 400 {
		return textResult(fmt.Sprintf("get offer: HTTP %d: %s", resp.StatusCode(), string(resp.Body))), nil, nil
	}
	if resp.JSON200 != nil {
		text, err := jsonText(resp.JSON200)
		if err != nil {
			return textResult(err.Error()), nil, nil
		}
		return textResult(text), nil, nil
	}
	return textResult(string(resp.Body)), nil, nil
}

func (d *Deps) getOffersByIDs(ctx context.Context, _ *mcp.CallToolRequest, in getOffersByIDsIn) (*mcp.CallToolResult, any, error) {
	cli, err := d.Clients.Get(in.Tenant)
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	if len(in.OfferIDs) == 0 {
		return textResult("offer_ids is required"), nil, nil
	}
	ids := make([]int, len(in.OfferIDs))
	for i, id := range in.OfferIDs {
		ids[i] = int(id)
	}
	params := &affiseapi.Get30OffersParams{IntId: &ids}
	limit := int64(in.Limit)
	if limit <= 0 {
		limit = int64(len(ids))
		if limit < 100 {
			limit = 100
		}
	}
	params.Limit = &limit
	resp, err := cli.Get30OffersWithResponse(ctx, params)
	if err != nil {
		return textResult(fmt.Sprintf("get offers by ids: %v", err)), nil, nil
	}
	if resp.StatusCode() >= 400 {
		return textResult(fmt.Sprintf("get offers by ids: HTTP %d: %s", resp.StatusCode(), string(resp.Body))), nil, nil
	}
	if resp.JSON200 != nil {
		text, err := jsonText(resp.JSON200)
		if err != nil {
			return textResult(err.Error()), nil, nil
		}
		return textResult(text), nil, nil
	}
	return textResult(string(resp.Body)), nil, nil
}
