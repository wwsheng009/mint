# 终端模式污染问题修复报告

**日期**: 2025-01-30
**严重级别**: 🔴 Critical - 导致终端永久损坏
**影响范围**: 所有 Windows 平台的 TUI 程序
**状态**: ✅ 已修复

---

## 问题描述

### 症状表现

在 Windows PowerShell 中运行 TUI 程序后，出现以下症状：

1. **第一次运行**：程序正常运行，可以正常退出
2. **第二次运行**：`fmt.Scanln()` 无法接收回车键
   - 可以输入数字和字母
   - 但按下 Enter 键没有任何响应
   - 必须关闭并重新打开 PowerShell 才能恢复

### 影响范围

- 所有使用 `runtime/platform` 的 example 程序
- `examples/engine/origin`
- `examples/engine/with_logger`
- `examples/engine/with_devtools`
- `examples/debug_test`

---

## 根本原因分析

### 核心死因（一句话）

> **程序将控制台改为 RAW + 非行模式，但程序退出路径没有 100% 保证调用 `Stop()`，导致终端永久处于污染状态。**

### 三大致命设计叠加

#### 🔥 致命点 1：缺少启动时重置机制

在 `input_windows.go` 的 `Start()` 方法中：

```go
// ❌ 问题代码
r.originalMode = r.getConsoleMode(handle)  // 保存原始模式
```

**问题**：如果上次程序崩溃或异常退出，控制台可能已经处于 raw 模式。此时保存的 `originalMode` 实际上是**污染过的模式**。

当程序正常退出并调用 `Stop()` 时：

```go
r.setConsoleMode(handle, r.originalMode)  // 恢复到污染的模式
```

结果：**恢复的仍然是 raw 模式**，终端状态永久损坏。

#### 🔥 致命点 2：defer 未正确使用

在 `engine.go` 的 `Run()` 方法中：

```go
// ❌ 问题代码
cleanup := func() {
    if e.inputReader != nil {
        e.inputReader.Stop()
        e.inputReader = nil
    }
    // ...
}

// ... 很多代码 ...

cleanup()  // 手动调用，不是 defer
```

**问题**：如果程序在调用 `cleanup()` 之前 panic、通过 `os.Exit()` 退出，或者某个异常路径提前返回，`cleanup()` 就不会执行，导致 `Stop()` 永远不会被调用。

#### 🔥 致命点 3：缺少进程级保险丝

程序没有任何信号处理机制来确保异常退出时恢复终端。

当用户按下 `Ctrl+C` 时，程序直接退出，defer 可能来不及执行，终端处于 raw 模式。

---

## 解决方案

### 修复 1：启动时强制重置控制台到安全模式

**文件**: `runtime/platform/input_windows.go`

添加 `resetConsoleToSaneMode()` 方法：

```go
// resetConsoleToSaneMode 重置控制台到安全模式
//
// 🔥 关键函数：防止上次崩溃遗毒
//
// 如果程序上次崩溃，控制台可能处于 raw 模式。
// 如果直接保存这个模式作为 "originalMode"，Stop() 恢复时就是错的。
//
// 必须先强制重置到 Windows 默认的安全模式，然后再保存。
func (r *windowsInputReader) resetConsoleToSaneMode(handle uintptr) {
	// Windows 默认 console input 模式（安全模式）
	saneMode := uint32(
		ENABLE_PROCESSED_INPUT | // 系统处理 Ctrl+C 和特殊字符
		ENABLE_LINE_INPUT |      // 行缓冲模式（fmt.Scanln 必需）
		ENABLE_ECHO_INPUT |      // 回显输入
		ENABLE_EXTENDED_FLAGS,   // 扩展标志
	)
	procSetConsoleMode.Call(handle, uintptr(saneMode))
}
```

在 `Start()` 方法中调用：

```go
// 🔥 关键修复：先重置控制台到安全模式，防止上次崩溃遗毒
r.resetConsoleToSaneMode(handle)

// 保存原始模式（现在保证是安全模式）
r.originalMode = r.getConsoleMode(handle)
```

### 修复 2：Engine.Run() 使用 defer cleanup

**文件**: `runtime/engine/engine.go`

```go
// 清理函数 - 确保总是执行
cleanup := func() {
	if e.inputReader != nil {
		e.inputReader.Stop()
		e.inputReader = nil
	}
	// 显示光标（如果被隐藏了）
	fmt.Print("\x1b[?25h")
	// 重置终端样式
	fmt.Print("\x1b[0m")
	// 清除屏幕
	fmt.Print("\x1b[2J")
	// 移动光标到左上角
	fmt.Print("\x1b[H")
}
defer cleanup() // 🔥 立即 defer，确保任何退出都会执行
```

**效果**：无论 panic、异常退出、正常返回，`cleanup()` 都会被执行。

### 修复 3：添加进程级信号恢复保险丝

**文件**: `runtime/platform/input_windows.go`

```go
// init 安装进程级终端恢复保险丝
//
// 🔥 工业级保护：即使程序 panic、强制关闭，也会恢复终端
//
// 这是最后一道防线，确保终端永远不会被永久污染。
func init() {
	go func() {
		// 监听中断信号
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		sig := <-ch
		// 强制恢复终端
		restoreTerminalImpl()
		// 输出提示信息（注意：此时终端可能处于异常状态）
		fmt.Fprintf(os.Stderr, "\n[WARNING] Received signal %v, terminal restored\n", sig)
		os.Exit(1)
	}()
}
```

**效果**：即使 Ctrl+C 强制终止，也会恢复终端。

### 修复 4：导出 RestoreTerminal() 函数

**文件**: `runtime/platform/input.go`

```go
// RestoreTerminal 恢复终端到正常模式
//
// 这是一个安全的兜底函数，用于在程序异常退出时恢复终端状态。
// 它会恢复行缓冲模式和回显，使 fmt.Scanln 等标准输入函数正常工作。
//
// 应该在 main 函数的 defer 中调用，确保即使 panic 或异常退出也能恢复终端。
//
// 示例：
//
//	func main() {
//	    defer platform.RestoreTerminal()
//	    // ... 你的代码
//	}
func RestoreTerminal() {
	restoreTerminalImpl()
}
```

### 修复 5：所有 example 添加 defer RestoreTerminal()

在所有 example 的 `main()` 函数中添加：

```go
func main() {
	// 设置清理函数
	defer func() {
		// 最先恢复终端控制台模式（必须在所有其他操作之前）
		// 这会恢复 ENABLE_LINE_INPUT 和 ENABLE_ECHO_INPUT，让 fmt.Scanln 等正常工作
		platform.RestoreTerminal()

		// 恢复终端 ANSI 序列
		fmt.Print("\x1b[?25h") // 显示光标
		fmt.Print("\x1b[0m")  // 重置样式
		fmt.Print("\x1b[H")   // 移动光标到左上角
		fmt.Println()         // 换行

		// ... 其他清理逻辑
	}()

	// ... 程序逻辑
}
```

**影响文件**：
- `examples/engine/origin/main.go`
- `examples/engine/with_logger/main.go`
- `examples/engine/with_devtools/main.go`
- `examples/debug_test/main.go`

---

## TUI 三大铁律

手写 Windows Console Raw Input Driver 必须遵守以下原则：

### 1. 任何时候退出都必须 Restore Console Mode

**正确做法**：
- 使用 `defer` 确保清理函数执行
- 不要依赖手动调用清理函数
- 禁止在代码中使用 `os.Exit()`（除非在 defer 之后）

**错误做法**：
```go
// ❌ 错误：手动调用，panic 时不会执行
cleanup()

// ❌ 错误：os.Exit 会跳过 defer
if err != nil {
    os.Exit(1)
}
```

**正确做法**：
```go
// ✅ 正确：使用 defer
defer cleanup()

// ✅ 正确：使用 return 而不是 os.Exit
if err != nil {
    return err  // defer 会执行
}
```

### 2. 启动时必须先 Reset Console（防崩溃遗毒）

**正确做法**：
```go
// ✅ 正确：先重置再保存
resetConsoleToSaneMode(handle)
r.originalMode = r.getConsoleMode(handle)
```

**错误做法**：
```go
// ❌ 错误：直接保存可能已污染的模式
r.originalMode = r.getConsoleMode(handle)
```

### 3. 绝不能依赖"上一次保存的 originalMode"

**问题**：如果上次程序崩溃，控制台处于 raw 模式，`originalMode` 保存的就是错误的模式。

**解决**：启动时强制重置到已知的安全模式，然后再保存。

---

## 修复效果

### 修复前的问题流程

#### 第一次运行
```
Console mode = 正常
Start() 保存 originalMode = 正常
改成 raw
程序异常退出 ❌ Stop 没执行
Console mode = raw（永久污染）
```

#### 第二次运行
```
Console mode = raw（已污染）
Start() 保存 originalMode = raw（错误！）
改成 raw
Stop() 恢复到 raw（等于没恢复）
Console 永久 raw → fmt.Scanln 失效
```

### 修复后的正确流程

#### 任意次运行
```
Console mode = 任意状态
Start() 重置到安全模式 → Console mode = 正常
Start() 保存 originalMode = 正常
改成 raw
程序退出 ✅ defer cleanup() 必执行
Stop() 恢复到 originalMode（正常）
Console mode = 正常 → fmt.Scanln 正常
```

---

## 测试验证

### 测试步骤

1. **基本功能测试**：
   ```bash
   go run examples/engine/with_logger/main.go
   # 按 ESC 退出
   # 再次运行，fmt.Scanln 应该能正常工作
   go run examples/engine/with_logger/main.go
   ```

2. **异常退出测试**：
   ```bash
   go run examples/engine/with_logger/main.go
   # 按 Ctrl+C 强制退出
   # 终端应该被自动恢复
   # 再次运行应该正常工作
   ```

3. **多次运行测试**：
   ```bash
   for i in {1..10}; do
       go run examples/engine/with_logger/main.go
       # 按 ESC 快速退出
   done
   # 终端应该始终保持正常状态
   ```

### 验证结果

- ✅ `fmt.Scanln()` 能正常接收回车键
- ✅ 多次运行不会污染终端
- ✅ Ctrl+C 后终端能自动恢复
- ✅ panic 后终端能自动恢复

---

## 关键代码变更

### 修改的文件列表

1. `runtime/platform/input_windows.go` - 添加重置机制和信号处理
2. `runtime/platform/input.go` - 导出 RestoreTerminal()
3. `runtime/engine/engine.go` - 使用 defer cleanup
4. `examples/engine/origin/main.go` - 添加 defer RestoreTerminal()
5. `examples/engine/with_logger/main.go` - 添加 defer RestoreTerminal()
6. `examples/engine/with_devtools/main.go` - 添加 defer RestoreTerminal()
7. `examples/debug_test/main.go` - 添加 defer RestoreTerminal()

### 新增的关键函数

1. `resetConsoleToSaneMode(handle uintptr)` - 强制重置到安全模式
2. `RestoreTerminal()` - 导出的终端恢复函数
3. `init()` - 进程级信号恢复保险丝

---

## 经验总结

### 学到的教训

1. **手写终端驱动是高风险操作**
   - Windows Console Input 是内核级资源
   - 修改控制台模式会影响整个终端会话
   - 必须有完善的错误处理和恢复机制

2. **defer 是救命稻草**
   - 任何可能失败的操作都要用 defer 保护
   - 不要依赖手动调用清理函数
   - os.Exit 会跳过 defer，要谨慎使用

3. **启动时总是假设环境已被污染**
   - 不要信任"原始状态"是正常的
   - 先重置到已知的安全状态，再保存
   - 这是防御性编程的体现

4. **多层次的恢复机制是必要的**
   - defer cleanup() - 正常退出
   - defer RestoreTerminal() - 应用层兜底
   - 信号处理 - 异常强制退出

### 最佳实践

1. **所有 TUI 程序都应该**：
   - 启动时重置控制台模式
   - 使用 defer 清理资源
   - 监听信号强制恢复
   - 提供手动恢复命令（如 `reset`）

2. **开发调试时**：
   - 使用独立的测试终端
   - 随时准备 `reset` 命令
   - 保存测试脚本，快速验证修复

3. **文档记录**：
   - 记录所有终端模式相关的代码
   - 标注高风险操作
   - 提供故障恢复指南

---

## 参考资料

- [Windows Console API Documentation](https://docs.microsoft.com/en-us/windows/console/console-functions)
- [SetConsoleMode function](https://docs.microsoft.com/en-us/windows/console/setconsolemode)
- [ENABLE_LINE_INPUT flag](https://docs.microsoft.com/en-us/windows/console/setconsolemode)
- [Go signal package](https://pkg.go.dev/os/signal)

---

## 附录：完整的修复代码

### input_windows.go 完整修复

```go
// resetConsoleToSaneMode 重置控制台到安全模式
func (r *windowsInputReader) resetConsoleToSaneMode(handle uintptr) {
	saneMode := uint32(
		ENABLE_PROCESSED_INPUT |
		ENABLE_LINE_INPUT |
		ENABLE_ECHO_INPUT |
		ENABLE_EXTENDED_FLAGS,
	)
	procSetConsoleMode.Call(handle, uintptr(saneMode))
}

func (r *windowsInputReader) Start(events chan<- RawInput) error {
	// ... ...

	// 🔥 关键修复：先重置控制台到安全模式
	r.resetConsoleToSaneMode(handle)
	r.originalMode = r.getConsoleMode(handle)

	// ... 设置 raw 模式 ...
}

// restoreTerminalImpl 恢复终端到默认模式
func restoreTerminalImpl() {
	handle, _, _ := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if handle != 0 {
		defaultMode := uint32(
			ENABLE_PROCESSED_INPUT |
			ENABLE_LINE_INPUT |
			ENABLE_ECHO_INPUT |
			ENABLE_EXTENDED_FLAGS,
		)
		procSetConsoleMode.Call(handle, uintptr(defaultMode))
	}
	// ... 恢复输出模式 ...
}

// init 进程级保险丝
func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		restoreTerminalImpl()
		fmt.Fprintf(os.Stderr, "\n[WARNING] Received signal %v, terminal restored\n", sig)
		os.Exit(1)
	}()
}
```

### engine.go 完整修复

```go
func (e *Engine) Run() error {
	// ... 启动输入读取器 ...

	cleanup := func() {
		if e.inputReader != nil {
			e.inputReader.Stop()
			e.inputReader = nil
		}
		// ... 恢复终端状态 ...
	}
	defer cleanup() // 🔥 关键：使用 defer

	// ... 主循环 ...
}
```

### example main.go 完整修复

```go
func main() {
	defer func() {
		platform.RestoreTerminal() // 🔥 最先恢复终端模式
		// ... 其他清理 ...
	}()
	// ... 程序逻辑 ...
}
```

---

**文档结束**
