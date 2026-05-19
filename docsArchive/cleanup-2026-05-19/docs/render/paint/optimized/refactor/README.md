# Fiber-First 渲染管线重构方案

## 文档版本
- 版本: 1.0
- 日期: 2024年
- 状态: 待实施

## 目录
1. [重构目标](#重构目标)
2. [总体架构](#总体架构)
3. [实施阶段](#实施阶段)
4. [风险评估](#风险评估)
5. [验收标准](#验收标准)

---

## 重构目标

### 核心原则
```
Fiber 是唯一的运行时实体
VNode 只在 Reconcile 阶段存在
PaintableBox 是渲染单元
```

### 目标指标
| 指标 | 当前 | 目标 | 提升 |
|------|------|------|------|
| 内存占用 | VNode + Fiber + ComputedBox | Fiber + PaintableBox | ~30% ↓ |
| 渲染时间 | VNode 创建 + Layout + Paint | Layout + Paint | ~20% ↓ |
| GC 压力 | 每帧创建 VNode | 仅更新 Fiber | ~40% ↓ |

---

## 总体架构

### 当前架构（3层）
```
用户代码 → renderFn() → VNode 树
                          ↓
                  Fiber Reconciler
                          ↓
                    VNode 树 + Fiber 树
                          ↓
                  Layout (依赖 VNode)
                          ↓
                    ComputedBox
                          ↓
                  Paint (依赖 VNode)
                          ↓
                    Buffer
```

### 目标架构（2层）
```
用户代码 → renderFn() → VNode (临时)
                          ↓
                   Reconciler (协调)
                          ↓
                   Fiber 树 (持久化)
                    ↓         ↓
              Instance    PaintableBox
                          ↓
                   Layout (只读 Fiber)
                          ↓
                   LayoutResult
                          ↓
                   Paint (只用 PaintableBox)
                          ↓
                    Buffer
```

---

## 实施阶段

### Phase 1: Fiber 结构优化 [详情](/docsArchive/render/paint/optimized/refactor/phase1_fiber_structure.md)
**时间**: 1-2 天
**目标**: 清理 Fiber 结构，删除冗余字段

关键任务：
- 删除重复的 ComponentInstance 字段
- 明确 ComputedBox 类型
- 删除 deprecated 字段

### Phase 2: Layout 引擎优化 [详情](/docsArchive/render/paint/optimized/refactor/phase2_layout_engine.md)
**时间**: 2-3 天
**目标**: 实现 Fiber-first Layout

关键任务：
- 实现 FiberLayoutAdapter
- 实现 Fiber → PaintableBox 转换器
- 集成到渲染管线

### Phase 3: Paint 引擎优化 [详情](/docsArchive/render/paint/optimized/refactor/phase3_paint_engine.md)
**时间**: 1-2 天
**目标**: 确保 Paint 完全解耦

关键任务：
- 删除 PaintVNode 方法
- 验证 Paint 只用 PaintableBox

### Phase 4: 渲染管线集成 [详情](/docsArchive/render/paint/optimized/refactor/phase4_rendering_pipeline.md)
**时间**: 3-5 天
**目标**: 重构核心渲染流程

关键任务：
- 重构 DeclarativeNode.Paint()
- 移除 VNode 运行时依赖
- 双轨运行验证

### Phase 5: 组件迁移 [详情](/docsArchive/render/paint/optimized/refactor/phase5_component_migration.md)
**时间**: 5-7 天
**目标**: 所有组件使用新架构

关键任务：
- 迁移基础组件
- 迁移交互组件
- 迁移复杂组件

---

## 风险评估

### 高风险点
| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 组件迁移工作量大 | 高 | 渐进式迁移，保持双轨运行 |
| 布局兼容性 | 高 | 充分测试，对比新旧结果 |
| 性能回退 | 中 | 实时监控，随时回滚 |
| API 兼容性 | 中 | 保持接口稳定，渐进废弃 |

### 回滚策略
每个阶段完成后打 tag，可随时回滚到任意阶段。

---

## 验收标准

### 技术标准
- [ ] VNode 在 commit 后被完全丢弃
- [ ] Layout 只读 Fiber
- [ ] Paint 只用 PaintableBox
- [ ] 所有组件实现 PaintableInstance
- [ ] 性能提升 > 15%

### 功能标准
- [ ] 所有现有功能正常
- [ ] 所有测试通过
- [ ] 示例应用正常运行
- [ ] 无内存泄漏

### 代码标准
- [ ] 无循环依赖
- [ ] 代码覆盖率 > 80%
- [ ] 文档完整
- [ ] 代码审查通过

---

## 文档索引

### 设计文档
- [当前系统分析](/docsArchive/declarative_node_paint_analysis.md)
- [Fiber-First 架构](../FIBER_FIRST_RENDER_PIPELINE.md)
- [实施指南](../IMPLEMENTATION_GUIDE.md)

### 重构步骤
- [Phase 1: Fiber 结构优化](/docsArchive/render/paint/optimized/refactor/phase1_fiber_structure.md)
- [Phase 2: Layout 引擎优化](/docsArchive/render/paint/optimized/refactor/phase2_layout_engine.md)
- [Phase 3: Paint 引擎优化](/docsArchive/render/paint/optimized/refactor/phase3_paint_engine.md)
- [Phase 4: 渲染管线集成](/docsArchive/render/paint/optimized/refactor/phase4_rendering_pipeline.md)
- [Phase 5: 组件迁移](/docsArchive/render/paint/optimized/refactor/phase5_component_migration.md)

### 检查清单
- Phase 1 检查清单（待补充）
- Phase 2 检查清单（待补充）
- Phase 3 检查清单（待补充）
- Phase 4 检查清单（待补充）
- Phase 5 检查清单（待补充）

---

**维护者**: Fiber-first 架构团队
**最后更新**: 2024年
