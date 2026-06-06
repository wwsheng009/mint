# Modal

`ui/components/modal` provides a Fiber-first modal dialog with:

- dedicated header rendering
- body surface fill and optional shadow
- explicit content padding
- independent close policies for `ESC` and backdrop clicks
- stable topmost-first behavior when multiple modals are open

## Basic usage

```go
modal.NewBuilder().
    Title("Settings").
    Content(ui.Text("Edit your preferences here.")).
    Padding(1).
    InnerSize(36, 6).
    Rounded().
    Open(true).
    OnClose(CloseModalIntent{}).
    Build()
```

The root `ui` package also exposes `ui.ModalOf(...)`, `ui.ModalOfSize(...)`,
`ui.ModalTitled(...)`, explicit opened helpers `ui.ModalOpenedOf(...)`,
`ui.ModalOpenedOfSize(...)`, `ui.ModalOpenedTitled(...)`, and static helpers
such as `ui.ModalConfirm(...)`.

`ModalOf(...)` / `ModalTitled(...)` build a declarative modal node and keep the
default `Open(false)` state. For an immediate popup, use `Opened()` on the
builder or the opened shortcuts above. For business dialogs that need title,
footer actions, close policies, or explicit size, prefer `ui.NewModalBuilder()`
or `modal.NewBuilder()`.

```go
ui.ModalOpenedTitled(
    "Provider Picker",
    ui.List().
        Rows([]string{"anthropic", "openai", "vertex"}).
        SortAscending().
        MaxRows(6).
        ShowBorder(true).
        Build(),
)
```

## Operational content patterns

Modal is only the container. The content can be a form, list, table, or any
other `VNode`; choose the content component by operation intent:

- Use `FormDialog` for short bounded forms such as an audit reason or one
  operation parameter. It opens by default and provides title, body, targets,
  submit/cancel footer, and close policies.
- Use `ModalOpenedTitled(...)` with `List` for short single-column pickers such
  as provider/action/scope selection. `List.SortAscending()`,
  `List.SortDescending()`, or `List.SortRows(enabled, descending)` sort by
  row text or `RowItem.Title`.
- Use `ModalOpenedTitled(...)` with `DataTable` for multi-column comparison,
  stable row keys, pagination, activation, or column sorting.
- Use `Drawer` instead of `Modal` for longer create/edit workflows where the
  operator should keep the underlying list context visible.

Short form:

```go
ui.FormDialogDangerReasonAction(
    "runtime.reload.reason",
    "Reload Runtime",
    "Reload runtime configuration after checking the target endpoint.",
    "reload-runtime",
    "reloadReason",
    state.ReloadReason,
    "Prepare Reload",
    prepareReloadIntent{},
    closeDialogIntent{},
)
```

Sorted list popup:

```go
providers := ui.List().
    Header("Providers").
    Items([]ui.ListItem{
        ui.NewListItem("openai").WithPrefix("[ok]").WithDescription("healthy"),
        ui.NewListItemWithDescription("anthropic", "degraded").WithPrefix("[warn]"),
    }).
    SortAscending().
    MaxRows(8).
    ShowBorder(true).
    Build()

popup := ui.ModalOpenedTitled("Select Provider", providers)
```

Sortable table popup:

```go
tableView := ui.DataTable(
    []ui.TableColumn{
        {Title: "provider", Width: 18, Sortable: true},
        {Title: "status", Width: 12, Sortable: true},
        {Title: "available", Width: 10, Sortable: true},
    },
    [][]string{
        {"openai", "healthy", "12/12"},
        {"anthropic", "degraded", "7/10"},
    },
    ui.DataTableComponentID("provider.picker.table"),
    ui.DataTableRowKeys([]string{"provider:openai", "provider:anthropic"}),
    ui.DataTableSelectedKey(state.SelectedProviderKey),
    ui.DataTableSelectedKeyField("selectedProviderKey"),
    ui.DataTableActivateKeyField("activatedProviderKey"),
    ui.DataTableSortBy(state.ProviderSortColumn, state.ProviderSortDescending),
    ui.DataTableOnChange(providerPickerTableChangeIntent{}),
    ui.DataTableOperationalStyle(),
)

popup := ui.ModalOpenedTitled("Compare Providers", tableView)
```

For service-side pagination or service-side sorting, keep the modal table
controlled with `DataTableComponentID(...)`, `DataTableCurrentPage(...)`,
`DataTableSortBy(...)`, and `DataTableOnChange(...)`; only translate sort state
to API query parameters when the backend contract explicitly supports those
fields.

## Static helpers

`Confirm / Info / Success / Warning / Error / Alert` 现在除了默认标题和 footer 之外，还支持静态模板选项：

- `WithConfirmText(...)` / `WithCancelText(...)`
- `WithConfirmVariant(...)` / `WithCancelVariant(...)`
- `WithHelperPrefix(...)` / `WithHelperPrefixNode(...)`
- `WithFooterLayout(...)`
- `WithConfirmIntent(...)` / `WithCancelIntent(...)`

示例：

```go
modal.Confirm(
    "Delete Item",
    "This action cannot be undone.",
    modal.WithHelperPrefix("[!]"),
    modal.WithConfirmText("Delete"),
    modal.WithConfirmVariant(button.VariantDanger),
    modal.WithFooterLayout(modal.StaticFooterLayoutCenter),
)
```

## Main builder options

- `Title(string)`: sets the header title.
- `Content(rtui.VNode)`: sets the main body.
- `Footer(rtui.VNode)`: sets a footer area rendered below content.
- `Padding(int)`: adds inner space between the border and child content.
- `Size(w, h)`: sets outer size.
- `InnerSize(w, h)`: sets content size and automatically includes border/header/padding chrome.
- `Centered(bool)`: controls fixed centered display vs normal flow layout.
- `Closeable(bool)`: master switch for middleware-driven close behavior.
- `CloseOnEsc(bool)`: enables or disables `ESC` closing.
- `CloseOnBackdrop(bool)`: enables or disables click-outside closing.
- `Shadow(bool)`: enables or disables modal shadow rendering.
- `ShadowStyle(style.Style)`: customizes shadow color/style.
- `Style(style.Style)`: customizes the modal body/chrome style.
- `OnClose(intent.Intent)`: emits an intent when middleware closes the modal.

## Behavior notes

- If a modal is open, the middleware targets the topmost modal first.
- Clicking outside a modal is swallowed even when backdrop-close is disabled, so the background stays blocked.
- Header text is rendered with display-width awareness, so wide characters remain aligned.
- The modal instance provides its own `BoxModel` and `FlexStyle`, which means header rows and footer layout are accounted for by the layout engine instead of only by paint-time drawing.
- Static helpers remain ordinary `*Builder`, so helper模板和普通 builder 配置可以继续叠加。

## Middleware

Register the modal middleware once in app setup:

```go
ui.WithPluginSetup(func(app *framework.App) {
    app.AddMiddleware(modal.NewModalMiddleware())
})
```

Without the middleware, the modal still renders, but global `ESC` and backdrop-close behavior will not run.
