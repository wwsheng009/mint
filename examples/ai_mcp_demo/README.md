# AI MCP Demo

This example starts a small Mint UI app with the embedded MCP server enabled.

## Run

```bash
go run ./examples/ai_mcp_demo
```

## Notes

Endpoints are written to `tmp/ai_mcp_demo_endpoints.txt` to avoid breaking the TUI
layout with stdout/stderr output.

## Environment Options

- `MINT_MCP_TRANSPORT`: `http` (default) or `pipe`
- `MINT_MCP_HOST`: host for `http`, or socket/pipe path for `pipe`
- `MINT_MCP_PORT`: port for `http` (default `0` = random)
- `MINT_AI_TOKEN`: auth token (sent as `Authorization: Bearer <token>` or `X-Mint-AI-Token`)
- `MINT_MCP_EXPOSE_TREES`: `true`/`false`
- `MINT_MCP_EXPOSE_WRITE`: `true`/`false`
