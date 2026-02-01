# 如何使用 Sandbox 运行你的应用

> Mint TUI 应用 Sandbox 集成快速指南
>
> 版本: 1.1 - 支持 SandboxEventSource 集成

## 概述

Sandbox 是 Mint TUI 框架的测试环境，支持三种运行模式：

1. **Mock 沙箱** - 自动化测试，无需终端
2. **真实沙箱** - 交互式运行，真实环境
3. **回放沙箱** - 事件回放，调试分析

---

## 第一部分：Mock 沙箱测试（推荐用于单元测试）

### 方式一：使用 RunTest (推荐 - 新版 API)

新版 API 提供完整的框架功能支持，包括 Fiber 模式、完整事件处理、渲染缓冲区等。

#### 1. 导入必要的包

```go
import (
    "testing"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/platform"
)
```

#### 2. 创建测试函数

```go
func TestMyApp(t *testing.T) {
    // 创建测试应用（完整框架支持）
    testApp, err := ui.RunTest(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()  // 重要：确保清理

    // 注入事件
    testApp.InjectSpecialKey(platform.KeyTab)
    testApp.InjectSpecialKey(platform.KeyEnter)

    // 获取渲染结果
    rendered := testApp.GetRenderString()

    // 断言
    if err := testApp.AssertRender("Expected Text"); err != nil {
        t.Error(err)
    }
}
```

### 方式二：使用 RunTestWithSandbox (高级功能)

支持事件录制/回放等 MockSandbox 高级功能。

```go
func TestMyApp_WithSandbox(t *testing.T) {
    // 使用 MockSandbox 作为事件源
    testApp, err := ui.RunTestWithSandbox(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 获取 MockSandbox
    sb := testApp.GetSandbox()

    // 通过 Sandbox 注入事件
    sb.InjectSpecialKey(platform.KeyTab)
    sb.InjectSpecialKey(platform.KeyEnter)

    // 获取队列统计
    stats := sb.QueueStats()
    t.Logf("Queue length: %d", stats.Length)

    // 直接注入也支持
    testApp.InjectKey('a')
}
```

### 方式三：使用 TestRun (Deprecated - 仅用于简单组件)

```go
// Deprecated: 推荐使用 ui.RunTest() 获取完整的框架功能支持
func TestMyApp_Simple(t *testing.T) {
    testApp, err := ui.TestRun(MyApp, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    helper := testApp.Helper()
    result := helper.
        Process().
        AssertRender("Expected Text").
        Result()

    if !result.OK() {
        t.Error(result.Error())
    }
}
```

#### 3. 运行测试

```bash
go test -v
```

### 完整示例：测试计数器应用

```go
package main_test

import (
    "testing"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/platform"
)

func TestCounterIncrement(t *testing.T) {
    testApp, err := ui.RunTest(Counter,
        ui.WithWidth(40),
        ui.WithHeight(18),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 验证初始状态
    if err := testApp.AssertRender("Count: 0"); err != nil {
        t.Error(err)
    }

    // 点击 "+" 按钮
    testApp.InjectSpecialKey(platform.KeyTab)  // 切换焦点
    testApp.InjectSpecialKey(platform.KeyEnter) // 触发点击

    // 验证状态更新
    if err := testApp.AssertRender("Count: 1"); err != nil {
        t.Error(err)
    }
}
```

### 常用事件注入方法

```go
// 字符键
testApp.InjectKey('a')

// 特殊按键
testApp.InjectSpecialKey(platform.KeyEnter)
testApp.InjectSpecialKey(platform.KeyTab)
testApp.InjectSpecialKey(platform.KeyEscape)
testApp.InjectSpecialKey(platform.KeyBackspace)

// 带修饰符
raw := platform.RawInput{
    Type:  platform.InputKeyPress,
    Key:   's',
    Modifiers: platform.ModCtrl,
}
testApp.GetFrameworkApp().InjectEvent(raw)

// 字符串
testApp.InjectString("hello world")

// 鼠标事件
testApp.InjectMouse(10, 5, platform.MouseLeft, platform.MousePress)
```

---

## 第二部分：真实沙箱运行（交互式调试）

### 基本步骤

#### 1. 导入必要的包

```go
import (
    "log"
    "github.com/wwsheng009/mint/sandbox/real"
)
```

#### 2. 创建真实沙箱

```go
func main() {
    // 创建真实沙箱
    sb, err := real.New(80, 24)
    if err != nil {
        log.Fatal(err)
    }
    defer sb.Close()

    // 初始化并运行
    sb.Initialize(nil)
    sb.Run()

    // 在这里运行你的应用
    // 注意：需要将事件从沙箱转发到你的应用
}
```

### 完整示例：带事件处理的真实沙箱

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/real"
    "github.com/wwsheng009/mint/ui"
)

func main() {
    // 创建真实沙箱
    sb, err := real.New(80, 24)
    if err != nil {
        log.Fatal(err)
    }
    defer sb.Close()

    // 设置事件处理器
    sb.SetEventHandler(func(event sandbox.RawInput) error {
        // 将事件发送到 UI 引擎
        // 在实际应用中，这里会调用 ui.DispatchEvent(event)
        log.Printf("Event: %+v", event)
        return nil
    })

    // 初始化并运行
    sb.Initialize(nil)
    sb.Run()

    // 运行 UI 应用（实际实现）
    // ui.Run(MyApp, ui.WithInputSource(sb.Events()))

    // 等待中断信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    log.Println("Running... Press Ctrl+C to exit")
    <-sigCh

    // 获取录制的事件
    events := sb.RecordedEvents()
    log.Printf("Recorded %d events", len(events))

    // 可以保存事件供回放使用
    // saveEvents(events, "recording.json")
}
```

### 使用快照记录状态

```go
func main() {
    sb, err := real.New(80, 24)
    if err != nil {
        log.Fatal(err)
    }
    defer sb.Close()

    sb.Initialize(nil)
    sb.Run()

    // 创建初始快照
    snap1, _ := sb.Snapshot(sandbox.SnapshotStandard, "initial")

    // ... 用户交互 ...

    // 创建交互后快照
    snap2, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-interaction")

    // 列出所有快照
    snapshots := sb.ListSnapshots()
    for _, meta := range snapshots {
        log.Printf("Snapshot: %s, Tags: %v", meta.ID, meta.Tags)
    }
}
```

---

## 第三部分：回放沙箱（调试和问题复现）

### 基本步骤

#### 1. 录制事件

```go
// 从真实沙箱获取录制的事件
realSB, _ := real.New(80, 24)
realSB.Initialize(nil)
realSB.Run()

// ... 用户交互 ...

events := realSB.RecordedEvents()
```

#### 2. 创建回放沙箱

```go
import "github.com/wwsheng009/mint/sandbox/replay"

replaySB := replay.New(events, 80, 24)
replaySB.Initialize(nil)
replaySB.Run()
```

#### 3. 控制回放

```go
player := replaySB.Player()

// 播放控制
player.Play()    // 开始播放
player.Pause()   // 暂停
player.Stop()    // 停止

// 导航
player.Next()      // 下一个事件
player.Previous()  // 上一个事件
player.Seek(10)    // 跳转到索引 10

// 速度控制
player.SetSpeed(2.0)  // 2 倍速
player.SetSpeed(0.5)  // 0.5 倍速
```

### 完整示例：事件回放

```go
package main

import (
    "log"

    "github.com/wwsheng009/mint/sandbox/replay"
)

func main() {
    // 假设我们已经从真实沙箱获得了事件
    events := loadEventsFromFile("recording.json")

    // 创建回放沙箱
    replaySB := replay.New(events, 80, 24)
    replaySB.Initialize(nil)
    replaySB.Run()

    player := replaySB.Player()

    // 单步回放
    for player.HasNext() {
        event, err := player.Next()
        if err != nil {
            log.Printf("Error: %v", err)
            break
        }

        log.Printf("Event %d: %+v", player.Index(), event)

        // 在实际应用中，这里会将事件发送到 UI 引擎
        // ui.DispatchEvent(event)

        // 可以在这里设置断点进行调试
        if shouldBreakAt(event) {
            break
        }
    }
}

func shouldBreakAt(event Event) bool {
    // 自定义断点逻辑
    return event.Type == InputKeyPress && event.Key == 'q'
}
```

---

## 第四部分：完整工作流示例

### 示例 1：完整的测试工作流

```go
package main_test

import (
    "fmt"
    "testing"

    "github.com/wwsheng009/mint/ui"
)

// TestCompleteWorkflow 完整的测试工作流
func TestCompleteWorkflow(t *testing.T) {
    testApp, err := ui.TestRun(MyApp, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    sb := testApp.Sandbox()
    helper := testApp.Helper()

    // 1. 测试初始化
    t.Run("Initialization", func(t *testing.T) {
        result := helper.
            Process().
            AssertRender("Welcome").
            Result()

        if !result.OK() {
            t.Error(result.Error())
        }
    })

    // 2. 测试用户登录
    t.Run("Login", func(t *testing.T) {
        result := helper.
            Tab().
            Type("user@example.com").
            Tab().
            Type("password").
            Enter().
            Process().
            AssertRender("Dashboard").
            Result()

        if !result.OK() {
            t.Error(result.Error())
        }
    })

    // 3. 使用快照保存登录状态
    t.Run("SaveLoginState", func(t *testing.T) {
        snap, err := sb.Snapshot(sandbox.SnapshotStandard, "logged-in")
        if err != nil {
            t.Fatal(err)
        }

        // 执行一些操作
        helper.Tab().Tab().Enter().Process()

        // 恢复到登录状态
        sb.Restore(snap)

        // 验证已恢复
        err = sb.AssertRender("Dashboard")
        if err != nil {
            t.Error(err)
        }
    })

    // 4. 测试多个场景
    scenarios := []struct {
        name     string
		actions  func(*mock.TestHelper) *mock.TestHelper
		expected string
	}{
		{
			name: "Create Item",
			actions: func(h *mock.TestHelper) *mock.TestHelper {
				return h.Tab().Tab().Type("New Item").Enter().Process()
			},
			expected: "Item created",
		},
		{
			name: "Delete Item",
			actions: func(h *mock.TestHelper) *mock.TestHelper {
				return h.Tab().Tab().Tab().Enter().Process()
			},
			expected: "Item deleted",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// 恢复到登录状态
			sb.Restore(snap)

			// 执行场景
			result := sc.actions(helper).AssertRender(sc.expected).Result()
			if !result.OK() {
				t.Errorf("%s failed: %v", sc.name, result.Error())
			}
		})
	}
}
```

---

## 常见使用模式

### 模式 1：测试单一功能

```go
func TestFeatureX(t *testing.T) {
    testApp, _ := ui.TestRun(MyApp)
    defer testApp.Close()

    result := testApp.Helper().
        // 导航到功能 X
        Tab().Tab().Tab().
        // 执行操作
        Enter().
        Process().
        // 验证结果
        AssertRender("Expected Result").
        Result()

    if !result.OK() {
        t.Error(result.Error())
    }
}
```

### 模式 2：测试状态转换

```go
func TestStateTransition(t *testing.T) {
    testApp, _ := ui.TestRun(MyApp)
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 保存初始状态
    initial, _ := sb.Snapshot(sandbox.SnapshotStandard, "initial")

    // 执行操作
    testApp.Helper().
        Tab().Enter().
        Process().
        Result()

    // 保存新状态
    newState, _ := sb.Snapshot(sandbox.SnapshotStandard, "new-state")

    // 验证可以回到初始状态
    sb.Restore(initial)
    if err := sb.AssertRender("Initial State"); err != nil {
        t.Error(err)
    }

    // 验证可以回到新状态
    sb.Restore(newState)
    if err := sb.AssertRender("New State"); err != nil {
        t.Error(err)
    }
}
```

### 模式 3：测试复杂表单

```go
func TestComplexForm(t *testing.T) {
    testApp, _ := ui.TestRun(MyApp)
    defer testApp.Close()

    helper := testApp.Helper()

    result := helper.
        // 填写表单字段
        Tab().Type("John Doe").
        Tab().Type("john@example.com").
        Tab().Type("123-456-7890").
        Tab().Tab().Type("New York").
        Tab().Tab().Type("10001").
        // 提交表单
        Tab().Tab().Tab().Enter().
        Process().
        // 验证结果
        AssertRender("Success").
        AssertNotRender("Error").
        Result()

    if !result.OK() {
        t.Error(result.Error())
    }
}
```

### 模式 4：性能测试

```go
func TestPerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }

    testApp, _ := ui.TestRun(MyApp)
    defer testApp.Close()

    helper := testApp.Helper()

    start := time.Now()

    // 执行大量操作
    for i := 0; i < 100; i++ {
        result := helper.
            Tab().Enter().
            Process().
            Result()

        if !result.OK() {
            t.Errorf("Iteration %d failed", i)
            break
        }
    }

    elapsed := time.Since(start)
    t.Logf("100 iterations took: %v", elapsed)
}
```

---

## 集成到现有项目

### 步骤 1：更新应用结构

确保你的应用是一个返回 `ui.VNode` 的函数：

```go
func MyApp() ui.VNode {
    // ... 应用逻辑 ...
}
```

### 步骤 2：添加测试文件

创建 `*_test.go` 文件：

```go
package main_test

import (
    "testing"
    "github.com/wwsheng009/mint/ui"
)

func TestMyApp(t *testing.T) {
    testApp, err := ui.TestRun(MyApp, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // ... 测试逻辑 ...
}
```

### 步骤 3：运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run TestMyApp

# 运行所有测试并显示覆盖率
go test -cover ./...
```

---

## 最佳实践

1. **总是清理资源**
```go
testApp, _ := ui.TestRun(MyApp)
defer testApp.Close()
```

2. **使用有意义的快照标签**
```go
sb.Snapshot(sandbox.SnapshotStandard, "after-login", "user-authenticated")
```

3. **验证每个断言**
```go
result := helper.Process().AssertRender("expected").Result()
if !result.OK() {
    t.Error(result.Error())
}
```

4. **使用快照避免重复初始化**
```go
initial, _ := sb.Snapshot("initial")
// ... 测试 ...
sb.Restore(initial)  // 回到初始状态
```

5. **组合使用 Mock 和 Real 沙箱**
- Mock 用于自动化测试
- Real 用于手动测试和录制
- Replay 用于调试和问题复现

---

## 故障排查

### 问题：测试失败，但手动运行正常

**解决：** 检查事件处理时机
```go
helper.Tab().Enter().Process()  // 确保调用 Process()
```

### 问题：断言失败，找不到期望的文本

**解决：** 使用 RenderString() 查看实际输出
```go
output := sb.RenderString()
t.Logf("Actual output:\n%s", output)
```

### 问题：事件队列满了

**解决：** 增加队列容量
```go
config := sandbox.DefaultConfig()
config.Event.QueueMaxSize = 50000
testApp, _ := ui.TestRunWithConfig(MyApp, config)
```

### 问题：快照恢复失败

**解决：** 检查快照级别
- Minimal: 只有 Buffer
- Standard: Buffer + Events
- Full: Buffer + Events + State

---

## 总结

| 场景 | 推荐沙箱 | 命令 |
|------|---------|------|
| 单元测试 | Mock | `go test -v` |
| 手动测试 | Real | `go run main.go` |
| 问题复现 | Replay | `go run replay.go` |
| CI/CD | Mock | `go test -race -cover` |

更多详细示例，请参考：
- `examples/sandbox_demo/` - 完整示例代码
- `docs/sandbox/USER_GUIDE.md` - 详细使用手册
- `docs/sandbox/API_REFERENCE.md` - API 参考手册
