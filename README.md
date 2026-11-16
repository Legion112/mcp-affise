# mcp-affise

[![CI](https://github.com/Legion112/mcp-affise/actions/workflows/ci.yml/badge.svg)](https://github.com/Legion112/mcp-affise/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Open-source **MCP** (Model Context Protocol) server for the [Affise](https://affise.com/) Admin API. Speaks **stdio** JSON-RPC for Cursor, Claude Desktop, and other MCP clients.

Affise does **not** publish OpenAPI JSON. This repo scrapes the official HTML docs (`docs3.1`), emits OpenAPI, applies hand-written **overlays** for docs↔live API drift, then generates a Go client with `oapi-codegen`.

Requires **Go 1.27+**. MIT licensed.

## Tools

| Tool | Description |
|------|-------------|
| `list_tenants` | Configured tenants (id/host/description; no tokens) |
| `list_offers` | `GET /3.0/offers` (optional `q`, `status`, `page`, `limit`) |
| `get_offer` | `GET /3.0/offer/{ID}` |
| `get_offers_by_ids` | Batch via `int_id` filter |

## Codegen pipeline

```text
HTML docs3.1 → affise.gen.yaml → overlays.yaml → affise.yaml → oapi-codegen → internal/affiseapi
```

| Make target | What it does |
|-------------|--------------|
| `make fetch-docs` | Download Affise HTML docs fixture |
| `make openapi` | Parse HTML → gen → apply overlay → final YAML |
| `make generate` | `openapi` + `oapi-codegen` |
| `make generate-check` | Fail if generated files are dirty |
| `make capture-sample TENANT=… PATH=/3.0/offers` | Dump live JSON for drafting overlays |

**Hand-edit only** [`openapi/overlays.yaml`](openapi/overlays.yaml). Never edit `*.gen.yaml`, `affise.yaml`, or `internal/affiseapi`.

## Configuration

1. Copy [`tenants.example.yaml`](tenants.example.yaml) → `tenants.yaml`
2. Copy [`.env.example`](.env.example) → `.env` and set tokens

| Variable | Meaning |
|----------|---------|
| `AFFISE_TENANTS_FILE` | Path to tenants YAML (default `tenants.yaml`) |
| `AFFISE_TENANT_<ID>_HOST` | API host (overrides YAML) |
| `AFFISE_TENANT_<ID>_TOKEN` | API-Key (required) |
| `AFFISE_DOTENV` | Path to `.env` |

`<ID>` is the tenant id uppercased with non-alnum → `_` (e.g. `my-traffgun` → `MY_TRAFFGUN`).

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
