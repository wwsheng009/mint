# Layer 系统实施总结

**Fiber-First 架构下 Tooltip/Toast 组件的多层 Layer 支持**

---

## 📋 用户需求

**原始请求**: "调整 /ui/components/tooltip 组件支持可配置的多层渲染"

**背景**:
- Tooltip 和 Toast 组件的 SetLayer() 方法未实际生效
- 需要 Builder API 提供便捷的 Layer 设置方法
- 基于 Fiber-First 架构实现

---

## 🔍 实施前分析

### 发现的问题

| 问题 | 位置 | 问题描述 |
|------|------|---------|
| SetLayer() 无效 | `tooltip/vnode.go` | 字段不存在，方法没有副作用 |
| 默认 Layer 错误 | `tooltip/vnode.go` | 返回硬编码值为 `LayerOverlay` |
| Builder API 缺失 | `tooltip/builder.go` | 没有 Layer() 便捷方法 |

### Fiber-First 架构 (当前实现)

```
NewDeclarativeNodeFromFuncWithFiber(fn, fwApp)
    ↓
fiberFirstPaint(ctx, buf)
    ├─ Phase 1: Fiber Reconciliation (VNode → Fiber)
    │       └─ reconciler.Render(nullBuf, fn)
    │
    ├─ Phase 2: Layout (Fiber → LayoutBox)
    │       └─ newLayoutEngine.LayoutFiber(fiber, constraints)
    │           └─ FiberToNodeAdapterPure(fiber)
    │               └─ GetLayer() → layout.Layer
    │
    └─ Phase 3: Paint (LayoutBox → PaintableBox → Buffer)
        ├─ FiberToPaintableConverter.ConvertToLayout()
        │       └─ layout.Layer → paintableBox.Layer (int)
        │
        ├─ buildPlanes() 遍历 PaintableBox 树
        │       └─ planes.AddToLayer(RenderLayer(box.Layer), box)
        │
        └─ paintEngine.PaintPaintablePlanes(planes, buf)
            └─ 按 renderOrder [0,1,2,3,4] 依次渲染各层
```

---

## ✅ 实施内容

### 1. VNode 结构修改

**文件**: `ui/components/tooltip/vnode.go`

#### TooltipVNode

```go
type VNode struct {
    content   rtui.VNode
    text      string
    position  Position
    tooltip   *rtText.Text
    show      bool
    layer     rtui.Layer  // ← 新增持久化字段
}

// GetLayer 返回当前层级
func (t *VNode) GetLayer() rtui.Layer {
    return t.layer
}

// SetLayer 设置层级并返回自身
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
    t.layer = l
    return t
}

// Props 添加 layer 属性
func (t *VNode) Props() rtui.Props {
    return rtui.Props{
        "content":  t.content,
        "text":     t.text,
        "position": t.position,
        "show":     t.show,
        "layer":    t.layer,  // ← 新增
    }
}

// SetProps 支持 layer 属性
func (t *VNode) SetProps(props rtui.Props) {
    if layer, ok := props["layer"].(rtui.Layer); ok {
        t.layer = layer  // ← 新增
    }
    // ...
}

// New() 更新默认层
func New(content rtui.VNode, text string) *VNode {
    return &VNode{
        content: content,
        text:    text,
        layer:   rtui.LayerTooltip,  // ← 从 LayerOverlay 改为 LayerTooltip
    }
}
```

#### ToastVNode

```go
type ToastVNode struct {
    text      string
    toastType ToastType
    expireAt  time.Time
    show      bool
    layer     rtui.Layer  // ← 新增持久化字段
}

// GetLayer / SetLayer 实现类似
func (t *ToastVNode) GetLayer() rtui.Layer {
    return t.layer
}

func (t *ToastVNode) SetLayer(l rtui.Layer) rtui.VNode {
    t.layer = l
    return t
}

// NewToast() 默认层为 LayerOverlay
func NewToast(text string) *ToastVNode {
    return &ToastVNode{
        text:      text,
        toastType: ToastInfo,
        expireAt:  time.Now().Add(3 * time.Second),
        layer:     rtui.LayerOverlay,  // ← 默认覆盖层
    }
}
```

### 2. Builder API 扩展

**文件**: `ui/components/tooltip/builder.go`

#### Tooltip Builder

```go
type Builder struct {
    content  rtui.VNode
    text     string
    position Position
    layer    rtui.Layer  // ← 新增
}

// Layer() 通用方法
func Layer(l rtui.Layer) func(*Builder) {
    return func(b *Builder) {
        b.layer = l
    }
}

// 便捷方法
func BaseLayer() func(*Builder)       { return Layer(rtui.LayerBase) }
func OverlayLayer() func(*Builder)    { return Layer(rtui.LayerOverlay) }
func ModalLayer() func(*Builder)      { return Layer(rtui.LayerModal) }
func TooltipLayer() func(*Builder)    { return Layer(rtui.LayerTooltip) }
func InspectorLayer() func(*Builder)  { return Layer(rtui.LayerInspector) }

// SetRenderLayer() 别名方法
func SetRenderLayer() func(*Builder) {
    return TooltipLayer()
}

// Build() 使用 layer 字段
func (b *Builder) Build() *VNode {
    return &VNode{
        content:  b.content,
        text:     b.text,
        position: b.position,
        layer:    b.layer,  // ← 使用 builder 的 layer
    }
}
```

#### Toast Builder

```go
type ToastBuilder struct {
    text      string
    toastType ToastType
    layer     rtui.Layer  // ← 新增
}

// Tooltip Builder 同样支持:
// - Layer(l rtui.Layer)
// - BaseLayer(), OverlayLayer(), ModalLayer(), TooltipLayer(), InspectorLayer()
// - SetRenderLayer() (别名)

func SetRenderLayer() func(*ToastBuilder) {
    return OverlayLayer()  // Toast 默认为 Overlay 层
}
```

### 3. 使用示例文档

**文件**: `ui/components/tooltip/layer_demo.go`

创建完整的使用示例和指南，包括:
- Tooltip Layer 使用示例 (4 个场景)
- Toast Layer 使用示例 (4 个场景)
- 层级选择指南
- 完整示例

---

## 🎯 使用指南

### 方法 1: 使用 SetLayer()

```go
content := rtui.NewText("Hover me")

// Tooltip
t := tooltip.New(content, "Help info")
t.SetLayer(rtui.LayerTooltip)

// Toast
toast := tooltip.NewToast("Saved!")
toast.SetLayer(rtui.LayerOverlay)
```

### 方法 2: 使用 Builder API

```go
// Tooltip
tooltip.NewBuilder(content, "Help info").
    Layer(rtui.LayerTooltip).
    Build()

tooltip.NewBuilder(content, "Important").
    ModalLayer().
    Build()

// Toast
tooltip.NewToastBuilder("Success!").
    Info().
    OverlayLayer().
    Build()

tooltip.NewToastBuilder("Error!").
    Error().
    ModalLayer().
    Build()
```

### 方法 3: 使用便捷方法

```go
// Tooltip 便捷方法
tooltip.NewBuilder(content, "Info").
    TooltipLayer().
    Build()

// Toast 便捷方法
tooltip.NewToastBuilder("Saved!").
    SetRenderLayer().  // 使用别名
    Build()
```

---

## 📊 层级选择指南

| 场景 | 推荐层级 | 常量 |
|------|---------|------|
| 普通 Tooltip | Tooltip 层 (3) | `rtui.LayerTooltip` |
| 重要提示 | Modal 层 (2) | `rtui.LayerModal` |
| 普通 Toast | Overlay 层 (1) | `rtui.LayerOverlay` |
| 紧急通知 | Modal 层 (2) | `rtui.LayerModal` |
| 调试信息 | Inspector 层 (4) | `rtui.LayerInspector` |

### 完整层级定义

```go
const (
    LayerBase      Layer = iota  // 0: 基础层
    LayerOverlay               // 1: 覆盖层
    LayerModal                 // 2: 模态层
    LayerTooltip               // 3: 提示层
    LayerInspector             // 4: 检查器层
)
```

### Z-Order 计算方式

```
zOrder = int(Layer) * 10000 + Depth
```

---

## ✅ 验证结果

### 编译检查

```bash
cd E:\projects\yao\wwsheng009\mint
go build ./ui/components/tooltip/...
# ✅ 无编译错误
```

### 测试运行

```bash
go test ./ui/components/tooltip/... -v
# ✅ 所有 25 个测试通过
```

### 测试覆盖

```
TestNewTooltip                      ✅
TestTooltipBuilder                  ✅
TestTooltipPositionShortcuts        ✅
TestTooltipStyle                    ✅
TestTooltipConvenienceFunc          ✅
TestNewTooltipInstance              ✅
TestTooltipShowHide                 ✅
TestTooltipCalculatePosition        ✅
TestNewToast                        ✅
TestToastBuilder                    ✅
TestToastTypeShortcuts              ✅
TestToastConvenienceFuncs           ✅
TestNewToastInstance                ✅
TestToastShowHide                   ✅
TestToastExpiration                 ✅
TestNewToastManager                 ✅
TestToastManagerAdd                 ✅
TestToastManagerConvenience         ✅
TestToastManagerClear               ✅
TestToastManagerRemove              ✅
TestToastManagerHideAndRemove       ✅
TestToastManagerCleanExpired        ✅
```

---

## 🚧 技术细节

### fiberFirstPaint() 核心代码位置

| 步骤 | 位置 | 说明 |
|------|------|------|
| Phase 1: Reconciliation | `declarative_node.go:fiberFirstPaint()` | `reconciler.Render(nullBuf, fn)` |
| Phase 2: Layout | `layout_switcher.go:LayoutFiber()` | `NewLayoutEngineAdapter.LayoutFiber()` |
| Fiber→Node Adapter | `fiber_adapter.go:FiberToNodeAdapterPure` | `GetLayer()` 实现 |
| Phase 3.1: Layout→Paintable | `converter.go:FiberToPaintableConverter` | `Convert()` 方法 |
| Phase 3.2: Build Planes | `declarative_node.go:buildPlanes()` | 遍历树并分组 |
| Phase 3.3: Paint Planes | `paint_engine.go:PaintPaintablePlanes()` | 按 renderOrder 渲染 |

### Layer 传播流程 (零拷贝)

```
VNode.GetLayer()
    ↓
NewFiber() (runtime/ui/fiber_util.go)
    ↓
Fiber.Layer = vnode.GetLayer()
    ↓ (持久化存储在 Fiber)


FiberToNodeAdapterPure.GetLayer()
    ↓
layout.Layer(a.fiber.Layer)  // ← 零拷贝映射
    ↓
LayoutBox.Layer (从 layout.Node 获取)
    ↓
PaintableBox.Layer (int conversion)
    ↓
PaintablePlanes 分组
    ↓
PaintEngine.PaintPaintablePlanes()
```

### 统一类型体系

```go
// runtime/types/layer.go - 统一类型定义
type Layer int

const (
    LayerBase      Layer = iota  // 0
    LayerOverlay                 // 1
    LayerModal                   // 2
    LayerTooltip                 // 3
    LayerInspector               // 4
)

// runtime/layout/types.go - 类型别名
type Layer = types.Layer

// runtime/paint/paintable_planes.go - 类型别名
type RenderLayer = types.Layer

// runtime/ui/fiber.go - 类型别名
type Layer = types.Layer
```

### Modal 特殊处理

**位置**: `paint_engine.go:paintModalBackdropBox()`

```go
// 在 PaintEngine.PaintPaintablePlanes() 中
if layer == paint.RenderLayerModal && len(boxes) > 0 {
    e.paintModalBackdropBox(boxes[0], buffer)
}
```

Modal 层会额外绘制背景遮罩（灰化非 Modal 区域）。

---

## 📈 修改总结

### 文件修改

| 文件 | 变更内容 |
|------|---------|
| `ui/components/tooltip/vnode.go` | ✅ V/T 添加 `layer` 字段<br>✅ 修复 `GetLayer()` / `SetLayer()`<br>✅ 更新 `Props()` / `SetProps()`<br>✅ 修改默认层 |
| `ui/components/tooltip/builder.go` | ✅ Builder 添加 `Layer()` 方法<br>✅ 添加便捷方法<br>✅ Toast Builder 同样支持 |
| `ui/components/tooltip/layer_demo.go` | ✅ 新建使用示例文档 |

### API 变更

| 类别 | 方法 | 说明 |
|------|------|------|
| VNode | `GetLayer()` | 返回当前层级 |
| VNode | `SetLayer(l Layer)` | 设置层级 |
| Builder | `Layer(l Layer)` | 通用方法 |
| Builder | `BaseLayer()` | Base 层便捷方法 |
| Builder | `OverlayLayer()` | Overlay 层便捷方法 |
| Builder | `ModalLayer()` | Modal 层便捷方法 |
| Builder | `TooltipLayer()` | Tooltip 层便捷方法 |
| Builder | `InspectorLayer()` | Inspector 层便捷方法 |
| Builder | `SetRenderLayer()` | 别名方法 |

---

## 📚 相关文档

- `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md` - Fiber-First Layer 系统完整说明
- `docs/layer/LAYER_SYSTEM_ARCHITECTURE.md` - Layer 系统架构
- `docs/layer/TWO_RENDERING_SYSTEMS_EXPLAINED.md` - ~~历史分析（已过时）~~
- `ui/components/tooltip/layer_demo.go` - 完整使用示例

---

## 🎉 总结

### 核心成就

1. ✅ **修复了 SetLayer() 无效问题**
   - V/ToastVNode 添加持久化 `layer` 字段
   - `SetLayer()` 现在正确修改字段

2. ⚙️ **更新了默认层级**
   - Tooltip 默认: `LayerBase` → `LayerTooltip`
   - Toast 默认: `LayerBase` → `LayerOverlay`

3. 🛠️ **扩展了 Builder API**
   - 添加 `Layer()` 通用方法
   - 添加 6 个便捷方法
   - Toast Builder 同样支持

4. 📚 **创建了使用示例文档**
   - 完整的 layer_demo.go
   - 层级选择指南
   - 多个实际场景示例

### 关键要点

- **Layer 存储在 Fiber 节点中**，跨帧持久化
- **fiberFirstPaint() 三阶段渲染**:
  1. Phase 1: Fiber Reconciliation (VNode → Fiber)
  2. Phase 2: Layout (Fiber → LayoutBox)
  3. Phase 3: Paint (LayoutBox → PaintablePlanes → Buffer)
- **零拷贝传递**: Fiber.Layer → layout.Layer → paintableBox.Layer
- **统一类型体系**: 所有包使用 `types.Layer` 类型别名
- **自动多层级渲染**: PaintEngine 按 renderOrder 渲染各层
- **Modal 特殊处理**: 自动绘制背景遮罩（灰化非 Modal 区域）

---

**实施日期**: 2026-02-23
**最后更新**: 2026-03-01
**状态**: ✅ 完成 - Fiber-first 架构
**版本**: 2.0
**关键更新**: 添加 fiberFirstPaint 三阶段渲染流程分析
