# Deprecated 代码安全移除分析

## 概述

本文档记录已标记 deprecated 的代码的移除情况。

---

## 已完成移除

### ✅ 2026-02-21 清理完成

#### 已删除的文件
| 文件 | 原因 |
|------|------|
| `examples/engine_test/main.go` | 仅为测试 LayoutSwitcher |
| `internal/render/layout_switcher_test.go` | LayoutSwitcher 的测试 |

#### 已删除的代码（layout_switcher.go）

| 代码 | 行数 | 原因 |
|------|------|------|
| `LayoutEngineType` 类型及常量 | ~50 | 不再需要引擎切换 |
| `ParseLayoutEngineType()` | ~15 | 不再需要 |
| `LayoutEngine` 接口 | ~25 | 不再需要 |
| `ComputeEngineAdapter` | ~70 | 被 NewLayoutEngineAdapter 替代 |
| `computeLayoutResultAdapter` | ~30 | 不再需要 |
| `computedBoxAdapter` | ~25 | 不再需要 |
| `LayoutSwitcher` | ~200 | 被 NewLayoutEngineAdapter 替代 |
| `SwitcherStats` | ~15 | LayoutSwitcher 的统计 |
| `abs()` 辅助函数 | ~5 | LayoutSwitcher 内部使用 |
| Deprecated 注释块 | ~20 | 清理 |
| **总计** | **~493 行** | |

#### 已修改的文件

| 文件 | 修改内容 |
|------|----------|
| `rendering_pipeline.go` | 移除 `switcher`/`useSwitcher` 字段、简化 `Render()` 和 `RenderLayers()`、移除 `MINT_LAYOUT_ENGINE` 检查 |
| `declarative_node.go` | 移除未使用的 `layoutSwitcher` 字段 |

---

## 保留的代码

### NewLayoutEngineAdapter（仍在使用）

**用途:** Fiber-first 渲染管道的布局引擎适配器

**调用位置:**
- `declarative_node.go:175,194,390` - fiberFirstPaint 中使用
- `runtime/ui/fiber_render_pipeline_test.go:392` - 测试

**保留原因:** 这是 Fiber-first 架构的核心组件

### legacyPaint（保留作为回退）

**用途:** 作为 fiberFirstPaint 的安全回退

**调用位置:**
- `declarative_node.go` 多处 - 当 Fiber-first 失败时回退

**保留原因:** 确保渲染稳定性

### NewDeclarativeNodeFromFunc（保留用于测试）

**用途:** 测试辅助函数

**调用位置:**
- 27 处测试文件

**保留原因:** 大量测试依赖

### VNodeToNodeAdapter（保留用于回退）

**用途:** 当没有 Fiber 时将 VNode 转换为 layout.Node

**保留原因:** legacy 路径仍需要

---

## 架构简化

### 移除前
```
RenderingPipeline
├── layoutEngine (compute.Engine)
├── switcher (LayoutSwitcher) ← 已删除
│   ├── computeEngine (ComputeEngineAdapter) ← 已删除
│   └── newEngine (NewLayoutEngineAdapter)
└── paintEngine
```

### 移除后
```
RenderingPipeline
├── layoutEngine (compute.Engine)
└── paintEngine

Fiber-first 路径（DeclarativeNode.fiberFirstPaint）:
├── newLayoutEngine (NewLayoutEngineAdapter) ← 直接使用
├── paintEngine
└── converter (FiberToPaintableConverter)
```

---

## 环境变量清理

| 环境变量 | 状态 |
|----------|------|
| `MINT_LAYOUT_ENGINE` | ❌ 已移除（不再支持） |
| `MINT_FIBER_FIRST` | ✅ 保留（控制 Fiber-first 模式） |

---

## 验证状态

- ✅ 编译通过 (`go build ./...`)
- ✅ 测试结果与修改前一致（预先存在的失败与本次修改无关）
- ✅ Fiber-first demo 编译通过

---

## 历史记录

| 日期 | 操作 | 删除行数 |
|------|------|----------|
| 2026-02-21 | 删除 engine_test, layout_switcher_test | ~200 行 |
| 2026-02-21 | 删除 ParallelRenderingPipeline, BenchmarkResult | ~226 行 |
| 2026-02-21 | 删除 LayoutSwitcher, ComputeEngineAdapter 等废弃类型 | ~493 行 |
| 2026-02-21 | 简化 RenderingPipeline | ~150 行 |
| **总计** | | **~1069 行** |
