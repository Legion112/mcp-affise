package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Legion112/mcp-affise/internal/apispec"
)

type listReadEndpointsIn struct {
	Q string `json:"q,omitempty" jsonschema:"Optional substring filter on path or summary"`
}

type affiseGetIn struct {
	Tenant     string         `json:"tenant" jsonschema:"Tenant id from list_tenants"`
	Path       string         `json:"path" jsonschema:"Affise path template or concrete path, e.g. /3.0/offers or /3.0/offer/123"`
	PathParams map[string]any `json:"path_params,omitempty" jsonschema:"Path parameter values when path is a template with {param}"`
	Query      map[string]any `json:"query,omitempty" jsonschema:"Query string parameters (strings, numbers, bools, or arrays)"`
}

func (d *Deps) listReadEndpoints(ctx context.Context, _ *mcp.CallToolRequest, in listReadEndpointsIn) (*mcp.CallToolResult, any, error) {
	_ = ctx
	cat, err := apispec.Load()
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	q := strings.ToLower(strings.TrimSpace(in.Q))
	var out []apispec.Endpoint
	for _, ep := range cat.Endpoints {
		if q == "" || strings.Contains(strings.ToLower(ep.Path), q) || strings.Contains(strings.ToLower(ep.Summary), q) || strings.Contains(strings.ToLower(ep.OperationID), q) {
			out = append(out, ep)
		}
	}
	text, err := jsonText(map[string]any{
		"count":     len(out),
		"note":      "Read-only catalog (GET). Use affise_get to call an endpoint.",
		"endpoints": out,
	})
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	return textResult(text), nil, nil
}

func (d *Deps) affiseGet(ctx context.Context, _ *mcp.CallToolRequest, in affiseGetIn) (*mcp.CallToolResult, any, error) {
	cat, err := apispec.Load()
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	concrete, err := buildConcretePath(in.Path, in.PathParams)
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	if _, _, err := cat.Resolve(concrete); err != nil {
		return textResult(err.Error()), nil, nil
	}
	qv, err := encodeQuery(in.Query)
	if err != nil {
		return textResult(err.Error()), nil, nil
	}
	status, body, err := d.Clients.GET(ctx, in.Tenant, concrete, qv)
	if err != nil {
		return textResult(fmt.Sprintf("affise_get: %v", err)), nil, nil
	}
	if status >= 400 {
		return textResult(fmt.Sprintf("affise_get: HTTP %d: %s", status, string(body))), nil, nil
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return textResult(string(body)), nil, nil
	}
	text, err := jsonText(parsed)
	if err != nil {
		return textResult(string(body)), nil, nil
	}
	return textResult(text), nil, nil
}

func buildConcretePath(path string, pathParams map[string]any) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.Contains(path, "{") {
		return path, nil
	}
	out := path
	for k, v := range pathParams {
		repl := fmt.Sprint(v)
		out = strings.ReplaceAll(out, "{"+k+"}", repl)
	}
	// retry case-insensitive placeholder names
	if strings.Contains(out, "{") {
		for k, v := range pathParams {
			repl := fmt.Sprint(v)
			out = replacePathParamCI(out, k, repl)
		}
	}
	if strings.Contains(out, "{") {
		return "", fmt.Errorf("missing path_params for %q (still has placeholders: %s)", path, out)
	}
	return out, nil
}

func replacePathParamCI(path, name, value string) string {
	start := 0
	for {
		i := strings.Index(path[start:], "{")
		if i < 0 {
			return path
		}
		i += start
		j := strings.Index(path[i:], "}")
		if j < 0 {
			return path
		}
		j += i
		inner := path[i+1 : j]
		if strings.EqualFold(inner, name) {
			path = path[:i] + value + path[j+1:]
			start = i + len(value)
			continue
		}
		start = j + 1
	}
}

func encodeQuery(q map[string]any) (url.Values, error) {
	out := url.Values{}
	if len(q) == 0 {
		return out, nil
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := q[k]
		switch t := v.(type) {
		case nil:
			continue
		case string:
			out.Add(k, t)
		case float64:
			if t == float64(int64(t)) {
				out.Add(k, strconv.FormatInt(int64(t), 10))
			} else {
				out.Add(k, strconv.FormatFloat(t, 'f', -1, 64))
			}
		case bool:
			out.Add(k, strconv.FormatBool(t))
		case []any:
			for _, item := range t {
				out.Add(k, fmt.Sprint(item))
			}
		default:
			out.Add(k, fmt.Sprint(t))
		}
	}
	return out, nil
}
