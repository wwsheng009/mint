# Intent 系统架构分析与改进方案

**日期**: 2026年3月7日
**背景**: Phase 10 Intent Bubble 集成导致 typed_intent_demo 等示例功能失效

---

## 一、现有架构分析

### 1.1 全局 Intent Runtime 系统

#### 组件层次结构

```
globalIntentRuntime (ui/hooks.go)
├── Registry (intent.Registry)
│   └── handlers: map[string]*HandlerRegistration
│       └── Handler.Handle(ctx, intent) → IntentResult
│
├── Dispatcher (intent.Dispatcher)
│   └── Dispatch(intent) → 查找 handler → 执行 → 返回结果
│
└── Store (intent.SimpleStore / store.Store[T])
    └── Get() / Set(newState)
        └── 触发 UI 更新
```

#### Intent 处理流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Intent 处理流程                                  │
└─────────────────────────────────────────────────────────────────────┘

组件层面：
┌─────────────────────────────────────────────────────────────────────┐
│  Button/Select/OptionGroup                                          │
│    用户交互 → EmitIntent(Intent)                                     │
│      ↓                                                              │
│    IntentEmitter (fiber_util.go)                                    │
│      → GetGlobalIntentRuntime().Emit(intent)                        │
└─────────────────────────────────────────────────────────────────────┘

全局层面：
┌─────────────────────────────────────────────────────────────────────┐
│  Runtime.Emit(intent)                                                │
│    ↓                                                                │
│  Dispatcher.Dispatch(intent)                                         │
│    ↓                                                                │
│  Registry.Lookup(intentType)                                         │
│    ↓                                                                │
│  Handler.Handle(ctx, intent)                                         │
│    ├── 获取当前状态: store.Get()                                      │
│    ├── 调用 Reducer: reducer.Reduce(state, intent)                  │
│    └── 更新状态: store.Set(newState)                                 │
│       ↓                                                             │
│    触发 UI 重新渲染                                                  │
└─────────────────────────────────────────────────────────────────────┘

Handler 注册：
┌─────────────────────────────────────────────────────────────────────┐
│  appReducerBuilder                                          │
│    ├── BindFieldMap(map[string]func)                               │
│    │   └── → FieldChangeIntent 处理                                  │
│    └── On(SubmitIntent{}, handler)                                 │
│        └── → SubmitIntent 处理                                       │
│    ↓                                                                │
│  RegisterToGlobal(rt.GetStore())                                    │
│    ↓                                                                │
│  注册到 DefaultRegistry() → 在 ui.RunApp 中使用                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

### 1.2 Intent Bubble 系统

#### 现有实现（Phase 3）

```go
// runtime/intent/bubble.go

// Emit 发送 Intent，沿 Parent 链向上冒泡
func Emit(component TreeComponent, i Intent) {
    depth := 0
    for component != nil && depth < MaxBubbleDepth {
        if parent := component.Parent(); parent != nil {
            parent.EmitIntent(i)  // 继续冒泡
        }
        component = parent
        depth++
    }
}

// TreeComponent 接口
type TreeComponent interface {
    Parent() interface{}
}

// IntentHandler 接口
type IntentHandler interface {
    HandleIntent(i Intent) bool
}

// IntentEmitter 接口
type IntentEmitter interface {
    SetIntentEmitter(func(Intent))
}
```

#### 问题分析（根据 INTENT_BUBBLE_AUDIT_REPORT.md）

**P0-1**: FiberUtil 错误地将 IntentEmitter 连接到全局 Runtime

```go
// 错误的实现（Phase 10 修复前）
setter.SetIntentEmitter(func(i intent.Intent) {
    if runtime := GetGlobalIntentRuntime(); runtime != nil {
        result := runtime.Emit(i)  // ❌ 所有 Intent 都直接全局分发
    }
})

// 修复后的实现（Phase 10）
setter.SetIntentEmitter(func(i intent.Intent) {
    if component, ok := instance.(TreeComponent); ok {
        intent.Emit(component, i)  // ❌ 所有 Intent 都本地冒泡
    }
})

// 问题：
// 1. FieldChangeIntent 需要全局处理（FieldMap）
// 2. SubmitIntent 需要全局处理（Reducer）
// 3. 本地冒泡破坏了现有功能
```

---

### 1.3 关键问题识别

#### 问题 1: 统一路由与分类路由的冲突

**现有架构假设**：
- 所有 Intent 都应该发送到全局 Runtime
- Handler 通过 Registry 注册
- Dispatcher 路由到对应的 Handler

**Intent Bubble 的假设**：
- 子组件 Intent 应该优先冒泡到父组件
- 父组件可以选择处理或继续冒泡
- 用于组件内部通信和事件委托

**冲突点**：
```
    全局路由                      本地冒泡
    ────────                     ────────
    FieldChangeIntent → FieldMap  (需要全局)
    SubmitIntent → Reducer        (需要全局)
    ButtonPress → ???            (两者皆可)
    SelectChange → ???           (两者皆可)
```

#### 问题 2: Handler 与 HandleIntent 的混淆

```go
// 全局 Handler（Registry 注册）
func(ctx *ActionContext, i Intent) IntentResult

// 组件 HandleIntent（组件内部）
func(inst *Instance, i Intent) bool
```

**两者的关系**：
- Handler: 定义在 Reducer 中，注册到全局 Registry
- HandleIntent: 定义在组件中，用于本地处理
- 没有明确的接口或区分机制

#### 问题 3: 没有明确的 Intent 分类标准

| Intent 类型 | 是否需要全局处理 | 是否需要本地冒泡 | 当前处理方式 |
|-----------|----------------|----------------|-------------|
| FieldChangeIntent | ✅ | ❌ | 全局 |
| SubmitIntent | ✅ | ❌ | 全局 |
| ButtonPressIntent | ✅ (app 逻辑) | ✅ (焦点管理) | 混乱 |
| SelectChangeIntent | ✅ (状态更新) | ✅ (父组件监听) | 混乱 |
| FocusIntent/BlurIntent | ✅ (焦点管理) | ❌ | 全局 |

---

## 二、改进方案设计

### 2.1 方案概述

**目标**：
1. 不破坏现有功能（typed_intent_demo 等示例）
2. 支持可选的 Intent Bubble 功能
3. 提供清晰的 Intent 分类标准

**核心思想**：
- 定义 `GlobalIntent` 接口标记需要全局处理的 Intent
- 兼容现有代码（所有现有 Intent 默认为全局）
- 新 Intent 可以选择实现 `GlobalIntent` 接口

---

### 2.2 方案细节

#### 接口定义

```go
// runtime/intent/types.go

// GlobalIntent 标记需要全局处理的 Intent
// 如果 Intent 实现此接口并返回 true，将发送到全局 Runtime
// 否则在本地组件中冒泡（如果支持）
type GlobalIntent interface {
    Intent
    IsGlobal() bool  // true = 全局处理，false = 本地冒泡
}

// 默认实现：所有现有 Intent 都是全局 Intent
// 方案 1: 使用默认标记（推荐，最小改动）
//   - Intent 是否实现 GlobalIntent 接口可选
//   - 如果未实现，默认为全局处理（向后兼容）

// 方案 2: 类型断言检查
//   - 在 fiber_util.go 中进行类型断言
//   - 如果实现了 GlobalIntent.IsGlobal() 返回 false，则本地冒泡
```

#### FiberUtil 实现

```go
// runtime/ui/fiber_util.go

setter.SetIntentEmitter(func(i intent.Intent) {
    // 检查是否为全局 Intent
    var isGlobal bool
    if global, ok := i.(intent.GlobalIntent); ok {
        isGlobal = global.IsGlobal()
    } else {
        // 默认：未实现 GlobalIntent 的 Intent 都是全局 Intent
        // 这确保了向后兼容性
        isGlobal = true
    }

    if isGlobal {
        // 全局 Intent: 发送到全局 Runtime
        if runtime := GetGlobalIntentRuntime(); runtime != nil {
            result := runtime.Emit(i)
            if result.Error != nil {
                // 错误处理
            }
        }
    } else {
        // 本地 Intent: 冒泡到父组件
        if component, ok := instance.(intent.TreeComponent); ok {
            intent.Emit(component, i)
        }
    }
})
```

---

### 2.3 兼容性保证

#### 向后兼容

| 现有代码 | 行为 | 兼容性 |
|---------|------|--------|
| FieldChangeIntent | 全局处理 | ✅ 默认为 GlobalIntent |
| SubmitIntent | 全局处理 | ✅ 默认为 GlobalIntent |
| Button.OnPress(SubmitIntent) | 全局处理 | ✅ 默认为 GlobalIntent |
| OptionGroup.ForField | 全局处理 | ✅ 默认为 GlobalIntent |

#### 新功能支持

| 新功能 | 实现 | 说明 |
|-------|------|------|
| 组件本地通信 | 实现 `LocalIntent` | `IsGlobal() = false` |
| 事件委托 | 实现 `LocalIntent` | 父组件 HandleIntent |
| 自定义冒泡 | 实现 `GlobalIntent` | 选择全局或本地 |

---

### 2.4 使用示例

#### 示例 1: 现有代码（无需修改）

```go
// typed_intent_demo/main.go
// 无需修改，保持不变
submitButton := button.NewBuilder("Submit").
    OnPress(SubmitIntent{}).  // 默认为全局 Intent
    Build()

optiongroup.NewBuilder(options).
    Mode(ModeSingle).
    ForField(intent.BindField("City")).  // FieldChangeIntent 默认全局
    Build()
```

#### 示例 2: 本地组件通信（新功能）

```go
// 定义本地 Intent
type ChildSelectIntent struct {
    SelectedValue string
}

func (ChildSelectIntent) IntentType() string { return "ChildSelect" }
func (ChildSelectIntent) IsGlobal() bool { return false }  // 本地冒泡

// 子组件发射
childInst.EmitIntent(ChildSelectIntent{SelectedValue: "value"})

// 父组件接收
func (p *Parent) HandleIntent(i intent.Intent) bool {
    if select, ok := i.(ChildSelectIntent); ok {
        // 处理本地选择事件
        p.selectedValue = select.SelectedValue
        return true
    }
    return false
}
```

---

### 2.5 实现步骤

#### 步骤 1: 添加 GlobalIntent 接口

**文件**: `runtime/intent/types.go`

```go
// GlobalIntent 标记需要全局处理的 Intent
// 默认行为：未实现此接口的 Intent 都是全局 Intent
type GlobalIntent interface {
    Intent
    IsGlobal() bool
}
```

#### 步骤 2: 更新 fiber_util.go

**文件**: `runtime/ui/fiber_util.go`

```go
setter.SetIntentEmitter(func(i intent.Intent) {
    // 检查是否为本地 Intent（实现 GlobalIntent 并返回 false）
    isLocal := false
    if global, ok := i.(intent.GlobalIntent); ok {
        isLocal = !global.IsGlobal()
    }

    if isLocal {
        // 本地 Intent: 冒泡到父组件
        if component, ok := instance.(intent.TreeComponent); ok {
            intent.Emit(component, i)
        }
    } else {
        // 全局 Intent: 发送到全局 Runtime
        if runtime := GetGlobalIntentRuntime(); runtime != nil {
            result := runtime.Emit(i)
            if result.Error != nil {
                if log.UILogger.Enabled() {
                    log.UILogger.Debug("[IntentEmitter] Failed to emit intent %s: %v",
                        i.IntentType(), result.Error)
                } else {
                    fmt.Printf("[IntentEmitter] Failed to emit intent %s: %v\n",
                        i.IntentType(), result.Error)
                }
            }
        }
    }
})
```

#### 步骤 3: 测试验证

1. 运行 `typed_intent_demo`: 确保所有功能正常
2. 运行所有现有测试：确保没有破坏
3. 添加 Intent Bubble 测试：验证本地冒泡功能

---

## 三、方案对比

### 选项 1: 添加 GlobalIntent 接口（推荐）

| 特性 | 评分 |
|-----|-----|
| 向后兼容 | ✅✅✅ |
| 实现复杂度 | ✅✅✅ |
| 清晰度 | ✅✅ |
| 灵活性 | ✅✅ |

**优势**：
- 最小改动
- 完全向后兼容
- 清晰的 Intent 分类

**劣势**：
- 需要为每个本地 Intent 实现 `IsGlobal() = false`

---

### 选项 2: 添加 LocalIntent 接口

```go
type LocalIntent interface {
    Intent
    IsLocal() bool
}
```

**区别**：
- GlobalIntent: 默认全局，需要显式标记本地
- LocalIntent: 默认本地，需要显式标记全局

**推荐**: 选项 1（GlobalIntent），因为：
1. 现有代码默认都是全局处理
2. 最小化改动
3. 更符合"默认全局，可选本地"的语义

---

### 选项 3: 基于的元数据标记

```go
// 使用 props 或 VNode 属性
optiongroup.NewBuilder(options).
    LocalBubbling(true).
    Build()
```

**劣势**：
- 需要修改 Builder API
- 无法细粒度控制到 Intent 级别
- 不够类型安全

---

## 四、推荐实现方案

### 选择: 选项 1（添加 GlobalIntent 接口）

**理由**：
1. ✅ 完全向后兼容：现有代码无需修改
2. ✅ 最小实现复杂度：只需添加接口和更新 fiber_util.go
3. ✅ 清晰的分类标准：GlobalIntent.IsGlobal()
4. ✅ 灵活的扩展性：新 Intent 可以自由选择全局或本地
5. ✅ 类型安全：编译时检查

**关键设计决策**：
1. 默认行为：未实现 GlobalIntent 的 Intent 都是全局 Intent
2. 向后兼容：所有现有 Intent 自动工作
3. 可选功能：Intent Bubble 作为可选特性
4. 渐进迁移：可以逐步将某些 Intent 改为本地

---

## 五、实施计划

### 阶段 1: 基础设施（本次）

1. ✅ 添加 `GlobalIntent` 接口
2. ✅ 更新 `fiber_util.go` 实现双路由
3. ✅ 测试验证（编译 + 现有测试）
4. ✅ 运行 `typed_intent_demo` 确保功能正常

### 阶段 2: 示例和文档（后续）

1. 编写 Intent Bubble 使用示例
2. 更新组件开发文档
3. 添加最佳实践指南

### 阶段 3: 优化和扩展（可选）

1. 为常用本地 Intent 提供便捷类型
2. 探索更高级的冒泡控制机制
3. 性能优化（减少类型断言开销）

---

## 六、风险和缓解

| 风险 | 缓解措施 |
|-----|---------|
| 性能影响 | 类型断言开销可忽略（已在其他地方使用） |
| 开发者困惑 | 提供清晰的文档和示例代码 |
| API 不稳定 | GlobalIntent 接口一旦确定不要轻易修改 |
| 兼容性问题 | 默认行为确保向后兼容 |

---

## 七、总结

### 架构现状

1. **全局 Intent Runtime**: 成功的基础架构，支持 State → Reducer → UI 流程
2. **Intent Bubble**: 有价值但集成混乱，需要明确分类
3. **分类标准缺失**: 没有统一的 Intent 路由规则

### 改进方案

1. **添加 GlobalIntent 接口**: 实现清晰的 Intent 分类
2. **默认全局行为**: 确保向后兼容
3. **可选本地冒泡**: 支持灵活的组件通信
4. **最小改动**: 只修改 fiber_util.go 和添加接口

### 预期效果

1. ✅ 功能完整性: 所有现有示例正常工作
2. ✅ 清晰的架构: Intent 路由规则明确
3. ✅ 向后兼容: 现有代码无需修改
4. ✅ 可扩展性: 可以渐进式采用 Intent Bubble

---

**下一步**: 实施阶段 1，验证功能完整性
