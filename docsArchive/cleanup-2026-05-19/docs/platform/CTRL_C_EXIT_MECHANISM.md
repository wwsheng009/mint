# Ctrl+C 退出机制

## 概述

Mint 在 Linux/Darwin 平台使用 tcell 库处理终端输入。Ctrl+C 的处理需要特别注意，因为 tcell 将其作为键盘事件而非信号返回。

## tcell 的行为

tcell 对于 Ctrl+C 的处理方式：
- **返回 `tcell.KeyCtrlC` 键盘事件**，而不是发送 SIGINT 信号
- 这意味着传统的 `signal.Notify(SIGINT)` 无法捕获 Ctrl+C

## 退出机制设计

### 数据流

```
用户按下 Ctrl+C
    ↓
tcell 检测到按键，返回 EventKey{Key: KeyCtrlC}
    ↓
parseKeyEvent() 检测到 KeyCtrlC
    ├─ 设置 input.Key = 'c'
    ├─ 设置 input.Modifiers = ModCtrl
    └─ 设置 exitFlag = 1 (原子操作)
    ↓
RawInput 传递到 Pump
    ↓
convertToKeyMsg() 检测 Ctrl+C
    ├─ 关闭 quitApp 通道
    └─ 调用 p.Stop() 停止 Pump
    ↓
App.Run() 主循环
    ├─ 检测到 quitApp 通道关闭
    ├─ 设置 a.state = StateStopping
    └─ 退出循环
    ↓
App.Close() 清理资源
    └─ 调用 restoreTerminal() 恢复终端
```

## 代码实现

### 1. 平台层 (input_linux_tcell.go / input_darwin_tcell.go)

```go
// 全局退出标志
var exitFlag int32 = 0

func (r *tcellInputReader) parseKeyEvent(ev *tcell.EventKey, now time.Time) RawInput {
    // ... 修饰键处理 ...

    switch ev.Key() {
    case tcell.KeyCtrlC:
        input.Key = 'c'
        input.Modifiers |= ModCtrl
        // 设置退出标志，让应用能够退出
        atomic.StoreInt32(&exitFlag, 1)
    // ... 其他键处理 ...
    }
}

func (r *tcellInputReader) readLoop() {
    for {
        // 检查是否收到退出信号
        if atomic.LoadInt32(&exitFlag) != 0 {
            return
        }
        // ... 事件处理 ...
    }
}
```

### 2. 事件泵层 (framework/event/pump.go)

```go
type Pump struct {
    quit     chan struct{}     // 内部退出
    quitApp  chan struct{}     // Ctrl+C 退出通知通道
    // ...
}

func (p *Pump) convertToKeyMsg(raw platform.RawInput) runtimemsg.Msg {
    // ... 修饰键转换 ...

    // 检查 Ctrl+C 组合键
    if raw.Modifiers&platform.ModCtrl != 0 && raw.Key == 'c' {
        // Ctrl+C 被按下，通知应用退出
        close(p.quitApp)
        // 仍然返回消息让上层处理（如果需要）
    }

    return runtimemsg.NewKeyMsg(raw.Key, raw.Special, modifiers)
}

func (p *Pump) QuitAppRequested() <-chan struct{} {
    return p.quitApp
}
```

### 3. 应用层 (framework/app.go)

```go
func (a *App) Run() error {
    // ... 初始化 ...

    eventChan := a.pump.Events()
    quitAppChan := a.pump.QuitAppRequested()

    for a.state == StateRunning {
        select {
        case msg := <-eventChan:
            // ... 处理消息 ...
        case <-quitAppChan:
            // Ctrl+C 退出
            a.state = StateStopping
            return nil
        case <-ticker.C:
            // ... tick 处理 ...
        }
    }

    return nil
}
```

## 为什么需要三层设计？

1. **平台层** (`exitFlag`)
   - 让 `readLoop()` 能够快速退出
   - 防止在 `PollEvent()` 调用中阻塞

2. **事件泵层** (`quitApp` 通道)
   - 将 Ctrl+C 检测与内部 `quit` 通道分离
   - 允许应用优雅地关闭，而不是立即崩溃

3. **应用层** (主循环 select)
   - 统一的退出点
   - 可以处理多种退出方式（Ctrl+C、Quit 命令、其他信号）

## 与 Windows 平台的区别

| 特性 | Windows | Linux/Darwin |
|------|---------|---------------|
| Ctrl+C 检测方式 | 信号处理器 (`signal.Notify`) | 键盘事件 (`tcell.KeyCtrlC`) |
| 终端恢复 | `SetConsoleMode()` | `screen.Fini()` |
| 信号处理 | `init()` 中注册 | `init()` 中注册 + tcell 键盘事件 |

## 故障排除

### Ctrl+C 需要按两次才退出

**原因**: `quitApp` 通道没有被正确监听，或 Pump 没有正确停止。

**检查**:
1. 确认 `p.QuitAppRequested()` 返回正确的通道
2. 确认 App.Run() 中监听了这个通道
3. 确认 `convertToKeyMsg()` 中关闭了这个通道

### Ctrl+C 完全无反应

**原因**: tcell 未正确初始化，或键盘事件未正确处理。

**检查**:
1. 启用 `TUI_DEBUG_INPUT=true` 查看原始事件
2. 检查 `parseKeyEvent()` 中是否有 `case tcell.KeyCtrlC`
3. 检查是否设置了 `exitFlag`

## 相关文件

- `runtime/platform/input_linux_tcell.go` - Linux tcell 输入实现
- `runtime/platform/input_darwin_tcell.go` - Darwin tcell 输入实现
- `framework/event/pump.go` - 事件泵和消息转换
- `framework/app.go` - 应用主循环
- `docs/platform/SIGNAL_HANDLING.md` - 信号处理总览
