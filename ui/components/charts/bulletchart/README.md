# BulletChart

`bulletchart` 是第一批优先实现的图表组件之一。

目标定位：

- 当前值 vs 目标值
- KPI / SLA / 预算执行率展示
- 终端中比仪表盘式 gauge 更高密度的表达

当前已实现：

- 单指标
- 可配置 `target marker`
- `qualitative ranges`
- 自适应 `value label` 布局
- 显式方向语义 `higher / lower / neutral`

最小示例：

```go
bulletchart.NewBuilder().
    Label("Latency").
    Value(173).
    Target(200).
    Max(250).
    Width(22).
    Build()
```

带阈值区间和下置标签的示例：

```go
bulletchart.NewBuilder().
    Label("Availability").
    Value(996).
    Target(999).
    Max(1000).
    Width(22).
    BelowValueLabel().
    QualitativeRanges(
        bulletchart.QualitativeRange{Limit: 970},
        bulletchart.QualitativeRange{Limit: 990},
        bulletchart.QualitativeRange{Limit: 1000},
    ).
    TargetMarkerRune('┆').
    Build()
```

带方向语义的示例：

```go
bulletchart.NewBuilder().
    Label("Throughput").
    Value(82).
    Target(75).
    Max(100).
    Width(22).
    HigherIsBetter().
    BelowValueLabel().
    Build()
```

默认方向是 `neutral`，会保持中性的分段和目标线样式。

如果指标有明确方向，建议显式声明：

- `HigherIsBetter()`：默认分段从低到高映射为 `error -> warning -> success`
- `LowerIsBetter()`：默认分段从低到高映射为 `success -> warning -> error`
- `NeutralDirection()`：保持中性语义，不强行假设业务方向

显式 `QualitativeRange.Style` 和 `TargetMarkerStyle(...)` 仍然拥有最高优先级，会覆盖默认方向语义。

实现建议：

- 尽量复用 `progress` 在样式和动画上的经验
- 但保持单独组件语义，不直接复用 `progress` 对外 API
- 宽度受限时优先保证图条和目标线可读，再把值标签降级到下一行
