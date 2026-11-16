package docsgen

import (
	"strings"
	"testing"
)

func TestApplyOverlayMergeAndStale(t *testing.T) {
	base := []byte(`
openapi: "3.0.3"
info:
  title: Affise API
  version: "1"
paths:
  /3.0/offers:
    get:
      operationId: get_3_0_offers
      responses:
        "200":
          description: ok
components:
  schemas:
    AffiseObject:
      type: object
      additionalProperties: true
`)
	overlay := &Overlay{
		Overlay: "1.0.0",
		Actions: []OverlayAction{
			{
				Target: "$.components.schemas",
				Update: map[string]any{
					"Offer": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":   map[string]any{"type": "integer"},
							"logo": map[string]any{"type": "string"},
						},
					},
				},
			},
		},
	}
	out, err := ApplyOverlay(base, overlay)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "logo") {
		t.Fatalf("expected logo in output:\n%s", out)
	}

	stale := &Overlay{
		Actions: []OverlayAction{{
			Target: "$.paths['/no/such'].get",
			Update: map[string]any{"summary": "x"},
		}},
	}
	if _, err := ApplyOverlay(base, stale); err == nil {
		t.Fatal("expected stale target error")
	}
}

func TestEmitAndOverlayRoundTrip(t *testing.T) {
	ops := []Operation{{
		Method:  "GET",
		Path:    "/3.0/offers",
		Summary: "Get a list of offers",
		Parameters: []Parameter{{
			Name: "limit",
			In:   "query",
			Type: "Integer",
		}},
	}}
	gen, err := EmitOpenAPI(ops, "Affise API")
	if err != nil {
		t.Fatal(err)
	}
	overlay := &Overlay{Actions: []OverlayAction{{
		Target: "$.components.schemas",
		Update: map[string]any{
			"Offer": map[string]any{"type": "object"},
		},
	}}}
	final, err := ApplyOverlay(gen, overlay)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(final), "Offer") {
		t.Fatal(string(final))
	}
}
