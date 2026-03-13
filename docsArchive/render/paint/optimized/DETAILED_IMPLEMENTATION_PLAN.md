# Fiber-First 渲染管线优化 - 详细实施方案

## 当前系统状态分析

### 已完成的工作（约 70%）

1. **Fiber 结构已优化**
   - 文件: runtime/ui/fiber.go
   - 状态: 已包含 Instance 和 Style 字段，无 VNode 依赖

2. **Layout 引擎已优化**
   - 文件: runtime/compute/fiber_only_layout.go
   - 状态: BuildComputedBoxFiberOnly() 只使用 Fiber 树

3. **Paint 引擎已优化**
   - 文件: internal/render/paint_engine.go
   - 状态: PaintLayout() 使用 PaintableBox，支持组件自定义 Paint

4. **组件模板已完成**
   - 目录: ui/components/button/, ui/components/control/
   - 状态: Button 已按 Fiber-first 架构实现

### 需要完成的关键工作（约 30%）

#### 核心问题：渲染管线仍然依赖 VNode

**问题根源：**
1. DeclarativeNode.Paint() 仍然调用 renderWithFiberContext() 生成 VNode
2. PipelineRenderer.Render() 仍然接受 VNode 参数
3. RenderingPipeline.Render() 仍然接受 VNode 参数

**这违背了 Fiber-first 的核心原则：**
VNode 应该在 Reconcile 阶段被完全丢弃，Layout 和 Paint 阶段不应该访问 VNode
