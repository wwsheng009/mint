# Stretch Layout System - 拉伸布局系统

## 一、概述

### 1.1 系统简介

拉伸布局系统是 TUI 框架中用于实现组件自动扩展以填充可用空间的核心功能。该系统实现了类似 CSS Flexbox 的布局行为，允许子组件按照指定的比例分配父容器的可用空间。

### 1.2 核心概念

| 概念 | 说明 | 示例 |
|------|------|------|
| **Flex** | 弹性因子，控制子元素扩展比例 | `flex=1` 等分空间，`flex=2` 获取双倍空间 |
| **StretchCross** | 跨轴拉伸，所有子元素自动填充跨轴空间 | VStack 子元素横向拉伸，HStack 子元素纵向拉伸 |
| **Gap** | 子元素之间的间距 | `gap=0` 紧密排列，`gap=1` 间隔 1 字符 |
| **Constraints** | 约束传递，父容器向子元素传递空间限制 | `MaxWidth=80` 限制最大宽度为 80 |

### 1.3 设计目标

1. **约束驱动**: 父容器通过约束向子元素传递可用空间
2. **两阶段处理**: Measure 阶段计算尺寸，Layout 阶段计算位置
3. **职责单一**: 组件只负责测量，布局引擎负责位置计算
4. **性能优化**: 缓存测量结果，支持增量布局

---

## 二、架构设计

### 2.1 分层架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              用户代码层 (User Code)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ ui.Flex()   │  │ .Stretch()  │  │  .Gap(0)     │  │ .Align()     │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            VNode 抽象层 (runtime/ui)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ LayoutNode   │  │ BorderedNode │  │ ElementVNode │  │  GetLayout   │   │
│  │              │  │              │  │              │  │    Info()    │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          布局引擎层 (runtime/compute)                      │
│  ┌──────────────────────────────────┐  ┌──────────────────────────────────┐ │
│  │     Measure Phase (测量阶段)      │  │    Layout Phase (布局阶段)       │ │
│  │  ┌─────────────────────────────┐ │  │  ┌─────────────────────────────┐ │ │
│  │  │ measureVNode()              │ │  │  │ buildComputedBox()          │ │ │
│  │  │   ├─ measureLayoutChildren()│ │  │  │   ├─ getChildConstraints()  │ │ │
│  │  │   └─ measureBordered()      │ │  │  │   └─ calculatePositions()   │ │ │
│  │  │                             │ │  │  │                             │ │ │
│  │  │ Flex Distribution:          │ │  │  │ layoutHStack()              │ │ │
│  │  │ 1. 识别 flex 子元素           │ │  │  │ layoutVStack()              │ │ │
│  │  │ 2. 计算固定尺寸               │ │  │  │ layoutBordered()            │ │ │
│  │  │ 3. 分配剩余空间               │ │  │  │                             │ │ │
│  │  └─────────────────────────────┘ │  │  └─────────────────────────────┘ │ │
│  └──────────────────────────────────┘  └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          渲染引擎层 (internal/render)                       │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                      RenderingPipeline                                  │ │
│  │  ┌─────────────────────┐  ┌─────────────────────────────────────────┐  │ │
│  │  │  Layout Engine      │  │           Paint Engine                   │  │ │
│  │  │  (compute.Engine)   │  │     (render.PaintEngine)                │  │ │
│  │  │                     │  │                                         │  │ │
│  │  │  - ComputedLayout   │  │  - Paint computed boxes                │  │ │
│  │  │  - ComputedBox      │  │  - Handle borders                      │  │ │
│  │  │  - LayoutCache      │  │  - Render text                         │  │ │
│  │  └─────────────────────┘  └─────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
用户代码
   │
   │ ui.Flex(bordered, 1) ────> 设置 props["flex"] = 1
   │ .Stretch() ──────────────> 设置 node.stretchCross = true
   │ .Gap(0) ─────────────────> 设置 node.gap = 0
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Measure Phase (测量阶段)                                     │
│                                                             │
│  LayoutNode.Measure(constraints)                            │
│       │                                                     │
│       ├──► 1. 遍历子元素，识别 flex 子元素                     │
│       │                                                     │
│       ├──► 2. 测量非 flex 子元素，计算 fixedWidth             │
│       │                                                     │
│       ├──► 3. 计算剩余空间 = MaxWidth - fixedWidth - gaps    │
│       │                                                     │
│       └──► 4. 按 flex 比例分配剩余空间给 flex 子元素           │
│                                                             │
│  结果: 返回 Size{Width, Height}                             │
└─────────────────────────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Build Phase (构建阶段)                                       │
│                                                             │
│  buildComputedBox(vnode, constraints)                       │
│       │                                                     │
│       └──► getChildConstraints(parent, child, constraints)  │
│             │                                               │
│             ├──► HStack: 计算 flex 子元素的宽度约束          │
│             │     flexWidth = (remaining * flexFactor) / totalFlex
│             │                                               │
│             └──► VStack: 计算 flex 子元素的高度约束          │
│                   flexHeight = (remaining * flexFactor) / totalFlex
│                                                             │
│  结果: 创建 ComputedBox 树                                   │
└─────────────────────────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Layout Phase (布局阶段)                                     │
│                                                             │
│  calculatePositions(box, x, y)                             │
│       │                                                     │
│       └──► layoutVStack() / layoutHStack()                  │
│             │                                               │
│             └──► 应用 StretchCross，拉伸子元素到跨轴尺寸       │
│                                                             │
│  结果: 设置所有 Box.X, Box.Y                                │
└─────────────────────────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Paint Phase (绘制阶段)                                      │
│                                                             │
│  PaintEngine.Paint(layout, buffer)                         │
│       │                                                     │
│       └──► 根据 ComputedBox 位置绘制                        │
│                                                             │
│  结果: 渲染到 Buffer                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、核心实现

### 3.1 Flex 分布算法 (Measure 阶段)

Flex 分布在 `LayoutNode.Measure()` 和 `measureLayoutChildren()` 中实现：

```go
// runtime/ui/layout.go:287-379
func (l *LayoutNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    children := l.Children()

    // 第一遍：识别 flex 子元素，测量非 flex 子元素
    var flexChildren []struct {
        child  VNode
        factor int
    }
    var fixedWidth int   // HStack: 固定宽度总和
    var fixedHeight int  // VStack: 固定高度总和
    flexTotalFactor := 0

    for i, child := range children {
        childInfo := GetLayoutInfo(child)
        if childInfo.Flex > 0 {
            // 收集 flex 子元素
            flexChildren = append(flexChildren, struct {
                child  VNode
                factor int
            }{child, childInfo.Flex})
            flexTotalFactor += childInfo.Flex
        } else {
            // 测量非 flex 子元素
            childSize := l.measureChild(child, childConstraints)
            fixedWidth += childSize.Width  // 或 fixedHeight += childSize.Height
        }
        // 累加 gap
        if i < len(children)-1 {
            fixedWidth += l.gap
        }
    }

    // 第二遍：如果有 flex 子元素且父容器有界，分配剩余空间
    if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
        availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*l.gap
        remainingSpace := availableWidth - fixedWidth

        // 按比例分配给 flex 子元素
        for _, fc := range flexChildren {
            flexWidth := (remainingSpace * fc.factor) / flexTotalFactor

            // 使用固定约束测量 flex 子元素
            childConstraints := runtime.BoxConstraints{
                MinWidth:  flexWidth,
                MaxWidth:  flexWidth,
                MinHeight: 0,
                MaxHeight: innerMaxHeight,
            }
            childSize := l.measureChild(fc.child, childConstraints)
            totalWidth += childSize.Width
        }
    }

    return runtime.Size{Width: totalWidth, Height: totalHeight}
}
```

**关键点**:
1. 两遍测量：第一遍识别 flex/非 flex 子元素，第二遍分配空间
2. 约束传递：flex 子元素接收固定宽度的约束 `MinWidth=MaxWidth=flexWidth`
3. 跨轴填充：VStack 填充父容器宽度，HStack 填充父容器高度

### 3.2 Flex 约束传递 (Build 阶段)

`getChildConstraints()` 函数确保 flex 子元素在构建阶段也获得正确的约束：

```go
// runtime/compute/engine.go:578-732
func (e *Engine) getChildConstraints(parent, child VNode,
    parentConstraints runtime.BoxConstraints, parentSize runtime.Size) runtime.BoxConstraints {

    layoutInfo := rtui.GetLayoutInfo(child)

    if tagger, ok := parent.(interface{ Tag() string }); ok {
        switch tagger.Tag() {
        case "hstack":
            // HStack: 计算 flex 宽度分布
            if layoutInfo.Flex > 0 && parentConstraints.HasBoundedWidth() {
                // 计算所有兄弟元素的总 flex 和固定宽度
                parentChildren := parent.Children()
                var totalFlexFactor int
                var fixedWidth int

                for _, sibling := range parentChildren {
                    siblingInfo := rtui.GetLayoutInfo(sibling)
                    if siblingInfo.Flex > 0 {
                        totalFlexFactor += siblingInfo.Flex
                    } else {
                        // 测量非 flex 兄弟
                        siblingSize := e.measureVNode(sibling, siblingConstraints)
                        fixedWidth += siblingSize.Width
                    }
                }

                // 计算此 flex 子元素的宽度
                availableWidth := parentConstraints.MaxWidth - paddingWidth - gaps
                remainingSpace := availableWidth - fixedWidth
                flexWidth := (remainingSpace * layoutInfo.Flex) / totalFlexFactor

                return runtime.BoxConstraints{
                    MinWidth:  flexWidth,
                    MaxWidth:  flexWidth,  // 固定宽度
                    MinHeight: 0,
                    MaxHeight: childMaxHeight,
                }
            }

        case "vstack":
            // VStack: 计算 flex 高度分布 (类似逻辑)
            // ...
        }
    }

    return parentConstraints
}
```

**关键点**:
1. 动态计算：每次调用都重新计算所有兄弟元素的 flex 分布
2. 约束一致性：确保 Measure 和 Build 阶段的约束一致
3. 兄弟感知：需要遍历所有兄弟元素来确定正确的 flex 约束

### 3.3 StretchCross 跨轴拉伸 (Layout 阶段)

`StretchCross` 让所有子元素自动填充跨轴空间：

```go
// runtime/compute/engine.go:578-645
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    stretchCross := layoutInfo.StretchCross

    for _, child := range box.Children {
        childInfo := rtui.GetLayoutInfo(child.VNode)
        oldWidth := child.Box.Width

        // 拉伸条件：子元素有 flex OR 父容器启用 StretchCross
        if (childInfo.Flex > 0 || stretchCross) && box.Box.Width < runtime.Infinity {
            child.Box.Width = box.Box.Width  // 拉伸到容器宽度

            // 如果是文本节点，添加空格填充
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

**关键点**:
1. 双重条件：子元素有 `flex > 0` 或父容器启用 `StretchCross`
2. 文本填充：拉伸文本节点时，在内容后追加空格
3. 有限检查：只在 `Width < Infinity` 时拉伸

---

## 四、API 使用指南

### 4.1 基本用法

```go
// 1. Flex 包装 - 让子元素按比例扩展
ui.Flex(
    ui.Bordered().Child(content).Build(),
    1,  // flex factor (可选，默认 1)
)

// 2. Stretch - 所有子元素自动拉伸到跨轴尺寸
ui.VStackBuilder(
    item1, item2, item3,
).Stretch().Build()

// 3. Gap - 控制子元素间距
ui.HStackBuilder(
    left, right,
).Gap(0).Build()  // 无间距，紧密排列

// 4. 组合使用
ui.HStackBuilder(
    ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),
    ui.Flex(ui.Bordered().Child(content).Build(), 1),
).Gap(0).Build()
```

### 4.2 典型场景

#### 场景 1: 左右分栏布局

```go
// 目标：左右两栏等宽，无间隙
ui.HStackBuilder(
    ui.Flex(ui.Bordered().Child(leftPanel).Build(), 1),
    ui.Flex(ui.Bordered().Child(rightPanel).Build(), 1),
).Gap(0).Build()

// 结果：40 + 0 + 40 = 80 (完美平分)
```

#### 场景 2: 侧边栏 + 主内容 (1:2 比例)

```go
ui.HStackBuilder(
    ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),   // 1/3 宽度
    ui.Flex(ui.Bordered().Child(content).Build(), 2),    // 2/3 宽度
).Gap(1).Build()

// 结果：约 26 + 1 + 53 = 80
```

#### 场景 3: VStack 子元素横向填充

```go
// 目标：所有子元素拉伸到容器宽度
ui.VStackBuilder(
    item1, item2, item3,
).Stretch().Build()

// 效果：每个 item 都会拉伸到父容器的宽度
```

#### 场景 4: 嵌套布局

```go
// 根 VStack 启用 Stretch，子 HStack 使用 Flex
ui.VStackBuilder(
    ui.HStackBuilder(  // Header
        ui.Text("Title"),
        ui.Text("       "),  // 空格填充
        ui.Button("Action"),
    ).Stretch().Build(),

    ui.HStackBuilder(  // Body - 左右分栏
        ui.Flex(ui.Bordered().Child(sidebar).Build(), 1),
        ui.Flex(ui.Bordered().Child(content).Build(), 1),
    ).Gap(0).Build(),
).Stretch().Build()
```

---

## 五、注意事项与最佳实践

### 5.1 常见陷阱

| 问题 | 症状 | 解决方案 |
|------|------|----------|
| **Gap 导致不对齐** | 元素宽度总和 + gap ≠ 父容器宽度 | 使用 `.Gap(0)` 或调整 flex 计算 |
| **文本未拉伸** | VStack 子元素宽度不匹配 | 启用 `.Stretch()` 或给子元素添加 flex |
| **边框溢出** | Bordered 内容超出父容器 | 减少内容宽度或增加父容器约束 |
| **嵌套 Flex 失效** | 内层 flex 不生效 | 检查外层容器的约束是否传递 |

### 5.2 性能考虑

1. **缓存优化**: 叶子节点（无子元素的节点）的测量结果会被缓存
   ```go
   // 缓存 Key 包含：VNode 类型 + 约束 + 内容 hash
   type LayoutCacheKey struct {
       VNodeType   string
       Constraints BoxConstraints
       VNodeKey    string
       ContentHash uint64
       PropsHash   uint64
   }
   ```

2. **避免重复测量**:
   - 使用 `Key()` 为静态元素设置唯一标识
   - 避免在 Measure 中进行昂贵的计算

3. **增量布局**:
   - 使用 `DirtyTracker` 标记需要重新布局的节点
   - 未变化的子树可以跳过重新布局

### 5.3 调试技巧

```bash
# 启用布局调试
TUI_LAYOUT_DEBUG=true go run ./examples/demo1

# 启用拉伸调试
TUI_STRETCH_DEBUG=true go run ./examples/demo1

# 启用管道调试
TUI_PIPELINE_DEBUG=true go run ./examples/demo1
```

---

## 六、架构评估与建议

### 6.1 当前架构的优势

| 优势 | 说明 |
|------|------|
| **两阶段清晰** | Measure 计算尺寸，Layout 计算位置，职责分明 |
| **约束驱动** | 父容器通过约束向子元素传递可用空间 |
| **灵活扩展** | 支持多种布局方式（Flex, Stretch, Gap, Align） |
| **性能优化** | 缓存机制减少重复计算 |

### 6.2 已知限制

| 限制 | 影响 | 缓解方案 |
|------|------|----------|
| **Flex 只支持整数** | 无法实现精确的小数比例 | 使用更大的 flex 值（如 2:3 用 4:6 代替） |
| **Flex 计算重复** | getChildConstraints 每次都遍历兄弟元素 | 考虑在 Measure 阶段缓存计算结果 |
| **文本填充简单** | 只追加空格，不支持对齐 | 添加对齐选项（左/中/右） |
| **无最小尺寸保证** | Flex 子元素可能被压缩到 0 | 添加 MinWidth/MinHeight 支持 |

### 6.3 改进建议

#### 短期改进 (1-2 周)

1. **优化 Flex 计算**
   ```go
   // 在 Measure 阶段缓存 flex 计算结果
   type FlexDistribution struct {
       ChildFlex   []int
       ChildWidths []int
       TotalFlex   int
       Remaining   int
   }
   ```

2. **支持 MinWidth/MinHeight**
   ```go
   // Flex 子元素的最小尺寸
   if flexWidth < minWidth {
       flexWidth = minWidth
   }
   ```

3. **改进文本对齐**
   ```go
   type TextAlign int
   const (
       TextAlignLeft TextAlign = iota
       TextAlignCenter
       TextAlignRight
   )
   ```

#### 中期改进 (1-2 月)

1. **百分比尺寸支持**
   ```go
   // 支持 "50%" 这样的尺寸指定
   type Size struct {
       Value   int
       Percent float64  // 0.0 - 1.0
   }
   ```

2. **Grid 布局系统**
   ```go
   // 二维网格布局
   ui.Grid(
       rows: 2, cols: 3,
       gap: 1,
       children: [...],
   )
   ```

3. **绝对定位支持**
   ```go
   // 允许某些元素脱离正常布局流
   ui.Absolute().X(10).Y(5).Child(element)
   ```

#### 长期改进 (3-6 月)

1. **布局验证器**
   ```go
   // 检测布局冲突和循环依赖
   type LayoutValidator struct{}
   func (v *LayoutValidator) Validate(vnode VNode) []error
   ```

2. **布局动画**
   ```go
   // 平滑的布局过渡动画
   type LayoutAnimation struct {
       Duration time.Duration
       Easing  func(float64) float64
   }
   ```

3. **响应式布局**
   ```go
   // 根据容器尺寸自动调整布局
   type ResponsiveLayout struct {
       Breakpoints map[int]VNode  // width -> layout
   }
   ```

### 6.4 代码维护指南

#### 添加新的布局容器

1. 实现 `Measurable` 接口
2. 实现 `Measure()` 方法，返回约束驱动的尺寸
3. 在 `getChildConstraints()` 中添加对应的 case
4. 在 `calculatePositions()` 中添加布局逻辑
5. 添加单元测试

#### 修改 Flex 行为

1. 确认修改哪个阶段（Measure/Build/Layout）
2. 同步修改 `runtime/ui/layout.go` 和 `runtime/compute/engine.go`
3. 更新测试用例
4. 检查是否影响现有组件

#### 调试布局问题

1. 启用调试标志查看约束传递
2. 检查 `GetLayoutInfo()` 返回值
3. 验证 `HasBoundedWidth/Height()` 判断
4. 确认 `SubtractPadding()` 正确处理 Infinity

---

## 七、参考文件

| 文件 | 职责 |
|------|------|
| `runtime/ui/layout.go` | LayoutNode, HStack, VStack 实现 |
| `runtime/ui/layout_util.go` | GetLayoutInfo, LayoutInfo 结构 |
| `runtime/compute/engine.go` | 布局引擎，getChildConstraints |
| `runtime/compute/types.go` | BoxConstraints, ComputedBox |
| `internal/render/rendering_pipeline.go` | 渲染管线 |
| `internal/render/paint_engine.go` | 绘制引擎 |
| `ui/layout.go` | 公开 API 重新导出 |

---

## 八、版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2025-02-05 | 初始版本，基础 Flex 支持 |
| 1.1 | 2025-02-06 | 添加 HStackBuilder，优化 Flex 分发算法 |
| 1.2 | 2025-02-06 | 完善文档，添加架构图和示例 |

---

*文档维护者: Claude*
*最后更新: 2025-02-06*
