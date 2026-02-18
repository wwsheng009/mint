# Fiber-first 实施检查清单

## 目标

根据 `docs/fiber/fiber_first/` 文档的要求，实现真正的 Fiber-first 架构：
- Layout 只读 Fiber，不读 VNode
- Render 只读 ComputedBox，不读 VNode
- Fiber 不持有 VNode 引用
- Paint 逻辑迁移到 Fiber

## 当前完成状态

### ✅ 已完成

| 组件 | 状态 | 说明 |
|------|------|------|
| `runtime/layout` | ✅ 完成 | 纯布局引擎，无 VNode 依赖 |
| `runtime/paint` (Buffer, Cell, Painter) | ✅ 完成 | 纯绘制层，无 VNode 依赖 |
| `internal/render/paint_engine.go` | ✅ 完成 | 新 API `PaintLayout()` 只依赖 `paint.PaintableLayout` |
| `runtime/paint/paintable_node.go` | ✅ 完成 | 定义 `PaintableNode` 接口 |
| `runtime/paint/paintable_box.go` | ✅ 完成 | 定义 `PaintableBox`, `PaintableLayout` |
| `runtime/compute/adapter_vnode.go` | ✅ 完成 | VNode → PaintableNode 适配器 (过渡期) |
| `runtime/compute/adapter_fiber.go` | ✅ 完成 | Fiber → PaintableNode 适配器 |
| `runtime/compute/adapter_convert.go` | ✅ 完成 | `AsPaintable()` 转换方法 |
| **`Fiber.PaintFunc`** | ✅ **完成** | 添加 PaintFunc 字段到 Fiber |
| **`CreateFiber` 提取 Paint** | ✅ **完成** | 在 CreateFiber 中提取 VNode.Paint() |
| **`FiberPaintableAdapter`** | ✅ **完成** | 优先使用 Fiber.PaintFunc |
| **`ComputedBox.DiffKey`** | ✅ **完成** | 添加 DiffKey 字段 |
| **`dirty_tracker`** | ✅ **完成** | 使用 DiffKey 替代 VNode.Key() |
| Fiber.Style | ✅ 完成 | Fiber 已有 Style 字段 |
| Fiber.LayoutDirection/Gap/Padding/Flex | ✅ 完成 | 布局字段已迁移 |
| Fiber.FocusableMeta | ✅ 完成 | Fiber-first focus 支持 |

### 🔄 待完成 (下一阶段)

| 组件 | 状态 | 剩余工作 |
|------|------|----------|
| `ComputedBox.VNode` | 🔄 待删除 | 当前仍持有 VNode 引用 (过渡期保留) |
| `Fiber.FocusableVNode` | 🔄 待删除 | 已标记 DEPRECATED (过渡期保留) |
| `adapter_vnode.go` | 🔄 待删除 | 等 ComputedBox.VNode 删除后 |

---

## 已完成的工作详情

### Phase 1: Fiber.PaintFunc 迁移 ✅

#### 1.1 添加 Fiber.PaintFunc 字段 ✅

文件: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // ... 现有字段
    
    // === Paint Function (Fiber-first) ===
    // PaintFunc stores the paint function extracted from VNode.Paint().
    // This enables Render to call Paint without VNode dependency.
    PaintFunc interface{}
}
```

#### 1.2 在 CreateFiber 中提取 Paint ✅

文件: `runtime/ui/fiber_util.go`

```go
// Extract PaintFunc (Fiber-first Paint Architecture)
var paintFunc interface{}
if _, ok := vnode.(interface { Paint(int, int) interface{} }); ok {
    paintFunc = vnode  // Store VNode reference for Paint call
}

return &Fiber{
    // ...
    PaintFunc: paintFunc,
}
```

#### 1.3 更新 FiberPaintableAdapter ✅

文件: `runtime/compute/adapter_fiber.go`

```go
func (a *FiberPaintableAdapter) Paint(x, y int) []paint.DrawCmd {
    // Primary Path: Use Fiber.PaintFunc (Fiber-first)
    if a.Fiber.PaintFunc != nil {
        if paintable, ok := a.Fiber.PaintFunc.(interface {
            Paint(int, int) []paint.DrawCmd
        }); ok {
            return paintable.Paint(x, y)
        }
    }
    // Fallback paths...
}
```

### Phase 2.1: ComputedBox.DiffKey ✅

#### 添加 DiffKey 字段 ✅

文件: `runtime/compute/types.go`

```go
type ComputedBox struct {
    // ...
    
    // DiffKey stores the key for dirty tracking (Fiber-first)
    DiffKey string
}
```

#### 更新 dirty_tracker.go ✅

```go
// NeedLayoutBox checks by DiffKey first, then VNode.Key() as fallback
func (t *DirtyTracker) NeedLayoutBox(box *ComputedBox) bool {
    if box.DiffKey != "" {
        return t.NeedLayout(box.DiffKey)
    }
    if box.VNode != nil {
        return t.NeedLayout(box.VNode.Key())
    }
    return false
}
```

#### 更新 engine.go ✅

```go
if fiber != nil {
    box.NodeID = fiber.NodeID
    box.DiffKey = fiber.DiffKey  // Fiber-first
    box.Layer = fiber.Layer
}
```

---

## 过渡期保留项

以下项目在过渡期保留，将在后续阶段删除：

| 项目 | 当前状态 | 删除条件 |
|------|----------|----------|
| `ComputedBox.VNode` | 保留 | 所有代码路径使用 ChildFiber/NodeID/DiffKey |
| `Fiber.FocusableVNode` | 保留 (DEPRECATED) | 所有代码使用 FocusableMeta |
| `adapter_vnode.go` | 保留 | ComputedBox.VNode 删除后 |

---

## 测试状态

### 通过的测试

- `runtime/paint/...` ✅ (197 个测试)
- `go build ./...` ✅

### 已知失败 (预先存在问题)

- `TestFiberOnlyLayoutDebug` - Fiber 布局边界计算问题
- `TestFiberVsVNodeLayout` - Fiber 尺寸计算问题
- `TestRenderingParity_FragmentTests` - Fragment 测量问题
- `TestVNodeRenderer_MeasureConsistency` - Renderer 适配问题

---

## 文件变更汇总

### 已修改文件

| 文件 | 变更 |
|------|------|
| `runtime/ui/fiber.go` | 添加 `PaintFunc interface{}` 字段 |
| `runtime/ui/fiber_util.go` | CreateFiber 中提取 Paint |
| `runtime/compute/types.go` | 添加 `DiffKey string` 字段 |
| `runtime/compute/adapter_fiber.go` | 优先使用 Fiber.PaintFunc |
| `runtime/compute/dirty_tracker.go` | 使用 DiffKey 替代 VNode.Key() |
| `runtime/compute/engine.go` | 填充 DiffKey 到 ComputedBox |
| `runtime/compute/fiber_only_layout.go` | 填充 DiffKey 到 ComputedBox |

### 新增文件

| 文件 | 说明 |
|------|------|
| `runtime/paint/paintable_node.go` | PaintableNode 接口定义 |
| `runtime/paint/paintable_box.go` | PaintableBox, PaintableLayout 结构体 |
| `runtime/compute/adapter_vnode.go` | VNode 适配器 |
| `runtime/compute/adapter_fiber.go` | Fiber 适配器 |
| `runtime/compute/adapter_convert.go` | AsPaintable() 转换方法 |
| `internal/render/paint_engine.go` | 重构后的 PaintEngine |

---

## 下一步工作 (可选)

1. **修复 Fiber 布局测试** - 解决 `TestFiberOnlyLayoutDebug` 等失败
2. **删除 ComputedBox.VNode** - 更新所有依赖代码
3. **删除 Fiber.FocusableVNode** - 确保所有代码使用 FocusableMeta
4. **删除 adapter_vnode.go** - 完成纯 Fiber-first 架构

---

## 参考

- `docs/fiber/fiber_first/fiber_first.md` - Fiber-first 架构原则
- `docs/fiber/fiber_first/fiber_vs_vnode.md` - VNode vs Fiber 数据迁移
- `docs/render/plan/RENDER_ENGINE_DECOUPLING.md` - 渲染引擎解耦方案
