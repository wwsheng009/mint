# FilterBar

`filterbar/` provides a Fiber-first filter toolbar for operational data pages. It is intended to sit above a table, list, or virtual list so operators can choose scope before selecting records and opening details.

The component follows Mint's component model:

- `VNode` stores declarative fields, actions, sizing, and style.
- `Instance` implements `RuntimeChildrenProvider` and synthesizes standard `Input`, `Select`, `Button`, `Wrap`, and layout nodes.
- Field changes use typed `intent.FieldChangeIntent` when `ForField(...)` is configured. Action buttons use normal `intent.Intent`.
- `Summary(...)` renders a compact scope/status line below the title, useful for current tenant, time range, page, result count, or server-side filter state.
- `Action.WithDisabledReason(...)` disables the action and renders a compact reason line so operators understand the missing prerequisite; `WithDisabledReasonIf(condition, reason)` keeps condition checks chainable and appends matching reasons when loading, lookup, or selection state controls availability. Use `WithDisabledReasons(...)` when one action can have multiple simultaneous missing prerequisites and the operator should see all of them at once.
- `SummaryCompactSearch(...)`, `CompactPageSummary(...)`, and `LookupSummary(...)` keep user-entered search or lookup values width-bounded in the summary while the input itself keeps the full value.
- `SummaryValueUnless(...)` and `SummaryCompactUnless(...)` omit inactive default scopes, for example `status=all`, so wrapped multi-field pages can reserve summary width for filters that are actually narrowing the result set.
- `SummaryPresence(...)` reports required field readiness without displaying the field value itself, useful for operation reason, confirmation text, ticket id, or other short write-flow prerequisites.
- `ClearFieldAction(...)`, `ClearFieldActionWithLabel(...)`, and `SearchRefreshClear(...)` provide standard Clear actions for bound search/filter fields. Clear only emits `intent.FieldChangeIntent{Value: ""}`; the application reducer still owns pagination reset, selection reset, and data reload behavior. Use the labeled variant when one multi-field bar needs separate actions such as `Clear Search` and `Clear History`.
- `ResetAction(...)`, `ResetActionWhenChanged(...)`, `FieldsRefreshReset(...)`, and `FieldsRefreshResetWhenChanged(...)` provide standard Reset actions for multi-field filter bars. Reset emits the caller-provided intent so the application can restore business defaults such as `status=actionable`, `source=all`, page number, and selection.

## Example

```go
filters := filterbar.NewBuilder().
    Title("Provider Filters").
    Summary("scope: production | page: provider tokens").
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
    Action(filterbar.Button("export", "Export", ExportIntent{}).WithDisabledReason("Select at least one provider first.")).
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
filterbar.ClearFieldAction("providerQuery", state.Query, state.Busy, "Providers are loading.")
filterbar.ClearFieldActionWithLabel("clear-history", "Clear History", "historyQuery", state.HistoryQuery, state.Busy, "Providers are loading.")
filterbar.ResetAction(ResetFiltersIntent{}, state.Busy, "Providers are loading.")
filterbar.ResetActionWhenChanged(ResetFiltersIntent{}, state.FiltersChanged, state.Busy, "Providers are loading.")
filterbar.Button("reset", "Reset", ResetFiltersIntent{})
filterbar.Button("clear", "Clear", ClearIntent{}).WithDisabled(true)
filterbar.Button("export", "Export", ExportIntent{}).WithDisabledReason("Select rows first.")
filterbar.Button("load", "Load", LoadIntent{}).
    WithDisabledReasonIf(state.Lookup == "", "Enter a lookup id.").
    WithDisabledReasonIf(state.Busy, "Data is loading.")
filterbar.Button("load", "Load", LoadIntent{}).
    WithDisabledReasons(loadDisabledReasons(state)...)
```

Use `SearchRefreshClear(...)` for common data pages with one search field:

```go
filterbar.SearchRefreshClear(
    "providers.filters",
    filterbar.CompactPageSummary(state.Page, state.Total, state.Query, 28),
    126,
    6,
    filterbar.Search("query", "Search", state.Query).ForField("providerQuery"),
    RefreshIntent{},
    state.Busy,
    "Providers are loading.",
)
```

Use `FieldsRefreshReset(...)` for multi-field pages where resetting is not the
same as clearing every field:

```go
filterbar.FieldsRefreshReset(
    "alerts.filters",
    alertsSummary,
    126,
    8,
    fields,
    RefreshIntent{},
    ResetAlertFiltersIntent{},
    state.Busy,
    "Alerts are loading.",
)
```

The Reset intent should be handled by the application reducer. Components should
not guess default values for select fields or decide whether a reset reloads
server-side data.

Use `FieldsRefreshResetWhenChanged(...)` when the host can tell whether the
current filter scope differs from defaults. It disables Reset as `Nothing to
reset.` while leaving Refresh available.

High-risk actions should not live in a filter bar. Keep them in selection-specific detail panels or confirmations where the target and impact are visible.
Use disabled reasons for real prerequisites only: missing selection, permission boundaries, loading state, unsupported backend operation, or incomplete filters.

## Summary Helpers

Use the summary helpers to keep page scope text stable across pages:

```go
filterbar.Summary(
    filterbar.SummaryCount("page", state.Page),
    filterbar.SummaryCount("total", state.Total),
    filterbar.SummaryValueUnless("status", state.Status, "all"),
    filterbar.SummaryPresence("reason", state.OperationReason, "ready", "missing"),
    filterbar.SummaryCompactSearch(state.Search, 28),
)

filterbar.CompactPageSummary(state.Page, state.Total, state.Search, 28)

filterbar.LookupSummary(filterbar.LookupSummaryConfig{
    Lookup:        state.LookupID,
    LookupFallback: "required",
    LookupWidth:   36,
    Source:        state.LookupSource,
    SourceFallback: "none",
    Resolved:      state.ResolvedBy,
    ResolvedFallback: "pending",
    ResolvedWidth: 16,
    ItemsLabel:    "spans",
    Items:         state.SpanCount,
    ErrorsLabel:   "errors",
    Errors:        state.ErrorCount,
})
```

`CompactPageSummary(...)` only shortens the display value in the summary. The
caller should keep the full search value in the field state and API query.
`LookupSummary(...)` is intended for drilldown pages where the lookup target may
come from typed input, a selected row, or another page context. It shows blank
lookup/source/resolved values as `-` by default and clamps negative counts to
zero. Use `LookupFallback`, `SourceFallback`, and `ResolvedFallback` when a
missing lookup should read as an actionable state such as `required`, `none`, or
`pending` instead of a bare placeholder.
