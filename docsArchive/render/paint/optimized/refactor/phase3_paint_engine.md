# Phase 3: Paint 引擎优化

## 概述
**时间**: 1-2 天
**优先级**: P0（必须）
**依赖**: Phase 2 完成

## 目标
确保 Paint 引擎完全解耦，只使用 PaintableBox，删除所有 VNode 依赖。

---

## 架构约束

> **核心原则**：Paint 阶段禁止访问 VNode
>
> 根据 Fiber-first 架构设计，VNode 只存在于 Reconcile 阶段。
> Paint 阶段应该只使用：
> - `paint.PaintableBox` - 组件实例
> - `paint.PaintableLayout` - 布局结果
> - `layout.LayoutBox` - 布局信息

### 禁止的访问路径

```
❌ Paint 阶段访问 VNode
❌ PaintEngine 调用 vnode.Paint()
❌ PaintEngine 调用 vnode.Text()
❌ PaintEngine 调用 vnode.Props()
```

### 正确的数据流

```
Fiber → LayoutEngine → layout.LayoutBox → PaintableLayout → PaintEngine → Buffer
```

---

## 当前问题

### 1. PaintVNode 方法仍存在
```go
// internal/render/declarative_node.go
func (n *DeclarativeNode) PaintVNode(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
    // ❌ 依赖 VNode
    switch vnode.Type() {
    case rtui.VNodeText:
        n.paintText(vnode, x, y, buf)
    // ...
    }
}
```

### 2. PaintEngine 中仍有 VNode 残留
```go
// internal/render/paint_engine.go
// 可能有 VNode 相关代码
```

### 3. Legacy 渲染路径
```go
// DeclarativeNode.Paint() 中
n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf) // fallback
```

---

## 实施步骤

### Step 3.1: 审计 PaintEngine

**目标**: 确认 PaintEngine 不依赖 VNode

**检查清单**:
```
[ ] 搜索 paint_engine.go 中的 "VNode" 引用
[ ] 搜索 paint_engine.go 中的 "vnode" 变量
[ ] 确认所有输入都是 PaintableBox
```

**搜索命令**:
```bash
grep -n "VNode\|vnode" internal/render/paint_engine.go
```

**预期结果**: PaintEngine 应该只依赖 PaintableBox 和 PaintableLayout

---

### Step 3.2: 删除 PaintVNode 方法

**文件**: `internal/render/declarative_node.go`

**操作**:
1. 标记 PaintVNode 为 deprecated
2. 添加警告日志
3. 最终删除

**Step 3.2.1: 添加废弃警告**

```go
// PaintVNode paints a VNode tree to the buffer.
// DEPRECATED: Use PaintLayout with PaintableLayout instead.
func (n *DeclarativeNode) PaintVNode(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
    if os.Getenv("MINT_WARN_LEGACY") == "true" {
        log.RenderLogger.Warn("[DEPRECATED] PaintVNode is deprecated, use PaintLayout instead")
    }
    
    // 保持原有实现，用于兼容
    // ...
}
```

**Step 3.2.2: 替换所有调用点**

搜索所有 PaintVNode 调用:
```bash
grep -rn "\.PaintVNode\(" --include="*.go"
```

替换为 PaintLayout:
```go
// Before
n.PaintVNode(vnode, x, y, buf)

// After
layout := converter.Convert(fiberRoot)
paintEngine.PaintLayout(layout, buf)
```

**Step 3.2.3: 删除 PaintVNode**

在 Phase 4 完成后，删除此方法及其辅助方法:
- PaintVNode
- paintText
- paintElement
- paintChildren
- paintTable
- paintBordered

---

### Step 3.3: 验证 PaintEngine.PaintLayout

**文件**: `internal/render/paint_engine.go`

**确保**:
1. PaintLayout 只接受 PaintableLayout
2. 所有绘制通过 PaintableNode.Paint() 完成
3. 没有直接访问 VNode 的代码

**代码验证**:

```go
// ✅ 正确的 Paint 签名
func (e *PaintEngine) PaintLayout(layout *paint.PaintableLayout, buffer *paint.Buffer) error {
    if layout == nil || layout.Root == nil {
        return nil
    }
    return e.paintBox(layout.Root, buffer)
}

// ✅ paintBox 使用 PaintableNode
func (e *PaintEngine) paintBox(box *paint.PaintableBox, buffer *paint.Buffer) error {
    // ...
    commands := box.Node.Paint(box.X, box.Y)
    // ...
}
```

---

### Step 3.4: 添加 Paint 可视化调试

**文件**: `internal/render/paint_engine_debug.go`（新建）

```go
package render

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/paint"
)

// DebugPaintLayout 打印 PaintableLayout 结构
func DebugPaintLayout(layout *paint.PaintableLayout) string {
    if layout == nil || layout.Root == nil {
        return "<nil layout>"
    }
    return debugPaintBox(layout.Root, 0)
}

func debugPaintBox(box *paint.PaintableBox, indent int) string {
    prefix := ""
    for i := 0; i < indent; i++ {
        prefix += "  "
    }
    
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%sBox[%d,%d %dx%d] NodeID=%d Layer=%d\n",
        prefix, box.X, box.Y, box.Width, box.Height, box.NodeID, box.Layer))
    
    for _, child := range box.Children {
        sb.WriteString(debugPaintBox(child, indent+1))
    }
    
    return sb.String()
}
```

---

### Step 3.5: 添加性能监控

**文件**: `internal/render/paint_engine_metrics.go`（新建）

```go
package render

import (
    "time"
    "github.com/wwsheng009/mint/runtime/paint"
)

// PaintMetrics 记录绘制性能指标
type PaintMetrics struct {
    TotalTime     time.Duration
    BoxCount      int
    CommandCount  int
    Depth         int
}

// PaintLayoutWithMetrics 带性能监控的绘制
func (e *PaintEngine) PaintLayoutWithMetrics(layout *paint.PaintableLayout, buffer *paint.Buffer) (*PaintMetrics, error) {
    metrics := &PaintMetrics{}
    start := time.Now()
    
    err := e.PaintLayout(layout, buffer)
    
    metrics.TotalTime = time.Since(start)
    metrics.BoxCount = countBoxes(layout.Root)
    
    return metrics, err
}

func countBoxes(box *paint.PaintableBox) int {
    if box == nil {
        return 0
    }
    count := 1
    for _, child := range box.Children {
        count += countBoxes(child)
    }
    return count
}
```

---

## 测试计划

### 单元测试

**文件**: `internal/render/paint_engine_test.go`

```go
func TestPaintLayout_NilInput(t *testing.T) {
    engine := NewPaintEngine()
    buf := paint.NewBuffer(80, 24)
    
    err := engine.PaintLayout(nil, buf)
    assert.Nil(t, err)
    
    err = engine.PaintLayout(&paint.PaintableLayout{}, buf)
    assert.Nil(t, err)
}

func TestPaintLayout_SingleBox(t *testing.T) {
    engine := NewPaintEngine()
    buf := paint.NewBuffer(80, 24)
    
    node := &mockPaintableNode{text: "Hello"}
    box := paint.NewPaintableBoxWithBounds(node, 10, 5, 10, 1)
    layout := paint.NewPaintableLayout(box)
    
    err := engine.PaintLayout(layout, buf)
    assert.Nil(t, err)
    
    // 验证文本被绘制
    assert.Contains(t, buf.String(), "Hello")
}
```

### 集成测试

```bash
# 测试 Paint 引擎
go test ./internal/render -run TestPaintEngine -v

# 测试性能监控
go test ./internal/render -run TestMetrics -v

# 测试调试输出
go test ./internal/render -run TestDebug -v
```

### 性能测试

```bash
# 基准测试
go test ./internal/render -bench=BenchmarkPaintLayout -benchmem

# 目标：1000 个 Box < 5ms
```

---

## 验收标准

### 代码标准
- [ ] PaintVNode 标记为 deprecated
- [ ] PaintVNode 调用点替换为新 API
- [ ] PaintEngine.PaintLayout 是唯一入口
- [ ] 无 VNode 依赖

### 测试标准
- [ ] PaintLayout 单元测试通过
- [ ] 性能监控测试通过
- [ ] 基准测试满足要求

### 功能标准
- [ ] 所有组件正确渲染
- [ ] 无渲染错误
- [ ] 性能无退化

---

## 完成检查清单

### 代码修改
- [ ] declarative_node.go: 标记 PaintVNode deprecated
- [ ] paint_engine.go: 验证无 VNode 依赖
- [ ] paint_engine_debug.go: 新增调试工具
- [ ] paint_engine_metrics.go: 新增性能监控

### 测试
- [ ] paint_engine_test.go: 新增/更新测试
- [ ] 集成测试通过
- [ ] 基准测试通过

### 文档
- [ ] 更新 PaintEngine 文档
- [ ] 更新迁移指南

---

**下一步**: [Phase 4: 渲染管线集成](./phase4_rendering_pipeline.md)
