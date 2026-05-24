# Descriptions

`descriptions/` 提供键值对展示组件，适合详情页、摘要面板和只读属性列表。

支持：

- 多列布局与 `span`
- horizontal / vertical 两种 item 布局
- 可选标题与 header extra
- bordered / plain 两种外观
- 稳定 `LabelWidth` / `ContentWidth`
- 空值占位 `EmptyText`
- 敏感值脱敏 `SensitiveField` / `WithSensitive`

## 示例

```go
ui.NewDescriptionsBuilder().
    Title("Build Info").
    Column(2).
    Item(descriptions.Field("Version", "v1.2.3")).
    Item(descriptions.Field("Commit", "308cc4b5").WithSpan(2)).
    Build()
```

## 运维详情示例

```go
ui.NewDescriptionsBuilder().
    Title("Provider Key").
    Column(2).
    LabelWidth(14).
    EmptyText("n/a").
    MaskText("masked").
    Item(descriptions.Value("Status", "healthy")).
    Item(descriptions.Value("Last Error", nil)).
    Item(descriptions.SensitiveField("Token", "agw_example_token")).
    Build()
```

对 token、provider key、authorization header 等敏感字段，优先使用 `SensitiveField` 或 `WithSensitive(true)`，不要把原始值渲染到终端。
