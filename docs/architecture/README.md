# Architecture Documentation

This page summarizes the current Mint architecture as reflected by the source tree. Older design records still exist under `docsArchive/`, but this file should be treated as the current architecture entry point.

## Current Shape

Mint is a Go TUI framework built around a declarative UI layer, a framework application runtime, a Fiber-first render pipeline, typed state flow, and a testable input/sandbox stack.

```
Application code
  |
  | ui.Run / ui.RunApp
  v
ui package
  - options, app entry, shortcuts
  - default Portal roots
  - Intent runtime bootstrap
  |
  v
framework.App
  - lifecycle and event loop
  - input pump and Action routing
  - focus, selection, hit map
  - async render and terminal output
  - theme, AI, inspector integration
  |
  v
internal/render + internal/reconciler
  - VNode -> Fiber reconciliation
  - Fiber-first layout and paint path
  - Portal-aware two-phase layout
  - HitMap and dirty paint hints
  |
  v
runtime packages
  - layout, paint, render, platform
  - input, msg, action, intent
  - store, reducer, statemachine
  - focus, selection, scheduler
```

## Main Layers

| Layer | Source | Responsibility |
|---|---|---|
| Public UI | `../../ui` | Public entry points, component shortcuts, Hooks, testing helpers, `ui.Run`, `ui.RunApp`. |
| Component library | `../../ui/components` | 60 top-level component packages plus chart subpackages. Each component owns VNode, builder, instance, intent and tests where applicable. |
| Framework runtime | `../../framework` | App lifecycle, event pump, theme, validation, component bridge, AI integration, inspector integration. |
| Render bridge | `../../internal/render` | Bridges declarative VNode functions into framework components, owns Fiber-first paint/layout adapters and Portal-aware layout. |
| Reconciler | `../../internal/reconciler` | Fiber tree construction, diffing, component instance lifecycle, stable node identity, memo/error boundary support. |
| Runtime primitives | `../../runtime` | Lower-level UI, layout, paint, platform, input, action, intent, focus, scheduler, state and selection packages. |
| Testing stack | `../../sandbox`, `../../ui/e2e`, `../../ui/test.go` | Mock/real/replay sandbox, deterministic event injection, E2E driver, locators, traces and diagnostics. |
| Developer tools | `../../devtools`, `../../internal/inspector` | Timeline, replay, snapshot, remote observation and UI inspector support. |

## Application Startup

The primary application entry point is `ui.Run` in `../../ui/app.go`.

At startup it:

1. Applies options such as size, title, FPS, no-alternate-screen, interaction mode, plugin setup and AI config.
2. Creates `framework.NewApp()`.
3. Sets configured layout size and terminal buffer size.
4. Initializes the default theme through `framework/theme`.
5. Installs graphics bootstrap support.
6. Creates an `intent.Runtime`, registers built-in intent handlers and exposes it to `runtime/ui`.
7. Wraps the app component with default overlay, modal and tooltip Portal roots.
8. Creates `internal/render.NewDeclarativeNodeFromFuncWithFiber(...)`.
9. Connects the declarative node to `framework.App`, FocusManager and Intent runtime.
10. Runs `framework.App.Run()`.

`ui.RunApp` follows the same framework path but integrates `runtime/statemachine.AppRuntime[T]` so Store/Reducer state changes trigger rerenders.

## Rendering Path

Fiber-first is the default path used by `ui.Run`.

Current render flow:

```
ComponentFunc
  -> VNode tree
  -> internal/reconciler Fiber tree
  -> runtime/layout LayoutBox tree
  -> runtime/paint PaintableBox / Buffer / SceneFrame
  -> framework.App terminal output
```

Important implementation facts:

- `internal/render.NewDeclarativeNodeFromFuncWithFiber` creates the Fiber reconciler and `FiberFocusManager`.
- Portal-aware layout is enabled by default and can be disabled with `MINT_PORTAL_LAYOUT=false` or `MINT_PORTAL_LAYOUT=0`.
- `DeclarativeNode` stores the latest HitMap, layout result, paintable root and dirty rect hints for framework/app and tests.
- `framework.App` uses `paint.Renderer` and an async renderer by default; async rendering can be disabled with `MINT_ASYNC_RENDER=false`.
- Graphics/image presentation is experimental and controlled by `runtime/platform` graphics environment variables such as `MINT_GRAPHICS` and `MINT_CELL_PIXELS`.

Related docs:

- [Fiber-first consolidated docs](../fiber/fiber_first/consolidated/README.md)
- [Render system docs](../render/README.md)
- [Layer and Portal docs](../layer/LAYER_SYSTEM_ARCHITECTURE.md)
- [Layout docs](../layout/README.md)

## Event And Input Flow

The framework runtime uses a layered event/action path:

```
platform raw input
  -> framework/event.Pump
  -> runtime/msg
  -> runtime/action.InputProcessor
  -> runtime/action.Router / ScopeDispatcher
  -> Fiber target instance or global handler
  -> Intent / Store / reducer update as needed
  -> ScheduleUpdate / MarkDirty
```

Notable source locations:

- `../../framework/event`: event pump, key map, event abstractions and event logger.
- `../../runtime/input`: input tracker, keymap, mouse tracker and snapshots.
- `../../runtime/msg`: runtime message types.
- `../../runtime/action`: action model, router, middleware, scope dispatcher.
- `../../runtime/interaction`: pressed/click/cancel/reset interaction context.
- `../../runtime/intent`: typed intent runtime and registry.

## State Model

Mint currently supports two complementary state styles:

- Local component state through Hooks in `ui` / `runtime/ui`.
- App-level typed state through `runtime/store`, `runtime/reducer`, `runtime/intent` and `runtime/statemachine`.

Recommended app-level flow:

```
Component interaction
  -> Intent
  -> Registry / Dispatcher
  -> Reducer[T]
  -> Store[T]
  -> ui.UseStoreSelector / AppRuntime view
  -> Fiber update
```

Related docs:

- [Store/Reducer docs](../ui/store/README.md)
- [State docs](../state/README.md)
- [Runtime intent README](../../runtime/intent/README.md)

## Focus, Selection And Interaction Modes

`framework.App` exposes three interaction modes:

- `InteractionModeInteractive`: normal app interaction with mouse capture.
- `InteractionModeAppSelection`: app-managed selection through `runtime/selection`; Ctrl+C copies selection instead of quitting.
- `InteractionModeTerminalSelection`: terminal-native text selection with mouse capture disabled.

The public UI package exposes these through `ui.WithInteractionMode`, `ui.SetInteractionMode`, `ui.GetInteractionMode` and `ui.CycleInteractionMode`.

Related source:

- `../../framework/app.go`
- `../../runtime/focus`
- `../../runtime/selection`
- `../../runtime/input`

## Components

The component library lives under `../../ui/components`. The public pattern is:

- `NewBuilder(...)` or package-specific constructor.
- `Build()` returning `runtime/ui.VNode` / `ui.VNode`.
- Optional `BuildTyped()` or `BuildInstance()` for tests and internal use.
- State and interaction handled by component instance + Intent / Action integration.
- E2E tests live in `../../ui/e2e`.

Current component index: [components/README.md](../components/README.md).

## Testing Architecture

Mint has three testing surfaces:

- `ui.RunTest` / `ui.RunTestWithSandbox` in `../../ui/test.go` for framework-backed unit tests.
- `ui/e2e` for higher-level interactions, locators, traces and visual assertions.
- `sandbox` packages for mock, real and replay execution.

Recommended commands:

```bash
go test ./ui/components/... -count=1
go test ./ui/e2e/... -count=1
go test ./sandbox/... -count=1
go test ./... -count=1
```

Related docs:

- [Sandbox quick start](../sandbox/QUICK_START_GUIDE.md)
- [E2E docs](../testing/e2e/README.md)

## Related Documentation

- [Main docs index](../README.md)
- [Components](../components/README.md)
- [Debug](../debug/README.md)
- [Render](../render/README.md)
- [Layout](../layout/README.md)
- [Layer](../layer/LAYER_SYSTEM_ARCHITECTURE.md)
- [Store/Reducer](../ui/store/README.md)
