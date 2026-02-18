# Fiber-First 渲染管线实施指南

## 快速开始

本文档提供 Fiber-First 渲染管线的具体实施细节和代码示例。

---

## 1. 核心接口定义

### 1.1 paint.PaintableBox 接口

```go
package paint

// PaintableBox 是所有可渲染组件的核心接口
type PaintableBox interface {
    // Paint 在指定位置绘制组件
    // 返回绘制命令列表
    Paint(x, y int) []DrawCmd
    
    // GetSize 获取组件的固有大小（可选）
    GetSize() (width, height int)
}

// DrawCmd 表示一个绘制命令
type DrawCmd struct {
    Type     DrawCmdType
    X, Y     int
    Width    int
    Height   int
    Text     string
    Style    Style
    Children []DrawCmd
}

type DrawCmdType int

const (
    DrawText DrawCmdType = iota
    DrawRect
    DrawBorder
    DrawFill
)
```

### 1.2 LayoutResult 结构

```go
package layout

// LayoutResult 表示布局结果
type LayoutResult struct {
    Root *LayoutNode
}

// LayoutNode 表示单个节点的布局信息
type LayoutNode struct {
    Fiber    *Fiber              // 关联的 Fiber 节点
    Instance paint.PaintableBox  // 关联的实例
    Box      Box                 // 布局盒子
    Children []*LayoutNode       // 子节点
}

// Box 表示布局盒子
type Box struct {
    X, Y      int
    Width     int
    Height    int
    Direction Direction // 布局方向
}

type Direction int

const (
    DirectionHorizontal Direction = iota
    DirectionVertical
)
```

---

## 2. Fiber 结构实现

### 2.1 Fiber 定义

```go
package ui

import (
    "paint"
)

type Fiber struct {
    // Identity
    Type ElementType
    Key  string
    
    // Tree Structure
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber
    
    // Runtime Entity
    Instance paint.PaintableBox  // ✅ 核心字段
    
    // Layout Input
    Style Style
    Props MemoizedProps
    
    // Scheduling
    Flags Flags
    Lanes Lane
}

// ✅ 不再包含 VNode 字段
// ✅ 不再包含 LayoutBox 字段
```

### 2.2 Fiber 创建

```go
// CreateFiber 从 VNode 创建 Fiber
func CreateFiber(vnode VNode) *Fiber {
    fiber := &Fiber{
        Type:  vnode.Type(),
        Key:   vnode.Key(),
        Style: extractStyle(vnode),
        Props: extractProps(vnode),
    }
    
    // 创建 Instance
    if factory, ok := vnode.(InstanceFactory); ok {
        fiber.Instance = factory.CreateInstance()
    }
    
    return fiber
}

// CloneFiber 克隆 Fiber (复用 Instance)
func CloneFiber(old *Fiber) *Fiber {
    return &Fiber{
        Type:      old.Type,
        Key:       old.Key,
        Instance:  old.Instance,  // ✅ 复用，不克隆
        Style:     old.Style,
        Props:     old.Props,
        Alternate: old,
    }
}
```

---

## 3. Layout 引擎实现

### 3.1 Fiber-based Layout 引擎

```go
package layout

type FiberLayoutEngine struct {
    // 可配置的布局算法
}

// LayoutFiber 对 Fiber 树进行布局
func (e *FiberLayoutEngine) LayoutFiber(root *Fiber, constraints Constraints) *LayoutResult {
    layoutRoot := e.layoutNode(root, constraints)
    return &LayoutResult{Root: layoutRoot}
}

// layoutNode 递归布局单个节点
func (e *FiberLayoutEngine) layoutNode(fiber *Fiber, constraints Constraints) *LayoutNode {
    if fiber == nil {
        return nil
    }
    
    // 1. 创建布局节点
    node := &LayoutNode{
        Fiber:    fiber,
        Instance: fiber.Instance,
    }
    
    // 2. 计算当前节点大小
    node.Box = e.calculateBox(fiber.Style, constraints)
    
    // 3. 布局子节点
    if fiber.Child != nil {
        childConstraints := e.getChildConstraints(node.Box, fiber.Style)
        
        var prevChild *LayoutNode
        for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
            childNode := e.layoutNode(childFiber, childConstraints)
            
            if prevChild == nil {
                node.Children = []*LayoutNode{childNode}
            } else {
                node.Children = append(node.Children, childNode)
            }
            prevChild = childNode
            
            // 更新约束（根据布局方向）
            childConstraints = e.updateConstraints(childConstraints, childNode.Box, fiber.Style)
        }
    }
    
    return node
}

// calculateBox 计算节点的布局盒子
func (e *FiberLayoutEngine) calculateBox(style Style, constraints Constraints) Box {
    // 固有大小
    width := style.Width
    height := style.Height
    
    // Flex 布局
    if style.FlexGrow > 0 {
        if style.FlexDirection == DirectionHorizontal {
            width = constraints.MaxWidth * style.FlexGrow
        } else {
            height = constraints.MaxHeight * style.FlexGrow
        }
    }
    
    // 应用约束
    if width > constraints.MaxWidth {
        width = constraints.MaxWidth
    }
    if height > constraints.MaxHeight {
        height = constraints.MaxHeight
    }
    
    return Box{
        Width:     width,
        Height:    height,
        Direction: style.FlexDirection,
    }
}
```

---

## 4. Paint 引擎实现

### 4.1 PaintableBox 驱动的 Paint 引擎

```go
package paint

type PaintEngine struct {
    // 可配置的绘制策略
}

// PaintLayout 绘制布局结果
func (e *PaintEngine) PaintLayout(layoutResult *LayoutResult, buf *Buffer) {
    if layoutResult == nil || layoutResult.Root == nil {
        return
    }
    
    e.paintNode(layoutResult.Root, 0, 0, buf)
}

// paintNode 递归绘制节点
func (e *PaintEngine) paintNode(node *LayoutNode, offsetX, offsetY int, buf *Buffer) {
    if node == nil {
        return
    }
    
    // 计算绝对坐标
    absX := offsetX + node.Box.X
    absY := offsetY + node.Box.Y
    
    // 如果有实例，调用 Paint 方法
    if node.Instance != nil {
        drawCmds := node.Instance.Paint(absX, absY)
        
        // 执行绘制命令
        for _, cmd := range drawCmds {
            e.executeDrawCmd(cmd, buf)
        }
    }
    
    // 递归绘制子节点
    for _, child := range node.Children {
        e.paintNode(child, absX, absY, buf)
    }
}

// executeDrawCmd 执行单个绘制命令
func (e *PaintEngine) executeDrawCmd(cmd DrawCmd, buf *Buffer) {
    switch cmd.Type {
    case DrawText:
        buf.DrawText(cmd.Text, cmd.X, cmd.Y, cmd.Style)
    case DrawRect:
        buf.DrawRect(cmd.X, cmd.Y, cmd.Width, cmd.Height, cmd.Style)
    case DrawBorder:
        buf.DrawBorder(cmd.X, cmd.Y, cmd.Width, cmd.Height, cmd.Style)
    case DrawFill:
        buf.FillRect(cmd.X, cmd.Y, cmd.Width, cmd.Height, cmd.Style)
    }
}
```

---

## 5. 组件实现示例

### 5.1 Text 组件

```go
package components

import "paint"

// TextVNode - 描述层
type TextVNode struct {
    text  string
    style paint.Style
}

func (v *TextVNode) CreateInstance() paint.PaintableBox {
    return &TextInstance{
        text:  v.text,
        style: v.style,
    }
}

// TextInstance - 运行时层
type TextInstance struct {
    text  string
    style paint.Style
}

func (i *TextInstance) Paint(x, y int) []paint.DrawCmd {
    return []paint.DrawCmd{
        {
            Type:  paint.DrawText,
            X:     x,
            Y:     y,
            Text:  i.text,
            Style: i.style,
        },
    }
}

func (i *TextInstance) GetSize() (int, int) {
    // 返回文本的宽度和高度
    return len(i.text), 1
}
```

### 5.2 VStack 组件

```go
package components

import "paint"

// VStackVNode - 描述层
type VStackVNode struct {
    children []VNode
    spacing  int
    style    paint.Style
}

func (v *VStackVNode) CreateInstance() paint.PaintableBox {
    return &VStackInstance{
        spacing: v.spacing,
        style:   v.style,
    }
}

// VStackInstance - 运行时层
type VStackInstance struct {
    spacing int
    style   paint.Style
}

func (i *VStackInstance) Paint(x, y int) []paint.DrawCmd {
    // VStack 本身不绘制，只返回容器命令
    return []paint.DrawCmd{
        {
            Type:     paint.DrawFill,
            X:        x,
            Y:        y,
            Style:    i.style,
            Children: []paint.DrawCmd{}, // 子节点由 PaintEngine 处理
        },
    }
}

func (i *VStackInstance) GetSize() (int, int) {
    // 返回默认大小（实际大小由布局引擎计算）
    return 0, 0
}
```

### 5.3 Button 组件

```go
package components

import "paint"

// ButtonVNode - 描述层
type ButtonVNode struct {
    label    string
    variant  ButtonVariant
    onClick  func()
    disabled bool
}

func (v *ButtonVNode) CreateInstance() paint.PaintableBox {
    return &ButtonInstance{
        label:    v.label,
        variant:  v.variant,
        onClick:  v.onClick,
        disabled: v.disabled,
    }
}

// ButtonInstance - 运行时层
type ButtonInstance struct {
    label    string
    variant  ButtonVariant
    onClick  func()
    disabled bool
    
    // 运行时状态
    focused  bool
    hovered  bool
    pressed  bool
}

func (i *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
    // 根据状态确定样式
    style := i.getStyle()
    
    // 绘制背景
    bgCmd := paint.DrawCmd{
        Type:  paint.DrawFill,
        X:     x,
        Y:     y,
        Width: len(i.label) + 2, // padding
        Height: 1,
        Style: style,
    }
    
    // 绘制边框（如果有焦点）
    var borderCmd *paint.DrawCmd
    if i.focused {
        borderCmd = &paint.DrawCmd{
            Type:  paint.DrawBorder,
            X:     x,
            Y:     y,
            Width: len(i.label) + 2,
            Height: 1,
            Style: i.getFocusStyle(),
        }
    }
    
    // 绘制文本
    textCmd := paint.DrawCmd{
        Type:  paint.DrawText,
        X:     x + 1, // padding
        Y:     y,
        Text:  i.label,
        Style: style,
    }
    
    // 组合命令
    cmds := []paint.DrawCmd{bgCmd, textCmd}
    if borderCmd != nil {
        cmds = append(cmds, *borderCmd)
    }
    
    return cmds
}

func (i *ButtonInstance) getStyle() paint.Style {
    if i.disabled {
        return DisabledStyle
    }
    if i.pressed {
        return PressedStyle
    }
    if i.hovered {
        return HoveredStyle
    }
    return NormalStyle
}

func (i *ButtonInstance) GetSize() (int, int) {
    return len(i.label) + 2, 1
}

// HandleClick 处理点击事件
func (i *ButtonInstance) HandleClick() {
    if !i.disabled && i.onClick != nil {
        i.onClick()
    }
}
```

---

## 6. 渲染管线集成

### 6.1 DeclarativeNode 优化

```go
package render

import (
    "layout"
    "paint"
)

type DeclarativeNode struct {
    reconciler   *Reconciler
    layoutEngine *layout.FiberLayoutEngine
    paintEngine  *paint.PaintEngine
    fiberRoot    *Fiber
    
    // ✅ 不再需要 root VNode
    // root VNode
    
    useFiberFirst bool
}

func (n *DeclarativeNode) Paint(ctx PaintContext, buf *paint.Buffer) {
    if n.useFiberFirst {
        n.fiberFirstPaint(ctx, buf)
    } else {
        n.legacyPaint(ctx, buf)
    }
}

// fiberFirstPaint - 新的渲染路径
func (n *DeclarativeNode) fiberFirstPaint(ctx PaintContext, buf *paint.Buffer) {
    // Phase 1: Reconciliation
    // 生成临时 VNode → 更新 Fiber → 丢弃 VNode
    n.reconciler.Reconcile(func() VNode {
        return n.renderFn()
    })
    
    // Phase 2: Layout
    // Fiber → LayoutResult
    constraints := layout.Constraints{
        MaxWidth:  ctx.AvailableWidth,
        MaxHeight: ctx.AvailableHeight,
    }
    layoutResult := n.layoutEngine.LayoutFiber(n.fiberRoot, constraints)
    
    // Phase 3: Paint
    // LayoutResult → Buffer
    n.paintEngine.PaintLayout(layoutResult, buf)
}

// legacyPaint - 旧的渲染路径 (向后兼容)
func (n *DeclarativeNode) legacyPaint(ctx PaintContext, buf *paint.Buffer) {
    // 保持原有逻辑
    // ...
}
```

---

## 7. 测试策略

### 7.1 单元测试

```go
func TestTextInstancePaint(t *testing.T) {
    instance := &TextInstance{
        text:  "Hello",
        style: paint.DefaultStyle,
    }
    
    cmds := instance.Paint(10, 20)
    
    assert.Equal(t, 1, len(cmds))
    assert.Equal(t, paint.DrawText, cmds[0].Type)
    assert.Equal(t, 10, cmds[0].X)
    assert.Equal(t, 20, cmds[0].Y)
    assert.Equal(t, "Hello", cmds[0].Text)
}

func TestFiberLayoutEngine(t *testing.T) {
    // 创建测试 Fiber 树
    root := createTestFiberTree()
    
    engine := &layout.FiberLayoutEngine{}
    result := engine.LayoutFiber(root, layout.Constraints{
        MaxWidth:  80,
        MaxHeight: 24,
    })
    
    assert.NotNil(t, result)
    assert.NotNil(t, result.Root)
    // 更多断言...
}
```

### 7.2 集成测试

```go
func TestFullRenderingPipeline(t *testing.T) {
    // 创建测试应用
    app := createTestApp()
    
    // 执行渲染
    buf := paint.NewBuffer(80, 24)
    app.render(buf)
    
    // 验证输出
    assert.Contains(t, buf.String(), "Expected text")
}
```

---

## 8. 性能监控

### 8.1 关键指标

```go
type RenderingMetrics struct {
    ReconcileTime time.Duration
    LayoutTime    time.Duration
    PaintTime     time.Duration
    TotalTime     time.Duration
    
    FiberCount    int
    InstanceCount int
    DrawCmdCount  int
}

func (n *DeclarativeNode) PaintWithMetrics(ctx PaintContext, buf *paint.Buffer) *RenderingMetrics {
    metrics := &RenderingMetrics{}
    
    start := time.Now()
    
    // Phase 1: Reconcile
    reconcileStart := time.Now()
    n.reconciler.Reconcile(n.renderFn)
    metrics.ReconcileTime = time.Since(reconcileStart)
    
    // Phase 2: Layout
    layoutStart := time.Now()
    layoutResult := n.layoutEngine.LayoutFiber(n.fiberRoot, ctx.Constraints)
    metrics.LayoutTime = time.Since(layoutStart)
    
    // Phase 3: Paint
    paintStart := time.Now()
    n.paintEngine.PaintLayout(layoutResult, buf)
    metrics.PaintTime = time.Since(paintStart)
    
    metrics.TotalTime = time.Since(start)
    
    return metrics
}
```

---

## 9. 调试技巧

### 9.1 可视化工具

```go
// DumpLayoutResult 打印布局结果
func DumpLayoutResult(result *LayoutResult) string {
    return dumpNode(result.Root, 0)
}

func dumpNode(node *LayoutNode, indent int) string {
    prefix := strings.Repeat("  ", indent)
    
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%sFiber: %v\n", prefix, node.Fiber.Type))
    sb.WriteString(fmt.Sprintf("%sBox: X=%d Y=%d W=%d H=%d\n", 
        prefix, node.Box.X, node.Box.Y, node.Box.Width, node.Box.Height))
    
    if node.Instance != nil {
        sb.WriteString(fmt.Sprintf("%sInstance: %T\n", prefix, node.Instance))
    }
    
    for _, child := range node.Children {
        sb.WriteString(dumpNode(child, indent+1))
    }
    
    return sb.String()
}
```

---

## 10. 迁移检查清单

### 组件迁移检查

- [ ] 所有组件实现 `paint.PaintableBox` 接口
- [ ] 移除组件中的 VNode 运行时依赖
- [ ] 确保 Instance 状态正确管理
- [ ] 测试所有交互功能
- [ ] 验证视觉效果一致性

### 系统集成检查

- [ ] Fiber 树正确构建
- [ ] Layout 引擎输出正确
- [ ] Paint 引擎渲染正常
- [ ] 事件系统工作正常
- [ ] 性能指标符合预期

---

**下一步**: 参考 [FIBER_FIRST_RENDER_PIPELINE.md](./FIBER_FIRST_RENDER_PIPELINE.md) 了解完整的架构设计
