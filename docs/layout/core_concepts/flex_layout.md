# Flex Layout Implementation

> **版本**: 2.0
> **最后更新**: 2025-02-06
> **状态**: ✅ 已实现

## Overview

Flex layout 是一个类似 CSS Flexbox 的布局系统，允许子组件按照指定的比例分配父容器的可用空间。

### 核心特性

| 特性 | 状态 | 说明 |
|------|------|------|
| Flex 比例分配 | ✅ | 支持 flex factor 控制空间分配比例 |
| 主轴自适应 | ✅ | HStack/VStack 自动填充主轴空间 |
| 跨轴拉伸 | ✅ | StretchCross 让子元素填充跨轴空间 |
| 间距控制 | ✅ | Gap 控制子元素之间的间距 |
| 约束驱动 | ✅ | 基于 BoxConstraints 的布局计算 |

---

## API 使用

### 基本示例

```go
import "github.com/wwsheng009/mint/ui"

// 1. Flex 包装 - 子元素按比例扩展
ui.HStack(
    ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),   // 1/2 宽度
    ui.Flex(ui.Bordered().Child(content).Build(), 1),    // 1/2 宽度
)

// 2. 不同比例
ui.HStack(
    ui.Flex(left, 1),   // 1/3 宽度
    ui.Flex(middle, 1), // 1/3 宽度
    ui.Flex(right, 1),  // 1/3 宽度
)

// 3. 使用 HStackBuilder 控制 Gap
ui.HStackBuilder(
    ui.Flex(left, 1),
    ui.Flex(right, 1),
).Gap(0).Build()  // 无间距
```

### VStack 横向拉伸

```go
// VStack 启用 Stretch，子元素自动填充宽度
ui.VStackBuilder(
    item1,
    item2,
    item3,
).Stretch().Build()

// 效果：每个 item 都会拉伸到父容器的宽度
```

---

## 实现架构

### 两阶段渲染

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Measure (测量阶段)                                  │
│                                                             │
│  LayoutNode.Measure(constraints)                            │
│       │                                                     │
│       ├──► 1. 识别 flex 子元素                               │
│       ├──► 2. 测量非 flex 子元素                             │
│       ├──► 3. 计算剩余空间                                  │
│       └──► 4. 按 flex 比例分配                              │
│                                                             │
│  结果: Size{Width, Height}                                  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Layout (布局阶段)                                   │
│                                                             │
│  buildComputedBox()                                         │
│       │                                                     │
│       └──► getChildConstraints()                            │
│             │                                               │
│             ├──► HStack: 计算 flexWidth                     │
│             └──► VStack: 计算 flexHeight                    │
│                                                             │
│  calculatePositions()                                       │
│       │                                                     │
│       └──► 应用 StretchCross                                │
│                                                             │
│  结果: ComputedBox{x, y, w, h}                              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Paint (绘制阶段)                                    │
│                                                             │
│  PaintEngine.Paint(layout, buffer)                          │
│                                                             │
│  结果: 渲染到屏幕                                            │
└─────────────────────────────────────────────────────────────┘
```

### Flex 分布算法

**HStack 主轴宽度分配**:

```go
// runtime/ui/layout.go:287-379
func (l *LayoutNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // 第一遍：识别 flex/非 flex 子元素
    var flexChildren []struct{ child VNode; factor int }
    var fixedWidth int
    flexTotalFactor := 0

    for i, child := range children {
        childInfo := GetLayoutInfo(child)
        if childInfo.Flex > 0 {
            flexChildren = append(flexChildren, ...)
            flexTotalFactor += childInfo.Flex
        } else {
            // 测量非 flex 子元素
            fixedWidth += childSize.Width
        }
        fixedWidth += gap
    }

    // 第二遍：分配剩余空间给 flex 子元素
    if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
        availableWidth := constraints.MaxWidth - padding - gaps
        remainingSpace := availableWidth - fixedWidth

        for _, fc := range flexChildren {
            flexWidth := (remainingSpace * fc.factor) / flexTotalFactor

            // 使用固定约束测量
            childConstraints := BoxConstraints{
                MinWidth:  flexWidth,
                MaxWidth:  flexWidth,
                ...
            }
            childSize := measureChild(fc.child, childConstraints)
            totalWidth += childSize.Width
        }
    }

    return Size{Width: totalWidth, Height: totalHeight}
}
```

**Flex 宽度计算公式**:

```
flexWidth = (remainingSpace × childFlexFactor) / totalFlexFactor

其中：
- remainingSpace = MaxWidth - fixedWidth - (n-1) × gap
- fixedWidth = Σ(非 flex 子元素宽度)
- totalFlexFactor = Σ(flex 子元素的 flex factor)
```

---

## 文件结构

### 核心文件

| 文件 | 职责 |
|------|------|
| `runtime/ui/layout.go` | LayoutNode.Measure() 实现 |
| `runtime/ui/layout_util.go` | GetLayoutInfo() 提取布局信息 |
| `runtime/layout/layout_engine.go` | getChildConstraints() 约束传递 |
| `ui/layout.go` | 公开 API 重新导出 |

### 新增 API

| API | 文件 | 说明 |
|-----|------|------|
| `HStackBuilder()` | `runtime/ui/layout.go` | HStack 构建器 |
| `HStackBuilder()` | `ui/layout.go` | 公开 API 重新导出 |
| `.Gap(int)` | `runtime/ui/layout.go` | 设置间距 |
| `.Stretch()` | `runtime/ui/layout.go` | 启用跨轴拉伸 |

---

## 数据流详解

### 1. Flex 属性设置

```go
// 用户代码
ui.Flex(borderedNode, 1)

// Flex 函数 (runtime/ui/layout.go:239-258)
func Flex(vnode VNode, flexFactors ...int) VNode {
    flex := 1
    if len(flexFactors) > 0 {
        flex = flexFactors[0]
    }
    // 设置到 props
    if n, ok := vnode.(interface{ SetProp(string, interface{}) }); ok {
        n.SetProp("flex", flex)
        return vnode
    }
    // ...
}
```

### 2. LayoutInfo 提取

```go
// GetLayoutInfo (runtime/ui/layout_util.go:50-146)
func GetLayoutInfo(vnode VNode) LayoutInfo {
    info := LayoutInfo{Flex: 0}

    // 检查 LayoutNode
    if layoutNode, ok := vnode.(*LayoutNode); ok {
        info.Flex = layoutNode.Flex()
        // 检查 props 覆盖
        if props := vnode.Props(); props != nil {
            if f, ok := props["flex"].(int); ok {
                info.Flex = f
            }
        }
        return info
    }

    // 检查 BorderedNode
    if _, ok := vnode.(*BorderedNode); ok {
        if props := vnode.Props(); props != nil {
            if f, ok := props["flex"].(int); ok {
                info.Flex = f
            }
        }
        return info
    }

    // 检查 ElementVNode
    // ...
}
```

### 3. 约束传递

```go
// getChildConstraints (runtime/layout/layout_engine.go:592-658)
case "hstack":
    childInfo := rtui.GetLayoutInfo(child)

    if childInfo.Flex > 0 && parentConstraints.HasBoundedWidth() {
        // 计算所有兄弟元素的 flex 分布
        parentChildren := parent.Children()
        var totalFlexFactor int
        var fixedWidth int

        for _, sibling := range parentChildren {
            siblingInfo := rtui.GetLayoutInfo(sibling)
            if siblingInfo.Flex > 0 {
                totalFlexFactor += siblingInfo.Flex
            } else {
                siblingSize := e.measureVNode(sibling, ...)
                fixedWidth += siblingSize.Width
            }
        }

        // 计算此 flex 子元素的宽度
        availableWidth := parentConstraints.MaxWidth - paddingWidth - gaps
        remainingSpace := availableWidth - fixedWidth
        flexWidth := (remainingSpace * childInfo.Flex) / totalFlexFactor

        return BoxConstraints{
            MinWidth:  flexWidth,
            MaxWidth:  flexWidth,
            MinHeight: 0,
            MaxHeight: childMaxHeight,
        }
    }
```

### 4. StretchCross 应用

```go
// layoutVStack (runtime/layout/layout_engine.go:579-645)
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    stretchCross := layoutInfo.StretchCross

    for _, child := range box.Children {
        childInfo := rtui.GetLayoutInfo(child.VNode)

        // 拉伸条件：子元素有 flex OR 父容器启用 StretchCross
        if (childInfo.Flex > 0 || stretchCross) && box.Box.Width < runtime.Infinity {
            child.Box.Width = box.Box.Width

            // 文本节点：添加空格填充
            if text := rtui.GetTextContent(child.VNode); text != "" {
                padding := child.Box.Width - e.measureTextWidth(text)
                if padding > 0 && padding < 1000 {
                    for i := 0; i < padding; i++ {
                        text += " "
                    }
                    child.RenderedText = text
                }
            }
        }

        e.calculatePositions(child, childX, childY)
        childY += child.Box.Height + gap
    }
}
```

---

## 使用示例

### 示例 1: 左右分栏

```go
// 目标：两栏等宽，无间隙
ui.HStackBuilder(
    ui.Flex(
        ui.Bordered().Color("blue").Child(sidebar).Build(),
        1,
    ),
    ui.Flex(
        ui.Bordered().Color("blue").Child(content).Build(),
        1,
    ),
).Gap(0).Build()

// 布局结果 (80 宽容器):
// ┌────────────────┬────────────────┐
// │  Sidebar       │  Content        │
// │  (40宽)        │  (40宽)         │
// └────────────────┴────────────────┘
```

### 示例 2: 1:2 比例分栏

```go
ui.HStack(
    ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),   // 约 26
    ui.Flex(ui.Bordered().Child(content).Build(), 2),    // 约 53
)

// 布局结果 (80 宽容器, gap=1):
// ┌──────────────┬────────────────────────────────┐
// │  Sidebar      │  Content                         │
// │  (26宽)       │  (53宽)                          │
// └──────────────┴────────────────────────────────┘
```

### 示例 3: VStack 横向填充

```go
ui.VStackBuilder(
    ui.Text("Item 1"),
    ui.Text("Item 2"),
    ui.Text("Item 3"),
).Stretch().Build()

// 效果：每个 Text 拉伸到容器宽度
// ┌────────────────────────────────────────────────┐
// │ Item 1                                        │
// │ Item 2                                        │
// │ Item 3                                        │
// └────────────────────────────────────────────────┘
```

### 示例 4: 完整布局

```go
// 根 VStack 启用 Stretch
mainContent := ui.VStackBuilder(
    // Header: 自动填充宽度
    ui.Bordered().Color("blue").Child(
        ui.HStackBuilder(
            ui.Text("Title"),
            ui.Text("       "),  // 填充空格
            ui.Button("Action"),
        ).Stretch().Build(),
    ).Build(),

    // Body: 左右分栏，等宽无间隙
    ui.HStackBuilder(
        ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),
        ui.Flex(ui.Bordered().Child(content).Build(), 1),
    ).Gap(0).Build(),
).Stretch().Build()

// 布局结果 (80x24):
// ┌────────────────────────────────────────────────────────────┐
// │ Title                                     [Action]        │ ← Header
// ├──────────────────────────────┬─────────────────────────────┤
// │ Sidebar                       │ Content                      │ ← Body
// │                               │                              │
// │                               │                              │
// └──────────────────────────────┴─────────────────────────────┘
```

---

## 调试

### 启用调试输出

```bash
# 布局调试
TUI_LAYOUT_DEBUG=true go run ./examples/demo1

# 拉伸调试
TUI_STRETCH_DEBUG=true go run ./examples/demo1

# 管道调试
TUI_PIPELINE_DEBUG=true go run ./examples/demo1
```

### 调试输出示例

```
[Layout.Measure] Element: constraints={0 80 0 24}, size={80 13}
[getChildConstraints.HStack] child flex=1/2, flexWidth=39
[Layout.Position] Element at (0,0,80×3)
[layoutVStack] box=(0,0,80x13) stretchCross=true
[layoutVStack]   stretch child: 4 -> 80 (text="Title")
```

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2025-02-05 | 初始 Flex 支持 |
| 1.1 | 2025-02-05 | 添加 StretchCross |
| 2.0 | 2025-02-06 | 完整 Flex 分布算法，HStackBuilder |
| 2.1 | 2025-02-06 | 添加文本对齐选项 (TextAlign, TextCenter, TextRight) |
| 2.2 | 2025-02-06 | 优化 Flex 缓存，O(N²) → O(N) |

---

## 相关文档

- [Stretch Layout System](./stretch_layout.md) - 完整的拉伸布局系统文档
- Layout refactor history - archived under `../../../docsArchive/cleanup-2026-05-19/docs/layout/refactor/`
- [Rendering Pipeline](/docsArchive/LAYOUT_RENDERING_REFACTOR.md) - 渲染管线
