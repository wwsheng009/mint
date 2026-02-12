# 信号处理与 Ctrl+C 退出

## 概述

Mint 实现了多层防御的信号处理机制，确保终端在任何情况下都能正确恢复。

## Ctrl+C 退出机制

### Windows 平台

Windows 平台通过 `init()` 函数注册信号监听：

```go
// runtime/platform/input_windows.go
func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		_ = <-ch
		restoreTerminalImpl()
		os.Exit(0)
	}()
}
```

### Linux 平台

Linux 平台使用相同的机制：

```go
// runtime/platform/input_linux_tcell.go
func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		_ = <-ch
		restoreTerminalImpl()
		os.Exit(0)
	}()
}
```

### Darwin (macOS) 平台

Darwin 平台使用相同的机制：

```go
// runtime/platform/input_darwin_tcell.go
func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		_ = <-ch
		restoreTerminalImpl()
		os.Exit(0)
	}()
}
```

## 多层防御机制

终端恢复采用三层防御系统：

### 第 1 层：Engine 层
- `Engine.Run()` 的 `defer cleanup()`
- 调用 `inputReader.Stop()` → 恢复 originalMode

### 第 2 层：应用层
- `main()` 的 `defer platform.RestoreTerminal()`
- 强制恢复到安全模式

### 第 3 层：进程层
- `init()` 的信号处理
- Ctrl+C 时强制恢复（兜底）

## 信号类型

| 信号 | 名称 | 触发方式 |
|------|------|----------|
| SIGINT | 中断信号 | Ctrl+C |
| SIGTERM | 终止信号 | kill 命令 |

## 终端恢复函数

每个平台实现自己的 `restoreTerminalImpl()`：

- **Windows**: 调用 Windows API 设置默认控制台模式
- **Linux**: 调用 tcell `screen.Fini()`
- **Darwin**: 调用 tcell `screen.Fini()`

## 为什么需要 init() 中的信号处理？

1. **Panic 保护**: 如果程序 panic，defer 可能不执行
2. **强制关闭**: 用户强制关闭（Ctrl+C）时确保清理
3. **终端污染**: 一次终端污染可能导致 shell 永久损坏

## 测试

```bash
# 启动程序，然后按 Ctrl+C
go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/

# 终端应该正确恢复，不会出现显示异常
```

## 相关文件

- `runtime/platform/input_windows.go` - Windows 信号处理
- `runtime/platform/input_linux_tcell.go` - Linux 信号处理
- `runtime/platform/input_darwin_tcell.go` - Darwin 信号处理
- `runtime/platform/signal.go` - 通用信号处理器
- `runtime/platform/platform.go` - 平台抽象
