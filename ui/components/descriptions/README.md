# Descriptions

`descriptions/` 提供键值对展示组件，适合详情页、摘要面板和只读属性列表。

支持：

- 多列布局与 `span`
- horizontal / vertical 两种 item 布局
- 可选标题与 header extra
- bordered / plain 两种外观

## 示例

```go
ui.NewDescriptionsBuilder().
    Title("Build Info").
    Column(2).
    Item(descriptions.Field("Version", "v1.2.3")).
    Item(descriptions.Field("Commit", "308cc4b5").WithSpan(2)).
    Build()
```
