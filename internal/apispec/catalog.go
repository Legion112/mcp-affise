package apispec

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed affise.yaml
var embeddedSpec []byte

// Endpoint is one read (GET) operation from the OpenAPI catalog.
type Endpoint struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationID string   `json:"operation_id"`
	Summary     string   `json:"summary"`
	PathParams  []string `json:"path_params,omitempty"`
	QueryParams []string `json:"query_params,omitempty"`
}

// Catalog indexes GET endpoints from the generated OpenAPI spec.
type Catalog struct {
	Endpoints []Endpoint
	byPath    map[string]Endpoint // template path → endpoint
}

var (
	loadOnce sync.Once
	loaded   *Catalog
	loadErr  error
)

// Load returns the embedded Affise GET catalog.
func Load() (*Catalog, error) {
	loadOnce.Do(func() {
		loaded, loadErr = parse(embeddedSpec)
	})
	return loaded, loadErr
}

func parse(raw []byte) (*Catalog, error) {
	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `yaml:"operationId"`
			Summary     string `yaml:"summary"`
			Parameters  []struct {
				Name string `yaml:"name"`
				In   string `yaml:"in"`
			} `yaml:"parameters"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi catalog: %w", err)
	}
	cat := &Catalog{byPath: map[string]Endpoint{}}
	for path, item := range doc.Paths {
		op, ok := item["get"]
		if !ok {
			continue
		}
		ep := Endpoint{
			Method:      "GET",
			Path:        path,
			OperationID: op.OperationID,
			Summary:     op.Summary,
		}
		for _, p := range op.Parameters {
			switch strings.ToLower(p.In) {
			case "path":
				ep.PathParams = append(ep.PathParams, p.Name)
			case "query":
				ep.QueryParams = append(ep.QueryParams, p.Name)
			}
		}
		cat.Endpoints = append(cat.Endpoints, ep)
		cat.byPath[path] = ep
	}
	sort.Slice(cat.Endpoints, func(i, j int) bool {
		return cat.Endpoints[i].Path < cat.Endpoints[j].Path
	})
	if len(cat.Endpoints) == 0 {
		return nil, fmt.Errorf("openapi catalog has no GET operations")
	}
	return cat, nil
}

// Resolve matches a concrete request path (e.g. /3.0/offer/12) or template
// (/3.0/offer/{ID}) to a catalog endpoint. Returns path params extracted from concrete paths.
func (c *Catalog) Resolve(requestPath string) (Endpoint, map[string]string, error) {
	requestPath = strings.TrimSpace(requestPath)
	if requestPath == "" {
		return Endpoint{}, nil, fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}
	if ep, ok := c.byPath[requestPath]; ok {
		return ep, map[string]string{}, nil
	}
	reqParts := splitPath(requestPath)
	for _, ep := range c.Endpoints {
		tplParts := splitPath(ep.Path)
		if len(tplParts) != len(reqParts) {
			continue
		}
		params := map[string]string{}
		match := true
		for i := range tplParts {
			t, r := tplParts[i], reqParts[i]
			if strings.HasPrefix(t, "{") && strings.HasSuffix(t, "}") {
				name := strings.TrimSuffix(strings.TrimPrefix(t, "{"), "}")
				params[name] = r
				continue
			}
			if t != r {
				match = false
				break
			}
		}
		if match {
			return ep, params, nil
		}
	}
	return Endpoint{}, nil, fmt.Errorf("unknown or non-GET Affise path %q (see list_read_endpoints)", requestPath)
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}
