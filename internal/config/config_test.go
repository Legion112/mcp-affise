package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvKey(t *testing.T) {
	if got := EnvKey("my-traffgun"); got != "MY_TRAFFGUN" {
		t.Fatalf("got %s", got)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	tenants := filepath.Join(dir, "tenants.yaml")
	if err := os.WriteFile(tenants, []byte(`
tenants:
  - id: demo
    host: api.example.com
    description: Demo
`), 0o644); err != nil {
		t.Fatal(err)
	}
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("AFFISE_TENANT_DEMO_TOKEN=secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AFFISE_DOTENV", dotenv)
	t.Setenv("AFFISE_TENANTS_FILE", tenants)
	// clear host env
	t.Setenv("AFFISE_TENANT_DEMO_HOST", "")
	t.Setenv("AFFISE_TENANT_DEMO_TOKEN", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tenants) != 1 || cfg.Tenants[0].Host != "api.example.com" {
		t.Fatalf("%+v", cfg.Tenants)
	}
	ten, err := cfg.Tenant("demo")
	if err != nil || ten.Token != "secret" {
		t.Fatalf("%+v %v", ten, err)
	}
}
