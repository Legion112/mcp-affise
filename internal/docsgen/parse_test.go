package docsgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOffersSlice(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "offers_slice.html"))
	if err != nil {
		t.Fatal(err)
	}
	ops, err := ParseHTML(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	ops = FilterPaths(ops, []string{"/3.0/offers", "/3.0/offer/{ID}"})
	if len(ops) < 2 {
		t.Fatalf("expected >=2 ops, got %d", len(ops))
	}

	var list, get *Operation
	for i := range ops {
		switch ops[i].Path {
		case "/3.0/offers":
			list = &ops[i]
		case "/3.0/offer/{ID}":
			get = &ops[i]
		}
	}
	if list == nil {
		t.Fatal("missing GET /3.0/offers")
	}
	if list.Method != "GET" {
		t.Fatalf("method=%s", list.Method)
	}
	if list.Summary == "" {
		t.Fatal("expected summary for offers list")
	}
	foundStatus := false
	for _, p := range list.Parameters {
		if p.Name == "status" {
			foundStatus = true
			if p.Type != "Array[string]" {
				t.Fatalf("status type=%s", p.Type)
			}
			if p.Default != "active" {
				t.Fatalf("status default=%q", p.Default)
			}
		}
	}
	if !foundStatus {
		t.Fatal("missing status param")
	}

	if get == nil {
		t.Fatal("missing GET /3.0/offer/{ID}")
	}
	hasPath := false
	for _, p := range get.Parameters {
		if p.In == "path" {
			hasPath = true
			if !p.Required {
				t.Fatal("path param should be required")
			}
		}
	}
	if !hasPath {
		t.Fatal("expected path parameter for {ID}")
	}
}

func TestParseFullDocsAllowlist(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "docs3.1.html"))
	if err != nil {
		t.Fatal(err)
	}
	ops, err := ParseHTML(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) < 100 {
		t.Fatalf("expected many ops, got %d", len(ops))
	}
	filtered := FilterPaths(ops, []string{"/3.0/offers", "/3.0/offer/{ID}"})
	if len(filtered) != 2 {
		t.Fatalf("filtered=%d want 2", len(filtered))
	}
	// exact /3.0/offers must not pull /3.0/offers/count
	onlyOffers := FilterPaths(ops, []string{"/3.0/offers"})
	if len(onlyOffers) != 1 || onlyOffers[0].Path != "/3.0/offers" {
		t.Fatalf("exact filter: %+v", onlyOffers)
	}
}

func TestMapAffiseType(t *testing.T) {
	typ, format, items := MapAffiseType("Array[integer]")
	if typ != "array" || items != "integer" {
		t.Fatalf("got %s %s %s", typ, format, items)
	}
	typ, format, _ = MapAffiseType("Integer")
	if typ != "integer" || format != "int64" {
		t.Fatalf("got %s %s", typ, format)
	}
}
