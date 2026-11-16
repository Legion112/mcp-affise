package tools

import (
	"encoding/json"
	"testing"

	"github.com/Legion112/mcp-affise/internal/affiseapi"
)

func TestJSONTextOffer(t *testing.T) {
	logo := "https://cdn.example/logo.png"
	id := int64(42)
	o := affiseapi.OffersListResponse{
		Offers: &[]affiseapi.Offer{{
			Id:   &id,
			Logo: &logo,
		}},
	}
	s, err := jsonText(o)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
}
