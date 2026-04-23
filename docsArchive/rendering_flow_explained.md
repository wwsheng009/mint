# Mint TUI 渲染流程完整解析

## 📋 目录

1. [完整渲染流程](#完整渲染流程)
2. [关键数据结构](#关键数据结构)
3. [Layout阶段详解](#layout阶段详解)
4. [Paint阶段详解](#paint阶段详解)
5. [Flex布局实例分析](#flex布局实例分析)
6. [最近修复的问题](#最近修复的问题)
7. [调试技巧](#调试技巧)

---

## 完整渲染流程

```
用户代码
  ↓
RenderingPipeline.Render(vnode, constraints, buffer)
  │
  ├─→ Phase 1: Layout (布局计算)
  │   └─→ Engine.Layout(vnode, constraints)
  │       ├─→ buildComputedBox() [第一遍：测量]
  │       │   ├─→ 调用 vnode.Measure(constraints)
  │       │   ├─→ 对于 LayoutNode：调用 MeasureLayout()
  │       │   └─→ 返回每个子节点的约束和大小
  │       │
  │       └─→ calculatePositions() [第二遍：定位]
  │           ├─→ 计算每个节点的 (x, y) 坐标
  │           ├─→ 调用 SetBounds(x, y, width, height)
  │           └─→ 返回 ComputedLayout
  │
  └─→ Phase 2: Paint (渲染绘制)
      └─→ PaintEngine.Paint(layout, buffer)
          └─→ paintNode(box, buffer)
              ├─→ 检查是否实现 Paintable 接口
              │   ├─→ Button.Paint(x, y)
              │   └─→ 返回 []paint.DrawCmd
              │
              └─→ 将 DrawCmd 写入 buffer
```

---

## 关键数据结构

### ComputedBox (布局计算结果)

```go
type ComputedBox struct {
    VNode        VNode          // 虚拟节点
    Parent       *ComputedBox   // 父节点
    Children     []*ComputedBox // 子节点
    Box          runtime.Box    // 位置和大小 {X, Y, Width, Height}
    NaturalWidth int            // 自然宽度（无约束时的宽度）
}
```

### BoxConstraints (布局约束)

```go
type BoxConstraints struct {
    MinWidth  int  // 最小宽度
    MaxWidth  int  // 最大宽度
    MinHeight int  // 最小高度
    MaxHeight int  // 最大高度
}
```

### ComputedLayout (完整布局)

```go
type ComputedLayout struct {
    Root *ComputedBox  // 根节点（包含整个布局树）
}
```

---

## Layout阶段详解

### 第一步：测量 (buildComputedBox)

```go
func (e *Engine) buildComputedBox(vnode VNode, parent *ComputedBox, constraints BoxConstraints) *ComputedBox {
    box := &ComputedBox{
        VNode: vnode,
        Box:   {X: 0, Y: 0, Width: 0, Height: 0},
    }

    // 测量自然宽度（无约束）
    if measurable, ok := vnode.(interface{ Measure(BoxConstraints) Size }); ok {
        naturalSize := measurable.Measure({0, Infinity, 0, Infinity})
        box.NaturalWidth = naturalSize.Width
    }

    // 对于 LayoutNode：使用单次测量
    if layoutMeasurer, ok := vnode.(LayoutMeasurer); ok {
        measurement := layoutMeasurer.MeasureLayout(measurer, constraints)
        // measurement.ChildConstraints 包含每个子节点的约束
        // measurement.ChildSizes 包含每个子节点的大小
    }

    return box
}
```

### 第二步：定位 (calculatePositions)

```go
func (e *Engine) calculatePositions(box *ComputedBox, x, y int) {
    // 设置当前节点的位置
    box.Box.X = x
    box.Box.Y = y

    // ⭐ 关键：将布局结果存储到 VNode 中
    if boundsAware, ok := box.VNode.(interface{ SetBounds(int, int, int, int) }); ok {
        boundsAware.SetBounds(x, y, box.Box.Width, box.Box.Height)
    }

    // 递归处理子节点
    switch box.VNode.Tag() {
    case "hstack":
        e.layoutHStack(box, x, y)  // 水平布局
    case "vstack":
        e.layoutVStack(box, x, y)  // 垂直布局
    }
}
```

### HStack定位逻辑

```go
func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
    layoutInfo := GetLayoutInfo(box.VNode)
    gap := layoutInfo.Gap

    currentX := x
    for _, child := range box.Children {
        // 设置子节点位置
        child.Box.X = currentX
        child.Box.Y = y + calculateCrossOffset(child)

        // 调用 SetBounds
        child.SetBounds(currentX, child.Box.Y, child.Box.Width, child.Box.Height)

        // 移动到下一个位置
        currentX += child.Box.Width + gap
    }
}
```

---

## Paint阶段详解

### PaintEngine.Paint

```go
func (e *PaintEngine) Paint(layout *ComputedLayout, buffer *Buffer) error {
    return e.paintNode(layout.Root, buffer)
}

func (e *PaintEngine) paintNode(box *ComputedBox, buffer *Buffer) error {
    // 优先检查是否实现 Paintable 接口
    if paintable, ok := box.VNode.(interface{ Paint(int, int) []DrawCmd }); ok {
        // 组件有自定义渲染逻辑
        commands := paintable.Paint(box.Box.X, box.Box.Y)
        for _, cmd := range commands {
            buffer.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
        }
        return nil  // Paintable组件自己处理渲染（包括子节点）
    }

    // 否则，使用默认渲染逻辑
    e.paintDefault(box, buffer)
    return e.paintChildren(box, buffer)
}
```

### Button.Paint 实现

```go
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // 1. 构建按钮文本
    buttonText := focusIndicator + labelText  // " [ Left ]"
    contentWidth := len(buttonText)          // 9

    // 2. 获取用户指定的 padding（从 BoxModelMixin）
    padding := b.Padding()
    paddingLeft := padding[3]   // 用户设置的左padding
    paddingRight := padding[1]  // 用户设置的右padding

    // 3. 计算自然宽度（只包含内容，不包含padding）
    naturalWidth := contentWidth  // 9

    // 4. 获取布局分配的宽度（从 bounds，由 layout engine 设置）
    layoutWidth := naturalWidth
    if b.bounds[2] > 0 {
        layoutWidth = b.bounds[2]  // 26 (由 flex 分配)
    }

    // 5. 如果按钮被拉伸，应用文本对齐
    if layoutWidth > naturalWidth {
        textAlign := b.TextAlign()
        availableSpace := layoutWidth - naturalWidth  // 26 - 9 = 17

        switch textAlign {
        case AlignCenter:
            // 居中：左右两边分配空间
            leftSpace := paddingLeft + availableSpace/2
            rightSpace := paddingRight + (availableSpace - availableSpace/2)
            buttonText = strings.Repeat(" ", leftSpace) + buttonText +
                         strings.Repeat(" ", rightSpace)

        case AlignStart:
            // 左对齐：所有空间加到右边
            buttonText = strings.Repeat(" ", paddingLeft) + buttonText +
                         strings.Repeat(" ", paddingRight + availableSpace)

        case AlignEnd:
            // 右对齐：所有空间加到左边
            leftSpace := paddingLeft + availableSpace
            buttonText = strings.Repeat(" ", leftSpace) + buttonText +
                         strings.Repeat(" ", paddingRight)
        }
    }

    // 6. 返回绘制命令
    return []paint.DrawCmd{
        paint.NewTextCmd(x, y, buttonText, buttonStyle),
    }
}
```

---

## Flex布局实例分析

### 场景：3个按钮平均分配80宽度

```go
HStackBuilder(
    Button("Left").Flex(1),
    Button("Center").Flex(1),
    Button("Right").Flex(1),
).Gap(1).Build()
```

### Layout阶段计算

```
约束：{MinWidth: 0, MaxWidth: 80}

HStack.MeasureLayout({0, 80, ...}):
  1. 识别flex子节点：3个，每个 Flex=1
  2. 计算可用空间：
     availableWidth = 80 - 2*gap = 80 - 2 = 78
  3. 每个flex子节点分配：
     flexWidth = 78 / 3 = 26
  4. 设置子节点约束：
     ChildConstraints[0] = {26, 26, ...}  // "Left"
     ChildConstraints[1] = {26, 26, ...}  // "Center"
     ChildConstraints[2] = {26, 26, ...}  // "Right"
```

### Paint阶段渲染

```
Button "Left":
  contentWidth = 9     // " [ Left ]"
  layoutWidth = 26     // 由 SetBounds 设置
  availableSpace = 26 - 9 = 17
  textAlign = AlignStart
  → 渲染: " [ Left ]" + 17个空格

Button "Center":
  contentWidth = 11    // " [ Center ]"
  layoutWidth = 26
  availableSpace = 26 - 11 = 15
  textAlign = AlignCenter
  → 渲染: 7个空格 + " [ Center ]" + 8个空格

Button "Right":
  contentWidth = 10    // " [ Right ]"
  layoutWidth = 26
  availableSpace = 26 - 10 = 16
  textAlign = AlignEnd
  → 渲染: 16个空格 + " [ Right ]"
```

### 最终渲染结果

```
[ Left ]                 [ Center ]          [ Right
↑                         ↑                    ↑
0                        27                  54
+26                       +26                 +26
```

---

## 最近修复的问题

### 问题描述

按钮在 flex 布局中没有填充分配的宽度：
- Layout 阶段正确分配了宽度（26字符）
- 但 Paint 阶段没有使用这个宽度
- 文本对齐（左/中/右）不工作

### 根本原因

在 `Button.Paint()` 中，`naturalWidth` 的计算**错误地包含了用户指定的 padding**：

```go
// ❌ 错误的代码
naturalWidth := contentWidth + paddingLeft + paddingRight
```

这导致 padding 被计算了两次：
1. 第一次在 `naturalWidth` 计算时
2. 第二次在应用文本对齐时

结果是 `layoutWidth > naturalWidth` 判断失败，文本对齐逻辑从未执行。

### 解决方案

**修复 `naturalWidth` 计算**：只包含 `contentWidth`

```go
// ✅ 正确的代码
naturalWidth := contentWidth  // 不包含 padding
```

**修复 padding 应用**：在渲染时正确应用用户指定的 padding

```go
// AlignCenter: 左右两边都包含用户指定的 padding
leftSpace := paddingLeft + availableSpace/2
rightSpace := paddingRight + (availableSpace - availableSpace/2)
buttonText = strings.Repeat(" ", leftSpace) + buttonText +
             strings.Repeat(" ", rightSpace)

// AlignStart: 左边加 paddingLeft，右边加 paddingRight + availableSpace
buttonText = strings.Repeat(" ", paddingLeft) + buttonText +
             strings.Repeat(" ", paddingRight + availableSpace)

// AlignEnd: 左边加 paddingLeft + availableSpace，右边加 paddingRight
buttonText = strings.Repeat(" ", paddingLeft + availableSpace) + buttonText +
             strings.Repeat(" ", paddingRight)
```

### 修复结果

✅ **elegant_api_demo**: 按钮填充分配宽度，文本正确对齐
✅ **demo2**: 按钮平均分布在整个容器宽度
✅ **文本对齐**: AlignStart/AlignCenter/AlignEnd 正常工作

---

## 调试技巧

### 启用 Layout 调试

```bash
TUI_LAYOUT_DEBUG=true go run main.go
```

输出示例：
```
[HStack.MeasureLayout] flex child 0: flexWidth=26, size={26 1}
[Layout.Position] Button at (0,3,26x1)
```

### 启用 Paint 调试

```bash
TUI_PAINT_DEBUG=true go run main.go
```

输出示例：
```
[Paint.paintNode] ✅ Paintable: YES, calling Paint(0, 3)
[DEBUG-PAINT] label="Left", bounds=[0 3 26 1], x=0, y=3
[DEBUG-PAINT]   contentWidth=9, naturalWidth=9, layoutWidth=26, paddingLeft=1, paddingRight=2
```

### 启用完整渲染流程调试

```bash
TUI_DEBUG_RENDER=true go run main.go
```

### 常见调试命令

```bash
# 查看 Layout 阶段的 flex 计算
TUI_LAYOUT_DEBUG=true go run main.go 2>&1 | grep "flex"

# 查看 Paint 阶段的 bounds 信息
TUI_PAINT_DEBUG=true go run main.go 2>&1 | grep "DEBUG-PAINT"

# 查看完整的渲染流程
TUI_PIPELINE_DEBUG=true go run main.go 2>&1 | grep -E "(Layout|Paint)"
```

---

## 关键要点总结

1. **两阶段渲染**：
   - Layout 阶段计算位置和大小，存储在 `ComputedBox` 中
   - Paint 阶段使用计算出的位置进行渲染

2. **SetBounds 的作用**：
   - Layout 阶段调用 `SetBounds(x, y, width, height)` 存储布局结果
   - Paint 阶段通过 `bounds` 字段读取这些信息

3. **Flex 布局流程**：
   - `LayoutNode.MeasureLayout()` 计算每个 flex 子节点应分配的宽度
   - `calculatePositions()` 设置子节点的 bounds
   - `Button.Paint()` 读取 bounds 并应用文本对齐

4. **Natural vs Layout Width**：
   - `naturalWidth`: 组件的自然宽度（无约束时）
   - `layoutWidth`: 布局引擎分配的宽度（通过 bounds）
   - 当 `layoutWidth > naturalWidth` 时，应用文本对齐

5. **Padding 的正确处理**：
   - Padding 在渲染时应用，不包含在 naturalWidth 计算中
   - 避免 padding 被重复计算

---

## 相关文件

- `internal/render/rendering_pipeline.go`: 渲染流程入口
- `runtime/compute/engine.go`: Layout 阶段实现
- `internal/render/paint_engine.go`: Paint 阶段实现
- `runtime/ui/layout_measurement.go`: LayoutNode 测量逻辑
- `components/button/button.go`: Button 组件实现
- `runtime/ui/box_model.go`: BoxModel 接口和 Mixin

---

**文档版本**: v1.0
**最后更新**: 2025-02-07
**作者**: Claude Sonnet 4.5
