# Sandbox 高级功能指南

> Mint TUI 框架 MockSandbox 高级功能完整参考
>
> 版本: 1.0
> 更新日期: 2026-02-01

---

## 目录

1. [概述](#概述)
2. [事件录制与回放](#事件录制与回放)
3. [事件队列管理](#事件队列管理)
4. [测试辅助器 (TestHelper)](#测试辅助器-testhelper)
5. [快照系统](#快照系统)
6. [事件注入策略](#事件注入策略)
7. [配置系统](#配置系统)
8. [完整示例](#完整示例)

---

## 1. 概述

MockSandbox 提供了以下高级功能：

| 功能 | 描述 | 使用场景 |
|------|------|----------|
| **事件录制** | 记录用户操作序列 | 重现 bug、演示录制 |
| **事件回放** | 按录制序列重放操作 | 自动化测试、调试 |
| **队列统计** | 实时监控事件队列 | 性能分析、内存监控 |
| **TestHelper** | 链式 API 测试辅助 | 快速编写测试用例 |
| **快照系统** | 保存和恢复应用状态 | 调试状态转换、测试隔离 |
| **注入策略** | 控制事件注入行为 | 环境隔离 |

---

## 2. 事件录制与回放

### 2.1 事件录制

```go
import (
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/mock"
)

func TestEventRecording(t *testing.T) {
    // 创建带录制功能的沙箱
    recorder := sandbox.NewEventRecorder(10000) // 最多录制10000个事件

    sb := mock.New(80, 24)
    sb.SetRecorder(recorder)

    sb.Initialize(nil)
    defer sb.Close()

    // 运行应用并录制事件...
    sb.InjectKey('a')
    sb.InjectKey('b')
    sb.InjectSpecialKey(platform.KeyEnter)

    // 获取录制的事件
    events := recorder.Events()
    t.Logf("录制了 %d 个事件", len(events))

    // 保存到文件
    recorder.SaveToFile("test_events.json")
}
```

### 2.2 事件回放

```go
import (
    "github.com/wwsheng009/mint/sandbox/replay"
)

func TestEventReplay(t *testing.T) {
    // 加载录制的事件
    recorder := sandbox.NewEventRecorder(10000)
    recorder.LoadFromFile("test_events.json")
    events := recorder.Events()

    // 创建回放沙箱
    replaySandbox := replay.New(events, 80, 24)

    // 设置回放速度
    replaySandbox.SetSpeed(2.0) // 2倍速

    // 单步执行
    replaySandbox.Step() // 前进一步
    replaySandbox.StepBack() // 后退一步

    // 自动回放
    replaySandbox.Play() // 自动播放所有事件
}
```

### 2.3 录制器 API

```go
// 创建录制器
recorder := sandbox.NewEventRecorder(maxSize)

// 录制事件
recorder.Record(event platform.RawInput)

// 获取事件
events := recorder.Events()

// 清空录制
recorder.Clear()

// 文件操作
recorder.SaveToFile(path string) error
recorder.LoadFromFile(path string) error

// 统计信息
count := recorder.Count()
bytes := recorder.Size()
```

---

## 3. 事件队列管理

### 3.1 队列配置

```go
import "github.com/wwsheng009/mint/sandbox"

// 创建有界队列
queue := sandbox.NewBoundedQueue(sandbox.QueueConfig{
    MaxSize:     10000,                // 最大事件数
    MaxMemory:   100 * 1024 * 1024,    // 最大内存 100MB
    EvictPolicy: sandbox.EvictOldest,  // 淘汰策略
})

// 淘汰策略选项
sandbox.EvictOldest   // 淘汰最旧的事件（默认）
sandbox.EvictPriority // 淘汰低优先级事件
sandbox.EvictPersist  // 持久化后淘汰
```

### 3.2 队列操作

```go
// 基本操作
queue.Push(event platform.RawInput) error
event := queue.Pop() platform.RawInput
length := queue.Len() int
isEmpty := queue.IsEmpty() bool

// 查看事件
event := queue.Peek() platform.RawInput // 查看最旧事件但不移除

// 批量操作
events := queue.PopAll() []platform.RawInput
queue.Clear()
```

### 3.3 队列统计

```go
// 通过 MockSandbox 获取统计
sb := mock.New(80, 24)
sb.Initialize(nil)

stats := sb.QueueStats()

// stats 包含：
type QueueStats struct {
    Length      int     // 当前队列长度
    MemoryUsed  int64   // 已用内存（字节）
    MemoryLimit int64   // 内存限制
    EvictCount  int     // 已淘汰事件数
    MaxSize     int     // 队列容量限制
}

t.Logf("队列: %d/%d 事件, %d/%d 字节",
    stats.Length, stats.MaxSize,
    stats.MemoryUsed, stats.MemoryLimit)
```

### 3.4 内存监控示例

```go
func TestQueueMemory(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    defer sb.Close()

    // 注入大量事件
    for i := 0; i < 10000; i++ {
        sb.InjectKey('a')
    }

    stats := sb.QueueStats()

    // 检查内存使用
    if stats.MemoryUsed > stats.MemoryLimit {
        t.Errorf("内存超限: %d > %d",
            stats.MemoryUsed, stats.MemoryLimit)
    }

    // 检查淘汰情况
    if stats.EvictCount > 0 {
        t.Logf("警告: %d 个事件被淘汰", stats.EvictCount)
    }
}
```

---

## 4. 测试辅助器 (TestHelper)

### 4.1 链式 API 基础

```go
import "github.com/wwsheng009/mint/sandbox/mock"

func TestHelperBasic(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    defer sb.Close()

    // 获取测试辅助器
    helper := sb.Helper()

    // 链式调用
    helper.
        Type("hello").
        Press(platform.KeyEnter).
        Wait(100 * time.Millisecond).
        AssertRender("hello")
}
```

### 4.2 可用操作

```go
helper := sb.Helper()

// 文本输入
helper.Type("text")          // 输入字符串
helper.TypeFast("text")      // 快速输入（无延迟）

// 按键操作
helper.Press(platform.KeyEnter)
helper.Press(platform.KeyTab)
helper.Press(platform.KeyEscape)

// 组合键
helper.PressWithMod('c', platform.ModCtrl)  // Ctrl+C
helper.PressWithMod('s', platform.ModCtrl)  // Ctrl+S

// 鼠标操作
helper.Click(10, 20)                          // 点击
helper.ClickButton(platform.MouseLeft)        // 指定按钮
helper.Drag(5, 5, 15, 15)                     // 拖拽

// 等待
helper.Wait(100 * time.Millisecond)
helper.WaitIdle()                             // 等待空闲

// 断言
helper.AssertRender("expected")
helper.AssertNotRender("unexpected")

// 结果获取
result := helper.Result()
if !result.OK() {
    t.Error(result.Error())
}
```

### 4.3 链式调用模式

```go
// 模式1: 简单操作序列
helper.
    Type("username").
    Press(platform.KeyTab).
    Type("password").
    Press(platform.KeyEnter).
    AssertRender("Welcome")

// 模式2: 带条件的操作
helper.
    Type("search term").
    Press(platform.KeyEnter).
    Wait(200 * time.Millisecond).
    AssertRender("results")

// 模式3: 获取中间结果
result1 := helper.Type("a").Press(platform.KeyEnter).Result()
if result1.OK() {
    helper.Type("b").Press(platform.KeyEnter)
}
```

### 4.4 自动事件收集

```go
// TestHelper 会自动收集所有操作的事件
events := helper.CollectedEvents()
t.Logf("生成了 %d 个事件", len(events))

// 可以重放到另一个沙箱
sb2 := mock.New(80, 24)
sb2.Initialize(nil)
for _, ev := range events {
    sb2.InjectRaw(ev)
}
```

---

## 5. 快照系统

### 5.1 快照级别

```go
import "github.com/wwsheng009/mint/sandbox"

// 三种快照级别
sandbox.SnapshotMinimal  // 仅渲染缓冲区
sandbox.SnapshotStandard // 缓冲区 + 事件历史
sandbox.SnapshotFull     // 包括应用状态
```

### 5.2 创建和恢复快照

```go
func TestSnapshots(t *testing.T) {
    sb := mock.New(80, 24)
    sb.Initialize(nil)
    defer sb.Close()

    // 执行一些操作
    sb.InjectKey('a')
    sb.InjectKey('b')

    // 创建快照
    snapshot, err := sb.Snapshot(
        sandbox.SnapshotStandard,
        "state-after-ab",
    )
    if err != nil {
        t.Fatal(err)
    }

    // 继续操作
    sb.InjectKey('c')

    // 恢复快照
    err = sb.Restore(snapshot)
    if err != nil {
        t.Fatal(err)
    }

    // 现在状态应该回到 'ab' 之后
    output := sb.RenderString()
    // ...
}
```

### 5.3 快照元数据

```go
type Snapshot struct {
    // 元数据
    ID        string    // 唯一标识
    Timestamp time.Time // 创建时间
    Level     SnapshotLevel // 快照级别
    Tags      []string  // 用户标签
    Size      int64     // 快照大小（字节）
    Checksum  string    // 校验和

    // 内容
    Buffer  *paint.Buffer    // 渲染缓冲区
    Events  []platform.RawInput // 事件历史（标准/完整）
    State   interface{}      // 应用状态（完整级别）
}

// 访问元数据
fmt.Printf("快照 ID: %s\n", snapshot.ID)
fmt.Printf("大小: %d 字节\n", snapshot.Size)
fmt.Printf("标签: %v\n", snapshot.Tags)
```

### 5.4 快照管理

```go
// 列出所有快照
snapshots := sb.ListSnapshots()
for _, snap := range snapshots {
    t.Logf("快照: %s (%v)", snap.ID, snap.Tags)
}

// 按标签查找
byTag := sb.FindSnapshots("tag1")

// 删除快照
sb.DeleteSnapshot(snapshot.ID)

// 清空所有快照
sb.ClearSnapshots()
```

### 5.5 快照配置

```go
config := &sandbox.Config{
    Snapshot: sandbox.SnapshotConfig{
        MaxCount:  200,             // 最多保存200个快照
        Level:     sandbox.SnapshotFull,
        AutoTag:   true,            // 自动生成标签
        Compress:  true,            // 启用压缩
    },
}
```

---

## 6. 事件注入策略

### 6.1 注入策略类型

```go
// 三种注入策略
sandbox.InjectProhibited  // 禁止注入（生产环境）
sandbox.InjectAllowed     // 允许注入（测试环境，默认）
sandbox.InjectRecorded    // 仅录制注入
```

### 6.2 配置注入策略

```go
// 方式1: 通过配置
config := &sandbox.Config{
    Event: sandbox.EventConfig{
        Strategy: sandbox.InjectAllowed,
    },
}
sb := mock.NewWithConfig(config)

// 方式2: 动态设置
sb := mock.New(80, 24)
sb.SetInjectStrategy(sandbox.InjectProhibited)
```

### 6.3 注入方法

```go
sb := mock.New(80, 24)

// 字符键
sb.InjectKey('a')

// 特殊键
sb.InjectSpecialKey(platform.KeyEnter)
sb.InjectSpecialKey(platform.KeyTab)
sb.InjectSpecialKey(platform.KeyEscape)

// 带修饰符
sb.InjectKeyWithMod('c', platform.ModCtrl)

// 字符串（自动分解为按键序列）
sb.InjectString("hello world")

// 鼠标
sb.InjectMouse(x, y, platform.MouseLeft, platform.MousePress)
sb.InjectClick(10, 20)
sb.InjectDrag(5, 5, 15, 15)

// 原始事件
raw := platform.RawInput{...}
sb.InjectRaw(raw)
```

---

## 7. 配置系统

### 7.1 预定义配置

```go
// 真实环境配置
config := sandbox.RealConfig()
// - 禁止注入
// - 大容量队列
// - 禁用快照

// 模拟环境配置
config := sandbox.MockConfig()
// - 允许注入
// - 标准队列
// - 标准快照

// 回放环境配置
config := sandbox.ReplayConfig()
// - 禁止注入
// - 小队列
// - 禁用快照
```

### 7.2 自定义配置

```go
config := &sandbox.Config{
    // 尺寸
    Width:  120,
    Height: 40,

    // 事件配置
    Event: sandbox.EventConfig{
        Strategy:       sandbox.InjectAllowed,
        QueueMaxSize:   50000,
        QueueMaxMemory: 50 * 1024 * 1024, // 50MB
        EvictPolicy:    sandbox.EvictOldest,
        RecordEnabled:  true,
    },

    // 快照配置
    Snapshot: sandbox.SnapshotConfig{
        MaxCount:  200,
        Level:     sandbox.SnapshotFull,
        AutoTag:   true,
        Compress:  true,
    },

    // 调试配置
    Debug: sandbox.DebugConfig{
        LogLevel:     "debug",
        LogEvents:    true,
        LogQueue:     true,
    },
}

sb := mock.NewWithConfig(config)
```

---

## 8. 完整示例

### 8.1 完整的测试示例

```go
package main_test

import (
    "testing"
    "time"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
)

func TestCompleteExample(t *testing.T) {
    // 1. 创建测试应用
    testApp, err := ui.RunTestWithSandbox(MyApp,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    sb := testApp.GetSandbox()
    helper := sb.Helper()

    // 2. 使用 TestHelper 模拟用户操作
    helper.
        Type("username").
        Press(platform.KeyTab).
        Type("password").
        Press(platform.KeyEnter).
        Wait(200 * time.Millisecond).
        AssertRender("Welcome")

    // 3. 检查队列统计
    stats := sb.QueueStats()
    t.Logf("处理了 %d 个事件", stats.Length)

    // 4. 创建快照
    snapshot, err := sb.Snapshot(
        sandbox.SnapshotStandard,
        "after-login",
    )
    if err != nil {
        t.Fatal(err)
    }

    // 5. 继续操作
    helper.Type("search term").Press(platform.KeyEnter)

    // 6. 恢复到登录后状态
    err = sb.Restore(snapshot)
    if err != nil {
        t.Fatal(err)
    }

    // 7. 验证状态已恢复
    if err := testApp.AssertRender("Welcome"); err != nil {
        t.Error("快照恢复失败")
    }
}
```

### 8.2 事件录制与回放示例

```go
func TestRecordingAndReplay(t *testing.T) {
    // 1. 录制会话
    recorder := sandbox.NewEventRecorder(10000)

    testApp1, _ := ui.RunTestWithSandbox(MyApp)
    defer testApp1.Close()

    sb1 := testApp1.GetSandbox()
    sb1.SetRecorder(recorder)

    // 执行操作（自动录制）
    sb1.Helper().
        Type("test").
        Press(platform.KeyEnter).
        Wait(100 * time.Millisecond)

    // 2. 保存录制
    events := recorder.Events()
    t.Logf("录制了 %d 个事件", len(events))

    // 3. 回放到新实例
    testApp2, _ := ui.RunTestWithSandbox(MyApp)
    defer testApp2.Close()

    sb2 := testApp2.GetSandbox()

    for _, ev := range events {
        sb2.InjectRaw(ev)
        time.Sleep(1 * time.Millisecond) // 模拟原始时序
    }

    // 4. 验证结果一致
    render1 := testApp1.GetRenderString()
    render2 := testApp2.GetRenderString()

    if render1 != render2 {
        t.Error("回放结果与录制不一致")
    }
}
```

### 8.3 性能监控示例

```go
func TestPerformanceMonitoring(t *testing.T) {
    testApp, _ := ui.RunTestWithSandbox(MyApp)
    defer testApp.Close()

    sb := testApp.GetSandbox()

    // 记录初始状态
    initialStats := sb.QueueStats()

    // 执行大量操作
    for i := 0; i < 1000; i++ {
        sb.InjectKey('a')
        sb.InjectSpecialKey(platform.KeyBackspace)
    }

    // 记录最终状态
    finalStats := sb.QueueStats()

    // 分析性能
    processed := finalStats.Length - initialStats.Length
    evicted := finalStats.EvictCount - initialStats.EvictCount
    memoryUsed := finalStats.MemoryUsed - initialStats.MemoryUsed

    t.Logf("处理事件: %d", processed)
    t.Logf("淘汰事件: %d", evicted)
    t.Logf("内存使用: %d 字节", memoryUsed)

    // 性能断言
    if evicted > 100 {
        t.Errorf("过多事件被淘汰: %d", evicted)
    }
}
```

---

## 附录

### A. 相关文档

- [APP_LIFECYCLE_AND_SANDBOX.md](./APP_LIFECYCLE_AND_SANDBOX.md) - 应用生命周期与 Sandbox 集成
- [ENGINE_SANDBOX_INTEGRATION.md](./ENGINE_SANDBOX_INTEGRATION.md) - Engine 与 Sandbox 集成架构
- [SANDBOX_DEBUG_GUIDE.md](./SANDBOX_DEBUG_GUIDE.md) - Sandbox 调试技巧
- [API_REFERENCE.md](./API_REFERENCE.md) - 完整 API 参考

### B. 快速参考

```go
// 创建沙箱
sb := mock.New(width, height)
sb.Initialize(nil)
defer sb.Close()

// 事件注入
sb.InjectKey('a')
sb.InjectSpecialKey(platform.KeyEnter)
sb.InjectString("text")

// 测试辅助器
helper := sb.Helper()
helper.Type("text").Press(platform.KeyEnter).AssertRender("expected")

// 快照
snapshot, _ := sb.Snapshot(sandbox.SnapshotStandard, "tag")
sb.Restore(snapshot)

// 队列统计
stats := sb.QueueStats()

// 事件录制
recorder := sandbox.NewEventRecorder(10000)
sb.SetRecorder(recorder)
events := recorder.Events()
```
