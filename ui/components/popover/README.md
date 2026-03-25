# Popover

`popover/` 提供比 `Tooltip` 更丰富的浮层卡片，适合解释信息、详情补充和轻量操作提示。

支持：

- `title` + `body`
- `click` / `hover` / `manual` 三种触发模式
- `PlacementAuto` + top / bottom 系列 6 个 placement：`top/topLeft/topRight`、`bottom/bottomLeft/bottomRight`
- 非受控 `InitialOpen(...)` 与受控 `Open(...)`
- `Install(app)` 后可统一收口 ESC / outside-click 关闭
- 本地 `PopoverToggleIntent` / `PopoverOpenIntent` / `PopoverCloseIntent`
- overlay 坐标换算已复用 `ui/components/internal/overlayposition`

## 示例

```go
ui.NewPopoverBuilder(ui.Text("Hover me")).
    Title("Mint UI").
    Body("Popover 比 Tooltip 更适合放多行说明。").
    Trigger(popover.TriggerHover).
    Placement(popover.PlacementBottomLeft).
    Build()
```

如果 anchor 本身是会吞掉点击的交互组件，例如 `Button`，可把按钮的 press intent 设为 `popover.ToggleWithID(...)`。

如果希望 ESC 和 outside-click 能关闭已打开的 Popover，需要在应用启动时调用一次 `popover.Install(app)`。
