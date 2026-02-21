# Mint 布局系统分析报告

## 目录
- [1. 执行摘要](#1-执行摘要)
- [2. 系统架构概述](#2-系统架构概述)
- [3. 问题详细分析](#3-问题详细分析)
- [4. 约束系统深度分析](#4-约束系统深度分析)
- [5. 维度转换链分析](#5-维度转换链分析)
- [6. 实际案例研究](#6-实际案例研究)
- [7. 影响评估](#7-影响评估)

---

## 1. 执行摘要

### 1.1 核心问题
Mint 布局系统当前存在以下核心问题：

1. **约束传播不一致**：Measure 阶段和 Paint 阶段对约束的处理存在差异
2. **维度语义混淆**：组合架构中"外部维度"和"内部维度"频繁转换
3. **API 设计复杂**：Auto-measure 的语义不清晰，容易导致误用
4. **调试困难**：缺少约束传播的可视化工具

### 1.2 根本原因
- 组件组合架构（如 Panel = Border + VStack）导致层层维度转换
- 约束传递链中"父约束"与"自身显式维度"的优先级不明确
- Text.Wrap 组件同时受 MaxWidth 和 maxHeight 双重约束影响
- Measure 和 Paint 阶段共享但不协调的约束系统

### 1.3 影响
- **用户体验**：Panel 组件中 Text 内容溢出或高度计算错误
- **开发体验**：API 调用不符合直觉，需要通过测试验证行为
- **维护成本**：Bug 修复需要深入理解约束传播链

---

## 2. 系统架构概述

### 2.1 布局流水线

```
┌─────────────────────────────────────────────────────────────┐
│                    Fiber Render Pipeline                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. Reconcile: VNode Tree → Fiber Tree                       │
│                                                               │
│  2. Layout Phase                                             │
│     ├─ Measure: 计算每个组件的"理想尺寸"                      │
│     │   └─ 使用 Constraints 约束计算空间需求                 │
│     │                                                         │
│     └─ Layout: 分配实际位置和尺寸                             │
│         └─ 根据 Measure 结果分配可用空间                     │
│                                                               │
│  3. Paint Phase                                              │
│     └─ 使用 Bounds 渲染内容到 Buffer                         │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心组件层次

```
Panel (组合容器)
  │
  └─ Border (边框容器)
      │
      └─ VStack (垂直布局)
          │
          ├─ Header (可选)
          ├─ Flex(Content, 1)  ← Content 占据剩余空间
          └─ Footer (可选)
```

### 2.3 关键接口

```go
// 约束系统
type Constraints struct {
    MinWidth  int  // 最小宽度
    MaxWidth  int  // 最大宽度
    MinHeight int  // 最小高度
    MaxHeight int  // 最大高度
}

// 尺寸
type Size struct {
    Width  int  // 测量得到的大小
    Height int
}

// 渲染边界 (Paint 阶段)
type Rect [4]int  // [x, y, width, height]
```

---

## 3. 问题详细分析

### 3.1 约束传播不一致

#### 症状
- Panel 的 Text.Wrap 内容计算错误
- Auto-height Panel 返回错误的高度

#### 根因分析

**问题场景 1：Border 的约束传递错误**

```go
// ui/components/border/instance.go - Measure (修复前)
func (i *Instance) Measure(constraints Constraints) Size {
    ...
    if i.needMeasureHeight && i.innerWidth > 0 {
        childConstraints := Constraints{
            MaxWidth:  constraints.MaxWidth,  // 错误！应该用 innerWidth
            MaxHeight: constraints.MaxHeight,
        }
        childSize := i.child.Measure(childConstraints)
        ...
    }
}
```

**问题**：当 Border 有显式 `width=18`（内部宽度）但 height=0 时，Measure 子元素时传递了 `constraints.MaxWidth`（可能是 50），而不是 `innerWidth=18`。

**修复后**：
```go
// 修复后
if i.needMeasureHeight && i.innerWidth > 0 {
    childConstraints := Constraints{
        MaxWidth:  i.innerWidth,  // 使用显式的内部宽度
        MaxHeight: constraints.MaxHeight,
    }
    ...
}
```

#### 约束传递规则表

| 场景 | 子元素的 MaxWidth | 子元素的 MaxHeight |
|------|------------------|------------------|
| 父有显式 width，auto height | 父的 width | 父的 MaxHeight |
| 父有显式 height，auto width | 父的 MaxWidth | 父的 height |
| 父完全 auto | 父的 MaxWidth | 父的 MaxHeight |
| 父有固定 width/height | 父的 width | 父的 height |

### 3.2 维度语义混淆

#### 外部维度 vs 内部维度

**Panel 的维度定义**：
```go
Panel.Width(20)  // 外部总宽度 = 边框 + 内容
Panel.Height(10) // 外部总高度 = 边框 + 内容
```

**内部 Border 的维度**：
```go
// Panel.getComposed()
borderWidth := v.width - borderPadding  // 20 - 2 = 18
borderHeight := v.height - borderPadding // 10 - 2 = 8

border.SetWidth(borderWidth)   // Border 的内容宽度
border.SetHeight(borderHeight) // Border 的内容高度
```

**边框内边距计算**：
```go
// ui/components/border/utils.go
func GetBorderWidth(style BorderStyle) int {
    switch style {
    case BorderNone:
        return 0
    default:
        return 1  // 所有可见边框占用 1 个字符单元格
    }
}

// 总 padding = 左右各 1 * 2 = 2
borderPadding := 2 * GetBorderWidth(style)
```

#### 混淆点

| API 含义 | 实际含义 | 用户期望 |
|---------|---------|---------|
| `Panel.Width(20)` | Panel 的外部总宽度 | 内容宽度为 20 |
| `Border.SetWidth(18)` | Border 的内部内容宽度 | Border 总宽度为 18 |

**建议改进**：
```go
// 方案 1：明确命名
Panel.SetOuterWidth(20)   // 外部总宽度
Panel.SetInnerWidth(18)   // 内部内容宽度

// 方案 2：维度计算自动化
Panel.SetContentWidth(18) // 自动计算外部宽度 = 18 + border
```

### 3.3 Text.Wrap 双重约束

#### 约束 1：MaxWidth（决定换行）

```go
// ui/components/text/instance.go - Measure
func (i *Instance) Measure(constraints Constraints) Size {
    if i.wrap {
        maxWidth := constraints.MaxWidth
        lines := wordWrap(i.content, maxWidth)
        return Size{Width: maxWidth, Height: len(lines)}
    }
}
```

**示例**：
```
content = "This is a very long text that will be wrapped"
maxWidth = 18
→ 换行成 4 行:
  Line 1: "This is a very"
  Line 2: "long text that"
  Line 3: "will be wrapped"
  Line 4: "" (或最后一行)
→ 返回 Size{18, 4}
```

#### 约束 2：maxHeight（决定实际渲染行数）

```go
// ui/components/text/instance.go - Paint
func (i *Instance) Paint(ctx PaintContext, buf *Buffer) {
    if i.wrap {
        lines := wordWrap(i.content, ctx.Bounds[2]) // 使用 MaxWidth
        for i, line := range lines {
            if y+i >= ctx.Bounds[3] {  // 检查高度约束
                break  // 裁剪超出部分
            }
            render(line, x, y+i)
        }
    }
}
```

**不一致场景**：

```
Measure 阶段：
  MaxWidth = 18, MaxHeight = 100
  → Text 计算出 4 行
  → 返回 Size{18, 4}

Paint 阶段：
  Bounds = [0, 0, 18, 3]  // height 只给 3
  → Text 只渲染 3 行
  → 第 4 行被裁剪
```

#### 问题根源

1. **Measure 的 MaxHeight 和 Paint 的 Bounds[3] 可能不一致**
2. **Measure 用于布局规划，Paint 用于实际渲染**
3. **Flex 布局可能根据 Measure 结果分配空间，但实际空间更小**

### 3.4 Flex 布局的 Measure/Layout 耦合

#### FlexLayout 实现

```go
// ui/components/stack/instance.go - FlexLayout
func (i *Instance) FlexLayout(bounds Rect, children []ComponentInstance) {
    // 1. 测量所有子元素
    var totalWidth int
    for _, child := range children {
        size := child.Measure(constraints)  // 使用 Measure 结果
        totalWidth += size.Width
    }

    // 2. 计算可用空间
    availableSpace := bounds[2] - totalWidth  // ❌ 重新计算，与 Measure 阶段不一致
    flexSpace := availableSpace * flexFactor / totalFlex

    // 3. 分配空间给 flex 子元素
    for _, child := range flexChildren {
        child.Bounds = Rect{0, 0, flexSpace, bounds[3]}
        child.Layout(child.Bounds)
    }
}
```

#### 问题分析

1. **可用空间计算不一致**：
   - Measure 阶段：使用 `constraints.MaxWidth`
   - Layout 阶段：使用 `bounds[2]`（实际可用宽度）

2. **Flex 分配基数错误**：
   - 应该基于 `totalMeasuredWidth` 而非 `totalWidth`

3. **Auto-height 子元素的问题**：
   - Measure 返回 4 行（基于 MaxWidth=18）
   - Layout 只给 3 行高度
   - 第 4 行被裁剪，布局不合理

### 3.5 Auto-measure 语义不清晰

#### 当前 API

```go
Panel.New().
    SetWidth(20).   // 显式宽度
    // Height 未设置 → Auto height
    SetContent(text.Wrap("long content"))
```

#### 用户期望 vs 实际行为

| 用户期望 | 实际行为 | 原因 |
|---------|---------|------|
| SetWidth(20) 设置内容宽度 | 设置外部总宽度（含边框） | 组合架构导致 |
| Auto height 会包裹所有内容 | 但 Measure 阶段可能获取错误的 MaxWidth | 约束传播错误 |
| 显式维度优先于父约束 | 有时父约束覆盖显式维度 | 约束优先级不明确 |

#### 边界情况

**情况 1：显式维度 < 父约束**
```go
HStack(width=50)  // 父约束
  ├─ Panel(width=20)  // 显式 20 < 50
  └─ Panel(width=auto)  // 自动剩余

// 预期：Panel1 保持 20，Panel2 占据 30
```

**情况 2：显式维度 > 父约束**
```go
HStack(width=50)  // 父约束
  └─ Panel(width=60)  // 显式 60 > 50

// 预期：Panel 被压缩到 50，还是溢出？
```

**情况 3：显式维度 + auto**
```go
HStack(width=50)
  ├─ Panel(width=20)
  ├─ Panel(flex=1)
  └─ Panel(width=15)

// 预期：固定 20+15=35，flex 占据 15
```

---

## 4. 约束系统深度分析

### 4.1 约束传播链

```
Root Bounds (Terminal)
  ↓
Frame/Container Constraints
  ↓
┌─────────────────────────────────┐
│  HStack.Measure(constraints)   │
│    ↓                           │
│  Child 1 Constraints            │
│    ↓                           │
│  Child 2 Constraints            │
│    ↓                           │
│  Child 3 Constraints            │
└─────────────────────────────────┘
  ↓
Border.Measure(constraints)
  ↓
Modify constraints (width/height)
  ↓
VStack.Measure(newConstraints)
  ↓
FlexChild.Measure(childConstraints)
  ↓
Text.Measure(finalConstraints)
```

### 4.2 约束修改规则

#### 规则 1：显式维度修改约束

```go
// 原始约束
constraints = {MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 100}

// 组件有显式 width=20
newConstraints = {
    MinWidth: 20,          // 显式宽度作为最小和最大
    MaxWidth: 20,
    MinHeight: 0,
    MaxHeight: 100,        // height 继承父约束
}
```

#### 规则 2：组合容器的内部维度

```go
// Panel width=20 (外部)
borderPadding = 2  // 左右各 1

// Border Measure 时
innerWidth = 20 - 2 = 18

newConstraints = {
    MinWidth: 18,
    MaxWidth: 18,
    ...
}
```

#### 规则 3：Flex 元素的约束

```go
// HStack width=50, 包含两个子元素
totalFixedWidth = 20  // 固定元素总宽度
availableSpace = 50 - 20 = 30

// Flex 子元素的约束
flexConstraints = {
    MinWidth: 0,
    MaxWidth: availableSpace,  // 可用空间
    ...
}
```

### 4.3 约束冲突解决

#### 冲突 1：显式维度 vs 父约束

```
场景：
  Panel.SetWidth(20)  // 用户意图
  父约束 MaxWidth = 15  // 实际可用空间

解决策略：
  优先级：父约束 > 显式维度（保护布局不破坏）
  生成约束：MaxWidth = 15
  结果：Panel 被压缩
```

#### 冲突 2：MinWidth > MaxWidth

```
场景：
  MinWidth = 25
  MaxWidth = 20

解决策略：
  MaxWidth = max(MinWidth, MaxWidth) = 25
  警告：MinWidth 不能大于 MaxWidth
```

#### 冲突 3：Auto vs Fixed 混合

```
场景：
  HStack(width=50)
    ├─ Panel(width=30)   // 固定
    ├─ Panel(flex=1)     // Auto
    └─ Panel(width=25)   // 固定

问题：固定总和 = 55，超过 50

解决策略：
  1. 压缩固定元素（按比例）
  2. Flex 元素分配剩余空间（可能是负数）
  3. 或报错：布局不可行（推荐）
```

### 4.4 约束传递可视化工具设计

```go
// 约束追踪器（调试工具）
type ConstraintTracer struct {
    path     []string      // 组件路径
    steps    []TraceStep   // 约束传递步骤
}

type TraceStep struct {
    From      string       // 来源组件
    To        string       // 目标组件
    Input     Constraints  // 输入约束
    Output    Constraints  // 输出约束
    Dimension Size         // 测量结果
    Reason    string       // 约束修改原因
}

func (ct *ConstraintTracer) TraceMeasure(vnode VNode, constraints Constraints) {
    // 记录 Measure 阶段的约束传递
    ct.steps = append(ct.steps, TraceStep{
        From:   "Parent",
        To:     vnode.Tag(),
        Input:  constraints,
        Output: vnode.Props().Constraints,
        Reason: "Passing constraints",
    })
}
```

---

## 5. 维度转换链分析

### 5.1 Panel 组件的维度转换

```
用户 API:
  Panel.SetWidth(20).SetHeight(10)

维度转换链:
  Panel (外部维度)
    Width:  20
    Height: 10
      ↓
  Border (内部维度 = 外部 - 边框padding)
    padding = 2 * GetBorderWidth(Rounded) = 2
    Width:  20 - 2 = 18
    Height: 10 - 2 = 8
      ↓
  VStack (继承 Border 维度)
    Width:  18
    Height: 8 (或 auto)
      ↓
  Flex(Text, 1) (占据 VStack 的可用空间)
    MaxWidth: 18
    MaxHeight: 8
      ↓
  Text (根据约束换行)
    MaxWidth = 18
    如果内容长度 > 18，则换行
    返回高度 = ceil(len(content) / 18)
```

### 5.2 边框样式的维度影响

| 边框样式 | GetBorderWidth() | 左右 padding | 上下 padding | 总影响 |
|---------|-----------------|-------------|-------------|-------|
| BorderNone | 0 | 0 | 0 | 无影响 |
| BorderSingle | 1 | 2 | 2 | +2 width, +2 height |
| BorderDouble | 1 | 2 | 2 | +2 width, +2 height |
| BorderRounded | 1 | 2 | 2 | +2 width, +2 height |
| BorderDashed | 1 | 2 | 2 | +2 width, +2 height |

**注意**：Double 边框虽然视觉上更粗，但仍然占用 1 个字符单元格。

### 5.3 维度转换的代码路径

```go
// Panel.getComposed()
func (v *VNode) getComposed() VNode {
    // 1. 计算 Border 维度
    borderPadding := 2 * GetBorderWidth(v.borderStyle)

    // 2. 外部维度 → 内部维度
    innerWidth := v.width - borderPadding
    innerHeight := v.height - borderPadding

    // 3. 设置到 Border
    border.SetWidth(innerWidth)
    border.SetHeight(innerHeight)

    // 4. Border 内部处理（可能再转换）
    return border
}

// Border.Measure()
func (i *Instance) Measure() Size {
    // 1. 使用内部维度测量子元素
    childConstraints := Constraints{
        MaxWidth:  i.innerWidth,  // 18
        MaxHeight: i.innerHeight, // 8
    }
    childSize := i.child.Measure(childConstraints)

    // 2. 内部尺寸 → 外部尺寸（加回边框）
    outerWidth := childSize.Width + borderPadding
    outerHeight := childSize.Height + borderPadding

    return Size{outerWidth, outerHeight}
}
```

### 5.4 维度转换的问题

#### 问题 1：双重转换

```
Panel (20) → Border (18) → Measure → 返回 (18+2=20) → Panel
              ↓
           添加边框
              ↓
           返回外部尺寸
```

**问题**：为什么 Border 需要内部尺寸测量后又要加回边框？

**原因**：
- Border 的 `innerWidth` 可能是 0（auto），需要从子元素测量
- 测量后需要返回**外部尺寸**给父组件布局

#### 问题 2：Auto 维度的时机

```
场景：
  Panel(width=20, height=auto)  // height 未设置
    ↓
  Border(width=18, height=0)     // height=0 表示需要测量
    ↓
  VStack.Measure()               // 返回内容高度 4
    ↓
  Border.Size = {18, 4+2=6}      // 加回边框
    ↓
  Panel.Size = {20, 6}           // Panel 总高度
```

**问题**：如果在 Panel.Measure 阶段就已经计算了高度，为什么 Border.Measure 还要重复计算？

**原因**：
- Panel 本身没有 Instance，只有 VNode
- Panel.CreateInstance() 返回的是 Border 的 Instance
- 实际测量逻辑在 Border 中

---

## 6. 实际案例研究

### 6.1 案例 1：Panel 内容溢出

#### 用户代码

```go
panel.NewBuilder().
    Width(20).
    Height(3).
    Content(text.New("This is very long content that overflows").Wrap(true))
```

#### 期望行为

```
┌──────────────────┐
│ Title            │
├──────────────────┤
│ This is very     │
│ long content tha │
│ t overflows      │
└──────────────────┘
```

#### 实际行为（修复前）

```
┌──────────────────┐
│ Title            │
├──────────────────┤
│ This is very     │
│ long content tha │
│ t overflows      │
│ should be here   │  ← 溢出！
│ but is clipped   │
└──────────────────┘
```

#### 根因分析

```go
// Text.Paint() (修复前)
func (i *Instance) Paint(ctx PaintContext, buf *Buffer) {
    if i.wrap {
        lines := wordWrap(i.content, ctx.Bounds[2])  // 使用 MaxWidth=18
        for _, line := range lines {
            // ❌ 没有检查 ctx.Bounds[3] (maxHeight)
            render(line, x, y)
            y++
        }
    }
}
```

**问题**：
1. Text 根据 MaxWidth 计算出 6 行
2. 没有检查 maxHeight，渲染了所有 6 行
3. 超出 Panel 的 3 行高度

#### 修复方案

```go
// Text.Paint() (修复后)
func (i *Instance) Paint(ctx PaintContext, buf *Buffer) {
    if i.wrap {
        lines := wordWrap(i.content, ctx.Bounds[2])  // 使用 MaxWidth=18
        for i, line := range lines {
            if y+i >= ctx.Bounds[3] {  // ✅ 检查高度约束
                break  // 超出部分裁剪
            }
            render(line, x, y+i)
        }
    }
}
```

### 6.2 案例 2：Auto-height Panel 高度错误

#### 用户代码

```go
HStack(width=50).
    SetChildrenList([]VNode{
        panel.NewBuilder().Width(20).Height(3).Title("Fixed").Build(),
        panel.NewBuilder().Width(20).Title("Auto").
            Content(text.New("Short").Wrap(true)).Build(),
    })
```

#### 期望行为

```
┌──────────┐  ┌──────────┐
│ Fixed    │  │ Auto     │
├──────────┤  ├──────────┤
│ Content  │  │ Short    │
│ Content  │  │          │
└──────────┘  └──────────┘
  Height=3      Height=2  (内容只有 1 行 + 边框)
```

#### 实际行为（修复前）

```
┌──────────┐  ┌──────────┐
│ Fixed    │  │ Auto     │
├──────────┤  ├──────────┤
│ Content  │  │ Short    │
│ Content  │  │          │  ← 高度显示不足
└──────────┘  │          │
              └──────────┘
```

#### 根因分析

```go
// Border.Measure() (修复前)
func (i *Instance) Measure(constraints Constraints) Size {
    if i.needMeasureHeight && i.innerWidth > 0 {
        childConstraints := Constraints{
            MaxWidth:  constraints.MaxWidth,  // ❌ 错误！
            MaxHeight: constraints.MaxHeight,
        }
        childSize := i.child.Measure(childConstraints)  // 使用错误的约束
        ...
    }
}
```

**约束传播错误**：
```
HStack.Measure(constraints={MaxWidth: 50})
  ↓
Auto Panel (width=20, height=auto)
  ↓
Border (innerWidth=18, height=0, needMeasureHeight=true)
  ↓
childConstraints = {MaxWidth: 50, MaxHeight: 100}  // ❌ 应该是 {18, 100}
  ↓
VStack.Measure({50, 100})
  ↓
Text.Measure({50, 100})
  → Text 不需要换行（宽度 50 足够）
  → 返回 Size{50, 1}
  ↓
Border 返回 Size{50+2, 1+2} = {52, 3}
  ❌ 与 Panel 的 width=20 不符！
```

#### 修复方案

```go
// Border.Measure() (修复后)
func (i *Instance) Measure(constraints Constraints) Size {
    if i.needMeasureHeight && i.innerWidth > 0 {
        childConstraints := Constraints{
            MaxWidth:  i.innerWidth,  // ✅ 使用显式内部宽度
            MaxHeight: constraints.MaxHeight,
        }
        childSize := i.child.Measure(childConstraints)
        ...
    }
}
```

**修复后的约束传播**：
```
HStack.Measure(constraints={MaxWidth: 50})
  ↓
Auto Panel (width=20, height=auto)
  ↓
Border (innerWidth=18, height=0, needMeasureHeight=true)
  ↓
childConstraints = {MaxWidth: 18, MaxHeight: 100}  // ✅ 正确
  ↓
VStack.Measure({18, 100})
  ↓
Text.Measure({18, 100})
  → "Short" 长度 5 < 18，不需要换行
  → 返回 Size{5, 1}
  ↓
VStack 返回 Size{18, 1}
  ↓
Border 返回 Size{18+2, 1+2} = {20, 2}
  ✅ 与 Panel 的 width=20 一致，高度=2 正确
```

### 6.3 案例 3：Flex 布局的高度分配问题

#### 用户代码

```go
HStack(width=40, height=5).
    SetGap(2).
    SetChildrenList([]VNode{
        panel.NewBuilder().Width(15).Title("Fixed").Build(),
        stack.NewHStack().Flex(1, panel.NewBuilder().
            Title("Auto").
            Content(text.New("Text").Wrap(true)).
            Build()),
    })
```

#### 期望行为

```
┌───────────┐┌────────────────┐
│ Fixed     ││ Auto           │
├───────────┤├────────────────┤
│ Content   ││ Text           │
│ Content   ││                │
│ Content   │└────────────────┘
│ Content   │
└───────────┘
```

#### 实际行为（潜在问题）

```
问题分析：
1. HStack.Measure 返回 height=max(childHeights)=5
2. Layout 阶段分配给每个子元素 5 行
3. 但 Auto Panel 在 Measure 返回 height=2
4. 布局时 Flex 试图拉伸它到 5
5. 如果 Flex 的高度分配逻辑有问题，可能导致：
   - 内容过度拉伸
   - 或垂直对齐错误
```

#### 需要检查的代码

```go
// HStack.FlexLayout()
func (i *Instance) FlexLayout(bounds Rect, children []...) {
    for _, child := range children {
        childHeight := bounds[3]  // 所有子元素高度相同
        childBounds := Rect{x, y, childWidth, childHeight}

        // 检查是否支持 Stretch
        if child.SupportsStretch() {
            childBounds[3] = childHeight  // 拉伸到填满
        }

        child.Layout(childBounds)
    }
}
```

---

## 7. 影响评估

### 7.1 受影响的组件

| 组件 | 影响程度 | 具体问题 |
|------|---------|---------|
| Panel | 高 | 维度转换复杂，API 不符合直觉 |
| Border | 高 | 约束传递错误已修复，但仍需测试 |
| Text | 中 | Wrap 双重约束已处理 |
| VStack | 中 | Flex 布局可能存在高度分配问题 |
| HStack | 中 | Flex 布局需要进一步验证 |
| Wrap | 低 | 约束传播基本正常 |

### 7.2 用户影响

#### 初级用户
- **问题描述**：不理解 `Panel.SetWidth(20)` 的实际含义
- **建议**：提供更清晰的 API 和文档

#### 中级用户
- **问题描述**：Auto-measure 行为不符合直觉
- **建议**：提供调试工具和示例

#### 高级用户
- **问题描述**：复杂的组合布局难以调试
- **建议**：提供约束追踪和可视化工具

### 7.3 兼容性影响

#### 向后兼容性

**破格变更风险**：
1. **Panel 维度语义**：从"外部维度"改为"内部维度"
   - 破格变更：现有代码需要修改

2. **Border 约束传递行为**：已修复，但改变了部分场景的 Measure 结果
   - 非破格：bug 修复

**建议**：
- 保持现有 API 语义（外部维度）
- 提供新的显式 API（如 `SetContentWidth()`）
- 在文档中明确说明维度计算规则

### 7.4 性能影响

#### Measure 阶段
- **当前问题**：组合架构导致重复测量
- **潜在优化**：
  - 缓存中间结果
  - 减少不必要的维度转换

#### Paint 阶段
- **当前问题**：高度检查（`y+i >= ctx.Bounds[3]`）轻微影响性能
- **优化建议**：
  - 在 Measure 阶段就限制高度
  - Paint 阶段假设高度已正确计算

---

## 附录

### A. 相关代码文件

```
/ui/components/panel/
  vnode.go           - Panel 组合架构
  panel_test.go      - Panel 测试

/ui/components/border/
  instance.go        - Border 的 Measure/Paint
  vnode.go           - Border VNode API
  utils.go           - GetBorderWidth()

/ui/components/text/
  instance.go        - Text 的 Wrap 处理
  text_test.go       - Text 测试

/ui/components/stack/
  instance.go        - VStack/HStack 的 Flex 布局
  stack_test.go      - Stack 测试
```

### B. 相关提交

```
3c64b3c4 fix: Text content overflow in Panel by respecting layout bounds
6a776c0f fix: Auto-height Panel now correctly measures content height
27951e8a test: Fix Panel component tests to correctly reflect composition architecture
```

### C. 相关文档

```
/docs/layout/
  plan/             - 本文档目录
    01-analysis.md  - 分析报告（本文档）
    02-optimization.md  - 优化方案
    03-testing.md   - 测试方案
    04-debug-tools.md - 调试工具
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-21
**作者**: Qwen Code
