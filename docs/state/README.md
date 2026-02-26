# Mint Fiber-first 状态管理文档

本目录包含 Mint Fiber-first 渲染模式下状态管理的完整文档。

## 文档目录

| 文档 | 描述 |
|------|------|
| **[FIBER_STATE_ARCHITECTURE.md](./FIBER_STATE_ARCHITECTURE.md)** | 核心架构说明 - 完整的状态管理机制、Fiber 树构建流程、Intent 系统和 InstanceManager 详解 |
| **[BEST_PRACTICES.md](./BEST_PRACTICES.md)** | 最佳实践指南 - 如何选择正确的状态类型、常见架构模式、性能优化技巧和代码示例 |
| **[MIGRATION.md](./MIGRATION.md)** | 迁移指南 - 从闭包模式迁移到纯状态模式的完整步骤、对比和常见问题解答 |

## 快速导航

### 新手入门

1. 首先阅读 **[FIBER_STATE_ARCHITECTURE.md](./FIBER_STATE_ARCHITECTURE.md)** 了解核心概念
2. 然后阅读 **[BEST_PRACTICES.md](./BEST_PRACTICES.md)** 学习最佳实践

### 迁移现有代码

1. 阅读 **[MIGRATION.md](./MIGRATION.md)** 了解迁移步骤
2. 参考迁移前后的代码对比
3. 使用迁移检查清单

### 深入理解

- **状态流转**：`VNode → Fiber → LayoutBox → PaintableBox`
- **双重状态管理**：Hooks（局部）vs Global（全局）
- **Intent 系统**：单向数据流和事件批处理

## 关键概念

### Fiber 树结构

```
VNode (声明式描述)
    ↓
Fiber (运行时实例)
    ↓
LayoutBox (布局计算)
    ↓
PaintableBox (绘制)
```

### 状态存储层级

| 层级 | 类型 | 用途 |
|-----|------|------|
| Hooks 状态 | `[]Hook` | useState/useEffect 的局部值 |
| 全局状态 | `map[string]interface{}` | Intent Handler 更新的全局数据 |
| Props | `Props` | 父组件传递的数据 |
| MemoizedProps | `Props` | Props 的副本用于 diff |
| Render 缓存 | `interface{}` | Text 内容、Layout 结果等 |

### 状态类型选择

- **useState**：组件内部状态（toggle、input 值、展开/折叠）
- **全局状态**：跨组件状态（当前步骤、用户信息、表单数据）
- **Props**：父子通信（列表项、模态框标题、图表数据）

## 代码示例

### 组件内部状态

```go
func Counter() ui.VNode {
    count, setCount := rtui.UseState(0)

    return ui.Button(fmt.Sprintf("Count: %d", count)).
        OnClick(func() { setCount(count + 1) }).
        Build()
}
```

### 全局状态

```go
type UpdateStepIntent struct {
    Step int
}
func (UpdateStepIntent) IntentType() string { return "UpdateStep" }

ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
    ctx.SetState("step", i.Step)
    return intent.HandledResult()
})

func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)

    return ui.Button("Next").
        OnPress(UpdateStepIntent{Step: step + 1}).
        Build()
}
```

## 优化亮点

### 批量更新

```go
// ✅ 多次 SetState 会被自动批处理
ctx.SetState("field1", "value1")
ctx.SetState("field2", "value2")
ctx.SetState("field3", "value3")
// 这些更新会被合并，只触发一次重新渲染
```

### 语义化 API

```go
// 新增的语义化方法
ctx.SetGlobalState(key, value)       // 设置全局状态
ctx.GetGlobalState(key, default)     // 获取全局状态
ctx.GetGlobalInt(key, default)       // 获取 int 类型的全局状态
ctx.GetGlobalString(key, default)    // 获取 string 类型的全局状态
ctx.GetGlobalBool(key, false)        // 获取 bool 类型的全局状态
```

## 相关资源

- **项目根目录**: `E:\projects\yao\wwsheng009\mint`
- **核心代码**:
  - `runtime/ui/hooks.go` - Hooks 和状态上下文
  - `internal/state/instance_manager.go` - 组件实例管理
  - `runtime/ui/reconcile/*.go` - Fiber 调度器

---

**最后更新**: 2026-02-26
