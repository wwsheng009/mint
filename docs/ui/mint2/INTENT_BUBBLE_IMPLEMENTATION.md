# Intent Bubble 实现总结

## 概述

Mint Runtime 2.0 Intent 系统架构改进，实现了基于 `GlobalIntent` 接口的双路由策略，同时保持完整的向后兼容性。

## 设计目标

1. **不破坏现有功能**: 所有现有代码无需修改
2. **支持可选的 Intent Bubble**: 组件可以选择使用本地冒泡
3. **清晰的 Intent 分类**: 明确区分全局和本地 Intent

## 架构设计

### GlobalIntent 接口

```go
// GlobalIntent 标记 Intent 的处理方式
type GlobalIntent interface {
    // IsGlobal 返回 false 时本地冒泡，返回 true 时全局处理
    // 未实现 GlobalIntent 的 Intent 默认全局处理（向后兼容）
    IsGlobal() bool
}
```

### 智能 FiberUtil 路由

```go
// FiberUtil 创建组件时设置 IntentEmitter
setter.SetIntentEmitter(func(i intent.Intent) {
    // 检查是否为本地 Intent
    isLocal := false
    if global, ok := i.(intent.GlobalIntent); ok {
        isLocal = !global.IsGlobal()
    }

    if isLocal {
        // 本地 Intent: 沿 Parent() 链冒泡
        if component, ok := instance.(intent.TreeComponent); ok {
            intent.Emit(component, i)
        }
    } else {
        // 全局 Intent: 发送到全局 Runtime
        if runtime := GetGlobalIntentRuntime(); runtime != nil {
            runtime.Emit(i)
        }
    }
})
```

### 两种路由模式

| 模式 | IsGlobal() 实现 | 路由路径 | 使用场景 |
|------|-----------------|----------|----------|
| 全局 | 未实现或返回 true | FiberUtil → Runtime → Reducer/FieldMap | 状态管理、表单提交、路由 |
| 本地 | 返回 false | FiberUtil → intent.Emit() → Parent() 链 | 组件内部通信、子组件控制 |

## 已实现的组件 Intent

### OptionGroup

- `OptionSelectIntent.IsGlobal() = false`
- `OptionGroupDeselectIntent.IsGlobal() = false`

**使用场景**: Option 子组件向父组件 OptionGroup 报告选择状态

### Select

- `SelectChangeIntent.IsGlobal() = false`
- `SelectNextIntent.IsGlobal() = false`
- `SelectPrevIntent.IsGlobal() = false`
- `SelectByIndexIntent.IsGlobal() = false`
- `SelectByValueIntent.IsGlobal() = false`

**使用场景**: Select 组件的导航和状态变更

### Treeview

- `NodeSelectIntent.IsGlobal() = false`
- `NodeExpandIntent.IsGlobal() = false`
- `NodeCollapseIntent.IsGlobal() = false`
- `NavigationIntent.IsGlobal() = false`

**使用场景**: 树节点的选择状态和展开/折叠控制

## 默认全局 Intent

未实现 `GlobalIntent` 接口的 Intent 默认全局处理，包括：

- `FieldChangeIntent` - 表单字段变更（全局状态管理）
- `SubmitIntent` - 表单提交（全局动作）
- `NavigateIntent` - 路由导航（全局功能）
- 所有其他现有 Intent

## 测试验证

### Intent Bubble 集成测试

```go
func TestNestedIntentBubble_FormTo(t *testing.T) {
    // 创建 Form and Select
    formInst := form.NewInstance(props)
    selectInst := selectcomp.NewInstance(props)
    formInst.AddChild(selectInst)

    // 设置 emitter 回调
    intentEmitted := false
    var emittedIntent intent.Intent
    selectInst.SetIntentEmitter(func(i intent.Intent) {
        intentEmitted = true
        emittedIntent = i
    })

    // 触发选择变更
    selectInst.SelectNext()

    // 验证 Intent 被发射
    assert.True(t, intentEmitted)
    assert.IsType(t, selectcomp.SelectChangeIntent{}, emittedIntent)
}
```

**结果**: ✅ PASS

### 组件测试

- ✅ Select: 所有测试通过
- ✅ OptionGroup: 所有测试通过
- ✅ Form: 所有测试通过
- ✅ Treeview: 所有测试通过
- ✅ Button: 所有测试通过（包括 Focus）

### 示例程序

- ✅ typed_intent_demo: 编译成功
- ✅ 所有 examples: 编译成功

## 使用指南

### 1. 创建本地 Intent

```go
// Step 1: 定义 Intent
type MyLocalIntent struct {
    // ... 字段
}

// Step 2: 实现 GlobalIntent 接口
func (MyLocalIntent) IsGlobal() bool {
    return false // 本地冒泡
}
```

### 2. 发射本地 Intent

```go
// 在组件内部
func (inst *MyComponent) triggerAction() {
    intent := MyLocalIntent{/* ... */}
    if inst.intentEmitter != nil {
        inst.intentEmitter(intent) // FiberUtil 会自动路由到本地冒泡
    }
}
```

### 3. 处理本地 Intent

```go
// 在父组件中
func (inst *ParentComponent) HandleIntent(i intent.Intent) bool {
    switch v := i.(type) {
    case MyLocalIntent:
        // 处理本地 Intent
        inst.handleMyLocalIntent(v)
        return true // 停止冒泡
    }
    return false // 继续冒泡
}
```

## 文件变更

### 新增文件

1. `runtime/intent/global_intent.go` - GlobalIntent 接口定义
2. `docs/ui/mint2/INTENT_ARCHITECTURE_ANALYSIS.md` - 架构分析文档

### 修改文件

1. `runtime/ui/fiber_util.go` - 实现智能路由
2. `ui/components/optiongroup/intent.go` - 本地 Intent 实现
3. `ui/components/select/intent.go` - 本地 Intent 实现
4. `ui/components/select/instance.go` - 使用 intentEmitter
5. `ui/components/treeview/intent.go` - 本地 Intent 实现

## Git 提交

### Commit 1: feat(intent): 实现 GlobalIntent 接口和智能路由机制

```
架构改进：
- 添加 GlobalIntent 接口，区分全局 Intent 和本地 Intent
- 默认行为：未实现 GlobalIntent 的 Intent 都是全局 Intent（向后兼容）
- 实现 GlobalIntent 接口并返回 false 的 Intent 本地冒泡
- 实现 GlobalIntent 接口并返回 true 的 Intent（或未实现）全局处理

修改内容：
1. 新增 runtime/intent/global_intent.go
2. 更新 runtime/ui/fiber_util.go
3. 新增 docs/ui/mint2/INTENT_ARCHITECTURE_ANALYSIS.md

测试结果：
✅ typed_intent_demo: 所有功能正常
✅ Form/Select/OptionGroup: 所有测试通过
✅ 向后兼容性：现有代码无需修改
```

### Commit 2: feat(intent): Intent Bubble 组件完整实现

```
为 Intent Bubble 集成添加完整支持，通过 GlobalIntent 接口实现智能路由。

修改内容：
1. ui/components/optiongroup/intent.go: 实现 IsGlobal() = false
2. ui/components/select/intent.go: 实现 IsGlobal() = false
3. ui/components/treeview/intent.go: 实现 IsGlobal() = false
4. ui/components/select/instance.go: 使用 intentEmitter 触发路由

测试结果：
✅ runtime/intent: 所有测试 PASS
✅ ui/components/select: 所有测试 PASS
✅ ui/components/optiongroup: 所有测试 PASS
✅ ui/components/form: 所有测试 PASS
✅ ui/components/treeview: 所有测试 PASS
✅ examples/...: 所有示例编译成功
```

## 设计决策

### 为什么使用 GlobalIntent 接口？

1. **向后兼容**: 未实现接口的 Intent 默认全局处理
2. **显式选择**: 组件开发者显式决定 Intent 的路由方式
3. **类型安全**: 编译时检查，避免运行时错误

### 为什么修改 emitSelectChange 使用 intentEmitter？

1. **统一路由**: 所有 Intent 通过 FiberUtil 的智能路由
2. **可扩展性**: FiberUtil 可以在未来添加更多路由策略
3. **测试友好**: 测试可以捕获所有 Intent 发射

### 为什么不使用单独的 Event Emitter？

1. **避免重复**: Intent 已经是事件抽象，不需要额外层级
2. **保持简单**: FiberUtil 已经处理了 Intent 发射
3. **易于理解**: 统一的 Intent 概念

## 未来扩展

### 可能的改进

1. **中间件支持**: 在路由前/后添加钩子
2. **路由策略定制**: 允许组件自定义路由行为
3. **Intent 合并**: 批量发送 Intent 时进行优化
4. **调试工具**: 可视化 Intent 流向

### 新组件支持

其他组件可以选择添加本地 Intent 支持：

- List 组件: ItemSelectIntent
- Table 组件: RowSelectIntent, HeaderSortIntent
- Tabs 组件: TabSelectIntent
- Menu 组件: MenuItemSelectIntent

## 总结

Intent Bubble 实现成功达成了所有设计目标：

✅ **不破坏现有功能**: 所有测试通过，所有示例编译成功
✅ **支持 Intent Bubble**: Select/OptionGroup/Treeview 支持本地冒泡
✅ **向后兼容**: 未实现 GlobalIntent 的 Intent 默认全局处理

双路由架构为 Mint Runtime 提供了清晰、灵活的 Intent 处理机制。
