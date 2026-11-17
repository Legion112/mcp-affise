# mcp-affise

[![CI](https://github.com/Legion112/mcp-affise/actions/workflows/ci.yml/badge.svg)](https://github.com/Legion112/mcp-affise/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Open-source **MCP** (Model Context Protocol) server for the [Affise](https://affise.com/) Admin API. Speaks **stdio** JSON-RPC for Cursor, Claude Desktop, and other MCP clients.

Affise does **not** publish OpenAPI JSON. This repo scrapes the official HTML docs (`docs3.1`), emits OpenAPI, applies hand-written **overlays** for docs↔live API drift, then generates a Go client with `oapi-codegen`.

Requires **Go 1.27+**. MIT licensed.

## Tools

Read-only (codegen includes **GET** endpoints only; no POST/PUT/DELETE tools).

| Tool | Description |
|------|-------------|
| `list_tenants` | Configured tenants (id/host/description; no tokens) |
| `list_read_endpoints` | Catalog of all Affise GET paths from OpenAPI (`q` filter) |
| `affise_get` | Call any catalogued GET (`path`, optional `path_params_json` / `query_json`) |
| `list_offers` | `GET /3.0/offers` |
| `get_offer` | `GET /3.0/offer/{ID}` (includes `partner_payments`) |
| `get_offers_by_ids` | Batch via `int_id` |
| `get_offer_privacy` | `GET /3.1/offers/{ID}/privacy` |
| `list_partner_postbacks` | `GET /3.0/admin/postbacks` (requires `partner_id` and/or `offer_id`) |

## Codegen pipeline

```text
HTML docs3.1 → (GET only) → affise.gen.yaml → overlays.yaml → affise.yaml
  → oapi-codegen → internal/affiseapi
  → copy → internal/apispec/affise.yaml (MCP catalog)
```

| Make target | What it does |
|-------------|--------------|
| `make fetch-docs` | Download Affise HTML docs fixture |
| `make openapi` | Parse HTML → gen → apply overlay → final YAML (`METHODS=GET`) |
| `make generate` | `openapi` + `oapi-codegen` + embeddable catalog copy |
| `make generate-check` | Fail if generated files are dirty |
| `make capture-sample TENANT=… PATH=/3.0/offers` | Dump live JSON for drafting overlays |

**Hand-edit only** [`openapi/overlays.yaml`](openapi/overlays.yaml). Never edit `*.gen.yaml`, `affise.yaml`, or `internal/affiseapi`.

## Configuration

1. Copy [`tenants.example.yaml`](tenants.example.yaml) → `tenants.yaml`
2. Copy [`.env.example`](.env.example) → `.env` and set the vars referenced from YAML

`host` / `token` support Symfony-style placeholders: `%env(VAR_NAME)%` (whole value). Literals work too.

```yaml
tenants:
  - id: example
    host: "%env(AFFISE_EXAMPLE_HOST)%"   # or: api.affise.com
    token: "%env(AFFISE_EXAMPLE_TOKEN)%"
    description: Example Affise tenant
```

| Variable | Meaning |
|----------|---------|
| `AFFISE_TENANTS_FILE` | Path to tenants YAML (default `tenants.yaml`) |
| `AFFISE_DOTENV` | Path to `.env` (default `.env`) |
| *(any)* | Only vars named inside `%env(...)%` in the YAML |

## Build

```bash
make generate
make build    # → bin/mcp-affise
make test
```

## Cursor

```json
{
  "mcpServers": {
    "affise": {
      "command": "/absolute/path/to/mcp-affise/bin/mcp-affise",
      "env": {
        "AFFISE_DOTENV": "/absolute/path/to/mcp-affise/.env",
        "AFFISE_TENANTS_FILE": "/absolute/path/to/mcp-affise/tenants.yaml"
      }
    }
  }
}
```

## License

MIT
