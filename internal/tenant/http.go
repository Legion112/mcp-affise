package tenant

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GET performs an authenticated GET against a concrete Affise path (e.g. /3.0/offers?limit=1).
func (c *Clients) GET(ctx context.Context, tenantID, path string, query url.Values) (status int, body []byte, err error) {
	t, ok := c.meta[tenantID]
	if !ok {
		return 0, nil, fmt.Errorf("unknown tenant %q", tenantID)
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, nil, fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{
		Scheme:   "https",
		Host:     t.Host,
		Path:     path,
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("API-Key", t.Token)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err = io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}
