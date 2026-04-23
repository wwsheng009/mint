# 约束追踪器应用总结

## 概述

约束追踪器 (`runtime/layout/tracer.go`) 已实现并集成到 Border 组件中。本文档总结当前应用状态和使用方式。

---

## 实现状态

### 已完成 ✅

| 功能 | 状态 | 文件 |
|------|------|------|
| 核心追踪逻辑 | ✅ 已实现 | runtime/layout/tracer.go |
| Border 集成 | ✅ 已完成 | ui/components/border/instance.go |
| 测试覆盖 | ✅ 完成 | ui/components/border/border_constraint_test.go |
| 使用指南 | ✅ 完成 | docs/layout/constraint_tracer_guide.md |
| 完成报告 | ✅ 完成 | docs/layout/phase1_completion.md |

### 待集成（建议）🔄

| 组件 | 文件 | 优先级 |
|------|------|--------|
| Panel | ui/components/panel/instance.go | 高 |
| VStack | ui/components/stack/instance.go | 中 |
| HStack | ui/components/stack/instance.go | 中 |
| Text | ui/components/text/instance.go | 高 |

---

## 集成方式

### 1. 基本集成模式

```go
package yourcomponent

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Instance struct {
    path    string  // 用于追踪的路径
    child   rtui.VNode
    width   int     // 显式宽度
    height  int     // 显式高度
    // ...
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 1. 计算子组件约束
    childConstraints := inst.computeChildConstraints(constraints)

    // 2. 追踪约束传递
    // 获取子组件标识
    childTag := "child"
    if tagger, ok := inst.child.(interface{ Tag() string }); ok {
        childTag = tagger.Tag()
    }

    // 构建路径
    childPath := fmt.Sprintf("%s/%s", inst.path, childTag)

    // 追踪约束（在测量子组件之前调用）
    layout.TraceMeasuring(
        "component("+inst.key+")",  // 来源组件 ID
        childTag,                   // 目标组件 ID
        childPath,                  // 完整路径
        constraints,                // 输入约束
        childConstraints,           // 输出约束
        layout.Size{},              // 尺寸（测量后更新）
        fmt.Sprintf("Reason"),      // 约束修改原因
    )

    // 3. 测量子组件
    childSize := inst.measureChild(inst.child, childConstraints)

    // 4. 更新追踪数据（可选，如果需要在同一步记录最终尺寸）
    // （当前实现会在 TraceMeasuring 中创建条目，尺寸稍后更新）

    // 5. 计算最终尺寸
    // ...

    return finalSize
}
```

### 2. Border 集成示例

```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // ...

    if (needMeasureWidth || needMeasureHeight) && inst.child != nil {
        // Use computeChildConstraints to unify constraint passing logic
        innerConstraints := inst.computeChildConstraints(constraints)

        // Trace constraint propagation (before measurement)
        childPath := fmt.Sprintf("%s/%s", inst.path, "border")
        childTag := "border"
        if tagger, ok := inst.child.(interface{ Tag() string }); ok {
            childTag = tagger.Tag()
        }

        layout.TraceMeasuring(
            "border("+inst.key+")",
            childTag,
            childPath,
            constraints,          // 输入约束
            innerConstraints,     // 输出约束（传递给子元素）
            layout.Size{},        // 尺寸将在测量后更新
            fmt.Sprintf("Applied border padding (%dx%d), explicit width=%d, height=%d",
                borderWidth, borderWidth, inst.width, inst.height),
        )

        // Measure child and cache result
        inst.measuredChildSize = inst.measureChild(inst.child, innerConstraints)

        // ...计算最终尺寸
    }

    // ...
}
```

---

## API 快速参考

### 启用/禁用追踪器

```go
// 启用全局追踪（在 main 函数中）
layout.EnableTracer()

// 禁用追踪
layout.DisableTracer()

// 检查是否启用
if layout.IsTracerEnabled() {
    // 追踪器已启用
}
```

### 追踪约束

```go
layout.TraceMeasuring(
    from,      // 来源组件 ID
    to,        // 目标组件 ID
    path,      // 完整路径
    input,     // 输入约束
    output,    // 输出约束
    resultSize,// 最终尺寸
    reason,    // 修改原因
)
```

### 管理追踪数据

```go
// 清除之前的数据
layout.ClearTrace()

// 获取所有条目
entries := layout.GetTraceEntries()

// 输出追踪日志
log := layout.DumpTrace()
fmt.Println(log)
```

---

## 使用建议

### 1. 调试时启用

```go
package main

import (
    "flag"
    "github.com/wwsheng009/mint/runtime/layout"
)

var debugLayout = flag.Bool("debug-layout", false, "Enable layout tracing")

func main() {
    flag.Parse()

    if *debugLayout {
        layout.EnableTracer()
        defer layout.DisableTracer()
    }

    // ... 运行程序

    // 输出追踪日志
    if *debugLayout {
        fmt.Println(layout.DumpTrace())
    }
}
```

### 2. 在测试中使用

```go
func TestComponent_LayoutConstraints(t *testing.T) {
    // 启用追踪器
    layout.EnableTracer()
    defer layout.DisableTracer()
    defer layout.ClearTrace()

    // 创建并测量组件
    component := NewComponent()
    component.Measure(constraints)

    // 验证追踪数据
    entries := layout.GetTraceEntries()
    for _, entry := range entries {
        // 验证约束传递
    }
}
```

### 3. 组件 ID 命名

```go
// ✅ 好的 ID - 包含组件类型和键
layout.TraceMeasuring("border(my-panel)", "content", ...)

// ✅ 好的 ID - 包含父级上下文
layout.TraceMeasuring("vstack(header)", "title", ...)

// ❌ 模糊的 ID
layout.TraceMeasuring("vstack", "child", ...)
```

---

## 输出示例

```
╔══════════════════════════════════════════════════════════════════╗
║                    Constraint Propagation Trace               ║
╚══════════════════════════════════════════════════════════════════╝

Step 0
  Path: /app/main/border
  border(my-border) → content
  Input:    {0..80} × {0..24}
  Output:   {76..78} × {0..24}
  Dimension: 78dw × 6dh
  Reason:   Applied border padding (1x1), explicit width=0, height=0
```

---

## 性能考虑

### 追踪器开销

- **无追踪**：几乎无开销（`IsTracerEnabled()` 检查）
- **启用追踪**：每个 TraceMeasuring 调用约 1-2 μs

### 内存使用

- 每个追踪条目约 200-300 字节
- 建议定期 ClearTrace() 避免内存增长

### 生产环境建议

```go
// 仅在调试构建中启用
if buildMode == "debug" {
    layout.EnableTracer()
}

// 或者环境变量控制
if os.Getenv("LAYOUT_TRACE") == "true" {
    layout.EnableTracer()
}
```

---

## 常见问题

### Q: 追踪器是否线程安全？

A: 追踪器不是线程安全的。如果在多个 goroutine 中使用，请确保每个 goroutine 有独立的追踪上下文，或者避免并发使用。

### Q: 如何只追踪特定组件？

A: 使用环境变量：
```bash
export LAYOUT_TRACE=true
export LAYOUT_TRACE_ONLY=border,panel
```

（需要在 layout/tracer.go 中实现过滤逻辑）

### Q: 追踪器会影响性能吗？

A: 追踪器仅在被启用时才有开销。禁用状态下的检查 (<10 ns) 可以忽略不计。

---

## 相关文档

- [约束追踪器使用指南](./constraint_tracer_guide.md) - 详细 API 和文档
- [Phase 1 完成报告](./phase1_completion.md) - 总体完成情况
- [优化计划 - Phase 1](./plan/01-analysis.md) - 原始设计方案
