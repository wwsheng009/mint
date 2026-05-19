# Mint Development Guide

This guide is the SDK-level development reference for building applications with Mint and for making framework changes in this repository.

## Mental Model

Mint applications are declarative terminal applications.

```text
App view function
  -> VNode tree
  -> Fiber reconciliation
  -> layout tree
  -> paint buffer
  -> terminal output
```

User input follows the opposite direction:

```text
terminal input
  -> framework event pump
  -> action routing
  -> focused component or hit target
  -> Intent
  -> reducer/store update
  -> render update
```

For application code, stay at the `ui` and `runtime/{intent,store,reducer}` level. Only framework contributors should normally work directly in `framework`, `internal/render`, or `internal/reconciler`.

## Build A New TUI App

Create a separate Go module and reference Mint:

```bash
mkdir my-tui-app
cd my-tui-app
go mod init example.com/my-tui-app
go mod edit -replace github.com/wwsheng009/mint=E:\projects\yao\wwsheng009\mint
go get github.com/wwsheng009/mint
go mod tidy
```

Use this package split for non-trivial applications:

```text
my-tui-app/
  main.go
  internal/app/
    state.go
    intents.go
    reducer.go
    view.go
  internal/services/
    ...
```

`main.go` should remain thin:

```go
package main

import (
	"example.com/my-tui-app/internal/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	if err := ui.Run(app.View,
		ui.WithWidth(100),
		ui.WithHeight(32),
		ui.WithTitle("My App"),
		ui.WithInit(app.RegisterHandlers),
	); err != nil {
		panic(err)
	}
}
```

## Application State Pattern

Use this pattern for production code:

1. Define one app state type.
2. Define typed Intent structs for user actions.
3. Create one `store.Store[T]`.
4. Register reducers with `reducer.NewBuilder[T]()`.
5. Register handlers in `ui.WithInit(...)`.
6. Render from state snapshots.

```go
package app

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

type State struct {
	Count int
	Query string
}

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

var Store = store.NewStore(State{})

var Reducer = reducer.NewBuilder[State]().
	On(IncrementIntent{}, func(s State, i intent.Intent) State {
		s.Count++
		return s
	}).
	On(intent.FieldChangeIntent{}, func(s State, i intent.Intent) State {
		f := i.(intent.FieldChangeIntent)
		if f.Field == "query" {
			s.Query = f.Value
		}
		return s
	})

func RegisterHandlers() {
	Reducer.RegisterToGlobal(Store)
}

func View() ui.VNode {
	state := Store.Get()

	return ui.VStack(
		ui.NewTextBuilder("Search").Bold(true).Build(),
		ui.NewInputBuilder().
			Value(state.Query).
			ForField(intent.BindField("query")).
			Build(),
		ui.Text(fmt.Sprintf("Count: %d", state.Count)),
		ui.NewButtonBuilder("Increment").OnPress(IncrementIntent{}).Build(),
	)
}
```

### When To Use Hooks

Hooks are appropriate for local, short-lived component state:

| Use Hooks For | Use Store/Reducer For |
|---|---|
| temporary input focus details | business state |
| local animation counters | form data used across sections |
| component-local refs | selected records |
| local effects and cleanup | persistence, workflows, undo, cross-component updates |

Do not use Hooks as a substitute for app-level state.

## Component Usage Rules

Prefer the public `ui` package:

```go
ui.NewButtonBuilder("Save").OnPress(SaveIntent{}).Build()
ui.NewInputBuilder().Placeholder("Name").Build()
ui.NewTableBuilder().Columns(cols).Rows(rows).Build()
ui.VStack(...)
ui.HStack(...)
```

Use component packages directly only when the shortcut does not expose what you need:

```go
import "github.com/wwsheng009/mint/ui/components/table"
```

When API behavior is unclear, read files in this order:

1. `ui/components/<name>/README.md`
2. `ui/components/<name>/builder.go`
3. `ui/components/<name>/vnode.go`
4. `ui/components/<name>/instance.go`
5. `ui/components/<name>/*_test.go`
6. `ui/e2e/*_<name>_e2e_test.go`

## Intent Naming

Intent names share a global registry. Use stable, domain-specific names:

```go
func (SaveUserIntent) IntentType() string { return "User.Save" }
func (OpenAuditIntent) IntentType() string { return "Audit.Open" }
```

Avoid generic names like `Click`, `Submit`, or `Change` in application code unless they are scoped by package or domain.

For button-like interactions, implement `StayPressed()` when the visual pressed state should remain long enough to be visible:

```go
func (SaveUserIntent) StayPressed() bool { return true }
```

## Layout Guidance

Use high-level layout components first:

| Need | Preferred API |
|---|---|
| vertical stack | `ui.VStack(...)` |
| horizontal row | `ui.HStack(...)` |
| shell layout | `ui.NewLayoutBuilder()` |
| dense controls | `ui.NewSpaceBuilder()` or HStack/VStack |
| scrollable region | `ui.NewScrollViewBuilder()` |
| data table | `ui.NewTableBuilder()` |
| large lists | `ui.NewVirtualListBuilder()` |
| overlay | Modal, Drawer, Tooltip, Popover, Popconfirm |

Do not hand-roll focus, overlay placement, portal roots, or mouse hit maps in application code. Use the component library unless you are implementing framework internals.

## Testing Strategy

Use the smallest test surface that validates the behavior.

| Change Type | Recommended Tests |
|---|---|
| app state reducer | normal Go unit test for reducer behavior |
| component builder | package unit test under `ui/components/<name>` |
| component interaction | `ui.RunTest` or `ui/e2e` driver |
| event routing/focus/overlay | E2E test with snapshots or locator assertions |
| sandbox behavior | `sandbox/...` tests |
| render/layout internals | targeted tests under `internal/render`, `runtime/layout`, or `ui/layout` |

Fast checks:

```bash
go test ./ui -count=1
go test ./runtime/intent ./runtime/store ./runtime/reducer ./runtime/statemachine -count=1
go test ./sandbox/... -count=1
```

Representative component checks:

```bash
go test ./ui/components/button ./ui/components/input ./ui/components/table ./ui/components/treeview ./ui/components/virtuallist ./ui/components/modal ./ui/components/charts/... -count=1
```

Full check:

```bash
go test ./... -count=1
```

Run the full suite before releases, broad refactors, or changes to render, focus, event routing, or component base behavior.

## Debugging

Use current logger categories from `docs/debug/environment_variables.md`.

PowerShell:

```powershell
$env:TUI_LOG_OUTPUT="console"
$env:TUI_DEBUG_INTENT="true"
$env:TUI_DEBUG_ACTION="true"
go run ./examples/store_reducer_demo
```

Render/layout investigation:

```powershell
$env:TUI_LOG_OUTPUT="console"
$env:TUI_DEBUG_RENDER="true"
$env:TUI_DEBUG_LAYOUT="true"
$env:TUI_DEBUG_PAINT="true"
$env:MINT_ASYNC_RENDER="false"
go run ./examples/counter
```

Mouse or overlay investigation:

```powershell
$env:TUI_LOG_OUTPUT="console"
$env:TUI_DEBUG_HITMAP="true"
$env:TUI_DEBUG_PUMP="true"
go run ./examples/menu_demo
```

## Documentation Policy

Current documentation belongs in:

| Location | Content |
|---|---|
| `README.md` | SDK overview and quick start |
| `DEVELOPMENT.md` | App and framework development workflow |
| `docs/README.md` | Focused current docs index |
| `docs/<area>/README.md` | Area-level current state and references |
| `ui/components/<component>/README.md` | Component-specific API |
| `examples/README.md` | Curated runnable example map |

Historical fix reports, one-off investigations, design alternatives, implementation tickets, and duplicate demos belong under `docsArchive/`.

Do not add new root-level status files unless they are meant to be maintained. Prefer updating the relevant current document.

## Example Policy

Examples should be either:

1. A recommended learning path.
2. A maintained feature demonstration.
3. A regression fixture that cannot be expressed well as a test.
4. A complex integration scenario.

Avoid adding single-component toy examples when the same behavior is already covered by `mvp_components_demo`, `mvp_form_demo`, `charts_gallery_demo`, or component tests.

Archived examples from the 2026-05-19 cleanup are under:

```text
docsArchive/cleanup-2026-05-19/_examples/
```

## Framework Contribution Workflow

1. Read the relevant current docs and nearby tests.
2. Identify the public behavior that must change.
3. Make the smallest scoped code change.
4. Add or update tests at the same layer.
5. Run targeted tests first.
6. Run the fast SDK checks.
7. Run the full suite for broad changes.
8. Update README, `DEVELOPMENT.md`, area docs, or component README only when public behavior changes.

## API Stability Rules

Keep the public SDK stable around:

| Surface | Stability Expectation |
|---|---|
| `ui.Run` and app options | High |
| `ui.New...Builder` component constructors | High |
| `runtime/intent` Intent interface and field change flow | High |
| `runtime/store` and `runtime/reducer` | High |
| `ui/test.go` and sandbox public helpers | Medium-high |
| `framework/*` | Advanced, less stable |
| `internal/*` | Not public |

When changing public builders, preserve compatibility where practical. If a breaking change is necessary, update examples, component README, and migration notes in the same change.

## Release Checklist

Before publishing or handing off a version:

1. `go test ./ui -count=1`
2. `go test ./runtime/intent ./runtime/store ./runtime/reducer ./runtime/statemachine -count=1`
3. `go test ./sandbox/... -count=1`
4. `go test ./ui/components/... -count=1`
5. `go test ./ui/e2e/... -count=1`
6. `go test ./... -count=1`
7. Run the curated examples that cover changed behavior.
8. Confirm `README.md`, `DEVELOPMENT.md`, `docs/README.md`, and `examples/README.md` still point to existing files.
