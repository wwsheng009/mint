# Result

`result/` 提供结果状态页组件，适合成功、失败、权限不足或资源不存在等场景的集中展示。

支持：

- `status` 预设：`info/success/warning/error/403/404/500`
- 自定义 `icon`
- `title` / `subtitle`
- `extra` 扩展操作区域
- `bordered` 样式

## 示例

```go
ui.NewResultBuilder().
    Status(result.StatusSuccess).
    Title("Saved successfully").
    Subtitle("Your configuration has been applied.").
    Extra(ui.NewButtonBuilder("Back").Build()).
    Build()
```
