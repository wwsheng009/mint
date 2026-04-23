# StatusBar

`statusbar` 是一个轻量级复合组件，用于页头、页脚、模式栏、快捷提示栏等单行状态信息展示。

## 能力

- 三槽布局：`LeftSections(...)` / `CenterSections(...)` / `RightSections(...)`
- 追加式 API：`Left(...)` / `Center(...)` / `Right(...)`
- 固定宽度节：`Section.Width` 支持按显示宽度补齐或截断
- 宽字符安全：宽度计算基于终端显示宽度，中文等双宽字符可正确对齐
- 溢出控制：`WithEllipsis()` / `WithClip()` / `WithOverflow(...)`
- Interactive sections: `OnPress(...)` turns the status bar into a lightweight toolbar
- Keyboard accessible: interactive sections can be reached with `Tab` and triggered with `Enter`
- Hover feedback: interactive sections highlight on mouse hover when mouse capture is enabled
- Inline tooltip/help: `WithHelp(...)` / `WithTooltip(...)` + `BuildWithHelp()`
- Help display modes: `HelpDisplayInline` / `HelpDisplayOverlay` / `HelpDisplayBoth`
- Overlay placement: `TooltipPlacementAuto` / `TooltipPlacementTop` / `TooltipPlacementBottom`
- Overlay fallback / clamp: `TooltipPlacementAuto` 会按可见空间在上下方间回退，横向候选与视口 clamp 已复用 `ui/components/internal/overlayposition`；显式 `TooltipPlacementTop` / `TooltipPlacementBottom` 在左右边界场景下都会保持各自 family，并分别向左/向右选择合适的横向候选；在顶角 / 底角这类双轴受限场景下，则会在保留 edges-first 候选顺序的前提下回退到对侧 vertical family
- Overlay viewport source: tooltip-layer 下的 overlay help 现在会优先使用运行时注入的真实 viewport，而不是依赖局部 layout bounds，因此顶角 / 底角 fallback 不会再因为零尺寸 layer bounds 而把内容画到可视区外
- Overlay wrapping: `TooltipMaxWidth(...)` limits content width and wraps to multiple lines
- Overlay gap: `TooltipGapRows(...)` adds a clearer vertical separation from the anchor
- Overlay visibility: overlay tooltips only appear while the mouse is hovering the section
- Overlay bubble arrow: tooltips render an anchor-pointing arrow by default
- Overlay arrow themes: `TooltipArrowStyleSharp` / `TooltipArrowStyleRounded`
- 主题默认值：`Theme(...)` 可为未显式着色的节提供统一样式
- 外层留白：`Padding(...)` 可用于和正文区域分隔

## Section API

```go
type Section struct {
    Key         string
    Text        string
    FgColor     string
    BgColor     string
    Bold        bool
    Width       int
    Align       ui.Align
    Overflow    statusbar.OverflowMode
    PressIntent intent.Intent
    Disabled    bool
    HelpText    string
}
```

### 常用构造

```go
statusbar.Text(" READY ")
statusbar.Badge(" INSERT ", "black", "green")
statusbar.ActionBadge(" SAVE ", "black", "green", SaveIntent{})
statusbar.Text("Long path").WithWidth(12).WithEllipsis()
statusbar.Text("Details").OnPress(OpenDetailIntent{}).WithHelp("Open detail panel")
```

### Section 链式方法

- `WithColors(fg, bg)`
- `WithForeground(color)`
- `WithBackground(color)`
- `WithBold(bool)`
- `WithWidth(width)`
- `WithAlign(ui.AlignStart | ui.AlignCenter | ui.AlignEnd)`
- `WithEllipsis()`
- `WithClip()`
- `WithOverflow(statusbar.OverflowEllipsis | statusbar.OverflowClip)`
- `OnPress(intent.Intent)`
- `WithDisabled(bool)`
- `WithKey(string)`
- `WithHelp(text)`
- `WithTooltip(text)`

## Theme API

```go
type Theme struct {
    FgColor       string
    BgColor       string
    Bold          bool
    HoverStyle    style.Style
    FocusStyle    style.Style
    PressedStyle  style.Style
    DisabledStyle style.Style
    HelpStyle          style.Style
    TooltipBorderStyle style.Style
    TooltipShadowStyle style.Style
}
```

内置主题：

- `statusbar.DefaultTheme()`
- `statusbar.MutedTheme()`
- `statusbar.ContrastTheme()`
- `Theme.WithHoverStyle(...)` / `WithFocusStyle(...)` / `WithPressedStyle(...)` / `WithDisabledStyle(...)` / `WithHelpStyle(...)`
- `Theme.WithTooltipBorderStyle(...)` / `WithTooltipShadowStyle(...)` / `WithTooltipArrowStyle(...)`
- `Builder.HelpDisplayMode(statusbar.HelpDisplayInline | HelpDisplayOverlay | HelpDisplayBoth)`
- `Builder.TooltipPlacement(statusbar.TooltipPlacementAuto | TooltipPlacementTop | TooltipPlacementBottom)`
- `Builder.TooltipMaxWidth(width)`
- `Builder.TooltipGapRows(rows)`
- `Theme.WithTooltipArrowStyle(statusbar.TooltipArrowStyleSharp | statusbar.TooltipArrowStyleRounded)`

## Builder 示例

```go
bar := statusbar.NewBuilder().
    DefaultTheme().
    HelpDisplayMode(statusbar.HelpDisplayOverlay).
    Padding(0, 1, 0, 1).
    Left(statusbar.ActionBadge(" MODE ", "black", "yellow", CycleModeIntent{}).WithHelp("Cycle to the next mode")).
    Left(statusbar.ActionText(" Interactive ", CycleModeIntent{}).WithWidth(18).WithEllipsis().WithHelp("Current mode")).
    Center(statusbar.Text("Mint Demo").WithWidth(24).WithAlign(ui.AlignCenter).WithBold(true)).
    Right(statusbar.ActionBadge(" UI ", "black", "yellow", SetModeIntent{Mode: "interactive"}).WithHelp("Regular UI interaction mode")).
    Right(statusbar.ActionBadge(" APP ", "bright-white", "bright-black", SetModeIntent{Mode: "app_selection"}).WithHelp("App-managed copy mode")).
    Right(statusbar.ActionBadge(" TERM ", "bright-white", "bright-black", SetModeIntent{Mode: "terminal_selection"}).WithHelp("Native terminal selection mode")).
    BuildWithHelp()
```

### Help / Tooltip 行

```go
bar := statusbar.NewBuilder().
    Theme(statusbar.DefaultTheme().WithTooltipArrowStyle(statusbar.TooltipArrowStyleRounded)).
    HelpDisplayMode(statusbar.HelpDisplayOverlay).
    TooltipPlacement(statusbar.TooltipPlacementAuto).
    TooltipGapRows(1).
    TooltipMaxWidth(38).
    HelpPrefix("? ").
    HelpFallback("Hover actions to inspect overlay help").
    Left(statusbar.ActionText(" Open ", OpenIntent{}).WithHelp("Open current file")).
    Right(statusbar.ActionText(" Save ", SaveIntent{}).WithHelp("Save current document")).
    BuildWithHelp()
```

`BuildWithHelp()` 会根据 `HelpDisplayMode(...)` 决定是否渲染 inline help、overlay tooltip 或两者同时渲染。

- Inline help: hovered first, focused second, fallback last
- Overlay tooltip: only appears while the mouse is hovering the section
- Both: inline keeps keyboard fallback semantics, overlay remains hover-only

## 测试入口

- 单测：`go test ./ui/components/statusbar`
- 重点覆盖：overlay tooltip 的 wrap、arrow theme、`TooltipPlacementAuto` 的 bottom bias、显式 `TooltipPlacementTop` / `TooltipPlacementBottom` 在左右边界的 family 内横向回退，以及顶角 / 底角下回退到对侧 vertical family 的双轴受限几何；overlay help 会优先读取真实 viewport；极窄 viewport 下显式 `TooltipPlacementTop` / `TooltipPlacementBottom` 的 left-edge clamp 仍保持原 vertical family；当 viewport 过窄且过矮、同侧和对侧候选都放不下时，也会通过双轴 clamp 保持 `Top` / `Bottom` family
- E2E：`go test ./ui/e2e -run "TestE2EStatusbar|TestE2EOverlay"`
- 重点覆盖：hover 显隐、overlay help 跟随 hovered section、左右边界下显式 `TooltipPlacementTop` / `TooltipPlacementBottom` 的横向候选回退；顶角 / 底角双轴受限场景现在既有 dedicated `statusbar` e2e，也已纳入 `overlay_consistency_e2e_test.go` 的跨组件 corner 回归，覆盖回退到对侧 vertical family 后仍保持左右侧 family 语义；另外还补了极窄且过矮 viewport 下的 dedicated dual-axis clamp e2e，验证 `TooltipPlacementTop` / `TooltipPlacementBottom` 在没有任何 vertical candidate 能完整放入时仍保留原箭头方向

## `ui` 顶层快捷入口

```go
bar := ui.StatusBarWithHelpMode(
    ui.StatusBarThemeDefault(),
    "Hover actions to inspect overlay help",
    ui.StatusBarHelpOverlay,
    ui.StatusBarSections(
        ui.StatusBarActionBadge(" MODE ", "black", "yellow", CycleModeIntent{}).WithHelp("Cycle to the next mode"),
        ui.StatusBarActionText(" Interactive ", CycleModeIntent{}).WithWidth(18).WithEllipsis().WithHelp("Current mode"),
    ),
    ui.StatusBarSections(
        ui.StatusBarText("Mint Demo").WithWidth(24).WithAlign(ui.AlignCenter).WithBold(true),
    ),
    ui.StatusBarSections(
        ui.StatusBarActionBadge(" UI ", "black", "yellow", SetModeIntent{Mode: "interactive"}).WithHelp("Regular UI interaction mode"),
        ui.StatusBarActionBadge(" APP ", "bright-white", "bright-black", SetModeIntent{Mode: "app_selection"}).WithHelp("App-managed copy mode"),
        ui.StatusBarActionBadge(" TERM ", "bright-white", "bright-black", SetModeIntent{Mode: "terminal_selection"}).WithHelp("Native terminal selection mode"),
    ),
)
```

## 适用场景

- 交互模式状态栏
- 编辑器 / IDE 底部信息栏
- 文件浏览器路径栏
- 列表或表格的快捷键提示栏
- 终端友好的 tooltip/help 提示栏
