# NoAlternateScreen Option

## What is Alternate Screen?

**Alternate Screen** (`\x1b[?1049h`) is a terminal feature that switches to a separate screen buffer:
- Original terminal content is hidden and preserved
- TUI app renders in the alternate buffer
- On exit, original content is restored
- **Trade-off**: Cannot select/copy text, no scrollback

**Normal Screen** (with `NoAlternateScreen`):
- TUI renders in normal terminal buffer
- Output becomes part of scrollback history
- Can select/copy text with mouse
- Content persists after app exits

## Usage

### Basic Example

```go
import "github.com/wwsheng009/mint/ui"

ui.Run(
    myApp,
    ui.WithWidth(80),
    ui.WithHeight(24),
    ui.WithNoAlternateScreen(), // Enable normal screen mode
)
```

### When to Use

| Mode | Use Case |
|------|----------|
| Default (Alternate Screen) | Interactive apps, games, dashboards |
| NoAlternateScreen | Debugging, logging, demos, output inspection |

## Differences

| Feature | Default | NoAlternateScreen |
|---------|---------|-------------------|
| Screen clearing | Yes (`\x1b[2J`) | No |
| Content on exit | Cleared | Preserved |
| Mouse selection | No | Yes |
| Scrollback | No | Yes |
| Copy/paste | Limited | Full terminal support |

## Implementation

The option sets `MINT_NO_ALTERNATE_SCREEN=true` environment variable, which is checked by:
1. `ui/app.go` - Sets the env var from option
2. `framework/app.go` - Skips screen clearing when set

## Files Modified

- `ui/app.go` - Added `WithNoAlternateScreen()` option
- `framework/app.go` - Check env var, skip clear screen
