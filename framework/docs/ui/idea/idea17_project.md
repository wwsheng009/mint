好，这一步我们把 **v0.1 的“抽象体系”彻底压扁成“可以直接开工的工程结构”**。

这是非常关键的一步，因为**目录结构 = 心智模型**。
结构一旦定型，后面的人（包括未来的你）都会顺着它写代码。

---

# 🛠 v0.1 工程目录结构设计（可直接落地）

目标：

* 模块边界清晰
* SDK / Engine 强隔离
* 后续可演进，不推翻

---

## 🧱 顶层结构（强烈建议）

```
mint-ui/
├── cmd/
│   └── mint/            # CLI / 示例入口
│
├── internal/            # Engine 内核（不对外）
│   ├── runtime/
│   ├── render/
│   ├── layout/
│   ├── scheduler/
│   ├── state/
│   ├── event/
│   ├── buffer/
│   └── platform/
│
├── ui/                  # SDK（对开发者）
│   ├── app.go
│   ├── view.go
│   ├── hooks.go
│   ├── layout.go
│   ├── style.go
│   ├── components/
│   └── debug.go
│
├── components/          # 内置高阶组件
│   ├── list/
│   ├── input/
│   ├── modal/
│   ├── table/
│   └── progress/
│
├── examples/            # 示例（极其重要）
│   ├── counter/
│   ├── dashboard/
│   └── logs/
│
├── docs/
│   ├── getting-started.md
│   ├── cookbook.md
│   └── architecture.md  # 给贡献者，不是用户
│
├── go.mod
└── README.md
```

---

# 🔥 关键设计思想（请认真看）

## 1️⃣ `internal/` = 内核态（禁止 SDK 直接引用）

**规则：**

> `ui/` 不能 import `internal/*`

这是为了防止：

* SDK 泄漏引擎概念
* API 被底层绑死

---

## 2️⃣ `ui/` = 对外唯一入口

开发者 **只接触这里**。

```go
import "mint/ui"
```

SDK 内部可以偷偷用 engine，但用户永远不知道。

---

# 🧠 内核模块拆解（internal）

### `internal/runtime`

职责：

* App 生命周期
* Root RNode
* 主循环

```go
type Runtime struct {
    Scheduler *Scheduler
    Renderer  *Renderer
}
```

---

### `internal/scheduler`

职责：

* Dirty queue
* 批处理
* 帧调度

> ❗这里是性能中枢

---

### `internal/state`

职责：

* Hook 状态
* setState 队列
* Effect 生命周期

---

### `internal/layout`

职责：

* Measure / Layout
* Flex 算法
* Scroll 计算

不允许：

* 状态更新
* IO
* 动画

---

### `internal/render`

职责：

* Paint
* DrawCmd
* Clip
* Z-layer

---

### `internal/buffer`

职责：

* Buffer
* Cell
* Unicode 宽度
* Diff
* ANSI 输出

---

### `internal/event`

职责：

* 输入解析
* Focus 管理
* Hit Test
* Event 派发

---

### `internal/platform`

职责：

* OS / Terminal 差异
* Windows / Unix
* Resize 事件

---

# 🧩 SDK 层（ui/）设计细节

### `ui/app.go`

```go
func Run(app func()) {
    runtime := runtime.New()
    runtime.Run(app)
}
```

---

### `ui/view.go`

```go
func View(fn func()) {
    currentRuntime.Render(fn)
}
```

负责：

* 开启 VNode 构建上下文
* 绑定当前 Fiber

---

### `ui/hooks.go`

```go
func UseState[T any](init T) (T, func(T))
func UseEffect(fn func() func(), deps []any)
```

只暴露最少 hook。

---

### `ui/layout.go`

```go
func Row(children ...Node) Node
func Column(children ...Node) Node
func Box() *BoxNode
```

---

### `ui/style.go`

```go
func (n Node) Color(c Color) Node
func (n Node) Bold() Node
```

---

### `ui/components/*`

包装内置组件，最终还是 Node。

---

# 📦 内置组件与 SDK 的关系

```
internal/   ← engine
ui/         ← API 封装
components/ ← 高阶复合组件
```

组件永远用 `ui` 写，**不允许直连 internal**。

---

# 🧠 示例项目是“第二 API 文档”

`examples/` 的重要性 ≈ SDK 本身。

必须做到：

* 复制即跑
* 不依赖隐藏技巧
* 展示最佳实践

---

# 🚨 一个非常重要的工程纪律

> ❗ **任何新功能，先想“放在哪个目录”**

如果你犹豫：

* 是 internal 还是 ui？
* 是 runtime 还是 scheduler？

说明这个功能 **设计还不清晰**，不要写。

---

# 🎯 到这里你已经完成了什么？

你现在已经具备：

* 明确的发布目标（v0.1）
* 冻结的功能范围
* 可执行的工程结构
* 清晰的模块边界
* 可扩展的演进路线

这已经不是“讨论方案”，而是：

> **随时可以开仓、开工、开源的状态**

---

## 如果你愿意继续，下一步是最后一块「落地动作」：

> ✅ **v0.1 第一周开发清单（Day-by-Day ToDo）**

这是把“大工程”拆成 **今天就能写的代码**。
