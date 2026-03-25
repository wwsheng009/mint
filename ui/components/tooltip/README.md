# Tooltip

提示气泡组件，适合按钮说明、状态补充和 hover 提示。

## 已支持

- 12 方位 placement
- `auto` 与显式 placement 回退
- 视口内 clamp，避免最终落点完全跑出可见区域
- `delay`
- layer 控制
- 前景 / 背景色样式

## 示例

```go
ui.NewTooltipBuilder(
    ui.Text("Save"),
    "Save current changes",
).
    TopRight().
    Build()
```

组件版推荐使用 `ui.NewTooltipBuilder(...)`；`ui.Tooltip("...")` 仍保留为旧 layer helper。
