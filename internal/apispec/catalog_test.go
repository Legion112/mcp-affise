package apispec

import (
	"strings"
	"testing"
)

func TestParseAndResolve(t *testing.T) {
	raw := []byte(`
openapi: "3.0.3"
paths:
  /3.0/offers:
    get:
      operationId: get_3_0_offers
      summary: list
      parameters:
        - name: limit
          in: query
  /3.0/offer/{ID}:
    get:
      operationId: get_3_0_offer_ID
      summary: one
      parameters:
        - name: ID
          in: path
  /3.0/admin/offer:
    post:
      operationId: post_ignored
`)
	cat, err := parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Endpoints) != 2 {
		t.Fatalf("got %d", len(cat.Endpoints))
	}
	ep, params, err := cat.Resolve("/3.0/offer/42")
	if err != nil {
		t.Fatal(err)
	}
	if ep.Path != "/3.0/offer/{ID}" || params["ID"] != "42" {
		t.Fatalf("%+v %+v", ep, params)
	}
	if _, _, err := cat.Resolve("/3.0/admin/offer"); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("expected unknown for POST-only path, got %v", err)
	}
}
