# Phase 4: 渲染管线集成

## 概述
**时间**: 3-5 天
**优先级**: P0（必须）
**依赖**: Phase 1-3 完成

## 目标
重构 DeclarativeNode.Paint()，实现完全的 Fiber-first 渲染流程，移除 VNode 运行时依赖。

---

## 架构约束

> **核心原则**：VNode 只存在于 Reconcile 阶段
>
> 根据 Fiber-first 架构设计：
> - VNode 在 Reconcile 后被丢弃
> - Layout 只读 Fiber（通过 FiberToNodeAdapter）
> - Paint 只用 PaintableBox

### 禁止的访问路径

```
❌ DeclarativeNode 持有 VNode 引用
❌ Paint 方法中访问 VNode
❌ Layout 阶段访问 VNode
```

### 正确的数据流

```
用户代码 → renderFn() → VNode (临时)
                          ↓
                   Reconciler
                          ↓ (VNode 被丢弃)
                   Fiber 树 (持久化)
                          ↓
                   FiberToNodeAdapter (internal/render)
                          ↓
                   layout.Engine (runtime/layout)
                          ↓
                   layout.LayoutBox
                          ↓
                   PaintableLayout
                          ↓
                   PaintEngine
                          ↓
                    Buffer
```

---

## 当前问题

### 1. DeclarativeNode 仍持有 VNode
```go
type DeclarativeNode struct {
    root     rtui.VNode  // ❌ 持有 VNode
    renderFn rtui.ComponentFunc
    // ...
}
```

### 2. Paint 方法流程仍依赖 VNode
```go
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
    // Phase 1: 获取 VNode 树
    if n.useFiber {
        n.root = n.renderWithFiberContext()  // ❌ 保存 VNode
    }
    
    // Phase 3: 使用 VNode 渲染
    pipeline.RenderWithConstraints(n.root, ...)  // ❌ 依赖 VNode
}
```

### 3. Legacy 路径仍存在
```go
// Legacy fallback
n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
```

---

## 目标架构

### 新的渲染流程
```
DeclarativeNode.Paint(ctx, buf)
    │
    ├── Phase 1: Fiber Reconciliation
    │   └── reconciler.Reconcile(renderFn)
    │       → VNode 临时存在 → 更新 Fiber → VNode 丢弃
    │
    ├── Phase 2: Fiber-based Layout
    │   └── layoutEngine.LayoutFiber(fiberRoot, constraints)
    │       → PaintableLayout
    │
    └── Phase 3: PaintableBox-based Paint
        └── paintEngine.PaintLayout(layout, buf)
```

---

## 实施步骤

### Step 4.1: 重构 DeclarativeNode 结构

**文件**: `internal/render/declarative_node.go`

**代码变更**:

```go
// Before
type DeclarativeNode struct {
    mu       sync.RWMutex
    root     rtui.VNode              // ❌ 删除
    renderFn rtui.ComponentFunc
    instance *rtui.ComponentContext
    focusMgr *rtui.FiberFocusManager
    fwApp    *framework.App
    reconciler rtui.Reconciler
    renderer   rtui.VNodeRenderer
    useFiber   bool
    layerMgr   *layer.Manager
}

// After
type DeclarativeNode struct {
    mu       sync.RWMutex
    // root 删除 - VNode 不再持久化
    fiberRoot *rtui.Fiber           // ✅ 新增：持久化的 Fiber 根
    renderFn  rtui.ComponentFunc
    
    instance  *rtui.ComponentContext
    focusMgr  *rtui.FiberFocusManager
    fwApp     *framework.App
    
    // Fiber-first 渲染组件
    reconciler     rtui.Reconciler
    layoutEngine   *FiberLayoutEngine      // ✅ 新增
    paintEngine    *PaintEngine
    converter      *FiberToPaintableConverter // ✅ 新增
    
    layerMgr *layer.Manager
}
```

---

### Step 4.2: 重构 Paint 方法

**文件**: `internal/render/declarative_node.go`

**代码**:

```go
// Paint 渲染组件到 buffer
func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    n.mu.Lock()
    defer n.mu.Unlock()

    // ========== Phase 1: Fiber Reconciliation ==========
    // VNode 在此阶段临时存在，协调后丢弃
    n.reconciler.Reconcile(n.renderFn)
    // VNode 已被丢弃，Fiber 树已更新

    // 获取 Fiber 根
    fiberRoot := n.getFiberRoot()
    if fiberRoot == nil {
        return
    }

    // ========== Phase 2: Fiber-based Layout ==========
    constraints := runtime.BoxConstraints{
        MinWidth:  0,
        MaxWidth:  ctx.AvailableWidth,
        MinHeight: 0,
        MaxHeight: ctx.AvailableHeight,
    }

    // 执行布局，返回 layout.LayoutResult
    layoutResult := n.layoutEngine.LayoutFiber(fiberRoot, constraints)
    if layoutResult == nil || layoutResult.Root == nil {
        return
    }

    // 转换为 PaintableLayout（使用 internal/render 中的转换器）
    paintableLayout := n.converter.Convert(fiberRoot, layoutResult.Root)

    // ========== Phase 3: PaintableBox-based Paint ==========
    // 检查是否有多层内容
    if n.hasMultipleLayers(paintableLayout) {
        // 多层渲染
        n.paintEngine.PaintMultiLayer(paintableLayout, buf)
    } else {
        // 单层渲染（性能更优）
        n.paintEngine.PaintLayout(paintableLayout, buf)
    }

    // 更新 Layer Manager（用于事件处理）
    n.updateLayerManager(paintableLayout)
}

// getFiberRoot 获取 Fiber 树根节点
func (n *DeclarativeNode) getFiberRoot() *rtui.Fiber {
    if n.reconciler == nil {
        return nil
    }
    
    // 从 Reconciler 获取当前 Fiber 根
    if fiberReconciler, ok := n.reconciler.(interface{ GetFiberRoot() *rtui.Fiber }); ok {
        return fiberReconciler.GetFiberRoot()
    }
    
    return n.fiberRoot
}

// hasMultipleLayers 检查是否有多层内容
func (n *DeclarativeNode) hasMultipleLayers(layout *paint.PaintableLayout) bool {
    if layout == nil || layout.Root == nil {
        return false
    }
    return n.hasLayerNodes(layout.Root, rtui.LayerBase)
}

// hasLayerNodes 递归检查是否有非 Base 层的节点
func (n *DeclarativeNode) hasLayerNodes(box *paint.PaintableBox, baseLayer rtui.Layer) bool {
    if box == nil {
        return false
    }
    
    if box.Layer > int(baseLayer) {
        return true
    }
    
    for _, child := range box.Children {
        if n.hasLayerNodes(child, baseLayer) {
            return true
        }
    }
    
    return false
}
```

---

### Step 4.3: 实现双轨运行机制

**文件**: `internal/render/declarative_node.go`

**目标**: 在重构期间保持向后兼容

**代码**:

```go
type RenderMode int

const (
    RenderModeLegacy RenderMode = iota
    RenderModeFiberFirst
    RenderModeBoth // 双轨对比
)

type DeclarativeNode struct {
    // ...
    renderMode RenderMode
}

func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    n.mu.Lock()
    defer n.mu.Unlock()
    
    switch n.renderMode {
    case RenderModeFiberFirst:
        n.fiberFirstPaint(ctx, buf)
        
    case RenderModeBoth:
        // 双轨运行，对比结果
        n.comparePaint(ctx, buf)
        
    default:
        n.legacyPaint(ctx, buf)
    }
}

// fiberFirstPaint 新的 Fiber-first 渲染路径
func (n *DeclarativeNode) fiberFirstPaint(ctx component.PaintContext, buf *paint.Buffer) {
    // Phase 1: Reconcile
    n.reconciler.Reconcile(n.renderFn)

    // Phase 2: Layout（使用 runtime.BoxConstraints）
    constraints := runtime.BoxConstraints{
        MaxWidth:  ctx.AvailableWidth,
        MaxHeight: ctx.AvailableHeight,
    }
    
    // 布局并转换
    paintableLayout := n.layoutEngine.LayoutFiberAndConvert(n.getFiberRoot(), constraints)

    // Phase 3: Paint
    n.paintEngine.PaintLayout(paintableLayout, buf)
}

// legacyPaint 旧的渲染路径（向后兼容）
func (n *DeclarativeNode) legacyPaint(ctx component.PaintContext, buf *paint.Buffer) {
    // 保持原有实现
    // ...
}

// comparePaint 双轨对比
func (n *DeclarativeNode) comparePaint(ctx component.PaintContext, buf *paint.Buffer) {
    // 创建两个 buffer
    legacyBuf := paint.NewBuffer(buf.Width, buf.Height)
    fiberBuf := paint.NewBuffer(buf.Width, buf.Height)
    
    // 分别渲染
    n.legacyPaint(ctx, legacyBuf)
    n.fiberFirstPaint(ctx, fiberBuf)
    
    // 对比结果
    if !buffersEqual(legacyBuf, fiberBuf) {
        log.RenderLogger.Warn("Render result mismatch!")
        n.logDiff(legacyBuf, fiberBuf)
    }
    
    // 使用新结果
    copyBuffer(buf, fiberBuf)
}
```

---

### Step 4.4: 删除 renderWithFiberContext 和 nonFiberRender

**文件**: `internal/render/declarative_node.go`

**操作**:
1. 删除 renderWithFiberContext 方法
2. 删除 nonFiberRender 方法
3. 删除相关的 VNode 处理逻辑

```go
// 删除这些方法
// func (n *DeclarativeNode) renderWithFiberContext() rtui.VNode { ... }
// func (n *DeclarativeNode) nonFiberRender() rtui.VNode { ... }
```

---

### Step 4.5: 更新 RenderingPipeline

**文件**: `internal/render/rendering_pipeline.go`

**目标**: 支持 Fiber 输入

**代码变更**:

```go
// 新增方法：RenderFiber
func (p *RenderingPipeline) RenderFiber(fiberRoot *rtui.Fiber, constraints Constraints, buf *paint.Buffer) error {
    // 1. Layout
    layout := p.layoutEngine.LayoutFiberAndConvert(fiberRoot, constraints)
    if layout == nil {
        return errors.New("layout failed")
    }
    
    // 2. Paint
    return p.paintEngine.PaintLayout(layout, buf)
}
```

---

### Step 4.6: 更新 NewDeclarativeNodeFromFunc

**文件**: `internal/render/declarative_node.go`

**代码**:

```go
// NewDeclarativeNodeFromFuncWithFiber 创建 Fiber-first 的 DeclarativeNode
func NewDeclarativeNodeFromFuncWithFiber(fn rtui.ComponentFunc) *DeclarativeNode {
    node := &DeclarativeNode{
        renderFn:    fn,
        instance:    rtui.NewComponentContextForRoot(),
        focusMgr:    rtui.NewFiberFocusManager(),
        renderMode:  RenderModeFiberFirst, // 默认使用 Fiber-first
    }
    
    // 初始化 Fiber-first 组件
    node.layoutEngine = NewFiberLayoutEngine()
    node.paintEngine = NewPaintEngine()
    node.converter = NewFiberToPaintableConverter()
    
    return node
}
```

---

## 测试计划

### 单元测试

```go
func TestDeclarativeNode_FiberFirstPaint(t *testing.T) {
    node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
        return text.New("Hello")
    })
    
    buf := paint.NewBuffer(80, 24)
    ctx := component.PaintContext{
        AvailableWidth:  80,
        AvailableHeight: 24,
    }
    
    node.Paint(ctx, buf)
    
    assert.Contains(t, buf.String(), "Hello")
}

func TestDeclarativeNode_CompareMode(t *testing.T) {
    node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
        return text.New("Test")
    })
    node.renderMode = RenderModeBoth
    
    buf := paint.NewBuffer(80, 24)
    ctx := component.PaintContext{
        AvailableWidth:  80,
        AvailableHeight: 24,
    }
    
    node.Paint(ctx, buf)
    
    // 验证无警告日志
}
```

### 集成测试

```bash
# 测试所有渲染模式
go test ./internal/render -run TestDeclarativeNode -v

# 测试双轨对比
go test ./internal/render -run TestCompareMode -v

# 测试多层渲染
go test ./internal/render -run TestMultiLayer -v
```

### 示例验证

```bash
# 运行所有示例
cd examples
for dir in */; do
    cd "$dir"
    go run main.go &
    sleep 2
    pkill -f "go run main.go"
    cd ..
done
```

---

## 验收标准

### 代码标准
- [ ] DeclarativeNode 不再持有 VNode
- [ ] Paint 方法使用 Fiber-first 流程
- [ ] 双轨运行机制可用
- [ ] Legacy 路径可配置

### 测试标准
- [ ] 所有单元测试通过
- [ ] 双轨对比无差异
- [ ] 示例应用正常运行

### 性能标准
- [ ] 渲染性能无退化
- [ ] 内存使用减少

---

## 回滚计划

### 配置切换
```bash
# 切换到 Legacy 模式
export MINT_RENDER_MODE=legacy

# 切换到双轨模式
export MINT_RENDER_MODE=both

# 切换到 Fiber-first 模式
export MINT_RENDER_MODE=fiber
```

### Git Tag
```bash
git tag -a refactor-phase4-start -m "Start Phase 4"
git tag -a refactor-phase4-done -m "Complete Phase 4"
```

---

## 完成检查清单

### 代码修改
- [ ] declarative_node.go: 重构结构和方法
- [ ] rendering_pipeline.go: 支持 Fiber 输入
- [ ] 新增双轨运行机制

### 测试
- [ ] declarative_node_test.go
- [ ] compare_mode_test.go
- [ ] 示例验证通过

### 文档
- [ ] 更新架构文档
- [ ] 更新 API 文档
- [ ] 更新迁移指南

---

**下一步**: [Phase 5: 组件迁移](./phase5_component_migration.md)
