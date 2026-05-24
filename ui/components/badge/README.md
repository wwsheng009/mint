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

## 状态徽标快捷构造

运维和后台场景中，状态字段可以使用 `ui/components/statusbadge` 提供的声明式组合能力；根包 `ui.StatusBadge(...)` 是面向 SDK 使用者的薄转发入口，它会把常见状态字符串映射为统一语义色：

```go
ui.StatusBadge("healthy")       // success
ui.StatusBadge("rate_limited")  // warning
ui.StatusBadge("unauthorized")  // error
ui.StatusBadge("syncing")       // processing
```

默认映射遵循：

- `healthy` / `active` / `available` / `effective` / `enabled`：正常
- `degraded` / `rate_limited` / `pending_restart` / `cooldown`：警告
- `unhealthy` / `disabled` / `unauthorized` / `unavailable` / `failed`：异常
- 其它值：中性

可用选项：

- `StatusBadgeTone(tone)`
- `StatusBadgeKey(key)`
- `StatusBadgeLabel(label)`
- `StatusBadgeText(text)`
- `StatusBadgeDot()`
- `StatusBadgeMapper(mapper)`
- `StatusBadgeLabelStyle(style)`
- `StatusBadgeStyle(style)`
