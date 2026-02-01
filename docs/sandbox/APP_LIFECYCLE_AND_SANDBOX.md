# 应用生命周期与 Sandbox 集成指南

> Mint TUI 框架应用生命周期、事件循环与 Sandbox 测试集成
>
> 版本: 1.1
> 更新日期: 2026-02-01

---

## 目录

1. [概述](#概述)
2. [应用状态机](#应用状态机)
3. [应用生命周期](#应用生命周期)
4. [事件循环机制](#事件循环机制)
5. [应用退出机制](#应用退出机制)
6. [Sandbox 集成](#sandbox-集成)
7. [测试环境 vs 生产环境](#测试环境-vs-生产环境)
8. [消息监听与处理](#消息监听与处理)
9. [最佳实践](#最佳实践)

---

## 1. 概述

Mint TUI 应用采用事件驱动架构，应用通过主事件循环持续监听和处理消息（事件）。Sandbox 提供了一个隔离的测试环境，可以模拟真实环境的事件输入。

### 核心概念

```
┌─────────────────────────────────────────────────────────────────┐
│                        应用 (App)                                │
├────────────────────────────┬────────────────────────────────────────┤
│                      主事件循环                         │
│                      (持续监听)                           │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  ┌─────────┐    ┌──────────┐    ┌────────────┐    ┌───────┐  │ │
│  │  │ 事件源  │───→│ 事件泵   │───→│  事件处理   │───→│ 渲染  │  │ │
│  │  └─────────┘    └──────────┘    └────────────┘    └───────┘  │ │
│  │       │             │                │                │        │  │
│  │       ▼             ▼                ▼                ▼        │  │
│  │  platform.Input   Pump.Events    framework.Event   paint.Buffer │ │
│  │  (生产环境)       (通道)         (事件对象)       (缓冲区)  │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  测试模式:                                                         │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  ┌─────────┐    ┌──────────┐    ┌────────────┐    ┌───────┐  │ │
│  │  │ Sandbox │───→│SandboxEvent│───│  事件处理   │───→│ 渲染  │  │ │
│  │  │         │    │   Source  │    │             │    │      │  │ │
│  │  └─────────┘    └──────────┘    └────────────┘    └───────┘  │ │
│  └────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
```

---

## 2. 应用状态机

### 2.1 AppState 定义

```go
// framework/app.go

type AppState int

const (
    StateCreated     AppState = iota  // 应用已创建，未初始化
    StateInitializing                  // 正在初始化
    StateRunning                      // 运行中（主循环活跃）
    StatePaused                       // 已暂停（主循环挂起）
    StateStopping                      // 正在停止
    StateStopped                      // 已停止
    StateError                        // 错误状态
)
```

### 2.2 状态转换图

```
                    ┌───────────────┐
                    │  StateCreated  │
                    └───────┬───────┘
                            │
                            ▼
                    ┌───────────────┐
                    │StateInitializing│
                    └───────┬───────┘
                            │
                    ┌───────▼────────┐
                    │                 │
                    ▼                 │
              ┌─────────┐         │
              │StateError│◄────┐  │
              └─────┬───┘     │  │
                    │           │  │
                    │           ▼  │
                    │    ┌──────────┐│
                    │    │StateRunning││
                    │    └─────┬────┘│
                    │          │     │
                    │    ┌──────▼─────┐│
                    │    │StatePaused ││
                    │    └──────┬─────┘│
                    │          │     │
                    │          ▼     │
                    │    ┌──────────┐│
                    │    │StateStopping││
                    │    └─────┬────┘│
                    │          │      │
                    ▼          ▼      ▼
                ┌───────────────────────┐
                │     StateStopped       │
                └───────────────────────┘
```

---

## 3. 应用生命周期

### 3.1 完整生命周期流程

```
┌───────────────────────────────────────────────────────────────────┐
│                       应用启动与运行流程                             │
└───────────────────────────────────────────────────────────────────┘

1. 创建阶段 (StateCreated)
   ├─ framework.NewApp()
   └─ 返回处于 Created 状态的 App 实例

2. 初始化阶段 (StateCreated → StateInitializing)
   ├─ 调用 Init()
   ├─ 创建事件泵 (Pump)
   │   ├─ 生产模式: platform.InputReader → PlatformEventSource
   │   └─ 测试模式: customSource → SandboxEventSource
   ├─ 设置根组件 (SetRoot)
   └─ 状态转换: StateInitializing → StateRunning

3. 运行阶段 (StateRunning)
   ├─ 启动主事件循环 (for/select)
   ├─ 监听三个通道:
   │   ├─ eventChan: 处理输入事件
   │   ├─ ticker.C: 按帧率渲染 (60fps)
   │   └─ quit: 退出信号
   └─ 循环处理直到收到退出信号

4. 停止阶段 (StateRunning → StateStopping)
   ├─ 调用 Quit() 发送退出信号
   ├─ 或收到上下文取消
   └─ 跳出主循环

5. 关闭阶段 (StateStopping → StateStopped)
   ├─ 调用 Close()
   ├─ 停止事件泵 (Pump.Stop())
   ├─ 清理资源
   └─ 应用终止
```

### 3.2 代码示例

```go
// ui/app.go - 生产环境启动

func Run(app ComponentFunc, opts ...Option) error {
    // 1. 创建应用 (StateCreated)
    fwApp := framework.NewApp()
    fwApp.Resize(80, 24)

    // 2. 初始化主题
    fwApp.InitTheme("dark")

    // 3. 设置根组件
    declarativeRoot := newDeclarativeRoot(app, fwApp)
    fwApp.SetRoot(declarativeRoot)

    // 4. 运行应用 (会调用 Init() → StateRunning → 主循环)
    return fwApp.Run()
}
```

---

## 4. 事件循环机制

### 4.1 主事件循环结构

```go
// framework/app.go (简化版)

func (a *App) Run() error {
    // 初始化
    if err := a.Init(); err != nil {
        return err
    }
    defer a.Close()

    // 创建定时器 (60fps)
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    // 获取事件通道
    eventChan := a.pump.Events()

    // 主事件循环
    for a.state == StateRunning {
        select {
        case ev := <-eventChan:
            // 处理输入事件
            if ev == nil {
                break // 通道关闭
            }
            a.handleEvent(ev)

        case <-ticker.C:
            // 按帧率渲染
            if a.dirty && a.throttler.ShouldRender() {
                a.render()
            }

        case <-a.quit:
            // 收到退出信号
            a.state = StateStopping
            return nil

        case <-a.contextMgr.Context().Done():
            // 上下文取消
            a.state = StateStopping
            return nil
        }
    }

    return nil
}
```

### 4.2 事件流向图

```
┌──────────────────────────────────────────────────────────────────┐
│                        事件流向                                     │
└───────────────────────────────────────────────────────────────────┘

输入源 → RawInput 通道 → Pump.convertLoop() → Pump.events 通道
                                                              │
                         ┌──────────────────────────────────────┘
                         ▼
                    ┌────────────────────────────────────┐
                    │          framework.App.handleEvent()    │
                    └────────────────────────────────────┘
                         │
                         ▼
              ┌────────────────────────────────────┐
              │              事件路由               │
              │  ├─ 键盘事件 → KeyMap → 快捷键处理   │
              │  ├─ 窗口调整 → 直接处理              │
              │  └─ 其他事件 → Root组件           │
              └────────────────────────────────────┘
                         │
                         ▼
                    ┌────────────────────────────────────┐
                    │           组件事件分发               │
                    │  • onClick / onChange 等回调         │
                    │  • 组件状态更新 (setState)          │
                    │  • 虚拟 DOM diff (Fiber模式)       │
                    └────────────────────────────────────┘
                         │
                         ▼
                    ┌────────────────────────────────────┐
                    │              渲染层                 │
                    │  • Reconciler 协调差异              │
                    │  • CommitRoot 提交到 Buffer       │
                    │  • Renderer 输出到终端             │
                    └────────────────────────────────────┘
```

---

## 5. 应用退出机制

### 5.1 退出触发方式

#### 方式一：用户按键退出

```go
// ui/keymap.go - 注册退出快捷键

func initKeyMap(app *framework.App) {
    // Ctrl+C 或 Ctrl+D 退出
    app.KeyMap().Register("quit", platform.KeyModCtrl, 'c',
        func(ev frameworkevent.Event) bool {
            app.Quit()
            return true
        })
}
```

#### 方式二：组件调用退出

```go
// 在组件内部触发退出

func MyLogoutButton() ui.VNode {
    return ui.ButtonBuilder("退出").
        OnClick(func() {
            // 获取应用实例并退出
            framework.GetCurrentApp().Quit()
        }).
        Build()
}
```

#### 方式三：外部信号退出

```go
// 监听系统信号

func main() {
    app := framework.NewApp()
    // ... 设置应用 ...

    // 监听中断信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        app.Quit() // 触发退出
    }()

    app.Run()
}
```

#### 方式四：测试环境自动退出

在测试环境中，应用运行在后台 goroutine 中，测试结束时 `defer testApp.Close()` 会自动清理资源，无需显式调用退出。

```go
// 测试代码示例
func TestRunTestBasic(t *testing.T) {
    testApp, err := ui.RunTest(DebugCounter,
        ui.WithWidth(40),
        ui.WithHeight(12),
    )
    if err != nil {
        t.Fatalf("RunTest failed: %v", err)
    }
    defer testApp.Close()  // ← 测试结束时自动清理

    // 测试逻辑（无需调用 Quit）
    time.Sleep(100 * time.Millisecond)
    rendered := testApp.GetRenderString()
    // ...
}  // ← defer 触发 Close()
```

**测试退出流程详解：**

```
测试函数结束
    ↓
defer testApp.Close()
    ↓
TestableApp.Close()
    ↓
framework.App.Close()
    ├─ a.state = StateStopped  ────────┐
    └─ a.pump.Stop()                   │  两个退出条件
          ├─ close(p.quit)              │
          ├─ source.Stop()              │
          └─ close(p.events)  ───┐     │
                              │     │
                         ┌────▼─────▼─────┐
                         │  主循环检查       │
                         │  for a.state == │
                         │  StateRunning   │
                         └────────┬────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │ a.state ≠ StateRunning   │ ← 条件满足
                    │ OR eventChan 关闭         │
                    └─────────────┬─────────────┘
                                  ▼
                         主循环退出
                         goroutine 结束
```

**测试模式的关键差异：**

| 特性 | 生产环境 (ui.Run) | 测试环境 (ui.RunTest) |
|------|-------------------|----------------------|
| 运行方式 | 阻塞主线程 | 后台 goroutine |
| 退出方式 | 显式调用 Quit() | defer Close() 自动清理 |
| 主循环条件 | for a.state == StateRunning | 同左，但 Close() 改变状态 |
| 事件通道关闭 | Quit() 触发 | Close() → Pump.Stop() 关闭 |
| goroutine 管理 | 无（主线程运行） | 有（后台运行，需等待退出） |

**源码实现：**

```go
// ui/test.go - RunTest 在后台运行应用
func RunTest(app ComponentFunc, opts ...Option) (*TestableApp, error) {
    fwApp := framework.NewApp()
    // ... 初始化 ...

    // 在后台运行（主循环在 goroutine 中）
    go func() {
        fwApp.Run()
    }()

    // 等待应用启动
    for i := 0; i < 100; i++ {
        if fwApp.GetState() == framework.StateRunning {
            break
        }
        time.Sleep(10 * time.Millisecond)
    }

    return &TestableApp{fwApp: fwApp, ...}, nil
}

// ui/test.go - Close 清理资源
func (ta *TestableApp) Close() error {
    return ta.fwApp.Close()  // → framework.App.Close()
}

// framework/app.go - Close 改变状态并停止事件泵
func (a *App) Close() error {
    a.state = StateStopped  // ← 关键：改变主循环条件

    // 停止事件泵（关闭 events 通道）
    if a.pump != nil {
        a.pump.Stop()
    }

    // ... 其他清理 ...
    return nil
}
```

### 5.2 退出流程

```
Quit() 被调用
        │
        ▼
┌───────────────────┐
│ quit 通道         │ ← 发送空结构体
└─────────┬─────────┘
          │
          ▼
    主循环收到退出信号
          │
          ▼
┌───────────────────┐
│ a.state =         │
│ StateStopping     │
│ return nil        │ ← 跳出主循环
└─────────┬─────────┘
          │
          ▼
      defer Close()
          │
          ▼
┌───────────────────┐
│ 停止事件泵        │
│ Pump.Stop()      │
└───────────┬───────┘
            │
            ▼
┌───────────────────┐
│ 关闭资源          │
│ • root.OnBlur()   │
│ • renderer        │
│ • theme           │
└───────────────────┘
```

### 5.3 退出相关代码

```go
// framework/app.go

func (a *App) Quit() {
    select {
    case a.quit <- struct{}{}:
        // 非阻塞发送
    default:
        // 避免重复发送
    }
}

func (a *App) Close() error {
    a.state = StateStopped

    // 让根组件失去焦点
    if a.root != nil {
        if focusable, ok := a.root.(interface{ OnBlur() }); ok {
            focusable.OnBlur()
        }
    }

    // 停止事件泵
    if a.pump != nil {
        a.pump.Stop()
    }

    return nil
}
```

---

## 6. Sandbox 集成

### 6.1 生产环境 vs 测试环境

#### 生产环境 (ui.Run)

```
┌─────────────────────────────────────────────────────────────┐
│                      生产环境 (ui.Run)                       │
└─────────────────────────────────────────────────────────────┘

应用启动流程:
1. ui.Run(app)
   ↓
2. framework.NewApp()
   ↓
3. fwApp.Init()
   ├─ NewPump(platform.InputReader)
   ├─ InputReader.Start() → 打开真实终端
   └─ Pump.Start() → 启动 convertLoop
   ↓
4. fwApp.Run()
   ├─ 主循环监听 Pump.Events()
   └─ 阻塞运行直到 Quit()

事件来源:
platform.InputReader → RawInput → Pump → framework.Event
```

#### 测试环境 (ui.RunTest)

```
┌─────────────────────────────────────────────────────────────┐
│                    测试环境 (ui.RunTest)                      │
└─────────────────────────────────────────────────────────────┘

应用启动流程:
1. ui.RunTest(app, opts...)
   ↓
2. framework.NewApp()
   ↓
3. fwApp.Run() (后台 goroutine)
   ├─ Init() 创建 Pump (默认 EventSource)
   ├─ go fwApp.Run() 后台运行
   └─ 等待 StateRunning
   ↓
4. 返回 TestableApp

事件注入方式:
testApp.InjectKey('a')
    ↓
fwApp.InjectEvent(raw)
    ↓
Pump.Inject(raw) → 直接进入事件队列
```

#### 测试环境 - Sandbox 模式 (ui.RunTestWithSandbox)

```
┌─────────────────────────────────────────────────────────────┐
│               测试环境 Sandbox 模式 (ui.RunTestWithSandbox)         │
└─────────────────────────────────────────────────────────────┘

应用启动流程:
1. ui.RunTestWithSandbox(app, opts...)
   ↓
2. 创建 MockSandbox(80, 24)
   ↓
3. 创建 SandboxEventSource(sb)
   ↓
4. framework.NewAppWithSource(source)
   ↓
5. fwApp.Run() (后台 goroutine)
   ├─ Init() 使用 customSource
   ├─ NewPumpWithSource(source)
   └─ Pump.Start() → 启动 convertLoop
   ↓
6. 返回 TestableApp (含 MockSandbox)

事件注入方式:
方式1: testApp.InjectKey('a')
    ↓
方式2: sb.InjectKey('a')
    ↓
MockSandbox.EventHandler → rawInputs 通道 → Pump
```

### 6.2 Sandbox 与应用的关系

```
┌─────────────────────────────────────────────────────────────────┐
│                   Sandbox 与应用的关系                             │
└─────────────────────────────────────────────────────────────────┘

Sandbox 不直接管理应用，而是作为事件源提供服务:

┌────────────┐        ┌──────────────┐        ┌─────────────┐
│ MockSandbox │───────→│SandboxEvent │───────→│ Pump        │
│            │ Inject │  Source      │ Start  │              │
└────────────┘        └──────────────┘        └──────┬───────┘
                                                            │
                                                            ▼
                                                   ┌────────────────────────┐
                                                   │     framework.App         │
                                                   │  (持续运行监听事件)    │
                                                   └────────────────────────┘
```

**关键点**：
1. Sandbox 独立于应用生命周期
2. 应用通过 Pump 持续监听事件
3. Sandbox 提供事件注入能力用于测试
4. 应用退出不影响 Sandbox（反之亦然）

---

## 7. 测试环境 vs 生产环境

### 7.1 对比表

| 特性 | 生产环境 (ui.Run) | 测试环境 (ui.RunTest) | Sandbox模式 (RunTestWithSandbox) |
|------|-------------------|----------------------|-------------------------------|
| **事件源** | platform.InputReader | Pump.Inject() | MockSandbox.EventHandler |
| **阻塞** | ✅ 阻塞运行 | ❌ 后台运行 | ❌ 后台运行 |
| **交互** | 真实终端 | 事件注入 | 事件注入 |
| **退出** | Quit() 或 Ctrl+C | Close() | Close() |
| **Fiber 支持** | ✅ | ✅ | ✅ |
| **事件录制** | ❌ | ❌ | ✅ (MockSandbox) |
| **调试能力** | 基础日志 | 事件注入 | 队列统计、快照 |

### 7.2 代码对比

#### 生产环境 - 阻塞运行

```go
func main() {
    // 阻塞运行，直到用户退出
    err := ui.Run(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
        ui.WithTitle("My TUI App"),
    )

    if err != nil {
        log.Fatal(err)
    }

    // Run 返回后，应用已退出
    log.Println("App exited cleanly")
}
```

#### 测试环境 - 后台运行

```go
func TestMyApp(t *testing.T) {
    // 后台运行，可以注入事件
    testApp, err := ui.RunTest(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 应用在后台运行
    time.Sleep(100 * time.Millisecond)

    // 注入事件测试
    testApp.InjectKey('a')
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 验证结果
    rendered := testApp.GetRenderString()
    if !strings.Contains(rendered, "Expected") {
        t.Error("Test failed")
    }
}
```

#### 测试环境 - Sandbox 高级功能

```go
func TestWithRecording(t *testing.T) {
    testApp, err := ui.RunTestWithSandbox(MyApp)
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 获取 Sandbox 用于高级功能
    sb := testApp.GetSandbox()

    // 获取队列统计
    stats := sb.QueueStats()
    t.Logf("Queue length: %d, Memory: %d bytes",
        stats.Length, stats.MemoryUsed)

    // 注入事件
    sb.InjectSpecialKey(platform.KeyTab)
    sb.InjectKey('a')
}
```

---

## 8. 消息监听与处理

### 8.1 持续监听机制

应用通过 `select` 语句实现持续监听多个通道：

```go
for a.state == StateRunning {
    select {
    case ev := <-eventChan:
        // 监听事件通道
        a.handleEvent(ev)

    case <-ticker.C:
        // 监听定时器 (60fps)
        a.handleTick()

    case <-a.quit:
        // 监听退出信号
        return

    case <-a.contextMgr.Context().Done():
        // 监听上下文取消
        return
    }
}
```

### 8.2 事件处理优先级

```
┌─────────────────────────────────────────────────────────────────┐
│                     事件处理优先级                                 │
└─────────────────────────────────────────────────────────────────┘

优先级 1: 退出信号 (quit, context.Done())
    → 立即退出主循环

优先级 2: 定时器 (ticker.C)
    → 确保 UI 按帧率刷新
    → 即使没有输入也要渲染动画

优先级 3: 输入事件 (eventChan)
    → 处理用户交互
    → 触发组件回调
```

### 8.3 组件事件处理流程

```
用户输入 → RawInput
    ↓
Pump.convertLoop() → Event
    ↓
App.handleEvent()
    ↓
┌─────────────────────────────────────┐
│ 1. 事件过滤器 (eventFilter)          │
│    → 可以拦截特定事件                │
└─────────────────────────────────────┘
    ↓ (通过)
┌─────────────────────────────────────┐
│ 2. 快捷键处理 (KeyMap)              │
│    → Ctrl+C 退出                      │
│    → Ctrl+S 保存                      │
└─────────────────────────────────────┘
    ↓ (未处理)
┌─────────────────────────────────────┐
│ 3. 窗口调整 (直接处理)              │
│    → 更新缓冲区大小                  │
└─────────────────────────────────────┘
    ↓ (其他事件)
┌─────────────────────────────────────┐
│ 4. 组件事件分发                    │
│    ├─ 鼠标事件 → Root.HandleMouseEvent │
│    ├─ 键盘事件 → Root.HandleKeyEvent  │
│    └─ 自定义事件 → 分发到目标组件    │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│ 5. 组件回调                          │
│    ├─ onClick()                       │
│    ├─ onChange()                      │
│    ├─ onKeyPress()                    │
│    └─ 其他钩子函数                    │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│ 6. 状态更新与渲染                    │
│    ├─ setState() → dirty=true         │
│    ├─ reconciler 协调 (Fiber模式)     │
│    └─ render() → 输出到终端           │
└─────────────────────────────────────┘
```

---

## 9. 最佳实践

### 9.1 测试交互式应用

#### 基本模式测试

```go
func TestInteractiveApp(t *testing.T) {
    testApp, err := ui.RunTest(MyInteractiveApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 等待应用初始化完成
    time.Sleep(100 * time.Millisecond)

    // 模拟用户操作序列
    testApp.InjectKey('u')      // 用户名
    testApp.InjectKey('s')      // 切换到密码框
    testApp.InjectKey('p')      // 输入密码
    testApp.InjectSpecialKey(platform.KeyEnter) // 提交

    // 等待处理
    time.Sleep(100 * time.Millisecond)

    // 验证结果
    if err := testApp.AssertRender("Login Success"); err != nil {
        t.Error(err)
    }
}
```

#### 使用 Sandbox 进行高级测试

```go
func TestWithEventRecording(t *testing.T) {
    testApp, err := ui.RunTestWithSandbox(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    sb := testApp.GetSandbox()

    // 记录初始队列状态
    stats1 := sb.QueueStats()

    // 执行一系列操作
    testApp.InjectKey('a')
    testApp.InjectSpecialKey(platform.KeyTab)
    testApp.InjectKey('b')
    testApp.InjectSpecialKey(platform.KeyEnter)

    // 等待处理
    time.Sleep(200 * time.Millisecond)

    // 记录最终队列状态
    stats2 := sb.QueueStats()

    // 验证事件处理
    if stats2.Length >= stats1.Length {
        t.Logf("Events processed: %d",
            stats2.Length-stats1.Length)
    }

    // 验证渲染结果
    if err := testApp.AssertRender("Expected Result"); err != nil {
        t.Error(err)
    }
}
```

### 9.2 测试退出最佳实践

#### 无需显式退出

在测试中，**不需要显式发送退出消息**。`defer testApp.Close()` 会自动处理所有清理工作。

```go
// ✅ 推荐：让 defer 自动处理退出
func TestCounter(t *testing.T) {
    testApp, err := ui.RunTest(CounterApp)
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()  // ← 自动清理，无需显式退出

    // 测试逻辑...
    testApp.InjectSpecialKey(platform.KeyTab)
    testApp.InjectSpecialKey(platform.KeyEnter)
    time.Sleep(100 * time.Millisecond)

    // 验证结果
    rendered := testApp.GetRenderString()
    // ...
}  // ← defer 触发，主循环自动退出
```

#### 何时需要显式退出

只有在测试"退出功能本身"时，才需要显式发送退出消息：

```go
// 测试退出功能时需要显式退出
func TestQuitFunctionality(t *testing.T) {
    testApp, err := ui.RunTest(MyApp)
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()  // 仍然保留 defer 作为保险

    fwApp := testApp.GetFrameworkApp()

    // 验证应用正在运行
    if fwApp.GetState() != framework.StateRunning {
        t.Error("App not running")
    }

    // 模拟用户按 Ctrl+C 退出
    testApp.InjectKeyWithMod('c', platform.KeyModCtrl)
    time.Sleep(100 * time.Millisecond)

    // 验证应用已停止
    if fwApp.GetState() == framework.StateRunning {
        t.Error("App should have stopped after Ctrl+C")
    }

    // Close 会再次调用（幂等，不会出错）
    testApp.Close()
}
```

#### 测试退出状态转换

```go
// 验证退出过程中的状态变化
func TestExitStateTransition(t *testing.T) {
    testApp, _ := ui.RunTest(MyApp)
    defer testApp.Close()

    fwApp := testApp.GetFrameworkApp()

    // 初始状态：Running
    t.Logf("Initial state: %v", fwApp.GetState())
    if fwApp.GetState() != framework.StateRunning {
        t.Errorf("Expected StateRunning, got %v", fwApp.GetState())
    }

    // 调用 Quit（模拟退出键）
    fwApp.Quit()
    time.Sleep(50 * time.Millisecond)

    // 状态应该不再是 Running
    finalState := fwApp.GetState()
    t.Logf("After Quit: %v", finalState)

    if finalState == framework.StateRunning {
        t.Error("App should not be running after Quit")
    }
}
```

#### 并发测试注意事项

```go
// 并发测试时确保每个测试独立清理
func TestMultipleApps(t *testing.T) {
    // 创建多个独立的应用
    apps := make([]*ui.TestableApp, 3)
    for i := 0; i < 3; i++ {
        app, err := ui.RunTest(MyApp)
        if err != nil {
            t.Fatalf("Failed to create app %d: %v", i, err)
        }
        apps[i] = app
    }

    // 确保所有应用都在运行
    for i, app := range apps {
        fwApp := app.GetFrameworkApp()
        if fwApp.GetState() != framework.StateRunning {
            t.Errorf("App %d not running", i)
        }
    }

    // 按逆序关闭（避免依赖问题）
    for i := len(apps) - 1; i >= 0; i-- {
        apps[i].Close()
    }
}
```

### 9.3 测试状态持久化

```go
func TestStatePersistence(t *testing.T) {
    testApp, err := ui.RunTest(CounterApp)
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 初始渲染
    time.Sleep(50 * time.Millisecond)
    if err := testApp.AssertRender("Count: 0"); err != nil {
        t.Error(err)
    }

    // 点击增加按钮
    testApp.InjectSpecialKey(platform.KeyTab)  // 切换到按钮
    testApp.InjectSpecialKey(platform.KeyEnter) // 点击
    time.Sleep(100 * time.Millisecond)

    // 验证状态改变
    if err := testApp.AssertRender("Count: 1"); err != nil {
        t.Error(err)
    }

    // 再次渲染应该保持状态
    time.Sleep(50 * time.Millisecond)
    if err := testApp.AssertRender("Count: 1"); err != nil {
        t.Error(err)
    }
}
```

---

## 10. 附录

### A. 事件类型映射

```
platform.RawInput          Pump.convertLoop()      framework.Event
├───────────────────┬───────────────────┬──────────────────┐
│ InputKeyPress      │                   │ KeyPressEvent    │
│   ├─ Key (rune)    │ → convertKeyEvent  │ ├─ Key           │
│   └─ SpecialKey    │                   │ └─ Modifiers    │
│ InputResize        │ → convertResizeEvent │ ResizeEvent      │
│ InputMouse         │ → convertMouseEvent  │ MouseEvent       │
│   ├─ MouseX/Y     │                   │ ├─ X/Y          │
│   ├─ MouseButton │                   │ ├─ Button        │
│   └─ MouseAction  │                   │ └─ Action        │
└───────────────────┴───────────────────┴──────────────────┘
```

### B. 关键常量

```go
// 应用状态
const (
    StateCreated     framework.StateCreated     = 0
    StateRunning     framework.StateRunning     = 2
    StateStopped     framework.StateStopped     = 5
)

// 特殊按键
const (
    platform.KeyEnter    // 回车键
    platform.KeyTab      // Tab键
    platform.KeyEscape   // Esc键
    platform.KeyBackspace // 退格键
)

// 修饰符
const (
    platform.KeyModCtrl   // Ctrl
    platform.KeyModShift  // Shift
    platform.KeyModAlt    // Alt
)

// 鼠标按钮
const (
    platform.MouseLeft    // 左键
    platform.MouseRight   // 右键
    platform.MouseMiddle  // 中键
)
```

### C. 相关文档

- [ENGINE_SANDBOX_INTEGRATION.md](./ENGINE_SANDBOX_INTEGRATION.md) - Engine 与 Sandbox 集成架构
- [SANDBOX_DEBUG_GUIDE.md](./SANDBOX_DEBUG_GUIDE.md) - Sandbox 调试技巧
- [API_REFERENCE.md](./API_REFERENCE.md) - API 完整参考
- [USER_GUIDE.md](./USER_GUIDE.md) - 使用手册

---

**文档结束**

如有疑问或建议，请提交 Issue 或 Pull Request。
