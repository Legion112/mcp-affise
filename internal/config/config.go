package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Tenant is one Affise API host + credentials.
type Tenant struct {
	ID          string
	Host        string
	Token       string
	Description string
}

// Config holds multi-tenant Affise settings.
type Config struct {
	Tenants []Tenant
}

type tenantsFile struct {
	Tenants []struct {
		ID          string `yaml:"id"`
		Host        string `yaml:"host"`
		Description string `yaml:"description"`
	} `yaml:"tenants"`
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

// EnvKey returns AFFISE_TENANT_<ID>_TOKEN style suffix from tenant id.
func EnvKey(tenantID string) string {
	s := strings.ToUpper(nonAlnum.ReplaceAllString(tenantID, "_"))
	s = strings.Trim(s, "_")
	return s
}

// Load reads optional .env, tenants YAML, and per-tenant env credentials.
func Load() (*Config, error) {
	dotenvPath := os.Getenv("AFFISE_DOTENV")
	if dotenvPath == "" {
		dotenvPath = ".env"
	}
	if err := godotenv.Load(dotenvPath); err != nil && !isMissingFile(err) {
		return nil, fmt.Errorf("load dotenv %s: %w", dotenvPath, err)
	}

	path := os.Getenv("AFFISE_TENANTS_FILE")
	if path == "" {
		path = "tenants.yaml"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tenants file %s not found (copy tenants.example.yaml)", path)
		}
		return nil, err
	}
	var tf tenantsFile
	if err := yaml.Unmarshal(raw, &tf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(tf.Tenants) == 0 {
		return nil, fmt.Errorf("%s: no tenants defined", path)
	}

	cfg := &Config{}
	for _, t := range tf.Tenants {
		id := strings.TrimSpace(t.ID)
		if id == "" {
			return nil, fmt.Errorf("%s: tenant with empty id", path)
		}
		key := EnvKey(id)
		host := strings.TrimSpace(os.Getenv("AFFISE_TENANT_" + key + "_HOST"))
		if host == "" {
			host = strings.TrimSpace(t.Host)
		}
		host = normalizeHost(host)
		token := strings.TrimSpace(os.Getenv("AFFISE_TENANT_" + key + "_TOKEN"))
		if host == "" {
			return nil, fmt.Errorf("tenant %s: host missing (YAML or AFFISE_TENANT_%s_HOST)", id, key)
		}
		if token == "" {
			return nil, fmt.Errorf("tenant %s: missing AFFISE_TENANT_%s_TOKEN", id, key)
		}
		cfg.Tenants = append(cfg.Tenants, Tenant{
			ID:          id,
			Host:        host,
			Token:       token,
			Description: strings.TrimSpace(t.Description),
		})
	}
	return cfg, nil
}

// Tenant returns a tenant by id.
func (c *Config) Tenant(id string) (Tenant, error) {
	if c == nil {
		return Tenant{}, fmt.Errorf("config is nil")
	}
	for _, t := range c.Tenants {
		if t.ID == id {
			return t, nil
		}
	}
	return Tenant{}, fmt.Errorf("unknown tenant %q", id)
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimSuffix(host, "/")
}

func isMissingFile(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "no such file")
}
