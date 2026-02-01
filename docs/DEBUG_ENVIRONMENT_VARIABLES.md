# Debug Environment Variables

This document lists all environment variables used to control debug output in the Mint TUI Framework.

## General Debug Control

| Environment Variable | Description | Default | Files Using It |
|-------------------|-------------|----------|-----------------|
| `TUI_DEBUG` | Enable general debug mode in framework app | `false` | `framework/app.go` |
| `TUI_DEBUG_LOG` | Path to debug log file | `""` | `framework/app.go` |

## Output Control

| Environment Variable | Description | Values | Files Using It |
|-------------------|-------------|----------|-----------------|
| `TUI_OUTPUT_MODE` | Output mode for rendering | `direct`, `diff`, `debug` | `framework/app.go` |
| `TUI_OUTPUT_DEBUG` | Enable detailed output debug information | `1` | `framework/app.go`, `runtime/paint/renderer.go`, `runtime/paint/dirty.go` |

## Platform-Specific Debug

| Environment Variable | Description | Default | Files Using It |
|-------------------|-------------|----------|-----------------|
| `TUI_DEBUG_WINDOWS` | Enable Windows platform debug output | `false` | `runtime/platform/input_windows.go` |
| `TUI_DEBUG_EVENTS` | Enable raw input event debug output | `false` | `runtime/platform/input_windows.go` |
| `TUI_DEBUG_MOUSE` | Enable mouse event debug output | `false` | `runtime/platform/input_windows.go` |

## Component-Specific Debug

| Environment Variable | Description | Default | Files Using It |
|-------------------|-------------|----------|-----------------|
| `TUI_DEBUG_PUMP` | Enable event pump debug output | `false` | `framework/event/pump.go` |
| `TUI_DEBUG_UI` | Enable UI framework debug output | `false` | `ui/app.go` |
| `TUI_RENDER_DEBUG` | Enable rendering debug output | `false` | `runtime/paint/renderer.go`, `runtime/paint/dirty.go` |
| `TUI_FORM_DEBUG` | Enable form component debug output | `false` | `framework/form/form.go` |
| `TUI_CURSOR_DEBUG` | Enable cursor component debug output | `false` | `framework/cursor/cursor.go` |
| `TUI_INPUT_DEBUG` | Enable input component debug output | `false` | `framework/input/textinput.go` |
| `TUI_INPUT_DEBUG_FILE` | Path to input debug log file | `""` | `framework/input/textinput.go` |

## Usage Examples

### Enable general debug mode
```bash
export TUI_DEBUG=true
./your-app
```

### Enable Windows platform debug
```bash
export TUI_DEBUG_WINDOWS=true
./your-app
```

### Enable rendering debug
```bash
export TUI_RENDER_DEBUG=1
./your-app
```

### Enable detailed output debug
```bash
export TUI_OUTPUT_MODE=debug
export TUI_OUTPUT_DEBUG=1
./your-app
```

### Enable mouse event debug
```bash
export TUI_DEBUG_MOUSE=true
./your-app
```

### Multiple debug flags
```bash
export TUI_DEBUG=true
export TUI_DEBUG_WINDOWS=true
export TUI_RENDER_DEBUG=1
export TUI_OUTPUT_DEBUG=1
./your-app
```

## Notes

- Most debug variables use `"true"` or `"1"` as the enabled value
- Debug output goes to `stderr` to avoid interfering with the TUI display
- Some debug outputs (like `TUI_INPUT_DEBUG`) write to separate log files
- Check individual component documentation for specific debug information
