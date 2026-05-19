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
| `timetravel_demo` | Time-travel state history |

## Components And Forms

| Example | Purpose |
|---|---|
| `mvp_components_demo` | Broad component gallery |
| `mvp_form_demo` | Form workflow |
| `validation_demo` | Validation helpers |
| `date_time_picker_demo` | DatePicker and TimePicker |
| `radiogroup_demo` | Radio group behavior |
| `checkbox` | Checkbox behavior |
| `multiselect_demo` | Multi-select flow |
| `input` | Input behavior |
| `select` | Select behavior |
| `progress` | Progress variants |
| `toast` | Toast feedback |
| `modal` | Modal behavior |
| `menu_demo` | Menu, popup and navigation behavior |
| `tabs` | Tabs behavior |
| `list_interactive_demo` | Interactive list |
| `table_interactive_demo` | Table features |
| `virtuallist` | Virtual list |
| `clock_demo` | Clock component |
| `error_boundary` | Error boundary behavior |

## Layout And Rendering

| Example | Purpose |
|---|---|
| `layout_demo` | Layout overview |
| `layout_api_demo` | Layout API usage |
| `layout_component_fixtures_demo` | Layout component fixtures |
| `layout/*` | Layout visualizer and buffer demos |
| `grid` | Grid layout |
| `absolute` | Absolute positioning |
| `render` | Render path demo |
| `mouse` | Mouse interactions |
| `charts_gallery_demo` | Chart gallery |
| `charts_linechart_image_prototype` | Experimental image chart rendering |

## Advanced Runtime And Tooling

| Example | Purpose |
|---|---|
| `sandbox/*` | Sandbox recording, snapshots, injection and comprehensive tests |
| `devtools_demo/*` | DevTools and remote demo |
| `engine/*` | Engine-level integration demos |
| `fiber_firsts/*` | Fiber-first render and component demos |
| `fiber_demos/*` | Fiber/layout comparison demos |
| `fiber_counter_intent` | Fiber + Intent counter |
| `lane_demo` | Lane scheduling |
| `lane_scheduler_demo` | Scheduler demo |
| `interruptible_demo` | Interruptible rendering/update demo |
| `ai_mcp_demo` | AI/MCP integration |
| `webdashboard_demo` | Web dashboard integration |
| `ant_design_demo` | Ant Design-style theme/component demo |
| `component_fixtures` | Component fixtures for tests |
| `data/list_render_demo` | Data list rendering |

## Example Policy

Keep examples that are useful as a learning path, maintained feature demo, integration scenario, or regression fixture. Prefer tests over new examples for narrow bug reproductions. Prefer extending `mvp_components_demo`, `mvp_form_demo`, or `charts_gallery_demo` over adding a tiny one-component demo.
