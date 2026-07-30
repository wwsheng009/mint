# Progress

进度展示组件，适合任务进度、资源占用和阶段性完成度展示。

## 已支持

- `line` / `block` / `circle` / `dashboard`
- `normal` / `success` / `exception` / `active`
- `warning` 语义状态，适合 quota、pending_restart、rate_limited、degraded 等运维告警进度
- `value` / `max` 变化时的平滑百分比过渡
- `active` 状态逐帧动画
- `indeterminate` 不确定进度动画，适合 reload、reset、后台任务等待等无法得到总量的运维动作
- 默认 Unicode 字符集：line 使用 `━/·/●`，block 使用 `█/░/▓`，circle/dashboard 使用块元素和点状轨道
- ASCII 兼容字符集：通过 `ASCIIGlyphs()` 或 `GlyphStyle(GlyphStyleASCII)` 显式启用
- 自定义 `label`
- 自定义 `width`
- `showPercent`
- `showValue` + `unit`，用于展示 `used/total`、队列长度、quota、资源占用等运维值标签
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

ASCII 兼容输出：

```go
ui.NewProgressBuilder().
    Label("Uploading").
    Value(60).
    Max(100).
    ASCIIGlyphs().
    Build()
```

快捷函数：

```go
ui.Progress(60, 100)
ui.ProgressIndeterminate("Reloading")
ui.ProgressForState("Config sync", 50, 100, "pending_restart")
ui.ProgressForStateWithValue("Config sync", 50, 100, "pending_restart", "items")
ui.ProgressUsage("CPU", 82, 100)
ui.ProgressUsageWithValue("Queue", 7, 10, "jobs")
ui.ProgressBusy("Reloading")
ui.ProgressComplete("Done")
ui.ProgressFailed("Failed")
```

运维值标签：

```go
ui.NewProgressBuilder().
    Label("Queue").
    Value(7).
    Max(10).
    ShowValue(true).
    Unit("jobs").
    Width(16).
    Build()
```

渲染标签示例：`Queue: 7/10 jobs (70%)`。单位为 `%`、`ms`、`MB` 等紧凑后缀时会贴近数值，例如 `Latency: 42ms/100ms (42%)`。

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
