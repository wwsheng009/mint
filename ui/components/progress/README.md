# Progress

进度展示组件，适合任务进度、资源占用和阶段性完成度展示。

## 已支持

- `line` / `block` / `circle` / `dashboard`
- `normal` / `success` / `exception` / `active`
- `warning` 语义状态，适合 quota、pending_restart、rate_limited、degraded 等运维告警进度
- `value` / `max` 变化时的平滑百分比过渡
- `active` 状态逐帧动画
- `indeterminate` 不确定进度动画，适合 reload、reset、后台任务等待等无法得到总量的运维动作
- 自定义 `label`
- 自定义 `width`
- `showPercent`
- 运维语义快捷入口：`StatusForState(...)`、`ForState(...)`、`Usage(...)`、`UsageWithThresholds(...)`、`Busy(...)`、`Complete(...)`、`Failed(...)`

## 示例

```go
ui.NewProgressBuilder().
    Value(60).
    Max(100).
    Label("Uploading").
    Block().
    Active().
    Width(24).
    Build()
```

不确定进度：

```go
ui.NewProgressBuilder().
    Label("Reloading").
    Indeterminate().
    Width(24).
    Build()
```

快捷函数：

```go
ui.Progress(60, 100)
ui.ProgressIndeterminate("Reloading")
ui.ProgressForState("Config sync", 50, 100, "pending_restart")
ui.ProgressUsage("CPU", 82, 100)
ui.ProgressBusy("Reloading")
ui.ProgressComplete("Done")
ui.ProgressFailed("Failed")
```

## 运维语义映射

`StatusForState(...)` 面向运维页面的常见状态文案提供默认映射：

- `healthy` / `available` / `enabled` / `success` / `completed` / `in_sync` -> `success`
- `running` / `processing` / `loading` / `syncing` / `refreshing` / `reloading` -> `active`
- `degraded` / `rate_limited` / `pending_restart` / `pending` / `warning` / `lagging` / `queued` -> `warning`
- `unhealthy` / `disabled` / `unauthorized` / `unavailable` / `failed` / `error` / `blocked` / `out_of_sync` -> `exception`
- 未识别状态 -> `normal`

资源使用率默认使用 `80%` warning、`95%` exception 阈值；需要自定义阈值时使用 `UsageWithThresholds(...)`。

## 测试

- 单元测试：`go test ./ui/components/progress`
- E2E：`go test ./ui/e2e -run TestE2EProgress`
