# Mint Documentation

This directory contains current, maintained documentation for Mint. Historical design notes, fix reports, implementation investigations, and duplicate materials are archived under `docsArchive/`.

For SDK users, start with:

1. `../README.md`
2. `../DEVELOPMENT.md`
3. `components/README.md`
4. `ui/store/README.md`
5. `debug/README.md`
6. `sandbox/QUICK_START_GUIDE.md`
7. `testing/e2e/README.md`

## Current Source Facts

| Item | Current Value |
|---|---|
| Module | `github.com/wwsheng009/mint` |
| Go version | `go 1.24.0`, `toolchain go1.24.2` |
| Primary app entry | `ui.Run` |
| Store/Reducer app entry | `ui.Run` with `ui.WithInit(...)`, or `ui.RunApp` for `statemachine.AppRuntime[T]` |
| Public SDK package | `ui` |
| Recommended state packages | `runtime/intent`, `runtime/store`, `runtime/reducer` |
| Component source | `ui/components` |
| Example source | `examples` |
| Historical archive | `docsArchive` |

## Current Documentation Map

### SDK And API

| Document | Purpose |
|---|---|
| `../README.md` | SDK overview, installation, quick start, examples, test commands |
| `../DEVELOPMENT.md` | Application architecture, contribution workflow, testing and docs policy |
| `api/component.md` | VNode and component API reference |
| `api/hooks.md` | Hooks API reference and current usage notes |
| `api/border.md` | Border API |
| `api/memory-safety.md` | Memory safety, cleanup, subscription and goroutine guidance |

### Architecture

| Document | Purpose |
|---|---|
| `architecture/README.md` | Current layered architecture and runtime flow |
| `architecture/design/FIBER_ARCHITECTURE.md` | Fiber architecture detail |
| `fiber/fiber_first/consolidated/README.md` | Fiber-first reference entry |
| `fiber/fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md` | Fiber-first architecture |
| `fiber/fiber_first/consolidated/FIBER_FIRST_QUICK_REFERENCE.md` | Fiber quick reference |
| `fiber/fiber_first/FIBER_PAINT_ARCHITECTURE.md` | Fiber paint path |

### Components

| Document | Purpose |
|---|---|
| `components/README.md` | Component inventory and source map |
| `components/SCROLL_VIEW_COMPONENT.md` | ScrollView notes |
| `components/VIRTUAL_LIST_COMPONENT.md` | VirtualList notes |
| `components/TABS_COMPONENT.md` | Tabs notes |
| `components/TREEVIEW_NAVIGATION.md` | TreeView navigation |
| `components/DATEPICKER_COMPONENT.md` | DatePicker notes |
| `components/TIMEPICKER_COMPONENT.md` | TimePicker notes |
| `components/grid/ARCHITECTURE.md` | Grid architecture |
| `components/grid/ACCEPTANCE.md` | Grid acceptance notes |
| `components/grid/DEBUGGING_GUIDE.md` | Grid debugging |

The most accurate component API is normally under `../ui/components/<component>/README.md` and `../ui/components/<component>/builder.go`.

### State And Intent

| Document | Purpose |
|---|---|
| `ui/store/README.md` | Store/Reducer overview |
| `ui/store/guides/README.md` | Store guide entry |
| `ui/store/guides/DEVELOPMENT_GUIDE.md` | Store/Reducer development guide |
| `ui/store/guides/STORE_REDUCER_GUIDE.md` | Store/Reducer usage |
| `ui/store/guides/RUNAPP_GUIDE.md` | `ui.RunApp` guide |
| `ui/store/guides/HOOK_USAGE_GUIDE.md` | Hook usage with the current state model |
| `ui/store/guides/MIGRATION_GUIDE.md` | Migration notes |
| `ui/store/api/API_REFERENCE.md` | Store API reference |
| `ui/store/features/TYPE_SAFE_INTENT.md` | Type-safe intent notes |
| `ui/store/features/LOGGING_AND_ERROR_HANDLING_GUIDE.md` | Logging and error handling |
| `ui/store/hybrid/STATE_MANAGEMENT_GUIDE.md` | Hybrid state guidance |
| `ui/store/hybrid/HYBRID_MODE_IMPLEMENTATION.md` | Hybrid mode implementation notes |
| `ui/store/optimization/FIELD_BINDING_OPTIMIZATION.md` | Field binding optimization |

### Layout, Render, Theme

| Document | Purpose |
|---|---|
| `layout/README.md` | Layout docs entry |
| `layout/user_guide/README.md` | User-facing layout guide |
| `layout/core_concepts/flex_layout.md` | Flex layout |
| `layout/core_concepts/stretch_layout.md` | Stretch layout |
| `layout/core_concepts/wrap_component.md` | Wrap layout |
| `layout/core_concepts/layer_system_guide.md` | Layer concepts |
| `layout/box_model/box_model_quick_reference.md` | Box model quick reference |
| `layout/border/border_processing_flow.md` | Border processing |
| `layout/modal/modal_centering_guide.md` | Modal centering |
| `layout/visualizer_usage_guide.md` | Layout visualizer |
| `render/README.md` | Render docs entry |
| `render/diff/diff.md` | Diff rendering |
| `render/hook/README.md` | Render hooks |
| `render/paint/optimized/README.md` | Optimized paint path |
| `render/paint/optimized/FIBER_FIRST_RENDER_PIPELINE.md` | Fiber-first render pipeline |
| `render/pixel/README.md` | Pixel/graphics rendering notes |
| `theme/theme_system_guide.md` | Theme system |
| `theme/ant_design_implementation_guide.md` | Ant Design theme mapping |
| `theme/spacing_and_background_design.md` | Spacing and background |

### Runtime Features

| Document | Purpose |
|---|---|
| `debug/README.md` | Debugging entry |
| `debug/environment_variables.md` | Current environment variables |
| `debug/quick_start.md` | Debug quick start |
| `features/mouse-text-selection.md` | Mouse text selection |
| `features/focus/README.md` | Focus feature notes |
| `inspector/README.md` | Inspector entry |
| `inspector/QUICK_START.md` | Inspector quick start |
| `inspector/KEYBOARD_SHORTCUTS.md` | Inspector keyboard shortcuts |
| `layer/LAYER_SYSTEM_ARCHITECTURE.md` | Layer architecture |
| `layer/FIBER_FIRST_LAYER_SYSTEM.md` | Fiber-first layer system |
| `layer/PORTAL_IMPLEMENTATION.md` | Portal implementation |
| `platform/platform.md` | Platform input notes |
| `platform/key_release.md` | Key release notes |

### Testing And Tooling

| Document | Purpose |
|---|---|
| `sandbox/QUICK_START_GUIDE.md` | Sandbox quick start |
| `sandbox/API_REFERENCE.md` | Sandbox API |
| `sandbox/USER_GUIDE.md` | Sandbox user guide |
| `sandbox/SANDBOX_ADVANCED_FEATURES.md` | Advanced sandbox features |
| `sandbox/SANDBOX_DEBUG_GUIDE.md` | Sandbox debugging |
| `sandbox/APP_LIFECYCLE_AND_SANDBOX.md` | App lifecycle and sandbox |
| `testing/TESTING_TOOL.md` | Testing tool guide |
| `testing/e2e/README.md` | E2E testing entry |
| `testing/e2e/API_REFERENCE.md` | E2E API |
| `testing/e2e/INTERACTIVE_E2E_SUITE_DESIGN.md` | E2E suite design |
| `log/LOGGER_ENV_VAR_STANDARD.md` | Logger env var standard |

### AI And Integrations

| Document | Purpose |
|---|---|
| `ai/design/framework_app_ai_mcp_design.md` | AI/MCP app integration design |

## Archive Policy

The active docs tree should contain maintained user guides, API references, and current architecture documents.

Move the following to `docsArchive/` instead of keeping them in `docs/`:

| Archive Content | Reason |
|---|---|
| fix reports | useful history, not SDK reference |
| implementation completion reports | stale quickly |
| one-off investigation notes | too specific for current docs |
| alternative design drafts | confusing when mixed with current architecture |
| duplicate component demos | better represented by curated examples or tests |
| temporary debug probes | should not be part of learning path |

The 2026-05-19 cleanup moved archived materials to:

```text
docsArchive/cleanup-2026-05-19/
docsArchive/cleanup-2026-05-19/_examples/
```
