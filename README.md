# Mint

Mint is a modern declarative TUI framework for Go. It provides a React-like VNode model, a Fiber-first render pipeline, a terminal component library, typed Intent dispatch, Store/Reducer state management, sandbox testing, E2E utilities, DevTools, and Inspector support.

This repository is usable as an SDK-style Go module. The public application surface is intentionally concentrated around `github.com/wwsheng009/mint/ui` plus a small set of runtime packages for state and intent handling.

## Status

Current repository status as of 2026-05-20:

| Area | Status |
|---|---|
| Module path | `github.com/wwsheng009/mint` |
| Go version | `go 1.24.0`, `toolchain go1.24.2` |
| Primary entry point | `ui.Run` |
| Store/Reducer entry point | `ui.Run` with `ui.WithInit(...)`, or `ui.RunApp` for `statemachine.AppRuntime[T]` |
| Rendering path | Fiber-first by default |
| Component library | Broad TUI component set under `ui/components` |
| Test support | Unit tests, component tests, sandbox, replay, E2E driver, render snapshots |
| CI gates | `go mod tidy` diff check, `go vet ./...`, regular package tests, serialized interaction/E2E package tests, and sharded focused `go test -race` over runtime/framework/sandbox/devtools/components |
| Documentation posture | SDK-facing docs are kept in this README, `DEVELOPMENT.md`, and the focused docs under `docs/`; historical implementation notes are archived under `docsArchive/` |

Recent local verification covered the module tidy check, `go vet ./...`, the full test suite using the same regular/interactive package split as CI, and the focused race gate split by subsystem and component shards. Full validation can take several minutes and should be run before releases.

## Installation

For an external application, reference Mint as a Go module.

```bash
mkdir my-tui-app
cd my-tui-app
go mod init example.com/my-tui-app
go get github.com/wwsheng009/mint
```

When developing against a local checkout, use a `replace` directive:

```bash
go mod edit -replace github.com/wwsheng009/mint=E:\projects\yao\wwsheng009\mint
go mod tidy
```

Example `go.mod`:

```go
module example.com/my-tui-app

go 1.24.0

require github.com/wwsheng009/mint v0.0.0

replace github.com/wwsheng009/mint => E:\projects\yao\wwsheng009\mint
```

Use the local `replace` form while Mint is consumed from a workspace instead of a tagged release.

## Minimal App

```go
package main

import "github.com/wwsheng009/mint/ui"

func App() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("Hello Mint").Bold(true).FgColor("cyan").Build(),
		ui.Text("A declarative terminal UI."),
	)
}

func main() {
	if err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Mint App"),
	); err != nil {
		panic(err)
	}
}
```

Run it:

```bash
go run .
```

## Recommended App Architecture

For production applications, use typed Intents and a Store/Reducer state flow.

```text
User input
  -> component emits Intent
  -> reducer handles Intent
  -> Store updates state
  -> view reads Store
  -> Fiber render update
```

Small local UI state can use Hooks. Cross-component or business state should use `runtime/store`, `runtime/reducer`, and `runtime/intent`.

```go
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

type AppState struct {
	Count int
}

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "AppIncrement" }
func (IncrementIntent) StayPressed() bool  { return true }

var appStore = store.NewStore(AppState{})

var appReducer = reducer.NewBuilder[AppState]().
	On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count++
		return s
	})

func registerHandlers() {
	appReducer.RegisterToGlobal(appStore)
}

func App() ui.VNode {
	state := appStore.Get()

	return ui.VStack(
		ui.NewTextBuilder("Counter").Bold(true).FgColor("cyan").Build(),
		ui.Text(fmt.Sprintf("Count: %d", state.Count)),
		ui.NewButtonBuilder(" +1 ").
			Variant(ui.ButtonVariantPrimary).
			OnPress(IncrementIntent{}).
			Build(),
	)
}

func main() {
	if err := ui.Run(App,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Counter"),
		ui.WithInit(registerHandlers),
	); err != nil {
		panic(err)
	}
}
```

## Public Packages

| Package | Use |
|---|---|
| `github.com/wwsheng009/mint/ui` | Main SDK entry: `ui.Run`, VNode helpers, component builders, Hooks, app options, testing helpers |
| `github.com/wwsheng009/mint/runtime/intent` | Typed user intent model, dispatch, built-in field change intent |
| `github.com/wwsheng009/mint/runtime/store` | Generic observable state store |
| `github.com/wwsheng009/mint/runtime/reducer` | Generic reducer builder and field binding helpers |
| `github.com/wwsheng009/mint/runtime/statemachine` | Optional AppRuntime with history and time-travel support |
| `github.com/wwsheng009/mint/sandbox` | Test sandbox, replay, event injection |

Avoid depending on `internal/*`. Treat `framework/*` and lower-level `runtime/*` packages as advanced extension points unless a document or example explicitly uses them.

## Common App Options

```go
ui.Run(App,
	ui.WithWidth(80),
	ui.WithHeight(24),
	ui.WithTitle("My App"),
	ui.WithFPS(60),
	ui.WithNoAlternateScreen(),
	ui.WithInit(registerHandlers),
)
```

Key options:

| Option | Purpose |
|---|---|
| `ui.WithWidth(width)` | Initial layout width |
| `ui.WithHeight(height)` | Initial layout height |
| `ui.WithSize(width, height)` | Initial layout size |
| `ui.WithTitle(title)` | Terminal/app title |
| `ui.WithFPS(fps)` | Render frame rate limit |
| `ui.WithNoAlternateScreen()` | Keep output in the normal terminal buffer |
| `ui.WithInit(fn)` | Register intents or app-level setup after Mint initializes its intent runtime |
| `ui.WithInteractionMode(mode)` | Choose interactive, app-selection, or terminal-selection mouse behavior |

## Component Surface

Most components are exposed through `ui.New...Builder` helpers:

```go
ui.NewButtonBuilder("Save").OnPress(SaveIntent{}).Build()
ui.NewInputBuilder().Placeholder("Search").Build()
ui.NewTableBuilder().Columns(cols).Rows(rows).Build()
ui.NewTabsBuilder().AddTab("overview", "Overview").Build()
ui.NewModalBuilder().Title("Confirm").Build()
```

Current component groups:

| Group | Components |
|---|---|
| Basic display | Text, Divider, Badge, Tag, Empty, Descriptions, Statistic, Timeline, Clock, Timer |
| Layout | VStack/HStack, Space, SplitPane, Layout, Grid, Row/Col, Panel, ScrollView, Wrap, Absolute |
| Form and input | Input, Textarea, Checkbox, Radio, Switch, Slider, Rate, Select, DatePicker, TimePicker, Cascader, Transfer, Form, FilterBar, Validation |
| Data | Table, List, VirtualList, TreeView |
| Feedback | Alert, Progress, Spin, Skeleton, Result, Notification, Toast |
| Navigation | Tabs, Menu, Breadcrumb, Pagination, Steps, Anchor, Toolbar, StatusBar |
| Overlay | Modal, Drawer, Tooltip, Popover, Popconfirm, ConfirmDialog |
| Charts | Sparkline, BulletChart, BarChart, LineChart, Heatmap, ScatterPlot, Candlestick |

For exact builder methods, read the component README and `builder.go` under `ui/components/<component>/`.

## Curated Examples

The examples directory is intentionally curated. Simple duplicates and historical probes are archived under `docsArchive/cleanup-2026-05-19/_examples/`.

Start here:

```bash
go run ./examples/counter
go run ./examples/store_reducer_demo
go run ./examples/runapp_demo
go run ./examples/mvp_components_demo
go run ./examples/mvp_form_demo
go run ./examples/menu_demo
go run ./examples/table_interactive_demo
go run ./examples/charts_gallery_demo
go run ./examples/sandbox/06_comprehensive
```

See `examples/README.md` for the maintained example map.

## Documentation Map

| Document | Purpose |
|---|---|
| `DEVELOPMENT.md` | SDK usage, application architecture, contribution workflow, test strategy |
| `docs/README.md` | Focused docs index |
| `docs/architecture/README.md` | Current architecture and runtime flow |
| `docs/components/README.md` | Component inventory and source map |
| `docs/ui/store/README.md` | Store/Reducer state management |
| `docs/debug/README.md` | Debugging and environment variables |
| `docs/sandbox/QUICK_START_GUIDE.md` | Sandbox and deterministic testing |
| `docs/testing/e2e/README.md` | E2E driver and interaction testing |
| `devtools/docs/README.md` | DevTools observability, remote debugging, and standalone debugging entry point |

Historical design notes, fix reports, probes, and duplicate examples live under `docsArchive/`. Archived example directories use an underscore-prefixed folder so `go test ./...` does not treat them as active packages.

## Testing

Fast SDK checks:

```bash
go test ./ui -count=1
go test ./runtime/intent ./runtime/store ./runtime/reducer ./runtime/statemachine -count=1
go test ./sandbox/... -count=1
```

Representative component checks:

```bash
go test ./ui/components/button ./ui/components/input ./ui/components/table ./ui/components/treeview ./ui/components/virtuallist ./ui/components/modal ./ui/components/charts/... -count=1
```

Full validation:

```bash
go mod tidy
git diff --exit-code -- go.mod go.sum
go vet ./...
mapfile -t packages < <(go list ./... | grep -v '^github.com/wwsheng009/mint/ui$' | grep -v '^github.com/wwsheng009/mint/ui/e2e$' | grep -v '^github.com/wwsheng009/mint/examples/error_boundary$' | grep -v '^github.com/wwsheng009/mint/examples/ui_demos/demo1_full_featured$' | grep -v '^github.com/wwsheng009/mint/examples/ui_demos/demo2_runtime_internals$')
go test "${packages[@]}" -count=1
go test ./ui ./examples/error_boundary ./examples/ui_demos/demo1_full_featured ./examples/ui_demos/demo2_runtime_internals -count=1 -p 1
go test ./ui/e2e -count=1 -p 1
go test -race ./runtime/... -count=1 -p 1
go test -race ./framework/... -count=1 -p 1
go test -race ./sandbox/... -count=1 -p 1
go test -race ./devtools/... -count=1 -p 1
go test -race ./ui/components/... -count=1 -p 1
```

The full suite and race gate are large. CI lets regular packages run with Go's default package parallelism, then runs runtime-sensitive `ui` tests, interaction-heavy examples such as `examples/error_boundary`, the active UI demos, and `ui/e2e` serially to keep timing and global terminal/runtime state isolated. The component race gate is also sharded for better failure isolation and to avoid a single long-running race process. Local validation can split `./ui/components/...` into smaller package groups when the single command is slow.

## Debugging

Debugging is controlled through environment variables. See `docs/debug/environment_variables.md`.

PowerShell examples:

```powershell
$env:TUI_LOG_OUTPUT="console"
$env:TUI_DEBUG_INTENT="true"
go run ./examples/store_reducer_demo
```

```powershell
$env:TUI_LOG_OUTPUT="console"
$env:TUI_DEBUG_RENDER="true"
$env:MINT_ASYNC_RENDER="false"
go run ./examples/counter
```

Useful categories:

| Variable | Purpose |
|---|---|
| `TUI_DEBUG_RENDER` | Render bridge and app rendering |
| `TUI_DEBUG_PAINT` | Paint output |
| `TUI_DEBUG_LAYOUT` | Layout details |
| `TUI_DEBUG_ACTION` | Action routing |
| `TUI_DEBUG_INTENT` | Intent dispatch |
| `TUI_DEBUG_HITMAP` | Mouse hit testing |
| `TUI_DEBUG_FOCUS` | Focus management |
| `MINT_ASYNC_RENDER=false` | Disable async rendering while debugging ordering issues |
| `MINT_NO_ALTERNATE_SCREEN=true` | Keep terminal output visible after exit |

## Repository Layout

```text
ui/                 Public SDK surface, shortcuts, hooks, test helpers
ui/components/      Component library
runtime/            Intent, store, reducer, layout, paint, input, focus primitives
framework/          App lifecycle, event loop, terminal integration
internal/render/    VNode/Fiber render bridge
internal/reconciler/Fiber reconciler and diff engine
sandbox/            Test sandbox and replay support
devtools/           Timeline, replay, diagnostics
examples/           Curated runnable examples
docs/               Focused current documentation
docsArchive/        Historical notes, archived docs, duplicate examples
```

## Versioning Guidance

Until Mint is consumed as a tagged external SDK, downstream applications should:

1. Use a local `replace` directive.
2. Pin the Mint checkout through your workspace or repository policy.
3. Run the fast SDK checks after pulling Mint changes.
4. Avoid depending on `internal/*`.
5. Prefer Intent-based component interaction over historical closure callback examples.

## License

MIT License
