# 约束追踪器使用指南

## 概述

约束追踪器 (`runtime/layout/tracer.go`) 是一个用于调试布局约束传播的工具。它可以记录约束在组件树中的传递过程，帮助开发者理解布局计算。

## 核心功能

- **追踪约束传递**：记录每个组件接收和传递给子组件的约束
- **记录测量结果**：保存组件的最终尺寸
- **追踪完整路径**：记录约束从根到叶子的完整传播路径

## API 参考

### 启用追踪

```go
// 启用全局追踪器
layout.EnableTracer()

// 禁用全局追踪器
layout.DisableTracer()

// 检查追踪器是否启用
if layout.IsTracerEnabled() {
    // 追踪器已启用
}
```

### 追踪约束

```go
layout.TraceMeasuring(
    from,      // 来源组件 ID，如 "border(my-panel)"
    to,        // 目标组件 ID，如 "content"
    path,      // 完整路径，如 "/app/main/panel/border"
    input,     // 输入约束（父组件传递给当前组件的约束）
    output,    // 输出约束（当前组件传递给子组件的约束）
    resultSize,// 测量结果尺寸
    reason,    // 约束修改原因，如 "Applied border padding (1x1)"
)
```

### 管理追踪数据

```go
// 清除追踪数据
layout.ClearTrace()

// 获取所有追踪条目
entries := layout.GetTraceEntries()

// 获取追踪条目数
count := layout.GetEntryCount()

// 输出追踪日志
log := layout.DumpTrace()
fmt.Println(log)
```

### 配置追踪器

```go
// 设置紧凑模式（省略详细信息）
layout.SetCompactMode(true)

// 设置是否显示路径
layout.SetShowPath(false)
```

## 使用示例

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/ui/components/panel"
)

func main() {
    // 启用追踪器
    layout.EnableTracer()
    defer layout.DisableTracer()

    // 创建并测量组件
    p := panel.New().SetTitle("My Panel").SetContent("Hello")
    inst := p.CreateInstance()

    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 24,
    }

    size := inst.Measure(constraints)

    // 输出追踪日志
    fmt.Println(layout.DumpTrace())
    fmt.Printf("Final size: %dx%d\n", size.Width, size.Height)
}
```

### 输出示例

```
╔══════════════════════════════════════════════════════════════════╗
║                    Constraint Propagation Trace               ║
╚══════════════════════════════════════════════════════════════════╝

Step 0
  Path: /app
  app → panel(main-panel)
  Input:    {0..80} × {0..24}
  Output:   {76..78} × {0..24}
  Dimension: 78dw × 6dh
  Reason:   Applied border padding (1x1), explicit width=0, height=0
```

### 在组件中集成

```go
type Instance struct {
    path string  // 用于约束追踪的路径
    // ...
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 计算子组件约束
    childConstraints := inst.computeChildConstraints(constraints)

    // 追踪约束传递
    if tagger, ok := inst.child.(interface{ Tag() string }); ok {
        layout.TraceMeasuring(
            "border("+inst.key+")",
            tagger.Tag(),
            inst.path+"/"+inst.key,
            constraints,
            childConstraints,
            layout.Size{},  // 尺寸将在测量后更新
            fmt.Sprintf("Applied border padding (%dx%d),
                borderWidth, borderWidth),
        )
    }

    // 测量子组件
    childSize := inst.measureChild(inst.child, childConstraints)

    // ...
}
```

### 设置路径

```go
// 为实例设置路径
inst.SetPath("/app/main/panel/border")

// Path 将被追踪器记录
layout.TraceMeasuring(
    ...
    inst.path,
    ...
)
```

## 输出格式

### 完整输出

```
Step 0
  Path: /app/main/panel/border
  border(my-border) → content
  Input:    {0..80} × {0..24}
  Output:   {75..78} × {0..22}
  Dimension: 78dw × 6dh
  Reason:   Applied border padding (1x1), explicit width=0, height=0
```

### 检测的问题

追踪器会自动检测并报告潜在的布局问题：

```
⚠️  Height 30 exceeds MaxHeight 24
```

## 故障排除

### 没有追踪数据出现

确保追踪器已启用：
```go
layout.EnableTracer()
```

检查组件是否调用了 `TraceMeasuring`：
```go
layout.TraceMeasuring("from", "to", ...)
```

### 追踪日志过多

使用紧凑模式减少输出：
```go
layout.SetCompactMode(true)
layout.SetShowPath(false)
```

### 内存占用过多

定期清除追踪数据：
```go
layout.ClearTrace()
```

## 最佳实践

### 1. 在调试时启用追踪

```go
func main() {
    // 仅在调试模式启用
    if os.Getenv("DEBUG_LAYOUT") != "" {
        layout.EnableTracer()
        defer layout.DisableTracer()
    }
    // ...
}
```

### 2. 使用有意义的 ID

```go
// ✅ 好的 ID
layout.TraceMeasuring("border(my-panel)", "content", ...)

// ❌ 模糊的 ID
layout.TraceMeasuring("border", "child", ...)
```

### 3. 清晰的原因说明

```go
// ✅ 清晰的原因
reason := fmt.Sprintf("Applied border padding (%dx%d), explicit width=%d",
    borderWidth, borderWidth, inst.width)

// ❌ 模糊的原因
reason := "calculated"
```

### 4. 路径层级合理

```go
// 从根节点开始
root.SetPath("/")
border.SetPath("/app/panel/border")
content.SetPath("/app/panel/border/content")
```

## 已集成追踪器的组件

- ✅ `ui/components/border/Instance` - `instance.go:410`

## 未来集成建议

建议在以下组件中也集成约束追踪器：

- `ui/components/panel/Instance` - Panel 组合组件
- `ui/components/stack/Instance` - VStack/HStack 布局
- `ui/components/text/Instance` - Text 文本组件

## 相关文档

- [布局系统优化计划 - Phase 1](/docsArchive/01-analysis.md)
- [约束系统设计](/docsArchive/layout/constraint_tracer_summary.md)
- [调试工具](/docsArchive/04-debug-tools.md)

## 故障案例

### 案例 1：约束意外的偏移

**症状**：组件没有使用预期的宽度

**调试步骤**：
1. 启用追踪器
2. 查看约束传递链
3. 检查每个步骤的 `Output` 约束
4. 找到非预期的修改点

**示例输出**：
```
Step 0
  parent → border
  Input:    {0..80} × {0..24}
  Output:   {0..78} × {0..24}  ← 注意宽度变为 78
  Reason:   Applied border padding (1x1)
```

### 案例 2：子约束不正确

**症状**：子组件被裁剪或溢出

**调试步骤**：
1. 检查子组件的 `Output` 约束
2. 验证是否正确减去了 padding
3. 检查显式维度是否正确应用

**示例输出**：
```
Step 1
  border → content
  Input:    {0..78} × {0..24}
  Output:   {0..76} × {0..22}  ← 子组件约束
  Reason:   Applied border padding (1x1)
```

### 案例 3：高度超出约束

**症状**：组件高度超过父约束

**调试步骤**：
1. 查看追踪日志中的警告
2. 检查 `Dimension` 是否超过 `Input.MaxHeight`
3. 调整组件的高度计算逻辑

**示例输出**：
```
Step 0
  parent → text
  Input:    {0..50} × {0..10}
  Output:   {0..48} × {0..8}
  Dimension: 48dw × 12dh
  Reason:   Word wrap applied
  ⚠️  Height 12 exceeds MaxHeight 10
```
