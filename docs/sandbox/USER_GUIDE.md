# Sandbox 使用手册

> Mint TUI 框架沙箱测试环境使用指南
>
> 版本: 1.0
> 更新日期: 2026-02-01

---

## 目录

1. [快速开始](#快速开始)
2. [沙箱类型](#沙箱类型)
3. [Mock 沙箱测试](#mock-沙箱测试)
4. [真实终端测试](#真实终端测试)
5. [事件录制与回放](#事件录制与回放)
6. [快照功能](#快照功能)
7. [链式测试 API](#链式测试-api)
8. [UI 层集成](#ui-层集成)
9. [最佳实践](#最佳实践)
10. [常见问题](#常见问题)

---

## 快速开始

### 什么是 Sandbox？

Sandbox 是 Mint TUI 框架的测试环境，提供三种运行模式：

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| **Mock 沙箱** | 模拟测试环境 | 单元测试、组件测试、CI/CD |
| **真实沙箱** | 真实终端环境 | 手动测试、演示、交互调试 |
| **回放沙箱** | 事件回放环境 | 问题复现、调试分析 |

### 最小化示例

```go
package main_test

import (
    "testing"
    "github.com/wwsheng009/mint/sandbox/mock"
)

func TestBasicUsage(t *testing.T) {
    // 创建 Mock 沙箱
    sb := mock.New(80, 24)
    defer sb.Close()

    // 初始化并运行
    sb.Initialize(nil)
    sb.Run()

    // 注入事件
    sb.InjectString("hello")
    sb.ProcessEvents()

    // 检查结果
    if err := sb.AssertRender("hello"); err != nil {
        t.Error(err)
    }
}
```

---

## 沙箱类型

### 1. Mock 沙箱

**最适合：** 单元测试、自动化测试

**特点：**
- ✅ 无需终端交互
- ✅ 完全可控的事件注入
- ✅ 适合 CI/CD 环境
- ✅ 快速、稳定

**使用场景：**
```go
// 创建 Mock 沙箱
sb := mock.New(80, 24)
```

### 2. 真实沙箱

**最适合：** 手动测试、演示、交互式调试

**特点：**
- ✅ 真实终端输入
- ✅ 自动录制事件
- ✅ 完整用户体验
- ⚠️ 需要终端环境

**使用场景：**
```go
// 创建真实沙箱
sb, err := real.New(80, 24)
if err != nil {
    log.Fatal(err)
}
defer sb.Close()

// 运行并等待用户输入
sb.Run()
// ... 用户交互 ...
```

### 3. 回放沙箱

**最适合：** 问题复现、调试分析

**特点：**
- ✅ 精确复现已录制的会话
- ✅ 支持单步调试
- ✅ 可变速回放
- ✅ 前进后退导航

**使用场景：**
```go
// 从真实沙箱获取录制的事件
realSB, _ := real.New(80, 24)
realSB.Run()
// ... 用户交互 ...
events := realSB.RecordedEvents()

// 创建回放沙箱
replaySB := replay.New(events, 80, 24)
replaySB.Run()

// 逐步回放
for replaySB.Player().HasNext() {
    event, _ := replaySB.Step()
    fmt.Printf("Event: %+v\n", event)
}
```

---

## Mock 沙箱测试

### 基本使用流程

```go
// 1. 创建沙箱
sb := mock.New(80, 24)

// 2. 初始化
sb.Initialize(nil)

// 3. 启动
sb.Run()

// 4. 注入事件
sb.InjectString("username")
sb.ProcessEvents()

// 5. 检查结果
sb.AssertRender("username")

// 6. 清理
sb.Close()
```

### 设置事件处理器

```go
// 定义事件处理函数
handler := func(event platform.RawInput) error {
    fmt.Printf("Received event: %v\n", event)
    return nil
}

// 设置处理器
sb.SetEventHandler(handler)

// 现在注入的事件会调用处理器
sb.InjectKey('a')
sb.ProcessEvents()
```

### 注入不同类型的事件

```go
// 按键事件
sb.InjectKey('a')
sb.InjectKey('b')
sb.InjectKey('c')

// 特殊按键
sb.InjectSpecialKey(platform.KeyEnter)
sb.InjectSpecialKey(platform.KeyTab)
sb.InjectSpecialKey(platform.KeyEscape)

// 带修饰符的按键
sb.InjectKeyWithMod('s', platform.KeyModCtrl) // Ctrl+S

// 字符串（自动转换为按键序列）
sb.InjectString("hello world")

// 鼠标事件
sb.InjectMouse(10, 5, platform.MouseLeft, platform.MousePress)
sb.InjectMouse(10, 5, platform.MouseLeft, platform.MouseRelease)

// 窗口调整事件
sb.InjectResize(120, 40)

// 处理所有事件
sb.ProcessEvents()
```

### 测试渲染输出

```go
// 获取渲染字符串
output := sb.RenderString()
fmt.Println(output)

// 断言包含特定文本
err := sb.AssertRender("Welcome")
if err != nil {
    t.Errorf("Expected 'Welcome' in output: %v", err)
}

// 断言不包含特定文本
err = sb.AssertNotRender("Error")
if err != nil {
    t.Errorf("Unexpected 'Error' in output: %v", err)
}

// 直接检查
output := sb.RenderString()
if !strings.Contains(output, "expected") {
    t.Error("Output does not contain expected text")
}
```

### 检查队列统计

```go
stats := sb.QueueStats()
fmt.Printf("Queue length: %d\n", stats.Length)
fmt.Printf("Memory used: %d bytes\n", stats.MemoryUsed)
fmt.Printf("Memory limit: %d bytes\n", stats.MemoryLimit)
fmt.Printf("Evicted count: %d\n", stats.EvictCount)
```

### 完整测试示例

```go
func TestLoginForm(t *testing.T) {
    // 设置沙箱
    sb := mock.New(80, 24)
    defer sb.Close()
    sb.Initialize(nil)
    sb.Run()

    // 模拟用户登录
    sb.InjectString("user@example.com")  // 输入用户名
    sb.InjectSpecialKey(platform.KeyTab)   // 切换到密码框
    sb.InjectString("secret123")          // 输入密码
    sb.InjectSpecialKey(platform.KeyEnter) // 提交表单

    // 处理所有事件
    if err := sb.ProcessEvents(); err != nil {
        t.Fatal(err)
    }

    // 验证登录成功
    if err := sb.AssertRender("Welcome"); err != nil {
        t.Errorf("Login failed: %v", err)
    }

    // 验证错误消息不存在
    if err := sb.AssertNotRender("Invalid"); err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

---

## 真实终端测试

### 启动真实沙箱

```go
package main

import (
    "log"
    "github.com/wwsheng009/mint/sandbox/real"
)

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

    // 等待用户交互（事件循环在后台运行）
    select {}
}
```

### 事件录制

真实沙箱会自动录制所有事件：

```go
// 用户交互后获取录制的事件
events := sb.RecordedEvents()
fmt.Printf("Recorded %d events\n", len(events))

// 保存到文件（自定义）
if err := saveEventsToFile(events, "session.json"); err != nil {
    log.Fatal(err)
}
```

### 在录制期间创建快照

```go
// 创建带标签的快照
snap, err := sb.Snapshot(sandbox.SnapshotStandard, "before-login")
if err != nil {
    log.Fatal(err)
}

// 稍后恢复
if err := sb.Restore(snap); err != nil {
    log.Fatal(err)
}
```

### 真实沙箱配置

```go
config := sandbox.RealConfig()
config.Width = 120
config.Height = 40
config.Event.RecordEnabled = true
config.Snapshot.MaxCount = 50

sb, err := real.NewWithConfig(config)
```

---

## 事件录制与回放

### 录制工作流

```go
// 1. 在真实沙箱中录制
realSB, _ := real.New(80, 24)
realSB.Initialize(nil)
realSB.Run()

// ... 用户交互 ...

// 2. 获取录制的事件
events := realSB.RecordedEvents()

// 3. 保存到文件（可选）
// saveEvents(events, "recording.json")

// 4. 创建回放沙箱
replaySB := replay.New(events, 80, 24)
replaySB.Initialize(nil)
replaySB.Run()
```

### 回放控制

```go
player := replaySB.Player()

// 控制播放
player.Play()   // 开始播放
player.Pause()  // 暂停
player.Stop()   // 停止并重置

// 导航
player.Next()      // 下一个事件
player.Previous()  // 上一个事件
player.Seek(10)    // 跳转到索引 10

// 状态查询
index := player.Index()
total := player.Length()
hasNext := player.HasNext()
isPlaying := player.IsPlaying()
```

### 变速回放

```go
player.SetSpeed(2.0)  // 2 倍速
player.SetSpeed(0.5)  // 0.5 倍速
player.SetSpeed(1.0)  // 正常速度

// 获取当前速度
speed := player.Speed()
```

### 逐步调试

```go
for player.HasNext() {
    // 获取下一个事件
    event, err := player.Next()
    if err != nil {
        break
    }

    // 检查事件
    fmt.Printf("Event %d: Type=%v\n", player.Index(), event.Type)

    // 在关键点设置断点
    if event.Type == platform.InputKeyPress && event.Key == 's' {
        // 处理特殊按键
    }
}
```

---

## 快照功能

### 快照级别

| 级别 | 内容 | 用途 |
|------|------|------|
| **Minimal** | 仅渲染缓冲区 | 快速检查渲染 |
| **Standard** | 缓冲区 + 事件历史 | 调试、测试 |
| **Full** | 缓冲区 + 事件 + 状态 | 完整状态恢复 |

### 创建快照

```go
// 创建 Minimal 快照
snap, err := sb.Snapshot(sandbox.SnapshotMinimal, "checkpoint1")

// 创建 Standard 快照（默认）
snap, err := sb.Snapshot(sandbox.SnapshotStandard)

// 创建 Full 快照
snap, err := sb.Snapshot(sandbox.SnapshotFull, "important-state")

// 带多个标签
snap, err := sb.Snapshot(sandbox.SnapshotStandard, "tag1", "tag2", "tag3")
```

### 恢复快照

```go
// 恢复到快照状态
if err := sb.Restore(snap); err != nil {
    log.Fatal(err)
}

// 恢复后可以继续操作
sb.InjectString("new input")
sb.ProcessEvents()
```

### 列出快照

```go
snapshots := sb.ListSnapshots()
for _, meta := range snapshots {
    fmt.Printf("ID: %s, Time: %s, Level: %s\n",
        meta.ID, meta.Timestamp, meta.Level)
    fmt.Printf("Tags: %v\n", meta.Tags)
    fmt.Printf("Size: %d bytes\n", meta.Size)
}
```

### 快照元数据

```go
type SnapshotMetadata struct {
    ID        string         // 唯一标识
    Timestamp time.Time      // 创建时间
    Level     SnapshotLevel  // 快照级别
    Tags      []string       // 自定义标签
    Size      int64          // 大小（字节）
}
```

### 快照验证

```go
// 快照管理器会自动验证完整性
snapMgr := sandbox.NewSnapshotManager(100)

// 创建快照时自动计算校验和
snap, _ := snapMgr.Create(level, buffer, events, state, "test")

// 验证快照
isValid := snapMgr.Verify(snap)
if !isValid {
    log.Fatal("Snapshot corrupted!")
}
```

### 完整快照工作流

```go
func TestWithSnapshots(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    sb.Run()

    // 1. 创建初始快照
    initial, _ := sb.Snapshot(sandbox.SnapshotStandard, "initial")

    // 2. 执行操作
    sb.InjectString("action1")
    sb.ProcessEvents()

    // 3. 创建中间快照
    mid, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-action1")

    // 4. 继续操作
    sb.InjectString("action2")
    sb.ProcessEvents()

    // 5. 验证最终状态
    sb.AssertRender("result2")

    // 6. 回滚到中间状态
    sb.Restore(mid)
    sb.AssertRender("result1")  // 应该看到 action1 的结果

    // 7. 回滚到初始状态
    sb.Restore(initial)
    // 应该看不到任何操作的结果
}
```

---

## 链式测试 API

### 基本链式调用

```go
result := sb.Helper().
    Type("hello").
    Tab().
    Type("world").
    Enter().
    Process().
    Result()

if !result.OK() {
    t.Errorf("Chain failed: %v", result.Error())
}
```

### 可用的链式方法

#### 动作方法

```go
helper.
    Type("text").           // 输入文本
    Press(platform.KeyTab). // 按下特殊键
    PressKey('a').          // 按下字符键
    Click(10, 5).           // 点击
    Tab().                  // 按 Tab
    Enter().                // 按 Enter
    Escape().               // 按 Escape
    Process().              // 处理所有事件
    Wait(100*time.Millisecond) // 等待一段时间
```

#### 断言方法

```go
helper.
    AssertRender("expected").      // 断言包含文本
    AssertNotRender("unexpected")  // 断言不包含文本
```

### Result 处理

```go
result := helper.Type("test").Enter().Process().Result()

// 检查是否成功
if result.OK() {
    fmt.Println("All operations succeeded")
}

// 获取第一个错误
if err := result.Error(); err != nil {
    log.Fatal(err)
}

// 获取所有错误
for _, err := range result.Errors {
    fmt.Printf("Error: %v\n", err)
}
```

### 复杂测试场景

```go
func TestComplexForm(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    sb.Run()

    result := sb.Helper().
        // 填写表单
        Type("John Doe").
        Tab().
        Type("john@example.com").
        Tab().
        Type("password123").
        Tab().
        Type("123 Main St").
        Tab().
        Type("New York").
        Tab().
        Type("10001").
        // 导航到提交按钮
        Tab().
        Tab().
        // 提交
        Enter().
        Process().
        // 验证结果
        AssertRender("Success").
        AssertNotRender("Error").
        Result()

    if !result.OK() {
        t.Fatalf("Form submission failed: %v", result.Error())
    }
}
```

### 在断言之间清理错误

```go
helper := sb.Helper()

// 第一组测试
helper.Type("test1").Enter().Process().AssertRender("result1").Result()
if helper.HasErrors() {
    t.Errorf("Test 1 failed: %v", helper.Errors())
}

// 清理错误，继续测试
helper.ClearErrors()

// 第二组测试
helper.Type("test2").Enter().Process().AssertRender("result2").Result()
if helper.HasErrors() {
    t.Errorf("Test 2 failed: %v", helper.Errors())
}
```

### 结合快照使用链式 API

```go
func TestChainWithSnapshots(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    sb.Run()

    // 执行操作并创建快照
    sb.Helper().
        Type("initial").
        Enter().
        Process().
        AssertRender("initialized").
        Result()

    snap1, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-init")

    // 继续操作
    sb.Helper().
        Type("next").
        Enter().
        Process().
        AssertRender("processed").
        Result()

    // 恢复并重新测试
    sb.Restore(snap1)
    sb.Helper().
        Type("alternative").
        Enter().
        Process().
        AssertRender("alternative-result").
        Result()
}
```

---

## UI 层集成

### 使用 RunTest (推荐 - 新版 API)

```go
package ui_test

import (
    "testing"
    "github.com/wwsheng009/mint/ui"
)

func TestComponent(t *testing.T) {
    // 创建测试应用（完整框架支持）
    testApp, err := ui.RunTest(MyComponent,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

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

### 使用 RunTestWithSandbox (高级功能)

```go
func TestWithSandbox(t *testing.T) {
    // 使用 MockSandbox 作为事件源
    testApp, err := ui.RunTestWithSandbox(MyComponent,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 获取 MockSandbox（用于高级功能）
    sb := testApp.GetSandbox()

    // 通过 Sandbox 注入事件
    sb.InjectSpecialKey(platform.KeyTab)
    sb.InjectSpecialKey(platform.KeyEnter)

    // 直接注入也支持
    testApp.InjectKey('a')

    // 获取队列统计
    stats := sb.QueueStats()
    t.Logf("Queue length: %d", stats.Length)
}
```

### 使用 TestRun (Deprecated - 仅用于简单组件)

```go
// Deprecated: 推荐使用 ui.RunTest() 获取完整的框架功能支持
func TestComponent_Simple(t *testing.T) {
    testApp, err := ui.TestRun(app, ui.TestWithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 使用沙箱进行测试
    helper := testApp.Helper()
    helper.Type("test").Enter().Process().AssertRender("result")
}
```

### 使用 TestRunWithConfig (Deprecated)

```go
func TestWithCustomConfig(t *testing.T) {
    config := sandbox.DefaultConfig()
    config.Width = 120
    config.Height = 40
    config.Event.QueueMaxMemory = 50 * 1024 * 1024  // 50MB

    testApp, err := ui.TestRunWithConfig(app, config)
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // ... 测试 ...
}
```

### TestOptions (新版 API)

```go
// 设置宽度
ui.RunTest(app, ui.WithWidth(100))

// 设置高度
ui.RunTest(app, ui.WithHeight(30))

// 设置标题
ui.RunTest(app, ui.WithTitle("My Test"))

// 设置帧率
ui.RunTest(app, ui.WithFPS(60))

// 多个选项
ui.RunTest(app,
    ui.WithWidth(100),
    ui.WithHeight(30),
    ui.WithTitle("My Test"),
)
```

### TestOptions (旧版 API - Deprecated)

```go
// Deprecated: 推荐使用新版 WithWidth/WithHeight
ui.TestRun(app, ui.TestWithWidth(100))
ui.TestRun(app, ui.TestWithHeight(30))
ui.TestRun(app, ui.TestWithSize(100, 30))
```

### 访问底层沙箱

```go
// 新版 API (RunTestWithSandbox)
testApp, _ := ui.RunTestWithSandbox(app)
sb := testApp.GetSandbox()
if sb != nil {
    // 使用 MockSandbox 的高级功能
    sb.InjectString("test")
    stats := sb.QueueStats()
}

// 旧版 API (TestRun)
testApp, _ := ui.TestRun(app)
sb := testApp.Sandbox()
sb.InjectString("test")
sb.ProcessEvents()
```

---

## 最佳实践

### 1. 测试隔离

```go
// ✅ 好：每个测试独立
func TestFeatureA(t *testing.T) {
    sb := mock.New(80, 24)
    defer sb.Close()
    // ... 测试 A ...
}

func TestFeatureB(t *testing.T) {
    sb := mock.New(80, 24)
    defer sb.Close()
    // ... 测试 B ...
}

// ❌ 差：共享沙箱导致测试相互影响
var sb = mock.New(80, 24)

func TestFeatureA(t *testing.T) {
    // 测试 A 修改了 sb 的状态
}

func TestFeatureB(t *testing.T) {
    // 测试 B 受到测试 A 的影响
}
```

### 2. 资源清理

```go
// ✅ 好：使用 defer 清理
func TestWithCleanup(t *testing.T) {
    sb := mock.New(80, 24)
    defer sb.Close()  // 确保清理

    sb.Initialize(nil)
    sb.Run()
    // ... 测试 ...
}

// ❌ 差：忘记清理
func TestWithoutCleanup(t *testing.T) {
    sb := mock.New(80, 24)
    // 忘记 Close()
}
```

### 3. 使用辅助器简化代码

```go
// ✅ 好：使用辅助器创建沙箱
func createTestSandbox() (*mock.MockSandbox, error) {
    sb := mock.New(80, 24)
    if err := sb.Initialize(nil); err != nil {
        return nil, err
    }
    if err := sb.Run(); err != nil {
        return nil, err
    }
    return sb, nil
}

func TestHelper(t *testing.T) {
    sb, err := createTestSandbox()
    if err != nil {
        t.Fatal(err)
    }
    defer sb.Close()
    // ... 测试 ...
}
```

### 4. 链式 API vs 直接调用

```go
// ✅ 好：简单场景使用链式 API
result := sb.Helper().Type("test").Enter().Process().AssertRender("ok").Result()

// ✅ 好：复杂场景使用直接调用，更灵活
sb.InjectString("test")
if err := sb.ProcessEvents(); err != nil {
    return err
}
output := sb.RenderString()
// ... 复杂的验证逻辑 ...
```

### 5. 快照命名

```go
// ✅ 好：使用有意义的标签
snap, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-login", "user-authenticated")

// ❌ 差：无意义的标签
snap, _ := sb.Snapshot(sandbox.SnapshotStandard, "snap1")
```

### 6. 错误处理

```go
// ✅ 好：检查所有错误
if err := sb.Initialize(nil); err != nil {
    t.Fatalf("Initialize failed: %v", err)
}

if err := sb.Run(); err != nil {
    t.Fatalf("Run failed: %v", err)
}

// ❌ 差：忽略错误
sb.Initialize(nil)
sb.Run()
```

### 7. 并发测试

```go
// ✅ 好：沙箱是线程安全的
func TestConcurrent(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    sb.Run()

    var wg sync.WaitGroup

    // 多个 goroutine 同时注入事件
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            sb.InjectString(fmt.Sprintf("user%d", id))
        }(i)
    }

    wg.Wait()
    sb.ProcessEvents()
}
```

---

## 常见问题

### Q1: Mock 沙箱无法渲染？

**问题：** RenderString() 返回空字符串

**原因：** 没有设置事件处理器或组件未渲染到缓冲区

**解决：**
```go
// 设置事件处理器并确保组件渲染到缓冲区
sb.SetEventHandler(func(event platform.RawInput) error {
    // 渲染组件到 sb.Buffer()
    component.Render(sb.Buffer())
    return nil
})
```

### Q2: 事件队列满了？

**问题：** 注入事件时返回 `ErrQueueFull`

**原因：** 事件队列达到容量限制

**解决：**
```go
// 增加队列容量
config := sandbox.MockConfig()
config.Event.QueueMaxSize = 50000  // 从 10000 增加到 50000

sb := mock.NewWithConfig(config)

// 或者处理更多事件
for !sb.QueueStats().Length >= 1000 {
    sb.ProcessEvents()
}
```

### Q3: 内存占用过高？

**问题：** 测试内存占用持续增长

**原因：** 大量事件未处理或快照未清理

**解决：**
```go
// 定期处理事件
sb.ProcessEvents()

// 限制快照数量
config.Snapshot.MaxCount = 10

// 减少录制器长度
config.Event.RecordMaxLen = 1000
```

### Q4: 快照恢复失败？

**问题：** Restore() 返回错误

**原因：** 快照数据损坏或格式不匹配

**解决：**
```go
// 验证快照
if !sb.snapMgr.Verify(snap) {
    log.Fatal("Snapshot corrupted")
}

// 确保快照格式匹配
// Minimal 快照只包含 Buffer，不能依赖 Events 或 State
```

### Q5: 回放沙箱事件不匹配？

**问题：** 回放时行为与录制不一致

**原因：** 非确定性状态或时间依赖

**解决：**
```go
// 使用 Full 级别快照捕获完整状态
snap, _ := sb.Snapshot(sandbox.SnapshotFull)

// 在回放前恢复初始状态
sb.Restore(initialSnap)

// 使用单步回放调试
for player.HasNext() {
    event, _ := player.Next()
    // 检查每个事件
}
```

### Q6: 如何测试组件焦点？

```go
// 模拟 Tab 键切换焦点
sb.InjectSpecialKey(platform.KeyTab)
sb.ProcessEvents()

// 验证焦点组件
rendered := sb.RenderString()
if !strings.Contains(rendered, "[focused]") {
    t.Error("Focus not set correctly")
}
```

### Q7: 如何模拟长时间运行？

```go
// 使用快照避免重复初始化
setup, _ := sb.Snapshot(sandbox.SnapshotFull, "setup")

// 多次测试使用相同初始状态
for i := 0; i < 100; i++ {
    sb.Restore(setup)
    sb.InjectString(fmt.Sprintf("test%d", i))
    sb.ProcessEvents()
    // 验证
}
```

### Q8: 如何捕获所有事件？

```go
// 自定义事件处理器
var capturedEvents []platform.RawInput

sb.SetEventHandler(func(event platform.RawInput) error {
    capturedEvents = append(capturedEvents, event)
    // 正常处理事件
    return nil
})

sb.InjectString("test")
sb.ProcessEvents()

fmt.Printf("Captured %d events\n", len(capturedEvents))
```

---

## 进阶主题

### 自定义事件处理

```go
// 创建事件过滤器
filter := func(event platform.RawInput) error {
    // 忽略某些事件
    if event.Type == platform.InputMouse {
        return nil
    }
    // 处理其他事件
    return actualHandler(event)
}

sb.SetEventHandler(filter)
```

### 压力测试

```go
func TestStress(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    sb.Run()

    // 注入大量事件
    for i := 0; i < 100000; i++ {
        sb.InjectKey('a')
    }

    // 验证内存使用
    stats := sb.QueueStats()
    if stats.MemoryUsed > 100*1024*1024 {
        t.Errorf("Memory too high: %d", stats.MemoryUsed)
    }
}
```

### 调试技巧

```go
// 启用调试输出
debug := func(event platform.RawInput) error {
    fmt.Printf("[DEBUG] Event: %+v\n", event)
    return nil
}

sb.SetEventHandler(debug)

// 或使用快照对比
snap1, _ := sb.Snapshot(sandbox.SnapshotStandard, "before")
// ... 操作 ...
snap2, _ := sb.Snapshot(sandbox.SnapshotStandard, "after")

// 比较快照
// (自定义比较逻辑)
```

---

## 附录

### A. 错误代码参考

```go
// 生命周期错误
ErrInvalidTransition  // 无效的状态转换
ErrNotInitialized     // 未初始化
ErrAlreadyRunning     // 已在运行
ErrNotRunning         // 未运行

// 事件注入错误
ErrInjectionNotAllowed  // 不允许注入
ErrInvalidStrategy      // 无效策略
ErrQueueFull           // 队列已满
ErrQueueEmpty          // 队列为空

// 快照错误
ErrSnapshotNotFound  // 快照未找到
ErrSnapshotCorrupt   // 快照损坏
ErrRestoreFailed     // 恢复失败

// 配置错误
ErrInvalidConfig  // 无效配置

// 断言错误
ErrAssertionFailed  // 断言失败
ErrTimeout          // 超时
```

### B. 默认配置值

```go
DefaultConfig():
  Width:  80
  Height: 24
  FPS:    60

  Event:
    QueueMaxSize:   10000
    QueueMaxMemory: 100MB
    EvictPolicy:    EvictOldest
    Strategy:       InjectAllowed
    RecordMaxLen:   10000

  Snapshot:
    MaxCount: 100
    Level:    SnapshotStandard

  Performance:
    Throttle:      true
    MaxFPS:        60
    RenderTimeout: 100ms
```

### C. 平台按键常量

```go
platform.KeyTab
platform.KeyEnter
platform.KeyEscape
platform.KeySpace
platform.KeyBackspace
platform.KeyDelete
platform.KeyUp
platform.KeyDown
platform.KeyLeft
platform.KeyRight

// 修饰符
platform.KeyModCtrl
platform.KeyModAlt
platform.KeyModShift

// 鼠标按钮
platform.MouseLeft
platform.MouseMiddle
platform.MouseRight

// 鼠标动作
platform.MousePress
platform.MouseRelease
platform.MouseScroll
```

---

**手册结束**

如有问题或建议，请参考 API 手册或提交 Issue。
