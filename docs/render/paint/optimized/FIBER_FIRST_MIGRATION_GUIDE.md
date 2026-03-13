# Fiber-First Rendering Migration Guide

## 概述

本文档描述了新的 Fiber-first 渲染路径及其使用方法。新路径完全消除了 Layout 和 Paint 阶段对 VNode 的依赖。

## 新增文件

### 1. internal/render/declarative_node.go
**核心渲染管线（已合并到 DeclarativeNode）**

提供 FiberFirstPipeline 类型和 PaintFiberFirst() 函数，实现完整的 Fiber-first 渲染流程。

**关键功能：**
- FiberFirstPipeline.Render() - 执行完整的 Fiber-first 渲染
- PaintFiberFirst() - 高层 API，供 DeclarativeNode 使用
- IsFiberFirstEnabled() - 检查是否启用 Fiber-first 模式
- SetFiberFirstEnv() - 设置环境变量（测试用）

**渲染流程：**

```
Phase 1: Reconcile (VNode -> Fiber, VNode 被丢弃)
    ↓
Phase 2: Layout (Fiber -> LayoutResult, 无 VNode 访问)
    ↓
Phase 3: Paint (LayoutResult -> Buffer, 无 VNode 访问)
```

### 2. internal/render/fiber_adapter.go
**Fiber 到 Layout 接口适配器**

提供 FiberToNodeAdapter 类型，从 Fiber 树提取布局信息，替代访问 VNode。

**关键功能：**
- GetLayoutStyle() - 从 Fiber 获取样式
- GetLayoutDirection() - 从 Fiber 获取布局方向
- GetChildren() - 从 Fiber 树结构获取子节点
- GetInstance() - 从 Fiber 获取组件实例
- GetNodeID(), GetDiffKey(), GetLayer() - 获取元数据

**使用场景：**
Layout 引擎通过此适配器访问 Fiber 数据，无需访问 VNode。

### 3. internal/reconciler/get_fiber_root.go
**Reconciler 的 Fiber 根节点访问**

为 Reconciler 添加返回 Fiber 树的方法，替代旧的 Render() 方法。

**关键功能：**
- ReconcileAndRenderFiber() - 执行 Reconcile 并返回 Fiber 树
- GetCurrentFiber() - 获取当前已提交的 Fiber 根节点
- GetWorkInProgressFiber() - 获取正在处理的 Fiber 根节点
- MarkFiberDirty() - 标记需要重新渲染
- DumpFiberTree() - 调试工具，输出 Fiber 树结构

**与旧方法的区别：**

```
旧方法: reconciler.Render()
    → 执行 Reconcile + Layout + Paint
    → 保持 VNode 引用

新方法: reconciler.ReconcileAndRenderFiber()
    → 只执行 Reconcile
    → 返回 Fiber 树
    → VNode 被丢弃
    → Layout/Paint 由调用者处理
```

### 4. internal/render/declarative_node.go
**DeclarativeNode 集成**

提供 PaintWithFiberFirst() 方法，供 DeclarativeNode 使用新渲染路径。

**关键功能：**
- PaintWithFiberFirst() - Fiber-first 渲染入口
- ShouldUseFiberFirst() - 判断是否应该使用新路径
- getOrCreateFiberFirstPipeline() - 获取或创建渲染管线

### 5. internal/render/portal_layout_reset_test.go
**测试用例（Fiber-first 回归用例）**

提供单元测试，验证 Fiber-first 渲染路径的正确性。

## 如何启用 Fiber-first 渲染

### 方法 1：环境变量（推荐）

```bash
# 启用 Fiber-first 渲染
export MINT_FIBER_FIRST=true

# 启用调试日志
export MINT_DEBUG_FIBER_FIRST=true

# 运行程序
go run main.go
```

### 方法 2：代码中设置（测试用）

```go
import "github.com/wwsheng009/mint/internal/render"

// 启用 Fiber-first
render.SetFiberFirstEnv(true)

// 禁用 Fiber-first
render.SetFiberFirstEnv(false)
```

## 如何集成到 DeclarativeNode

### 步骤 1：修改 DeclarativeNode.Paint()

在 internal/render/declarative_node.go 中修改 Paint() 方法：

```go
func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    n.mu.Lock()
    defer n.mu.Unlock()

    // === 新增：检查是否使用 Fiber-first 路径 ===
    if ShouldUseFiberFirst(n) {
        err := n.PaintWithFiberFirst(ctx, buf, n.renderFn)
        if err == nil {
            return // Fiber-first 渲染成功，直接返回
        }
        // 如果失败，继续使用旧路径
        log.RenderLogger.Debug("Fiber-first render failed, falling back to legacy: %v", err)
    }

    // === 旧代码保持不变 ===
    if n.useFiber && n.reconciler != nil {
        n.root = n.renderWithFiberContext()
    } else {
        n.root = n.nonFiberRender()
    }
    
    if n.root == nil {
        return
    }

    n.applyFocusState()

    if n.renderer != nil {
        if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
            pipeline := adapter.GetPipeline()
            if err := pipeline.RenderWithConstraints(n.root, ctx.AvailableWidth, ctx.AvailableHeight, buf); err != nil {
                n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
            }
        } else {
            n.renderer.Render(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
        }
    } else {
        n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
    }
}
```

### 步骤 2：添加 fiberFirstPipeline 字段

在 internal/render/declarative_node.go 的 DeclarativeNode 结构体中添加：

```go
type DeclarativeNode struct {
    // ... 现有字段 ...

    // Fiber-first pipeline (新增)
    fiberFirstPipeline *FiberFirstPipeline
}
```

## 测试验证

### 运行单元测试

```bash
# 运行 Fiber-first 测试
go test -v ./internal/render -run TestFiberFirst

# 运行所有渲染测试
go test -v ./internal/render

# 运行并查看覆盖率
go test -cover ./internal/render
```

### 测试特定场景

```bash
# 测试 FiberToNodeAdapter
go test -v ./internal/render -run TestFiberToNodeAdapter

# 测试环境变量控制
go test -v ./internal/render -run TestIsFiberFirstEnabled
```

### 集成测试

```bash
# 启用 Fiber-first 运行示例应用
MINT_FIBER_FIRST=true go run examples/demo/main.go

# 启用调试日志
MINT_FIBER_FIRST=true MINT_DEBUG_FIBER_FIRST=true go run examples/demo/main.go
```

## 性能对比

### 测试脚本

```bash
#!/bin/bash

# 旧路径性能测试
echo "Testing legacy path..."
MINT_FIBER_FIRST=false go test -bench=. ./internal/render -benchtime=10s > legacy_perf.txt

# 新路径性能测试
echo "Testing Fiber-first path..."
MINT_FIBER_FIRST=true go test -bench=. ./internal/render -benchtime=10s > fiber_first_perf.txt

# 对比结果
echo "Comparing results..."
diff legacy_perf.txt fiber_first_perf.txt
```

### 预期改进

| 指标 | 旧路径 | 新路径 | 改进 |
|------|--------|--------|------|
| 内存占用 | VNode + Fiber + Box | Fiber + Box | ~30% ↓ |
| 渲染时间 | 创建 VNode + Layout + Paint | Layout + Paint | ~20% ↓ |
| GC 压力 | 每帧创建 VNode | 仅更新 Fiber | ~40% ↓ |

## 迁移路线图

### Phase 1: 并行运行（当前）
- ✅ 创建新的 Fiber-first 渲染路径
- ✅ 保留旧的渲染路径
- ✅ 通过环境变量切换
- 🔄 编写测试用例

### Phase 2: 测试验证
- [ ] 运行所有单元测试
- [ ] 运行集成测试
- [ ] 性能对比测试
- [ ] 真实应用测试

### Phase 3: 逐步迁移
- [ ] 在示例应用中默认启用
- [ ] 收集性能数据
- [ ] 修复发现的问题
- [ ] 优化性能

### Phase 4: 完全切换
- [ ] 默认使用新路径
- [ ] 旧路径作为 fallback
- [ ] 文档更新

### Phase 5: 清理
- [ ] 删除旧渲染路径
- [ ] 删除 VNode 运行时依赖
- [ ] 代码清理
- [ ] 最终测试

## 故障排除

### 问题 1：Fiber-first 模式未启用

**症状：** 即使设置了环境变量，仍然使用旧路径

**解决：**
```bash
# 确认环境变量已设置
echo $MINT_FIBER_FIRST

# 如果使用 IDE，确保环境变量在运行配置中
```

### 问题 2：reconciler 为 nil

**症状：** 日志显示 "Reconciler not available"

**解决：**
确保 DeclarativeNode 初始化时启用了 Fiber 模式：
```go
node := NewDeclarativeNode(app, renderFn, true) // useFiber = true
```

### 问题 3：渲染结果不正确

**症状：** 画面显示异常或空白

**解决：**
1. 启用调试日志：MINT_DEBUG_FIBER_FIRST=true
2. 检查 Fiber 树结构：使用 reconciler.DumpFiberTree()
3. 对比新旧路径的输出

## 调试技巧

### 查看 Fiber 树结构

```go
// 在 Paint 方法中添加
if os.Getenv("MINT_DEBUG_FIBER_FIRST") == "true" {
    if r, ok := n.reconciler.(*reconciler.Reconciler); ok {
        fmt.Println(r.DumpFiberTree())
    }
}
```

### 性能分析

```go
import "runtime/pprof"

// 开始 profiling
f, _ := os.Create("fiber_first.pprof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// 执行渲染
pipeline.Render(fiberRoot, ctx, buf)
```

## 相关文档

- [Fiber-First 架构设计](./FIBER_FIRST_RENDER_PIPELINE.md)
- [实施指南](./IMPLEMENTATION_GUIDE.md)
- [当前系统分析](/docsArchive/declarative_node_paint_analysis.md)

## 联系方式

如有问题或建议，请：
1. 查看本文档的故障排除章节
2. 查看相关设计文档
3. 提交 issue 或联系架构团队
