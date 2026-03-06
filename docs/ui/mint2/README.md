# Mint Runtime 2.0 架构升级

> **状态**: 🚧 架构设计完成，待实施
> **目标**: 建立完整的组件通信基础设施，支持组合组件和高级 UI 特性

---

## 📋 概述

Mint Runtime 2.0 是一次系统级架构升级，旨在解决 Mint UI 核心缺失的组件通信能力。

### 核心发现

OptionGroup 组件暴露的问题**不是 Bug**，而是揭示了 Mint 缺少以下核心能力：

- ❌ **Instance Tree** - 组件逻辑树（父子关系）
- ❌ **Context System** - 依赖注入（解决 Props Drilling）
- ❌ **Intent Bubble** - 事件冒泡（解耦父子通信）

这三个能力的缺失会影响所有组合组件的实现：
- 🔴 Form（字段联动和验证）
- 🔴 Menu（子菜单展开/收起）
- 🔴 Tabs（动态 Tab 管理）
- 🔴 Tree（展开/折叠逻辑）
- 🟡 Portal（Dialog、Tooltip、Overlay）

---

## 🏗️ 架构设计

### Mint Runtime 2.0 核心架构

```
┌─────────────────────────────────────────────────────────┐
│                    Mint Runtime 2.0                     │
├─────────────────────────────────────────────────────────┤
│  VNode (声明式 UI)                                      │
│          ↓                                              │
│  Fiber Runtime (调度)                                   │
│          ↓                                              │
│  Component Instance (组件)                              │
│          ↓                                              │
│  Instance Tree (逻辑组件树) ──────────────────────┐    │
│          ↓                                         │    │
│  ┌─────────────────────────────────────────────┐ │    │
│  │  Context System (依赖注入)                   │ │    │
│  │  Intent/Event System (事件冒泡)              │ │    │
│  │  Layout Tree (渲染树)                        │ │    │
│  │  Portal Layer (门户层)                       │ │    │
│  └─────────────────────────────────────────────┘ │    │
└────────────────────────────────────────────────────┼────┘
                                                     ↓
                                            ┌────────┐
                                            │ Store  │
                                            │Reducer │
                                            └────────┘
```

### 三层通信模型

| 层级 | 能力 | 用途 | 文档 |
|------|------|------|------|
| **1️⃣ Instance Tree** | 组件逻辑树 | 父子关系、生命周期 | [INSTANCE_TREE_IMPLEMENTATION.md](./INSTANCE_TREE_IMPLEMENTATION.md) |
| **2️⃣ Context System** | 依赖注入 | 跨层级访问、Provider | [CONTEXT_IMPLEMENTATION.md](./CONTEXT_IMPLEMENTATION.md) |
| **3️⃣ Intent Bubble** | 事件冒泡 | 解耦通信、事件拦截 | [INTENT_BUBBLE_IMPLEMENTATION.md](./INTENT_BUBBLE_IMPLEMENTATION.md) |

---

## 📚 文档导航

### 核心文档

1. **[MINT_RUNTIME_2.0_ARCHITECTURE.md](./MINT_RUNTIME_2.0_ARCHITECTURE.md)**
   - 总体架构分析报告
   - 当前架构缺陷分析
   - 三层通信模型设计
   - 实施路线图和优先级

2. **[INSTANCE_TREE_IMPLEMENTATION.md](./INSTANCE_TREE_IMPLEMENTATION.md)** (Phase 1)
   - BaseInstance 和 ComponentInstance 接口设计
   - Fiber 创建逻辑修改（mountInstance/unmountInstance）
   - OptionGroup/Option 组件更新示例
   - 单元测试和集成测试
   - 验证检查清单

3. **[CONTEXT_IMPLEMENTATION.md](./CONTEXT_IMPLEMENTATION.md)** (Phase 2)
   - FiberContext 核心接口
   - Context Provider 组件设计
   - Fiber 集成方案
   - 泛型类型安全
   - 测试用例

4. **[INTENT_BUBBLE_IMPLEMENTATION.md](./INTENT_BUBBLE_IMPLEMENTATION.md)** (Phase 3)
   - Intent 接口设计
   - Bubble 冒泡机制
   - Strategy 模式支持
   - OptionGroup/Field/Menu Intent 示例
   - 测试用例

5. **[review2.md](./review2.md)**
   - 架构评审和建议
   - 完整的设计方案参考

---

## 🚀 实施路线图

### Phase 1: Instance Tree (2天)

**目标**: 建立组件逻辑树，解决父子关系问题

- [ ] 定义 `BaseInstance` 和 `ComponentInstance` 接口
- [ ] 修改 Fiber 创建逻辑（`mountInstance`/`unmountInstance`）
- [ ] 更新 OptionGroup/Option 组件
- [ ] 添加单元测试
- [ ] 验证：组件可以通过 `Parent()`/`Children()` 访问

**预期收益**:
- ✅ 无需全局注册表 hack
- ✅ 子组件可以直接访问父组件
- ✅ 为 Context 和 Intent 打下基础

---

### Phase 2: Context System (3天)

**目标**: 实现依赖注入机制，解决 Props Drilling

- [ ] 定义 `ContextKey` 和 `FiberContext`
- [ ] 实现 `Provide`/`UseContext`/`UseContextValue`
- [ ] 集成到 Fiber
- [ ] 创建 Context Provider 组件
- [ ] OptionGroup/Theme/Form Context 示例
- [ ] 添加泛型类型安全测试

**预期收益**:
- ✅ 跨层级通信无需 Props Drilling
- ✅ 类型安全的依赖注入
- ✅ 支持 Theme、Router 等全局状态

---

### Phase 3: Intent Bubble (3天)

**目标**: 实现事件冒泡机制，解耦父子组件

- [ ] 定义 `Intent` 接口和 `IntentEmitter`/`IntentHandler`
- [ ] 实现 `Emit` 冒泡机制
- [ ] 实现 `Strategy` 模式
- [ ] 定义 OptionGroup/Field/Menu Intent
- [ ] 添加冒泡测试用例

**预期收益**:
- ✅ 解耦父子组件通信
- ✅ 支持事件拦截
- ✅ 易于撤销/重做

---

### Phase 4-8: 后续优化

| 阶段 | 任务 | 预计工作量 | 优先级 |
|------|------|-----------|--------|
| Phase 4 | 实现 Portal Layer | 2天 | 🟡 P1 |
| Phase 5 | 重构 OptionGroup | 1天 | 🟢 P2 |
| Phase 6 | 重构 Form 组件 | 2天 | 🟢 P2 |
| Phase 7 | 添加 Menu/Tabs | 3天 | 🟢 P2 |
| Phase 8 | 性能优化和测试 | 2天 | 🟢 P2 |

**总工作量**: 约 18 天

---

## 🎯 核心能力

### Instance Tree - 组件逻辑树

```go
// 组件可以访问树结构
func (inst *OptionInstance) GetParentGroup() (*Instance, bool) {
    parent := inst.Parent()
    if parent == nil {
        return nil, false
    }
    group, ok := parent.(*Instance)
    return group, ok
}
```

### Context System - 依赖注入

```go
// 提供者
func (inst *Instance) OnMount() {
    f := inst.GetFiber()
    f.Context.Provide(OptionGroupContext, inst)
}

// 消费者
func (inst *OptionInstance) OnMount() {
    f := inst.GetFiber()
    group, ok := fcontext.UseContextValue[*Instance](
        f.Context,
        OptionGroupContext,
    )
    // ...
}
```

### Intent Bubble - 事件冒泡

```go
// 发射器
func (inst *OptionInstance) HandleAction(act *Action) bool {
    intent.Emit(inst, OptionSelectIntent{
        GroupKey: inst.groupKey,
        Value:    inst.value,
    })
    return true
}

// 处理器
func (inst *Instance) HandleIntent(i Intent) bool {
    if v, ok := i.(OptionSelectIntent); ok {
        if v.GroupKey == inst.key {
            inst.SelectOption(v.Value)
            return true // 已处理，停止冒泡
        }
    }
    return false // 未处理，继续冒泡
}
```

---

## 🏡 设计原则

### 核心设计重点

1. **Fiber-first**: 保持 Fiber 调度引擎为核心
2. **声明式组件**: VNode 描述 UI，声明式编程
3. **类型安全**: 使用 Go 1.18+ 泛型确保类型安全
4. **高性能**: 针对单线程调度优化，无锁或轻量锁
5. **可扩展**: 易于添加新的 Intent、Context 和组件

### 参考架构

| 框架 | 借鉴的核心 |
|------|-----------|
| **React Fiber** | 调度引擎、事件冒泡、Diff 算法 |
| **Flutter** | 三树架构（Widget → Element → RenderObject） |
| **SwiftUI** | Environment 系统（依赖注入） |
| **BLoC** | Intent/Event 模式（状态管理） |

---

## 💡 使用场景示例

### OptionGroup 示例

```go
// 创建 OptionGroup
options := []optiongroup.Option{
    {Value: "fire", Label: "Fire 🔥"},
    {Value: "ice", Label: "Ice ❄️"},
    {Value: "thunder", Label: "Thunder ⚡"},
}

group := optiongroup.NewBuilder(options).
    Key("weapon-selector").
    Mode(optiongroup.ModeMultiple).
    ForField(intent.BindField("weapons")).
    Vertical().
    Build()

// 内部机制：
// 1. OptionGroup 提供 Context (OptionGroupContext)
// 2. Option 通过 Context 访问父实例
// 3. Option 发射 OptionSelectIntent
// 4. OptionGroup 接收 Intent 并处理
// 5. 更新状态并发射 FieldChangeIntent 到 Store
```

### Form 示例

```go
// 创建 Form
form := form.NewBuilder(fields).
    ForField(intent.BindField("formData")).
    Build()

// 内部机制：
// 1. Form 提供 Context (FormContext)
// 2. Field 通过 Context 访问 Form
// 3. Field 发射 FieldChangeIntent
// 4. Form 接收 Intent 并验证
// 5. Form 发射 FormSubmitIntent 到 Store
```

---

## 📊 对比分析

### 当前 vs Mint 2.0

| 场景 | 当前方案（Hack） | 新方案（原生） | 改善 |
|------|----------------|--------------|------|
| **子→父通信** | 全局注册表 / 闭包 | `Parent()` / Context / Intent | ✅ 80% |
| **深层组件访问** | Props Drilling | Context | ✅ 90% |
| **跨层事件** | 手动回调 | Intent Bubble | ✅ 75% |
| **Portal 渲染** | 部分支持 | 完整支持 | ✅ 100% |

### 组件实现难度对比

| 组件 | 当前难度 | 新难度 | 改善 |
|------|---------|-------|------|
| **OptionGroup** | 🔴 困难 | 🟢 简单 | ✅ 80% |
| **Form** | 🔴 困难 | 🟢 简单 | ✅ 70% |
| **Menu** | 🔴 困难 | 🟢 简单 | ✅ 75% |
| **Tabs** | 🟡 中等 | 🟢 简单 | ✅ 60% |
| **Tree** | 🔴 困难 | 🟡 中等 | ✅ 50% |
| **Portal** | ⚠️ 基础 | 🟢 完整 | ✅ 100% |

---

## 🔗 相关链接

### 文档

- [系统现状分析](../../current_system_analysis.md)
- [架构升级总体报告](./MINT_RUNTIME_2.0_ARCHITECTURE.md)
- [Phase 1 实施方案](./INSTANCE_TREE_IMPLEMENTATION.md)
- [Phase 2 实施方案](./CONTEXT_IMPLEMENTATION.md)
- [Phase 3 实施方案](./INTENT_BUBBLE_IMPLEMENTATION.md)

### 代码

- [Fiber Runtime](../../runtime/ui/fiber.go)
- [VNode 接口](../../runtime/ui/vnode.go)
- [OptionGroup 组件](../../ui/components/optiongroup/)

### 示例

- [Multi-Select Demo](../../examples/multiselect_demo/)
- [Type-Safe Form Demo](../../examples/typesafe_form_demo_runapp/)

---

## 🤝 贡献

欢迎对 Mint Runtime 2.0 提出改进建议：

1. 提交 Issue 描述问题
2. 提交 PR 改进文档
3. 在讨论区交流想法

---

## 📝 更新历史

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-03-06 | 1.0 | 初始版本，完整架构设计 |

---

**状态**: ✅ 架构设计完成，🚧 待实施

**下一步**: 开始实施 Phase 1 - Instance Tree

---

**祝实施顺利！** 🚀
