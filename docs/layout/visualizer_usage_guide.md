# 布局可视化工具使用指南

## 概述

布局可视化工具 (`ui/layout/visualizer`) 提供了一套强大的工具来检查和调试 Mint 组件的布局。它可以可视化布局树结构、追踪约束传播、检测潜在的布局问题。

**适用场景**：
- 调试布局问题
- 理解约束传播
- 检测尺寸异常
- 性能分析

**三种可视化方式**：

| 可视化方式 | 方法 | 描述 | 适用场景 |
|-----------|------|------|---------|
| **树形图** | `PrintTree()` | 传统缩进式树形结构 | 查看层级关系和约束传递 |
| **盒模型** | `PrintBoxModel()` | Chrome DevTools 风格的嵌套方块 | 直观查看嵌套布局和尺寸 |
| **网格图** | `PrintGrid()` | 2D 字符网格显示布局空间 | 理解组件在空间中的位置 |

---

## 基础用法

### 1. 创建可视化器

```go
import "github.com/wwsheng009/mint/ui/layout/visualizer"

// 创建新的可视化器
vis := visualizer.NewVisualizer()
```

### 2. 手动添加节点

```go
vis.AddNode(
    "panel_1",                           // 节点 ID
    "panel",                             // 节点类型（tag）
    layout.Rect{X: 0, Y: 0, Width: 40, Height: 15},  // 边界
    layout.Constraints{                  // 输入约束
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 24,
    },
    layout.Constraints{                  // 输出约束（传给子节点）
        MinWidth:  0,
        MaxWidth:  78,
        MinHeight: 0,
        MaxHeight: 22,
    },
    layout.Size{Width: 40, Height: 15}, // 测量尺寸
    "",                                  // 父节点 ID（空字符串表示根节点）
)
```

### 3. 打印布局树

```go
// 打印整个布局树
fmt.Println(vis.PrintTree())
```

**输出示例**：
```
Layout Tree:
════════════

┌─ panel (panel_1)
│  Position: (0, 0)
│  Size: 40w x 15dh
│  Input: {0..80} x {0..24}
│  To Children: {0..78} x {0..22}
│  Props: title=Settings

└─ border (border_1)
   │  Position: (0, 0)
   │  Size: 38w x 20dh
   │  Input: {0..78} x {0..22}
   │  To Children: {0..76} x {0..20}
   │  ⚠️  Height 20 exceeds MaxHeight 20
```

### 4. Chrome DevTools 风格盒模型

```go
// 打印盒模型可视化
fmt.Println(vis.PrintBoxModel())
```

**输出示例**：
```
Layout Box Model (Chrome DevTools style)
==================================================

┌────────────────────────────────────────┐
│                                        │
│ │ panel (40w x 15h)
│ │ Pos: (0,0)  {0..80} x {0..24}
│                                        │
│                                        │
││ ┌──────────────────────────────────────┐
││ │                                      │
││ │ │ bordered (38w x 13h) 【BORDER】
││ │ │ Pos: (1,1)  {0..78} x {0..22}
││ │                                      │
││ │                                      │
││ ││ ┌────────────────────────────────────┐
││ ││ │                                    │
││ ││ │ │ text (36w x 5h) 【TEXT】
││ ││ │ │ Pos: (2,2)  {0..76} x {0..20}
││ ││ │                                    │
││ ││ └────────────────────────────────────┘
││ │                                      │
││ └──────────────────────────────────────┘
│                                        │
└────────────────────────────────────────┘

==================================================
【BORDER】 = Border component with padding
【TEXT】   = Text content element
```

**特点**：
- 嵌套的方块显示组件层级
- 直接查看内部/外部尺寸关系
- 检测约束超限时显示警告 ⚠️
- 方框内显示位置、尺寸和约束

### 5. 2D 网格可视化

```go
// 打印网格可视化
fmt.Println(vis.PrintGrid())
```

**输出示例**：
```
Layout Grid (2D Visualization)
==================================================

      0    5   10   15   20   25   30   35
  ────────────────────────────────────────
 0│████████████████████████████████████████│
 1│████████████████████████████████████████│
 2│████████████████████████████████████████│
 3│█████████████████••••••••••••••••••••••│
 4│█████████████████••••••••••••••••••••••│
 5│█████████████████••••••••••••••••••••••│
  ────────────────────────────────────────

Legend:
  █ = panel/border     ║ = vstack    ░ = unknown
  ▓ = bordered         ═ = hstack    · = text
  ▒ = button
```

**图例说明**：
- `█` - panel/border 组件
- `▓` - bordered 组件
- `•` - text 内容
- `║` - vstack 垂直堆叠
- `═` - hstack 水平堆叠
- `▒` - button 按钮

**特点**：
- 2D 网格显示布局的空间分布
- 直观理解组件在画布上的位置
- 带坐标轴，便于定位

### 6. 打印约束传播链

```go
// 打印从根到指定节点的约束传播链
fmt.Println(vis.PrintConstraintChain("border_1"))
```

**输出示例**：
```
panel (panel_1)
  Input: {0..80} x {0..24}
  Output: {0..78} x {0..22}
  ↓

border (border_1)
  Input: {0..78} x {0..22}
  Output: {0..76} x {0..20}
  ↓
```

### 5. 打印布局摘要

```go
// 打印布局摘要统计
fmt.Println(vis.PrintSummary())
```

**输出示例**：
```
Layout Summary:
══════════════

Total Nodes: 5
Max Depth: 3
Root Size: 40w × 15dh
Root Position: (0, 0)
Root Constraints: {0..80} x {0..24}

Node Types:
  panel: 2
  border: 1
  text: 2
```

### 7. 打印布局摘要

```go
// 打印布局摘要统计
fmt.Println(vis.PrintSummary())
```

**输出示例**：
```
Layout Summary:
══════════════

Total Nodes: 5
Max Depth: 3
Root Size: 40w × 15dh
Root Position: (0, 0)
Root Constraints: {0..80} x {0..24}

Node Types:
  panel: 2
  border: 1
  text: 2
```

### 8. 查找布局问题

```go
// 自动检测布局问题
problems := vis.FindProblems()
for _, problem := range problems {
    fmt.Printf("⚠️  %s\n", problem)
}
```

**输出示例**：
```
⚠️  Node border_1 (border): height 20 exceeds MaxHeight 20
⚠️  Node text_1 (text): width 85 exceeds MaxWidth 80
```

---

---

## 实际应用场景

### 场景 0：切换不同可视化方式

根据不同的调试需求，选择最合适的可视化方法：

```go
// 创建可视化
vis := visualizer.VisualizeVNode(vnode, constraints)

// 场景 1: 快速查看层级结构 → 使用树形图
fmt.Println(vis.PrintTree())

// 场景 2: 理解嵌套布局和内部/外部尺寸 → 使用盒模型
fmt.Println(vis.PrintBoxModel())

// 场景 3: 查看组件在空间中的实际位置 → 使用网格图
fmt.Println(vis.PrintGrid())

// 场景 4: 统计布局信息 → 使用摘要
fmt.Println(vis.PrintSummary())
```

### 场景 1：调试 Panel 布局

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/components/panel"
    "github.com/wwsheng009/mint/ui/components/text"
    "github.com/wwsheng009/mint/ui/layout/visualizer"
)

func debugPanelLayout() {
    // 创建 Panel
    p := panel.NewBuilder().
        Title("Settings").
        OuterSize(40, 15).
        Content(text.New("Settings go here")).
        Build()

    // 从 VNode 自动构建可视化
    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 24,
    }

    vis := visualizer.VisualizeVNode(p, constraints)

    // 打印布局结构
    fmt.Println(vis.PrintTree())

    // 检查问题
    problems := vis.FindProblems()
    if len(problems) > 0 {
        fmt.Println("\n发现布局问题：")
        for _, problem := range problems {
            fmt.Printf("  - %s\n", problem)
        }
    }
}
```

### 场景 2：在组件中集成可视化

```go
package mycomponent

import (
    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/ui/layout/visualizer"
)

var debugVis *visualizer.Visualizer

// InitVisualization 初始化可视化（可选）
func InitVisualization() {
    debugVis = visualizer.NewVisualizer()
}

// DebugMeasuring 在测量时可视化
func DebugMeasuring(
    id string,
    tag string,
    constraints layout.Constraints,
    size layout.Size,
    parentID string,
) {
    if debugVis == nil {
        return
    }

    debugVis.AddNode(
        id,
        tag,
        layout.Rect{X: 0, Y: 0, Width: size.Width, Height: size.Height},
        constraints,
        layout.Constraints{}, // 输出约束需要在测量后更新
        size,
        parentID,
    )
}

// 在组件 Measure 方法中使用
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 测量逻辑...

    // 调试可视化
    if os.Getenv("DEBUG_LAYOUT") == "true" {
        DebugMeasuring(
            inst.Key(),
            inst.Tag(),
            constraints,
            size,
            "",
        )
    }

    return size
}
```

### 场景 3：对比不同约束下的布局

```go
func compareConstraints(vnode rtui.VNode) {
    constraints1 := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  40,
        MinHeight: 0,
        MaxHeight: 10,
    }

    constraints2 := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 20,
    }

    vis1 := visualizer.VisualizeVNode(vnode, constraints1)
    vis2 := visualizer.VisualizeVNode(vnode, constraints2)

    fmt.Println("=== 约束 1 ===")
    fmt.Println(vis1.PrintSummary())

    fmt.Println("\n=== 约束 2 ===")
    fmt.Println(vis2.PrintSummary())
}
```

### 场景 4：追踪约束传播问题

```go
func traceConstraintPropagation(vnode rtui.VNode, nodeName string) {
    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  50,
        MinHeight: 0,
        MaxHeight: 12,
    }

    vis := visualizer.VisualizeVNode(vnode, constraints)

    // 打印特定节点的约束传播链
    fmt.Printf("=== 约束传播到 %s ===\n", nodeName)
    fmt.Println(vis.PrintConstraintChain(nodeName))
}
```

### 场景 5：自动检测布局问题

```go
func autoDetectLayoutIssues(vnode rtui.VNode) layout.Constraints {
    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  60,
        MinHeight: 0,
        MaxHeight: 15,
    }

    vis := visualizer.VisualizeVNode(vnode, constraints)

    // 查找所有问题
    problems := vis.FindProblems()
    if len(problems) > 0 {
        fmt.Printf("发现 %d 个布局问题：\n", len(problems))
        for i, problem := range problems {
            fmt.Printf("%d. %s\n", i+1, problem)
        }

        // 建议修复
        fmt.Println("\n建议修复：")
        fmt.Println("- 增加父组件的 MaxWidth/MaxHeight")
        fmt.Println("- 检查子组件的尺寸设置")
        fmt.Println("- 考虑使用 Wrap 或 Auto 尺寸")
    }

    return constraints
}
```

---

## 集成到约束追踪器

布局可视化工具可以与约束追踪器 (`runtime/layout/tracer`) 配合使用。

```go
package main

import (
    "fmt"
    "os"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/ui/layout/visualizer"
)

func visualizeWithTracer(vnode rtui.VNode) {
    // 启用约束追踪器
    layout.EnableTracer()
    defer layout.DisableTracer()

    // 测量 VNode
    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  80,
        MinHeight: 0,
        MaxHeight: 24,
    }

    size := vnode.Measure(constraints)

    // 创建可视化器
    vis := visualizer.VisualizeVNode(vnode, constraints)

    // 结合追踪数据
    if os.Getenv("DEBUG_LAYOUT") == "true" {
        fmt.Println("=== 约束追踪 ===")
        fmt.Println(layout.DumpTrace())

        fmt.Println("\n=== 布局树 ===")
        fmt.Println(vis.PrintTree())

        fmt.Println("\n=== 布局问题 ===")
        problems := vis.FindProblems()
        if len(problems) == 0 {
            fmt.Println("✅ 未发现问题")
        } else {
            for _, problem := range problems {
                fmt.Printf("⚠️  %s\n", problem)
            }
        }
    }
}
```

---

## 高级用法

### 1. 自定义节点属性

```go
func addCustomProperties(vis *visualizer.Visualizer) {
    // 添加自定义属性到节点
    vis.SetNodeProperty("panel_1", "name", "MainPanel")
    vis.SetNodeProperty("panel_1", "priority", "high")
    vis.SetNodeProperty("panel_1", "visible", true)

    // 这些属性会在 PrintTree 中显示
    fmt.Println(vis.PrintTree())
    // 输出：│  Props: name=MainPanel priority=high visible=true
}
```

### 2. 嵌套子树可视化

```go
func visualizeSubtree(vis *visualizer.Visualizer, parentID string, subtree rtui.VNode) {
    constraints := layout.Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

    // 递归构建子树
    buildSubtree(vis, subtree, constraints, parentID, 0, 0)
}

func buildSubtree(vis *visualizer.Visualizer, vnode rtui.VNode, constraints layout.Constraints, parentID string, x, y int) string {
    if vnode == nil {
        return ""
    }

    nodeID := vnode.Key()
    if nodeID == "" {
        nodeID = fmt.Sprintf("node_%p", vnode)
    }

    tag := vnode.Tag()
    size := layout.Size{Width: 0, Height: 0}

    // 测量节点
    if measurable, ok := vnode.(interface{ Measure(layout.Constraints) layout.Size }); ok {
        size = measurable.Measure(constraints)
    }

    // 添加节点
    vis.AddNode(
        nodeID,
        tag,
        layout.Rect{X: x, Y: y, Width: size.Width, Height: size.Height},
        constraints,
        layout.Constraints{},
        size,
        parentID,
    )

    // 递归处理子节点
    children := vnode.Children()
    for _, child := range children {
        buildSubtree(vis, child, constraints, nodeID, x, y)
    }

    return nodeID
}
```

### 3. 性能分析

```go
func analyzePerformance(vnode rtui.VNode) {
    vis := visualizer.VisualizeVNode(vnode, layout.Constraints{})

    // 统计节点类型
    nodeTypes := make(map[string]int)
    for _, node := range vis.nodes {
        nodeTypes[node.Tag]++
    }

    fmt.Println("=== 性能分析 ===")
    fmt.Printf("总节点数: %d\n", len(vis.nodes))
    fmt.Printf("最大深度: %d\n", vis.calculateDepth(vis.rootID, 0))
    fmt.Println("\n节点类型分布：")
    for tag, count := range nodeTypes {
        fmt.Printf("  %s: %d\n", tag, count)
    }

    // 检测过深的树
    maxDepth := vis.calculateDepth(vis.rootID, 0)
    if maxDepth > 10 {
        fmt.Printf("⚠️  警告: 布局树过深（深度: %d），可能影响性能\n", maxDepth)
    }
}
```

---

## 最佳实践

### 1. 调试时启用，生产环境禁用

```go
// 通过环境变量控制
if os.Getenv("DEBUG_LAYOUT") == "true" {
    vis := visualizer.VisualizeVNode(vnode, constraints)
    fmt.Println(vis.PrintTree())
}
```

### 2. 结合约束追踪器使用

```go
// 添加布局追踪
layout.EnableTracer()
vis := visualizer.VisualizeVNode(vnode, constraints)

// 先看约束踪迹
fmt.Println(layout.DumpTrace())

// 再看布局树
fmt.Println(vis.PrintTree())
```

### 3. 自动检测问题

```go
// 在 CI/CD 中使用
vis := visualizer.VisualizeVNode(vnode, constraints)
problems := vis.FindProblems()
if len(problems) > 0 {
    log.Fatalf("发现布局问题: %v", problems)
}
```

### 4. 使用摘要快速了解结构

```go
// 用 PrintSummary 快速查看概览
vis := visualizer.VisualizeVNode(vnode, constraints)
fmt.Println(vis.PrintSummary())
```

---

## 与其他工具配合

### 与约束追踪器

```go
layout.EnableTracer()
vis := visualizer.VisualizeVNode(vnode, constraints)

// 约束追踪器显示传播过程
fmt.Println(layout.DumpTrace())

// 可视化器显示最终结果
fmt.Println(vis.PrintTree())
```

### 与增量布局

```go
ctx := incremental.NewLayoutContext()

// 标记需要重新布局的节点
ctx.MarkDirty(updatedNode)

// 检查需要重新布局的节点
if ctx.NeedsLayout(node) {
    // 测量并可视化
    vis.AddNode(...)
}
```

---

## 总结

布局可视化工具提供了三种主要的使用方式：

1. **手动构建** - 完全控制每个节点的信息
2. **自动构建** - 直接从 VNode 树生成可视化
3. **问题检测** - 自动发现布局异常

结合约束追踪器和其他调试工具，可以大大简化布局问题的调试过程。

---

**文档版本**: 1.0
**最后更新**: 2026-02-22
**维护者**: Qwen Code
