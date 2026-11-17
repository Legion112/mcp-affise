package tools

import "testing"

func TestBuildConcretePath(t *testing.T) {
	got, err := buildConcretePath("/3.0/offer/{ID}", map[string]any{"ID": 12})
	if err != nil || got != "/3.0/offer/12" {
		t.Fatalf("got %q %v", got, err)
	}
	got, err = buildConcretePath("/3.0/offer/{ID}", map[string]any{"id": 12})
	if err != nil || got != "/3.0/offer/12" {
		t.Fatalf("ci: %q %v", got, err)
	}
	if _, err := buildConcretePath("/3.0/offer/{ID}", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestEncodeQuery(t *testing.T) {
	v, err := encodeQuery(map[string]any{"limit": float64(1), "status": []any{"active", "stopped"}})
	if err != nil {
		t.Fatal(err)
	}
	if v.Get("limit") != "1" {
		t.Fatalf("%v", v)
	}
	if len(v["status"]) != 2 {
		t.Fatalf("%v", v)
	}
}

func TestParseJSONObject(t *testing.T) {
	m, err := parseJSONObject(`{"ID":"12","n":1}`)
	if err != nil || m["ID"] != "12" {
		t.Fatalf("%v %v", m, err)
	}
	m, err = parseJSONObject("")
	if err != nil || m != nil {
		t.Fatalf("%v %v", m, err)
	}
	if _, err := parseJSONObject("{"); err == nil {
		t.Fatal("expected error")
	}
}
