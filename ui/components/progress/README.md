# Progress

进度展示组件，适合任务进度、资源占用和阶段性完成度展示。

## 已支持

- `line` / `block` / `circle` / `dashboard`
- `normal` / `success` / `exception` / `active`
- `value` / `max` 变化时的平滑百分比过渡
- `active` 状态逐帧动画
- `indeterminate` 不确定进度动画，适合 reload、reset、后台任务等待等无法得到总量的运维动作
- 自定义 `label`
- 自定义 `width`
- `showPercent`

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
```

## 测试

- 单元测试：`go test ./ui/components/progress`
- E2E：`go test ./ui/e2e -run TestE2EProgress`
