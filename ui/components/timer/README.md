# Timer

`timer` is a Fiber-first timer and countdown display for operational TUI status.
It is designed for auto-refresh countdowns, retry windows, uptime, SLA windows,
and short async operation timers.

## Capabilities

- Elapsed mode: time since `StartedAt(...)`.
- Countdown mode: time until `Deadline(...)` or `StartedAt(...) + Duration(...)`.
- Live ticking through `TickableInstance`.
- Static deterministic rendering with `Live(false)` or `Static()`.
- Optional Unicode progress bar for countdown or bounded elapsed windows.
- ASCII progress fallback through `ASCIIProgress()` or `ProgressGlyphStyle(ProgressGlyphStyleASCII)`.
- Semantic styles for normal, warning, and expired states.
- Fixed width rendering for status bars and dense toolbars.
- Operational presets for auto-refresh, retry-after/cooldown, and operation elapsed timers.
- SDK entry through `ui.NewTimerBuilder()`.

## Examples

```go
ui.NewTimerBuilder().
    Label("Refresh").
    Countdown(30 * time.Second).
    ShowProgress(true).
    Build()
```

ASCII progress fallback:

```go
ui.NewTimerBuilder().
    Label("Refresh").
    Countdown(30 * time.Second).
    ShowProgress(true).
    ASCIIProgress().
    Build()
```

```go
ui.NewTimerBuilder().
    Label("Uptime").
    Elapsed().
    StartedAt(startedAt).
    Build()
```

```go
ui.NewTimerBuilder().
    Label("Retry").
    Until(retryAfter).
    WarningBelow(5 * time.Second).
    ExpiredText("ready").
    Width(18).
    Build()
```

Operational presets:

```go
ui.AutoRefreshTimer("Refresh", 30*time.Second)
ui.RetryAfterTimer("Retry", retryAfter)
ui.OperationElapsedTimer("Reload", startedAt)
```

Component package helpers expose the same presets as `timer.AutoRefresh(...)`, `timer.RetryAfter(...)`, and `timer.OperationElapsed(...)`.

## Fiber-first shape

- `VNode` is the immutable declarative description.
- `Instance` owns runtime time state.
- `Instance` implements `PaintableInstance` and `TickableInstance`.
- No component state is stored in package globals.

## Tests

- Unit: `go test ./ui/components/timer`
- SDK shortcut: `go test ./ui -run Timer`
- E2E: `go test ./ui/e2e -run "^TestE2ETimer"`
