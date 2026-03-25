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
