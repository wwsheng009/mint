# FilterBar

`filterbar/` provides a Fiber-first filter toolbar for operational data pages. It is intended to sit above a table, list, or virtual list so operators can choose scope before selecting records and opening details.

The component follows Mint's component model:

- `VNode` stores declarative fields, actions, sizing, and style.
- `Instance` implements `RuntimeChildrenProvider` and synthesizes standard `Input`, `Select`, `Button`, `Wrap`, and layout nodes.
- Field changes use typed `intent.FieldChangeIntent` when `ForField(...)` is configured. Action buttons use normal `intent.Intent`.

## Example

```go
filters := filterbar.NewBuilder().
    Title("Provider Filters").
    Width(96).
    LabelWidth(8).
    Field(filterbar.Search("query", "Query", state.Query).
        WithPlaceholder("provider").
        ForField("providerQuery")).
    Field(filterbar.Select("status", "Status", []filterbar.Option{
        {Value: "all", Label: "All"},
        {Value: "healthy", Label: "Healthy"},
        {Value: "degraded", Label: "Degraded"},
    }).WithSelectedIndex(state.StatusIndex).ForField("providerStatus")).
    Action(filterbar.Button("refresh", "Refresh", RefreshIntent{}).Primary()).
    Action(filterbar.Button("reset", "Reset", ResetFiltersIntent{})).
    Build()
```

## Operational Layout Guidance

Use `FilterBar` before the primary data surface. For manager-style operations pages, a common page sequence is:

1. Page title and scope summary.
2. `FilterBar` for query, status, group, provider, time range, and refresh/reset.
3. Primary table/list.
4. Detail drawer/panel only after selection.

Avoid placing permanent edit forms, JSON previews, or destructive operation panels beside the first-screen table. Those belong after selection or explicit action.

## Field Types

- `Search(key, label, value)` renders a search input.
- `Text(key, label, value)` renders a text input.
- `Select(key, label, options)` renders a select control.
- `Custom(key, label, node)` lets applications provide a specialized control.

Common field modifiers:

- `WithPlaceholder(text)`
- `WithWidth(width)`
- `WithLabelWidth(width)`
- `WithSelectedIndex(index)`
- `ForField(fieldName)`
- `OnChange(intent)`
- `OnSubmit(intent)`
- `WithDisabled(true)`

`FilterBar` defaults to a stable single-row layout. Use `Wrap(true)` only when the host page has enough vertical space and the wrapped controls have been checked in the target terminal size.

## Actions

Use `Button(key, label, intent)` for toolbar commands such as refresh, reset, export, or open advanced filters.

```go
filterbar.Button("refresh", "Refresh", RefreshIntent{}).Primary()
filterbar.Button("reset", "Reset", ResetFiltersIntent{})
filterbar.Button("clear", "Clear", ClearIntent{}).WithDisabled(true)
```

High-risk actions should not live in a filter bar. Keep them in selection-specific detail panels or confirmations where the target and impact are visible.
