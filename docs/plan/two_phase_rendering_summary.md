# 两阶段渲染重构 - 执行摘要

**状态**: 设计阶段，待审批
**预计工期**: 8-13 天
**风险等级**: 中-高
**优先级**: P1 (高优先级)

---

## 🎯 核心目标

**一句话**: 修复渲染时序，实现 `Measure → Layout → Paint` 的正确渲染管线

**当前问题**:
```
❌ Measure → Paint → Layout/SetBounds (顺序错误)
   ↓              ↓
计算完成       绘制时 bounds 未设置
```

**目标状态**:
```
✅ Measure → Layout/SetBounds → Paint (顺序正确)
   ↓              ↓                ↓
计算完成    bounds 已设置      使用正确 bounds 绘制
```

---

## 💡 技术方案

### 核心改动 (3 个文件)

#### 1. `runtime/compute/layout.go` (+100 行)
新增 `ComputedLayout.Paint()` 方法：
```go
func (cl *ComputedLayout) Paint() []paint.DrawCmd {
    return cl.paintBox(cl.Root, 0, 0)
}
```

#### 2. `runtime/render/renderer.go` (~30 行修改)
修改 `Render()` 方法调用顺序：
```go
func (r *Renderer) Render(vnode ui.VNode, buffer Buffer) error {
    // Phase 1: Measure + Layout
    layout, _ := r.engine.Layout(vnode, constraints)

    // Phase 2: Paint (AFTER bounds are set)
    cmds := layout.Paint()
    r.applyCommands(cmds, buffer)

    return nil
}
```

#### 3. `components/button/button.go` (确认逻辑)
确保 `Paint()` 使用 bounds：
```go
if b.bounds[2] > 0 {
    layoutWidth := b.bounds[2]
    // Pad text to fill layout width
    buttonText += strings.Repeat(" ", layoutWidth - buttonWidth)
}
```

---

## 📅 实施计划

### 阶段 1: 准备 (1-2 天)
- [ ] 创建功能分支
- [ ] 建立测试基线
- [ ] 代码审查

### 阶段 2: 核心实施 (3-5 天)
- [ ] 新增 `ComputedLayout.Paint()`
- [ ] 修改 `Renderer.Render()`
- [ ] 确认 `Engine.Layout()` 完整性
- [ ] 单元测试

### 阶段 3: 组件适配 (2-3 天)
- [ ] Button, Input, Text 组件
- [ ] 布局组件 (HStack, VStack, Grid)
- [ ] 第三方组件

### 阶段 4: 集成测试 (2-3 天)
- [ ] 回归测试
- [ ] 功能测试 (Wrap FillWidth)
- [ ] 性能测试

### 阶段 5: 发布 (1-2 天)
- [ ] 文档更新
- [ ] 代码审查
- [ ] 合并发布

**总计**: 9-15 天

---

## ⚠️ 风险评估

### 高风险 🔴
- **现有组件依赖旧时序** - 可能导致功能中断
- **Fiber 渲染路径不兼容** - 可能需要保留旧路径

### 缓解措施
- ✅ 全面的回归测试套件
- ✅ 保留旧渲染路径作为 fallback
- ✅ 渐进式迁移策略

---

## ✅ 成功标准

### 功能
- ✅ Wrap FillWidth 完全工作
- ✅ Button 按布局宽度拉伸
- ✅ 所有 demo 应用正常运行

### 性能
- ✅ 渲染性能退化 < 5%
- ✅ 内存使用增加 < 10%

### 质量
- ✅ 测试覆盖率 > 80%
- ✅ 所有单元测试通过

---

## 🔄 回滚计划

### 触发条件
- 任何核心 demo 失效
- 性能退化 > 10%
- 发现无法修复的架构冲突

### 回滚命令
```bash
git checkout main
git branch -D feature/two-phase-rendering
```

---

## 📊 影响范围

### 需要修改的文件 (5 个核心文件)
| 文件 | 改动量 | 风险 |
|------|--------|------|
| `runtime/compute/layout.go` | +100 行 | 中 |
| `runtime/render/renderer.go` | +30 行 | 中 |
| `runtime/compute/engine.go` | +50 行 | 低 |
| `components/button/button.go` | +10 行 | 低 |
| `components/text/text.go` | +10 行 | 低 |

### 不需要修改
- ✅ 约束系统 (`runtime/constraints.go`)
- ✅ 布局算法 (`components/layout/stack.go`)
- ✅ Wrap 组件 (`components/layout/wrap.go`)
- ✅ API 接口

---

## 🎯 建议行动

1. **立即行动**:
   - [ ] 审查设计方案
   - [ ] 批准实施计划
   - [ ] 创建功能分支

2. **本周完成**:
   - [ ] 准备阶段
   - [ ] 核心实施开始

3. **跟踪进度**:
   - 使用 GitHub Projects 跟踪任务
   - 每日站会同步进度
   - 每周风险评审

---

**完整文档**: [two_phase_rendering_refactor.md](./two_phase_rendering_refactor.md)
**负责人**: 待指定
**审查人**: 待指定
**批准人**: 待指定
