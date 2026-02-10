# 事件系统重构文档索引

> **最后更新**: 2025-02-10
> **状态**: 规划完成，待实施

---

## 📚 文档结构

本目录包含 Mint TUI 事件系统重构的完整规划和实施方案。

```
docs/plan/event/
├── README.md                    (本文件)
├── COMPREHENSIVE_ANALYSIS_2025.md  # 综合分析报告
├── IMPLEMENTATION_PLAN.md        # 实施方案详解
├── MSG_UNIFICATION.md            # Msg 统一与 Sandbox 集成
└── TASKS.md                      # 完整任务列表
```

---

## 🚀 快速导航

### 如果你想...

**了解问题全貌**
→ 阅读 [综合分析报告](./COMPREHENSIVE_ANALYSIS_2025.md)
  - 三个文档对比
  - 根本原因分析
  - 预期收益评估

**查看实施细节**
→ 阅读 [实施方案详解](./IMPLEMENTATION_PLAN.md)
  - 6 个阶段详细步骤
  - 代码示例
  - 验收标准

**了解 Msg/Cmd 设计**
→ 阅读 [Msg 统一设计](./MSG_UNIFICATION.md)
  - Msg 类型定义
  - Cmd 副作用系统
  - Sandbox 注入 API

**查看具体任务**
→ 阅读 [任务列表](./TASKS.md)
  - 47 个任务分解
  - 优先级排序
  - 时间估算

---

## 📊 执行摘要

### 核心问题
Inspector overlay 无法响应鼠标事件的根本原因：
1. **命中测试分散** - 各组件手写 bounds
2. **Event → Action 缺失** - 缺少语义化层
3. **异步操作分散** - 无统一 Cmd 管理

### 推荐方案
分 **6 个阶段** 重构，预计 **7 周** 完成：

| 阶段 | 内容 | 工期 |
|------|------|------|
| Phase 0 | 前置修复 | 1-2 天 ✅ |
| Phase 1 | HitMap 系统 | 2 周 |
| Phase 2 | Action 系统 | 2 周 |
| Phase 3 | Router 三阶段 | 1 周 |
| Phase 4 | Msg/Cmd 系统 | 1 周 |
| Phase 5 | 测试与工具 | 持续 |
| Phase 6 | 性能优化 | 1 周 |

### 预期收益

| 指标 | 当前 | 目标 | 提升 |
|------|------|------|------|
| Inspector 鼠标事件成功率 | ~30% | 100% | +233% |
| 组件需手写 bounds | 每个组件 | 0 个 | -100% |
| 测试用例坐标硬编码 | 90% | 0% | -100% |
| 按 ID 注入事件测试 | 不支持 | 100% | +∞ |

---

## 🎯 下一步行动

### 立即开始（本周）

**Phase 0 验收**（已完成代码修复，待测试）：
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go

# 测试清单：
# [ ] 按 F12 打开/关闭 Inspector
# [ ] 点击 tab 栏切换标签
# [ ] 移动鼠标查看 hover 信息
# [ ] 鼠标坐标显示正确
```

**Phase 1 实施**（Week 1-2）：
1. 创建 `runtime/event/hitmap.go`
2. 扩展 `framework/event/mouse_event.go`
3. 修改 `framework/app.go` 集成 HitMap
4. 编写单元测试

### 查看详细任务

→ [任务列表](./TASKS.md) - 47 个任务，优先级排序

---

## 📝 文档版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v1.0 | 2025-02-10 | 初始版本，完整规划 |

---

## 🔗 相关文档

### 框架设计
- [framework/docs/EVENT_SYSTEM.md](../../framework/docs/EVENT_SYSTEM.md) - V3 事件系统规范
- [framework/docs/ui/idea/idea4.1_event.md](../../framework/docs/ui/idea/idea4.1_event.md) - 事件系统设计理念
- [docs/event/long_term_event_architecture.md](../../event/long_term_event_architecture.md) - 长期架构规划

### 当前实现
- [runtime/event/hittest.go](../../runtime/event/hittest.go) - 当前命中测试实现
- [runtime/event/dispatch.go](../../runtime/event/dispatch.go) - 当前事件分发
- [internal/inspector/standalone_inspector.go](../../internal/inspector/standalone_inspector.go) - Inspector 实现

---

## ✅ 实施前检查清单

在开始实施前，请确认：

- [ ] 已阅读并理解 [综合分析报告](./COMPREHENSIVE_ANALYSIS_2025.md)
- [ ] 已查看 [实施方案详解](./IMPLEMENTATION_PLAN.md)
- [ ] 已了解 [Msg 统一设计](./MSG_UNIFICATION.md)
- [ ] 已审阅 [任务列表](./TASKS.md)
- [ ] 已备份当前代码（或创建新分支）
- [ ] 已设置合理的测试环境

---

**准备就绪？** 让我们从 [Phase 1 的第一个任务](./TASKS.md#phase-1-hitmap-系统week-1-2) 开始！
