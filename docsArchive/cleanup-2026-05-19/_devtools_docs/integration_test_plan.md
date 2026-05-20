# DevTools 集成测试方案

> **项目**: Mint TUI Runtime - DevTools
> **文档版本**: 1.0
> **创建日期**: 2026-01-30
> **状态**: 设计中

---

## 目录

1. [测试概述](#一测试概述)
2. [测试环境](#二测试环境)
3. [单元集成测试](#三单元集成测试)
4. [端到端测试](#四端到端测试)
5. [性能测试](#五性能测试)
6. [测试工具](#六测试工具)
7. [CI/CD 集成](#七cicd-集成)

---

## 一、测试概述

### 1.1 测试目标

| 测试类型 | 目标 | 覆盖率要求 |
|----------|------|------------|
| 单元测试 | 验证每个模块的独立功能 | > 80% |
| 集成测试 | 验证模块间交互正确性 | > 70% |
| 端到端测试 | 验证完整用户场景 | 关键路径 100% |
| 性能测试 | 确保 DevTools 开销可控 | < 5% CPU, < 10MB 内存 |
| 回归测试 | 防止已修复问题复发 | 全部 P0/P1 问题 |

### 1.2 测试架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         测试金字塔                               │
├─────────────────────────────────────────────────────────────────┤
│                        端到端测试 (10%)                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  E2E Tests: 完整用户场景 (表单、导航、交互)               │   │
│  └─────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                        集成测试 (30%)                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Runtime 集成: Action → Event → Layout → Render         │   │
│  │  DevTools 集成: CausalGraph, TimeTravel, Replay         │   │
│  │  Framework 集成: Component, Binding, Theme              │   │
│  └─────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                        单元测试 (60%)                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  EventBus, CausalGraph, FrameTimeline, LayoutCollector  │   │
│  │  AsyncCollector, MutationTap, Snapshot, Replay          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 测试分层策略

```
┌──────────────────────────────────────────────────────────────┐
│                        应用层                                 │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  framework/testing/ - 场景化 E2E 测试                   │  │
│  │  - TestScenario, TestStep, TestContext                 │  │
│  └────────────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────────────┤
│                        DevTools 层                            │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  devtools/integration_test.go - DevTools 集成测试       │  │
│  │  - Runtime集成, CausalGraph验证, TimeTravel测试         │  │
│  └────────────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────────────┤
│                        Runtime 层                             │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  runtime/*/integration_test.go - Runtime 模块集成       │  │
│  │  - Action系统, Event处理, Layout引擎, Focus管理        │  │
│  └────────────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────────────┤
│                        基础层                                 │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  *_test.go - 单元测试                                   │  │
│  │  - 数据结构, 算法, 工具函数                             │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

---

## 二、测试环境

### 2.1 目录结构

```
mint/
├── devtools/
│   ├── integration_test.go          # DevTools 集成测试主文件
│   ├── e2e_test.go                  # 端到端场景测试
│   ├── benchmark_test.go            # 性能基准测试
│   ├── testing/
│   │   ├── mock.go                  # Mock 工具
│   │   ├── fixture.go               # 测试固件
│   │   └── assertion.go             # 自定义断言
│   └── docs/
│       └── integration_test_plan.md # 本文档
│
├── runtime/
│   ├── integration_test.go          # Runtime 集成测试
│   └── testing/
│       ├── test_runtime.go          # 测试用 Runtime
│       └── test_context.go          # 测试上下文
│
├── framework/
│   └── testing/
│       ├── scenario.go              # 场景测试框架
│       └── recorder.go              # 交互记录器
│
└── tests/
    └── integration/                 # 跨模块集成测试
        ├── devtools_runtime_test.go
        └── full_stack_test.go
```

### 2.2 测试依赖

```go
// go.mod 添加测试依赖
require (
    github.com/stretchr/testify v1.8.4  // 断言库
    github.com/golang/mock v1.6.0       // Mock 生成
    gopkg.in/yaml.v3 v3.0.1             // 测试配置
)
```

### 2.3 测试配置

```yaml
# tests/config/test_config.yaml
test:
  timeout: 30s
  parallel: true
  verbose: false

devtools:
  enabled: true
  log_level: "debug"
  collect_stats: true

scenarios:
  path: "tests/scenarios"
  format: "yaml"

performance:
  enabled: true
  baseline: "tests/baseline/performance.json"
  threshold:
    cpu: 5.0      # 5%
    memory: 10485760  # 10MB
    frame_time: 500000  # 0.5ms
```

---

## 三、单元集成测试

### 3.1 Runtime 集成测试

```go
// devtools/integration_runtime_test.go
package devtools_test

import (
    "testing"
    "time"

    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/runtime"
)

// TestDevToolsRuntimeIntegration 测试 DevTools 与 Runtime 的完整集成
func TestDevToolsRuntimeIntegration(t *testing.T) {
    // 1. 创建 Runtime 实例
    rt := runtime.NewTestRuntime()
    defer rt.Close()

    // 2. 创建并启用 DevTools
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    // 3. 注册 DevTools 到 Runtime
    rt.RegisterDevTools(dt)

    // 4. 创建测试组件树
    root := createTestComponentTree()

    // 5. 初始布局收集
    layoutResult := rt.Layout(root)
    dt.CollectLayout(layoutResult)
    dt.BeginFrame()
    dt.EndFrame()

    // 6. 验证初始状态
    stats := dt.GetEventBus().GetStats()
    if stats.EventsSent.Load() == 0 {
        t.Error("Expected events to be sent")
    }

    // 7. 模拟用户交互
    simulateKeyPress(rt, 'a')
    simulateKeyPress(rt, 'b')
    simulateKeyPress(rt, 'c')

    // 8. 验证事件记录
    dt.EndFrame()
    events := dt.GetCollectedEvents()
    if len(events) < 3 {
        t.Errorf("Expected at least 3 events, got %d", len(events))
    }

    // 9. 验证因果链
    graph := dt.GetCausalGraph()
    if graph == nil {
        t.Fatal("Expected causal graph to exist")
    }

    summary := graph.GetFrameSummary()
    if summary.EventCount < 3 {
        t.Errorf("Expected at least 3 events in causal graph, got %d", summary.EventCount)
    }
}

// TestDevToolsLayoutTracking 测试布局变化追踪
func TestDevToolsLayoutTracking(t *testing.T) {
    rt := runtime.NewTestRuntime()
    defer rt.Close()

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    rt.RegisterDevTools(dt)

    // 创建可变大小的组件
    container := runtime.NewContainer()
    container.SetBounds(0, 0, 100, 100)

    child1 := runtime.NewLabel("Hello")
    child2 := runtime.NewLabel("World")
    container.Add(child1, child2)

    // 初始布局
    rt.Layout(container)
    dt.CollectLayout(rt.GetLastLayoutResult())

    // 修改组件大小
    container.SetBounds(0, 0, 200, 100)
    rt.Layout(container)
    dt.CollectLayout(rt.GetLastLayoutResult())

    // 验证布局变化被捕获
    deltas := dt.GetLayoutDeltas()
    if len(deltas) < 2 {
        t.Errorf("Expected at least 2 layout deltas, got %d", len(deltas))
    }

    // 验证变化包含容器尺寸变化
    found := false
    for _, delta := range deltas {
        for _, changed := range delta.Changed {
            if changed.Mask&devtools.ChangeRect != 0 {
                found = true
                break
            }
        }
    }
    if !found {
        t.Error("Expected to find rectangle change in layout deltas")
    }
}

// TestDevToolsCausalChain 测试因果链完整性
func TestDevToolsCausalChain(t *testing.T) {
    rt := runtime.NewTestRuntime()
    defer rt.Close()

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    rt.RegisterDevTools(dt)

    // 创建输入框和按钮
    input := runtime.NewInput()
    button := runtime.NewButton("Submit")

    container := runtime.NewVBox(
        runtime.NewLabel("Username:"),
        input,
        button,
    )

    rt.Layout(container)

    // 模拟输入序列
    dt.BeginFrame()

    // 1. 输入用户名
    rt.DispatchInput(runtime.KeyPress{Key: 'u'})
    rt.DispatchInput(runtime.KeyPress{Key: 's'})
    rt.DispatchInput(runtime.KeyPress{Key: 'e'})
    rt.DispatchInput(runtime.KeyPress{Key: 'r'})

    // 2. 按下 Tab 切换焦点
    rt.DispatchInput(runtime.KeyPress{Key: '\t'})

    // 3. 按下 Enter 提交
    rt.DispatchInput(runtime.KeyPress{Key: '\r'})

    dt.EndFrame()

    // 验证因果链
    // 事件 → 突变 → 布局 → 渲染
    graph := dt.GetCausalGraph()
    summary := graph.GetFrameSummary()

    t.Logf("Frame summary: Events=%d, Mutations=%d, Layouts=%d, Repaints=%d, Edges=%d",
        summary.EventCount, summary.MutationCount,
        summary.LayoutCount, summary.RepaintCount, summary.EdgeCount)

    // 验证边数量（因果关系）
    // 至少: 6 events → 1 focus mutation + 1 submit action
    if summary.EdgeCount < 2 {
        t.Errorf("Expected at least 2 causal edges, got %d", summary.EdgeCount)
    }

    // 验证事件顺序
    events := graph.GetAllEvents()
    if len(events) < 6 {
        t.Errorf("Expected at least 6 events, got %d", len(events))
    }

    // 验证事件类型
    hasInput := false
    hasFocusChange := false
    for _, event := range events {
        if event.Type == "input" {
            hasInput = true
        }
        if event.Type == "focus" {
            hasFocusChange = true
        }
    }

    if !hasInput {
        t.Error("Expected to find input events")
    }
    if !hasFocusChange {
        t.Error("Expected to find focus change events")
    }
}

// TestDevToolsMutationTracking 测试状态突变追踪
func TestDevToolsMutationTracking(t *testing.T) {
    rt := runtime.NewTestRuntime()
    defer rt.Close()

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    // 启用 Mutation Tap
    devtools.EnableMutationTap()
    defer devtools.DisableMutationTap()

    rt.RegisterDevTools(dt)

    // 创建可变状态的组件
    counter := runtime.NewLabel("Count: 0")
    button := runtime.NewButton("Increment")

    button.OnClick(func() {
        count := getCounterValue()
        setCounterValue(count + 1)
        counter.SetText("Count: " + string(rune(count+1)))
    })

    container := runtime.NewHBox(counter, button)
    rt.Layout(container)

    dt.BeginFrame()

    // 点击按钮
    rt.DispatchClick(button)

    dt.EndFrame()

    // 验证突变被捕获
    mutations := dt.GetCollectedMutations()
    if len(mutations) == 0 {
        t.Error("Expected mutations to be captured")
    }

    // 验证突变的因果关系
    foundClickMutation := false
    foundTextMutation := false

    for _, mut := range mutations {
        if mut.Component == "button" && mut.Kind == devtools.MutationAction {
            foundClickMutation = true
        }
        if mut.Component == "label" && mut.Field == "text" {
            foundTextMutation = true
        }
    }

    if !foundClickMutation {
        t.Error("Expected to find button click mutation")
    }
    if !foundTextMutation {
        t.Error("Expected to find label text mutation")
    }
}
```

### 3.2 Framework 集成测试

```go
// devtools/integration_framework_test.go
package devtools_test

import (
    "testing"

    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/framework"
)

// TestDevToolsFormIntegration 测试表单场景的 DevTools 集成
func TestDevToolsFormIntegration(t *testing.T) {
    // 使用 framework/testing 的测试上下文
    ctx := framework.testing.NewTestContext(framework.testing.TestOptions{
        DevTools:     true,
        RecordEvents: true,
    })
    defer ctx.Cleanup()

    // 创建表单
    form := framework.NewForm()
    form.AddTextField("username", "Username", "")
    form.AddTextField("password", "Password", "")
    form.AddTextField("email", "Email", "")
    form.AddButton("submit", "Submit")

    // 构建组件树
    app := framework.NewApp(form)
    ctx.SetRoot(app)

    // 获取 DevTools
    dt := ctx.GetDevTools()

    // 场景：填写表单
    scenarios := []framework.TestScenario{
        {
            Name: "填写用户名",
            Actions: []framework.TestAction{
                framework.KeyPress{Key: 't'},
                framework.KeyPress{Key: 'e'},
                framework.KeyPress{Key: 's'},
                framework.KeyPress{Key: 't'},
            },
            Validate: func(ctx *framework.TestContext) error {
                value := form.GetFieldValue("username")
                if value != "test" {
                    t.Errorf("Expected username='test', got '%s'", value)
                }
                return nil
            },
        },
        {
            Name: "切换到密码框",
            Actions: []framework.TestAction{
                framework.KeyPress{Key: '\t'},
            },
            Validate: func(ctx *framework.TestContext) error {
                focused := form.GetFocusedField()
                if focused != "password" {
                    t.Errorf("Expected password field focused, got '%s'", focused)
                }
                return nil
            },
        },
        {
            Name: "填写密码",
            Actions: []framework.TestAction{
                framework.KeyPress{Key: 'p'},
                framework.KeyPress{Key: 'a'},
                framework.KeyPress{Key: 's'},
                framework.KeyPress{Key: 's'},
            },
            Validate: func(ctx *framework.TestContext) error {
                value := form.GetFieldValue("password")
                if value != "pass" {
                    t.Errorf("Expected password='pass', got '%s'", value)
                }
                return nil
            },
        },
        {
            Name: "验证 DevTools 记录",
            Actions: []framework.TestAction{},
            Validate: func(ctx *framework.TestContext) error {
                // 验证事件记录
                events := dt.GetCollectedEvents()
                if len(events) < 8 { // 4 用户名 + 1 tab + 4 密码
                    t.Errorf("Expected at least 8 events, got %d", len(events))
                }

                // 验证焦点变化记录
                focusEvents := filterEventsByType(events, "focus")
                if len(focusEvents) < 1 {
                    t.Error("Expected at least 1 focus event")
                }

                // 验证因果链
                graph := dt.GetCausalGraph()
                if graph.GetFrameSummary().EventCount == 0 {
                    t.Error("Expected events in causal graph")
                }

                return nil
            },
        },
    }

    // 执行场景
    for _, scenario := range scenarios {
        t.Run(scenario.Name, func(t *testing.T) {
            if err := ctx.RunScenario(scenario); err != nil {
                t.Errorf("Scenario '%s' failed: %v", scenario.Name, err)
            }
        })
    }
}
```

---

## 四、端到端测试

### 4.1 完整用户场景测试

```go
// devtools/e2e_test.go
package devtools_test

import (
    "testing"
    "time"

    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/framework"
)

// TestE2E_LoginFlow 测试完整的登录流程
func TestE2E_LoginFlow(t *testing.T) {
    // 创建测试应用
    app := createLoginApp()

    // 创建测试上下文
    ctx := framework.NewE2EContext(app, framework.E2EOptions{
        DevTools:       true,
        RecordEvents:   true,
        CaptureScreens: true,
        Timeout:        30 * time.Second,
    })
    defer ctx.Cleanup()

    // 定义登录流程
    loginFlow := framework.E2EScenario{
        Name:        "用户登录流程",
        Description: "从启动到成功登录的完整流程",
        Steps: []framework.E2EStep{
            {
                Name: "应用启动",
                Action: func(ctx *framework.E2EContext) error {
                    // 验证应用已启动
                    if !ctx.IsRunning() {
                        return fmt.Errorf("app not running")
                    }
                    // 验证初始界面
                    if !ctx.RenderContains("Username:") {
                        return fmt.Errorf("login form not rendered")
                    }
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.RenderContains("Username:"),
                    framework.RenderContains("Password:"),
                    framework.ComponentExists("username-input"),
                    framework.ComponentExists("password-input"),
                    framework.ComponentExists("login-button"),
                },
            },
            {
                Name: "输入用户名",
                Action: func(ctx *framework.E2EContext) error {
                    ctx.Type("testuser")
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.ComponentValue("username-input", "testuser"),
                },
            },
            {
                Name: "切换到密码框",
                Action: func(ctx *framework.E2EContext) error {
                    ctx.PressKey("\t")
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.ComponentFocused("password-input"),
                },
            },
            {
                Name: "输入密码",
                Action: func(ctx *framework.E2EContext) error {
                    ctx.Type("password123")
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.ComponentValue("password-input", "password123"),
                },
            },
            {
                Name: "点击登录按钮",
                Action: func(ctx *framework.E2EContext) error {
                    ctx.PressKey("\t") // 切换到按钮
                    ctx.PressKey("\r") // 按下 Enter
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.RenderContains("Loading..."),
                },
            },
            {
                Name: "验证登录成功",
                Action: func(ctx *framework.E2EContext) error {
                    // 等待异步操作
                    time.Sleep(100 * time.Millisecond)
                    return nil
                },
                Assertions: []framework.Assertion{
                    framework.RenderContains("Welcome"),
                    framework.RenderContains("testuser"),
                },
            },
            {
                Name: "验证 DevTools 记录",
                Action: func(ctx *framework.E2EContext) error {
                    dt := ctx.GetDevTools()

                    // 验证事件记录完整
                    events := dt.GetCollectedEvents()
                    t.Logf("Total events recorded: %d", len(events))

                    // 验证因果链完整
                    graph := dt.GetCausalGraph()
                    summary := graph.GetFrameSummary()
                    t.Logf("Causal summary: Events=%d, Mutations=%d, Edges=%d",
                        summary.EventCount, summary.MutationCount, summary.EdgeCount)

                    // 验证可回放
                    replay := dt.CreateReplay()
                    if replay.GetFrameCount() == 0 {
                        return fmt.Errorf("no frames recorded for replay")
                    }

                    // 验证时间旅行
                    snapshots := dt.GetSnapshots()
                    if len(snapshots) == 0 {
                        return fmt.Errorf("no snapshots captured")
                    }

                    return nil
                },
            },
        },
    }

    // 执行场景
    if err := ctx.RunScenario(loginFlow); err != nil {
        t.Fatalf("Login flow failed: %v", err)
    }

    // 生成测试报告
    report := ctx.GenerateReport()
    t.Logf("E2E Test Report:\n%s", report)
}

// TestE2E_TimeTravel 测试时间旅行功能
func TestE2E_TimeTravel(t *testing.T) {
    app := createCounterApp()
    ctx := framework.NewE2EContext(app, framework.E2EOptions{
        DevTools: true,
    })
    defer ctx.Cleanup()

    dt := ctx.GetDevTools()

    // 执行一系列操作
    initialCount := getCounterValue(app)

    // 增加计数器 5 次
    for i := 0; i < 5; i++ {
        ctx.ClickButton("increment")
        ctx.WaitForFrame()
    }

    finalCount := getCounterValue(app)
    if finalCount != initialCount+5 {
        t.Errorf("Expected count=%d, got %d", initialCount+5, finalCount)
    }

    // 时间旅行：回到第 3 次点击后的状态
    snapshots := dt.GetSnapshots()
    if len(snapshots) < 4 {
        t.Fatalf("Expected at least 4 snapshots, got %d", len(snapshots))
    }

    // 恢复到第 3 次点击后的状态
    targetSnapshot := snapshots[3]
    if err := dt.RestoreSnapshot(targetSnapshot); err != nil {
        t.Fatalf("Failed to restore snapshot: %v", err)
    }

    // 验证状态正确恢复
    restoredCount := getCounterValue(app)
    expectedCount := initialCount + 3
    if restoredCount != expectedCount {
        t.Errorf("After time travel, expected count=%d, got %d", expectedCount, restoredCount)
    }

    // 验证可以继续操作
    ctx.ClickButton("increment")
    ctx.WaitForFrame()
    newCount := getCounterValue(app)
    if newCount != expectedCount+1 {
        t.Errorf("After restore+increment, expected count=%d, got %d", expectedCount+1, newCount)
    }
}

// TestE2E_Replay 测试确定性行回放
func TestE2E_Replay(t *testing.T) {
    app := createTodoApp()
    ctx := framework.NewE2EContext(app, framework.E2EOptions{
        DevTools:     true,
        RecordEvents: true,
    })
    defer ctx.Cleanup()

    dt := ctx.GetDevTools()

    // 原始运行：添加 3 个待办事项
    originalOutputs := []string{}
    for i := 1; i <= 3; i++ {
        todoText := fmt.Sprintf("Todo item %d", i)
        ctx.TypeInto("new-todo", todoText)
        ctx.PressKey("\r")
        ctx.WaitForFrame()

        // 记录渲染输出
        originalOutputs = append(originalOutputs, ctx.GetRenderOutput())
    }

    // 创建回放
    replay := dt.CreateReplay()

    // 重置应用
    app.Reset()

    // 执行回放
    replayOutputs := []string{}
    for i := 1; i <= 3; i++ {
        if err := replay.NextFrame(); err != nil {
            t.Fatalf("Replay frame %d failed: %v", i, err)
        }
        replayOutputs = append(replayOutputs, ctx.GetRenderOutput())
    }

    // 验证回放结果与原始运行一致
    if len(replayOutputs) != len(originalOutputs) {
        t.Fatalf("Replay output count mismatch: %d vs %d",
            len(replayOutputs), len(originalOutputs))
    }

    for i := range originalOutputs {
        if replayOutputs[i] != originalOutputs[i] {
            t.Errorf("Frame %d output mismatch:\nOriginal: %s\nReplay:   %s",
                i, originalOutputs[i], replayOutputs[i])
        }
    }

    t.Log("Deterministic replay verified successfully")
}
```

### 4.2 压力测试场景

```go
// devtools/stress_test.go
package devtools_test

import (
    "sync"
    "testing"
    "time"

    "github.com/wwsheng009/mint/devtools"
)

// TestStress_RapidEvents 测试快速连续事件
func TestStress_RapidEvents(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    // 快速发送大量事件
    eventCount := 10000

    start := time.Now()
    for i := 0; i < eventCount; i++ {
        dt.BeginFrame()
        dt.RecordEvent("keypress", "node", "bubble", map[string]interface{}{
            "key": rune('a' + i%26),
        })
        dt.EndFrame()
    }
    duration := time.Since(start)

    eventsPerSecond := float64(eventCount) / duration.Seconds()
    t.Logf("Processed %d events in %v (%.0f events/sec)",
        eventCount, duration, eventsPerSecond)

    // 验证性能
    if duration > 5*time.Second {
        t.Errorf("Event processing too slow: %v for %d events", duration, eventCount)
    }

    // 验证数据完整性
    stats := dt.GetEventBus().GetStats()
    if stats.EventsSent.Load() < uint64(eventCount) {
        t.Errorf("Expected at least %d events sent, got %d",
            eventCount, stats.EventsSent.Load())
    }
}

// TestStress_ConcurrentAccess 测试并发访问
func TestStress_ConcurrentAccess(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    const (
        goroutines   = 100
        opsPerGoroutine = 1000
    )

    var wg sync.WaitGroup
    start := time.Now()

    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < opsPerGoroutine; j++ {
                dt.BeginFrame()
                dt.RecordEvent("test", "node", "bubble", nil)
                dt.EndFrame()
            }
        }(i)
    }

    wg.Wait()
    duration := time.Since(start)

    totalOps := goroutines * opsPerGoroutine
    opsPerSecond := float64(totalOps) / duration.Seconds()

    t.Logf("Completed %d operations in %v (%.0f ops/sec) with %d goroutines",
        totalOps, duration, opsPerSecond, goroutines)

    // 验证没有死锁或崩溃
    if dt.IsEnabled() {
        t.Log("DevTools still enabled after concurrent access")
    }
}

// TestStress_MemoryGrowth 测试内存增长限制
func TestStress_MemoryGrowth(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    // 运行大量帧
    frameCount := 10000
    for i := 0; i < frameCount; i++ {
        dt.BeginFrame()
        dt.RecordEvent("test", "node", "bubble", nil)
        dt.CollectLayout(nil)
        dt.EndFrame()
    }

    // 获取内存统计
    stats := dt.GetStats()

    t.Logf("After %d frames:", frameCount)
    t.Logf("  Timeline frames: %d", stats.TimelineFrameCount)
    t.Logf("  Memory usage: %d bytes", stats.MemoryUsage)

    // 验证内存有上限（环形缓冲区应该限制增长）
    timeline := dt.GetTimeline()
    maxFrames := timeline.GetCapacity()
    currentFrames := timeline.GetFrameCount()

    if currentFrames > maxFrames {
        t.Errorf("Timeline exceeded capacity: %d > %d", currentFrames, maxFrames)
    }

    t.Logf("Timeline memory bounded: %d / %d frames", currentFrames, maxFrames)
}
```

---

## 五、性能测试

### 5.1 基准测试

```go
// devtools/benchmark_test.go
package devtools_test

import (
    "runtime"
    "testing"
    "time"

    "github.com/wwsheng009/mint/devtools"
)

// BenchmarkDevTools_Disabled 测试禁用 DevTools 的开销
func BenchmarkDevTools_Disabled(b *testing.B) {
    dt := devtools.New()
    // 不启用 DevTools

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        dt.RecordEvent("keypress", "node", "bubble", nil)
        dt.EndFrame()
    }
}

// BenchmarkDevTools_Enabled 测试启用 DevTools 的开销
func BenchmarkDevTools_Enabled(b *testing.B) {
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        dt.RecordEvent("keypress", "node", "bubble", nil)
        dt.EndFrame()
    }
}

// BenchmarkDevTools_WithLayout 测试包含布局收集的性能
func BenchmarkDevTools_WithLayout(b *testing.B) {
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    layoutResult := createMockLayoutResult(100) // 100 个组件

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        dt.CollectLayout(layoutResult)
        dt.EndFrame()
    }
}

// BenchmarkDevTools_CausalGraph 测试因果链构建性能
func BenchmarkDevTools_CausalGraph(b *testing.B) {
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    b.Run("Small", func(b *testing.B) {
        benchmarkCausalGraph(b, dt, 10, 5, 3)
    })
    b.Run("Medium", func(b *testing.B) {
        benchmarkCausalGraph(b, dt, 50, 20, 10)
    })
    b.Run("Large", func(b *testing.B) {
        benchmarkCausalGraph(b, dt, 100, 50, 20)
    })
}

func benchmarkCausalGraph(b *testing.B, dt *devtools.DevTools, events, mutations, edges int) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        for j := 0; j < events; j++ {
            dt.RecordEvent("keypress", "node", "bubble", nil)
        }
        dt.EndFrame()
    }
}

// BenchmarkDevTools_EventBus 测试 EventBus 性能
func BenchmarkDevTools_EventBus(b *testing.B) {
    bus := devtools.NewEventBus(4096)
    bus.Enable()
    defer bus.Close()

    ev := devtools.DebugEvent{Type: devtools.EventLayout}

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            bus.Emit(ev)
        }
    })
}

// BenchmarkDevTools_FrameTimeline 测试 FrameTimeline 性能
func BenchmarkDevTools_FrameTimeline(b *testing.B) {
    ft := devtools.NewFrameTimeline()
    ft.Enable()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entry := ft.BeginFrame(devtools.FrameID(i))
        ft.EndFrame()
    }
}

// BenchmarkDevTools_CausalGraphPool 测试对象池效果
func BenchmarkDevTools_CausalGraphPool(b *testing.B) {
    b.Run("WithoutPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cg := devtools.NewCausalGraph(devtools.FrameID(i))
            cg.AddEvent("test", "node", "bubble")
            // 不释放，让 GC 处理
        }
    })

    b.Run("WithPool", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cg := devtools.AcquireCausalGraph(devtools.FrameID(i))
            cg.AddEvent("test", "node", "bubble")
            devtools.ReleaseCausalGraph(cg)
        }
    })
}

// BenchmarkDevTools_MemoryAllocation 测试内存分配
func BenchmarkDevTools_MemoryAllocation(b *testing.B) {
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    b.ReportAllocs()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dt.BeginFrame()
        dt.RecordEvent("keypress", "node", "bubble", nil)
        dt.CollectLayout(createMockLayoutResult(50))
        dt.EndFrame()
    }
}

// BenchmarkDevTools_Overhead 测试 DevTools 总体开销
func BenchmarkDevTools_Overhead(b *testing.B) {
    b.Run("NoDevTools", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            simulateFrameWithoutDevTools()
        }
    })

    b.Run("WithDevTools", func(b *testing.B) {
        dt := devtools.New()
        dt.Enable()
        defer dt.Shutdown()

        for i := 0; i < b.N; i++ {
            dt.BeginFrame()
            simulateFrameWithDevTools(dt)
            dt.EndFrame()
        }
    })
}

func simulateFrameWithoutDevTools() {
    // 模拟正常的帧处理
    time.Sleep(1 * time.Microsecond)
}

func simulateFrameWithDevTools(dt *devtools.DevTools) {
    dt.RecordEvent("test", "node", "bubble", nil)
}
```

### 5.2 性能对比测试

```go
// devtools/performance_comparison_test.go
package devtools_test

import (
    "runtime"
    "testing"
    "time"

    "github.com/wwsheng009/mint/devtools"
)

// TestPerformance_Comparison 测试 DevTools 性能影响对比
func TestPerformance_Comparison(t *testing.T) {
    iterations := 10000

    // 无 DevTools
    runtime.GC()
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)

    start1 := time.Now()
    for i := 0; i < iterations; i++ {
        simulateFrameWithoutDevTools()
    }
    duration1 := time.Since(start1)

    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)

    // 有 DevTools
    dt := devtools.New()
    dt.Enable()
    defer dt.Shutdown()

    runtime.GC()
    var m3 runtime.MemStats
    runtime.ReadMemStats(&m3)

    start2 := time.Now()
    for i := 0; i < iterations; i++ {
        dt.BeginFrame()
        simulateFrameWithDevTools(dt)
        dt.EndFrame()
    }
    duration2 := time.Since(start2)

    runtime.GC()
    var m4 runtime.MemStats
    runtime.ReadMemStats(&m4)

    // 计算差异
    timeOverhead := float64(duration2-duration1) / float64(duration1) * 100
    memOverhead := float64(m4.TotalAlloc-m3.TotalAlloc) / float64(m2.TotalAlloc-m1.TotalAlloc) * 100

    t.Logf("Performance Comparison (%d iterations):", iterations)
    t.Logf("  Time:")
    t.Logf("    Without DevTools: %v (%.2f ns/op)", duration1, float64(duration1.Nanoseconds())/float64(iterations))
    t.Logf("    With DevTools:    %v (%.2f ns/op)", duration2, float64(duration2.Nanoseconds())/float64(iterations))
    t.Logf("    Overhead:         %.2f%%", timeOverhead)
    t.Logf("  Memory:")
    t.Logf("    Without DevTools: %d bytes", m2.TotalAlloc-m1.TotalAlloc)
    t.Logf("    With DevTools:    %d bytes", m4.TotalAlloc-m3.TotalAlloc)
    t.Logf("    Overhead:         %.2f%%", memOverhead)

    // 验证开销在可接受范围内
    if timeOverhead > 10.0 {
        t.Errorf("Time overhead too high: %.2f%%", timeOverhead)
    }
    if memOverhead > 20.0 {
        t.Errorf("Memory overhead too high: %.2f%%", memOverhead)
    }
}
```

---

## 六、测试工具

### 6.1 Mock 工具

```go
// devtools/testing/mock.go
package devtools_testing

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/devtools"
)

// MockRuntime 模拟 Runtime 接口
type MockRuntime struct {
    mu            sync.Mutex
    layoutResult  *devtools.LayoutResultAdapter
    events        []devtools.EventEntry
    componentTree map[string]interface{}
}

func NewMockRuntime() *MockRuntime {
    return &MockRuntime{
        events:        make([]devtools.EventEntry, 0),
        componentTree: make(map[string]interface{}),
    }
}

func (m *MockRuntime) Layout() *devtools.LayoutResultAdapter {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.layoutResult
}

func (m *MockRuntime) SetLayoutResult(lr *devtools.LayoutResultAdapter) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.layoutResult = lr
}

func (m *MockRuntime) GetEvents() []devtools.EventEntry {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.events
}

func (m *MockRuntime) AddEvent(event devtools.EventEntry) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.events = append(m.events, event)
}

// MockDevTools 模拟 DevTools 用于测试
type MockDevTools struct {
    *devtools.DevTools
    capturedEvents   []devtools.EventEntry
    capturedLayouts  []*devtools.LayoutDelta
    causalGraphs     []*devtools.CausalGraph
    snapshots        []devtools.Snapshot
}

func NewMockDevTools() *MockDevTools {
    return &MockDevTools{
        DevTools:        devtools.New(),
        capturedEvents:  make([]devtools.EventEntry, 0),
        capturedLayouts: make([]*devtools.LayoutDelta, 0),
        causalGraphs:    make([]*devtools.CausalGraph, 0),
        snapshots:       make([]devtools.Snapshot, 0),
    }
}

func (m *MockDevTools) CaptureEvent(event devtools.EventEntry) {
    m.capturedEvents = append(m.capturedEvents, event)
}

func (m *MockDevTools) CaptureLayout(delta *devtools.LayoutDelta) {
    m.capturedLayouts = append(m.capturedLayouts, delta)
}

func (m *MockDevTools) GetCapturedEvents() []devtools.EventEntry {
    return m.capturedEvents
}

func (m *MockDevTools) GetCapturedLayouts() []*devtools.LayoutDelta {
    return m.capturedLayouts
}

// AssertCausalChain 验证因果链
func (m *MockDevTools) AssertCausalChain(t testing.TB, minEvents, minMutations int) {
    t.Helper()

    if len(m.causalGraphs) == 0 {
        t.Fatal("No causal graphs recorded")
    }

    graph := m.causalGraphs[len(m.causalGraphs)-1]
    summary := graph.GetFrameSummary()

    if summary.EventCount < minEvents {
        t.Errorf("Expected at least %d events, got %d", minEvents, summary.EventCount)
    }

    if summary.MutationCount < minMutations {
        t.Errorf("Expected at least %d mutations, got %d", minMutations, summary.MutationCount)
    }
}

// AssertLayoutDelta 验证布局增量
func (m *MockDevTools) AssertLayoutDelta(t testing.TB, nodeID devtools.NodeID, expectedChanges devtools.ChangeMask) {
    t.Helper()

    for _, delta := range m.capturedLayouts {
        for _, changed := range delta.Changed {
            if changed.ID == nodeID {
                if changed.Mask&expectedChanges != 0 {
                    return // 找到匹配
                }
            }
        }
    }

    t.Errorf("Layout delta not found for node %s with changes %x", nodeID, expectedChanges)
}
```

### 6.2 测试固件

```go
// devtools/testing/fixture.go
package devtools_testing

import (
    "github.com/wwsheng009/mint/devtools"
    "github.com/wwsheng009/mint/runtime"
)

// Fixture 提供测试环境和清理
type Fixture struct {
    Runtime  *MockRuntime
    DevTools *MockDevTools
}

// Setup 创建测试环境
func Setup(t testing.TB) *Fixture {
    t.Helper()

    rt := NewMockRuntime()
    dt := NewMockDevTools()
    dt.Enable()

    return &Fixture{
        Runtime:  rt,
        DevTools: dt,
    }
}

// Teardown 清理测试环境
func (f *Fixture) Teardown(t testing.TB) {
    t.Helper()
    f.DevTools.Disable()
    f.DevTools.Shutdown()
}

// ComponentBuilder 构建测试组件
type ComponentBuilder struct {
    components []interface{}
}

func NewComponentBuilder() *ComponentBuilder {
    return &ComponentBuilder{
        components: make([]interface{}, 0),
    }
}

func (b *ComponentBuilder) AddLabel(id, text string) *ComponentBuilder {
    b.components = append(b.components, map[string]interface{}{
        "type": "label",
        "id":   id,
        "text": text,
    })
    return b
}

func (b *ComponentBuilder) AddButton(id, text string) *ComponentBuilder {
    b.components = append(b.components, map[string]interface{}{
        "type": "button",
        "id":   id,
        "text": text,
    })
    return b
}

func (b *ComponentBuilder) AddInput(id, placeholder string) *ComponentBuilder {
    b.components = append(b.components, map[string]interface{}{
        "type":        "input",
        "id":          id,
        "placeholder": placeholder,
    })
    return b
}

func (b *ComponentBuilder) Build() []interface{} {
    return b.components
}

// ScenarioBuilder 构建测试场景
type ScenarioBuilder struct {
    name        string
    description string
    steps       []ScenarioStep
}

type ScenarioStep struct {
    Name        string
    Actions     []Action
    Assertions  []Assertion
    WaitFor     time.Duration
}

type Action interface {
    Execute(*Fixture) error
}

type Assertion interface {
    Verify(*Fixture) error
}

func NewScenarioBuilder(name, description string) *ScenarioBuilder {
    return &ScenarioBuilder{
        name:        name,
        description: description,
        steps:       make([]ScenarioStep, 0),
    }
}

func (b *ScenarioBuilder) AddStep(name string, actions []Action, assertions []Assertion) *ScenarioBuilder {
    b.steps = append(b.steps, ScenarioStep{
        Name:       name,
        Actions:    actions,
        Assertions: assertions,
    })
    return b
}

func (b *ScenarioBuilder) Build() *Scenario {
    return &Scenario{
        Name:        b.name,
        Description: b.description,
        Steps:       b.steps,
    }
}

// Scenario 测试场景
type Scenario struct {
    Name        string
    Description string
    Steps       []ScenarioStep
}

func (s *Scenario) Run(t testing.TB, fixture *Fixture) error {
    t.Helper()

    for _, step := range s.Steps {
        t.Run(step.Name, func(t *testing.T) {
            // 执行操作
            for _, action := range step.Actions {
                if err := action.Execute(fixture); err != nil {
                    t.Errorf("Action failed: %v", err)
                }
            }

            // 等待
            if step.WaitFor > 0 {
                time.Sleep(step.WaitFor)
            }

            // 验证断言
            for _, assertion := range step.Assertions {
                if err := assertion.Verify(fixture); err != nil {
                    t.Errorf("Assertion failed: %v", err)
                }
            }
        })
    }

    return nil
}
```

### 6.3 自定义断言

```go
// devtools/testing/assertion.go
package devtools_testing

import (
    "testing"

    "github.com/wwsheng009/mint/devtools"
)

// Assertion 断言接口
type Assertion interface {
    Verify(*Fixture) error
}

// EventCountAssertion 断言事件数量
type EventCountAssertion struct {
    MinCount int
    MaxCount int
}

func (a *EventCountAssertion) Verify(f *Fixture) error {
    count := len(f.DevTools.GetCapturedEvents())
    if a.MinCount > 0 && count < a.MinCount {
        return fmt.Errorf("event count %d < minimum %d", count, a.MinCount)
    }
    if a.MaxCount > 0 && count > a.MaxCount {
        return fmt.Errorf("event count %d > maximum %d", count, a.MaxCount)
    }
    return nil
}

func EventCount(min, max int) Assertion {
    return &EventCountAssertion{MinCount: min, MaxCount: max}
}

// CausalChainAssertion 断言因果链
type CausalChainAssertion struct {
    MinEvents     int
    MinMutations  int
    MinEdges      int
}

func (a *CausalChainAssertion) Verify(f *Fixture) error {
    graphs := f.DevTools.causalGraphs
    if len(graphs) == 0 {
        return fmt.Errorf("no causal graphs recorded")
    }

    summary := graphs[len(graphs)-1].GetFrameSummary()

    if summary.EventCount < a.MinEvents {
        return fmt.Errorf("event count %d < minimum %d", summary.EventCount, a.MinEvents)
    }
    if summary.MutationCount < a.MinMutations {
        return fmt.Errorf("mutation count %d < minimum %d", summary.MutationCount, a.MinMutations)
    }
    if summary.EdgeCount < a.MinEdges {
        return fmt.Errorf("edge count %d < minimum %d", summary.EdgeCount, a.MinEdges)
    }

    return nil
}

func CausalChain(minEvents, minMutations, minEdges int) Assertion {
    return &CausalChainAssertion{
        MinEvents:    minEvents,
        MinMutations: minMutations,
        MinEdges:     minEdges,
    }
}

// LayoutDeltaAssertion 断言布局变化
type LayoutDeltaAssertion struct {
    NodeID         devtools.NodeID
    ExpectedChange devtools.ChangeMask
}

func (a *LayoutDeltaAssertion) Verify(f *Fixture) error {
    for _, delta := range f.DevTools.GetCapturedLayouts() {
        for _, changed := range delta.Changed {
            if changed.ID == a.NodeID {
                if changed.Mask&a.ExpectedChange != 0 {
                    return nil
                }
            }
        }
    }
    return fmt.Errorf("layout delta not found for node %s", a.NodeID)
}

func LayoutChanged(nodeID devtools.NodeID, changeMask devtools.ChangeMask) Assertion {
    return &LayoutDeltaAssertion{
        NodeID:         nodeID,
        ExpectedChange: changeMask,
    }
}

// AssertDevToolsStats DevTools 统计断言
func AssertDevToolsStats(t *testing.T, dt *devtools.DevTools, expectedEvents, expectedMutations int) {
    t.Helper()

    graph := dt.GetCausalGraph()
    if graph == nil {
        t.Fatal("No causal graph available")
    }

    summary := graph.GetFrameSummary()

    if summary.EventCount < expectedEvents {
        t.Errorf("Expected at least %d events, got %d", expectedEvents, summary.EventCount)
    }

    if summary.MutationCount < expectedMutations {
        t.Errorf("Expected at least %d mutations, got %d", expectedMutations, summary.MutationCount)
    }
}

// AssertTimelineIntegrity 时间轴完整性断言
func AssertTimelineIntegrity(t *testing.T, dt *devtools.DevTools) {
    t.Helper()

    timeline := dt.GetTimeline()
    frames := timeline.GetAllFrames()

    // 验证帧序列
    for i := 1; i < len(frames); i++ {
        prev := frames[i-1]
        curr := frames[i]

        // 验证帧 ID 单调递增
        if curr.FrameID <= prev.FrameID {
            t.Errorf("Frame ID not monotonic: %d <= %d", curr.FrameID, prev.FrameID)
        }

        // 验证时间顺序
        if curr.StartTime.Before(prev.StartTime) {
            t.Errorf("Frame time not ordered: frame %d before frame %d",
                curr.FrameID, prev.FrameID)
        }
    }
}
```

---

## 七、CI/CD 集成

### 7.1 GitHub Actions 配置

```yaml
# .github/workflows/devtools-test.yml
name: DevTools Tests

on:
  push:
    branches: [main, develop]
    paths:
      - 'devtools/**'
      - 'runtime/**'
      - 'framework/**'
  pull_request:
    paths:
      - 'devtools/**'
      - 'runtime/**'
      - 'framework/**'

jobs:
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out ./devtools/...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run integration tests
        run: |
          go test -v -race ./devtools/... -run Integration

  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run E2E tests
        run: |
          go test -v -timeout 30m ./devtools/... -run E2E

  benchmarks:
    name: Benchmarks
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem ./devtools/... | tee benchmark.txt

      - name: Store benchmark result
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: true

  race-detector:
    name: Race Detector
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run with race detector
        run: |
          go test -race -short ./devtools/...
```

### 7.2 Makefile 目标

```makefile
# devtools/Makefile

.PHONY: test test-unit test-integration test-e2e test-race benchmark coverage

# 运行所有测试
test: test-unit test-integration test-e2e

# 单元测试
test-unit:
	go test -v -race -coverprofile=coverage.out ./devtools/...
	go tool cover -html=coverage.out -o coverage.html

# 集成测试
test-integration:
	go test -v -race ./devtools/... -run Integration

# 端到端测试
test-e2e:
	go test -v -timeout 30m ./devtools/... -run E2E

# 竞态检测
test-race:
	go test -race -short ./devtools/...

# 基准测试
benchmark:
	go test -bench=. -benchmem ./devtools/... | tee benchmark.txt

# 覆盖率报告
coverage:
	go test -coverprofile=coverage.out ./devtools/...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 性能分析
profile:
	go test -cpuprofile=cpu.prof -memprofile=mem.prof ./devtools/...
	go tool pprof -text cpu.prof
	go tool pprof -text mem.prof

# 清理
clean:
	rm -f coverage.out coverage.html benchmark.txt cpu.prof mem.prof
```

---

## 八、总结

### 8.1 测试覆盖目标

| 模块 | 单元测试 | 集成测试 | E2E 测试 | 目标 |
|------|----------|----------|----------|------|
| EventBus | ✓ | ✓ | - | > 90% |
| CausalGraph | ✓ | ✓ | ✓ | > 90% |
| FrameTimeline | ✓ | ✓ | - | > 85% |
| LayoutCollector | ✓ | ✓ | ✓ | > 85% |
| AsyncCollector | ✓ | ✓ | - | > 80% |
| TimeTravel | - | ✓ | ✓ | > 80% |
| Replay | - | ✓ | ✓ | > 80% |
| Client | - | - | ✓ | > 70% |

### 8.2 性能基准

| 指标 | 目标 | 当前 | 状态 |
|------|------|------|------|
| 禁用开销 | < 10 ns/op | - | 待测 |
| 启用开销 | < 500 ns/op | - | 待测 |
| 内存增长 | < 10MB/10k frames | - | 待测 |
| CPU 占用 | < 5% | - | 待测 |
| 事件吞吐 | > 10k events/sec | - | 待测 |

### 8.3 下一步行动

1. **第一阶段**: 创建测试基础设施
   - [ ] 创建 `devtools/testing` 包
   - [ ] 实现 Mock 工具
   - [ ] 实现测试固件和断言

2. **第二阶段**: 实现集成测试
   - [ ] Runtime 集成测试
   - [ ] Framework 集成测试
   - [ ] DevTools 模块间集成测试

3. **第三阶段**: 实现 E2E 测试
   - [ ] 用户场景测试
   - [ ] 时间旅行测试
   - [ ] 回放测试

4. **第四阶段**: 性能测试
   - [ ] 基准测试
   - [ ] 性能对比
   - [ ] 压力测试

5. **第五阶段**: CI/CD 集成
   - [ ] GitHub Actions 配置
   - [ ] 性能基准跟踪
   - [ ] 覆盖率报告
