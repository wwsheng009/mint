# Mint Runtime 2.0 - Component Patterns Guide

本文档描述了 Mint Runtime 2.0 架构中的组件设计模式和实践模式。

## 目录

1. [Instance Tree 构建模式](#instance-tree-构建模式)
2. [Intent Bubble 集成模式](#intent-bubble-集成模式)
3. [Context 系统使用模式](#context-系统使用模式)
4. [示例总结](#示例总结)

---

## Instance Tree 构建模式

### 概述

Mint Runtime 2.0 支持两种 Instance Tree 构建模式：

| 模式 | 描述 | 适用场景 | 典型组件 |
|------|------|----------|----------|
| **Reconciler Automatic** | Reconciler 在 diff 阶段自动调用 `mountInstanceToTree` 建立父子关系 | 动态子组件、列表渲染、表单字段 | `Form` |
| **OnMount Manual** | 组件在 `OnMount` 阶段手动调用 `rebuildChildInstances` 建立子树 | 静态子组件、固定结构 | `OptionGroup` |

### 模式 1: Reconciler Automatic

在 Reconciler Automatic 模式中，组件只需要实现 `TreeContainer` 接口（`AddChild`, `RemoveChild`）和 `TreeNode` 接口（`Parent`, `Children`），Reconciler 会在适当的时候自动调用这些方法。

```go
// 示例: Form 组件
type Instance struct {
    key            string
    childInstances []rtui.ComponentInstance
    // ... 其他字段
}

// AddChild 实现 TreeContainer 接口
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
    inst.mu.Lock()
    defer inst.mu.Unlock()

    for _, existing := range inst.childInstances {
        if existing == child {
            return // 已存在
        }
    }
    inst.childInstances = append(inst.childInstances, child)
}

// Children 实现 TreeNode 接口
func (inst *Instance) Children() []rtui.ComponentInstance {
    inst.mu.RLock()
    defer inst.mu.RUnlock()
    return inst.childInstances
}
```

**特点**：
- ✅ 简单、自动化
- ✅ 适用于子组件动态变化的场景
- ✅ 内存管理由 Reconciler 处理
- ⚠️ 需要注意并发访问（使用 mutex）

### 模式 2: OnMount Manual

在 OnMount Manual 模式中，组件在 `OnMount` 时遍历 VNode 的子节点，手动创建和管理子实例。

```go
// 示例: OptionGroup 组件
type Instance struct {
    key            string
    childInstances []*OptionInstance
    vnode         *VNode
    // ... 其他字段
}

func (inst *Instance) OnMount() {
    inst.behaviors.OnMount(inst)
    // 手动重建子实例树
    inst.rebuildChildInstances()
}

func (inst *Instance) rebuildChildInstances() {
    // 1. 清空现有子实例
    for _, child := range inst.childInstances {
        parent := child.Parent()
        if parent != nil {
            // 清理 parent 引用
            child.SetParent(nil)
        }
    }
    inst.childInstances = inst.childInstances[:0]

    // 2. 遍历 VNode 子节点
    for i, childVNode := range inst.vnode.Children() {
        if childInst, ok := childVNode.(*optiongroup.VNode); ok {
            child := childInst.CreateInstance()
            // 设置 parent 引用
            child.SetParent(inst)
            inst.childInstances = append(inst.childInstances, child)
        }
    }
}
```

**特点**：
- ✅ 对于静态结构更简洁
- ✅ 可以精确控制子实例的创建和销毁
- ✅ 适用于组件内部有特定子节点类型的场景
- ⚠️ 需要手动管理 parent 引用，避免内存泄漏

---

## Intent Bubble 集成模式

### 概述

Intent Bubble 是一种本地事件冒泡机制，允许子组件通过 Intent 与父组件通信。

### 基本接口

```go
// TreeComponent - 标识组件支持冒泡
type TreeComponent interface {
    Parent() interface{}
}

// IntentHandler - 处理 Intent
type IntentHandler interface {
    HandleIntent(i Intent) bool  // 返回 true 表示已处理，停止冒泡
}

// IntentEmitter - 发射 Intent
type IntentEmitter interface {
    EmitIntent(i Intent)
}
```

### 模式 1: Intent 定义

为每个组件定义特定类型的 Intent：

```go
// 示例: Select 组件的 Intent
type SelectChangeIntent struct {
    SelectedIndex  int
    SelectedValue string
    SelectedLabel string
    ComponentID   string  // 用于路由
}

func (i SelectChangeIntent) IntentType() string {
    return "select.SelectChangeIntent"
}

func (i SelectChangeIntent) Priority() intent.ActionPriority {
    return intent.PriorityNormal
}

func (i SelectChangeIntent) IsTransition() bool {
    return false
}
```

### 模式 2: Intent 发射

组件在状态变更时发射 Intent：

```go
// 示例: Select 组件
func (inst *Instance) SelectNext() {
    if len(inst.options) == 0 {
        return
    }
    inst.selectedIndex = (inst.selectedIndex + 1) % len(inst.options)
    inst.dirty = true
    inst.emitSelectChange() // 发射 Intent
}

func (inst *Instance) emitSelectChange() {
    if inst.intentEmitter == nil {
        return
    }

    var selectIntent SelectChangeIntent
    selectedValue := inst.SelectedValue()
    selectedLabel := inst.SelectedLabel()

    if inst.componentID != "" {
        selectIntent = SelectChangeWithID(inst.componentID, inst.selectedIndex, selectedValue, selectedLabel)
    } else {
        selectIntent = SelectChange(inst.selectedIndex, selectedValue, selectedLabel)
    }

    intent.Emit(inst, selectIntent)
}
```

### 模式 3: Intent 处理

组件实现 `IntentHandler` 接口来处理特定 Intent：

```go
func (inst *Instance) HandleIntent(i Intent) bool {
    // 只处理属于本组件的 Intent（通过 componentID）
    if inst.componentID != "" {
        if id, ok := i.(interface{ GetComponentID() string }); ok {
            if id.GetComponentID() != "" && id.GetComponentID() != inst.componentID {
                return false // 不是给本组件的 Intent，忽略
            }
        }
    }

    switch v := i.(type) {
    case SelectNextIntent:
        inst.SelectNext()
        return true
    case SelectPrevIntent:
        inst.SelectPrev()
        return true
    case SelectByIndexIntent:
        if v.Index >= -1 && v.Index < len(inst.options) {
            inst.SetSelectedIndex(v.Index)
            return true
        }
    }
    return false
}
```

### 模式 4: 嵌套组件 Intent 冒泡

子组件的 Intent 自动冒泡到父组件：

```go
// 子组件发射到父组件的链路：
// 子组件.emitSelectChange() -> intent.Emit(child, intent) 
//   -> 调用 child.parent.HandleIntent(intent) 
//   -> 如果父组件实现了 IntentHandler，可以处理 Intent
//   -> 如果处理返回 false，继续向上冒泡（最多 100 层）
```

---

## Context 系统使用模式

### 概述

Context 系统提供了一种依赖注入机制，允许父组件向子组件传递数据和方法。

### 模式 1: Provider 模式（全局注册）

对于需要在多个子组件间共享状态的容器组件（如 `Form`），使用全局注册模式：

```go
// 定义 Context
type FormContext interface {
    GetValue(field string) (interface{}, bool)
    SetValue(field string, value interface{})
    GetValues() map[string]interface{}
    // ... 其他方法
}

// 全局注册表
var (
    formRegistry = make(map[string]*Instance)
    formMu       sync.RWMutex
)

// 注册（在 OnMount）
func (inst *Instance) OnMount() {
    formID := inst.Key()
    if formID != "" {
        formMu.Lock()
        formRegistry[formID] = inst
        formMu.Unlock()
    }
}

// 注销（在 OnUnmount）
func (inst *Instance) OnUnmount() {
    formID := inst.Key()
    if formID != "" {
        formMu.Lock()
        delete(formRegistry, formID)
        formMu.Unlock()
    }
}

// 子组件通过实例树解析祖先 Form
func GetFormContext(formID string) FormContext {
    // 当前 owner 所在实例树内查找匹配 Form
    // 若需要显式跨树兼容访问，则调用 GetRegisteredFormContext(formID)
    ...
}
```

### 模式 2: 子组件使用 Context

```go
// 子组件通过 formID 和 ForForm 绑定到表单
type SelectInstance struct {
    key         string
    componentID string
    formID      string // 表单 ID
    // ...
}

// 通过实例树 + Intent 获取表单数据
func (inst *Instance) emitFieldValueChanged() {
    if inst.formID != "" {
        formCtx := GetFormContext(inst.formID) // owner-bound 场景
        if formCtx != nil {
            // 可以在这里访问表单状态
            value := formCtx.GetValue("field")
            // 更新表单值
            formCtx.SetValue("field", inst.SelectedValue())
        }
    }
}
```

### 模式 3: Intent 作为上下文更新

Intent 系统与 Context 系统可以协同工作：

```go
// 子组件通过 Intent 通知父组件更新
type FormFieldChangeIntent struct {
    FormID string
    Field  string
    Value  interface{}
    IsDirty bool
}

// 表单处理 Intent
func (inst *FormInstance) HandleIntent(i Intent) bool {
    switch v := i.(type) {
    case FormFieldChangeIntent:
        if v.FormID == inst.Key() {
            inst.SetValue(v.Field, v.Value)
            // 子组件已经触发了 Intent，这里只需要更新表单状态
            // 子组件也可以通过 GetFormContext 访问表单
        }
    }
    return false
}
```

---

## 示例总结

### 示例 1: Form + Select (Reconciler Automatic + Intent Bubble + Context)

```go
// 创建表单
formBuilder := form.NewForm("loginForm")
formBuilder.Label("Login")

// 创建绑定到表单的 Select
selectBuilder := selectcomp.NewBuilder()
selectBuilder.SetKey("username")
selectBuilder.ForForm(intent.BindForm("loginForm")) // 设置表单 ID
selectBuilder.SetOptions([]selectcomp.Option{
    {Value: "user1", Label: "User 1"},
    {Value: "user2", Label: "User 2"},
})

// 添加到表单
formBuilder.AddChild(selectBuilder)

// === 运行时行为 ===
// 1. Reconciler 在 diff 阶段调用 form.AddChild(selectInstance)
// 2. Form 的 AddChild 实现（自动模式）维护 childInstances
// 3. Select 在 OnMount 时，用户改变选项时发射 SelectChangeIntent
// 4. 由于 formID 已设置，Select 在 owner-bound 场景下可通过 GetFormContext("loginForm") 访问祖先表单
//    若需要跨树兼容访问，则显式调用 GetRegisteredFormContext("loginForm")
```

### 示例 2: OptionGroup + Option (OnMount Manual + Intent Bubble)

```go
// 创建 OptionGroup Builder
groupBuilder := optiongroup.NewBuilder()
groupBuilder.SetKey("group1")
groupBuilder.SetStyle(style.Style{FG: color.Blue})

// 添加 Option
groupBuilder.AddOption("option1", "Option 1")
groupBuilder.AddOption("option2", "Option 2")

// === 运行时行为 ===
// 1. OptionGroup 在 OnMount 时调用 rebuildChildInstances
// 2. 遍历 VNode 子节点，手动创建 Option 实例
// 3. 设置 Option 的 parent 为 OptionGroup 实例
// 4. 用户点击 Option 时，发射 OptionSelectIntent
// 5. Intent 自动冒泡到 OptionGroup，触发选中更改
```

---

## 最佳实践

### 1. 内存管理

- **Always clear parent references**: 当子实例被移除时，确保清理 `parent` 引用
- **Use mutex for concurrent access**: 对于 `childInstances` 等共享数据，使用 `sync.RWMutex`
- **Avoid cycles in Instance Tree**: 使用 `wouldCauseCycle` 检查防止循环引用

### 2. Intent Bubble

- **Use componentID for routing**: 在多实例场景中，使用 `componentID` 区分不同实例
- **Keep bubbling depth limited**: Intent Bubble 系统默认限制 100 层，防止无限循环
- **Return true for handled Intents**: 如果 Intent 被处理，应返回 `true` 停止冒泡

### 3. Context System

- **Register in OnMount, Unregister in OnUnmount**: 确保资源正确清理
- **Use unique formIDs**: 在全局注册表中，确保每个表单有唯一的 key
- **Thread-safe access**: Context 注册表使用读写锁，支持并发访问

### 4. 接口实现

- **Use `interface{}` for Parent()**: 为了避免 import 循环，`Parent()` 返回 `interface{}`
- **Implement TreeComponent for bubbling**: 需要 Intent Bubble 的组件必须实现 `TreeComponent`
- **Provide IntentEmitter field**: `SetIntentEmitter(fn)` 用于设置 Intent 冒泡回调
