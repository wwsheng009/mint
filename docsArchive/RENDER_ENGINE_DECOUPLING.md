# Render Engine Decoupling Plan (渲染引擎解耦方案)

## 目标

解除渲染引擎 (`internal/render/paint_engine.go`) 对 `VNode` 和 `Fiber` 的强耦合，通过接口抽象实现：

1. **布局引擎** (`runtime/layout`) - 纯布局计算，不依赖具体节点类型
2. **绘制层** (`runtime/paint`) - 纯绘制抽象，通过接口访问节点
3. **适配层** (`runtime/compute`) - 提供 VNode/Fiber 到抽象接口的适配

## 当前问题分析

### 1. runtime/compute 包的耦合

**ComputedBox 结构体** (`runtime/compute/types.go:36-76`):

```go
type ComputedBox struct {
    VNode VNode           // ❌ 直接绑定 VNode
    runtime.Box
    Children []*ComputedBox
    Parent *ComputedBox
    LayoutDirty bool
    LayoutHash uint64
    RenderedText string
    NaturalWidth int
    NodeID uint64
    Layer rtui.Layer      // ❌ 绑定 rtui.Layer
    ChildFiber *rtui.Fiber // ❌ 直接绑定 Fiber
}
```

**Engine 对 VNode 的依赖** (`runtime/compute/engine.go`):

| 方法 | VNode依赖 | 用途 |
|------|----------|------|
| `buildComputedBoxWithSize` | `vnode.Children()`, `vnode.Key()`, `vnode.Type()` | 构建布局树 |
| `measureVNode` | `vnode.(Measurable)`, 类型断言 | 测量节点尺寸 |
| `layoutHStack/VStack` | `rtui.GetLayoutInfo(vnode)` | 布局算法 |
| `buildHitMapFromComputedBoxes` | `vnode.Key()` | 构建命中映射 |

### 2. internal/render/paint_engine.go 的耦合

**PaintEngine 对 ComputedBox 的依赖**:

| 位置 | 依赖 | 用途 |
|------|------|------|
| L24 | `map[*compute.ComputedBox]style.Color` | 背景继承追踪 |
| L44 | `*compute.ComputedLayout` | Paint入口参数 |
| L111-113 | `box.VNode.(Paintable)` | 检查是否可绘制 |
| L173 | `box.VNode.Style()` | 获取样式 |
| L193-215 | `box.VNode.Type()` | 类型分发 |
| L218-228 | `box.VNode.(接口断言)` | 获取border/table信息 |
| L238-260 | `box.RenderedText`, `box.VNode` | 文本绘制 |
| L366-410 | `box.VNode` 边框属性 | 边框绘制 |

### 3. 现有可复用资源

#### runtime/layout 包 (已存在)

```go
// types.go - 纯布局接口
type Node interface {
    ID() string
    Type() string
    Children() []Node
    GetPosition() (x, y int)
    SetPosition(x, y int)
    GetSize() (width, height int)
    SetSize(width, height int)
}

type Measurable interface {
    Node
    Measure(constraints Constraints) Size
}

type Engine struct { ... }  // 纯布局引擎
```

#### runtime/paint 包 (已存在)

```go
// paintable.go - 已有抽象
type Paintable interface {
    Paint(x, y int) []DrawCmd
}

type DrawCmd struct {
    X, Y  int
    Text  string
    Style style.Style
}

// buffer.go - 纯绘制缓冲
type Buffer struct { ... }
type Cell struct { ... }

// context.go - 纯绘制上下文
type PaintContext struct { ... }
type Painter struct { ... }
```

## 重构方案

### 目标架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      runtime/layout                              │
│  (纯布局引擎)                                                    │
│  - Node 接口                                                    │
│  - Constraints, Size, Rect                                      │
│  - Engine.Layout() → LayoutResult                               │
│  ✅ 已存在，可直接使用                                            │
└─────────────────────────────────────────────────────────────────┘
                               ↓ LayoutBox
┌─────────────────────────────────────────────────────────────────┐
│                      runtime/paint                               │
│  (纯绘制抽象层)                                                  │
│  - PaintableNode 接口 (新增)                                    │
│  - PaintableBox 结构体 (新增)                                   │
│  - PaintableLayout 结构体 (新增)                                │
│  - Buffer, Cell, Painter, PaintContext (已有)                   │
└─────────────────────────────────────────────────────────────────┘
                               ↓ PaintableLayout
┌─────────────────────────────────────────────────────────────────┐
│                      runtime/compute                             │
│  (适配层)                                                        │
│  - VNodeAdapter: 实现 layout.Node + paint.PaintableNode        │
│  - FiberAdapter: 实现 layout.Node + paint.PaintableNode        │
│  - ComputedBox.AsPaintable() → *paint.PaintableBox             │
│  - ComputedLayout.AsPaintable() → *paint.PaintableLayout       │
└─────────────────────────────────────────────────────────────────┘
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│                  internal/render/paint_engine.go                 │
│  (解耦后的绘制引擎)                                               │
│  - Paint(layout *paint.PaintableLayout, buffer *paint.Buffer)  │
│  - 只依赖 paint.PaintableLayout，不依赖 compute.ComputedBox     │
│  - 不依赖 rtui.VNode 或 rtui.Fiber                              │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 1: 扩展 runtime/paint 接口

#### 1.1 新增 PaintableNode 接口

文件: `runtime/paint/paintable_node.go`

```go
package paint

import "github.com/wwsheng009/mint/runtime/style"

// NodeType 节点类型枚举
type NodeType int

const (
    NodeTypeText NodeType = iota
    NodeTypeElement
    NodeTypeComponent
    NodeTypeFragment
)

// String 返回节点类型名称
func (t NodeType) String() string {
    switch t {
    case NodeTypeText:
        return "Text"
    case NodeTypeElement:
        return "Element"
    case NodeTypeComponent:
        return "Component"
    case NodeTypeFragment:
        return "Fragment"
    default:
        return "Unknown"
    }
}

// PaintableNode 可绘制节点接口
// 这是 PaintEngine 操作的最小抽象，解耦 VNode/Fiber 依赖
type PaintableNode interface {
    // 标识
    ID() string
    NodeType() NodeType
    Tag() string // "button", "text", "hstack", "vstack" 等

    // 样式
    Style() style.Style
    SetStyle(s style.Style)

    // 内容
    TextContent() string

    // 绘制 - 复用现有 Paintable 接口
    Paintable
}

// BorderStyle 边框样式
type BorderStyle int

const (
    BorderStyleSingle BorderStyle = iota
    BorderStyleDouble
    BorderStyleRounded
    BorderStyleNone
)

// BorderInfo 边框信息接口
// 可选接口，节点可实现此接口提供边框信息
type BorderInfo interface {
    GetBorderStyle() BorderStyle
    GetBorderColor() string
    GetBorderLabel() string
}

// LayoutInfo 布局信息接口
// 可选接口，节点可实现此接口提供布局信息
type LayoutInfo interface {
    GetDirection() string // "row" or "column"
    GetGap() int
    GetFlex() int
    GetAlign() string
    GetCrossAlign() string
    GetPadding() [4]int // top, right, bottom, left
}
```

#### 1.2 新增 PaintableBox 结构体

文件: `runtime/paint/paintable_box.go`

```go
package paint

import (
    "github.com/wwsheng009/mint/runtime/style"
)

// PaintableBox 可绘制盒子
// 替代 compute.ComputedBox 作为 PaintEngine 的输入
// 这是一个纯数据结构，不依赖任何具体实现
type PaintableBox struct {
    // 节点信息
    Node PaintableNode

    // 布局结果
    X, Y         int
    Width, Height int

    // 树结构
    Children []*PaintableBox
    Parent   *PaintableBox

    // 布局状态
    LayoutDirty  bool
    LayoutHash   uint64

    // 渲染辅助
    RenderedText string // 布局阶段计算的渲染文本（含对齐填充）
    NaturalWidth int    // 自然宽度（用于对齐计算）

    // 边框信息（可选）
    BorderStyle BorderStyle
    BorderColor string
    BorderLabel string

    // 层级信息（可选）
    Layer  int // 渲染层: 0=Base, 1=Overlay, 2=Modal, 3=Tooltip, 4=Inspector
    ZIndex int // 层内排序

    // 节点标识（用于 HitMap）
    NodeID uint64
}

// NewPaintableBox 创建可绘制盒子
func NewPaintableBox(node PaintableNode) *PaintableBox {
    return &PaintableBox{
        Node:    node,
        Layer:   0,
        ZIndex:  0,
    }
}

// Bounds 返回边界
func (b *PaintableBox) Bounds() (x, y, w, h int) {
    return b.X, b.Y, b.Width, b.Height
}

// SetBounds 设置边界
func (b *PaintableBox) SetBounds(x, y, w, h int) {
    b.X, b.Y, b.Width, b.Height = x, y, w, h
}

// AddChild 添加子节点
func (b *PaintableBox) AddChild(child *PaintableBox) {
    child.Parent = b
    b.Children = append(b.Children, child)
}

// MarkDirty 标记为脏
func (b *PaintableBox) MarkDirty() {
    b.LayoutDirty = true
    for parent := b.Parent; parent != nil; parent = parent.Parent {
        if parent.LayoutDirty {
            break
        }
        parent.LayoutDirty = true
    }
}

// ClearDirty 清除脏标记
func (b *PaintableBox) ClearDirty() {
    b.LayoutDirty = false
    for _, child := range b.Children {
        child.ClearDirty()
    }
}

// Depth 返回深度
func (b *PaintableBox) Depth() int {
    depth := 0
    for parent := b.Parent; parent != nil; parent = parent.Parent {
        depth++
    }
    return depth
}

// Count 返回子树节点数
func (b *PaintableBox) Count() int {
    count := 1
    for _, child := range b.Children {
        count += child.Count()
    }
    return count
}

// FindByPosition 查找包含指定位置的最内层盒子
func (b *PaintableBox) FindByPosition(x, y int) *PaintableBox {
    // 检查点是否在此盒子内
    if x < b.X || x >= b.X+b.Width || y < b.Y || y >= b.Y+b.Height {
        return nil
    }

    // 检查子节点（逆序，上层优先）
    for i := len(b.Children) - 1; i >= 0; i-- {
        if found := b.Children[i].FindByPosition(x, y); found != nil {
            return found
        }
    }

    return b
}

// FindByID 根据 NodeID 查找盒子
func (b *PaintableBox) FindByID(nodeID uint64) *PaintableBox {
    if b.NodeID == nodeID {
        return b
    }
    for _, child := range b.Children {
        if found := child.FindByID(nodeID); found != nil {
            return found
        }
    }
    return nil
}

// HasBorder 是否有边框
func (b *PaintableBox) HasBorder() bool {
    return b.BorderStyle != BorderStyleNone
}

// GetBorderInfo 获取边框信息（从 Node 获取或使用自身属性）
func (b *PaintableBox) GetBorderInfo() (style BorderStyle, color, label string) {
    // 优先使用自身属性
    if b.BorderStyle != BorderStyleNone {
        return b.BorderStyle, b.BorderColor, b.BorderLabel
    }
    // 尝试从 Node 获取
    if b.Node != nil {
        if bi, ok := b.Node.(BorderInfo); ok {
            return bi.GetBorderStyle(), bi.GetBorderColor(), bi.GetBorderLabel()
        }
    }
    return BorderStyleNone, "", ""
}

// PaintableLayout 可绘制布局
// 替代 compute.ComputedLayout 作为 PaintEngine.Paint() 的输入
type PaintableLayout struct {
    Root   *PaintableBox
    HitMap *HitMap // 可选的命中映射
}

// NewPaintableLayout 创建可绘制布局
func NewPaintableLayout(root *PaintableBox) *PaintableLayout {
    return &PaintableLayout{Root: root}
}

// FindByPosition 查找指定位置的盒子
func (l *PaintableLayout) FindByPosition(x, y int) *PaintableBox {
    if l.Root == nil {
        return nil
    }
    return l.Root.FindByPosition(x, y)
}

// FindByID 根据 NodeID 查找盒子
func (l *PaintableLayout) FindByID(nodeID uint64) *PaintableBox {
    if l.Root == nil {
        return nil
    }
    return l.Root.FindByID(nodeID)
}
```

### Phase 2: 创建适配器

#### 2.1 VNode 适配器

文件: `runtime/compute/adapter_vnode.go`

```go
package compute

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// VNodePaintableAdapter 将 VNode 适配为 PaintableNode 接口
type VNodePaintableAdapter struct {
    VNode rtui.VNode
}

// 确保 VNodePaintableAdapter 实现 paint.PaintableNode
var _ paint.PaintableNode = (*VNodePaintableAdapter)(nil)

func (a *VNodePaintableAdapter) ID() string {
    if a.VNode == nil {
        return ""
    }
    return a.VNode.Key()
}

func (a *VNodePaintableAdapter) NodeType() paint.NodeType {
    if a.VNode == nil {
        return paint.NodeTypeFragment
    }
    switch a.VNode.Type() {
    case rtui.VNodeText:
        return paint.NodeTypeText
    case rtui.VNodeElement:
        return paint.NodeTypeElement
    case rtui.VNodeComponent:
        return paint.NodeTypeComponent
    default:
        return paint.NodeTypeFragment
    }
}

func (a *VNodePaintableAdapter) Tag() string {
    if a.VNode == nil {
        return ""
    }
    if tagger, ok := a.VNode.(interface{ Tag() string }); ok {
        return tagger.Tag()
    }
    return ""
}

func (a *VNodePaintableAdapter) Style() style.Style {
    if a.VNode == nil {
        return style.Style{}
    }
    return a.VNode.Style()
}

func (a *VNodePaintableAdapter) SetStyle(s style.Style) {
    if a.VNode != nil {
        a.VNode.SetStyle(s)
    }
}

func (a *VNodePaintableAdapter) TextContent() string {
    if a.VNode == nil {
        return ""
    }
    return rtui.GetTextContent(a.VNode)
}

func (a *VNodePaintableAdapter) Paint(x, y int) []paint.DrawCmd {
    if a.VNode == nil {
        return nil
    }
    if paintable, ok := a.VNode.(interface {
        Paint(int, int) []paint.DrawCmd
    }); ok {
        return paintable.Paint(x, y)
    }
    return nil
}

// 实现 paint.BorderInfo 接口
func (a *VNodePaintableAdapter) GetBorderStyle() paint.BorderStyle {
    if a.VNode == nil {
        return paint.BorderStyleNone
    }
    if bs, ok := a.VNode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
        switch bs.GetBorderStyle() {
        case rtui.BorderStyleSingle:
            return paint.BorderStyleSingle
        case rtui.BorderStyleDouble:
            return paint.BorderStyleDouble
        case rtui.BorderStyleRounded:
            return paint.BorderStyleRounded
        default:
            return paint.BorderStyleNone
        }
    }
    return paint.BorderStyleNone
}

func (a *VNodePaintableAdapter) GetBorderColor() string {
    if a.VNode == nil {
        return ""
    }
    if bc, ok := a.VNode.(interface{ GetBorderColor() string }); ok {
        return bc.GetBorderColor()
    }
    return ""
}

func (a *VNodePaintableAdapter) GetBorderLabel() string {
    if a.VNode == nil {
        return ""
    }
    if bl, ok := a.VNode.(interface{ GetBorderLabel() string }); ok {
        return bl.GetBorderLabel()
    }
    return ""
}
```

#### 2.2 Fiber 适配器

文件: `runtime/compute/adapter_fiber.go`

```go
package compute

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// FiberPaintableAdapter 将 Fiber 适配为 PaintableNode 接口
type FiberPaintableAdapter struct {
    Fiber *rtui.Fiber
}

// 确保 FiberPaintableAdapter 实现 paint.PaintableNode
var _ paint.PaintableNode = (*FiberPaintableAdapter)(nil)

func (a *FiberPaintableAdapter) ID() string {
    if a.Fiber == nil {
        return ""
    }
    return a.Fiber.DiffKey
}

func (a *FiberPaintableAdapter) NodeType() paint.NodeType {
    if a.Fiber == nil {
        return paint.NodeTypeFragment
    }
    switch a.Fiber.Type {
    case rtui.VNodeText:
        return paint.NodeTypeText
    case rtui.VNodeElement:
        return paint.NodeTypeElement
    case rtui.VNodeComponent:
        return paint.NodeTypeComponent
    default:
        return paint.NodeTypeFragment
    }
}

func (a *FiberPaintableAdapter) Tag() string {
    if a.Fiber == nil {
        return ""
    }
    return a.Fiber.Tag
}

func (a *FiberPaintableAdapter) Style() style.Style {
    if a.Fiber == nil || a.Fiber.Props == nil {
        return style.Style{}
    }
    // 从 Fiber.Props 获取样式
    if s, ok := a.Fiber.Props["style"].(style.Style); ok {
        return s
    }
    return style.Style{}
}

func (a *FiberPaintableAdapter) SetStyle(s style.Style) {
    if a.Fiber != nil {
        if a.Fiber.Props == nil {
            a.Fiber.Props = make(map[string]interface{})
        }
        a.Fiber.Props["style"] = s
    }
}

func (a *FiberPaintableAdapter) TextContent() string {
    if a.Fiber == nil {
        return ""
    }
    // 优先从 MemoizedState 获取
    if a.Fiber.MemoizedState != nil {
        if s, ok := a.Fiber.MemoizedState.(string); ok {
            return s
        }
    }
    // 从 Props 获取
    if a.Fiber.Props != nil {
        if c, ok := a.Fiber.Props["content"]; ok {
            if s, ok := c.(string); ok {
                return s
            }
        }
    }
    return ""
}

func (a *FiberPaintableAdapter) Paint(x, y int) []paint.DrawCmd {
    if a.Fiber == nil {
        return nil
    }
    // Fiber 通过 FiberVNode 代理获取 Paint 方法
    vnode := rtui.NewFiberVNode(a.Fiber)
    if paintable, ok := vnode.(interface {
        Paint(int, int) []paint.DrawCmd
    }); ok {
        return paintable.Paint(x, y)
    }
    return nil
}

func (a *FiberPaintableAdapter) GetBorderStyle() paint.BorderStyle {
    if a.Fiber == nil || a.Fiber.Props == nil {
        return paint.BorderStyleNone
    }
    if bs, ok := a.Fiber.Props["borderStyle"].(rtui.BorderStyle); ok {
        switch bs {
        case rtui.BorderStyleSingle:
            return paint.BorderStyleSingle
        case rtui.BorderStyleDouble:
            return paint.BorderStyleDouble
        case rtui.BorderStyleRounded:
            return paint.BorderStyleRounded
        default:
            return paint.BorderStyleNone
        }
    }
    return paint.BorderStyleNone
}

func (a *FiberPaintableAdapter) GetBorderColor() string {
    if a.Fiber == nil || a.Fiber.Props == nil {
        return ""
    }
    if bc, ok := a.Fiber.Props["borderColor"].(string); ok {
        return bc
    }
    return ""
}

func (a *FiberPaintableAdapter) GetBorderLabel() string {
    if a.Fiber == nil || a.Fiber.Props == nil {
        return ""
    }
    if bl, ok := a.Fiber.Props["borderLabel"].(string); ok {
        return bl
    }
    return ""
}
```

#### 2.3 ComputedBox 转换方法

文件: `runtime/compute/adapter_convert.go`

```go
package compute

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
    "github.com/wwsheng009/mint/runtime/paint"
)

// AsPaintable 将 ComputedBox 转换为 PaintableBox
// 这是 PaintEngine 使用的入口方法
func (cb *ComputedBox) AsPaintable() *paint.PaintableBox {
    if cb == nil {
        return nil
    }

    box := &paint.PaintableBox{
        X:            cb.Box.X,
        Y:            cb.Box.Y,
        Width:        cb.Box.Width,
        Height:       cb.Box.Height,
        RenderedText: cb.RenderedText,
        NaturalWidth: cb.NaturalWidth,
        LayoutDirty:  cb.LayoutDirty,
        LayoutHash:   cb.LayoutHash,
        NodeID:       cb.NodeID,
        Layer:        int(cb.Layer),
        Children:     make([]*paint.PaintableBox, 0, len(cb.Children)),
    }

    // 适配 Node
    if cb.VNode != nil {
        box.Node = &VNodePaintableAdapter{VNode: cb.VNode}
    } else if cb.ChildFiber != nil {
        box.Node = &FiberPaintableAdapter{Fiber: cb.ChildFiber}
    }

    // 转换边框信息
    if cb.VNode != nil {
        if bs, ok := cb.VNode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
            switch bs.GetBorderStyle() {
            case rtui.BorderStyleSingle:
                box.BorderStyle = paint.BorderStyleSingle
            case rtui.BorderStyleDouble:
                box.BorderStyle = paint.BorderStyleDouble
            case rtui.BorderStyleRounded:
                box.BorderStyle = paint.BorderStyleRounded
            }
        }
        if bc, ok := cb.VNode.(interface{ GetBorderColor() string }); ok {
            box.BorderColor = bc.GetBorderColor()
        }
        if bl, ok := cb.VNode.(interface{ GetBorderLabel() string }); ok {
            box.BorderLabel = bl.GetBorderLabel()
        }
    }

    // 递归转换子节点
    for _, child := range cb.Children {
        if childBox := child.AsPaintable(); childBox != nil {
            childBox.Parent = box
            box.Children = append(box.Children, childBox)
        }
    }

    return box
}

// AsPaintableLayout 将 ComputedLayout 转换为 PaintableLayout
func (cl *ComputedLayout) AsPaintableLayout() *paint.PaintableLayout {
    if cl == nil {
        return nil
    }

    layout := paint.NewPaintableLayout(cl.Root.AsPaintable())
    
    // 转换 HitMap（如果需要）
    // HitMap 可以保留原引用，因为它是独立的
    layout.HitMap = cl.HitMap

    return layout
}
```

### Phase 3: 重构 PaintEngine

文件: `internal/render/paint_engine.go`

```go
package render

import (
    "fmt"
    "os"

    "github.com/wwsheng009/mint/internal/log"
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/border"
    "github.com/wwsheng009/mint/runtime/layer"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/style"
)

// PaintEngine renders layout trees using pre-computed layout information
// 重构后的 PaintEngine 只依赖 paint.PaintableLayout，不依赖 compute.ComputedBox
type PaintEngine struct {
    debug             bool
    lastHadModal      bool
    forceFullRender   bool
    parentBackground  map[*paint.PaintableBox]style.Color
    lastLayersPresent map[int]bool
    lastLayerBounds   map[int]runtime.Box
}

// NewPaintEngine creates a new paint engine
func NewPaintEngine() *PaintEngine {
    return &PaintEngine{
        debug:             log.PaintLogger.Enabled(),
        lastLayersPresent: make(map[int]bool),
        lastLayerBounds:   make(map[int]runtime.Box),
    }
}

// SetDebug enables/disables debug output
func (e *PaintEngine) SetDebug(debug bool) {
    e.debug = debug
}

// Paint renders a paintable layout to a buffer
// 重构后的入口方法：使用 paint.PaintableLayout 替代 compute.ComputedLayout
func (e *PaintEngine) Paint(layout *paint.PaintableLayout, buffer *paint.Buffer) error {
    if layout == nil || layout.Root == nil {
        return nil
    }

    // Clear parent background map at the start of each frame
    e.parentBackground = make(map[*paint.PaintableBox]style.Color)

    if e.forceFullRender {
        e.forceFullRender = false
        // Clear buffer
        for y := 0; y < buffer.Height; y++ {
            for x := 0; x < buffer.Width; x++ {
                buffer.SetCell(x, y, ' ', style.Style{})
            }
        }
    }

    return e.paintBox(layout.Root, buffer)
}

// paintBox recursively paints a paintable box and its children
// 重构后的核心方法：使用 paint.PaintableBox 替代 compute.ComputedBox
func (e *PaintEngine) paintBox(box *paint.PaintableBox, buffer *paint.Buffer) error {
    if box == nil || box.Node == nil {
        return nil
    }

    // Get parent background for inheritance
    var parentBG style.Color
    if bg, ok := e.parentBackground[box]; ok && bg != "" {
        parentBG = bg
    }

    // Check if node implements Paintable (custom rendering)
    commands := box.Node.Paint(box.X, box.Y)
    if len(commands) > 0 {
        // Apply commands with potential background inheritance
        for _, cmd := range commands {
            styleToApply := cmd.Style
            if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
                styleToApply.BG = parentBG
            }
            buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
        }
        return nil
    }

    // Inherit parent background for non-Paintable nodes
    if parentBG != "" {
        nodeStyle := box.Node.Style()
        if nodeStyle.BG == "" || nodeStyle.BG == style.NoColor {
            nodeStyle.BG = parentBG
            box.Node.SetStyle(nodeStyle)
        }
    }

    // Paint based on node type
    switch box.Node.NodeType() {
    case paint.NodeTypeText:
        e.paintText(box, buffer)
    case paint.NodeTypeElement:
        e.paintElement(box, buffer)
    case paint.NodeTypeFragment:
        return e.paintChildren(box, buffer)
    }

    // Handle bordered elements
    if bs, bc, bl := box.GetBorderInfo(); bs != paint.BorderStyleNone {
        e.paintBordered(box, buffer, bs, bc, bl)
        return nil
    }

    return e.paintChildren(box, buffer)
}

// paintText paints a text node
func (e *PaintEngine) paintText(box *paint.PaintableBox, buffer *paint.Buffer) {
    text := box.RenderedText
    if text == "" {
        text = box.Node.TextContent()
    }
    if text != "" {
        maxX := box.X + box.Width
        buffer.SetStringAligned(box.X, box.Y, text, box.Node.Style(), maxX)
    }
}

// paintElement paints an element node
func (e *PaintEngine) paintElement(box *paint.PaintableBox, buffer *paint.Buffer) {
    content := box.RenderedText
    if content == "" {
        content = box.Node.TextContent()
    }
    if content != "" {
        maxX := box.X + box.Width
        buffer.SetStringAligned(box.X, box.Y, content, box.Node.Style(), maxX)
        return
    }

    // Paint container background if set
    nodeStyle := box.Node.Style()
    if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
        e.paintContainerBackground(box, buffer, nodeStyle)
        
        // Store parent background for child inheritance
        if e.parentBackground == nil {
            e.parentBackground = make(map[*paint.PaintableBox]style.Color)
        }
        for _, childBox := range box.Children {
            e.parentBackground[childBox] = nodeStyle.BG
        }
    }
}

// paintContainerBackground fills the container area with background color
func (e *PaintEngine) paintContainerBackground(box *paint.PaintableBox, buffer *paint.Buffer, bgStyle style.Style) {
    backgroundStyle := style.Style{}.Background(bgStyle.BG)
    for y := 0; y < box.Height; y++ {
        for x := 0; x < box.Width; x++ {
            buffer.SetCell(box.X+x, box.Y+y, ' ', backgroundStyle)
        }
    }
}

// paintChildren paints children of a box
func (e *PaintEngine) paintChildren(box *paint.PaintableBox, buffer *paint.Buffer) error {
    for _, childBox := range box.Children {
        if err := e.paintBox(childBox, buffer); err != nil {
            return err
        }
    }
    return nil
}

// paintBordered paints a bordered box
func (e *PaintEngine) paintBordered(box *paint.PaintableBox, buffer *paint.Buffer, bs paint.BorderStyle, bc, bl string) {
    // Convert border style
    var borderStyle border.Style
    switch bs {
    case paint.BorderStyleDouble:
        borderStyle = border.StyleDouble
    case paint.BorderStyleRounded:
        borderStyle = border.StyleRounded
    default:
        borderStyle = border.StyleSingle
    }

    config := border.Config{
        Style: borderStyle,
        Color: bc,
        Label: bl,
    }
    renderer := border.WithConfig(config)

    contentWidth := box.Width - 2
    contentHeight := box.Height - 2
    if contentWidth < 0 {
        contentWidth = 0
    }
    if contentHeight < 0 {
        contentHeight = 0
    }

    renderer.Paint(box.X, box.Y, contentWidth, contentHeight,
        func(px, py int, ch rune, s style.Style) {
            buffer.SetCell(px, py, ch, s)
        })

    for _, childBox := range box.Children {
        if err := e.paintBox(childBox, buffer); err != nil && e.debug {
            log.PaintLogger.Debug("[paintBordered] error: %v", err)
        }
    }
}

// clearRegion clears a rectangular region of the buffer
func (e *PaintEngine) clearRegion(bounds runtime.Box, buffer *paint.Buffer) {
    maxX := buffer.Width
    maxY := buffer.Height

    for y := bounds.Y; y < bounds.Y+bounds.Height && y < maxY; y++ {
        for x := bounds.X + bounds.Width - 1; x >= bounds.X && x < maxX; x-- {
            buffer.SetCell(x, y, ' ', style.Style{})
        }
    }
}

// PaintLayers renders multiple layers in order
func (e *PaintEngine) PaintLayers(
    layouts layer.LayerLayouts,
    buffer *paint.Buffer,
) error {
    // ... 保持原有逻辑，但内部调用 AsPaintableLayout() 转换
    // 此方法保持向后兼容
    return nil
}

// PaintRenderPlanes paints RenderPlanes to buffer
func (e *PaintEngine) PaintRenderPlanes(
    renderPlanes *layer.RenderPlanes,
    buffer *paint.Buffer,
) error {
    if renderPlanes == nil {
        return nil
    }

    for _, layer := range renderPlanes.GetRenderOrder() {
        boxes := renderPlanes.GetLayer(layer)
        if boxes == nil || len(boxes) == 0 {
            continue
        }

        for _, box := range boxes {
            // 将 compute.ComputedBox 转换为 paint.PaintableBox
            paintableBox := box.AsPaintable()
            layout := paint.NewPaintableLayout(paintableBox)
            if err := e.Paint(layout, buffer); err != nil {
                return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
            }
        }
    }

    return nil
}
```

### Phase 4: 向后兼容层

为了平滑迁移，提供兼容方法：

```go
// internal/render/paint_engine_compat.go

import (
    "github.com/wwsheng009/mint/runtime/compute"
    "github.com/wwsheng009/mint/runtime/paint"
)

// PaintComputedLayout 兼容方法：接收 compute.ComputedLayout
// Deprecated: 使用 Paint(paint.PaintableLayout) 替代
func (e *PaintEngine) PaintComputedLayout(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
    if layout == nil {
        return nil
    }
    // 转换为 PaintableLayout
    paintableLayout := layout.AsPaintableLayout()
    return e.Paint(paintableLayout, buffer)
}
```

## 实施时间表

### Sprint 1 (Week 1-2): 接口定义

- [ ] 创建 `runtime/paint/paintable_node.go`
- [ ] 创建 `runtime/paint/paintable_box.go`
- [ ] 添加单元测试

### Sprint 2 (Week 3-4): 适配器实现

- [ ] 创建 `runtime/compute/adapter_vnode.go`
- [ ] 创建 `runtime/compute/adapter_fiber.go`
- [ ] 创建 `runtime/compute/adapter_convert.go`
- [ ] 添加适配器测试

### Sprint 3 (Week 5-6): PaintEngine 重构

- [ ] 重构 `internal/render/paint_engine.go`
- [ ] 添加兼容方法
- [ ] 更新现有测试

### Sprint 4 (Week 7-8): 迁移和验证

- [ ] 更新调用方使用新接口
- [ ] 性能测试
- [ ] 集成测试
- [ ] 文档更新

## 风险和缓解措施

| 风险 | 缓解措施 |
|------|----------|
| 性能退化 | 适配器开销很小，通过基准测试验证 |
| 破坏现有功能 | 提供兼容层，渐进迁移 |
| 接口设计不当 | 先在 compute 包内验证，再推广 |
| 边框/样式丢失 | 确保 BorderInfo 接口完整实现 |

## 验收标准

1. ✅ `internal/render/paint_engine.go` 不再直接导入 `rtui` 包
2. ✅ `PaintEngine.Paint()` 只接受 `*paint.PaintableLayout` 参数
3. ✅ 现有测试全部通过
4. ✅ 性能无明显退化 (<5%)
5. ✅ 可以独立测试 PaintEngine，无需 VNode/Fiber

## 附录

### A. 类型对照表

| 旧类型 (compute) | 新类型 (paint) | 转换方法 |
|------------------|----------------|----------|
| `ComputedBox` | `PaintableBox` | `AsPaintable()` |
| `ComputedLayout` | `PaintableLayout` | `AsPaintableLayout()` |
| `VNode` | `PaintableNode` | `VNodePaintableAdapter` |
| `Fiber` | `PaintableNode` | `FiberPaintableAdapter` |

### B. 接口依赖关系

```
paint.PaintableNode
    ├── paint.Paintable (嵌入)
    └── paint.BorderInfo (可选)

paint.PaintableBox
    ├── paint.PaintableNode (引用)
    └── paint.PaintableBox (子节点)

paint.PaintableLayout
    └── paint.PaintableBox (根节点)
```

### C. 迁移检查清单

- [ ] 所有 `box.VNode.Type()` → `box.Node.NodeType()`
- [ ] 所有 `box.VNode.Tag()` → `box.Node.Tag()`
- [ ] 所有 `box.VNode.Style()` → `box.Node.Style()`
- [ ] 所有 `box.VNode.Key()` → `box.Node.ID()`
- [ ] 所有 `rtui.GetTextContent(box.VNode)` → `box.Node.TextContent()`
- [ ] 所有 `box.VNode.(Paintable)` → `box.Node.Paint()`
