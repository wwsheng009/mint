# Mint UI 日志与错误处理完整指南

**版本**: v2.0
**创建日期**: 2026-03-04
**状态**: ✅ 已实现
**相关历史文档**: `../../../../docsArchive/cleanup-2026-05-19/docs/ui/store/optimization/SYSTEM_ANALYSIS_OPTIMIZATION.md`

---

## 目录

- [一、日志系统](#一日志系统)
  - [1.1 配置 Logger](#11-配置-logger)
  - [1.2 启用调试日志](#12-启用调试日志)
  - [1.3 日志级别控制](#13-日志级别控制)
- [二、错误处理策略](#二错误处理策略)
  - [2.1 错误处理策略类型](#21-错误处理策略类型)
  - [2.2 配置错误处理](#22-配置错误处理)
  - [2.3 自定义错误处理器](#23-自定义错误处理器)
- [三、完整示例](#三完整示例)
- [四、最佳实践](#四最佳实践)

---

## 一、日志系统

### 1.1 配置 Logger

Mint UI 使用 `internal/log.Logger` 作为统一的日志系统。Dispatcher 和 Registry 支持设置自定义 Logger。

#### Registry Logger

```go
import (
    "github.com/wwsheng009/mint/runtime/intent"
    mintlog "github.com/wwsheng009/mint/internal/log"
)

func registerHandlers() {
    // 创建自定义 Logger
    logger := mintlog.NewLogger("Registry", "REGISTRY")
    logger.Enabled(true)  // 启用日志

    // 设置到 Registry
    intent.DefaultRegistry().SetLogger(logger)

    // 现在注册、覆盖、取消注册都会被记录
    reducerBuilder.RegisterToGlobal(appStore)
}
```

#### Dispatcher Logger

```go
func configureDispatcher() *intent.Dispatcher {
    registry := intent.DefaultRegistry()
    dispatcher := intent.NewDispatcher(registry)

    // 创建 Logger
    logger := mintlog.NewLogger("IntentDispatcher", "INTENT")
    logger.SetEnabled(true)

    // 设置到 Dispatcher
    dispatcher.SetLogger(logger)

    // 或者使用内部 Intent.Logger (如果框架已配置)
    // dispatcher.SetLogger(internal.IntentLogger)

    return dispatcher
}
```

---

### 1.2 启用调试日志

Mint UI 支持通过环境变量启用调试日志：

```bash
# 启用所有调试日志
export TUI_DEBUG_ALL=true

# 启用特定类别的日志
export TUI_DEBUG_INTENT=true   # Intent 相关
export TUI_DEBUG_ACTION=true   # Action 相关
export TUI_DEBUG_FOCUS=true    # Focus 相关

# 运行程序
go run ./examples/store_reducer_demo/
```

#### 启用日志输出到文件

```go
import "github.com/wwsheng009/mint/internal/log"

func main() {
    // 输出到文件
    log.SetLogFile("./debug.log", log.SizeBasedRotation, 10*1024*1024) // 10MB

    // 输出到文件和控制台
    log.SetLogOutput(log.OutputBoth)

    ui.Run(App, ...)
}
```

---

### 1.3 日志级别控制

Logger 支持不同级别的输出：

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| `Debug` | 详细的调试信息 | Intent dispatch、Handler 覆盖 |
| `Info` | 一般信息 | 组件挂载、State 更新 |
| `Warn` | 警告信息 | Handler 覆盖被保护 |
| `Error` | 错误信息 | Intent 处理失败 |

#### 自定义启用逻辑

```go
logger := mintlog.NewLogger("MyApp", "APP")

// 运行时启用日志
logger.SetEnabled(os.Getenv("DEBUG") == "1")

// 或根据功能模块启用
if os.Getenv("DEBUG_INTENTS") == "1" {
    mintlog.IntentLogger.SetEnabled(true)
}
```

---

## 二、错误处理策略

### 2.1 错误处理策略类型

| 策略 | 说明 | 适用场景 |
|------|------|---------|
| `ErrorLogIgnore` | 记录错误，忽略 | 开发调试 |
| `ErrorLogPanic` | 记录错误，panic | 严格模式、测试 |
| `ErrorCustomCallback` | 调用自定义处理器 | 生产环境 |
| `ErrorLogRetry` | 重试（待实现） | 网络请求等 |

---

### 2.2 配置错误处理

```go
import "github.com/wwsheng009/mint/runtime/intent"

func configureDispatcher() *intent.Dispatcher {
    registry := intent.DefaultRegistry()
    dispatcher := intent.NewDispatcher(registry)

    // 设置错误处理策略
    dispatcher.SetErrorStrategy(intent.ErrorLogIgnore)  // 默认

    // 或使用严格模式
    dispatcher.SetErrorStrategy(intent.ErrorLogPanic)

    // 或使用自定义回调
    dispatcher.SetErrorStrategy(intent.ErrorCustomCallback)
    dispatcher.SetErrorHandler(handleIntentError)

    return dispatcher
}

func handleIntentError(intent intent.Intent, err error) {
    // 发送到错误追踪服务
    sentry.CaptureException(err)

    // 显示用户友好的错误提示
    showErrorModal("操作失败，请重试")
}
```

---

### 2.3 自定义错误处理器

#### 基础示例

```go
type ErrorCounter struct {
    mu      sync.Mutex
    errors  map[string]int
}

func (ec *ErrorCounter) Handle(intent Intent, err error) {
    ec.mu.Lock()
    defer ec.mu.Unlock()

    intentType := intent.IntentType()
    ec.errors[intent_type]++

    // 记录到监控系统
    metrics.Increment("intent.error", map[string]string{
        "type": intentType,
        "error": err.Error(),
    })
}
```

#### 与 Sentry 集成

```go
import "github.com/getsentry/sentry-go"

func initSentryErrorHandler(dispatcher *intent.Dispatcher) {
    dispatcher.SetErrorStrategy(intent.ErrorCustomCallback)
    dispatcher.SetErrorHandler(func(intent intent.Intent, err error) {
        sentry.WithScope(func(scope *sentry.Scope) {
            scope.SetTag("intent_type", intent.IntentType())
            scope.SetContext("intent", map[string]interface{}{
                "type": intent.IntentType(),
            })
            sentry.CaptureException(err)
        })
    })
}
```

---

## 三、完整示例

### 示例 1: Store + Reducer Demo 完整配置

```go
package main

import (
    "os"

    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
)

var appStore *store.Store[AppState]

func main() {
    // 1. 配置日志
    setupLogging()

    // 2. 初始化 Store
    initStore()

    // 3. 运行应用
    err := ui.Run(App,
        ui.WithWidth(60),
        ui.WithHeight(30),
        ui.WithTitle("Store + Reducer Demo"),
        ui.WithInit(registerHandlers),  // 在内置 handlers 之后注册
    )
    if err != nil {
        panic(err)
    }
}

func setupLogging() {
    // 启用 Intent 相关日志
    if os.Getenv("DEBUG") == "1" {
        intent.IntentLogger.SetEnabled(true)
        intent.ActionLogger.SetEnabled(true)
    }

    // 设置自定义日志输出到文件
    log.SetLogFile("./intents.log", log.SizeBasedRotation, 10*1024*1024)
    log.SetLogOutput(log.OutputBoth)  // 同时输出到文件和控制台
}

func initStore() {
    appStore = store.NewStore(AppState{
        Count:      0,
        Username:   "",
        Email:      "",
        ActiveTab:  0,
    })
}

func registerHandlers() {
    // 创建 Logger
    logger := intent.NewLogger("Registry", "REGISTRY")
    logger.SetEnabled(os.Getenv("DEBUG") == "1")
    intent.DefaultRegistry().SetLogger(logger)

    // 注册 handlers（会记录日志）
    reducerBuilder.RegisterToGlobal(appStore)
}

// ================================
// Reducer 定义
// ================================

var reducerBuilder = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count--
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        // ... 字段变更处理
        return s
    }).
    Build()

// ================================
// 视图函数
// ================================

func App() ui.VNode {
    state := appStore.Get()
    return ui.VStack(
        // ... UI 组件
    )
}
```

---

### 示例 2: 生产环境配置

```go
package main

import (
    "os"

    "github.com/getsentry/sentry-go"
    "github.com/wwsheng009/mint/runtime/intent"
    mintlog "github.com/wwsheng009/mint/internal/log"
    "github.com/wwsheng009/mint/ui"
)

func main() {
    // 初始化 Sentry
    initSentry()

    // 配置日志
    initLogging()

    // 配置 Dispatcher 错误处理
    initDispatcher()

    ui.Run(App, ...)
}

func initSentry() {
    sentry.Init(sentry.ClientOptions{
        Dsn: os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENVIRONMENT"),
        // ...
    })
}

func initLogging() {
    // 输出到文件
    mintlog.SetLogFile("./app.log", mintlog.SizeBasedRotation, 10*1024*1024)
    mintlog.SetLogOutput(mintlog.OutputFile)

    // 生产环境不启用 DEBUG 日志
    mintlog.IntentLogger.SetEnabled(false)
}

func initDispatcher() {
    dispatcher := intent.DefaultDispatcher()

    // 配置错误处理策略
    dispatcher.SetErrorStrategy(intent.ErrorCustomCallback)
    dispatcher.SetErrorHandler(func(intent intent.Intent, err error) {
        // 发送到 Sentry
        sentry.WithScope(func(scope *sentry.Scope) {
            scope.SetTag("component", "intent_dispatcher")
            scope.SetContext("intent", map[string]interface{}{
                "type": intent.IntentType(),
            })
            sentry.CaptureException(err)
        })

        // 记录到日志
        mintlog.IntentLogger.Error("Intent failed: type=%s, error=%v",
            intent.IntentType(), err)
    })

    // 设置 Logger
    logger := mintlog.NewLogger("Production", "PROD")
    dispatcher.SetLogger(logger)
}
```

---

## 四、最佳实践

### 4.1 日志配置

✅ **推荐做法**:

```go
// 根据环境配置
func setupLogger(env string) *log.Logger {
    logger := log.NewLogger("App", "APP")

    switch env {
    case "production":
        logger.SetEnabled(false)  // 关闭日志（或仅 Error）
    case "staging":
        logger.SetEnabled(true)
        log.SetLogFile("./staging.log", log.SizeBasedRotation, 10*1024*1024)
        log.SetLogOutput(log.OutputFile)
    case "development":
        logger.SetEnabled(true)
        log.SetLogOutput(log.OutputBoth)  // 控制台 + 文件
    }

    return logger
}
```

❌ **不推荐**:

```go
// 始终启用所有日志
mintlog.IntentLogger.SetEnabled(true)
mintlog.FocusLogger.SetEnabled(true)
// ...
// 性能开销大，生产环境会污染日志
```

---

### 4.2 错误处理策略

✅ **推荐做法**:

| 环境 | 策略 | 原因 |
|------|------|------|
| Development | `ErrorLogIgnore` | 快速迭代，查看日志即可 |
| Production | `ErrorCustomCallback` | 上报错误，不影响用户体验 |
| Production (Critical) | `ErrorLogPanic` | 关键错误立即暴露 |

```go
func getErrorStrategy(env string) intent.ErrorHandlingStrategy {
    switch env {
    case "production":
        return intent.ErrorCustomCallback
    case "production_critical":
        return intent.ErrorLogPanic
    default:
        return intent.ErrorLogIgnore
    }
}
```

---

### 4.3 监控和分析

```go
type Metrics struct {
    intentsDispatched  int64
    intentsFailed      int64
    handlerOverridden  int64
}

func (m *Metrics) CaptureIntentDispatch(intentType string) {
    atomic.AddInt64(&m.intentsDispatched, 1)
}

func (m *Metrics) CaptureIntentError(intentType string, err error) {
    atomic.AddInt64(&m.intentsFailed, 1)
    // 发送到监控系统
    metrics.CounterIncrement("intent.error", map[string]string{"type": intentType})
}
```

---

### 4.4 Handler 注册最佳实践

```go
// ✅ 好的做法：清晰的配置和日志
func registerHandlers() {
    logger := log.NewLogger("Registry", "REGISTRY")
    logger.SetEnabled(os.Getenv("DEBUG") == "1")
    intent.DefaultRegistry().SetLogger(logger)

    logger.Info("Registering app handlers")
    reducerBuilder.RegisterToGlobal(appStore)
    logger.Info("All handlers registered successfully")
}

// ❌ 不推荐：无日志，难以调试
func registerHandlers() {
    reducerBuilder.RegisterToGlobal(appStore)
    // 如果出错，不知道是哪个 handler 失败
}
```

---

## 附录

### A. 相关文档

- 系统分析与优化历史记录：`../../../../docsArchive/cleanup-2026-05-19/docs/ui/store/optimization/SYSTEM_ANALYSIS_OPTIMIZATION.md`
- [IMPLEMENTATION_REVIEW.md](/docsArchive/reviews/IMPLEMENTATION_REVIEW.md) - 实施复查报告
- [STORE_STORE_REDUCER_GUIDE.md](../guides/STORE_REDUCER_GUIDE.md) - Store + Reducer 指南

### B. API 参考

#### Registry 方法

| 方法 | 说明 |
|------|------|
| `SetLogger(*Logger)` | 设置 Registry Logger |
| `Register(string, Handler, ...Option)` | 注册 handler（支持选项） |

#### Dispatcher 方法

| 方法 | 说明 |
|------|------|
| `SetLogger(*Logger)` | 设置 Dispatcher Logger |
| `SetErrorStrategy(ErrorHandlingStrategy)` | 设置错误处理策略 |
| `SetErrorHandler(func(Intent, error))` | 设置自定义错误处理器 |

#### 错误处理策略

```go
const (
    ErrorLogIgnore         // 记录并忽略
    ErrorLogPanic          // 记录并 panic
    ErrorLogRetry          // 记录并重试（待实现）
    ErrorCustomCallback    // 自定义处理器
)
```

---

**文档版本**: v2.0
**最后更新**: 2026-03-04
**维护者**: Mint UI Team
