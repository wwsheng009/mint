# Tag

标签组件，适合状态标记、筛选条件展示和轻量元数据提示。

## 已支持

- 多种颜色变体
- 自定义文本
- 可关闭
- 图标前缀
- `closeIntent`

## 示例

```go
ui.NewTagBuilder("Beta").
    Primary().
    Icon("*").
    Closable(true).
    Build()
```

快速创建：

```go
ui.Tag("Stable")
```
