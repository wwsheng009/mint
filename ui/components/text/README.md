# Text

基础文本组件，适合普通文案、标题、标签和轻量样式化输出。

## 已支持

- 内容设置
- 前景 / 背景色
- `bold` / `italic` / `underline`
- padding
- 文本对齐
- `maxWidth`

## 示例

```go
ui.NewTextBuilder("Hello, Mint").
    Bold(true).
    FgColor("green").
    Padding(0, 1, 0, 1).
    Build()
```

快捷函数：

```go
ui.Text("Hello")
ui.TextMuted("helper text")
ui.TextMutedLines("status loaded", "Graphics: inline image unsupported")
ui.BoolText(true) // "yes"
ui.CompactText(providerName, 20) // "-" for blank, ellipsis when too long
ui.EnabledText(provider.Enabled) // "enabled" or "disabled"
ui.FirstNonEmptyText(providerName, providerID, "unknown")
ui.KeyValueSummaryText(ui.KeyValueTextPart{Label: "retries", Value: "3"}) // "retries=3"
ui.ListText(groups...) // "default, backup" or "-"
ui.NonZeroIntText(pid) // "1234" or ""
ui.OptionalCompactText(scopeText, 80) // "" for blank optional helper text
ui.OptionalUnitText(durationMS, hasDuration, "ms") // "13ms" or "-"
ui.OptionalTrimmedDecimalText(threshold, hasThreshold, 2) // "12.35" or "-"
ui.SlashText(provider, group, keyID) // "provider/group/key" with blank segments as "-"
ui.ShortTimeText(syncAt) // "15:04:05" or "-"
ui.ServerSortScopeText("runtime", true) // "server runtime desc"
ui.CurrentPageSortScopeText("latency", false) // "current page latency asc"
ui.ColumnSortScopeText("rows", sortColumn, sortDescending, "provider", "load", "avg wait")
```

`TextMutedLines(...)` 适合把同一语义下的多行低强调提示组合为一个稳定节点；空行会被忽略，全部为空时返回 `nil`，便于 `StackPanel` 这类容器自动过滤可选说明。

`FirstNonEmptyText(...)` 适合从多个 API 字段或派生值中选择第一个有效展示值；每个候选都会按显示文本规则裁剪空白和换行，全部为空时返回空串，便于调用方继续传入业务兜底值。

`CompactText(...)` 适合表格单元格、摘要字段这类必须占位的值；空值会显示为 `-`。`OptionalCompactText(...)` 适合表格 scope、面板 hint、可选辅助说明这类“没有生效就应整行省略”的值；空白输入保留为空串，非空时同样按显示宽度限宽。

`ListText(...)` 适合展示 provider groups、applied filters、unsupported filters 等 API 返回的短字符串列表；空白项会被跳过，全部为空时显示 `-`，避免列表摘要出现空逗号或多余分隔符。

`KeyValueSummaryText(...)` 适合展示紧凑的 `key=value / other=value` 运维摘要，例如重试次数/延迟、对象计数和降级范围；label 会被规范化，空 label 会被跳过，空 value 显示为 `-`。

详情面板空态中需要把恢复动作和当前范围组合成 hint 时，使用 `ui.DetailPanelEmptyHint(...)`，它会基于同一套 `KeyValueTextPart` 规范化 scope，并跳过空 value，避免每个页面手写 `Scope: ...` 拼接。

`EnabledText(...)` 适合表格、摘要和短状态中展示开关态；需要颜色语义的详情字段继续使用 `NewDescriptionsEnabledValue(...)`。

`ShortTimeText(...)` 适合状态栏、摘要行和表格中紧凑展示本地同步时间、刷新时间等时间点；零值时间显示为 `-`。

`NonZeroIntText(...)` 适合 PID、版本号、HTTP status code 等 `0` 表示未返回或不适用的字段；零值和负值返回空串，交给外层 MetricRow、Descriptions 或 `firstNonEmpty(...)` 处理最终兜底。

`OptionalUnitText(...)` 和 `OptionalTrimmedDecimalText(...)` 适合 API 使用 `has_value` / `has_threshold` 这类显式标记表示数值是否存在的场景；标记为 false 时显示 `-`，标记为 true 时复用对应的单位或裁剪小数格式。

`SlashText(...)` 适合把已经格式化好的运维显示片段组合为路径、阈值对或其它 `a/b/c` 摘要；每个片段都会按显示文本规则规范化，空片段显示为 `-`，避免页面各自手写 `/` 拼接时出现空段语义不一致。

服务端分页或混合分页表需要在 summary/pagination 中标明排序作用域时，使用 `ServerSortScopeText(...)` 或 `CurrentPageSortScopeText(...)`。前者适合业务层已确认 API 支持并实际下发 `sort_by`/`sort_order` 的列，后者适合只重排当前已加载 rows 的本地扫描，避免操作者误以为跨页全量结果已经排序。

当业务状态只保存表格列号时，使用 `ColumnSortScopeText(...)`、`ServerColumnSortScopeText(...)` 或 `CurrentPageColumnSortScopeText(...)`，按列顺序传入可展示标签；负数列、越界列或空标签会返回空串，便于未排序状态直接省略 scope。第 0 列是有效列，只有负数才表示未排序。
