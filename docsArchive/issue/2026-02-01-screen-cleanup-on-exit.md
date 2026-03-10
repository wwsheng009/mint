# 退出时清屏问题修复

## 问题描述

应用程序在以下情况下退出时，TUI 内容残留在终端屏幕上：
1. 正常退出（q 键、ESC 键）
2. Ctrl+C 中断退出
3. Panic 崩溃退出

## 根本原因

### 1. `framework/app.Close()` 缺少清屏

原代码只显示光标，不清除屏幕内容：

```go
func (a *App) Close() error {
    // ...
    a.ShowCursor()  // 仅显示光标
    // 缺少清屏操作
}
```

### 2. `runtime/core/runtime.Shutdown()` 缺少清屏

```go
func (r *Runtime) Shutdown(timeout ...time.Duration) error {
    // ...
    if err := r.platform.Close(); err != nil {
        return err
    }
    // 缺少清屏操作
}
```

### 3. Panic 恢复时 `ExitAltScreen()` 不清屏

```go
func (a *App) ExitAltScreen() {
    fmt.Print("\x1b[?1049l")  // 退出备用屏幕
    // 但我们从未进入备用屏幕模式，此序列无效
}
```

### 4. Ctrl+C 信号处理器不清屏

`runtime/platform/input_windows.go` 和 `input_unix.go` 中的 `restoreTerminalImpl()` 只恢复终端模式，不清屏：

```go
func restoreTerminalImpl() {
    // 恢复控制台模式
    // 缺少清屏和显示光标
}
```

且调用 `os.Exit(0)` 会绕过所有 defer，导致 `App.Close()` 不会执行。`os.Exit(1)`非零退出码会让 Windows cmd/powershell 打印 exit status 1。

### 5. `examples/toast/main.go` Hooks 调用位置错误

```go
func main() {
    // 错误：在 main 函数中调用 hooks
    infoToast, setInfoToast, _ := ui.UseStateInt(0)

    ui.Run(func() ui.VNode {
        // ... 组件函数
    })
}
```

Hooks 必须在组件函数内部调用，因为框架需要通过调用顺序来跟踪 hook 状态。

## 修复方案

### 1. `framework/app.go` - Close() 方法

```go
func (a *App) Close() error {
    // 显示终端光标
    a.ShowCursor()

    // 清屏，避免退出时残留内容
    a.clearScreen()

    // 关闭 panic 恢复管理器
    if a.recovery != nil {
        a.recovery.Close()
    }
    return nil
}
```

### 2. `runtime/core/runtime.go` - Shutdown() 方法

```go
func (r *Runtime) Shutdown(timeout ...time.Duration) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.running = false

    // 清屏，避免退出时残留内容
    _ = r.platform.Clear()

    // 关闭平台
    if err := r.platform.Close(); err != nil {
        return err
    }

    return r.contextManager.Shutdown(timeout...)
}
```

### 3. `framework/app.go` - ExitAltScreen() 方法

```go
func (a *App) ExitAltScreen() {
    fmt.Print("\x1b[?1049l")
    // 由于我们未使用备用屏幕模式（只用 \x1b[2J 清屏），
    // panic 时需要主动清屏以避免 TUI 内容残留
    a.clearScreen()
}
```

### 4. `runtime/platform/input_windows.go` - restoreTerminalImpl()

```go
func restoreTerminalImpl() {
    // 先清屏和显示光标（在修改控制台模式之前，确保 ANSI 序列有效）
    fmt.Print("\x1b[2J")    // 清屏
    fmt.Print("\x1b[H")     // 移动光标到左上角
    fmt.Print("\x1b[?25h")  // 显示光标

    // 然后恢复控制台模式
    // ...
}
```

### 5. `runtime/platform/input_unix.go` - restoreTerminalImpl()

```go
func restoreTerminalImpl() {
    // 先清屏和显示光标（在修改终端模式之前）
    fmt.Print("\x1b[2J")    // 清屏
    fmt.Print("\x1b[H")     // 移动光标到左上角
    fmt.Print("\x1b[?25h")  // 显示光标

    // 然后恢复终端模式
    // ...
}
```

### 6. `examples/toast/main.go` - Hooks 调用位置修复

```go
func main() {
    ui.Run(func() ui.VNode {
        // 正确：在组件函数内部调用 hooks
        infoToast, setInfoToast, _ := ui.UseStateInt(0)
        // ... 组件渲染
    })
}
```

### 7. `runtime/platform/input_unix.go` - ioctl 类型修复

```go
// 修复前：fd 是 int 类型，但 ioctl 需要 uintptr
ioctl(fd, syscall.TCGETS, ...)  // 错误

// 修复后：转换为 uintptr
ioctl(uintptr(fd), syscall.TCGETS, ...)  // 正确
```

## 关键点

### 1. ANSI 序列顺序

必须在修改终端/控制台模式**之前**打印 ANSI 序列，否则序列可能不被解析：

```
1. 打印 ANSI 序列（清屏、显示光标）
2. 恢复终端/控制台模式
```

### 2. 退出路径覆盖

| 退出方式 | 调用路径 | 清屏位置 |
|---------|---------|---------|
| 正常退出 | `App.Run()` → `App.Close()` | `Close()` |
| Runtime 退出 | `Runtime.Shutdown()` | `Shutdown()` |
| Panic 退出 | `recovery.Handle()` → `ExitAltScreen()` | `ExitAltScreen()` |
| Ctrl+C | 信号处理器 → `restoreTerminalImpl()` | `restoreTerminalImpl()` |

### 3. Hooks 规则

Hooks（如 `UseStateInt`）必须在组件函数的顶层调用，且每次渲染时以相同的顺序调用。这是 React Hooks 规则，同样适用于此框架。

## 修改文件列表

| 文件 | 修改内容 |
|------|---------|
| `framework/app.go` | `Close()` 添加清屏 |
| `framework/app.go` | `ExitAltScreen()` 添加清屏 |
| `runtime/core/runtime.go` | `Shutdown()` 添加清屏 |
| `runtime/platform/input_windows.go` | `restoreTerminalImpl()` 添加清屏 |
| `runtime/platform/input_unix.go` | `restoreTerminalImpl()` 添加清屏 + ioctl 类型修复 |
| `runtime/platform/input_darwin.go` | **新增** Darwin/macOS 专用实现 |
| `examples/toast/main.go` | 修复 hooks 调用位置 |

## Darwin/macOS 平台支持

创建了 `runtime/platform/input_darwin.go` 文件，因为 Darwin 使用不同的 ioctl 常量：

| 平台 | 获取终端属性 | 设置终端属性 |
|------|-------------|-------------|
| Linux | `TCGETS` | `TCSETS` |
| Darwin | `TIOCGETA` | `TIOCSETA` |

通过构建标签分离平台实现：

| 文件 | 构建标签 |
|------|---------|
| `input_windows.go` | `//go:build windows` |
| `input_darwin.go` | `//go:build darwin` |
| `input_unix.go` | `//go:build (unix \|\| linux \|\| freebsd) && !darwin && !windows` |

## 测试验证

```bash
# 测试正常退出
go run ./examples/input
# 按 q 键退出 → 应该清屏

# 测试 Ctrl+C 退出
go run ./examples/input
# 按 Ctrl+C → 应该清屏

# 测试 toast 示例
go run ./examples/toast
# 应该正常运行，不再 panic
```

## 相关 ANSI 转义序列

| 序列 | 功能 |
|------|------|
| `\x1b[2J` | 清除整个屏幕 |
| `\x1b[H` | 移动光标到左上角 |
| `\x1b[?25h` | 显示光标 |
| `\x1b[?25l` | 隐藏光标 |
| `\x1b[?1049h` | 进入备用屏幕 |
| `\x1b[?1049l` | 退出备用屏幕 |

## 信号处理器警告消息

原始终端模式下，换行符可能无效：
- `\n` - 可能不工作
- `\r\n` - 也可能不工作
- ANSI 换行序列 - 也可能不工作

**解决方案**：直接移除警告消息。屏幕已清空，用户能看出程序已退出。

## 架构改进建议

### 1. 统一清理接口

考虑创建一个 `TerminalCleanup` 接口，统一处理所有退出路径的清理工作：

```go
type TerminalCleanup interface {
    Cleanup()
}
```

### 2. 备用屏幕模式

考虑使用真正的备用屏幕模式（`\x1b[?1049h`），退出时会自动恢复原屏幕内容：

```go
// 启动时
fmt.Print("\x1b[?1049h")  // 进入备用屏幕

// 退出时
fmt.Print("\x1b[?1049l")  // 退出备用屏幕，自动恢复
```

这样可以完全避免清屏操作，但需要确保所有平台都支持。

### 3. 信号处理改进

考虑使用 `defer` + 恢复机制，而非直接调用 `os.Exit`：

```go
func init() {
    go func() {
        ch := make(chan os.Signal, 1)
        signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
        <-ch
        restoreTerminalImpl()
        // 让主循环自然退出，而非 os.Exit
    }()
}
```

## 相关 Issue

- Toast 示例 panic: `useState must be called within a component`
- Darwin/macOS 平台构建错误: `cannot use fd (variable of type int) as uintptr value`
- Ctrl+C 退出后终端内容残留
