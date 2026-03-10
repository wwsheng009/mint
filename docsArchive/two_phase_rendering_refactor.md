# 两阶段渲染重构设计方案

**文档版本**: 1.0
**创建日期**: 2025-01-07
**目标**: 修复渲染时序问题，实现 Measure → Layout → Paint 的正确渲染管线
**状态**: 设计阶段

---

## 📋 目录

1. [背景与问题](#背景与问题)
2. [当前架构分析](#当前架构分析)
3. [目标架构设计](#目标架构设计)
4. [技术方案](#技术方案)
5. [实施计划](#实施计划)
6. [风险评估](#风险评估)
7. [测试策略](#测试策略)
8. [回滚计划](#回滚计划)
9. [影响范围](#影响范围)

---

## 背景与问题

### 问题描述

当前 Mint TUI 框架存在**渲染时序错误**：

```
当前流程（错误）:
1. Measure     ✅ 测量阶段
2. Paint       ❌ 绘制阶段（太早）
3. Layout      ✅ 布局计算
4. SetBounds   ✅ 设置边界（太晚）
```

**后果**：
- 组件在 `Paint()` 时无法获取正确的布局宽度
- `FillWidth` 功能在逻辑正确但视觉失效
- 按钮无法拉伸填充可用空间

### 根本原因

渲染引擎的 `Render()` 方法调用顺序不正确，导致：
- `Paint()` 在 `Layout/SetBounds()` 之前执行
- VNode 的 `bounds` 字段在绘制时还未被设置
- 布局引擎正确计算了尺寸和位置，但组件看不到这些信息

### 受影响功能

1. ❌ **Wrap 组件的 FillWidth** - 按钮无法拉伸
2. ❌ **VStack 的 stretchCross** - 子元素无法横向拉伸
3. ❌ **HStack 的 stretchCross** - 子元素无法纵向拉伸
4. ❌ **所有依赖 SetBounds 的组件** - 无法获取布局位置

---

## 当前架构分析

### 当前渲染流程

```go
// pseudo-code of current rendering engine
func (e *Engine) Render(vnode VNode, buffer Buffer) {
    // Phase 1: Measure (测量)
    constraints := e.getRootConstraints()
    size := e.measureVNode(vnode, constraints)

    // Phase 2: Paint directly (直接绘制) ← 问题所在！
    vnode.Paint(0, 0, buffer)

    // Phase 3: Layout (布局) - 但 Paint 已经结束了！
    layout, _ := e.Layout(vnode, constraints)
    e.calculatePositions(layout.Root, 0, 0)
}
```

### 关键代码位置

| 文件 | 函数 | 行号 | 职责 |
|------|------|------|------|
| `runtime/render/renderer.go` | `Render()` | ~50 | 主渲染循环 |
| `runtime/compute/engine.go` | `Layout()` | ~39 | 布局计算 |
| `runtime/compute/engine.go` | `calculatePositions()` | ~850 | 位置计算和 SetBounds |
| `components/button/button.go` | `Paint()` | ~515 | 按钮绘制 |

### 数据流分析

```
用户代码 (Wrap + Buttons)
    ↓
Build() 阶段 → VStack(HStack(Button, Button, ...))
    ↓
Measure() 阶段 → 计算每个 Button 的 flexWidth = 18 ✅
    ↓
Paint() 阶段 → Button.Paint() 读取 bounds = [0,0,0,0] ❌
    ↓
Layout() 阶段 → calculatePositions() 设置 bounds = [1,12,18,1] ✅
    ↓
结果：按钮使用自然宽度绘制，看不到布局分配的宽度
```

---

## 目标架构设计

### 新渲染流程

```
目标流程（正确）:
1. Measure     ✅ 测量阶段 - 计算所有组件尺寸
2. Layout      ✅ 布局阶段 - 计算所有组件位置 + SetBounds
3. Paint       ✅ 绘制阶段 - 使用正确的 bounds 绘制
```

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Application (用户代码)                    │
│  WrapBuilder(...).FillWidth().Build()                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   Rendering Engine (渲染引擎)                │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Phase 1: Measure (测量)                            │   │
│  │  - 调用 Measure() 计算所有组件尺寸                   │   │
│  │  - 收集约束信息                                      │   │
│  │  - 返回 LayoutMeasurement (包含尺寸和子元素约束)     │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ↓                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Phase 2: Layout (布局)                             │   │
│  │  - 调用 calculatePositions() 计算位置               │   │
│  │  - 对每个 VNode 调用 SetBounds(x, y, w, h)          │   │
│  │  - 生成 ComputedLayout 树                           │   │
│  └─────────────────────────────────────────────────────┘   │
│                          ↓                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Phase 3: Paint (绘制)                              │   │
│  │  - 调用 Paint() 绘制所有组件                         │   │
│  │  - 此时 bounds 已正确设置 ✅                         │   │
│  │  - 输出到缓冲区                                      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   Terminal Output (终端输出)                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 技术方案

### 方案概述

**核心改动**：将渲染流程从混合模式改为清晰的三个阶段

```
当前: Measure + Paint 混合 → Layout
目标: Measure → Layout → Paint (顺序执行)
```

### 关键修改点

#### 1. 新增 `ComputedLayout.Paint()` 方法

**文件**: `runtime/compute/layout.go` (新建)

```go
package compute

import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/paint"
)

// Paint generates draw commands from a computed layout tree
func (cl *ComputedLayout) Paint() []paint.DrawCmd {
    if cl.Root == nil {
        return nil
    }
    return cl.paintBox(cl.Root, 0, 0)
}

// paintBox recursively paints a computed box and its children
func (cl *ComputedLayout) paintBox(box *ComputedBox, x, y int) []paint.DrawCmd {
    var cmds []paint.DrawCmd

    // Paint this box if it has a VNode with Paint method
    if box.VNode != nil {
        if paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
            // Call Paint with correct bounds already set by SetBounds
            childCmds := paintable.Paint(box.Box.X, box.Box.Y)
            cmds = append(cmds, childCmds...)
        }
    }

    // Recursively paint children
    for _, child := range box.Children {
        childCmds := cl.paintBox(child, 0, 0)
        cmds = append(cmds, childCmds...)
    }

    return cmds
}
```

#### 2. 修改 `Engine.Render()` 方法

**文件**: `runtime/render/renderer.go` (修改)

```go
// Render renders a VNode tree to a buffer using two-phase approach
func (r *Renderer) Render(vnode ui.VNode, buffer Buffer) error {
    // Phase 1: Measure and Layout
    constraints := r.getConstraints(buffer.Width(), buffer.Height())
    layout, err := r.engine.Layout(vnode, constraints)
    if err != nil {
        return fmt.Errorf("layout failed: %w", err)
    }

    // Phase 2: Paint (AFTER layout is complete and bounds are set)
    cmds := layout.Paint()

    // Phase 3: Apply draw commands to buffer
    r.applyCommands(cmds, buffer)

    return nil
}
```

#### 3. 确保 `calculatePositions()` 完成后再 Paint

**文件**: `runtime/compute/engine.go` (确认逻辑)

```go
// Layout performs layout calculation and position assignment
func (e *Engine) Layout(vnode VNode, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    // Build layout tree and measure (Phase 1)
    root := e.buildComputedBox(vnode, nil, constraints)
    if root == nil {
        return NewComputedLayout(nil), nil
    }

    // Calculate positions (Phase 2) - this calls SetBounds!
    e.calculatePositions(root, 0, 0)

    // Clear dirty flags
    root.ClearDirty()

    return NewComputedLayout(root), nil
}
```

#### 4. 修改 `Button.Paint()` 使用正确的 bounds

**文件**: `components/button/button.go` (优化)

```go
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // ... existing code ...

    // Use bounds that were set by calculatePositions() BEFORE Paint
    if b.bounds[2] > 0 {
        layoutWidth := b.bounds[2]  // bounds = [x, y, width, height]

        if layoutWidth > buttonWidth {
            padding := layoutWidth - buttonWidth
            buttonText += strings.Repeat(" ", padding)
            buttonWidth = layoutWidth
        }
    }

    // ... rest of painting logic ...
}
```

### 数据结构变更

#### ComputedLayout 增强

**文件**: `runtime/compute/layout.go` (修改)

```go
type ComputedLayout struct {
    Root *ComputedBox
    // 新增：缓存绘制命令，避免重复计算
    cachedCmds []paint.DrawCmd
    cached     bool
}

// Paint generates draw commands (with caching)
func (cl *ComputedLayout) Paint() []paint.DrawCmd {
    if cl.cached {
        return cl.cachedCmds
    }

    cl.cachedCmds = cl.paintBox(cl.Root, 0, 0)
    cl.cached = true
    return cl.cachedCmds
}

// Invalidate clears the cache when layout changes
func (cl *ComputedLayout) Invalidate() {
    cl.cached = false
    cl.cachedCmds = nil
}
```

---

## 实施计划

### 阶段 1: 准备阶段 (1-2 天)

**目标**: 建立测试基线和备份机制

- [ ] **1.1 创建功能分支**
  ```bash
  git checkout -b feature/two-phase-rendering
  ```

- [ ] **1.2 建立测试基线**
  - 运行所有现有测试，记录结果
  - 截图所有 demo 应用的当前状态
  - 创建回归测试套件

- [ ] **1.3 代码审查**
  - 标记所有调用 `Paint()` 的位置
  - 标记所有调用 `Layout()` 的位置
  - 识别可能依赖旧时序的代码

### 阶段 2: 核心实施 (3-5 天)

**目标**: 实现两阶段渲染的核心逻辑

- [ ] **2.1 新增 `ComputedLayout.Paint()` 方法**
  - 文件: `runtime/compute/layout.go`
  - 实现 `paintBox()` 递归绘制
  - 添加命令缓存机制

- [ ] **2.2 修改 `Renderer.Render()`**
  - 文件: `runtime/render/renderer.go`
  - 改为 Layout → Paint 顺序
  - 添加错误处理

- [ ] **2.3 确认 `Engine.Layout()` 完整性**
  - 文件: `runtime/compute/engine.go`
  - 确保 `calculatePositions()` 被调用
  - 确保 `SetBounds()` 被调用

- [ ] **2.4 单元测试**
  - 测试 `ComputedLayout.Paint()`
  - 测试 `Renderer.Render()` 新流程
  - 测试 bounds 时序

### 阶段 3: 组件适配 (2-3 天)

**目标**: 确保所有组件兼容新渲染流程

- [ ] **3.1 Button 组件**
  - 确认 `Paint()` 使用 bounds
  - 测试 FillWidth 功能

- [ ] **3.2 Input 组件**
  - 检查边界计算
  - 测试光标位置

- [ ] **3.3 Text 组件**
  - 检查文本填充逻辑
  - 测试对齐方式

- [ ] **3.4 布局组件**
  - HStack, VStack, Grid
  - Bordered 容器

- [ ] **3.5 第三方组件**
  - 检查是否有外部依赖
  - 更新文档

### 阶段 4: 集成测试 (2-3 天)

**目标**: 全面测试新渲染流程

- [ ] **4.1 回归测试**
  - 所有单元测试必须通过
  - 所有集成测试必须通过
  - 性能测试无退化

- [ ] **4.2 功能测试**
  - Wrap 组件 FillWidth ✅
  - VStack stretchCross ✅
  - HStack stretchCross ✅
  - 所有 demo 应用正常运行

- [ ] **4.3 边缘案例**
  - 空组件树
  - 嵌套布局
  - 极端约束条件

### 阶段 5: 文档和发布 (1-2 天)

- [ ] **5.1 更新文档**
  - 渲染流程说明
  - 组件开发指南
  - 迁移指南

- [ ] **5.2 代码审查**
  - 团队审查
  - 性能分析

- [ ] **5.3 发布**
  - 合并到 main 分支
  - 发布说明
  - 版本标记

---

## 风险评估

### 高风险 🔴

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| **现有组件依赖旧时序** | 功能中断 | 中 | 全面的回归测试 |
| **性能退化** | 用户体验下降 | 低 | 性能基准测试 |
| **Fiber 渲染路径不兼容** | 核心功能失效 | 中 | 保留旧路径作为 fallback |

### 中风险 🟡

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| **第三方组件不兼容** | 部分功能失效 | 低 | 提供迁移指南 |
| **调试困难** | 开发效率下降 | 中 | 增强调试工具 |
| **文档不完整** | 理解困难 | 中 | 详细注释和示例 |

### 低风险 🟢

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| **代码复杂度增加** | 维护成本上升 | 高 | 代码审查和重构 |
| **测试覆盖不足** | 边缘情况失效 | 低 | 持续集成测试 |

---

## 测试策略

### 单元测试

**文件**: `runtime/compute/engine_test.go` (新增)

```go
func TestTwoPhaseRenderingOrder(t *testing.T) {
    engine := NewEngine()

    // Create a simple VNode tree
    vnode := ui.HStack(
        app.Button("A"),
        app.Button("B"),
    )

    // Phase 1: Layout
    layout, err := engine.Layout(vnode, runtime.BoxConstraints{
        MinWidth: 0, MaxWidth: 100,
        MinHeight: 0, MaxHeight: 100,
    })
    require.NoError(t, err)

    // Phase 2: Paint
    cmds := layout.Paint()
    assert.NotNil(t, cmds)

    // Verify bounds were set before Paint
    for _, box := range layout.AllBoxes() {
        if box.VNode != nil {
            if tagger, ok := box.VNode.(interface{ Tag() string }); ok {
                if tagger.Tag() == "button" {
                    assert.Greater(t, box.Box.Width, 0, "Button should have width")
                }
            }
        }
    }
}
```

### 集成测试

**文件**: `examples/integration/test_two_phase_rendering.go` (新增)

```go
func TestWrapFillWidthRendering(t *testing.T) {
    // Create Wrap with FillWidth
    buttons := []ui.VNode{
        app.Button("1"),
        app.Button("2"),
        app.Button("3"),
        app.Button("4"),
    }

    wrap := layout.WrapBuilder(buttons...).
        Gap(1).
        ScreenWidth(80).
        FillWidth().
        Build()

    // Render to buffer
    buffer := NewTestBuffer(80, 20)
    renderer := render.NewRenderer()
    err := renderer.Render(wrap, buffer)
    require.NoError(t, err)

    // Verify buttons are stretched
    output := buffer.String()
    assert.Contains(t, output, "[1]")  // Buttons should be visible
    // TODO: Add precise width checking
}
```

### 性能测试

**文件**: `runtime/compute/engine_bench_test.go` (新增)

```go
func BenchmarkTwoPhaseRendering(b *testing.B) {
    engine := NewEngine()
    vnode := createComplexLayout() // 1000 components

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        layout, _ := engine.Layout(vnode, runtime.BoxConstraints{
            MinWidth: 0, MaxWidth: 80,
            MinHeight: 0, MaxHeight: 24,
        })
        layout.Paint()
    }
}
```

### 回归测试清单

- [ ] 所有现有单元测试通过
- [ ] demo1_full_featured 正常运行
- [ ] demo2_runtime_internals 正常运行
- [ ] demo3_form 正常运行
- [ ] demo4_table 正常运行
- [ ] demo5_scroll 正常运行
- [ ] Wrap 组件 FillWidth 生效
- [ ] 性能无退化 (< 5% slowdown)

---

## 回滚计划

### 触发条件

如果以下情况发生，立即回滚：

1. ❌ 任何核心 demo 应用无法运行
2. ❌ 性能退化超过 10%
3. ❌ 发现无法修复的架构冲突
4. ❌ 测试覆盖率低于 80%

### 回滚步骤

```bash
# 1. 停止开发
git checkout main

# 2. 删除功能分支
git branch -D feature/two-phase-rendering

# 3. 如果已合并，回退 commit
git revert <commit-hash>

# 4. 验证系统恢复正常
go test ./...
go run examples/ui_demos/demo1_full_featured
```

### 应急方案

如果新架构有严重问题但部分功能已依赖：

1. **保留旧渲染路径**
   ```go
   // 在 Renderer 中保留旧路径作为 fallback
   func (r *Renderer) Render(vnode ui.VNode, buffer Buffer) error {
       if os.Getenv("TUI_USE_LEGACY_RENDER") == "true" {
           return r.legacyRender(vnode, buffer)
       }
       return r.newRender(vnode, buffer)
   }
   ```

2. **渐进式迁移**
   - 先在新组件中使用新流程
   - 旧组件保持不变
   - 逐步迁移所有组件

---

## 影响范围

### 需要修改的文件

| 文件 | 类型 | 改动量 | 风险 |
|------|------|--------|------|
| `runtime/compute/layout.go` | 修改 | +100 行 | 中 |
| `runtime/compute/engine.go` | 修改 | +50 行 | 低 |
| `runtime/render/renderer.go` | 修改 | +30 行 | 中 |
| `components/button/button.go` | 修改 | +10 行 | 低 |
| `components/text/text.go` | 修改 | +10 行 | 低 |
| `components/input/input.go` | 修改 | +15 行 | 低 |

### 不需要修改的文件

- ✅ `runtime/constraints.go` - 约束系统
- ✅ `runtime/size.go` - 尺寸计算
- ✅ `runtime/paint/` - 绘制命令系统
- ✅ `components/layout/stack.go` - 布局算法
- ✅ `components/layout/wrap.go` - Wrap 组件

### 兼容性

**向后兼容**: ✅ 是
- 所有现有组件无需修改即可工作
- API 接口无变化
- 用户代码无需改动

**向前兼容**: ✅ 是
- 新架构支持未来扩展
- 可添加更多渲染优化
- 支持异步渲染潜力

---

## 成功标准

### 功能要求

- ✅ Wrap 组件 FillWidth 完全工作
- ✅ Button 按布局宽度拉伸
- ✅ 所有 demo 应用正常运行
- ✅ 无新的 bug 引入

### 性能要求

- ✅ 渲染性能退化 < 5%
- ✅ 内存使用增加 < 10%
- ✅ 首次渲染时间 < 100ms (1000 组件)

### 质量要求

- ✅ 测试覆盖率 > 80%
- ✅ 所有单元测试通过
- ✅ 代码审查通过
- ✅ 文档完整

---

## 附录

### A. 相关文档

- [Wrap 组件实现报告](./wrap_fillwidth_implementation_report.md)
- [布局引擎架构](../architecture/layout_engine.md)
- [渲染系统设计](../architecture/rendering_system.md)

### B. 参考资料

- Flutter 渲染管线: https://api.flutter.dev/flutter/rendering/rendering-library.html
- React Fiber 架构: https://github.com/acdlite/react-fiber-architecture
- Terminal UI 渲染最佳实践

### C. 术语表

| 术语 | 定义 |
|------|------|
| **Measure** | 测量阶段，计算组件的尺寸 |
| **Layout** | 布局阶段，计算组件的位置 |
| **Paint** | 绘制阶段，生成绘制命令 |
| **SetBounds** | 设置组件的边界信息 (x, y, width, height) |
| **ComputedBox** | 已计算尺寸和位置的布局盒子 |
| **VNode** | 虚拟节点，表示组件树中的节点 |

---

**文档维护**: 请在每次重大修改后更新文档版本号和修改日期
**反馈**: 请将问题和建议提交到 GitHub Issues
