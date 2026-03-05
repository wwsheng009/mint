# fmt.Println 在终端 UI 应用中 Buffer 刷新问题

## 问题描述

在 Mint UI 应用中使用 `fmt.Println` 时，输出显示异常：

```
Current State:
  Username:
  Email:
  Age:         0
==Accept Terms: false
UsSubscribe:   false


=== Form Submission ===
Username:   false
Email:      ============
Age:        0===========
Accept T&C: false
Subscribe:  false
========================
```

可以看到输出被**截断、重叠或混入控制字符**。

---

## 根本原因

### 1. os.Stdout 被 UI 渲染占用

Mint UI 框架会直接向 `os.Stdout` 写入**大量控制字符**用于渲染 UI：

```go
// platform/screen.go
func (s *DefaultScreen) Write(data []byte) (int, error) {
    return s.file.Write(data)  // file = os.Stdout
}

// 写入的控制字符包括：
// \x1b[2J      - 清屏
// \x1b[?1049h   - 进入备用屏幕
// \x1b[H        - 光标归位
// \x1b[10;20H   - 移动光标到(10,20)
// ... 数千个控制字符每秒
```

### 2. 标准输出缓冲机制改变

Go 的标准输出默认是**行缓冲**（基于 `\n`），但在终端 UI 场景下：

| 情况 | 标准 Go 程序 | Mint UI 应用 |
|------|------------|-------------|
| **缓冲模式** | 行缓冲 | 无缓冲（直接写入）|
| **os.Stdout** | 独占使用 | UI 渲染混用 |
| **控制字符** | 无 | 大量 |
| **刷新时机** | 遇到 `\n` 自动 | 手动调用 Flush |

### 3. 输出竞争

```
┌─────────────────────────────────────────────┐
│  时间轴                                      │
├─────────────────────────────────────────────┤
│  T1: UI 渲染 → os.Stdout (控制字符)         │
│  T2: fmt.Println → os.Stdout (用户输出)     │
│  T3: UI 渲染 → os.Stdout (覆盖 T2)          │
│  T4: fmt.Println → os.Stdout (用户输出)     │
└─────────────────────────────────────────────┘
```

UI 渲染和用户输出交替写入 `os.Stdout`，导致输出被覆盖或错位。

---

## 代码证据

### Mint UI 如何使用 os.Stdout

```go
// runtime/platform/screen.go:57
func (s *DefaultScreen) Write(data []byte) (int, error) {
    return s.file.Write(data)  // file = os.Stdout
}

// 渲染循环每帧调用：
// 1. Write("\x1b[2J")           - 清屏
// 2. Write("[光标移动]")        - 定位
// 3. Write("[UI 组件]")         - 绘制
// 4. Write("[更多控制字符]")    - 高亮、样式等
// ... 每秒 60 次
```

### 用户输出与之竞争

```go
// 用户代码
func handleSubmit(s AppState, i intent.Intent) AppState {
    fmt.Println("\n=== Form Submission ===")  // ← 写入 os.Stdout
    fmt.Printf("Username: %v\n", s.Username) // ← 写入 os.Stdout
    // ... 更多 fmt.Println
    return s
}

// 同时 UI 渲染循环也在写入 os.Stdout
// 结果：输出被 UI 渲染的 ANSI 控制字符"污染"
```

---

## 为什么 buffer 无法刷新？

### 问题不是 buffer 刷新，而是输出被覆盖

```
用户理解的"buffer 无法刷新"现象：
  输出显示：Username: ============

实际原因：
  1. fmt.Println 写入 "Username: user@example.com\n"
  2. 紧接着 UI 渲染写入 "\x1b[10;20Hsome_value"（移动光标、覆盖内容）
  3. 结果看起来像是 buffer 没有刷新，实际是内容被 UI 渲染覆盖

根本：
  os.Stdout 被多个写入源同时使用，形成输出竞争
```

### 验证方法

查看终端输出可以看到：
- `\x1b[` (ANSI 转义序列) 直接出现在输出中
- 输出被截断（因为光标被移动了）
- 文字重叠（两次写入同一位置）

---

## 解决方案

### ❌ 方案 1：强制刷新 os.Stdout（无效）

```go
func handleSubmit(s AppState, i intent.Intent) AppState {
    fmt.Println("Username:", s.Username)
    os.Stdout.Sync()  // ← 无效！仍然会被 UI 渲染覆盖
    return s
}
```

**问题**：刷新只是确保数据写入内核，但后续的 UI 渲染仍会覆盖输出。

### ✅ 方案 2：使用 Mint UI 组件显示状态（推荐）

**不要使用 fmt.Println，而是在 UI 中显示状态**

```go
// ❌ 错误方式
func handleSubmit(s AppState, i intent.Intent) AppState {
    fmt.Println("=== Form Submission ===")
    fmt.Printf("Username: %v\n", s.Username)
    return s
}

// ✅ 正确方式 1：添加状态字段
type AppState struct {
    Username     string
    Email        string
    Submitted    bool  // ← 添加提交状态
    // ... 其他字段
}

func handleSubmit(s AppState, i intent.Intent) AppState {
    s.Submitted = true
    return s  // ← UI 会自动重新渲染显示提交状态
}

// ✅ 正确方式 2：在 View 中显示状态
func renderAppView(state AppState) ui.VNode {
    var content []ui.VNode

    content = append(content, ui.VStack(
        ui.NewTextBuilder("Form Fields").Bold(true).Build(),
        // ... 表单字段
    ))

    if state.Submitted {
        content = append(content,
            ui.NewTextBuilder("──────────────────────").FgColor("gray").Build(),
            ui.NewTextBuilder("✅ Form Submitted").FgColor("green").Bold(true).Build(),
            ui.NewTextBuilder("Username: "+state.Username).FgColor("bright-black").Build(),
            ui.NewTextBuilder("Email:   "+state.Email).FgColor("bright-black").Build(),
        )
    }

    return ui.VStack(content...)
}
```

### ✅ 方案 3：使用 Mint UI 的日志系统

```go
import "github.com/wwsheng009/mint/internal/log"

func handleSubmit(s AppState, i intent.Intent) AppState {
    log.UILogger.IfEnabled().Info("Form submitted")
    log.UILogger.IfEnabled().Info("Username: %v", s.Username)
    log.UILogger.IfEnabled().Info("Email: %v", s.Email)
    return s
}
```

**优点**：
- 日志会被框架正确处理
- 不会破坏 UI 布局
- 支持日志级别和过滤

### ⚠️ 方案 4：使用 stderr 而不是 stdout（有限帮助）

```go
func handleSubmit(s AppState, i intent.Intent) AppState {
    fmt.Fprintln(os.Stderr, "=== Form Submission ===")
    fmt.Fprintln(os.Stderr, "Username:", s.Username)
    return s
}
```

**限制**：
- `stderr` 也可能被 UI 框架使用（如错误输出）
- 仍然可能与 UI 渲染竞争
- 不推荐在生产代码中使用

---

## 最佳实践

### 1. UI 状态显示 vs 日志输出

| 需求 | 推荐 | 不推荐 |
|------|------|--------|
| **表单提交状态** | 添加 state.Submitted 字段，在 UI 中显示 | fmt.Println |
| **调试信息** | 使用 log.UILogger | fmt.Println |
| **错误信息** | 添加 state.Error 字段 | fmt.Fprintln(os.Stderr) |
| **用户通知** | 使用 Toast/Modal 组件 | fmt.Println |

### 2. 表单处理示例

```go
// ✅ 推荐的表单处理流程

// 1. 提交时设置状态
func handleSubmit(s AppState, i intent.Intent) AppState {
    s.Submitted = true
    s.SubmittedAt = time.Now()
    // 注意：不使用 fmt.Println
    return s
}

// 2. 在 View 中显示提交结果
func renderAppView(state AppState) ui.VNode {
    // 表单内容
    form := ui.VStack(
        // ... 表单字段
    )

    // 提交状态显示
    var status ui.VNode
    if state.Submitted {
        status = ui.VStack(
            ui.Text(""),
            ui.NewTextBuilder("─────────────────").
                FgColor("gray").
                Build(),
            ui.NewTextBuilder("✅ Form submitted successfully").
                FgColor("green").
                Bold(true).
                Build(),
            ui.NewTextBuilder("Username: "+state.Username).
                FgColor("bright-black").
                Build(),
            ui.NewTextBuilder("Email:   "+state.Email).
                FgColor("bright-black").
                Build(),
            ui.NewTextBuilder("At:      "+state.SubmittedAt.Format(time.RFC3339)).
                FgColor("gray").
                Build(),
        )
    }

    return ui.VStack(form, status)
}
```

### 3. 调试时的临时方案

如果确实需要在终端输出，可以临时禁用 alternate screen 模式：

```go
func main() {
    rt := statemachine.NewAppRuntime(initialState, AppView, AppReducer)

    ui.RunApp(rt,
        ui.WithWidth(60),
        ui.WithHeight(30),
        ui.WithNoAlternateScreen(),  // ← 禁用 alternate screen，输出可见
        ui.WithInit(func() {
            appReducerBuilder.RegisterToGlobal(rt.GetStore())
        }),
    )
}
```

**优点**：
- `fmt.Println` 输出会显示在终端底部
- UI 不会覆盖输出
- 内容会被滚动，不会丢失

**缺点**：
- UI 功能受限（不能使用全屏）
- 仅适用于调试

---

## 完整示例对比

### ❌ 错误：在 Reducer 中使用 fmt.Println

```go
func handleSubmit(s AppState, i intent.Intent) AppState {
    fmt.Println("\n=== Form Submission ===")
    fmt.Printf("Username: %v\n", s.Username)
    fmt.Printf("Email: %v\n", s.Email)
    return s
}

// 输出结果（混乱）：
// Current State:
//   Username:
//   Email:
// === Form Submission ===
// Username:   false
// Email:      =======
// Age:        0 ======
```

### ✅ 正确：使用 UI 组件显示状态

```go
type AppState struct {
    Username   string
    Email      string
    Submitted  bool
    SubmittedAt time.Time
}

func handleSubmit(s AppState, i intent.Intent) AppState {
    s.Submitted = true
    s.SubmittedAt = time.Now()
    return s  // ← UI 会自动重新渲染
}

func renderAppView(state AppState) ui.VNode {
    form := ui.VStack(
        // ... 表单字段
    )

    var status ui.VNode
    if state.Submitted {
        status = ui.VStack(
            ui.Text(""),
            ui.NewTextBuilder("✅ Form Submitted").FgColor("green").Build(),
            ui.NewTextBuilder("Username: "+state.Username).FgColor("bright-black").Build(),
            ui.NewTextBuilder("Email:   "+state.Email).FgColor("bright-black").Build(),
        )
    }

    return ui.VStack(form, status)
}

// 输出结果（清晰）：
// ┌─────────────────────────────┐
// │   🚀 Type-Safe Form Demo    │
// │                             │
// │   Username: [user@example]   │
// │   Email:    [user@test]      │
// │                             │
// │   ────────────────────────   │
// │   ✅ Form Submitted           │
// │   Username: user@example     │
// │   Email:    user@test        │
// └─────────────────────────────┘
```

---

## 总结

### 问题本质

| 表现 | 根本原因 |
|------|---------|
| **输出截断** | UI 渲染的光标移动截断了 `fmt.Println` 输出 |
| **控制字符污染** | `fmt.Println` 输出混入了 ANSI 控制字符 |
| **内容重叠** | UI 渲染和 `fmt.Println` 同时写入 `os.Stdout` |

### 解决原则

1. **不要在 UI 运行时使用 `fmt.Println`**
2. **在 UI 组件中显示状态**（作为 AppState 的一部分）
3. **使用 Mint UI 的日志系统**（`log.UILogger`）
4. **需要终端输出时**使用 `ui.WithNoAlternateScreen()`

### 迁移检查清单

将 `fmt.Println` 迁移到 UI 组件：

- [ ] ~~`fmt.Println("State:", state)`~~ → 添加 `state.Display` 字段，在 UI 中显示
- [ ] ~~`fmt.Println("Error:", err)`~~ → 使用 `ui.NewTextBuilder().FgColor("red")` 显示错误
- [ ] ~~`fmt.Println("Success")`~~ → 添加 `state.Success` 布尔值，在 UI 中显示成功消息
- [ ] ~~`fmt.Printf("Value: %v", val)`~~ → 使用 `fmt.Sprintf(val)` 构建文本，在 UI 中显示

---

**总结**：问题的根本原因不是 buffer 无法刷新，而是 `os.Stdout` 被 UI 渲染系统占用，导致 `fmt.Println` 的输出被覆盖。正确的做法是在 UI 组件中显示状态，而不是使用 `fmt.Println`。
