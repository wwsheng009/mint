# Mint DevTools

This directory is the current DevTools documentation entry point. Historical implementation plans, phase summaries, reviews, and old visualization notes were archived to:

```text
docsArchive/cleanup-2026-05-19/_devtools_docs/
```

Use the archived files for background only. Current behavior should be verified against the source under `devtools/` and the package tests.

## Purpose

DevTools provides runtime observability for Mint applications and framework work. The public entry point is `github.com/wwsheng009/mint/devtools`.

Core capabilities:

| Capability | Source | Notes |
|---|---|---|
| Frame lifecycle | `../devtools.go` | `BeginFrame()` / `EndFrame()` delimit runtime work. |
| Event collection | `../devtools.go`, `../collector.go` | `RecordEvent(...)` captures input or framework events. |
| Layout and repaint collection | `../devtools.go`, `../collector.go` | `CollectLayout(...)` and `CollectRepaint(...)` emit frame diagnostics. |
| Causal graph | `../causal.go`, `../causal_builder.go`, `../causal_query.go` | Connects events, mutations, layout changes, and repaint work. |
| Snapshot and diff | `../snapshot/` | Captures component state and compares frames. |
| Replay and time travel | `../replay/`, `../timetravel/` | Supports deterministic replay and snapshot navigation. |
| Remote inspection | `../remote/`, `../protocol/` | Exposes browser-oriented HTTP and protocol views. |
| Standalone debugging | `../standalone/` | Reads debug logs outside the target process. |

## Minimal Embedded Usage

```go
package main

import "github.com/wwsheng009/mint/devtools"

func main() {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	dt.BeginFrame()
	dt.RecordEvent("keypress", "search-input", "bubble", map[string]interface{}{
		"key": "Enter",
	})
	dt.EndFrame()
}
```

In a real Mint application or framework integration, call `BeginFrame()` before frame work starts, emit event/layout/repaint diagnostics as they happen, and call `EndFrame()` after the frame is complete.

## Remote HTTP Debugging

The remote module can expose an Inspector page and JSON APIs:

```go
package main

import (
	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/remote"
	"github.com/wwsheng009/mint/devtools/snapshot"
)

func main() {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	snapshots := snapshot.NewManager(100)
	server := remote.NewHTTPServer(9222, dt, snapshots)
	server.StartInBackground()

	// Application loop continues here.
}
```

Useful endpoints:

| Endpoint | Purpose |
|---|---|
| `GET /debug` | Browser Inspector page |
| `GET /health` | Server health and snapshot count |
| `GET /api/snapshots` | Snapshot list |
| `GET /api/snapshot/?frame=N` | Snapshot details for a frame |
| `GET /api/diff?from=A&to=B` | Snapshot diff |
| `GET /api/export` | Chromium-oriented export |

## Standalone Debugging

For post-mortem or low-interference debugging, use the standalone logger/debugger flow instead of embedding a live DevTools UI in the target process.

Current entry points:

| Tooling | Source |
|---|---|
| Logger API | `../standalone/logger.go` |
| Debugger implementation | `../standalone/debugger.go` |
| CLI command | `../../cmd/mint-debugger` |
| Usage guide | `../standalone/README.md` |

## Module Documentation

| Module | Current Doc | Purpose |
|---|---|---|
| Client | `../client/readme.md` | TUI panels and visualization helpers |
| Memory | `../memory/readme.md` | Ring buffers, adaptive sampling, memory monitor |
| Observation | `../observation/readme.md` | Metrics and pattern detection |
| Protocol | `../protocol/readme.md` | Message model and protocol server |
| Remote | `../remote/readme.md` | HTTP, WebSocket, Chromium bridge, Inspector UI |
| Replay | `../replay/readme.md` | Event recording and replay |
| Snapshot | `../snapshot/readme.md` | State capture and diff |
| Testing | `../testing/readme.md` | DevTools test fixtures and assertions |
| Time Travel | `../timetravel/readme.md` | Snapshot navigation and replay |

## Verification

Run these checks after changing DevTools code:

```bash
go test ./devtools/... -count=1
go test -race ./devtools/... -count=1 -p 1
```

For repository-wide changes, also run the release checks in `../../DEVELOPMENT.md`.
