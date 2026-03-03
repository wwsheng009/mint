# Runtime Layout 边框属性处理流程

**版本**: 1.0
**日期**: 2026-02-24
**方案**: 方案 A - 边框作为容器属性

---

## 目录

1. [概述](#概述)
2. [边框类型系统](#边框类型系统)
3. [处理流程](#处理流程)
4. [核心算法](#核心算法)
5. [各布局类型的边框处理](#各布局类型的边框处理)
6. [接口适配](#接口适配)
7. [示例代码](#示例代码)

---

## 概述

在方案 A 中，边框成为容器的**原生属性**，而不是独立的包装组件。这意味着：

- ✅ 所有容器（Stack, Grid, Wrap, Absolute, Modal）原生支持边框
- ✅ 边框尺寸自动包含在容器的总尺寸中
- ✅ 子节点布局自动考虑边框偏移
- ✅ 无需使用 `Border` 组件包装

---

## 边框类型系统

### 1. BorderStyle 枚举

位置: `runtime/layout/border.go`

```go
type BorderStyle int

const (
    BorderNone    BorderStyle = iota  // 无边框
    BorderSingle                       // 单线边框 ┌─┐│└┘
    BorderDouble                       // 双线边框 ╔═╗║╚╝
    BorderRounded                      // 圆角边框 ╭─╮│╰╯
    BorderDashed                       // 虚线边框 ┌┐│└┘
)
```

**关键特性**：
- 所有边框字符都占用 **1 个字符单元格**
- 边框宽度始终为 **1**（无论视觉上是单线还是双线）

---

### 2. Border 结构体

```go
type Border struct {
    Style BorderStyle  // 边框样式
    Width int          // 边框宽度（始终为 1）
    Label string       // 边框标签（显示在顶部中间）
}
```

---

### 3. 边框空间计算方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `HasBorder()` | bool | 是否有可见边框 |
| `HorizontalPadding()` | int | 水平边框空间（左+右=2） |
| `VerticalPadding()` | int | 垂直边框空间（上+下=2） |
| `ContentOffset()` | (x, y) int | 内容区偏移（始终 1, 1） |

**空间分配示意图**：

```
┌─────────────────────────────────┐  ← y = 0 (Top border, 1 row)
│     Border Label                │
├─────────────────────────────────┤  ← y = 1 (Content start, offset = 1)
│                                 │
│        Content Area             │  ← Content region
│                                 │
├─────────────────────────────────┤
└─────────────────────────────────┘  ← y = height-1 (Bottom border, 1 row)

  x=0        x=1                        x=width-1
  ↑          ↑                           ↑
  Left       Content start               Right border
  border
```

---

## 处理流程

### 整体流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                      VNode 设置边框属性                          │
│    vstack.Border("single", "Title")                             │
└─────────────────────┬───────────────────────────────────────────┘
                      │ Props
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Fiber 创建并同步属性                          │
│  1. Fiber.BorderStyle ← Props["borderStyle"]                   │
│  2. Fiber.BorderLabel ← Props["label"]                         │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│               FiberToNodeAdapter 适配                          │
│  func GetBorder() layout.Border { ... }                       │
│  优先读取 Fiber.BorderStyle，兼容旧 Props 方式                 │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                  布局引擎测量阶段                              │
│  1. 获取边框: border = node.GetBorder()                      │
│  2. 计算内容约束: inner = constraints - border.padding(...)   │
│  3. 测量子节点到内容区域                                      │
│  4. 返回总尺寸: inner_size + border.padding                │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                  布局引擎布局阶段                              │
│  1. 获取盒模型: boxModel = node.GetBoxModel()                │
│  2. 计算内容偏移: boxModel.ContentOffsetX/Y()                 │
│  3. 子节点位置 = 父位置 + 内容位置 + 边框偏移 + padding       │
│  4. 保存盒模型到 LayoutBox：box.BoxModel = boxModel          │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Paint 阶段                                 │
│  读取 LayoutBox.BoxModel.Border 并渲染边框字符                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 核心算法

### 算法 1: GetBorderFromNode

**位置**: `runtime/layout/border.go`

**功能**: 安全获取节点的边框配置

**伪代码**:

```go
func GetBorderFromNode(node Node) Border {
    if node == nil {
        return Border{Style: BorderNone}
    }

    if bordered, ok := node.(Bordered); ok {
        return bordered.GetBorder()
    }

    return Border{Style: BorderNone}
}
```

---

### 算法 2: CalculateBorderConstraints

**位置**: `runtime/layout/border.go`

**功能**: 根据边框调整约束，用于计算子节点可用空间

**伪代码**:

```go
func CalculateBorderConstraints(constraints Constraints, border Border) Constraints {
    if !border.HasBorder() {
        return constraints  // 无边框，不调整
    }

    return Constraints{
        MinWidth:  max(0, constraints.MinWidth - border.HorizontalPadding()),  // 左右各减1
        MaxWidth:  max(0, constraints.MaxWidth - border.HorizontalPadding()),
        MinHeight: max(0, constraints.MinHeight - border.VerticalPadding()),    // 上下各减1
        MaxHeight: max(0, constraints.MaxHeight - border.VerticalPadding()),
    }
}
```

**示例**:
```go
// 容器约束：100x50
// 边框占用：2x2（左右各1，上下各1）
// 内容约束：98x48

constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50}
border := NewBorder(BorderSingle)
inner := CalculateBorderConstraints(constraints, border)
// inner.MinWidth = 98, inner.MaxWidth = 98
// inner.MinHeight = 48, inner.MaxHeight = 48
```

---

### 算法 3: CalculateBorderBoxSize

**位置**: `runtime/layout/border.go`

**功能**: 根据内容尺寸计算包含边框的总尺寸

**伪代码**:

```go
func CalculateBorderBoxSize(contentWidth, contentHeight int, border Border) (int, int) {
    if !border.HasBorder() {
        return contentWidth, contentHeight  // 无边框，不增加
    }
    return contentWidth + border.HorizontalPadding(),  // +2
           contentHeight + border.VerticalPadding()    // +2
}
```

**示例**:
```go
// 内容尺寸：50x30
// 边框占用：2x2
// 总尺寸：52x32

contentWidth, contentHeight := 50, 30
border := NewBorder(BorderSingle)
totalWidth, totalHeight := CalculateBorderBoxSize(contentWidth, contentHeight, border)
// totalWidth = 52, totalHeight = 32
```

---

### 算法 4: FiberToNodeAdapter.GetBorder()

**位置**: `internal/render/fiber_adapter.go`

**功能**: 从 Fiber 获取边框配置（方案 A 优先策略）

**伪代码**:

```go
func (a *FiberToNodeAdapter) GetBorder() layout.Border {
    // 优先级 1: 方案 A - 使用 Fiber.BorderStyle 字段（对所有容器有效）
    if a.fiber.BorderStyle != "" && a.fiber.BorderStyle != "none" {
        return layout.Border{
            Style: parseBorderStyleString(a.fiber.BorderStyle),
            Label: a.fiber.BorderLabel,
        }
    }

    // 优先级 2: 向后兼容 - 使用旧 Props 方式
    if a.fiber.Tag == "bordered" || a.fiber.Tag == "border" {
        style := propsToBorderStyle(a.fiber.Props)
        label := propsToLabel(a.fiber.Props)
        return layout.Border{Style: style, Label: label}
    }

    // Modal 特殊处理
    if a.fiber.Tag == "modal" {
        style := propsToBorderStyle(a.fiber.Props)
        if style == BorderNone {
            style = BorderDouble  // Modal 默认双线
        }
        label := propsToTitle(a.fiber.Props)
        return layout.Border{Style: style, Label: label}
    }

    return layout.Border{Style: BorderNone}
}
```

**策略说明**:
1. **方案 A 优先**: 首先读取 `Fiber.BorderStyle` 和 `Fiber.BorderLabel`
2. **向后兼容**: 对旧 `bordered` 组件读取 `Props["borderStyle"]`
3. **Modal 特例**: 默认使用双线边框，标题作为标签

---

### 算法 5: FiberToNodeAdapter.Measure()

**位置**: `internal/render/fiber_adapter.go`

**功能**: 测量节点尺寸，包含边框计算

**伪代码**:

```go
func (a *FiberToNodeAdapter) Measure(constraints Constraints) Size {
    // 1. 获取边框配置
    border := a.GetBorder()

    // 2. 特殊处理：absolute 容器填充父容器
    if a.fiber.Tag == "absolute" {
        return Size{
            Width:  constraints.MaxWidth,
            Height: constraints.MaxHeight,
        }
    }

    // 3. 优先从 Instance 获取尺寸（已迁移组件）
    if a.fiber.Instance != nil {
        contentSize := measureInstance(a.fiber.Instance, constraints)
        if border.HasBorder() {
            // 加上边框空间
            return Size{
                Width:  contentSize.Width + border.HorizontalPadding(),
                Height: contentSize.Height + border.VerticalPadding(),
            }
        }
        return contentSize
    }

    // 4. 从 Style/Props 获取固定尺寸（用户设置的总尺寸）
    if hasFixedSize(a.fiber) {
        return Size{
            Width:  a.fiber.Style.Width,   // 已包含边框
            Height: a.fiber.Style.Height,
        }
    }

    // 5. 测量子节点（自动尺寸）
    if len(a.children) > 0 {
        // 5.1 计算内容约束（减去边框）
        innerConstraints := constraints
        if border.HasBorder() {
            innerConstraints = Constraints{
                MinWidth:  max(0, constraints.MinWidth - border.HorizontalPadding()),
                MaxWidth:  max(0, constraints.MaxWidth - border.HorizontalPadding()),
                MinHeight: max(0, constraints.MinHeight - border.VerticalPadding()),
                MaxHeight: max(0, constraints.MaxHeight - border.VerticalPadding()),
            }
        }

        // 5.2 测量子节点
        contentSize := measureChildren(a.children, innerConstraints)

        // 5.3 加上边框
        if border.HasBorder() {
            return Size{
                Width:  contentSize.Width + border.HorizontalPadding(),
                Height: contentSize.Height + border.VerticalPadding(),
            }
        }
        return contentSize
    }

    // 6. 默认值
    return Size{Width: 0, Height: 0}
}
```

---

### 算法 6: Engine.layoutNodeWithDepth()

**位置**: `runtime/layout/types.go`

**功能**: 递归布局节点，处理边框偏移

**核心逻辑**:

```go
func (e *Engine) layoutNodeWithDepth(node Node, constraints Constraints, x, y int, depth int, visited map[string]bool) *LayoutBox {
    // 1. 获取边框配置
    nodeBorder := GetBorderFromNode(node)

    // 2. 计算边框内容偏移
    borderOffsetX, borderOffsetY := nodeBorder.ContentOffset()  // (1, 1) if has border

    // 3. 创建 LayoutBox 并保存边框
    box := &LayoutBox{
        ID:       node.ID(),
        X:        x,
        Y:        y,
        Width:    width,
        Height:   height,
        Border:   nodeBorder,  // ✨ 保存边框供渲染使用
        Children: make([]*LayoutBox, 0),
    }

    // 4. 根据容器类型布局子节点
    if flexProvider, ok := node.(FlexStyleProvider); ok {
        // 4.1 计算内容空间（减去边框）
        innerWidth := width - nodeBorder.HorizontalPadding()
        innerHeight := height - nodeBorder.VerticalPadding()

        // 4.2 使用 FlexLayout 布局子节点
        childBoxes := flexLayout.LayoutChildren(innerWidth, innerHeight)

        // 4.3 递归布局子节点（位置加上边框偏移）
        for i, childBox := range childBoxes {
            child := node.Children()[i]
            childX := x + childBox.X + borderOffsetX  // ✨ 加上边框偏移
            childY := y + childBox.Y + borderOffsetY
            subBox := e.layoutNodeWithDepth(child, constraints, childX, childY, depth+1, visited)
            subBox.X = childX
            subBox.Y = childY
            box.Children = append(box.Children, subBox)
        }
    } else if gridProvider, ok := node.(GridStyleProvider); ok {
        // Grid 布局逻辑（类似 Flex）
        innerWidth := width - nodeBorder.HorizontalPadding()
        innerHeight := height - nodeBorder.VerticalPadding()
        childBoxes := gridLayout.LayoutChildren(innerWidth, innerHeight)
        // ... 递归布局，加上边框偏移
    } else if wrapProvider, ok := node.(WrapStyleProvider); ok {
        // Wrap 布局逻辑
        childBoxes := wrapLayout.LayoutChildren(width, height)
        // ... 递归布局，加上边框偏移
    } else if absProvider, ok := node.(AbsoluteStyleProvider); ok {
        // Absolute 布局逻辑
        for _, child := range node.Children() {
            childX, childY := absStyle.CalculatePosition(...)
            subBox := e.layoutNodeWithDepth(child, constraints, x+childX+borderOffsetX, y+childY+borderOffsetY, depth+1, visited)
            // ...
        }
    } else {
        // 默认布局
        for _, child := range node.Children() {
            childX := x + borderOffsetX  // ✨ 加上边框偏移
            childY := y + borderOffsetY
            subBox := e.layoutNodeWithDepth(child, constraints, childX, childY, depth+1, visited)
            box.Children = append(box.Children, subBox)
        }
    }

    return box
}
```

---

## 各布局类型的边框处理

### 1. Flex (Stack)

**处理逻辑**:

```
容器的总宽度 = 子节点最大宽度 + Gap + Padding
容器的总高度 = 子节点总高度 + Gap + Padding

有边框时:
  contentWidth = totalWidth - 2  (减去左右边框)
  contentHeight = totalHeight - 2 (减去上下边框)

子节点布局在 contentWidth x contentHeight 区域内
```

**示例**:

```go
// VStack 边框示例
vstack.SetWidth(20).SetHeight(10).Border("single", "Title")

// 容器总尺寸：20x10
// 边框占用：2x2
// 内容区域：18x8
// 子节点布局在 (1,1) 到 (19,9) 之间的区域
```

---

### 2. Grid

**处理逻辑**:

```
容器的总宽度 = 列宽之和 + ColumnGap
容器的总高度 = 行高之和 + RowGap

有边框时:
  contentWidth = totalWidth - 2
  contentHeight = totalHeight - 2

单元格子节点布局在调整后的网格内
```

**示例**:

```go
// Grid 边框示例
grid.New().
    SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
    SetBorder("double", "Grid")

// 容器总尺寸：40x20
// 边框占用：2x2
// 内容区域：38x18
// 每列宽度：19
// 单元格内容考虑边框偏移
```

---

### 3. Wrap

**处理逻辑**:

```
有边框时:
  可用宽度 = 容器宽度 - 2 (减去左右边框)
  子节点在可用宽度内换行

  子节点位置 = 计算位置 + (1, 1) (边框偏移)
```

**示例**:

```go
// Wrap 边框示例
wrap.New().
    SetWidth(30).
    SetBorder("rounded", "Tags")

// 容器总宽度：30
// 边框占用：左右各1
// 可用内容宽度：28
// 子节点在28宽度内换行
```

---

### 4. Absolute

**处理逻辑**:

```
有边框时:
  子节点位置 = 原始位置 + (1, 1) (边框偏移)

  例如 top:0, left:0 → 实际位置 top:1, left:1
```

**示例**:

```go
// Absolute 边框示例
absolute.New(child).
    SetBorder("dashed", "Abs")

// 子节点位置 (0, 0) 会变成 (1, 1)
// 考虑边框偏移
```

---

### 5. Modal

**处理逻辑**:

```
Modal 默认使用双线边框
Title 作为 Border Label

Modal 尺寸 = 内容区 + 边框
```

**示例**:

```go
// Modal 边框示例
modal.New().
    SetTitle("Confirm").
    SetOpen(true)

// 默认边框：double
// Border Label: "Confirm"
```

---

## 接口适配

### Bordered 接口

**位置**: `runtime/layout/border.go`

```go
type Bordered interface {
    Node
    GetBorder() Border
}
```

**实现者**:
- `FiberToNodeAdapter` ✅
- `BorderedNode` (旧包装组件)
- 其他容器组件（通过 FiberToNodeAdapter）

---

### FiberToNodeAdapter 实现

```go
// FiberToNodeAdapter 实现 layout.Bordered 接口
func (a *FiberToNodeAdapter) GetBorder() layout.Border {
    // ✨ 方案 A: 优先使用 Fiber.BorderStyle
    if a.fiber.BorderStyle != "" && a.fiber.BorderStyle != "none" {
        return layout.Border{
            Style: parseBorderStyleString(a.fiber.BorderStyle),
            Label: a.fiber.BorderLabel,
        }
    }

    // 向后兼容：旧组件使用 Props
    // ...
}
```

---

### Props 导出

每个容器 VNode 都导出边框属性到 Props:

```go
// Stack Props()
func (s *VNode) Props() rtui.Props {
    return rtui.Props{
        // ... 其他属性
        "borderStyle": s.borderStyle,  // ✨
        "label":       s.borderLabel,  // ✨
    }
}

// Grid Props()
func (g *VNode) Props() rtui.Props {
    return rtui.Props{
        // ... 其他属性
        "borderStyle": g.borderStyle,  // ✨
        "label":       g.borderLabel,  // ✨
    }
}
```

---

### Reconciler 同步

**位置**: `internal/reconciler/complete_work.go`

```go
// syncBorderProperties 将 VNode Props 的边框属性同步到 Fiber
func syncBorderProperties(element *Element, fiber *rtui.Fiber) {
    if fiber == nil || element == nil {
        return
    }

    props := element.Props
    if props == nil {
        return
    }

    // 从 Props 读取边框样式
    if borderStyle, ok := props["borderStyle"].(string); ok {
        fiber.BorderStyle = borderStyle
    }

    // 从 Props 读取边框标签
    if label, ok := props["label"].(string); ok {
        fiber.BorderLabel = label
    }
}
```

---

## 示例代码

### 示例 1: 基本边框容器

```go
import (
    "github.com/wwsheng009/mint/ui/components/stack"
    "github.com/wwsheng009/mint/ui/components/text"
)

func App() ui.VNode {
    return stack.NewVStack().
        SingleBorder("Title").
        SetGap(0).
        SetChildrenList([]ui.VNode{
            text.New("Item 1"),
            text.New("Item 2"),
            text.New("Item 3"),
        })
}

// 输出:
// ┌──────┐
// │Item 1│
// │Item 2│
// │Item 3│
// └──────┘
```

---

### 示例 2: 嵌套边框

```go
func NestedBorders() ui.VNode {
    return stack.NewVStack().
        DoubleBorder("Outer").
        SetGap(0).
        SetChildrenList([]ui.VNode{
            stack.NewVStack().
                SingleBorder("Inner").
                SetGap(0).
                SetChildrenList([]ui.VNode{
                    text.New("Content"),
                }),
        })
}

// 输出:
// ╔═══════╗
// ║┌─────┐║
// ║│Content│║
// ║└─────┘║
// ╚═══════╝
```

---

### 示例 3: Grid 边框

```go
import "github.com/wwsheng009/mint/ui/components/grid"

func GridWithBorder() ui.VNode {
    return grid.New().
        SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
        DashedBorder("2x2 Grid").
        SetChildrenAuto([]ui.VNode{
            text.New("A1"), text.New("B1"),
            text.New("A2"), text.New("B2"),
        })
}
```

---

### 示例 4: 不同边框样式

```go
func BorderStyles() ui.VNode {
    return stack.NewVStack().SetGap(1).SetChildrenList([]ui.VNode{
        stack.NewVStack().
            SingleBorder("Single").
            SetChildrenList([]ui.VNode{text.New("Single")}),
        stack.NewVStack().
            DoubleBorder("Double").
            SetChildrenList([]ui.VNode{text.New("Double")}),
        stack.NewVStack().
            RoundedBorder("Rounded").
            SetChildrenList([]ui.VNode{text.New("Rounded")}),
        stack.NewVStack().
            DashedBorder("Dashed").
            SetChildrenList([]ui.VNode{text.New("Dashed")}),
    })
}
```

---

## 测试覆盖

### 单元测试文件

| 测试文件 | 覆盖内容 |
|---------|---------|
| `runtime/layout/border_test.go` | Border 类型定义、BorderedNode、辅助函数 |
| `runtime/layout/comprehensive_layout_test.go` | 边框在布局引擎中的集成 |
| `runtime/layout/style_layout_test.go` | 边框与 Paddding/Margin 组合 |
| `internal/render/fiber_adapter_test.go` | FiberToNodeAdapter 边框处理 |
| `ui/components/modal/modal_test.go` | Modal 边框样式测试 |

---

### 关键测试用例

```go
// 边框空间计算测试
func TestBorder_Padding(t *testing.T) {
    tests := []struct {
        border            Border
        expectHorizontal  int
        expectVertical    int
    }{
        {Border{Style: BorderNone}, 0, 0},
        {Border{Style: BorderSingle}, 2, 2},
        {Border{Style: BorderDouble}, 2, 2},  // 双线也是 1-char
        {Border{Style: BorderRounded}, 2, 2},
    }
    // ...
}

// 嵌套边框测试
func TestEngine_Border_Nested(t *testing.T) {
    outer := NewMockCompositeNode("outer", 100, 50)
    outer.SetBorder(BorderSingle)

    inner := NewMockCompositeNode("inner", 96, 46)
    inner.SetBorder(BorderSingle)

    outer.SetChildren([]Node{inner})

    engine := NewEngine()
    result := engine.Layout(outer, UnboundedConstraints())

    // 内层应该考虑外层边框偏移
    innerBox := result.Root.Children[0]
    assert.Equal(t, 1, innerBox.X)  // 外层边框偏移
    assert.Equal(t, 1, innerBox.Y)
}

// Grid 边框测试
func TestGrid_WithBorder(t *testing.T) {
    grid := NewMockGridNode("grid", 102, 52)
    grid.SetBorder(BorderSingle)
    // ...
}
```

---

## 总结

### 方案 A 的优势

1. **API 简洁**: 直接在容器上设置边框，无需包装
   ```go
   // 旧方案
   border.New().SetBorder("single").SetChild(content)

   // 新方案
   stack.NewVStack().Border("single").SetChildren(content)
   ```

2. **性能更好**: 减少一层 Fiber 节点，减少内存分配

3. **类型安全**: 边框样式使用枚举，编译时检查

4. **布局高效**: 边框空间计算在布局引擎层完成

5. **向后兼容**: 保留旧 `Border` 组件的兼容层

---

### 数据流向

```
VNode Props
    ↓ (Reconciler)
Fiber.BorderStyle / Fiber.BorderLabel
    ↓ (FiberToNodeAdapter)
layout.Border
    ↓ (Layout Engine)
LayoutBox (包含边框信息)
    ↓ (Painter)
渲染到屏幕
```

---

### 边框空间原则

1. **边框占用空间**: 2x2 (左右各 1 字符，上下各 1 字符)
2. **内容偏移**: 始终 (1, 1)
3. **总尺寸**: 内容尺寸 + 边框空间
4. **测量阶段**: 先减边框测量内容，再加边框返回总尺寸
5. **布局阶段**: 子节点位置加上边框偏移

---

## 参考文件

- `runtime/layout/border.go` - 边框类型定义
- `runtime/layout/types.go` - 布局引擎核心逻辑
- `internal/render/fiber_adapter.go` - Fiber 到 layout.Node 适配
- `internal/reconciler/complete_work.go` - 边框属性同步
- `ui/components/stack/vnode.go` - Stack 容器边框支持
- `ui/components/grid/vnode.go` - Grid 容器边框支持
- `ui/components/wrap/vnode.go` - Wrap 容器边框支持
- `ui/components/absolute/vnode.go` - Absolute 容器边框支持
- `ui/components/modal/vnode.go` - Modal 容器边框支持
