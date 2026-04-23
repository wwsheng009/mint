# Fiber架构集成状态与问题分析

## 概述

本文档记录了Fiber reconciler与mint UI框架集成的当前状态、存在的问题和下一步计划。

**创建时间**: 2026-02-01
**状态**: 进行中

---

## 问题现状

### ✅ 已解决的问题

1. **按钮显示正常** - 两个按钮（-和+）都能正确显示
2. **Tab切换正常** - 焦点可以在按钮间切换
3. **焦点样式显示** - 蓝色背景样式正确应用
4. **水平布局（HStack）** - 按钮水平排列正常

### ❌ 当前问题

1. **状态更新后UI不刷新**
   - setState被调用，hook.Value被更新
   - 但重新渲染时useState返回的还是旧值（value=0）
   - UI显示的count始终为0

2. **按钮点击可能无响应**
   - 未确认onClick是否被正确调用
   - HandleEvent可能未被调用

---

## 当前架构分析

### Fiber模式渲染流程

```
┌─────────────────────────────────────────────────────────────┐
│                    paintWithFiber                          │
├─────────────────────────────────────────────────────────────┤
│ 1. 重置元素收集: d.resetInteractiveElements()             │
│ 2. 设置渲染回调: d.reconciler.SetRenderCallback(...)        │
│ 3. reconciler.Render():                                    │
│    ├─ prepareFreshStack(renderFunc)                        │
│    │  ├─ r.ctx.ResetContext()                               │
│    │  ├─ SetCurrentContext(r.ctx)                          │
│    │  ├─ vnode := renderFunc()  // Counter函数              │
│    │  │  └─ useState() 使用 r.ctx 存储hooks                 │
│    │  └─ 创建/更新Fiber树                                  │
│    ├─ workLoopSync()                                        │
│    │  └─ BeginWork/CompleteWork 处理每个Fiber节点           │
│    └─ CommitRoot()                                         │
│       └─ renderFiberToBuffer() 遍历Fiber树并渲染           │
└─────────────────────────────────────────────────────────────┘
```

### Legacy模式渲染流程

```
┌─────────────────────────────────────────────────────────────┐
│                    paintLegacy                              │
├─────────────────────────────────────────────────────────────┤
│ 1. d.ctx.resetContext()                                    │
│ 2. SetCurrentContext(d.ctx)                                │
│ 3. vnode := d.appFn()  // Counter函数                        │
│ 4. renderVNode(vnode)  // 递归渲染                           │
│ 5. d.ctx.runEffects()                                        │
└─────────────────────────────────────────────────────────────┘
```

### 关键区别

| 方面 | Legacy模式 | Fiber模式 |
|------|-----------|----------|
| Context | `d.ctx` (declarativeRoot的) | `r.ctx` (Reconciler的) |
| Hooks存储 | `d.ctx.Hooks` | `r.ctx.Hooks` |
| 组件展开 | renderVNode递归处理 | beginWorkComponent处理 |
| 渲染方式 | 递归遍历VNode树 | 遍历Fiber树 |

---

## 问题分析

### 问题1: 状态更新后UI不刷新

**现象**:
```
useState: componentID=App-comp-1, hookIndex=1, value=0  // 始终返回0
getOrCreateHook: EXISTS, hookIndex=0, value=0             // hook存在但值未更新
```

**调试发现**:
1. `getOrCreateHook`返回已存在的hook
2. 但hook.Value始终为0
3. 没有看到"setState"的调试输出（onClick可能未被调用）

**可能原因**:
1. **指针失效** - hooks切片扩容导致指针失效
2. **闭包捕获** - setState捕获的hook指针在重新渲染时已失效
3. **Context不一致** - setState和useState使用了不同的context实例

### 问题2: 点击事件处理

**现象**:
- 没有"HandleEvent"调试输出
- 没有"Enter pressed"调试输出

**分析**:
- HandleEvent可能未被调用
- 或事件路由在Fiber模式下有问题

---

## 关键文件

### 核心文件

| 文件 | 作用 | 关键问题 |
|------|------|---------|
| `ui/app.go` | declarativeRoot和渲染逻辑 | paintWithFiber集成 |
| `ui/reconciler.go` | Reconciler实现 | context管理 |
| `ui/begin_work.go` | BeginWork阶段 | 组件展开逻辑 |
| `ui/hooks.go` | Hooks实现 | useState/setState |
| `ui/diff.go` | Diff算法 | reconcileChildren |

### 相关文件

| 文件 | 作用 |
|------|------|
| `ui/fiber.go` | Fiber数据结构 |
| `ui/vnode.go` | VNode类型定义 |
| `ui/instance_manager.go` | 组件实例管理 |
| `ui/interaction_state.go` | 交互状态管理 |

---

## Fiber机制说明

### Fiber树结构

```
Fiber
├── VNode          // 虚拟DOM节点
├── Type           // VNode类型 (Component/Element/Text/Fragment)
├── Key            // 唯一标识（用于reconciliation）
├── Props          // 属性
├── Lanes          // 优先级（ lanes）
├── Flags          // 副作用标记
├── Child          // 第一个子节点
├── Sibling        // 下一个兄弟节点
├── Return         // 父节点
├── Alternate      // 双缓冲（current/work-in-progress）
└── ComponentInstance // 组件实例（用于hooks状态）
```

### Reconciliation流程

```
1. prepareFreshStack: 创建VNode，创建Fiber树
2. workLoopSync: 遍历Fiber树
   ├─ BeginWork: Reconcile children
   └─ CompleteWork: 收集effects
3. CommitRoot: 渲染到buffer
```

### Key-based Reconciliation

```go
// shouldUpdate 决定是否可以复用Fiber节点
func shouldUpdate(current *Fiber, vnode VNode) bool {
    // 1. Key是主要标识 - 不同key = 不同元素
    if currentKey != newKey { return false }
    // 2. Type是次要标识 - 同key但不同type = 替换
    if current.Type != vnode.Type() { return false }
    // 3. 组件名必须相同
    if current.Type == VNodeComponent {
        return currentComp.Name() == newComp.Name()
    }
    return true
}
```

---

## 待解决问题

### 优先级1: 状态更新后UI不刷新

**调试步骤**:
1. 确认setState是否被调用
2. 确认hook.Value是否被正确更新
3. 确认重新渲染时是否读取到正确的hook.Value
4. 检查hooks切片指针是否失效

**可能的解决方案**:
- 方案A: 在getOrCreateHook中返回slice索引而不是指针
- 方案B: 使用map存储hooks而不是slice
- 方案C: 确保hooks切片不会扩容

### 优先级2: 事件处理

**调试步骤**:
1. 确认HandleEvent是否被调用
2. 确认按键事件是否被正确路由
3. 确认onClick是否被正确调用

---

## 环境变量

| 变量 | 值 | 作用 |
|------|---|----|
| `MINT_USE_FIBER` | `true` | 启用Fiber模式 |
| `TUI_DEBUG_UI` | `true` | 启用调试输出 |
| `TUI_DEBUG_KEYS` | `true` | 启用Key验证警告 |

---

## 调试命令

```bash
# 运行counter示例
MINT_USE_FIBER=true TUI_DEBUG_UI=true go run ./examples/counter/main.go

# 构建并运行
go build -o counter.exe ./examples/counter/main.go
MINT_USE_FIBER=true TUI_DEBUG_UI=true ./counter.exe
```

---

## 下一步计划

1. **修复useState指针问题** - 确保setState更新能正确反映到useState
2. **验证事件处理** - 确认HandleEvent在Fiber模式下正常工作
3. **完善LayoutNode处理** - 确保padding、gap等属性正确应用
4. **简化Fiber集成** - 减少对legacy逻辑的影响

---

## 参考文档

- [Fiber Architecture](../design/fiber-architecture.md)
- [Reconciliation Algorithm](../design/reconciliation.md)
- [Hooks Implementation](../design/hooks.md)
