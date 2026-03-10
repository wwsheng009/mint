# Mint UI 当前问题分析与重构方案

> ⚠️ **DEPRECATED** - 本文档已归档
>
> **当前状态**: **Store + Reducer 架构已完整实现 (93%)，所有 P0 问题已修复**
> **状态报告**: [`/docs/ui/store/status/CURRENT_STATUS.md`](../../ui/store/status/CURRENT_STATUS.md)
> **实现审查**: [`/docs/ui/store/reviews/IMPLEMENTATION_REVIEW.md`](../../ui/store/reviews/IMPLEMENTATION_REVIEW.md)
>
> 本文档记录了架构演进过程中的问题分析和重构方案。现在 Store + Reducer 架构已成熟，这些问题已解决。

---

**创建时间**: 2026-03-04
**归档时间**: 2026-03-08
**版本**: v1.0 (历史版本)
**范围**: 基于 store 和 mvp 架构文档与实际遇到的问题

---

## 一、执行摘要

### 1.1 问题本质

当前 Mint UI 在实际使用中暴露出一系列架构问题，这些问题暴露了以下核心矛盾：

```
设计文档愿景 vs 实际实现之间的断层
├─ Store 架构设计：单一数据源 + Reducer 模式 → 实际：三重状态系统
├─ MVP 架构设计：State 为事实源 → 实际：闭包依赖 + GlobalState 中转
├─ Intent 设计：声明式 → 实际：参数化 + 类型断言混乱
└─ 文档导向：从实现角度 → 实际：需要大量隐式知识
```

### 1.2 修复优先级

| 优先级 | 问题 | 影响范围 | 修复复杂度 |
|--------|------|---------|-----------|
| 🔴 P0 | 输入框无法输入 | 所有表单组件 | 中 |
| 🔴 P0 | Button Click Count 固定 | 有状态按钮 | 低 |
| 🟡 P1 | Checkbox 无法响应 | 表单组件 | 低 |
| 🟡 P1 | ClickIntent 无 handler | 所有按钮 | 低 |
| 🟢 P2 | 三重状态系统复杂 | 全局架构 | 高 |
| 🟢 P2 | 类型断言混乱 | Intent handlers | 中 |
| 🟢 P2 | 文档不完整 | 新手体验 | 中 |

---

## 二、实际遇到的问题清单

### 2.1 输入框无法输入文字

#### 问题表现
```
用户输入文字 → Instance 值更新 → 但显示不变
```

#### 错误代码
```go
// ❌ 错误实现
input.SetChangeIntent(intent.SetState("input1-value", "entered"))
//                     ^^^^^^^^^^^^^^^^^^^^^^
//                     意图：设置为一个固定值 "entered"
//                     实际：用户每次输入都被覆盖为 "entered"
```

#### 根本原因

| 层面 | 设计文档 | 实际代码 |
|------|---------|---------|
| **Intent 设计** | FieldChangeIntent {Field, Value} = 动态值 | SetStateIntent {Key, Value} = 静态值 |
| **Instance 行为** | 自动发射 FieldChangeIntent | 直接发射用户传入的 Intent |
| **Handler 映射** | 使用内置 `handleFieldChange` | 误用 `handleSetState` |

**问题类型**: **文档-实现不一致** + **API 混淆**

---

### 2.2 Button Click Count 固定为 1

#### 问题表现
```
按钮按一次 → count = 1
按钮再按 → count 仍然是 1
```

#### 错误代码
```go
// ❌ 错误实现（方案 1 - 误用 Toggle）
button.OnPress(intent.Toggle("btn-click-count"))
// 问题：Toggle 切换 boolean，不能递增 int

// ❌ 错误实现（方案 2 - 误用 GlobalState）
ui.RegisterIntent(func(ctx *intent.ActionContext, i ToggleIntent) intent.IntentResult {
    currentCount := 0
    if v, ok := ctx.GetState("btn-click-count"); ok {  // ❌ 永远未设置
        if c, ok := v.(int); ok {
            currentCount = c
        }
    }
    setter(currentCount + 1)  // ❌ 永远是 0 + 1 = 1
})
```

#### 根本原因

| 层面 | 设计文档要求 | 实际代码问题 |
|------|-------------|-------------|
| **Intent 类型** | 使用自定义 Intent 或 UseStateInt+闭包 | 错误使用 Toggle（类型不匹配） |
| **状态读取** | getter() 函数或 functional update | 手动从 GlobalState 读取（未初始化） |
| **Setter 类型** | 支持 int 或 func(int) int | 类型断言混乱，使用了错误的类型 |

**问题类型**: **API 误用** + **时序依赖** + **类型系统混乱**

---

### 2.3 Checkbox 无法响应

#### 问题表现
```
按 SPACE → checkbox 视觉状态切换 → 但 Hook 状态未更新
```

#### 错误代码
```go
// ❌ 错误实现
checkbox.OnToggle(intent.Toggle("chk1-checked"))
// 问题：只发射 Toggle intent，没有绑定 Hook 状态
```

#### 根本原因

| 层面 | 设计文档要求 | 实际代码问题 |
|------|-------------|-------------|
| **MVP 模式** | ForField(StateKey) + Checked(hookValue) | 只用 OnToggle，缺少 binding |
| **数据流** | FieldChangeIntent → Hook update | Toggle intent 只更新 GlobalState |
| **Handler** | FieldChangeIntent type-safe | 使用 wrong intent type |

**问题类型**: **MVP 模式理解偏差** + **API 不完整**

---

### 2.4 ClickIntent 警告

#### 问题表现
```
[IntentEmitter] Failed to emit intent Click: no handler registered for intent type: Click
```

#### 错误代码
```go
// ❌ 文档示例错误
button.OnPress(intent.Click("btn1"))  // ← 无 handler
```

#### 根本原因

| Intent | 内置 Handler | 文档建议 | 实际可用性 |
|--------|-------------|---------|-----------|
| `SetState` | ✅ | ✅ | ✅ |
| `Toggle` | ✅ | ✅ | ✅ |
| `FieldChange` | ✅ | ✅ | ✅ |
| `Click` | ❌ | ✗ 使用（误导） | ❌ 警告 |
| `Press` | ❌ | ✗ 未提及 | ❌ 警告 |

**问题类型**: **文档误导** + **意图不清晰**

---

## 三、架构层面存在的根本问题

### 3.1 状态系统的三重复杂性

#### 当前架构

```
┌─────────────────────────────────────────────────────┐
│  三重状态系统（实际实现）                           │
├─────────────────────────────────────────────────────┤
│                                                     │
│  1. UseState (Component Local)                     │
│     ├─ 存储: ctx.Hooks[hookIndex].Value           │
│     ├─ Setter: 闭包函数                          │
│     └─ 自动触发重渲染 ✅                           │
│                                                     │
│  2. GlobalState (Cross-component)                 │
│     ├─ 存储: ctx.GlobalState[key]                  │
│     ├─ Setter: ctx.SetState(key, value)           │
│     └─ 需要手动 ScheduleUpdate ❌                 │
│                                                     │
│  3. Instance (Component Internal)                 │
│     ├─ 存储: inst.value, inst.checked             │
│     ├─ 特点: 临时缓冲                             │
│     └─ 需要 Intent 同步到其他状态 ❌              │
│                                                     │
└─────────────────────────────────────────────────────┘

数据同步链路：
Instance → FieldChangeIntent → GlobalState → UseState
  ↑                                                    ↓
  └──────────────── VNode 渲染 ───────────────────────┘

复杂性：3 种状态 + 2 次同步 + 手动类型转换
```

#### 设计文档愿景

```
┌─────────────────────────────────────────────────────┐
│  单一状态源（设计文档）                              │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Store (Single Source of Truth)                    │
│     ├─ 存储: 统一的状态树                         │
│     ├─ 更新: Reducer(Action) → New State          │
│     └─ 绑定: 声明式订阅                          │
│                                                     │
│ 数据流：                                            │
│     Event → Action → Reducer → Store → VNode       │
│     Instance: 纯缓冲，不维护业务状态 ❌             │
│     Hooks: 自动订阅 Store ✅                        │
│                                                     │
└─────────────────────────────────────────────────────┘

简洁性：1 种状态 + 0 次手动同步 + 类型安全
```

**问题**: 设计文档愿景与实际实现存在巨大差距

---

### 3.2 Intent 系统的不一致性

#### 内置 Intent 清单

| Intent | Handler | 适用场景 | 参数类型 | 使用复杂度 |
|--------|--------|---------|----------|-----------|
| `SetState` | ✅ | 绝对状态设置 | Key, Value | 低 |
| `Toggle` | ✅ | Boolean 切换 | Key | 低 |
| `Increment` | ✅ | 数值递增 | Key, Delta | 低 |
| `FieldChange` | ✅ | 表单字段变更 | Field, Value | 低 |
| `Navigate` | ✅ | 页面导航 | Path, Params | 低 |
| `Focus`/`Blur` | ✅ | 焦点管理 | TargetID | 低 |
| `OpenModal`/`CloseModal` | ✅ | Modal 管理 | ModalID | 低 |
| **`Click`** | ❌ | **按钮点击** | TargetID | ❌ 需自定义 |
| **`Press`** | ❌ | **通用按压** | TargetID | ❌ 需自定义 |

#### 问题
1. `Click` 是最基本的操作，但没有 handler
2. Intent 清单不完整：文档中提及但未实现
3. 缺少"Intent 分类指南": 哪些有 handler？哪些没有？

---

### 3.3 MVP 模式的学习曲线陡峭

#### 实现 MVP 模式所需的步骤

```go
// 第 1 步：创建 UseState
value, setValue := ui.UseStateString("")

// 第 2 步：保存 setter 到 GlobalState（为什么？？？）
ctx.GlobalState["field-setter"] = setValue

// 第 3 步：注册 Handler（在 WithInit 中）
ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("field-setter"); ok {
            if setter, ok := fn.(func(string)); ok {  // ❌ 类型断言
                setter(i.Value)
            }
        }
        return intent.HandledResult()
    })
})

// 第 4 步：组件使用 ForField
input.ForField(intent.BindField("field")).
    Value(value).
    Build()

// 第 5 步：处理 UseStateInt 的 special case（functional update）
setter(func(c int) int { return c + 1 })
```

#### 问题
1. 为什么需要把 setter 存到 GlobalState？
   - 设计文档：简化跨组件访问
   - 实际效果：增加复杂度和类型断言

2. 为什么不直接用闭包？
   - 设计文档：防止过期闭包
   - 实际问题：WithInit 在 render 之前执行，setter 还未创建

3. FieldChangeIntent vs handleFieldChange 的关系是什么？
   - 设计文档：内置 handler 自动处理
   - 实际使用：仍然需要手动注册 handler

---

### 3.4 类型系统的混乱

```go
// UseStateInt 的 setter 可以接受 3 种类型：
setClickCount(5)                              // int
setClickCount(SetIntFunc(func(c int) int {   // SetIntFunc
    return c + 1
}))
setClickCount(func(c int) int { return c + 1 })  // raw function

// 但类型签名只显示：
func UseStateInt(initial int) (int, func(interface{}), func() int)
//                                      ^^^^^^^^^^^^^^^
//                                  IDE 看起来是 func(interface{})
```

#### 问题
1. 类型签名不反映实际能力
2. 需要类型断言 + 类型转换
3. IDE 无法提供完整提示

---

### 3.5 文档与实现脱节

| 问题 | 文档描述 | 实际情况 |
|------|---------|---------|
| Click Intent | 在示例中使用 | ❌ 无 handler，警告 |
| MVP 最佳实践 | 简单的 3 步流程 | ❌ 需要 5 步 + 类型断言 |
| Intent 管理模式 | 3 种清晰模式 | ❌ 文档未提及注册方式 |
| 类型安全 | StateKey 泛型 | ❌ 仍需手动类型断言 |

---

## 四、与设计文档的对齐分析

### 4.1 Store 架构文档对齐度

| 设计原则 | 文档要求 | 实际实现 | 对齐度 |
|---------|---------|---------|--------|
| 单一数据源 | Store 是唯一真实来源 | UseState + GlobalState + Instance | 🔴 30% |
| 消除闭包 | Handler 不捕获 state | setter 存 GlobalState + 类型断言 | 🟡 50% |
| Reducer 模式 | Reducer(Action) → New State | handlers 直接更新 State | 🟡 40% |
| 类型安全 | StateKey[T] 泛型 | 存在类型断言 | 🟢 70% |

### 4.2 MVP 架构文档对齐度

| 设计原则 | 文档要求 | 实际实现 | 对齐度 |
|---------|---------|---------|--------|
| State 为事实源 | 所有状态在 State | Instance 临时缓冲 + 2 次同步 | 🟡 50% |
| Intent 携带最少数据 | FieldChangeIntent {Field, Value} | 实现正确 ✅ | 🟢 90% |
| Instance 是纯缓冲 | 不维护业务状态 | Instance 有内部状态 | 🟡 60% |
| ForField 绑定 | 自动发射 Intent | 实现正确 ✅ | 🟢 90% |

### 4.3 Intent 管理文档对齐度

| 设计模式 | 文档描述 | 实际实现 | 对齐度 |
|---------|---------|---------|--------|
| 方案 1: 组件级状态 | ui.On + Simple* Intent | 未实现 Simple* Intent | 🔴 20% |
| 方案 2: 全局状态 | runtime/intent 内置 | handler 存在但文档不全 | 🟡 60% |
| 方案 3: 自定义 Intent | 灵活扩展 | 可用但无示例 | 🟡 50% |

---

## 五、根本原因分析

### 5.1 架构演进断层

```
设计文档（理想状态）
    ↓
  实现初期（部分实现）
    ↓
   快速迭代（补丁式修复）
    ↓
当前状态（文档-实现不匹配）
```

**问题**: 设计文档是"目标架构"，但实现还停留在"中间态"

---

### 5.2 文档维护滞后

| 文档 | 最后更新 | 覆盖度 | 问题 |
|------|---------|--------|------|
| MVP_MIGRATION_GUIDE.md | 2026-02-26 | 70% | 示例代码与实际 API 不符 |
| INTENT_MANAGEMENT_PATTERNS.md | 2026-02-27 | 60% | 注册方式未说明 |
| COMPONENT_INTENT_REVIEW.md | unknown | 40% | 缺少实现细节 |

---

### 5.3 类型系统限制

Go 的类型系统限制了函数式编程的表达能力：

```go
// React (JS)
const [count, setCount] = useState(0)
setCount(c => c + 1)  // 函数式更新，自然

// Mint (Go)
clickCount, setClickCount, _ := ui.UseStateInt(0)
setClickCount(func(c int) int { return c + 1 })  // 类型签名不清晰
```

---

## 六、重构方案

### 6.1 短期修复（P0/P1 问题）

#### 修复 1: 输入框无法输入

**问题**: `SetState` Intent 固定值覆盖用户输入

**方案**: 统一使用 `FieldChangeIntent`

```go
// ✅ 正确实现
input.ForField(intent.BindField("field")).
    Value(value).  // 绑定状态
    Build()

// Handler（内置 handleFieldChange 或自定义）
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    if fn, ok := ctx.GetState(i.Field + "-setter"); ok {
        if setter, ok := fn.(func(string)); ok {
            setter(i.Value)
        }
    }
    return intent.HandledResult()
})
```

#### 修复 2: Button Click Count 固定

**问题**: 错误使用 Toggle Intent + GlobalState 未初始化

**方案**: 使用 Functional Update

```go
// ✅ 正确实现
clickCount, setClickCount, _ := ui.UseStateInt(0)
ctx.GlobalState["setClickCount"] = setClickCount

ui.RegisterIntent(func(ctx *intent.ActionContext, i ClickButtonIntent) intent.IntentResult {
    if fn, ok := ctx.GetState("setClickCount"); ok {
        if setter, ok := fn.(func(interface{})); ok {
            setter(func(c int) int { return c + 1 })  // ✅ Functional update
        }
    }
    return intent.HandledResult()
})
```

#### 修复 3: Checkbox 无法响应

**问题**: 缺少 ForField 绑定

**方案**: 添加 ForField + Checked

```go
// ✅ 正确实现
checked, setChecked := ui.UseStateBool(false)
ctx.GlobalState["checked-setter"] = setChecked

checkbox.ForField(intent.BindField("checked")).
    Checked(checked).
    Build()

// Handler（FieldChangeIntent）
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    if i.Field == "checked" {
        if fn, ok := ctx.GetState("checked-setter"); ok {
            value := i.Value == "true"
            if setter, ok := fn.(func(bool)); ok {
                setter(value)
            }
        }
    }
    return intent.HandledResult()
})
```

#### 修复 4: ClickIntent 警告

**问题**: Click Intent 无 handler

**方案 A**: 添加内置 handler（简化方案）

```go
// runtime/intent/builtin_handlers.go
func handleClick(ctx *ActionContext, i ClickIntent) IntentResult {
    // 为按钮点击提供默认处理（无操作）
    // 用户可以覆盖
    return HandledResult()
}

// SetupBuiltinHandlers 中添加
RegisterTypedRuntime(rt, handleClick)
```

**方案 B**: 文档明确说明（推荐）

更新 `MVP_MIGRATION_GUIDE.md`，明确说明：
- `Click`/`Press` Intent 没有内置 handler
- 推荐使用自定义 Intent 推荐的内置 Intent

---

### 6.2 中期优化（架构简化）

#### 优化 1: 简化状态管理

**目标**: 减少手动状态同步

**方案**:

```go
// 提议：自动状态绑定 API
type StateBinding[T any] interface {
    Bind(key string) (T, func(T))
    Update(T)
}

// 使用
inputBinder := ui.UseStateBinding[string]("username")  // 自动创建
username, setUsername := inputBinder.Get(), inputBinder.Set()

// 自动注册 handler（无需 WithInit）
input.ForField(inputBinder.BindKey())  // 自动处理
```

#### 优化 2: 消除 GlobalState 中转

**问题**: setter 存 GlobalState 是复杂的根源

**方案**:

```go
// 提议：Handler 从 ComponentContext 直接获取 setter
ui.RegisterIntent(func(ctx *intent.ActionContext, i FieldChangeIntent) intent.IntentResult {
    // 从 ComponentContext 查找状态 setter
    if setter, ok := ctx.GetSetter(i.Field); ok {
        if stringSetter, ok := setter.(func(string)); ok {
            stringSetter(i.Value)
        }
    }
    return intent.HandledResult()
})

// ComponentContext 新方法：
func (ctx *ComponentContext) GetSetter(field string) (interface{}, bool)
```

#### 优化 3: 统一 Intent 分类文档

**方案**: 创建 `INTENT_CATALOG.md`

```markdown
# Intent 分类索引

## 已实现 Handler ✅
- SetState
- Toggle
- Increment
- FieldChange
- Navigate
- Focus/Blur

## 需要自定义 Handler ❌
- Click（推荐：自定义 Intent）
- Press（推荐：使用 Toggle 或自定义）

## 推荐使用方式
1. 表单输入：ForField + FieldChangeIntent
2. 状态切换：Toggle
3. 数值递增：Increment
4. 自定义逻辑：自定义 Intent
```

---

### 6.3 长期重构（架构演进）

#### 目标 1: 统一 Store 架构

**设计**: 
```
当前: UseState + GlobalState + Instance → 三重状态
目标: Store (Single Source of Truth)
```

**方案**:
```go
// 建议的 Store API
type Store[T any] interface {
    Get() T
    Set(T)
    Update(func(T) T)
    Subscribe(func(T)) Unsubscribe
}

// 使用
type AppState struct {
    Username string
    Email    string
    Count    int
}

store := ui.NewStore(AppState{Username: "", Email: "", Count: 0})

// Hooks 自动订阅
username := ui.UseStore(store, func(s AppState) string { return s.Username })

// Actions
ui.Dispatch(UpdateUsernameAction{Value: "john"})
```

#### 目标 2: 类型安全 Intent DSL

**设计**: 
```
当前: Intent + 类型断言 + 手动类型转换
目标: 编译期类型安全的 Intent DSL
```

**方案**:
```go
// 建议的 DSL DSL
intent := UpdateUsername("john")  // 编译期类型检查
ui.RegisterHandler(UpdateUsername, handleUsername)

// 自动类型推断，无需类型断言
func handleUsername(ctx Context, i UpdateUsername) Result {
    ctx.SetUsername(i.Value)  // 类型安全
    return Handled
}
```

---

## 七、实施路线图

### Phase 1: 紧急修复（1-2 天）

| 任务 | 优先级 | 预估时间 |
|------|--------|---------|
| 修复输入框无法输入 | 🔴 P0 | 2h |
| 修复 Button Count 固定 | 🔴 P0 | 1h |
| 修复 Checkbox 无法响应 | 🟡 P1 | 1h |
| ClickIntent 处理（添加 handler 或文档） | 🟡 P1 | 1h |

### Phase 2: 架构简化（1 周）

| 任务 | 优先级 | 预估时间 |
|------|--------|---------|
| 实现 StateBinding API | 🟡 P2 | 1d |
| 消除 GlobalState 中转 | 🟡 P2 | 2d |
| 统一 Intent 分类文档 | 🟡 P2 | 0.5d |
| 更新所有示例代码 | 🟡 P2 | 2d |

### Phase 3: 长期演进（2-4 周）

| 任务 | 优先级 | 预估时间 |
|------|--------|---------|
| 设计并实现 Store API | 🟢 P3 | 1w |
| 设计并实现 Intent DSL | 🟢 P3 | 1w |
| 提供迁移工具 | 🟢 P3 | 0.5w |
| 更新所有文档 | 🟢 P3 | 0.5w |

---

## 八、关键决策点

### 决策 1: 是否添加 Click Intent handler？

| 方案 | 优点 | 缺点 | 建议 |
|------|------|------|------|
| 添加 handler | 消除警告，向后兼容 | 增加内置 Intent 数量 | ⚠️ 不推荐 |
| 文档说明 | 保持清晰，鼓励自定义 | 需要用户学习 | ✅ 推荐 |

**理由**: `Click` 作为通用操作，其语义应由应用定义，而非框架提供。保持清晰的 Intent 分类边界更重要。

### 决策 2: 是否保留 GlobalState 中转？

| 方案 | 优点 | 缺点 | 建议 |
|------|------|------|------|
| 保留 | 跨组件访问容易 | 增加复杂度和类型断言 | ⚠️ 短期保留 |
| 移除 | 简化架构，消除类型断言 | 需要新的跨组件机制 | ✅ 长期目标 |

**理由**: 当前移除成本高，建议先引入 StateBinding 等高级 API 平滑过渡。

### 决策 3: 统一 Store 还是渐进演进？

| 方案 | 优点 | 缺点 | 建议 |
|------|------|------|------|
| 一次性重构 | 架构清晰 | 破坏性大，风险高 | ⚠️ 不现实 |
| 渐进演进 | 降低风险，保持可用 | 多套 API 并存 | ✅ 推荐 |

**理由**: 当前系统已稳定，大规模重构风险太高。建议通过提供更好的高层 API 逐步引导用户迁移。

---

## 九、风险评估

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 破坏现有应用 | 高 | 中 | 提供兼容层，渐进迁移 |
| 性能下降 | 中 | 低 | 性能测试，优化热路径 |
| 文档更新不及时 | 中 | 高 | 建立 CI 检查文档-代码一致性 |
| 用户学习成本 | 高 | 中 | 提供迁移工具，保留旧 API |

---

## 十、总结

### 10.1 关键发现

1. **架构断层**: 设计文档描述的是"目标架构"，实现还停留在"中间态"
2. **三重状态系统**: UseState + GlobalState + Instance，导致复杂度和类型断言问题
3. **文档-实现不一致**: 很多文档示例无法直接编译运行
4. **类型系统限制**: Go 的类型系统限制了函数式编程的表达能力

### 10.2 核心建议

1. **短期**: 修复 P0/P1 问题，确保基本功能可用
2. **中期**: 简化状态管理，减少手动同步
3. **长期**: 统一 Store 架构，类型安全

### 10.3 愿景

将 Mint UI 从"组件驱动的 TUI 库"升级为"声明式 UI Runtime"：

```
当前: 组件 + 复杂的状态同步
目标: State (单一) + Intent (声明式) + Reducer (逻辑集中)
```

**关键转变**:
- 从: 事件 handler 捕获 state 执行逻辑
- 到: 事件只发 Intent，Reducer 读取最新 state 处理逻辑

---

## 附录

### A. 相关文档索引

| 文档 | 路径 | 重点内容 |
|------|------|---------|
| Store 重构计划 | `docs/architecture/store/REFACTOR_PLAN.md` | Store + Reducer 设计 |
| MVP 迁移指南 | `docs/architecture/mvp/MVP_MIGRATION_GUIDE.md` | ForField + FieldChangeIntent |
| Intent 管理模式 | `docs/architecture/mvp/INTENT_MANAGEMENT_PATTERNS.md` | 3 种 Intent 使用模式 |
| 实现复查 | `docs/architecture/store/IMPLEMENTATION_REVIEW.md` | 当前实现状态 |

### B. 问题修复检查清单

- [ ] 输入框可以正常输入
- [ ] Button Click Count 可以递增
- [ ] Checkbox 可以正常响应
- [ ] ClickIntent 警告已处理
- [ ] 示例代码全部可编译运行
- [ ] 示例代码与文档一致
- [ ] 文档更新到位
