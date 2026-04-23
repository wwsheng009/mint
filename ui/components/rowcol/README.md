# Row / Col

24 栅格 Flex 行列布局，适合表单编排、卡片网格和左右分栏。

## 已支持

- `Row` / `Col` 双节点模型
- 24 栅格 `span`
- `offset`
- 横向 / 纵向 `gutter`
- `wrap`
- `justify` / `align`

## 示例

```go
ui.NewRowBuilder().
    Gutter(2, 1).
    Children(
        ui.NewColBuilder().
            Span(8).
            Children(ui.Text("Sidebar")).
            Build(),
        ui.NewColBuilder().
            Span(16).
            Children(ui.Text("Content")).
            Build(),
    ).
    Build()
```

当前根包提供 `NewRowBuilder()` 和 `NewColBuilder()`，没有额外的 `ui.Row()` 快捷函数。
