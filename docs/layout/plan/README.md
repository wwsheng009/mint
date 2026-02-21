# Mint 布局系统优化计划

## 文档导航

本文档包含了 Mint 布局系统的完整分析、优化方案、测试计划和调试工具设计。

### 文档列表

| 文档 | 描述 | 状态 |
|------|------|------|
| [01-analysis.md](./01-analysis.md) | 详细分析报告，深入探讨当前问题、架构和影响评估 | ✅ |
| [02-optimization.md](./02-optimization.md) | 优化实施方案，包含三个阶段的改进计划 | ✅ |
| [03-testing.md](./03-testing.md) | 测试方案，涵盖单元测试、集成测试和性能测试 | ✅ |
| [04-debug-tools.md](./04-debug-tools.md) | 调试工具开发方案，包含多种调试工具的设计 | ✅ |

---

## 快速开始

### 对于新成员

建议按以下顺序阅读：

1. **首先阅读 [01-analysis.md](./01-analysis.md)** - 了解当前系统的问题和架构
2. **然后阅读 [02-optimization.md](./02-optimization.md)** - 了解优化方案和实施计划
3. **参考 [03-testing.md](./03-testing.md)** - 了解如何编写和运行测试
4. **使用 [04-debug-tools.md](./04-debug-tools.md)** - 学习如何使用调试工具

### 对于开发者

- 如果要修复 bug，先查看 [01-analysis.md](./01-analysis.md) 的相关章节
- 如果要实现新功能，参考 [02-optimization.md](./02-optimization.md) 的设计原则
- 如果要编写测试，参考 [03-testing.md](./03-testing.md) 的测试模板
- 如果遇到问题时，使用 [04-debug-tools.md](./04-debug-tools.md) 中的工具

---

## 问题摘要

### 核心问题

1. **约束传播不一致**：Measure 和 Paint 阶段对约束的处理存在差异
2. **维度语义混淆**：外部/内部维度频繁转换，API 不符合直觉
3. **Text.Wrap 双重约束**：同时受 MaxWidth 和 maxHeight 影响
4. **Auto-measure 语义不清晰**：触发方式和行为不明确
5. **缺少调试工具**：约束传播难以追踪和调试

### 快速修复（已完成）

- ✅ Border 约束传递错误（commit: 6a776c0f）
- ✅ Text 内容溢出（commit: 3c64b3c4）
- ✅ Panel 测试修复（commit: 27951e8a）

---

## 优化计划概览

### 短期优化（Phase 1）- Week 1-2

**目标**：修复已知 bug，完善约束传播逻辑

- [ ] 统一约束传递规则
- [ ] 完善 Text.Wrap 的高度约束验证
- [ ] 实现约束追踪工具
- [ ] 添加边界检查测试

### 中期优化（Phase 2）- Week 3-4

**目标**：改进 API 设计，降低使用复杂度

- [ ] 实现 Panel API 增强（明确内部/外部维度语义）
- [ ] 实现 Builder API 增强（方便的链式调用）
- [ ] 更新文档和示例

### 长期优化（Phase 3）- Week 5-6

**目标**：引入新的布局引擎和可视化工具

- [ ] 设计并实现布局 DSL
- [ ] 实现布局可视化工具
- [ ] 性能优化（Measure 缓存、增量布局）

---

## 测试策略

### 测试金字塔

```
        /\
       /  \        E2E Tests (5%)
      /----\       - 用户场景测试
     /      \      - 跨组件集成
    /--------\     Integration Tests (20%)
   /          \    - 组件间交互
  /            \   - 布局场景
 /--------------\ Unit Tests (75%)
/                \ - 单个组件测试
```

### 测试覆盖率目标

- 单元测试：85%
- 集成测试：70%
- 边界测试：80%

---

## 调试工具

### 可用工具

1. **约束追踪器** (Constraint Tracer)
   - 追踪约束传播链
   - 输出控制台、JSON、HTML 格式
   - 检测约束异常

2. **布局可视化** (Layout Visualizer)
   - 可视化布局树
   - 显示约束和维度
   - 检测布局问题

3. **维度检查器** (Dimension Checker)
   - 检查约束有效性
   - 检查尺寸是否在约束范围内
   - 提供修复建议

4. **性能分析器** (Performance Analyzer)
   - 追踪 Measure/Paint 时间
   - 统计缓存命中率
   - 识别性能瓶颈

5. **交互式调试器** (TUI Debugger)
   - 实时查看布局树
   - 选择节点查看详情
   - 键盘快捷键操作

### 使用方法

```bash
# 环境变量启用调试
export LAYOUT_TRACE=true
export LAYOUT_TRACE_ONLY=panel,text
export LAYOUT_DEBUG_FORMAT=html

# 命令行工具
mint layout-debug trace panel
mint layout-debug visualize hstack
mint layout-debug check panel
mint layout-debug profile panel

# 交互式调试
mint layout-debug interactive
```

---

## 关键知识点

### Panel 组合架构

```
Panel (外部维度)
  │
  └─ Border (内部维度 = 外部 - 边框 padding)
      │
      └─ VStack (继承 Border 维度)
          │
          └─ Flex(Content, 1)  ← Content 占据剩余空间
```

### 维度转换规则

- 外部维度 = 内部维度 + 边框 padding
- 边框 padding = 2 × GetBorderWidth(style)
- 单线/双线边框都占用 1 个字符单元格

### 约束优先级

```
显式维度 > 父约束 > 自动测量
```

---

## 相关提交

```
3c64b3c4 fix: Text content overflow in Panel by respecting layout bounds
6a776c0f fix: Auto-height Panel now correctly measures content height
27951e8a test: Fix Panel component tests to correctly reflect composition architecture
```

---

## 相关代码文件

```
/ui/components/panel/
  vnode.go              - Panel 组合架构
  panel_test.go         - Panel 测试

/ui/components/border/
  instance.go           - Border 的 Measure/Paint
  vnode.go              - Border VNode API
  utils.go              - GetBorderWidth()

/ui/components/text/
  instance.go           - Text 的 Wrap 处理

/ui/components/stack/
  instance.go           - VStack/HStack 的 Flex 布局

/ui/layout/
  constraints/          - 约束系统
    tracer.go           - 约束追踪工具
  visualizer/           - 可视化工具
    tree.go             - 树形图生成
  debug/
    checker.go          - 维度检查器
  performance/
    analyzer.go         - 性能分析器
```

---

## 贡献指南

### 提交代码前的检查

- [ ] 所有现有测试通过
- [ ] 新增功能的测试覆盖率 >= 80%
- [ ] 运行 `go test -race ./ui/components/...` 检查竞态
- [ ] 如果修改了布局逻辑，运行约束追踪验证
- [ ] 文档已更新

### 提交信息格式

```
type: brief description

Detailed explanation of what and why.

Related issues: #123
```

类型 (type):
- `fix` - 修复 bug
- `feat` - 新功能
- `perf` - 性能优化
- `test` - 测试
- `docs` - 文档
- `refactor` - 重构
- `chore` - 杂项

---

## 联系方式

- 仓库地址: `https://github.com/wwsheng009/mint`
- 布局系统问题: [创建 Issue](https://github.com/wwsheng009/mint/issues/new?labels=layout)

---

**文档版本**: 1.0
**最后更新**: 2026-02-21
**维护者**: Qwen Code
