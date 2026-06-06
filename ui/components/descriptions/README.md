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
- 部分掩码值 `MaskedValue` / `MaskedFallbackValue`
- 运维计数与比例字段 `CountValue` / `RatioValue`
- 运维详情组合：`Panel`、`ContextStrip`、`PanelWithContext`、`DetailPanel`

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
    Item(descriptions.CountValue("Retries", 0)).
    Item(descriptions.RatioValue("Queue", 3, 8)).
    Item(descriptions.Value("Last Error", nil)).
    Item(descriptions.SensitiveField("Token", "provider-token-demo")).
    Build()
```

对 token、provider key、authorization header 等敏感字段，优先使用 `SensitiveField` 或 `WithSensitive(true)`，不要把原始值渲染到终端。

如果运维人员需要识别同类对象，但不能看到完整账号、token 或 provider key，可以使用部分掩码：

```go
descriptions.MaskedValue("Account", "account-billing-prod", 2, 6)
descriptions.MaskedFallbackValue("Provider Key", "provider-key-demo", "-", 6, 4)
```

当详情面板依赖当前选中对象时，建议把对象身份和关键状态放在详情字段之前，避免操作者先扫一长串属性才确认目标。新页面优先使用语义化的 `DetailPanel`；`PanelWithContext` 保留为较底层的兼容组合入口：

```go
ui.DetailPanel(ui.DetailPanelConfig{
    Key:          "jobs.selection",
    Title:        "Job Detail",
    Width:        62,
    LabelWidth:   14,
    ContentWidth: 40,
    EmptyWhen: len(jobs) == 0,
    EmptyText: "No job selected.",
    EmptyHint: "Clear filters or refresh jobs.",
    Context: ui.DescriptionsContextStripConfig{
        Column:       3,
        LabelWidth:   7,
        ContentWidth: 10,
        Items: []ui.DescriptionsItem{
            ui.NewDescriptionsCompactValue("ID", job.ID, 10),
            ui.NewDescriptionsStateValue("Status", job.Status, job.Status),
            ui.NewDescriptionsEnabledValue("Enabled", job.Enabled),
        },
    },
    Items: []ui.DescriptionsItem{
        ui.NewDescriptionsCompactValue("Name", job.Name, 40),
        ui.NewDescriptionsCompactFallbackValue("Last Error", job.LastError, "-", 40),
    },
    Actions: []ui.VNode{actions},
})
```

`ContextStrip` 只负责展示选中对象的稳定身份、状态和下一步操作所需的少量关键信息。完整属性仍放在 details 区；操作按钮应放在面板底部或后续 drawer / dialog 中。

当列表为空、筛选后没有命中，或当前选中 key 已经不在当前数据页内时，`DetailPanelConfig.EmptyWhen` 可以把 context/details 替换成标准 empty 状态，同时保留 `Actions`。这样选择导航或当前对象动作仍能展示禁用原因，例如 loading、无可选项或缺少 trace id，而详情字段不会退化成一屏 `-`。

`DetailPanelConfig.EmptyHint` 可在 empty description 下追加低强调提示行，适合说明当前筛选 scope、恢复动作或数据来源，例如 `Clear filters or refresh jobs.`。Hint 为空时不渲染额外节点。需要同时展示恢复动作和当前范围时，优先用 `ui.DetailPanelEmptyHint(...)` 生成一致的提示文本：

```go
ui.DetailPanelEmptyHint(
    "Refresh jobs or reset job filters.",
    ui.KeyValueTextPart{Label: "status", Value: "active"},
    ui.KeyValueTextPart{Label: "last", Value: "failed"},
) // "Refresh jobs or reset job filters. Scope: status=active / last=failed"
```

`DetailPanelEmptyHint(...)` 会忽略空 value，空 scope 时只返回 action；`DetailPanelEmptyHintWithScopeWidth(...)` 可调整 scope 摘要的显示宽度，适合搜索词或 provider scope 很长的详情面板。
