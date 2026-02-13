# 渲染路径深度分析

**日期**: 2025-01-07
**状态**: 🔍 调试中

---

## 核心发现

### 1. Demo2 **确实使用**新的两阶段渲染系统

**证据**: `internal/render/declarative_node.go:59`

```go
func NewDeclarativeNodeFromFunc(fn rtui.ComponentFunc) *DeclarativeNode {
    node := &DeclarativeNode{...}
    // 使用新的 PipelineRenderer（两阶段渲染）
    node.renderer = NewPipelineRendererAdapter()  // ← 默认
    return node
}
```

### 2. Button **确实实现了** Paintable 接口

**证据**: `components/button/button.go:515`

```go
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // 检查 bounds 并填充宽度
    if b.bounds[2] > 0 {
        layoutWidth := b.bounds[2]
        if layoutWidth > buttonWidth {
            padding := layoutWidth - buttonWidth
            buttonText += strings.Repeat(" ", padding)  // ← 填充
        }
    }
    return []paint.DrawCmd{...}
}
```

### 3. PaintEngine **应该调用** Button.Paint()

**证据**: `internal/render/paint_engine.go:68-78`

```go
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
    // 检查是否实现了 Paintable 接口
    if paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
        // 应该调用 Paint(x, y)
        commands := paintable.Paint(box.Box.X, box.Box.Y)
        ...
        return nil
    }
    ...
}
```

---

## 完整的渲染路径

### 理论路径（应该发生）

```
ui.Run()
  ↓
framework.App.Run()
  ↓
DeclarativeNode.Paint()
  ↓
PipelineRendererAdapter.Render()
  ↓
PipelineRenderer.Render()
  ↓
RenderingPipeline.Render()
  ├─→ Phase 1: Layout
  │     ↓
  │   LayoutEngine.Layout()
  │     ↓
  │   calculatePositions()
  │     ↓
  │   SetBounds(x, y, width, height) ← flex width (18) ✅
  │
  └─→ Phase 2: Paint
        ↓
      PaintEngine.Paint()
        ↓
      paintNode()
        ↓
      type assertion: Paintable interface ✅
        ↓
      Button.Paint(x, y) ← 应该被调用
        ↓
      使用 b.bounds[2] 填充宽度 ✅
```

### 实际路径（推测）

**问题**: Button.Paint() 没有被调用

**可能原因**:

#### 原因 1: RenderingPipeline.Render() 失败，fallback 到 legacy

```go
// internal/render/rendering_pipeline.go:54-60
layout, err := p.layoutEngine.Layout(vnode, constraints)
if err != nil {
    return p.renderLegacy(vnode, 0, 0, buffer)  // ← Fallback
}
```

#### 原因 2: PaintEngine.Paint() 没有被调用

Layout 阶段成功，但 Paint 阶段因某种原因失败或被跳过。

#### 原因 3: 类型断言失败

`box.VNode` 没有通过 `Paintable` 接口的类型断言。

---

## 调试日志

为了追踪实际执行路径，已添加以下日志：

### 日志点 1: DeclarativeNode.Paint()

```go
[DeclarativeNode.Paint] n.renderer = ...
[DeclarativeNode.Paint] renderer type = *render.PipelineRendererAdapter
[DeclarativeNode.Paint] ✅ Using PipelineRendererAdapter
[DeclarativeNode.Paint] ❌ Pipeline render FAILED: ..., falling back to legacy
[DeclarativeNode.Paint] ✅ Pipeline render SUCCESS
```

**含义**:
- 看到 "✅ Using PipelineRendererAdapter" → 使用新渲染系统
- 看到 "❌ Pipeline render FAILED" → 渲染失败，fallback 到 legacy
- 看到 "✅ Pipeline render SUCCESS" → 渲染成功

### 日志点 2: PipelineRenderer.Render()

```go
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[PipelineRenderer] Using standard Render
[PipelineRenderer] ❌ Render FAILED: ..., falling back to legacy
[PipelineRenderer] ✅ Render SUCCESS
```

**含义**:
- "❌ Render FAILED" → PipelineRenderer 或 RenderingPipeline 失败
- "✅ Render SUCCESS" → PipelineRenderer 成功

### 日志点 3: RenderingPipeline.Render()

```go
[RenderingPipeline] Render started
[RenderingPipeline] ❌ Layout FAILED: ..., falling back to legacy
[RenderingPipeline] ✅ Layout complete, starting Paint phase
[RenderingPipeline] Starting Paint phase...
[RenderingPipeline] Paint complete, err=...
```

**含义**:
- "❌ Layout FAILED" → 布局引擎失败
- "✅ Layout complete" → 布局成功，开始 Paint
- "Paint complete, err=..." → Paint 阶段完成

### 日志点 4: PaintEngine.Paint()

```go
[PaintEngine.Paint] START: layout.Root=*button.ButtonVNode, box=(1,12,18x1)
[PaintEngine.Paint] END: err=<nil>
```

**含义**:
- "START" → PaintEngine 开始渲染
- "END" → PaintEngine 完成渲染

### 日志点 5: PaintEngine.paintNode()

```go
[Paint.paintNode] VNodeComponent at (1,12) size 18x1, vnode_type=*button.ButtonVNode
[Paint.paintNode]   ✅ Paintable: YES, calling Paint(1, 12)
[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)
```

**含义**:
- "✅ Paintable: YES" → 类型断言成功，应该调用 Button.Paint()
- "❌ Paintable: NO" → 类型断言失败，不会调用 Button.Paint()

---

## 启用调试

### 完整调试

```bash
cd examples/ui_demos/demo2_runtime_internals
TUI_DEBUG_RENDER=true go run .
```

### 分步调试

```bash
# 仅 PipelineRenderer
TUI_PIPELINE_DEBUG=true

# 仅 LayoutEngine
TUI_LAYOUT_DEBUG=true

# 仅 PaintEngine
TUI_PAINT_DEBUG=true

# 全部
TUI_DEBUG_RENDER=true  # 推荐使用此选项
```

---

## 预期输出

### 场景 1: 一切正常（理想情况）

```
[DeclarativeNode.Paint] ✅ Using PipelineRendererAdapter
[PipelineRenderer] Using standard Render
[RenderingPipeline] Render started
[RenderingPipeline] ✅ Layout complete, starting Paint phase
[RenderingPipeline] Starting Paint phase...
[PaintEngine.Paint] START: layout.Root=*button.ButtonVNode, box=(1,12,18x1)
[Paint.paintNode] VNodeComponent at (1,12) size 18x1, vnode_type=*button.ButtonVNode
[Paint.paintNode]   ✅ Paintable: YES, calling Paint(1, 12)
[Button.Paint] label="Test", hasFocus=false, x=1, y=12, bounds=[1, 12, 18, 0]
[Button.Paint] layoutWidth=18 > buttonWidth=14, padding=4  ← 拉伸！
[PaintEngine.Paint] END: err=<nil>
[RenderingPipeline] Paint complete, err=<nil>
[PipelineRenderer] ✅ Render SUCCESS
[DeclarativeNode.Paint] ✅ Pipeline render SUCCESS
```

**结果**: 按钮拉伸到 18 字符宽度 ✅

### 场景 2: Layout 失败

```
[DeclarativeNode.Paint] ✅ Using PipelineRendererAdapter
[PipelineRenderer] Using standard Render
[RenderingPipeline] Render started
[RenderingPipeline] ❌ Layout FAILED: ..., falling back to legacy
```

**结果**: Fallback 到 legacy，SetBounds 被禁用，按钮不拉伸 ❌

### 场景 3: 类型断言失败

```
...
[Paint.paintNode] VNodeComponent at (1,12) size 18x1, vnode_type=*button.ButtonVNode
[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)
```

**结果**: 不会调用 Button.Paint()，使用默认渲染，按钮不拉伸 ❌

---

## 下一步行动

1. **运行调试命令**
   ```bash
   TUI_DEBUG_RENDER=true go run examples/ui_demos/demo2_runtime_internals/main.go
   ```

2. **分析日志输出**
   - 确认哪个阶段失败或成功
   - 找到 Button.Paint() 没被调用的原因

3. **根据结果修复**
   - 场景 1: 成功！问题已解决 ✅
   - 场景 2: 修复 Layout 失败原因
   - 场景 3: 修复类型断言问题

---

## 技术细节

### SetBounds 的问题

**临时修复** (internal/render/declarative_node.go:356-371):
```go
if _, ok := vnode.(interface{ SetBounds(x, y, width, height int) }); ok {
    // ⚠️ 跳过 SetBounds 以避免覆盖 flex 宽度
    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[PaintVNode] ⚠️ Skipping SetBounds to preserve flex widths\n")
    }
}
```

**影响**:
- Legacy 路径中不会覆盖 flex 宽度
- 但 Button.Paint() 也拿不到 bounds 信息
- **这是临时方案**，需要找到根本原因

### Button.Paint() 的依赖

Button.Paint() **依赖** SetBounds 被正确调用：

```go
// components/button/button.go:637-648
if b.bounds[2] > 0 {
    layoutWidth := b.bounds[2]  // 从 SetBounds 获取
    if layoutWidth > buttonWidth {
        padding := layoutWidth - buttonWidth
        buttonText += strings.Repeat(" ", padding)  // 填充
    }
}
```

**如果 SetBounds 没被调用**，`b.bounds` 保持默认值 `[0, 0, 0, 0]`，按钮不会拉伸。

---

## 结论

当前状态：
- ✅ 架构正确：两阶段渲染
- ✅ Layout 正确：flex 宽度已计算
- ✅ SetBounds 正确：被调用并传递 flex 宽度
- ❓ Paint 未知：需要调试日志确认 Button.Paint() 是否被调用

**下一步**: 运行 `TUI_DEBUG_RENDER=true go run .` 并分析输出
