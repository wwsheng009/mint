# Intent Bubble 模块审查报告

审查日期: 2026年3月7日
审查范围: Mint Runtime 2.0 - Phase 10 Intent Bubble 集成

---

## 执行摘要

Intent Bubble 系统的**基础架构完整**，但在**集成层面存在严重问题**，导致 Intent Bubble 系统实际上无法正常工作。核心问题是：

1. **FiberUtil 错误地将 IntentEmitter 连接到全局 Intent Runtime** 而不是 Bubble 系统
2. **Form.AddChild 不设置 Parent 引用**，导致 Bubble 链路断裂
3. **测试中的 Intent Bubble 是手动构造的**，不是通过正常的渲染流程触发

---

## 发现的问题

### 🔴 问题 1: IntentEmitter 接线错误（严重）

**位置**: `runtime/ui/fiber_util.go:165-176`

**问题描述**:
```go
if setter, ok := instance.(interface{ SetIntentEmitter(func(i intent.Intent)) }); ok {
    setter.SetIntentEmitter(func(intent intent.Intent) {
        // ❌ 错误：发射到全局 intent runtime，而不是 Bubble 系统
        if runtime := GetGlobalIntentRuntime(); runtime != nil {
            result := runtime.Emit(intent)
            // ...
        }
    })
}
```

**影响**:
- 组件发射的 Intent **不会沿着 Instance Tree 冒泡**
- 所有 Intent 都直接发送到 **全局 Intent Registry/Dispatcher**
- 完全违背了 Intent Bubble 的设计意图（本地父子通信）

**正确实现应该是**:
```go
setter.SetIntentEmitter(func(intent intent.Intent) {
    // ✅ 沿着 Instance Tree 向上冒泡
    if component, ok := instance.(intent.TreeComponent); ok {
        intent.Emit(component, intent)
    }
})
```

**根本原因**:
这个错误可能是因为：
1. Intent Bubble 系统是后添加的（Phase 3），而全局 Intent System 是先有的
2. FiberUtil 的代码没有更新以支持 Intent Bubble
3. 没有区分"本地冒泡"和"全局分发"的 Intent 类型

---

### 🔴 问题 2: Form.AddChild 不设置 Parent 引用（严重）

**位置**: `ui/components/form/instance.go:200-217`

**问题描述**:
```go
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
    if child == nil {
        return
    }
    inst.mu.Lock()
    defer inst.mu.Unlock()

    // 检查子组件是否已存在
    for _, existing := range inst.childInstances {
        if existing == child {
            return
        }
    }
    // ❌ 只是添加到 childInstances，没有设置 parent 引用！
    inst.childInstances = append(inst.childInstances, child)
}
```

**影响**:
- 子组件的 `Parent()` 调用返回 **nil**
- `intent.Emit(component, intent)` 无法向上冒泡（因为 parent is nil）
- Form 无法接收来自子组件的 Intent

**正确实现应该是**（参考 OptionGroup）:
```go
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
    if child == nil {
        return
    }
    inst.mu.Lock()
    defer inst.mu.Unlock()

    // 检查是否已存在
    for _, existing := range inst.childInstances {
        if existing == child {
            return
        }
    }

    // ✅ 添加到子组件列表
    inst.childInstances = append(inst.childInstances, child)

    // ✅ 设置 parent 引用（子组件必须是 *BaseComponentInstance 类型）
    if childBase, ok := child.(*rtui.BaseComponentInstance); ok {
        childBase.parent = inst
    }
}
```

**对比**: `ui/components/optiongroup/instance.go:83-118` 的实现是正确的

---

### 🟡 问题 3: 测试使用手动构造的 Intent Bubble（中）

**位置**: `runtime/intent/nested_integration_test.go`

**问题描述**:
测试中的 Intent Bubble 逻辑是**手动构造**的，不符合实际使用场景：

```go
// 设置 intent emitter on select
selectInst.SetIntentEmitter(func(i intent.Intent) {
    bubbleIntents = append(bubbleIntents, i)  // ❌ 手动收集 Intent
})

// 并没有调用 intent.Emit(selectInst, intent)
// 没有 parent -> child 的冒泡链路
```

**影响**:
- 测试不能覆盖**真实的冒泡流程**
- 无法检测 `intent.Emit()` 的正确集成
- 无法测试 `Parent()` 链路的完整性

**需要的测试**:
```go
// 正确的测试应该：
// 1. 通过正常的渲染流程设置 parent 引用
// 2. 调用 intent.Emit(component, intent) 触发冒泡
// 3. 验证父组件的 HandleIntent 被调用

func TestRealIntentBubble(t *testing.T) {
    formInst := form.NewInstance(props)
    selectInst := selectcomp.NewInstance(props)

    // AddChild 应该设置 parent 引用
    formInst.AddChild(selectInst)

    // 设置 Form 为 IntentHandler
    var capturedIntent intent.Intent
    formInst.SetIntentHandler(func(i intent.Intent) bool {
        capturedIntent = i
        return true
    })

    // 触发冒泡（这里需要 Select 调用 intent.Emit）
    selectInst.SelectNext()  // 应该内部调用 intent.Emit(selectInst, SelectChangeIntent)

    // 验证 Form 的 HandleIntent 被调用
    if capturedIntent == nil {
        t.Error("Form should have received the bubbled intent")
    }
}
```

---

### 🟡 问题 4: ComponentID 路由实现不完整（中）

**位置**: `ui/components/select/instance.go:379-425`

**问题描述**:
`HandleIntent` 中对 ComponentID 的检查只检查了特定的 Intent 类型：

```go
func (inst *Instance) HandleIntent(i intent.Intent) bool {
    if inst.componentID != "" {
        // ❌ 只检查了特定的 Intent 类型
        if id, ok := i.(interface{ GetComponentID() string }); ok {
            if id.GetComponentID() != "" && id.GetComponentID() != inst.componentID {
                return false
            }
        }
    }

    switch v := i.(type) {
    case SelectNextIntent:
        // ...
    }
}
```

**影响**:
- 如果 Intent 实现了 `GetComponentID()` 但位置在 switch 之前，会有不一致行为
- 没有统一的 Intent ID 路由机制

**建议**: 创建一个统一的 Intent ID 路由辅助函数：

```go
func shouldHandleIntentWithID(inst *Instance, i intent.Intent) bool {
    if inst.componentID == "" {
        return true  // 没有设置 componentID，处理所有 Intent
    }

    // 检查 Intent 是否有 ComponentID
    if id, ok := i.(interface{ GetComponentID() string }); ok {
        intentID := id.GetComponentID()
        if intentID != "" && intentID != inst.componentID {
            return false  // 不是给本组件的 Intent
        }
    }
    return true
}
```

---

### 🟡 问题 5: OptionGroup 使用 manual parent 设置机制（中）

**位置**: `ui/components/optiongroup/instance.go:83-118`

**问题描述**:
OptionGroup 的 AddChild 手动访问了 `OptionInstance.parent`：

```go
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
    if optInst, ok := child.(*OptionInstance); ok {
        // ...
        optInst.parent = inst  // ❌ 直接访问内部字段
    }
}
```

**影响**:
- 破坏了封装性
- OptionInstance 必须导出 `parent` 字段（小写）
- 与其他组件（如 Form）不一致（Form 没有设置 parent）

**建议**:
- 使用统一的 Parent 设置接口
- 可以在 `TreeNode` 接口中添加 `SetParent(ComponentInstance)` 方法

---

### 🟢 问题 6: Instance Tree 循环检测和重新绑定（已在 Phase 10 修复）✅

**位置**: `runtime/ui/instance.go:224-276`

**状态**: ✅ 已在 Phase 10 正确实现
- `wouldCauseCycle()` 双向循环检测
- 自动重新绑定机制
- BaseComponentInstance.AddChild 自动处理

---

### 🟢 问题 7: Intent 定义完整性（基本完整）✅

**状态**: ✅ 基本完整
- Select: 5 种 Intent（SelectChangeIntent, SelectNextIntent, SelectPrevIntent, SelectByIndexIntent, SelectByValueIntent）
- Treeview: 4 种 Intent（NodeSelectIntent, NodeExpandIntent, NodeCollapseIntent, NavigationIntent）
- Form: 5 种 Intent（FormFieldChangeIntent, FormFieldBlurIntent, FormValidateIntent, FormSubmitIntent, FormResetIntent）
- OptionGroup: OptionSelectIntent（已有）

---

## 设计层面的潜在问题

### 1. Intent Bubble 和全局 Intent System 混淆

**问题**: FiberUtil 通过 IntentEmitter 将所有 Intent 发送到**全局系统**，而实际上应该：
- **Intent Bubble Intent（本地）** → `intent.Emit(component, intent)` 向上冒泡
- **全局 System Intent（跨组件）** → 全局 Intent Runtime 处理

**建议**: 区分两种 Intent 类型
```go
type LocalIntent interface {
    Intent
    IsLocal() bool  // true = bubble, false = global
}

type GlobalIntent interface {
    Intent
    IsGlobal() bool
}

// FiberUtil 中
if intent.IsLocal() {
    // 沿着 Instance Tree 冒泡
    intent.Emit(component, intent)
} else {
    // 发送到全局 Runtime
    runtime.Emit(intent)
}
```

### 2. 组件不一致的 Parent 引用设置

**问题**:
- OptionGroup: 正确设置 parent（手动访问内部字段）
- Form: 不设置 parent
- BaseComponentInstance.AddChild: 自动设置

**建议**: 统一所有 TreeContainer 组件的 AddChild 实现

---

## 推荐修复优先级

### P0 - 阻断性（必须修复）:
1. **修复 FiberUtil 的 IntentEmitter 接线** - 让它使用 `intent.Emit(component, intent)`
2. **修复 Form.AddChild 设置 Parent 引用** - 确保所有父组件正确设置子组件的 parent

### P1 - 重要（强烈建议）:
3. **重构测试以使用真实的 Intent Bubble 流程** - 确保测试覆盖实际的冒泡链路
4. **统一 ComponentID 路由机制** - 创建辅助函数避免重复代码

### P2 - 改进（未来优化）:
5. **区分本地和全局 Intent** - 设计更清晰的 Intent 类型层次
6. **统一 Parent 引用设置接口** - 避免手动访问内部字段

---

## 结论

Intent Bubble 的**核心实现（气泡系统和接口）是正确的**，但**集成到 Fiber 层面的代码存在严重问题**，导致实际的冒泡流程无法工作。目前系统实际上仍然是在使用全局 Intent System，而不是本地冒泡。

建议优先修复 P0 级别的问题，然后再进行测试和验证。
