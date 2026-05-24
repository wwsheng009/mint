# SplitPane

`SplitPane` arranges two panes in a stable left/right or top/bottom layout. It is intended for operations consoles, detail views, inspectors, and other screens where a fixed navigation or summary pane drives a flexible detail pane.

The component follows Mint's Fiber-first pattern:

- `VNode` stores declarative props only.
- `Instance` persists runtime state and implements `RuntimeChildrenProvider`.
- Runtime children are synthesized as standard `HStack`/`VStack`/`Box` nodes, so existing Fiber diffing and layout behavior remains the source of truth.

## Basic Usage

```go
pane := splitpane.NewBuilder().
    Key("ops").
    Width(96).
    Height(24).
    PrimarySize(28).
    Panes(groupList, providerDetails).
    Build()
```

For SDK users:

```go
pane := ui.NewSplitPaneBuilder().
    PrimarySize(28).
    Panes(groupList, providerDetails).
    Build()
```

## Vertical Layout

```go
pane := splitpane.NewBuilder().
    Vertical().
    Width(80).
    PrimarySize(6).
    Panes(summary, detail).
    Build()
```

## Size Model

- `PrimarySize(n)` fixes primary width in horizontal mode or primary height in vertical mode.
- `SecondarySize(n)` fixes secondary width or height.
- `PrimaryFlex(n)` and `SecondaryFlex(n)` pass flex factors to the generated pane boxes.
- The default keeps the primary pane at natural size and gives the secondary pane `flex=1`, which fits common master/detail screens.

## Separator

Separators are enabled by default.

```go
pane := splitpane.NewBuilder().
    Separator(true).
    SeparatorColor(style.Color("bright-black")).
    Panes(left, right).
    Build()
```

Use `Separator(false)` for compact layouts, or `SeparatorGlyph(...)` when an application needs a custom divider.
