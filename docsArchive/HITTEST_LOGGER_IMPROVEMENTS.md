# HitTest 调试系统改进

## 已完成的更改

### 1. 创建了文件日志系统
**文件**: `internal/logger/logger.go`

创建了完整的日志系统，替代了之前使用 `fmt.Fprintf(os.Stderr, ...)` 的调试方式。

**特性**:
- 基于文件的日志记录
- 日志级别: DEBUG, INFO, WARN, ERROR
- 按Category分类（MOUSE, HITMAP, LAYER等）
- 通过环境变量启用: `TUI_DEBUG_LOG=path/to/log.log`
- 线程安全

**使用方法**:
```go
import "github.com/wwsheng009/mint/internal/logger"

// 初始化（通常在App.Init()中）
log := logger.InitFromEnv()

// 记录日志
log.Debug("CATEGORY", "message with %s", "args")
log.Info("CATEGORY", "info message")
log.Warn("CATEGORY", "warning message")
log.Error("CATEGORY", "error message")
```

### 2. 更新了 Pump 中的鼠标调试
**文件**: `framework/event/pump.go`

**之前**:
```go
if os.Getenv("TUI_DEBUG_MOUSE_POSITION") == "true" {
    fmt.Fprintf(os.Stderr, "[MOUSE] Raw position: (%d, %d)...\n", x, y)
}
```

**现在**:
```go
log := logger.Get()
log.Debug("MOUSE", "Raw position: (%d, %d) | Action: %v", x, y, action)
```

**新增功能**: 检测多个按钮重叠的情况
```go
// Also log all entries at this position for debugging overlapping buttons
allEntries := hitMap.FindAllAt(raw.MouseX, raw.MouseY)
if len(allEntries) > 1 {
    log.Debug("MOUSE", "Multiple hits at (%d,%d):", raw.MouseX, raw.MouseY)
    for i, e := range allEntries {
        log.Debug("MOUSE", "  [%d] ID='%s' Bounds=(%d,%d,%dx%d) ZOrder=%d",
            i, e.NodeID, e.Bounds.X, e.Bounds.Y, e.Bounds.Width, e.Bounds.Height, e.ZOrder)
    }
}
```

### 3. 更新了 Layer Manager 的调试
**文件**: `runtime/layer/manager.go`

**centerModal 函数**:
- 记录modal大小、container大小、计算的偏移量
- 记录shift后的最终位置

**buildHitMapFromComputedBox 函数**:
- 记录每个HitMap entry的位置
- 记录最终HitMap的条目数量

**GetMergedHitMap 函数**:
- 记录每层的HitMap大小
- 记录modal layer的entry详情
- 记录合并后的HitMap信息

### 4. 更新了 App 初始化
**文件**: `framework/app.go`

在 `Init()` 函数中添加了logger初始化:
```go
// Initialize logger from environment
log := logger.InitFromEnv()
if log != nil && log.IsEnabled() {
    log.Info("APP", "Init: Starting initialization")
}
```

### 5. 更新了 Demo
**文件**: `examples/ui_demos/demo1_full_featured/main.go`

**main 函数**:
```go
// Enable file-based logging
os.Setenv("TUI_DEBUG_LOG", "demo1_debug.log")

// Auto-open modal for testing
os.Setenv("AUTO_OPEN_MODAL", "true")
```

**DebugPanel**:
- 更新了说明文字
- 添加了关于multiple button overlap detection的信息
- 说明日志文件位置

## 如何使用

### 运行Demo并收集调试日志

```bash
cd examples/ui_demos/demo1_full_featured

# 方法1: 直接运行（已设置AUTO_OPEN_MODAL）
./demo1_new.exe

# 方法2: 手动设置环境变量
TUI_DEBUG_LOG=modal_debug.log AUTO_OPEN_MODAL=true ./demo1_new.exe
```

### 查看日志

```bash
# 实时查看日志
tail -f demo1_debug.log

# 查看鼠标相关的日志
grep "MOUSE" demo1_debug.log

# 查看HitMap相关的日志
grep "HITMAP" demo1_debug.log

# 查看Modal centering的日志
grep "centerModal" demo1_debug.log
```

### 日志格式

```
[HH:MM:SS.mmm] [LEVEL] [CATEGORY] message
```

示例:
```
[12:34:56.789] [DEBUG] [MOUSE] Raw position: (25, 11) | Action: 0
[12:34:56.789] [DEBUG] [MOUSE] HitTest: Found 'button' at Bounds=(25,11,15x1) Local=(0,0)
[12:34:56.790] [INFO] [LAYER] [centerModal] modal=(0,0) size=40x9 container=80x24 offset=(20,7)
```

## 问题诊断流程

### 问题1: Modal按钮点击位置不正确

**日志检查清单**:
1. 检查 `[LAYER] [centerModal]` 日志，确认:
   - Container大小是否正确（应该是buffer大小，不是实际终端大小）
   - Offset计算是否正确

2. 检查 `[HITMAP] [buildHitMapFromComputedBox]` 日志，确认:
   - Button的Bounds是否正确
   - 是否所有button都被记录

3. 检查 `[MOUSE]` 日志，确认:
   - 鼠标点击的原始坐标
   - HitTest是否找到正确的button
   - 是否有多个按钮重叠（看 "Multiple hits" 日志）

### 问题2: 多行按钮点击错位

**关键日志**:
```
[MOUSE] Multiple hits at (25, 15):
  [0] ID='button' Bounds=(25,10,15x1) ZOrder=2
  [1] ID='button' Bounds=(25,14,15x1) ZOrder=3
```

如果看到这样的日志，说明多个button有重叠的Bounds，导致点击总是命中ZOrder最高的那个。

**可能原因**:
- Button的Height设置太小（只有1）
- VStack中button之间的间距计算错误
- Button的Y坐标没有正确累积

## 下一步排查

请运行demo并提供日志文件的输出:
```bash
./demo1_new.exe > /dev/null 2>&1
# 按ESC关闭
cat demo1_debug.log
```

特别关注以下信息:
1. Container大小（centerModal日志）
2. Button的Bounds坐标
3. 鼠标点击时的坐标和HitTest结果
