# Mint Examples

This directory contains curated, runnable examples. Very small duplicates and historical debug probes were moved to `../docsArchive/cleanup-2026-05-19/_examples/`.

## Recommended Learning Path

Run these first:

```bash
go run ./examples/counter
go run ./examples/store_reducer_demo
go run ./examples/runapp_demo
go run ./examples/mvp_components_demo
go run ./examples/mvp_form_demo
go run ./examples/table_interactive_demo
go run ./examples/charts_gallery_demo
```

## Core Application Patterns

| Example | Purpose |
|---|---|
| `counter` | Small Store/Reducer counter |
| `store_reducer_demo` | Main Store/Reducer learning example |
| `runapp_demo` | `ui.RunApp` and `statemachine.AppRuntime[T]` |
| `store_mixed_demo` | Mixed local and app-level state |
| `typed_intent_demo` | Type-safe Intent flow |
| `typesafe_form_demo_runapp` | Type-safe form flow with RunApp |
| `timer` | Timed updates |
| `transition_demo` | Transition-style updates |

## Components And Forms

| Example | Purpose |
|---|---|
| `mvp_components_demo` | Broad component gallery |
| `mvp_form_demo` | Form workflow |
| `multiselect_demo` | Multi-select flow |
| `select` | Select behavior |
| `modal` | Modal behavior |
| `menu_demo` | Menu, popup and navigation behavior |
| `tabs` | Tabs behavior |
| `table_interactive_demo` | Table features |
| `virtuallist` | Virtual list |
| `error_boundary` | Error boundary behavior |

## Layout And Rendering

| Example | Purpose |
|---|---|
| `layout_demo` | Layout overview |
| `layout_component_fixtures_demo` | Layout component fixtures |
| `charts_gallery_demo` | Chart gallery |
| `charts_linechart_image_prototype` | Experimental image chart rendering |

## Advanced Runtime And Tooling

| Example | Purpose |
|---|---|
| `sandbox/*` | Sandbox recording, snapshots, injection and comprehensive tests |
| `devtools_demo/*` | DevTools and remote demo |
| `engine/*` | Engine-level integration demos |
| `fiber_firsts/*` | Fiber-first render and component demos |
| `fiber_counter_intent` | Fiber + Intent counter |
| `lane_scheduler_demo` | Scheduler demo |
| `ai_mcp_demo` | AI/MCP integration |
| `webdashboard_demo` | Web dashboard integration |
| `ant_design_demo` | Ant Design-style theme/component demo |
| `component_fixtures` | Component fixtures for tests |
| `data/list_render_demo` | Data list rendering |

## Example Policy

Keep examples that are useful as a learning path, maintained feature demo, integration scenario, or regression fixture. Prefer tests over new examples for narrow bug reproductions. Prefer extending `mvp_components_demo`, `mvp_form_demo`, `multiselect_demo`, or `charts_gallery_demo` over adding a tiny one-component demo.

Archived component-only entries:

- `checkbox`, `input`, and `date_time_picker_demo` are covered by `mvp_components_demo` and `mvp_form_demo`.
- `radiogroup_demo` is covered by the option and checkbox composition patterns in `multiselect_demo`.
- `mouse` is covered by the maintained interaction demos and E2E tests.
- `absolute`, `grid`, `layout_api_demo`, `layout/*`, `render`, and `fiber_demos/*` are historical layout/render probes now covered by `layout_demo`, `layout_component_fixtures_demo`, `fiber_firsts/*`, component tests, and runtime tests.
- `progress`, `toast`, `validation_demo`, and `list_interactive_demo` are single-component demos covered by component tests and broader form/table/list examples.
- `lane_demo`, `interruptible_demo`, and `timetravel_demo` are command-line API probes covered by runtime scheduler/debug tests or the maintained `lane_scheduler_demo` integration example.
- `clock_demo`, `fiber_firsts/list_demo`, `fiber_firsts/tabs_demo`, `fiber_firsts/textarea_demo`, `fiber_firsts/treeview_demo`, `fiber_firsts/virtuallist_demo`, and `fiber_firsts/wrap_demo` are single-component walkthroughs now covered by component docs/tests, `mvp_components_demo`, `ui_demos/*`, or the maintained top-level `tabs` and `virtuallist` examples.
