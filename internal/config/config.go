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
		Token       string `yaml:"token"`
		Description string `yaml:"description"`
	} `yaml:"tenants"`
}

// %env(VAR_NAME)% — Symfony-style env placeholder (whole-value only).
var envPlaceholder = regexp.MustCompile(`^%env\(([A-Za-z_][A-Za-z0-9_]*)\)%$`)

// Load reads optional .env then tenants YAML, resolving %env(VAR)% in host/token.
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
		host, err := resolveEnv(t.Host)
		if err != nil {
			return nil, fmt.Errorf("tenant %s host: %w", id, err)
		}
		host = normalizeHost(host)
		token, err := resolveEnv(t.Token)
		if err != nil {
			return nil, fmt.Errorf("tenant %s token: %w", id, err)
		}
		token = strings.TrimSpace(token)
		if host == "" {
			return nil, fmt.Errorf("tenant %s: host is empty", id)
		}
		if token == "" {
			return nil, fmt.Errorf("tenant %s: token is empty", id)
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

// resolveEnv expands a whole-value %env(VAR)% placeholder, or returns the literal.
func resolveEnv(s string) (string, error) {
	s = strings.TrimSpace(s)
	m := envPlaceholder.FindStringSubmatch(s)
	if m == nil {
		return s, nil
	}
	name := m[1]
	val := strings.TrimSpace(os.Getenv(name))
	if val == "" {
		return "", fmt.Errorf("environment variable %s is not set or empty", name)
	}
	return val, nil
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
