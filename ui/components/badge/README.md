# Badge

徽标组件，适合消息计数、状态点和列表项上的轻量提示。

## 已支持

- 数字 `count`
- 自定义 `text`
- `dot` 模式
- `showZero`
- `overflowCount`
- `status` 变体
- label 与 badge 组合展示

## 示例

```go
ui.NewBadgeBuilder("Inbox").
    Count(120).
    OverflowCount(99).
    ShowZero(true).
    Error().
    Build()
```

如果只需要快速创建，也可以直接用：

```go
ui.Badge("Inbox", 12)
```
