# Fiber-First 渲染管线文档审查报告

**审查日期**: 2026-02-19
**审查范围**: `docs/render/paint/optimized/` 目录下所有文档
**参考依据**: `docs/fiber/fiber_first/` 目录下设计文档

---

## 一、核心设计原则回顾

根据 `fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md`：

### 1.1 三层架构

| 层级 | 角色 | 生命周期 |
|------|------|----------|
| VNode | 描述 | 短期（render后丢弃） |
| Fiber | 结构 | 长期（运行期持久） |
| paint.PaintableBox | 行为+状态 | 长期（运行期持久） |

### 1.2 禁止的访问路径

```
❌ Layout 访问 VNode
❌ Paint 访问 VNode
❌ Event 访问 VNode
❌ runtime/layout 依赖 Fiber/VNode（应只依赖抽象接口）
❌ Fiber 持有 VNode 引用
❌ Fiber 存储业务回调函数
```

### 1.3 正确的数据流

```
VNode (临时) → Reconcile → Fiber (持久) → Layout → LayoutBox → Paint
```

---

## 二、文档审查结果

### 2.1 README.md ✅ 基本通过

**问题**：
- 多层渲染示例代码使用了反引号而非代码块（语法错误）
- "接口集成设计"章节中 `FiberLayoutAdapter` 位置描述不明确

**建议**：
- 修复 Markdown 语法
- 明确 `FiberLayoutAdapter` 应在 `internal/render` 而非 `runtime/layout`

---

### 2.2 FIBER_FIRST_RENDER_PIPELINE.md ⚠️ 需要修改

**问题 1：接口名称不一致**

文档中使用 `PaintableInstance`，但根据 fiber_first 设计应为 `paint.PaintableBox`。

```go
// 文档中（错误）
Instance paint.PaintableBox

// fiber_first 设计
Instance paint.PaintableBox  // 正确
```

**问题 2：LayoutNode 定义不明确**

文档定义了 `LayoutNode` 结构，但与现有 `layout.LayoutBox` 混淆：

```go
// 文档中定义（新增结构）
type LayoutNode struct {
    Fiber    *Fiber
    Instance paint.PaintableBox
    Box      Box
    Children []*LayoutNode
}

// 应该直接使用 layout.LayoutBox
type LayoutBox struct {
    ID       string
    X, Y     int
    Width    int
    Height   int
    Layer    Layer
    ZIndex   int
    Children []*LayoutBox
}
```

**问题 3：FiberLayoutAdapter 位置错误**

架构图中暗示 `FiberLayoutAdapter` 在 `runtime/layout`，但应该在 `internal/render`。

**建议修改**：
1. 统一使用 `paint.PaintableBox` 作为实例类型名
2. 直接使用 `layout.LayoutBox` 而非自定义 `LayoutNode`
3. 明确 `FiberLayoutAdapter` 在 `internal/render/fiber_adapter.go`

---

### 2.3 IMPLEMENTATION_GUIDE.md ⚠️ 需要修改

**问题 1：DrawCmd 定义位置**

```go
// 文档中定义在 paint 包
package paint
type DrawCmd struct { ... }
```

应确认与现有 `runtime/paint` 包的一致性。

**问题 2：LayoutResult 结构冲突**

文档定义：
```go
type LayoutResult struct {
    Root *LayoutNode
}
```

应使用现有的 `layout.LayoutResult`：
```go
type LayoutResult struct {
    Boxes []LayoutBox
    Root  *LayoutBox
    ContentSize Size
    Dirty bool
    HitMap *HitMap
}
```

**问题 3：FiberLayoutEngine 签名错误**

```go
// 文档中（错误）
func (e *FiberLayoutEngine) LayoutFiber(root *Fiber, constraints Constraints) *LayoutResult

// 应该使用 runtime.BoxConstraints
func (e *FiberLayoutEngine) LayoutFiber(root *Fiber, constraints runtime.BoxConstraints) *layout.LayoutResult
```

---

### 2.4 FIBER_FIRST_MIGRATION_GUIDE.md ✅ 基本通过

**问题**：
- 多处使用反引号导致 Markdown 渲染问题
- 代码块语言标识符使用了 `\`go` 而非 ```go

**建议**：
- 修复 Markdown 语法错误

---

### 2.5 DETAILED_IMPLEMENTATION_PLAN.md ✅ 通过

内容简洁，与设计原则一致。

---

### 2.6 refactor/phase1_fiber_structure.md ⚠️ 需要修改

**问题 1：ComputedBox 类型建议**

文档建议使用 `*paint.PaintableBox`，但根据 fiber_first 设计，布局结果应该是 `*layout.LayoutBox`：

```go
// 文档建议（方案 B）
LayoutBox *paint.PaintableBox

// 更正确的做法
// Fiber 存储布局输入，不存储布局结果
// 布局结果由 LayoutEngine 输出为 layout.LayoutBox
```

**问题 2：删除 FocusableMeta 的前提条件**

文档要求删除 `FocusableMeta`，但需要确认所有 Focusable 操作已迁移到 Instance。

**建议**：
- 明确 Fiber 只存储布局输入（Style），不存储布局输出
- 布局结果由 `layout.LayoutBox` 表示，不存储在 Fiber 中

---

### 2.7 refactor/phase2_layout_engine.md ✅ 已修改通过

之前的审查已修复了主要问题：
- 移除了 `runtime/layout/adapter_fiber.go` 的创建计划
- 改为完善 `internal/render/fiber_adapter.go`
- 明确了 `runtime/layout` 不依赖 Fiber 的架构约束

---

### 2.8 refactor/phase3_paint_engine.md ✅ 基本通过

**问题**：
- PaintEngine 应确保只接受 `PaintableLayout`，不访问 VNode

**建议**：
- 添加架构约束说明，明确 Paint 阶段禁止访问 VNode

---

### 2.9 refactor/phase4_rendering_pipeline.md ⚠️ 需要修改

**问题 1：LayoutNode 与 LayoutBox 混用**

```go
// 文档中使用 LayoutNode（自定义）
type LayoutNode struct {
    Fiber    *Fiber
    Instance paint.PaintableBox
    Box      Box
    Children []*LayoutNode
}

// 应统一使用 layout.LayoutBox
```

**问题 2：FiberLayoutEngine 方法签名**

```go
// 文档中
func (e *FiberLayoutEngine) LayoutFiberAndConvert(root *rtui.Fiber, constraints layout.Constraints) *paint.PaintableLayout

// 应该使用 runtime.BoxConstraints
func (e *FiberLayoutEngine) LayoutFiberAndConvert(root *rtui.Fiber, constraints runtime.BoxConstraints) *paint.PaintableLayout
```

**问题 3：hasMultipleLayers 方法中的类型错误**

```go
// 文档中
if box.Layer > int(baseLayer) { ... }

// baseLayer 是 rtui.Layer 类型，需要正确比较
```

---

### 2.10 refactor/phase5_component_migration.md ⚠️ 需要修改

**问题 1：VNode 仍包含 Intent 字段**

根据 fiber_first 设计，VNode 不应持有 Intent，应使用 ActionTargetID：

```go
// 文档中（不符合设计）
type VNode struct {
    actionIntent intent.Intent  // ❌
}

// 正确做法
type VNode struct {
    // Intent 通过 ActionTargetID 路由，不存储在 VNode
}
```

**问题 2：Instance 中意图发射器定义**

```go
// 文档中
intentEmitter func(intent.Intent)

// 根据 fiber_first 设计，事件应通过 ActionBridge 路由
// Instance 应实现 ActionHandlerInstance 接口
```

**问题 3：组件迁移目标位置不一致**

文档建议迁移到 `ui/components/`，但需要确认与现有组件结构的一致性。

---

## 三、架构一致性问题汇总

### 3.1 关键问题

| 问题 | 影响 | 涉及文档 |
|------|------|----------|
| LayoutNode vs LayoutBox 混用 | 高 | pipeline, phase4, implementation |
| FiberLayoutAdapter 位置错误 | 高 | README, pipeline |
| 约束类型不一致 (layout.Constraints vs runtime.BoxConstraints) | 中 | 多个文档 |
| VNode 中存储 Intent | 中 | phase5 |

### 3.2 接口命名不一致

| 文档使用 | 应使用 | 说明 |
|----------|--------|------|
| LayoutNode | layout.LayoutBox | 统一使用现有类型 |
| layout.Constraints | runtime.BoxConstraints | 输入约束 |
| PaintableInstance | paint.PaintableBox | 实例类型 |

### 3.3 缺失的架构约束说明

以下文档应添加架构约束章节：
- `phase3_paint_engine.md` - Paint 禁止访问 VNode
- `phase4_rendering_pipeline.md` - 渲染管线的 VNode 生命周期
- `phase5_component_migration.md` - 组件的 VNode 禁止存储闭包

---

## 四、修改建议

### 4.1 高优先级修改

1. **统一使用 layout.LayoutBox**
   - 删除自定义 `LayoutNode` 结构
   - 所有布局结果使用 `layout.LayoutBox`
   - 转换器负责将 `LayoutBox` 转为 `PaintableBox`

2. **修正 FiberLayoutAdapter 位置**
   - 明确适配器在 `internal/render/fiber_adapter.go`
   - `runtime/layout` 只定义接口，不依赖 Fiber

3. **修正约束类型**
   - 输入约束使用 `runtime.BoxConstraints`
   - 布局引擎内部使用 `layout.Constraints`
   - 添加转换函数

### 4.2 中优先级修改

1. **修复 Markdown 语法错误**
   - `FIBER_FIRST_MIGRATION_GUIDE.md`
   - `README.md`

2. **添加架构约束章节**
   - 每个 phase 文档添加"架构约束"章节
   - 明确禁止的访问路径

### 4.3 低优先级修改

1. **统一代码示例风格**
2. **添加更多调试和监控示例**

---

## 五、正确的数据流和组件关系

### 5.1 完整数据流

```
┌─────────────────────────────────────────────────────────────┐
│                    Fiber-First 正确数据流                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  用户代码 → renderFn() → VNode (临时，仅存在于 Reconcile)   │
│                           ↓                                 │
│                   Reconciler (协调)                         │
│                           ↓                                 │
│                   Fiber 树 (持久化)                         │
│                    ↓         ↓                              │
│              Style 数据   Instance (paint.PaintableBox)     │
│                    ↓         ↓                              │
│              ┌────────────────────────┐                     │
│              │  internal/render 层    │                     │
│              │  FiberToNodeAdapter    │ ← 实现 layout.Node  │
│              └────────────┬───────────┘                     │
│                           ↓                                 │
│              ┌────────────────────────┐                     │
│              │   runtime/layout 层    │                     │
│              │   Engine.Layout()      │ ← 纯布局算法        │
│              │   → LayoutResult       │                     │
│              │   → LayoutBox 树       │                     │
│              └────────────┬───────────┘                     │
│                           ↓                                 │
│              ┌────────────────────────┐                     │
│              │   internal/render 层    │                    │
│              │   FiberToPaintable     │                     │
│              │   Converter            │                     │
│              │   → PaintableLayout    │                     │
│              └────────────┬───────────┘                     │
│                           ↓                                 │
│              ┌────────────────────────┐                     │
│              │   PaintEngine          │                     │
│              │   PaintLayout()        │ ← 只用 PaintableBox │
│              │   → Buffer             │                     │
│              └────────────────────────┘                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 层级依赖关系

```
┌──────────────────────────────────────┐
│         Application Layer            │
│  (internal/render, internal/reconciler)│
│                                      │
│  - Fiber 数据结构                    │
│  - FiberToNodeAdapter                │
│  - FiberToPaintableConverter         │
│  - 业务逻辑                          │
└──────────────┬───────────────────────┘
               │ 依赖
               ↓
┌──────────────────────────────────────┐
│       Infrastructure Layer           │
│        (runtime/layout)              │
│                                      │
│  - 抽象接口 (Node, Measurable...)    │
│  - 纯布局算法                        │
│  - LayoutBox (输出)                  │
│  - ❌ 不依赖 Fiber/VNode             │
└──────────────────────────────────────┘
```

---

## 六、检查清单

### 文档修改检查

- [ ] README.md - 修复 Markdown 语法，明确适配器位置
- [ ] FIBER_FIRST_RENDER_PIPELINE.md - 统一 LayoutNode → LayoutBox
- [ ] IMPLEMENTATION_GUIDE.md - 修正类型和签名
- [ ] FIBER_FIRST_MIGRATION_GUIDE.md - 修复 Markdown 语法
- [ ] phase1_fiber_structure.md - 明确布局结果不存 Fiber
- [ ] phase3_paint_engine.md - 添加架构约束
- [ ] phase4_rendering_pipeline.md - 统一类型，添加约束
- [ ] phase5_component_migration.md - 修正 Intent 使用方式

### 架构约束检查

- [ ] runtime/layout 不导入 runtime/ui
- [ ] runtime/layout 不导入 internal/reconciler
- [ ] Layout 阶段不访问 VNode
- [ ] Paint 阶段不访问 VNode
- [ ] Fiber 不持有 VNode 引用

---

**审查完成日期**: 2026-02-19
**审查人**: AI 助手
**状态**: 待修改
