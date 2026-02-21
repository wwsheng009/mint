# Deprecated 代码安全移除分析

## 概述

本文档分析已标记 deprecated 的代码的调用链路，评估是否可以安全移除。

---

## 一、调用链路分析

### 1. LayoutSwitcher

**定义位置:** `internal/render/layout_switcher.go:468`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `examples/engine_test/main.go` | 39 | 测试示例 | ✅ 可删除 |
| `internal/render/rendering_pipeline.go` | 44 | 条件创建 (`MINT_LAYOUT_ENGINE` 非空时) | ⚠️ 需修改 |
| `internal/render/layout_switcher.go` | 733 | ParallelRenderingPipeline 内部 | ✅ 可一起删除 |
| `internal/render/layout_switcher_test.go` | 多处 | 测试文件 | ✅ 可删除 |

**移除建议:**
1. 删除 `examples/engine_test/main.go` (测试示例)
2. 删除 `internal/render/layout_switcher_test.go` (测试文件)
3. 修改 `rendering_pipeline.go` 移除 switcher 相关代码
4. 删除 `layout_switcher.go` 中的 LayoutSwitcher、ParallelRenderingPipeline 等

---

### 2. ComputeEngineAdapter

**定义位置:** `internal/render/layout_switcher.go:152`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `internal/render/layout_switcher.go` | 503 | LayoutSwitcher 内部 | ✅ 随 LayoutSwitcher 删除 |
| `internal/render/layout_switcher_test.go` | 172 | 测试 | ✅ 随测试文件删除 |

**移除建议:** 随 LayoutSwitcher 一起删除

---

### 3. ParallelRenderingPipeline

**定义位置:** `internal/render/layout_switcher.go:726`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `internal/render/layout_switcher_test.go` | 139,147,162 | 测试 | ✅ 随测试文件删除 |

**移除建议:** 随 LayoutSwitcher 一起删除

---

### 4. legacyPaint

**定义位置:** `internal/render/declarative_node.go:537`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `declarative_node.go` | 343 | Paint() 默认回退 | ⚠️ 需保留 |
| `declarative_node.go` | 370 | fiberFirstPaint 回退 (Fiber root 为 nil) | ⚠️ 需保留 |
| `declarative_node.go` | 398 | fiberFirstPaint 回退 (layout 错误) | ⚠️ 需保留 |
| `declarative_node.go` | 442 | fiberFirstPaint 回退 (paint 错误) | ⚠️ 需保留 |
| `declarative_node.go` | 454 | fiberFirstPaint 最终回退 | ⚠️ 需保留 |
| `declarative_node.go` | 515 | comparePaint 内部调用 | ⚠️ 随 comparePaint 删除 |

**移除建议:** **暂不可移除** - 作为 fiberFirstPaint 的安全回退，需要保留

---

### 5. comparePaint

**定义位置:** `internal/render/declarative_node.go:507`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `declarative_node.go` | 335 | RenderModeBoth 模式调用 | ⚠️ 需同时移除 RenderModeBoth |

**移除建议:** 可移除，但需同时移除 RenderModeBoth 相关代码

---

### 6. VNodeToNodeAdapter

**定义位置:** `internal/render/fiber_adapter.go:688`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `layout_switcher.go` | 295 | NewLayoutEngineAdapter 的 VNode fallback | ✅ 随 LayoutSwitcher 删除 |
| `fiber_adapter.go` | 718 | 递归初始化子节点 | ⚠️ 需修改 |

**移除建议:** 
- 随 LayoutSwitcher 删除后，VNode fallback 路径不再需要
- 但需保留 `VNodeToNodeAdapter` 定义，因为 `fiber_adapter.go` 内部递归使用
- **或者** 完全重构，移除 VNode 路径

---

### 7. NonFiberRenderer

**定义位置:** `internal/render/vnode_renderer.go:42`

**调用位置:**

| 文件 | 行号 | 用途 | 可移除? |
|------|------|------|---------|
| `declarative_node.go` | 99 | 已注释掉的代码 | ✅ 可删除 |

**移除建议:** 可安全删除，目前只有注释掉的代码使用

---

### 8. NewDeclarativeNodeFromFunc

**定义位置:** `internal/render/declarative_node.go:81`

**调用位置:**

| 文件类型 | 数量 | 可移除? |
|----------|------|---------|
| 测试文件 (`*_test.go`) | 27 处 | ⚠️ 需迁移测试 |
| `declarative_node.go` 内部 | 定义 + deprecated 注释 | 保留定义 |

**移除建议:** **保留** - 大量测试使用，需要保留用于向后兼容

---

## 二、可安全移除的代码

### 阶段 1: 立即可删除

| 代码 | 文件 | 原因 |
|------|------|------|
| `examples/engine_test/` | 整个目录 | 仅为测试 LayoutSwitcher |
| `layout_switcher_test.go` | 整个文件 | LayoutSwitcher 的测试 |

### 阶段 2: 需要修改后删除

| 代码 | 依赖 | 修改方案 |
|------|------|----------|
| `LayoutSwitcher` | `rendering_pipeline.go` | 移除 switcher 字段和相关逻辑 |
| `ParallelRenderingPipeline` | 无 | 直接删除 |
| `ComputeEngineAdapter` | LayoutSwitcher 内部 | 直接删除 |
| `RenderModeBoth` | `declarative_node.go` | 移除常量和 comparePaint 调用 |
| `comparePaint()` | RenderModeBoth | 移除方法 |

### 阶段 3: 暂不可移除

| 代码 | 原因 |
|------|------|
| `legacyPaint()` | fiberFirstPaint 的安全回退 |
| `NewDeclarativeNodeFromFunc()` | 大量测试依赖 |
| `VNodeToNodeAdapter` | 递归使用，需重构 |

---

## 三、推荐的移除步骤

### 步骤 1: 删除测试和示例 (低风险)

```bash
# 删除文件
rm examples/engine_test/main.go
rm internal/render/layout_switcher_test.go
```

### 步骤 2: 修改 rendering_pipeline.go

移除 switcher 相关代码:
- 删除 `switcher *LayoutSwitcher` 字段
- 删除 `useSwitcher bool` 字段
- 删除 `NewRenderingPipeline()` 中的 switcher 创建逻辑
- 简化 `Render()` 方法，直接使用 `layoutEngine`

### 步骤 3: 删除 layout_switcher.go 中的废弃代码

保留:
- `NewLayoutEngineAdapter` (Fiber-first 使用)
- `newLayoutResultAdapter`
- `layoutBoxAdapter`

删除:
- `LayoutEngineType` 及常量
- `LayoutSwitcher`
- `ComputeEngineAdapter`
- `ParallelRenderingPipeline`
- `BenchmarkResult` / `BenchmarkRender`

### 步骤 4: 移除 RenderModeBoth

修改 `declarative_node.go`:
- 删除 `RenderModeBoth` 常量
- 删除 `comparePaint()` 方法
- 简化 `Paint()` 方法

---

## 四、风险评估

| 操作 | 风险等级 | 说明 |
|------|----------|------|
| 删除 `examples/engine_test/` | 🟢 低 | 独立测试示例 |
| 删除 `layout_switcher_test.go` | 🟢 低 | 仅测试文件 |
| 删除 `ParallelRenderingPipeline` | 🟢 低 | 无外部调用 |
| 删除 `ComputeEngineAdapter` | 🟡 中 | 被 LayoutSwitcher 使用 |
| 删除 `LayoutSwitcher` | 🟡 中 | 被 rendering_pipeline 使用 |
| 删除 `comparePaint` | 🟡 中 | 需移除 RenderModeBoth |
| 删除 `legacyPaint` | 🔴 高 | 作为安全回退 |
| 删除 `NewDeclarativeNodeFromFunc` | 🔴 高 | 大量测试依赖 |

---

## 五、结论

**立即可移除的代码:**
1. `examples/engine_test/main.go`
2. `internal/render/layout_switcher_test.go`
3. `ParallelRenderingPipeline` (layout_switcher.go 中)

**需要重构后移除的代码:**
1. `LayoutSwitcher` - 需修改 `rendering_pipeline.go`
2. `ComputeEngineAdapter` - 随 LayoutSwitcher 移除
3. `RenderModeBoth` + `comparePaint()` - 简化 Paint 逻辑

**建议保留的代码:**
1. `legacyPaint()` - 作为 fiberFirstPaint 的安全回退
2. `NewDeclarativeNodeFromFunc()` - 测试兼容性
3. `VNodeToNodeAdapter` - 递归使用，需单独重构

---

## 六、下一步行动

1. [ ] 删除 `examples/engine_test/main.go`
2. [ ] 删除 `internal/render/layout_switcher_test.go`
3. [ ] 重构 `rendering_pipeline.go` 移除 switcher
4. [ ] 删除 `layout_switcher.go` 中的废弃类型
5. [ ] 移除 `RenderModeBoth` 和 `comparePaint()`
6. [ ] 更新文档
