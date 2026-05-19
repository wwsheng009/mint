# Debug Environment Variables

This page is a compatibility entry for older links. The current source-of-truth debug variable reference is:

- [../debug/environment_variables.md](../debug/environment_variables.md)
- [../debug/README.md](../debug/README.md)
- [../log/LOGGER_ENV_VAR_STANDARD.md](../log/LOGGER_ENV_VAR_STANDARD.md)

## Current Summary

Mint's general logging system is implemented by `internal/log` and uses:

- `TUI_DEBUG_ALL`
- `TUI_DEBUG`
- `TUI_DEBUG_<CATEGORY>`
- `TUI_LOG_OUTPUT`
- `TUI_LOG_MAX_SIZE`
- `TUI_LOG_MAX_FILES`
- `TUI_LOG_COMPRESS`

Common runtime and render variables include:

- `MINT_ASYNC_RENDER`
- `MINT_ASYNC_FPS`
- `MINT_NO_ALTERNATE_SCREEN`
- `MINT_PORTAL_LAYOUT`
- `MINT_GRAPHICS`
- `MINT_CELL_PIXELS`
- `MINT_GRAPHICS_STRICT`
- `MINT_GRAPHICS_ALLOW_TERMINAL_FRAME`
- `MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE`

## Important Migration Notes

Older docs in this repository may mention variables such as `TUI_UI_DEBUG_*`, `TUI_RENDER_DEBUG`, `TUI_OUTPUT_DEBUG`, or `TUI_DEBUG_LOG`. Those are not the current general framework logging contract.

When in doubt, check:

- `../../internal/log/logger.go`
- `../../internal/log/file.go`
- `../../internal/log/rotation.go`
- `../../framework/app.go`
- `../../runtime/platform/graphics_env.go`
