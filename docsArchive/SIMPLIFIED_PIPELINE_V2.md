# 简化渲染管线设计

## 设计原则

1. **职责分离** - 布局只管布局，绘制只管绘制
2. **最少转换** - 减少中间数据结构
3. **避免循环依赖** - 各包之间依赖清晰

## 当前问题

```
Fiber → LayoutBox → ComputedBox → PaintableBox → Paint
        (layout)    (混合??)      (paint)
        
问题：
- ComputedBox 职责不清
- 3次转换，4种数据结构
- LayoutBox → ComputedBox 转换逻辑复杂
```

## 目标架构

```
Fiber → LayoutBox → PaintableBox → Paint
        (layout)    (paint)
        
优势：
- 职责清晰
- 1次转换，3种数据结构
- 删除 ComputedBox 中间层
```

## 数据结构定义

### 1. LayoutBox (runtime/layout) - 纯布局

```go
// runtime/layout/types.go

// LayoutBox 布局结果 - 只包含位置和尺寸
// 不包含任何与绘制相关的信息
type LayoutBox struct {
    // === 身份标识 ===
    ID string  // 用于匹配 Fiber (使用 NodeID 或 DiffKey)
    
    // === 位置和尺寸 ===
    X, Y     int
    Width    int
    Height   int
    Baseline int
    
    // === 渲染层 ===
    Layer  Layer
    ZIndex int
    
    // === 树结构 ===
    Children []*LayoutBox
    Parent   *LayoutBox
}

// Layer 渲染层
type Layer int

const (
    LayerBase Layer = iota
    LayerDropdown
    LayerModal
    LayerTooltip
    // ...
)
```

### 2. PaintableBox (runtime/paint) - 纯绘制

```go
// runtime/paint/paintable_box.go

// PaintableBox 绘制数据 - 包含绘制所需的所有信息
// 不包含布局计算逻辑
type PaintableBox struct {
    // === 位置和尺寸 (从 LayoutBox 复制) ===
    X, Y     int
    Width    int
    Height   int
    
    // === 身份标识 (从 Fiber 复制) ===
    NodeID  uint64
    DiffKey string
    
    // === 绘制层 (从 LayoutBox 复制) ===
    Layer  rtui.Layer
    ZIndex int
    
    // === 绘制信息 (从 Fiber 复制) ===
    Node         PaintableNode  // 绘制接口
    RenderedText string         // 渲染后的文本
    BorderStyle  BorderStyle    // 边框样式
    BorderColor  string         // 边框颜色
    BorderLabel  string         // 边框标签
    
    // === 树结构 ===
    Children []*PaintableBox
    Parent   *PaintableBox
}

// PaintableNode 绘制接口
type PaintableNode interface {
    ID() string
    Tag() string
    Style() style.Style
    SetStyle(style.Style)
    TextContent() string
    Paint(x, y int) []DrawCmd
}
```

### 3. FiberToPaintableConverter (internal/render) - 转换器

```go
// internal/render/converter.go

// FiberToPaintableConverter 将 Fiber + LayoutBox 转换为 PaintableBox
type FiberToPaintableConverter struct {
    fiberMap map[string]*reconciler.Fiber  // DiffKey -> Fiber
}

// Convert 将 LayoutBox 树转换为 PaintableBox 树
func (c *FiberToPaintableConverter) Convert(
    lbox *layout.LayoutBox,
    parent *paint.PaintableBox,
) *paint.PaintableBox {
    if lbox == nil {
        return nil
    }
    
    // 创建 PaintableBox
    pbox := &paint.PaintableBox{
        X:       lbox.X,
        Y:       lbox.Y,
        Width:   lbox.Width,
        Height:  lbox.Height,
        Layer:   convertLayer(lbox.Layer),
        ZIndex:  lbox.ZIndex,
        Parent:  parent,
        Children: make([]*paint.PaintableBox, 0),
    }
    
    // 查找匹配的 Fiber 并填充绘制信息
    if fiber := c.findFiber(lbox.ID); fiber != nil {
        pbox.NodeID = fiber.NodeID
        pbox.DiffKey = fiber.DiffKey
        pbox.Node = NewFiberPaintableNode(fiber)
        
        // 从 Fiber 复制边框信息
        if fiber.Props != nil {
            pbox.BorderStyle = getBorderStyle(fiber)
            pbox.BorderColor = getBorderColor(fiber)
            pbox.BorderLabel = getBorderLabel(fiber)
        }
    }
    
    // 递归转换子节点
    for _, childLBox := range lbox.Children {
        childPBox := c.Convert(childLBox, pbox)
        if childPBox != nil {
            pbox.Children = append(pbox.Children, childPBox)
        }
    }
    
    return pbox
}

// findFiber 通过 ID 查找 Fiber
// 支持 NodeID 格式和 DiffKey 格式
func (c *FiberToPaintableConverter) findFiber(id string) *reconciler.Fiber {
    // 先尝试直接匹配 (DiffKey)
    if f, ok := c.fiberMap[id]; ok {
        return f
    }
    // 再尝试 NodeID 匹配
    for _, f := range c.fiberMap {
        if fmt.Sprintf("%d", f.NodeID) == id {
            return f
        }
    }
    return nil
}
```

## 渲染流程

```
┌─────────────────────────────────────────────────────────────┐
│                        渲染管线                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. 获取 Fiber 树                                            │
│     fiber := reconciler.GetFiber()                          │
│                                                             │
│  2. 布局计算 (layout 包)                                     │
│     adapter := NewFiberToNodeAdapter(fiber)                 │
│     layoutBox := layoutEngine.Layout(adapter)               │
│                                                             │
│  3. 转换为 PaintableBox (render 包)                          │
│     converter := NewFiberToPaintableConverter(fiber)        │
│     paintableBox := converter.Convert(layoutBox)            │
│                                                             │
│  4. 绘制 (paint 包)                                          │
│     paintEngine.Paint(paintableBox, buffer)                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 包依赖关系

```
┌──────────────┐
│ runtime/ui   │  (Fiber)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ runtime/     │
│ layout       │  (LayoutBox) - 无 paint 依赖
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ internal/    │
│ render       │  (Converter) - 依赖 layout + paint + ui
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ runtime/     │
│ paint        │  (PaintableBox, PaintEngine)
└──────────────┘

无循环依赖！
```

## 要删除的代码

| 文件/结构 | 原因 |
|-----------|------|
| `runtime/compute/types.go` 中的 `ComputedBox` | 中间层，删除 |
| `runtime/compute/adapter_convert.go` | 不再需要 |
| `runtime/compute/dirty_tracker.go` | 移动到其他地方或重构 |
| `runtime/compute/engine.go` | 布局引擎已在 `runtime/layout` |

## 要新增的代码

| 文件 | 功能 |
|------|------|
| `internal/render/converter.go` | LayoutBox → PaintableBox 转换器 |
| `runtime/paint/paintable_box.go` | 扩展 PaintableBox (已存在，需修改) |

## 实施步骤

### Phase 1: 扩展 PaintableBox

修改 `runtime/paint/paintable_box.go`，添加所有绘制需要的字段。

### Phase 2: 创建 Converter

创建 `internal/render/converter.go`，实现 `LayoutBox → PaintableBox` 转换。

### Phase 3: 更新 PaintEngine

修改 PaintEngine 直接使用 PaintableBox。

### Phase 4: 删除 ComputedBox

删除 `runtime/compute` 中的中间层代码。

## 对比

| 方面 | 之前 | 之后 |
|------|------|------|
| 数据结构 | 4种 | 3种 |
| 转换次数 | 3次 | 1次 |
| ComputedBox | 混合职责 | 删除 |
| LayoutBox | 混合 | 纯布局 |
| PaintableBox | 纯绘制 | 纯绘制 |
| 职责分离 | 不清晰 | 清晰 |
