# Popconfirm

`popconfirm/` 提供带确认/取消操作的气泡确认框，适合删除、重置和其他危险操作前的二次确认。

支持：

- `title` + `description`
- `click` / `hover` / `manual` 触发
- `OK` / `Cancel` 文案自定义
- 本地 open/toggle intents
- 全局 `PopconfirmConfirmIntent` / `PopconfirmCancelIntent`

## 示例

```go
ui.NewPopconfirmBuilder(
    ui.NewButtonBuilder("Delete").Danger().Build(),
).
    ComponentID("delete.confirm").
    Title("Delete this record?").
    Description("This action cannot be undone.").
    Build()
```
