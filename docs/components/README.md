# Components Documentation

This directory contains user-facing component notes and historical component design documents. The implementation source of truth is `../../ui/components`, where each component package has its own `README.md`, builder, VNode/instance implementation and tests.

## Current Component System

Current source facts:

- 68 top-level component packages live in `../../ui/components` excluding `docs` and `internal`.
- Every top-level component package currently has a local `README.md`.
- Component regression tests are split between package tests under `../../ui/components/**` and end-to-end tests under `../../ui/e2e`.
- Chart components live under `../../ui/components/charts` and use shared internal chart helpers.
- Common construction style is a fluent builder such as `button.NewBuilder("OK").Primary().OnPress(MyIntent{}).Build()`.
- Interaction-oriented components generally emit `runtime/intent` intents rather than storing ad-hoc callbacks in the component tree.

## Source Layout

```
ui/components/
  <component>/
    README.md
    builder.go
    vnode.go
    instance.go
    intent.go        # when the component has typed intent integration
    *_test.go
  charts/
    <chart>/
    internal/
  internal/
    overlayposition/
    proputil/
    scroll/
    selection/
```

Not every component has every file listed above; small display components can be simpler.

## Component Inventory

### Basic Display

| Component | Source README | Notes |
|---|---|---|
| Text | `../../ui/components/text/README.md` | Text rendering and style. |
| Divider | `../../ui/components/divider/README.md` | Horizontal/vertical divider. |
| Badge | `../../ui/components/badge/README.md` | Label/count badge. |
| Tag | `../../ui/components/tag/README.md` | Tag display. |
| Empty | `../../ui/components/empty/README.md` | Empty state. |
| Descriptions | `../../ui/components/descriptions/README.md` | Key-value description layout. |
| Statistic | `../../ui/components/statistic/README.md` | Numeric/statistic display. |
| Timeline | `../../ui/components/timeline/README.md` | Timeline display. |
| Clock | `../../ui/components/clock/README.md` | Clock/tickable display. |
| Cursor | `../../ui/components/cursor/README.md` | Cursor rendering helpers. |

### Layout And Containers

| Component | Source README | Notes |
|---|---|---|
| Absolute | `../../ui/components/absolute/README.md` | Absolute positioning. |
| Grid | `../../ui/components/grid/README.md` | Grid layout and cell borders. |
| Layout | `../../ui/components/layout/README.md` | Application layout shell. |
| Panel | `../../ui/components/panel/README.md` | Bordered/framed panel. |
| RowCol | `../../ui/components/rowcol/README.md` | Row/column layout helpers. |
| Space | `../../ui/components/space/README.md` | Spacing layout. |
| SplitPane | `../../ui/components/splitpane/README.md` | Two-pane master/detail or top/bottom layout. |
| Wrap | `../../ui/components/wrap/README.md` | Wrapping layout. |
| ScrollView | `../../ui/components/scrollview/README.md` | Scrollable viewport. |
| Collapse | `../../ui/components/collapse/README.md` | Expand/collapse panels. |

### Form And Input

| Component | Source README | Notes |
|---|---|---|
| Input | `../../ui/components/input/README.md` | Text and number input. |
| Textarea | `../../ui/components/textarea/README.md` | Multi-line text input. |
| Checkbox | `../../ui/components/checkbox/README.md` | Checkbox and checked state. |
| Radio | `../../ui/components/radio/README.md` | Radio options. |
| Switch | `../../ui/components/switch/README.md` | Boolean switch. |
| Slider | `../../ui/components/slider/README.md` | Numeric slider. |
| Rate | `../../ui/components/rate/README.md` | Rating input. |
| Select | `../../ui/components/select/README.md` | Select/dropdown with overlay support. |
| Cascader | `../../ui/components/cascader/README.md` | Cascaded selection. |
| DatePicker | `../../ui/components/datepicker/README.md` | Date input and popup calendar. |
| TimePicker | `../../ui/components/timepicker/README.md` | Time segment picker. |
| Transfer | `../../ui/components/transfer/README.md` | Transfer list. |
| Form | `../../ui/components/form/README.md` | Form item/context and field integration. |
| FilterBar | `../../ui/components/filterbar/README.md` | Search/filter toolbar for data pages. |
| OptionGroup | `../../ui/components/optiongroup/README.md` | Shared option/radio/checkbox style grouping. |
| Validation | `../../ui/components/validation/README.md` | Validation helpers. |

### Actions, Feedback And Status

| Component | Source README | Notes |
|---|---|---|
| Button | `../../ui/components/button/README.md` | Button variants, size, focus style, `OnPress`. |
| Control | `../../ui/components/control/README.md` | Shared control interaction primitives. |
| Alert | `../../ui/components/alert/README.md` | Alert message. |
| Progress | `../../ui/components/progress/README.md` | Progress display. |
| Spin | `../../ui/components/spin/README.md` | Loading spinner. |
| Skeleton | `../../ui/components/skeleton/README.md` | Loading placeholder. |
| Result | `../../ui/components/result/README.md` | Result state. |
| Notification | `../../ui/components/notification/README.md` | Notification manager/component. |
| Toast | `../../ui/components/toast/README.md` | Toast feedback. |

### Navigation

| Component | Source README | Notes |
|---|---|---|
| Anchor | `../../ui/components/anchor/README.md` | Anchor navigation. |
| Breadcrumb | `../../ui/components/breadcrumb/README.md` | Breadcrumb trail. |
| Menu | `../../ui/components/menu/README.md` | Menu bar, popup and context menu. |
| Pagination | `../../ui/components/pagination/README.md` | Pagination control. |
| Steps | `../../ui/components/steps/README.md` | Step navigation. |
| Tabs | `../../ui/components/tabs/README.md` | Tabs, positioning, variants, reorder/close intents. |
| Toolbar | `../../ui/components/toolbar/README.md` | Operation toolbar for data and admin pages. |
| TreeView | `../../ui/components/treeview/README.md` | Hierarchical tree view. |
| StatusBar | `../../ui/components/statusbar/README.md` | Status bar and help overlay. |

### Data Display

| Component | Source README | Notes |
|---|---|---|
| List | `../../ui/components/list/README.md` | List selection and virtual bridge. |
| VirtualList | `../../ui/components/virtuallist/README.md` | Virtual list for large data. |
| Table | `../../ui/components/table/README.md` | Table rows, filtering, sorting, pagination, selection and expansion. |

### Overlay And Popup

| Component | Source README | Notes |
|---|---|---|
| Drawer | `../../ui/components/drawer/README.md` | Drawer overlay. |
| Modal | `../../ui/components/modal/README.md` | Modal component and static helpers. |
| Popconfirm | `../../ui/components/popconfirm/README.md` | Confirmation popup. |
| Popover | `../../ui/components/popover/README.md` | Popover overlay. |
| Tooltip | `../../ui/components/tooltip/README.md` | Tooltip overlay. |

Overlay positioning support lives in `../../ui/components/internal/overlayposition`.

### Charts

| Chart | Source README | Notes |
|---|---|---|
| Charts overview | `../../ui/components/charts/README.md` | Shared chart package entry. |
| BarChart | `../../ui/components/charts/barchart/README.md` | Bar chart. |
| BulletChart | `../../ui/components/charts/bulletchart/README.md` | Bullet chart. |
| Candlestick | `../../ui/components/charts/candlestick/README.md` | OHLC/candlestick chart. |
| Heatmap | `../../ui/components/charts/heatmap/README.md` | Heatmap chart. |
| LineChart | `../../ui/components/charts/linechart/README.md` | Text backend and image plot backend. |
| ScatterPlot | `../../ui/components/charts/scatterplot/README.md` | Scatter plot. |
| Sparkline | `../../ui/components/charts/sparkline/README.md` | Compact sparkline chart. |
| Model | `../../ui/components/charts/model/README.md` | Shared chart data model notes. |

Shared chart internals live under `../../ui/components/charts/internal`: `axis`, `canvas`, `downsample`, `layout`, `palette`, `raster`, `scale`.

## Public Builder Pattern

Most components expose `NewBuilder(...)` and a fluent API:

```go
btn := button.NewBuilder("Save").
    Primary().
    OnPress(SaveIntent{}).
    Build()
```

```go
tabsNode := tabs.NewBuilder().
    AddTab("overview", "Overview").
    AddTab("logs", "Logs").
    ActiveTabID("overview").
    OnChange(ChangeTabIntent{}).
    Build()
```

```go
tableNode := table.NewBuilder().
    Columns([]table.TableColumn{
        {Title: "ID", Width: 6},
        {Title: "Name", Width: 20},
    }).
    AddRow("1", "Alice").
    AddRow("2", "Bob").
    ShowBorder(true).
    Build()
```

For the exact API surface, prefer the component package README and `builder.go` in source.

## Interaction And State Pattern

Current components usually combine these mechanisms:

- VNode props for static configuration.
- Component instances for persistent local runtime state.
- `runtime/action` for low-level interaction dispatch.
- `runtime/intent` for semantic user intent.
- `runtime/store` and `runtime/reducer` for app-level state.
- Field binding helpers for form-style components.

This keeps the component tree declarative and allows E2E tests to assert messages, actions and intents.

## Examples

Useful example entry points:

- `../../examples/mvp_components_demo`
- `../../examples/mvp_form_demo`
- `../../examples/menu_demo`
- `../../examples/modal`
- `../../examples/select`
- `../../examples/table_interactive_demo`
- `../../examples/charts_gallery_demo`
- `../../examples/layout_component_fixtures_demo`
- `../../examples/component_fixtures`

## Tests

Run component unit tests:

```bash
go test ./ui/components/... -count=1
```

Run component E2E tests:

```bash
go test ./ui/e2e/... -count=1
```

Run selected suites:

```bash
go test ./ui/e2e -run TestButton -count=1
go test ./ui/components/table -count=1
go test ./ui/components/charts/... -count=1
```

## E2E Coverage Notes

`ui/e2e` currently contains broad component and interaction coverage, but not every `ui/components/<name>` directory has a one-to-one `<name>_e2e_test.go` file.

Direct or clearly named E2E coverage exists for most user-facing components. The following directories should be treated carefully when describing coverage because they either rely on package unit tests, aggregate suites, or support-module coverage rather than a direct same-name E2E file:

- `control`
- `drawer`
- `form`
- `notification`
- `optiongroup`
- `toast`
- `validation`
- `virtuallist`

Some of these are infrastructure modules (`control`, `validation`) and may not need direct E2E coverage. User-facing components such as `drawer`, `form`, `notification`, `optiongroup`, `toast` and `virtuallist` should be listed explicitly in any coverage matrix instead of being implied as directly covered.

Useful E2E commands:

```bash
go test ./ui/e2e/... -count=1
go test ./ui/e2e -run "Overlay|Modal|Select|Tabs|Table|Charts" -count=1
```

## Docs In This Directory

This `docs/components` directory is intentionally smaller than the source component README set. It contains higher-level or historical documents:

- [SCROLL_VIEW_COMPONENT.md](SCROLL_VIEW_COMPONENT.md)
- [VIRTUAL_LIST_COMPONENT.md](VIRTUAL_LIST_COMPONENT.md)
- [TABS_COMPONENT.md](TABS_COMPONENT.md)
- [TREEVIEW_NAVIGATION.md](TREEVIEW_NAVIGATION.md)
- [DATEPICKER_COMPONENT.md](DATEPICKER_COMPONENT.md)
- [TIMEPICKER_COMPONENT.md](TIMEPICKER_COMPONENT.md)
- [grid/](grid/)

Historical component fix reports and duplicate navigation notes were moved to `../../docsArchive/cleanup-2026-05-19/docs/components/`.

When updating component behavior, update the component-local README first, then update this index only if the inventory or navigation changes.

## Related Documentation

- [Main docs index](../README.md)
- [Architecture](../architecture/README.md)
- [Layout](../layout/README.md)
- [Store/Reducer](../ui/store/README.md)
- [E2E testing](../testing/e2e/README.md)
