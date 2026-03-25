# Popconfirm

`popconfirm/` 提供带确认/取消操作的气泡确认框，适合删除、重置和其他危险操作前的二次确认。

支持：

- `title` + `description`
- `click` / `hover` / `manual` 触发
- `PlacementTop` 默认定位，也支持 `PlacementAuto` + top / bottom 系列 6 个 placement
- `OK` / `Cancel` 文案自定义
- confirm / cancel 按钮 variant 与 footer layout
- `Install(app)` 后统一收口 ESC / outside-click 关闭
- 本地 open/toggle intents
- 全局 `PopconfirmConfirmIntent` / `PopconfirmCancelIntent`
- overlay 坐标换算已复用 `ui/components/internal/overlayposition`

## 示例

```go
ui.NewPopconfirmBuilder(
    ui.NewButtonBuilder("Delete").Danger().Build(),
).
    ComponentID("delete.confirm").
    Title("Delete this record?").
    Description("This action cannot be undone.").
    OkVariant(button.VariantDanger).
    FooterLayout(popconfirm.FooterLayoutCenter).
    Build()
```

如果希望 ESC 和 outside-click 能关闭已打开的 Popconfirm，需要在应用启动时调用一次 `popconfirm.Install(app)`。
