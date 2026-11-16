package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveEnv(t *testing.T) {
	t.Setenv("AFFISE_TEST_TOKEN", "secret")
	got, err := resolveEnv("%env(AFFISE_TEST_TOKEN)%")
	if err != nil || got != "secret" {
		t.Fatalf("got %q %v", got, err)
	}
	got, err = resolveEnv("api.example.com")
	if err != nil || got != "api.example.com" {
		t.Fatalf("literal: %q %v", got, err)
	}
	if _, err := resolveEnv("%env(AFFISE_MISSING_VAR)%"); err == nil {
		t.Fatal("expected missing env error")
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	tenants := filepath.Join(dir, "tenants.yaml")
	if err := os.WriteFile(tenants, []byte(`
tenants:
  - id: demo
    host: api.example.com
    token: "%env(AFFISE_DEMO_TOKEN)%"
    description: Demo
`), 0o644); err != nil {
		t.Fatal(err)
	}
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("AFFISE_DEMO_TOKEN=secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AFFISE_DOTENV", dotenv)
	t.Setenv("AFFISE_TENANTS_FILE", tenants)
	t.Setenv("AFFISE_DEMO_TOKEN", "secret")

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

func TestLoadEnvHostAndMissingToken(t *testing.T) {
	dir := t.TempDir()
	tenants := filepath.Join(dir, "tenants.yaml")
	if err := os.WriteFile(tenants, []byte(`
tenants:
  - id: demo
    host: "%env(AFFISE_DEMO_HOST)%"
    token: "%env(AFFISE_DEMO_TOKEN)%"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("AFFISE_DEMO_HOST=api.env.example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AFFISE_DOTENV", dotenv)
	t.Setenv("AFFISE_TENANTS_FILE", tenants)
	t.Setenv("AFFISE_DEMO_HOST", "api.env.example")
	t.Setenv("AFFISE_DEMO_TOKEN", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "AFFISE_DEMO_TOKEN") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}
