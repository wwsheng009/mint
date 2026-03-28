# Popover

`popover/` 提供比 `Tooltip` 更丰富的浮层卡片，适合解释信息、详情补充和轻量操作提示。

支持：

- `title` + `body`
- `click` / `hover` / `manual` 三种触发模式
- `PlacementAuto` + top / bottom 系列 6 个 placement：`top/topLeft/topRight`、`bottom/bottomLeft/bottomRight`
- `auto` 与显式 placement 共享同一套 fallback / clamp 计算，顶边和左右边界场景不再各走一套定位逻辑
- 非受控 `InitialOpen(...)` 与受控 `Open(...)`
- `Install(app)` 后可统一收口 ESC / outside-click 关闭
- 本地 `PopoverToggleIntent` / `PopoverOpenIntent` / `PopoverCloseIntent`
- overlay 坐标换算已复用 `ui/components/internal/overlayposition`

## 状态语义

- `InitialOpen(...)` 用于非受控初始打开状态，`Open(...)` 用于显式受控打开状态。
- 建议始终配置稳定的 `ComponentID(...)`，这样 `ToggleWithID(...)`、`OpenWithID(...)`、`CloseWithID(...)` 和 `PopoverChangeIntent` 才能稳定路由到正确实例。
- `click` / `hover` / `manual` 三种触发模式共享同一套 overlay 定位与关闭链路。

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

## 测试入口

- 单测：`go test ./ui/components/popover`
- 重点覆盖：`popover_test.go` 中的触发模式、`PopoverChangeIntent`、共享定位计算与 middleware 关闭链路；显式 `top-left` / `top-right` 在顶角场景下回退到 `bottom-left` / `bottom-right`，`bottom-left` / `bottom-right` 在底角场景下回退到 `top-left` / `top-right` 的 corner 几何；极窄 viewport 下显式 `top-left` / `top-right` / `bottom-left` / `bottom-right` 仍保持各自 vertical family，除了 left-edge clamp 之外，也覆盖了“没有任何 vertical candidate 能完整放入时”的双轴 clamp 与箭头朝向
- E2E：`go test ./ui/e2e -run TestE2EPopover`
- 重点补充：arrow-capable dual-axis clamp 场景现在也有 dedicated e2e，验证 viewport 过窄且过矮、没有任何 vertical candidate 能完整放入时，`PopoverPlacementTop*` 仍保留 `▼`，`PopoverPlacementBottom*` 仍保留 `▲`
