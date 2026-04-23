# Mint 布局系统优化实施方案

## 目录
- [1. 优化目标](#1-优化目标)
- [2. 设计原则](#2-设计原则)
- [3. 短期优化（Phase 1）](#3-短期优化phase-1)
- [4. 中期优化（Phase 2）](#4-中期优化phase-2)
- [5. 长期优化（Phase 3）](#5-长期优化phase-3)
- [6. API 改进方案](#6-api-改进方案)
- [7. 实施时间表](#7-实施时间表)
- [8. 风险评估](#8-风险评估)

---

## 1. 优化目标

### 1.1 核心目标
1. **降低复杂度**：简化约束传播和维度转换逻辑
2. **提高一致性**：统一 Measure 和 Paint 阶段的约束处理
3. **改善用户体验**：使 API 更符合直觉
4. **增强可调试性**：提供约束追踪和可视化工具

### 1.2 量化指标

| 指标 | 当前状态 | 目标状态 | 提升 |
|------|---------|---------|------|
| 约束传播错误数 | 2+ | 0 | -100% |
| API 调用符合直觉度 | 60% | 90% | +50% |
| 调试时间（平均） | 2小时 | 30分钟 | -75% |
| Measure 阶段重复调用 | 高频 | 低频 | -50% |
| 代码复杂度（圈复杂度） | >15 | <10 | -33% |

---

## 2. 设计原则

### 2.1 约束优先级原则

**显式维度 > 父约束 > 自动测量**

```go
// 决策树
func ResolveDimension(explicitValue int, parentConstraint int, autoMeasure func() int) int {
    if explicitValue > 0 {
        // 显式维度优先
        return explicitValue
    }
    if parentConstraint > 0 {
        // 使用父约束
        return parentConstraint
    }
    // 自动测量
    return autoMeasure()
}
```

### 2.2 维度语义透明原则

**API 应明确说明是"外部维度"还是"内部维度"**

```go
// 方案 1：清晰命名
panel.SetOuterWidth(20)   // 明确是外部维度
panel.SetInnerWidth(18)   // 明确是内部维度

// 方案 2：自动计算
panel.SetContentWidth(18)  // 明确是内容宽度，自动计算外部宽度
```

### 2.3 单一职责原则

**每个组件只负责一层维度转换**

错误示例：
```go
// Panel.getComposed() 同时处理维度转换和边框计算
func (v *VNode) getComposed() VNode {
    innerWidth := v.width - borderPadding  // 维度转换
    border.SetWidth(innerWidth)            // 边框设置
    // ...
}
```

正确示例：
```go
// Panel.getComposed() 只负责组合
func (v *VNode) getComposed() VNode {
    border := v.createBorder()  // 创建 Border
    border.SetDimensions(v.calculateBorderDimensions())  // 委托维度计算
    return border
}

func (v *VNode) calculateBorderDimensions() BorderDimensions {
    return BorderDimensions{
        OuterWidth:  v.width,
        OuterHeight: v.height,
        Style:       v.borderStyle,
    }
}

// Border 负责自己的维度转换
func (v *VNode) SetDimensions(d BorderDimensions) {
    v.innerWidth = d.OuterWidth - 2 * GetBorderWidth(d.Style)
    v.innerHeight = d.OuterHeight - 2 * GetBorderWidth(d.Style)
}
```

### 2.4 可观测性原则

**所有约束修改都应被记录**

```go
type MeasureContext struct {
    ParentConstraints Constraints
    ChildConstraints  Constraints
    ComponentTag      string
    ModificationTrace []string  // 约束修改原因
}

func (i *Instance) Measure(ctx MeasureContext) Size {
    // 记录约束修改
    if ctx.ChildConstraints != ctx.ParentConstraints {
        ctx.ModificationTrace = append(ctx.ModificationTrace,
            "Using explicit width: " + strconv.Itoa(i.width))
    }

    // 执行测量
    size := i.measureInternal(ctx)

    // 返回结果
    return size
}
```

---

## 3. 短期优化（Phase 1）

### 3.1 目标
修复当前已知的 bug，完善约束传播逻辑

### 3.2 任务清单

#### 任务 1.1：统一约束传递规则

**文件**: `ui/components/border/instance.go`

**当前代码**：
```go
func (i *Instance) Measure(constraints Constraints) Size {
    if i.needMeasureHeight {
        childConstraints := Constraints{
            MaxWidth:  constraints.MaxWidth,  // 可能错误
            MaxHeight: constraints.MaxHeight,
        }
        ...
    }
}
```

**优化后**：
```go
func (i *Instance) Measure(constraints Constraints) Size {
    childConstraints := i.computeChildConstraints(constraints)
    childSize := i.child.Measure(childConstraints)
    ...
}

func (i *Instance) computeChildConstraints(constraints Constraints) Constraints {
    cc := Constraints{
        MinWidth:  constraints.MinWidth,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: constraints.MinHeight,
        MaxHeight: constraints.MaxHeight,
    }

    // 规则: 显式维度 > 父约束
    if i.innerWidth > 0 {
        cc.MinWidth = i.innerWidth
        cc.MaxWidth = i.innerWidth
    }

    if i.innerHeight > 0 {
        cc.MinHeight = i.innerHeight
        cc.MaxHeight = i.innerHeight
    }

    // 规则: 边框内边距
    cc.MinWidth = max(0, cc.MinWidth - i.borderPadding)
    cc.MaxWidth = max(0, cc.MaxWidth - i.borderPadding)
    cc.MinHeight = max(0, cc.MinHeight - i.borderPadding)
    cc.MaxHeight = max(0, cc.MaxHeight - i.borderPadding)

    return cc
}
```

#### 任务 1.2：完善 Text.Wrap 的高度约束

**文件**: `ui/components/text/instance.go`

**已修复**，但需要增加验证逻辑：

```go
func (i *Instance) ValidatePaintSize(measureSize Size, paintBounds Rect) error {
    measureHeight := measureSize.Height
    paintHeight := paintBounds[3]

    if measureHeight > paintHeight {
        // 内容被裁剪，检查这是否是预期的
        if !i.allowCrop {
            return fmt.Errorf("Text height (%d) exceeds paint bounds (%d), content will be cropped",
                measureHeight, paintHeight)
        }
    }

    return nil
}
```

#### 任务 1.3：添加约束追踪日志

**新文件**: `ui/layout/constraints/tracer.go`

```go
package constraints

import (
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/ui"
)

// Tracer 追踪约束传播
type Tracer struct {
    enabled bool
    entries []Entry
}

type Entry struct {
    From      string
    To        string
    Input     layout.Constraints
    Output    layout.Constraints
    Dimension layout.Size
    Reason    string
}

var globalTracer = &Tracer{}

func Enable() {
    globalTracer.enabled = true
}

func Disable() {
    globalTracer.enabled = false
}

// Trace Measure 阶段的约束传递
func TraceMeasure(from, to string, input, output layout.Constraints, resultSize layout.Size, reason string) {
    if !globalTracer.enabled {
        return
    }

    globalTracer.entries = append(globalTracer.entries, Entry{
        From:      from,
        To:        to,
        Input:     input,
        Output:    output,
        Dimension: resultSize,
        Reason:    reason,
    })
}

// Dump 输出约束传播历史
func Dump() string {
    if !globalTracer.enabled {
        return "Constraint tracer is disabled"
    }

    var buf strings.Builder
    buf.WriteString("Constraint Propagation Trace:\n")
    buf.WriteString("==========================================\n\n")

    for i, entry := range globalTracer.entries {
        buf.WriteString(fmt.Sprintf("Step %d: %s → %s\n", i+1, entry.From, entry.To))
        buf.WriteString(fmt.Sprintf("  Input:    %s\n", entry.Input))
        buf.WriteString(fmt.Sprintf("  Output:   %s\n", entry.Output))
        buf.WriteString(fmt.Sprintf("  Dimension: %s\n", entry.Dimension))
        if entry.Reason != "" {
            buf.WriteString(fmt.Sprintf("  Reason:   %s\n", entry.Reason))
        }
        buf.WriteString("\n")
    }

    return buf.String()
}

func Clear() {
    globalTracer.entries = nil
}
```

**集成到组件**：
```go
// ui/components/border/instance.go
import "github.com/wwsheng009/mint/ui/layout/constraints"

func (i *Instance) Measure(constraints Constraints) Size {
    childConstraints := i.computeChildConstraints(constraints)
    childSize := i.child.Measure(childConstraints)

    // 追踪约束传递
    constraints.TraceMeasure(
        "border", i.child.Tag(),
        constraints, childConstraints,
        childSize,
        "Applied child constraints",
    )

    return i.computeOuterSize(childSize)
}
```

#### 任务 1.4：添加边界检查测试

**文件**: `ui/components/border/border_test.go`

```go
func TestBorder_ConstraintPropagation(t *testing.T) {
    tests := []struct {
        name           string
        borderWidth    int
        parentConstraints layout.Constraints
        expectedChildConstraints layout.Constraints
    }{
        {
            name:   "Explicit width uses inner width",
            borderWidth: 20,
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            expectedChildConstraints: layout.Constraints{
                MinWidth:  18,  // 20 - borderPadding(2)
                MaxWidth:  18,
                MinHeight: 0,
                MaxHeight: 100,
            },
        },
        {
            name:   "Auto width uses parent constraint",
            borderWidth: 0,  // auto
            parentConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  50,
                MinHeight: 0,
                MaxHeight: 100,
            },
            expectedChildConstraints: layout.Constraints{
                MinWidth:  0,
                MaxWidth:  48,  // 50 - borderPadding(2)
                MinHeight: 0,
                MaxHeight: 98,  // 100 - borderPadding(2)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            border := NewInstance()
            border.SetWidth(tt.borderWidth)

            childConstraints := border.computeChildConstraints(tt.parentConstraints)

            if childConstraints != tt.expectedChildConstraints {
                t.Errorf("Child constraints mismatch:\n"+
                    "  Expected: %+v\n"+
                    "  Got:      %+v",
                    tt.expectedChildConstraints, childConstraints)
            }
        })
    }
}
```

### 3.3 验收标准
- [ ] 所有现有测试通过
- [ ] 新增的边界测试通过
- [ ] 约束追踪工具可以正确输出
- [ ] Panel 内容不再溢出
- [ ] Auto-height Panel 高度计算正确

---

## 4. 中期优化（Phase 2）

### 4.1 目标
改进 API 设计，降低使用复杂度

### 4.2 任务清单

#### 任务 2.1：Panel API 改进

**新文件**: `ui/components/panel/api_improvement.go`

```go
package panel

import "github.com/wwsheng009/mint/ui/components/text"

// VNode 增强方法

// ============================================
// 方案: 添加明确的维度命名方法
// ============================================

// SetOuterWidth 设置外部总宽度（包含边框）
// 这是 SetWidth 的别名，保持向后兼容
func (v *VNode) SetOuterWidth(w int) *VNode {
    return v.SetWidth(w)
}

// SetInnerWidth 设置内部内容宽度（不含边框）
// 自动计算外部宽度 = innerWidth + borderPadding
func (v *VNode) SetInnerWidth(w int) *VNode {
    borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
    v.width = w + borderPadding
    v.composed = nil
    return v
}

// SetContentWidth 设置内容宽度 (SetInnerWidth 的别名)
func (v *VNode) SetContentWidth(w int) *VNode {
    return v.SetInnerWidth(w)
}

// SetOuterHeight 设置外部总高度（包含边框）
// 这是 SetHeight 的别名，保持向后兼容
func (v *VNode) SetOuterHeight(h int) *VNode {
    return v.SetHeight(h)
}

// SetInnerHeight 设置内部内容高度（不含边框）
// 自动计算外部高度 = innerHeight + borderPadding
func (v *VNode) SetInnerHeight(h int) *VNode {
    borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
    v.height = h + borderPadding
    v.composed = nil
    return v
}

// SetContentSize 设置内容文本行数
// 自动计算高度 = lineCount + borderPadding
func (v *VNode) SetContentSize(lineCount int) *VNode {
    return v.SetInnerHeight(lineCount)
}

// ============================================
// 方案: 提供便捷的内容设置方法
// ============================================

// SetTextContent 简化文本内容设置
// 自动使用 Text 组件并设置 Wrap=true
func (v *VNode) SetTextContent(content string) *VNode {
    textNode := text.New(content).SetWrap(true)
    v.content = textNode
    v.composed = nil
    return v
}

// SetWrappedTextContent 设置带有固定宽度的 Wrap 文本
// 自动设置 Panel 宽度来适应文本
func (v *VNode) SetWrappedTextContent(content string, width int) *VNode {
    v.SetContentWidth(width)
    textNode := text.New(content).SetWrap(true)
    v.content = textNode
    v.composed = nil
    return v
}

// ============================================
// 方案: 维度计算辅助方法
// ============================================

// GetOuterDimensions 获取外部总尺寸
func (v *VNode) GetOuterDimensions() (int, int) {
    return v.width, v.height
}

// GetInnerDimensions 获取内部内容尺寸
func (v *VNode) GetInnerDimensions() (int, int) {
    borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
    innerWidth := max(0, v.width - borderPadding)
    innerHeight := max(0, v.height - borderPadding)
    return innerWidth, innerHeight
}

// GetPadding 获取边框内边距
func (v *VNode) GetPadding() (int, int) {
    padding := 2 * newborder.GetBorderWidth(v.borderStyle)
    return padding, padding  // width, height padding
}
```

**使用示例**：
```go
// 方式 1: 使用外部维度（现有方式）
panel.NewBuilder().
    Width(20).Height(6).
    Content(text.New("Hello").Wrap(true))

// 方式 2: 使用内部维度（新方式）
panel.NewBuilder().
    SetContentWidth(18).SetContentHeight(4).
    Content(text.New("Hello").Wrap(true))

// 方式 3: 使用便捷方法（推荐）
panel.NewBuilder().
    SetWrappedTextContent("Hello World", 18)

// 方式 4: 自动尺寸
panel.NewBuilder().
    SetTextContent("Auto-sized content")
```

#### 任务 2.2：Builder API 增强

**新文件**: `ui/components/panel/builder_enhanced.go`

```go
package panel

import rtui "github.com/wwsheng009/mint/runtime/ui"
import newtext "github.com/wwsheng009/mint/ui/components/text"

// Builder 增强方法

// WithInnerWidth 设置内部内容宽度
func (b *Builder) WithInnerWidth(w int) *Builder {
    borderPadding := 2 * newborder.GetBorderWidth(b.borderStyle)
    b.width = w + borderPadding
    return b
}

// WithContentWidth 设置内容宽度 (alias)
func (b *Builder) WithContentWidth(w int) *Builder {
    return b.WithInnerWidth(w)
}

// WithInnerHeight 设置内部内容高度
func (b *Builder) WithInnerHeight(h int) *Builder {
    borderPadding := 2 * newborder.GetBorderWidth(b.borderStyle)
    b.height = h + borderPadding
    return b
}

// WithWrappedText 设置 Wrap 文本并自动调整宽度
func (b *Builder) WithWrappedText(text string, width int) *Builder {
    b.width = width + 2*newborder.GetBorderWidth(b.borderStyle)
    b.content = newtext.New(text).SetWrap(true)
    return b
}

// WithAutoHeight 设置自动高度（内容自适应）
func (b *Builder) WithAutoHeight() *Builder {
    b.height = 0
    return b
}

// WithFixed 固定宽高（外部维度）
func (b *Builder) WithFixed(w, h int) *Builder {
    b.width = w
    b.height = h
    return b
}

// WithFixedInner 固定宽高（内部维度）
func (b *Builder) WithFixedInner(w, h int) *Builder {
    b.WithInnerWidth(w)
    b.WithInnerHeight(h)
    return b
}

// WithContent 仅设置内容，自动尺寸
func (b *Builder) WithContent(content rtui.VNode) *Builder {
    b.content = content
    // 保持 width=0, height=0 以启用 auto-measure
    return b
}

// WithTextContent 设置文本内容并自动尺寸
func (b *Builder) WithTextContent(text string) *Builder {
    b.content = newtext.New(text)
    return b
}
```

#### 任务 2.3：文档和示例更新

**新文件**: `docs/layout/panel_api_guide.md`

```markdown
# Panel API 指南

## 维度概念

Panel 有两种维度概念：

1. **外部维度 (Outer Dimensions)**: 包含边框的总尺寸
2. **内部维度 (Inner Dimensions)**: 内容区域的尺寸（不含边框）

```
┌─────────────────────┐
│ Title               │  ← 外部高度包含边框
├─────────────────────┤
│                     │
│   内部内容区域       │  ← 内部高度不含边框
│                     │
└─────────────────────┘
↑  左边框 ↑ 右边框
      边框内边距
```

## API 方法对照

| 场景 | 传统 API | 新 API | 说明 |
|------|---------|--------|------|
| 设置外部宽度 | `Width(20)` | `SetOuterWidth(20)` | 包含边框 |
| 设置内容宽度 | - | `SetContentWidth(18)` | 不含边框 |
| 设置外部高度 | `Height(6)` | `SetOuterHeight(6)` | 包含边框 |
| 设置内容行数 | - | `SetContentSize(4)` | 不含边框 |
| Wrap 文本 | `Width(20).Content(text.Wrap(...))` | `SetWrappedTextContent(text, 18)` | 自动尺寸 |

## 推荐使用方式

### 场景 1：固定尺寸
```go
panel.NewBuilder().
    WithFixed(20, 6).
    Content(newtext.New("Hello"))
```

### 场景 2：固定宽度，自动高度
```go
panel.NewBuilder().
    WithContentWidth(18).
    WithAutoHeight().
    Content(newtext.New("Long content...").Wrap(true))
```

### 场景 3：自动尺寸
```go
panel.NewBuilder().
    WithAutoHeight().
    WithTextContent("Auto-sized content")
```

### 场景 4：Wrap 文本
```go
panel.NewBuilder().
    WithWrappedText("This is a long text that should wrap", 18)
```
```

### 4.3 验收标准
- [ ] 新 API 通过所有测试
- [ ] 文档完整且清晰
- [ ] 示例代码可运行
- [ ] 向后兼容性检查通过
- [ ] 用户反馈 API 更符合直觉

---

## 5. 长期优化（Phase 3）

### 5.1 目标
引入新的布局引擎，从根本上解决复杂度问题

### 5.2 任务清单

#### 任务 3.1：引入布局 DSL

**概念**：使用声明式 DSL 定义布局

```go
// 新文件: ui/layout/dsl/layout.go

package dsl

import "github.com/wwsheng009/mint/runtime/ui"

type Node struct {
    tag       string
    props     Props
    children  []Node
}

func Panel(props Props, children ...Node) Node {
    return Node{tag: "panel", props: props, children: children}
}

func Column(props Props, children ...Node) Node {
    return Node{tag: "column", props: props, children: children}
}

func Row(props Props, children ...Node) Node {
    return Node{tag: "row", props: props, children: children}
}

func Text(content string) Node {
    return Node{tag: "text", props: Props{"content": content}}
}

type Props map[string]interface{}

func (p Props) Width(w int) Props {
    p["width"] = w
    return p
}

func (p Props) Height(h int) Props {
    p["height"] = h
    return p
}

func (p Props) Flex(f int) Props {
    p["flex"] = f
    return p
}

// 布局 DSL 示例
func ExampleLayout() Node {
    return Column{}.Flex(1).
        Children(
            Panel{}.Title("Header").
                Height(3).
                Content(Text("Header")),
            Row{}.Flex(1).
                Children(
                    Panel{}.Title("Sidebar").
                        Width(20).
                        Content(Text("Sidebar")),
                    Panel{}.Title("Main").
                        Flex(1).
                        Content(Text("Main content")),
                ),
        )
}
```

#### 任务 3.2：布局可视化工具

**概念**：将布局树和约束可视化

```go
// 新文件: ui/layout/visualizer/tree.go

package visualizer

type Visualizer struct {
    nodes map[string]NodeState
}

type NodeState struct {
    Tag         string
    Bounds      Rect
    Constraints Constraints
    Dimension   Size
    Parent      string
    Children    []string
}

func (v *Visualizer) PrintTree() string {
    var buf strings.Builder
    v.printRecursive(buf, "root", "")
    return buf.String()
}

func (v *Visualizer) printRecursive(buf strings.Builder, nodeID, indent string) {
    node := v.nodes[nodeID]

    buf.WriteString(fmt.Sprintf("%s┌─ %s\n", indent, node.Tag))
    buf.WriteString(fmt.Sprintf("%s│  Bounds:    %s\n", indent, node.Bounds))
    buf.WriteString(fmt.Sprintf("%s│  Constraints: %s\n", indent, node.Constraints))
    buf.WriteString(fmt.Sprintf("%s│  Dimension:  %s\n", indent, node.Dimension))

    for i, childID := range node.Children {
        isLast := i == len(node.Children)-1
        childIndent := indent
        if isLast {
            childIndent += "└─ "
        } else {
            childIndent += "├─ "
        }
        v.printRecursive(buf, childID, childIndent)
    }
}

func (v *Visualizer) PrintConstraintChain(nodeID string) string {
    var chain []string

    for nodeID != "" {
        node := v.nodes[nodeID]
        chain = append(chain, fmt.Sprintf("%s: %s", node.Tag, node.Constraints))
        nodeID = node.Parent
    }

    // 反转
    for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
        chain[i], chain[j] = chain[j], chain[i]
    }

    return strings.Join(chain, "\n  ↓\n")
}
```

#### 任务 3.3：性能优化

**优化 1：Measure 缓存**

```go
type MeasureCache struct {
    cache map[string]Size
}

func (mc *MeasureCache) Get(vnode VNode, constraints Constraints) (Size, bool) {
    key := mc.key(vnode, constraints)
    size, ok := mc.cache[key]
    return size, ok
}

func (mc *MeasureCache) Put(vnode VNode, constraints Constraints, size Size) {
    key := mc.key(vnode, constraints)
    mc.cache[key] = size
}

func (mc *MeasureCache) Invalidate(vnode VNode) {
    // 移除该节点的所有缓存
    for key := range mc.cache {
        if strings.Contains(key, vnode.Key()) {
            delete(mc.cache, key)
        }
    }
}
```

**优化 2：增量布局计算**

```go
type IncrementalLayout struct {
    dirtyNodes map[string]bool
}

func (il *IncrementalLayout) MarkDirty(nodeID string) {
    il.dirtyNodes[nodeID] = true
}

func (il *IncrementalLayout) NeedsLayout(nodeID string) bool {
    return il.dirtyNodes[nodeID]
}

func (il *IncrementalLayout) Layout(root VNode) {
    // 只布局标记为 dirty 的节点及其子树
    for nodeID := range il.dirtyNodes {
        il.layoutSubtree(nodeID)
        delete(il.dirtyNodes, nodeID)
    }
}
```

### 5.3 验收标准
- [ ] DSL 可以定义复杂布局
- [ ] 可视化工具可以生成清晰的树形图
- [ ] Measure 缓存命中率 > 70%
- [ ] 增量布局减少 50% 的重复计算
- [ ] 整体性能提升 > 30%

---

## 6. API 改进方案

### 6.1 当前 API 问题

```go
// 问题 1: SetWidth(20) 的语义不明确
panel.SetWidth(20)
// 用户期望: 内容宽度为 20
// 实际: 外部总宽度为 20（内容宽度 18）

// 问题 2: Auto-measure 的触发方式不清晰
panel.SetWidth(20)  // height 未设置 -> auto 吗？
// 是的，但文档中没有明确说明

// 问题 3: Wrap 文本需要多次设置
panel.SetWidth(20).
    SetContent(text.New("Content").SetWrap(true))
// 能否简化？
```

### 6.2 改进方案

#### 方案 A: 明确命名（推荐）

```go
// 保持向后兼容：SetWidth/SetHeight = 外部维度
panel.SetWidth(20)   // 外部宽度（推荐使用 SetOuterWidth）
panel.SetHeight(6)   // 外部高度（推荐使用 SetOuterHeight）

// 新 API：明确的内部维度
panel.SetInnerWidth(18)    // 内部内容宽度
panel.SetInnerHeight(4)    // 内部内容高度

// 别名（更直观）
panel.SetContentWidth(18)  // 内容宽度
panel.SetContentSize(4)    // 内容行数
```

**优点**：
- 明确区分外部/内部维度
- 向后兼容
- 用户可以自由选择使用

**缺点**：
- API 数量增加
- 需要文档说明区别

#### 方案 B: 智能推断（不推荐）

```go
// 自动根据设置推断是外部还是内部
panel.SetWidth(20)  // 智能推断是外部还是内部？❌ 太复杂
```

**不推荐原因**：
- 推断逻辑复杂，容易出错
- 用户体验不确定

#### 方案 C: Builder 链式调用（推荐）

```go
panel.NewBuilder().
    WithFixed(20, 6).
    Content(content)

panel.NewBuilder().
    WithContentWidth(18).
    WithAutoHeight().
    Content(content)

panel.NewBuilder().
    WithWrappedText(text, 18)  // 最简洁
```

**优点**：
- 链式调用简洁流畅
- 方法名明确表达意图
- 易于扩展

**缺点**：
- 需要创建 Builder 实例

### 6.3 推荐的 API 设计

```go
// 1. 基础 API（VNode 方法）
panel.SetWidth(20)           // 外部维度
panel.SetOuterWidth(20)      // 明确的外部维度
panel.SetContentWidth(18)    // 内部维度
panel.SetInnerWidth(18)      // 内部维度（别名）

panel.SetHeight(6)           // 外部维度
panel.SetOuterHeight(6)      // 明确的外部维度
panel.SetContentSize(4)      // 内部维度（行数）
panel.SetInnerHeight(4)      // 内部维度（别名）

// 2. 便捷 API（VNode 方法）
panel.SetTextContent("Auto-sized text")
panel.SetWrappedTextContent("Wrap text", 18)

// 3. Builder API（推荐）
panel.NewBuilder().
    WithFixed(20, 6).                       // 固定尺寸
    Content(content)

panel.NewBuilder().
    WithContentWidth(18).                 // 固定宽度
    WithAutoHeight().                      // 自动高度
    Content(content)

panel.NewBuilder().
    WithWrappedText(text, 18).            // Wrap 文本（最简洁）

panel.NewBuilder().
    Title("Panel").
    WithAutoHeight().                      // 完全自动
    WithTextContent(text)
```

---

## 7. 实施时间表

### 7.1 阶段划分

| 阶段 | 时间 | 任务 | 优先级 |
|------|------|------|--------|
| Phase 1 | Week 1 | 约束传播规则统一 | P0 |
| Phase 1 | Week 1 | Text.Wrap 高度约束验证 | P0 |
| Phase 1 | Week 2 | 约束追踪工具 | P1 |
| Phase 1 | Week 2 | 边界检查测试 | P1 |
| Phase 2 | Week 3-4 | Panel API 增强 | P0 |
| Phase 2 | Week 3-4 | Builder API 增强 | P0 |
| Phase 2 | Week 4 | 文档和示例更新 | P1 |
| Phase 3 | Week 5-6 | 布局 DSL | P2 |
| Phase 3 | Week 5-6 | 可视化工具 | P1 |
| Phase 3 | Week 6 | 性能优化 | P2 |

### 7.2 里程碑

| 里程碑 | 日期 | 目标 |
|--------|------|------|
| M1 | Week 2 结束 | 所有已知 bug 修复，测试通过 |
| M2 | Week 4 结束 | API 改进完成，文档更新 |
| M3 | Week 6 结束 | 可视化工具可用，性能提升 |

### 7.3 资源分配

- **开发**: 1 人全职
- **测试**: 1 人兼职（Week 1-2 全职，Week 3-6 兼职）
- **文档**: 兼职（Week 3-4）

---

## 8. 风险评估

### 8.1 技术风险

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| 约束优化引入新 bug | 中 | 高 | 完善测试，逐步推进 |
| API 破格变更影响用户 | 中 | 高 | 保持向后兼容，提供迁移指南 |
| 性能优化效果不明显 | 低 | 中 | 基线性能测试，持续监控 |
| DSL 学习成本高 | 低 | 低 | 提供充分文档和示例 |

### 8.2 进度风险

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| 任务延期 | 中 | 中 | 缓冲时间，分阶段交付 |
| 需求变更 | 低 | 中 | 锁定 Phase 1 需求，Phase 2-3 可调整 |
| 资源不足 | 低 | 高 | 提前规划，优先保障核心任务 |

### 8.3 质量风险

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| 测试覆盖不足 | 中 | 高 | 制定测试计划，代码审查 |
| 文档不完整 | 中 | 中 | 技术写作支持，用户反馈 |
| 约束追踪工具不准确 | 低 | 中 | 多场景测试，与手动追踪对比 |

---

## 附录

### A. 相关文件

```
/ui/layout/
  constraints/         - 约束系统
    tracer.go         - 约束追踪工具
  dsl/                - 布局 DSL
    layout.go         - DSL 定义
  visualizer/         - 可视化工具
    tree.go           - 树形图生成

/ui/components/panel/
  api_improvement.go  - API 增强
  builder_enhanced.go - Builder 增强

/docs/layout/panel_api_guide.md  - Panel API 指南
```

### B. 测试模板

```go
// 约束传播测试模板
func TestComponent_ConstraintPropagation(t *testing.T) {
    tests := []struct {
        name                    string
        component               Component
        parentConstraints       Constraints
        expectedChildConstraints Constraints
        expectedSize            Size
    }{
        // 测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given
            ctx := MeasureContext{
                Constraints: tt.parentConstraints,
                Tracer:      NewTracer(),
            }

            // When
            size := tt.component.Measure(ctx)

            // Then
            if size != tt.expectedSize {
                t.Errorf("Size mismatch: got %v, want %v", size, tt.expectedSize)
            }

            // 验证约束传递
            ctx.Tracer.AssertChildConstraints(tt.expectedChildConstraints)
        })
    }
}
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-21
**作者**: Qwen Code
