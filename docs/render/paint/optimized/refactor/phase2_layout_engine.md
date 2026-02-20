# Phase 2: Layout 引擎优化

## 概述
**时间**: 2-3 天
**优先级**: P0（必须）
**依赖**: Phase 1 完成

## 目标
实现 Fiber-first Layout 引擎，使 Layout 阶段只读 Fiber，不再依赖 VNode。

---

## 架构原则

### 关键原则：runtime/layout 不依赖具体数据结构

`runtime/layout` 是**纯布局引擎**，属于基础设施层，必须遵循以下原则：

1. **不依赖 Fiber/VNode** - `runtime/layout` 只定义抽象接口（`Node`, `Measurable`, `Layered` 等）
2. **暴露 LayoutBox** - 布局结果通过 `LayoutBox` 和 `LayoutResult` 暴露
3. **适配器在外层** - Fiber/VNode 到 `layout.Node` 的适配器放在 `internal/render` 层

```
┌─────────────────────────────────────────────────────────────┐
│                    依赖关系图                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   internal/render (应用层)                                  │
│       ├── FiberToNodeAdapter   ← 适配 Fiber                │
│       ├── VNodeToNodeAdapter   ← 适配 VNode                │
│       ├── FiberToPaintableConverter                         │
│       └── FiberLayoutEngine                                 │
│              ↓ 依赖                                         │
│   runtime/layout (基础设施层)                               │
│       ├── Node 接口 (抽象)                                  │
│       ├── Measurable 接口 (抽象)                            │
│       ├── Layered 接口 (抽象)                               │
│       ├── Engine                                            │
│       ├── LayoutBox (布局结果)                              │
│       └── LayoutResult (布局结果)                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 当前状态

### 已有实现

| 组件 | 文件 | 状态 |
|------|------|------|
| layout.Node 接口 | runtime/layout/types.go | ✅ 已定义 |
| layout.Measurable 接口 | runtime/layout/types.go | ✅ 已定义 |
| layout.Layered 接口 | runtime/layout/layer.go | ✅ 已定义 |
| layout.Dirtyable 接口 | runtime/layout/types.go | ✅ 已定义 |
| layout.Engine | runtime/layout/types.go | ✅ 已实现 |
| layout.LayoutBox | runtime/layout/types.go | ✅ 已定义 |
| layout.LayoutResult | runtime/layout/types.go | ✅ 已定义 |
| FiberToNodeAdapter | internal/render/fiber_adapter.go | ✅ 部分实现 |
| VNodeToNodeAdapter | internal/render/fiber_adapter.go | ✅ 已实现 |
| LayoutSwitcher | internal/render/layout_switcher.go | ✅ 已实现 |

### 待完成

| 组件 | 问题 |
|------|------|
| FiberToNodeAdapter | 缺少 `Measure()` 方法 (Measurable 接口) |
| FiberToNodeAdapter | 缺少脏标记方法 (Dirtyable 接口) |
| Fiber Flags | 缺少 `FlagLayoutDirty` 常量 |

---

## 整体流程

```
┌─────────────────────────────────────────────────────────────┐
│                    Fiber-first Layout                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Fiber 树 (持久化)                                          │
│       ↓                                                     │
│  FiberToNodeAdapter (internal/render/fiber_adapter.go)      │
│       ↓ 实现 layout.Node 接口                               │
│  runtime/layout.Engine.Layout()                             │
│       ↓                                                     │
│  LayoutResult (layout.LayoutBox 树)                         │
│       ↓                                                     │
│  FiberToPaintableConverter (internal/render/converter.go)   │
│       ↓                                                     │
│  paint.PaintableLayout (paint.PaintableBox 树)              │
│       ↓                                                     │
│  PaintEngine.PaintLayout()                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 关键组件

| 组件 | 文件 | 职责 |
|------|------|------|
| layout.Node | runtime/layout/types.go | 布局节点抽象接口 |
| layout.Engine | runtime/layout/types.go | 纯布局引擎，不依赖具体数据 |
| layout.LayoutBox | runtime/layout/types.go | 布局结果盒子 |
| FiberToNodeAdapter | internal/render/fiber_adapter.go | 将 Fiber 适配为 layout.Node |
| FiberToPaintableConverter | internal/render/converter.go | 将 Fiber+LayoutBox 转换为 PaintableBox |
| FiberLayoutEngine | internal/render/fiber_layout_engine.go | Fiber-first Layout 入口 |

---

## 实施步骤

### Step 2.1: 完善 FiberToNodeAdapter - 实现 Measurable 接口

**文件**: `internal/render/fiber_adapter.go`（修改现有文件）

**添加代码**:

```go
// Measure 实现 layout.Measurable 接口
// 测量节点在给定约束下的理想尺寸
func (a *FiberToNodeAdapter) Measure(constraints layout.Constraints) layout.Size {
    if a.fiber == nil {
        return layout.Size{}
    }

    // 1. 尝试从 Instance 获取尺寸
    if a.fiber.Instance != nil {
        // 检查 Instance 是否实现 Measurable 接口
        if measurable, ok := a.fiber.Instance.(interface {
            Measure(layout.Constraints) layout.Size
        }); ok {
            return measurable.Measure(constraints)
        }

        // 使用 GetSize 方法
        w, h := a.fiber.Instance.GetSize()
        return layout.Size{
            Width:  constraints.ConstrainWidth(w),
            Height: constraints.ConstrainHeight(h),
        }
    }

    // 2. 从 Style 获取固定尺寸
    if a.fiber.Style.Width > 0 || a.fiber.Style.Height > 0 {
        w := a.fiber.Style.Width
        h := a.fiber.Style.Height
        return layout.Size{
            Width:  constraints.ConstrainWidth(w),
            Height: constraints.ConstrainHeight(h),
        }
    }

    // 3. 从 Props 获取尺寸
    if a.fiber.Props != nil {
        if w, ok := a.fiber.Props["width"].(int); ok && w > 0 {
            if h, ok := a.fiber.Props["height"].(int); ok && h > 0 {
                return layout.Size{
                    Width:  constraints.ConstrainWidth(w),
                    Height: constraints.ConstrainHeight(h),
                }
            }
        }
    }

    // 4. 默认值
    return layout.Size{Width: 0, Height: 0}
}
```

---

### Step 2.2: 完善 FiberToNodeAdapter - 实现 Dirtyable 接口

**文件**: `internal/render/fiber_adapter.go`（修改现有文件）

**添加代码**:

```go
// ========== layout.Dirtyable 接口实现 ==========

// IsLayoutDirty 检查是否需要重新布局
func (a *FiberToNodeAdapter) IsLayoutDirty() bool {
    if a.fiber == nil {
        return false
    }
    return a.fiber.Flags&rtui.FlagLayoutDirty != 0
}

// ClearLayoutDirty 清除布局脏标记
func (a *FiberToNodeAdapter) ClearLayoutDirty() {
    if a.fiber == nil {
        return
    }
    a.fiber.Flags &^= rtui.FlagLayoutDirty
}

// MarkLayoutDirty 标记需要重新布局
func (a *FiberToNodeAdapter) MarkLayoutDirty() {
    if a.fiber == nil {
        return
    }
    a.fiber.Flags |= rtui.FlagLayoutDirty
}
```

---

### Step 2.3: 添加 Fiber Flags 常量

**文件**: `runtime/ui/fiber.go`（修改现有文件）

**添加代码**:

```go
// Fiber Flags - 布局和绘制脏标记
const (
    // FlagLayoutDirty 表示节点需要重新布局
    FlagLayoutDirty EffectFlag = 1 << 10
    
    // FlagPaintDirty 表示节点需要重新绘制
    FlagPaintDirty EffectFlag = 1 << 11
)
```

---

### Step 2.4: 实现 FiberToPaintableConverter

**文件**: `internal/render/converter.go`（新建）

**代码**:

```go
package render

import (
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/internal/reconciler"
)

// FiberToPaintableConverter 将 Fiber 树 + LayoutBox 树转换为 PaintableBox 树
type FiberToPaintableConverter struct{}

// NewFiberToPaintableConverter 创建转换器
func NewFiberToPaintableConverter() *FiberToPaintableConverter {
    return &FiberToPaintableConverter{}
}

// Convert 转换 Fiber 树 + LayoutBox 为 PaintableLayout
// layoutBox 是布局引擎的输出，包含位置和尺寸信息
func (c *FiberToPaintableConverter) Convert(fiber *reconciler.Fiber, layoutBox *layout.LayoutBox) *paint.PaintableLayout {
    if fiber == nil || layoutBox == nil {
        return nil
    }

    rootBox := c.convertNode(fiber, layoutBox)
    return paint.NewPaintableLayout(rootBox)
}

// convertNode 递归转换单个节点
func (c *FiberToPaintableConverter) convertNode(fiber *reconciler.Fiber, layoutBox *layout.LayoutBox) *paint.PaintableBox {
    if fiber == nil || layoutBox == nil {
        return nil
    }

    // 创建 PaintableBox
    box := paint.NewPaintableBox(nil)

    // 从 LayoutBox 设置位置和尺寸
    box.NodeID = fiber.NodeID
    box.DiffKey = fiber.DiffKey
    box.X = layoutBox.X
    box.Y = layoutBox.Y
    box.Width = layoutBox.Width
    box.Height = layoutBox.Height
    box.Layer = int(layoutBox.Layer)

    // 设置 Node（从 Instance 获取 PaintableNode）
    if fiber.Instance != nil {
        if paintable, ok := fiber.Instance.(paint.PaintableNode); ok {
            box.Node = paintable
        }
    }

    // 递归转换子节点
    // Fiber 使用 Child -> Sibling 链表
    // LayoutBox 使用 Children 数组
    fiberChild := fiber.Child
    for i, layoutChild := range layoutBox.Children {
        if fiberChild == nil {
            break
        }
        
        childBox := c.convertNode(fiberChild, layoutChild)
        if childBox != nil {
            box.AddChild(childBox)
        }
        
        // 移动到下一个 sibling
        if i < len(layoutBox.Children)-1 {
            fiberChild = fiberChild.Sibling
        }
    }

    return box
}

// ConvertLayoutBox 直接将 LayoutBox 树转换为 PaintableBox 树
// 用于不需要 Fiber 引用的场景
func (c *FiberToPaintableConverter) ConvertLayoutBox(layoutBox *layout.LayoutBox) *paint.PaintableBox {
    if layoutBox == nil {
        return nil
    }

    box := paint.NewPaintableBox(nil)
    box.X = layoutBox.X
    box.Y = layoutBox.Y
    box.Width = layoutBox.Width
    box.Height = layoutBox.Height
    box.Layer = int(layoutBox.Layer)

    for _, child := range layoutBox.Children {
        childBox := c.ConvertLayoutBox(child)
        if childBox != nil {
            box.AddChild(childBox)
        }
    }

    return box
}
```

---

### Step 2.5: 更新 FiberLayoutEngine（可选，使用现有 LayoutSwitcher）

现有 `internal/render/layout_switcher.go` 已经实现了引擎切换逻辑，可以直接使用。

如果需要专门的 FiberLayoutEngine 封装：

**文件**: `internal/render/fiber_layout_engine.go`（新建）

**代码**:

```go
package render

import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/internal/reconciler"
)

// FiberLayoutEngine Fiber-first 布局引擎封装
// 注意：实际布局由 runtime/layout.Engine 执行，此类提供便捷方法
type FiberLayoutEngine struct {
    engine    *layout.Engine
    converter *FiberToPaintableConverter
}

// NewFiberLayoutEngine 创建 Fiber 布局引擎
func NewFiberLayoutEngine() *FiberLayoutEngine {
    return &FiberLayoutEngine{
        engine:    layout.NewEngine(),
        converter: NewFiberToPaintableConverter(),
    }
}

// LayoutFiber 对 Fiber 树进行布局
// 返回 LayoutResult，包含 LayoutBox 树
func (e *FiberLayoutEngine) LayoutFiber(root *reconciler.Fiber, constraints runtime.BoxConstraints) *layout.LayoutResult {
    if root == nil {
        return nil
    }

    // 1. 将 Fiber 适配为 layout.Node
    adapter := NewFiberToNodeAdapterPure(root)

    // 2. 转换约束
    layoutConstraints := layout.Constraints{
        MinWidth:  constraints.MinWidth,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: constraints.MinHeight,
        MaxHeight: constraints.MaxHeight,
    }

    // 3. 使用 runtime/layout 引擎进行布局
    result := e.engine.Layout(adapter, layoutConstraints)

    return result
}

// LayoutFiberAndConvert 布局并转换为 PaintableLayout
func (e *FiberLayoutEngine) LayoutFiberAndConvert(root *reconciler.Fiber, constraints runtime.BoxConstraints) *paint.PaintableLayout {
    // 1. 执行布局
    result := e.LayoutFiber(root, constraints)
    if result == nil || result.Root == nil {
        return nil
    }

    // 2. 转换为 PaintableLayout
    return e.converter.Convert(root, result.Root)
}

// GetEngine 返回底层布局引擎
func (e *FiberLayoutEngine) GetEngine() *layout.Engine {
    return e.engine
}
```

---

## 测试计划

### 单元测试

**文件**: `internal/render/fiber_adapter_test.go`（修改现有文件）

```go
package render_test

import (
    "testing"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestFiberToNodeAdapter_Measure(t *testing.T) {
    fiber := &rtui.Fiber{
        Style: rtui.Style{
            Width:  100,
            Height: 50,
        },
    }

    adapter := render.NewFiberToNodeAdapterPure(fiber)
    
    constraints := layout.Constraints{
        MinWidth:  0,
        MaxWidth:  200,
        MinHeight: 0,
        MaxHeight: 200,
    }
    
    size := adapter.Measure(constraints)
    
    if size.Width != 100 {
        t.Errorf("Expected width 100, got %d", size.Width)
    }
    if size.Height != 50 {
        t.Errorf("Expected height 50, got %d", size.Height)
    }
}

func TestFiberToNodeAdapter_Dirtyable(t *testing.T) {
    fiber := &rtui.Fiber{
        Flags: 0,
    }

    adapter := render.NewFiberToNodeAdapterPure(fiber)
    
    // 初始状态不脏
    if adapter.IsLayoutDirty() {
        t.Error("Expected not dirty initially")
    }
    
    // 标记为脏
    adapter.MarkLayoutDirty()
    if !adapter.IsLayoutDirty() {
        t.Error("Expected dirty after marking")
    }
    
    // 清除脏标记
    adapter.ClearLayoutDirty()
    if adapter.IsLayoutDirty() {
        t.Error("Expected not dirty after clearing")
    }
}
```

**文件**: `internal/render/converter_test.go`（新建）

```go
package render_test

import (
    "testing"
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime/layout"
)

func TestFiberToPaintableConverter_ConvertLayoutBox(t *testing.T) {
    converter := render.NewFiberToPaintableConverter()
    
    layoutBox := &layout.LayoutBox{
        ID:     "root",
        X:      10,
        Y:      20,
        Width:  100,
        Height: 50,
        Children: []*layout.LayoutBox{
            {
                ID:     "child1",
                X:      5,
                Y:      5,
                Width:  30,
                Height: 20,
            },
        },
    }
    
    paintBox := converter.ConvertLayoutBox(layoutBox)
    
    if paintBox == nil {
        t.Fatal("Expected non-nil result")
    }
    
    if paintBox.X != 10 || paintBox.Y != 20 {
        t.Errorf("Expected position (10, 20), got (%d, %d)", paintBox.X, paintBox.Y)
    }
    
    if paintBox.Width != 100 || paintBox.Height != 50 {
        t.Errorf("Expected size (100, 50), got (%d, %d)", paintBox.Width, paintBox.Height)
    }
    
    if len(paintBox.Children) != 1 {
        t.Errorf("Expected 1 child, got %d", len(paintBox.Children))
    }
}
```

### 集成测试

```bash
# 测试适配器
go test ./internal/render -run TestFiberToNodeAdapter -v

# 测试转换器
go test ./internal/render -run TestConverter -v

# 测试 FiberLayoutEngine
go test ./internal/render -run TestFiberLayoutEngine -v

# 测试布局引擎切换器
go test ./internal/render -run TestLayoutSwitcher -v
```

---

## 验收标准

### 代码标准
- [ ] FiberToNodeAdapter 实现 layout.Node 接口（已完成）
- [ ] FiberToNodeAdapter 实现 layout.Layered 接口（已完成）
- [ ] FiberToNodeAdapter 实现 layout.Measurable 接口（本次新增）
- [ ] FiberToNodeAdapter 实现 layout.Dirtyable 接口（本次新增）
- [ ] FiberToNodeAdapter 实现 layout.Marginal 接口（已完成）
- [ ] FiberToNodeAdapter 实现 layout.Positionable 接口（已完成）
- [ ] FiberToPaintableConverter 正确转换 Fiber + LayoutBox 为 PaintableBox
- [ ] FiberLayoutEngine 正确封装布局流程
- [ ] runtime/layout 包不依赖 Fiber/VNode（架构约束）

### 测试标准
- [ ] 单元测试覆盖所有新增接口方法
- [ ] 集成测试通过
- [ ] 性能测试：布局 1000 个节点 < 10ms

### 功能标准
- [ ] Fiber 树可正确布局
- [ ] 布局结果（LayoutBox）可正确转换为 PaintableBox
- [ ] Layer 和 ZIndex 正确处理
- [ ] 脏标记机制工作正常

---

## 完成检查清单

### 代码实现
- [ ] internal/render/fiber_adapter.go（添加 Measure、Dirtyable 方法）
- [ ] internal/render/converter.go（新建）
- [ ] internal/render/fiber_layout_engine.go（新建，可选）
- [ ] runtime/ui/fiber.go（添加 FlagLayoutDirty、FlagPaintDirty 常量）

### 测试
- [ ] internal/render/fiber_adapter_test.go（添加测试）
- [ ] internal/render/converter_test.go（新建）
- [ ] internal/render/fiber_layout_engine_test.go（新建，如需要）

### 文档
- [ ] 更新架构文档
- [ ] 更新 API 文档

---

## 架构约束说明

### 为什么 runtime/layout 不能依赖 Fiber？

1. **依赖倒置原则**：高层模块不应依赖低层模块，两者都应依赖抽象
   - `runtime/layout` 是基础设施层，应该只依赖抽象接口
   - `internal/render` 是应用层，负责将具体数据结构适配为抽象接口

2. **可测试性**：纯接口依赖使单元测试更容易 mock

3. **可复用性**：`runtime/layout` 可以被其他项目复用，不绑定到 Fiber

4. **关注点分离**：
   - `runtime/layout` 关注"如何布局"
   - `internal/render` 关注"如何将 UI 数据转换为布局输入"

### 正确的分层

```
┌──────────────────────────────────────┐
│         Application Layer            │
│  (internal/render, internal/reconciler)│
│                                      │
│  - Fiber 数据结构                    │
│  - FiberToNodeAdapter                │
│  - 业务逻辑                          │
└──────────────┬───────────────────────┘
               │ 依赖
               ↓
┌──────────────────────────────────────┐
│       Infrastructure Layer           │
│        (runtime/layout)              │
│                                      │
│  - 抽象接口 (Node, Measurable...)    │
│  - 纯布局算法                        │
│  - LayoutBox (输出)                  │
└──────────────────────────────────────┘
```

---

**下一步**: [Phase 3: Paint 引擎优化](./phase3_paint_engine.md)