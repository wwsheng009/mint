# Alert

内联提示组件，适合表单校验提示、状态提醒和局部错误展示。

## 已支持

- `info` / `success` / `warning` / `error`
- 可选标题与正文
- `closable`
- `closeIntent`
- 自定义样式

## 示例

```go
ui.NewAlertBuilder("Disk usage is high").
    Title("Warning").
    Warning().
    Closable(true).
    Build()
```

快速创建也可以直接用：

```go
ui.Alert("Saved successfully")
```
