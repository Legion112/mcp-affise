package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/Legion112/mcp-affise/internal/config"
	"github.com/Legion112/mcp-affise/internal/docsgen"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "openapi":
		err = cmdOpenAPI(os.Args[2:])
	case "capture-sample":
		err = cmdCapture(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "affise-gen: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `affise-gen — Affise HTML docs → OpenAPI (+ overlays)

Usage:
  affise-gen openapi -i docs.html -o openapi/affise.gen.yaml --overlay openapi/overlays.yaml --final openapi/affise.yaml
  affise-gen capture-sample --tenant ID --path /3.0/offers --out openapi/captures
`)
}

func cmdOpenAPI(args []string) error {
	fs := flag.NewFlagSet("openapi", flag.ContinueOnError)
	in := fs.String("i", "", "input HTML docs path")
	out := fs.String("o", "openapi/affise.gen.yaml", "generated OpenAPI (docs only)")
	overlayPath := fs.String("overlay", "openapi/overlays.yaml", "overlay file")
	finalPath := fs.String("final", "openapi/affise.yaml", "final OpenAPI after overlay")
	paths := fs.String("paths", "", "comma-separated path prefixes allowlist")
	title := fs.String("title", "Affise API", "OpenAPI info.title")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" {
		return fmt.Errorf("-i is required")
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	ops, err := docsgen.ParseHTML(string(raw))
	if err != nil {
		return err
	}
	var allow []string
	for _, p := range strings.Split(*paths, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			allow = append(allow, p)
		}
	}
	if len(allow) > 0 {
		ops = docsgen.FilterPaths(ops, allow)
		if len(ops) == 0 {
			return fmt.Errorf("no operations left after --paths filter")
		}
	}
	gen, err := docsgen.EmitOpenAPI(ops, *title)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*out, gen, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d ops)\n", *out, len(ops))

	ov, err := docsgen.LoadOverlay(*overlayPath)
	if err != nil {
		return err
	}
	final, err := docsgen.ApplyOverlay(gen, ov)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*finalPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*finalPath, final, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d overlay actions)\n", *finalPath, len(ov.Actions))
	return nil
}

func cmdCapture(args []string) error {
	fs := flag.NewFlagSet("capture-sample", flag.ContinueOnError)
	tenant := fs.String("tenant", "", "tenant id")
	path := fs.String("path", "", "API path e.g. /3.0/offers")
	outDir := fs.String("out", "openapi/captures", "output directory")
	query := fs.String("query", "limit=1", "raw query string")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tenant == "" || *path == "" {
		return fmt.Errorf("--tenant and --path are required")
	}
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	t, err := cfg.Tenant(*tenant)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("https://%s%s", t.Host, *path)
	if *query != "" {
		u += "?" + *query
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("API-Key", t.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, *path)
	name := fmt.Sprintf("%s_%s_%s.json", *tenant, safe, time.Now().UTC().Format("20060102T150405Z"))
	outPath := filepath.Join(*outDir, name)
	meta := map[string]any{
		"tenant":     *tenant,
		"url":        u,
		"status":     resp.StatusCode,
		"capturedAt": time.Now().UTC().Format(time.RFC3339),
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		parsed = string(body)
	}
	payload := map[string]any{"meta": meta, "body": parsed}
	enc, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, enc, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes, http %d)\n", outPath, len(body), resp.StatusCode)
	return nil
}
