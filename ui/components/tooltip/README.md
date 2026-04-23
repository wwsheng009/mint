# Tooltip

提示气泡组件，适合按钮说明、状态补充和 hover 提示。

## 已支持

- 12 方位 placement
- `auto` 与显式 placement 回退
- 视口内 clamp，避免最终落点完全跑出可见区域
- `delay`
- layer 控制
- 前景 / 背景色样式

## 状态语义

- 默认走组件内 hover 显隐逻辑，不需要额外安装中间件。
- `Auto()` 与显式方位都会复用统一的 fallback / clamp 计算，边界场景不再各自走一套定位逻辑。
- `Delay(...)` 控制显示延迟；layer 默认是 tooltip layer，也可以通过 `OverlayLayer()`、`ModalLayer()` 等切换。
- 组件版推荐使用 `ui.NewTooltipBuilder(...)`；`ui.Tooltip(...)` 仍保留为旧 layer helper。

## 示例

```go
ui.NewTooltipBuilder(
    ui.Text("Save"),
    "Save current changes",
).
    TopRight().
    Delay(0).
    Build()
```

## 测试入口

- 单测：`go test ./ui/components/tooltip`
- 重点覆盖：`tooltip_test.go` 中的 `delay` timer 生命周期、runtime children hover 路径，以及 top/bottom 与 left/right 显式 placement 的共享 candidate fallback 计算；`right-top` / `left-bottom` 的镜像回退，以及在两侧 horizontal family 都放不下时继续回退到 `top` / `bottom`；`top-left` / `top-right` / `bottom-left` / `bottom-right` 的 corner family 回退，以及对应 corner clamp；极窄 viewport 下显式 `top-left` / `top-right` / `bottom-left` / `bottom-right` 仍保持原 vertical family 的 left-edge clamp；当 viewport 过窄且过矮、没有任何 vertical candidate 能完整放入时，仍会通过双轴 clamp 保持 `top` / `bottom` family
- E2E：`go test ./ui/e2e -run TestE2ETooltip`
- 重点覆盖：hover 显隐、`delay` 延迟显示、顶边 `top` fallback、右边界 `right` fallback，以及 `right-top` 角落 placement 的镜像回退；跨组件一致性回归已覆盖顶边 fallback、左右边界下显式 `top` / `bottom` 的 family 内横向回退，以及 `top-left` / `top-right` 在顶角场景下回退到下方同侧 family、`bottom-left` / `bottom-right` 在底角场景下回退到上方同侧 family，见 `go test ./ui/e2e -run TestE2EOverlay`
