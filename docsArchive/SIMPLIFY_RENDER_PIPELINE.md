# 简化渲染数据流方案

## 当前问题

当前渲染流程经过多次转换：

```
Fiber → LayoutBox → ComputedBox → PaintableBox → Paint
        (layout)    (compute)     (paint)
        
转换次数: 3次
数据结构: 4种 (Fiber, LayoutBox, ComputedBox, PaintableBox)
```

### 各结构字段对比

| 字段 | Fiber | LayoutBox | ComputedBox | PaintableBox |
|------|-------|-----------|-------------|--------------|
| X, Y | ❌ | ✅ | ✅ | ✅ |
| Width, Height | ❌ | ✅ | ✅ | ✅ |
| NodeID | ✅ | ❌ | ✅ | ✅ |
| DiffKey | ✅ | ❌ | ✅ | ✅ |
| Layer | ✅ | ✅ | ✅ | ✅ |
| Children | ✅ (树) | ✅ | ✅ | ✅ |
| PaintFunc | ✅ | ❌ | ❌ | ✅ |
| Style | ✅ | ❌ | ❌ | ✅ |

## 简化方案

### 目标架构

```
Fiber → LayoutBox (实现 PaintableNode) → Paint
        
转换次数: 1次
数据结构: 2种 (Fiber, LayoutBox)
```

### 方案：扩展 LayoutBox

在 `runtime/layout/types.go` 中扩展 LayoutBox：

```go
// LayoutBox 布局结果盒子
type LayoutBox struct {
    // === 位置和尺寸 ===
    ID       string
    X, Y     int
    Width    int
    Height   int
    Baseline int
    
    // === 渲染层 ===
    Layer  Layer
    ZIndex int
    
    // === Fiber-first 扩展字段 ===
    // 这些字段从 Fiber 复制，使 LayoutBox 成为完整的渲染数据结构
    NodeID   uint64      // 稳定运行时 ID
    DiffKey  string      // 用于脏追踪
    PaintFunc interface{} // func(x, y int) []paint.DrawCmd
    Style    interface{} // style.Style (使用 interface{} 避免循环导入)
    
    // === 子节点 ===
    Children []*LayoutBox
    Parent   *LayoutBox
}
```

### 添加 PaintableNode 实现

在 `runtime/layout/paintable.go` 中：

```go
package layout

import "github.com/wwsheng009/mint/runtime/paint"

// 确保 LayoutBox 实现 PaintableNode
var _ paint.PaintableNode = (*LayoutBox)(nil)

// NodeType 返回节点类型
func (b *LayoutBox) NodeType() paint.NodeType {
    // 从 PaintFunc 或其他字段推断
    return paint.NodeTypeElement
}

// Tag 返回标签
func (b *LayoutBox) Tag() string {
    return b.ID
}

// Style 返回样式
func (b *LayoutBox) Style() interface{} {
    return b.Style
}

// TextContent 返回文本内容
func (b *LayoutBox) TextContent() string {
    // 从 PaintFunc 或特定字段获取
    return ""
}

// Paint 调用绘制函数
func (b *LayoutBox) Paint(x, y int) []paint.DrawCmd {
    if b.PaintFunc == nil {
        return nil
    }
    if fn, ok := b.PaintFunc.(func(int, int) []paint.DrawCmd); ok {
        return fn(x, y)
    }
    return nil
}
```

### 简化后的数据流

```go
// 之前 (4种结构, 3次转换)
func (p *RenderingPipeline) Render(...) {
    fiber := getFiber()
    layoutBox := layoutEngine.Layout(fiber)           // 转换1
    computedBox := convertToComputedBox(layoutBox)    // 转换2
    paintableBox := computedBox.AsPaintable()         // 转换3
    paintEngine.Paint(paintableBox)
}

// 之后 (2种结构, 1次转换)
func (p *RenderingPipeline) Render(...) {
    fiber := getFiber()
    layoutBox := layoutEngine.Layout(fiber)           // 转换1 (LayoutBox 直接携带所有需要的数据)
    paintEngine.PaintLayout(layoutBox)                // LayoutBox 实现 PaintableNode
}
```

## 实施步骤

### Phase 1: 扩展 LayoutBox

1. 在 `LayoutBox` 中添加 `NodeID`, `DiffKey`, `PaintFunc`, `Style` 字段
2. 修改 `FiberToNodeAdapter` 填充这些字段

### Phase 2: 实现 PaintableNode

1. 创建 `runtime/layout/paintable.go`
2. 为 `LayoutBox` 实现 `PaintableNode` 接口

### Phase 3: 更新 PaintEngine

1. `PaintEngine` 添加 `PaintLayoutBox(*layout.LayoutBox)` 方法
2. 直接使用 `LayoutBox` 作为输入

### Phase 4: 删除中间结构

1. 删除 `ComputedBox`
2. 删除 `PaintableBox`
3. 删除 `adapter_convert.go`

## 优势

| 方面 | 之前 | 之后 |
|------|------|------|
| 数据结构 | 4种 | 2种 |
| 转换次数 | 3次 | 1次 |
| 内存分配 | 多次 | 1次 |
| 代码复杂度 | 高 | 低 |
| 维护成本 | 高 | 低 |

## 兼容性

### 保留旧 API (过渡期)

```go
// 旧 API (保留，标记 deprecated)
func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
    // 内部转换
    return e.PaintLayoutBox(convertToLayoutBox(layout.Root), buffer)
}

// 新 API
func (e *PaintEngine) PaintLayoutBox(root *layout.LayoutBox, buffer *paint.Buffer) error {
    // 直接使用 LayoutBox
}
```

## 结论

推荐采用**扩展 LayoutBox**方案，可以：
1. 减少数据结构数量 (4 → 2)
2. 减少转换次数 (3 → 1)
3. 降低代码复杂度
4. 提高性能 (减少内存分配)
