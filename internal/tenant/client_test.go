package tenant

import (
	"testing"

	"github.com/Legion112/mcp-affise/internal/config"
)

func TestNewClientsList(t *testing.T) {
	cfg := &config.Config{Tenants: []config.Tenant{{
		ID:          "demo",
		Host:        "api.example.com",
		Token:       "tok",
		Description: "d",
	}}}
	c, err := NewClients(cfg)
	if err != nil {
		t.Fatal(err)
	}
	list := c.List()
	if len(list) != 1 || list[0].Token != "" || list[0].ID != "demo" {
		t.Fatalf("%+v", list)
	}
	if _, err := c.Get("missing"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := c.Get("demo"); err != nil {
		t.Fatal(err)
	}
}
