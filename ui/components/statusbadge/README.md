# StatusBadge

`statusbadge/` 是状态字段语义化徽标组合组件，基于 `badge/` 组件映射常见运维状态。

## 默认语义

- `healthy` / `active` / `available` / `effective` / `enabled`：正常
- `degraded` / `rate_limited` / `pending_restart` / `cooldown`：警告
- `unhealthy` / `disabled` / `unauthorized` / `unavailable` / `failed`：异常
- `info` / `processing` / `loading` / `syncing`：处理中
- 其它值：中性

## 示例

```go
node := ui.StatusBadge(
    "rate_limited",
    ui.StatusBadgeLabel("Provider key"),
    ui.StatusBadgeDot(),
)
```

## Fiber-first 约束

- 只执行纯状态字符串到 badge VNode 的映射。
- 可通过 `Mapper(...)` 注入自定义语义，组件自身不读取业务配置。
- 不保存状态、不触发副作用。

## 测试

```powershell
go test ./ui/components/statusbadge
```
