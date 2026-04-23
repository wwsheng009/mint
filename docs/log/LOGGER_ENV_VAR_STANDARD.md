# Logger 环境变量使用统一规范

## 概述

为了统一项目中的日志调试环境变量使用方法，已将所有 `os.Getenv("TUI_*DEBUG")` 调用替换为使用标准化的 Logger 系统来检查。

## 环境变量命名规范

所有环境变量统一遵循以下命名规范：

1. **标准格式**: `TUI_DEBUG_<CATEGORY>` （如 `TUI_DEBUG_FOCUS`, `TUI_DEBUG_RENDER`）
2. **兼容格式**: `TUI_<MODULE>_DEBUG` （如 `TUI_WRAP_DEBUG`, `TUI_CURSOR_DEBUG`, `TUI_PAINT_DEBUG`）
3. **全局开关**: `TUI_DEBUG` - 启用所有调试日志

## 支持的值

所有环境变量支持以下值（不区分大小写）：
- `"true"`
- `"1"`
- `"yes"`
- `"on"`

## 已添加的 Logger

以下 Logger 已经添加到 `internal/log/logger.go`：

```go
// 现有 Logger
FocusLogger = NewLogger("FocusManager", "FOCUS")
ReconcilerLogger = NewLogger("Reconciler", "RECONCILER")
RenderLogger = NewLogger("Render", "RENDER")
KeyLogger = NewLogger("KeyEvent", "KEYS")
EventLogger = NewLogger("Event", "EVENTS")
UILogger = NewLogger("UI", "UI")
ButtonLogger = NewLogger("Button", "BUTTON")
HitMapLogger = NewLogger("HitMap", "HITMAP")
BorderLogger = NewLoggerWithEnv("Border", "BORDER", "TUI_BORDER_DEBUG")
PipelineLogger = NewLoggerWithEnv("Pipeline", "PIPELINE", "TUI_PIPELINE_DEBUG")
PaintLogger = NewLoggerWithEnv("Paint", "PAINT", "TUI_PAINT_DEBUG")
InspectorLogger = NewLogger("Inspector", "INSPECTOR")
LayoutLogger = NewLogger("Layout", "LAYOUT")
LayerLogger = NewLogger("Layer", "LAYER")
EngineLogger = NewLogger("Engine", "ENGINE")

// 新增 Logger
WrapLogger = NewLoggerWithEnv("Wrap", "WRAP", "TUI_WRAP_DEBUG")
PumpLogger = NewLogger("Pump", "PUMP")
FormLogger = NewLoggerWithEnv("Form", "FORM", "TUI_FORM_DEBUG")
CursorLogger = NewLoggerWithEnv("Cursor", "CURSOR", "TUI_CURSOR_DEBUG")
```

## 使用方法

### 旧方法（已废弃）

```go
if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
    log.UILogger.Debug("debug message")
}
```

### 新方法（推荐）

```go
// 直接使用 Logger，无需检查环境变量
log.InspectorLogger.Debug("debug message")

// 或者检查 Logger 是否启用
if log.InspectorLogger.Enabled() || log.UILogger.Enabled() {
    log.UILogger.Debug("debug message")
}
```

## 已更新的文件

以下文件已经从直接使用 `os.Getenv()` 更新为使用 Logger：

1. `ui/components/scrollview/vnode.go` - 使用 InspectorLogger
2. `framework/event/handler.go` - 使用 UILogger/InspectorLogger
3. `ui/components/wrap/vnode.go` - 使用 WrapLogger
4. `ui/components/button/vnode.go` - 使用 PaintLogger/RenderLogger
5. `framework/event/pump.go` - 使用 PumpLogger
6. `ui/components/form/instance.go` - 使用 FormLogger
7. `ui/components/cursor/instance.go` - 使用 CursorLogger
8. `ui/components/tabs/vnode.go` - 使用 InspectorLogger

## 需要进一步更新的文件

以下文件仍使用 `os.Getenv()`，建议统一更新：

- `runtime/layout/layer_manager.go` - TUI_DEBUG_HITMAP
- `framework/app.go` - TUI_DEBUG
- `internal/inspector/tree_view.go` - TUI_DEBUG_INSPECTOR
- `runtime/ui/fiber_util.go` - TUI_DEBUG_HITMAP
- `internal/render/paint_engine.go` - TUI_PAINT_DEBUG
- `ui/hooks.go` - TUI_DEBUG_UI
- `ui/app.go` - TUI_DEBUG_UI
- `runtime/ui/hooks.go` - TUI_DEBUG_UI
- `runtime/platform/input_windows.go` - TUI_DEBUG_WINDOWS
- `runtime/event/hitmap_debug.go` - TUI_DEBUG_HITMAP
- `runtime/layout/validator.go` - TUI_DEBUG_VALIDATION
- `internal/render/rendering_pipeline.go` - TUI_PIPELINE_DEBUG
- `internal/render/declarative_node.go` - TUI_DEBUG_UI
- `internal/reconciler/reconciler.go` - TUI_DEBUG_HITMAP
- `internal/reconciler/list_detector.go` - TUI_DEBUG_KEY
- `internal/reconciler/diff.go` - TUI_DEBUG_HITMAP
- `internal/reconciler/begin_work.go` - TUI_DEBUG_HITMAP
- `framework/inspector_integration.go` - TUI_DEBUG_UI
- `ui/components/input/instance.go` - TUI_INPUT_DEBUG
- `examples/fiber_counter/main.go` - TUI_DEBUG_UI
- `ui/layout.go` - TUI_DEBUG_LAYOUT

## 环境变量对照表

| 环境变量 | Logger | 说明 |
|---------|--------|------|
| TUI_DEBUG | 全局 | 启用所有调试日志 |
| TUI_DEBUG_FOCUS | FocusLogger | 焦点管理调试 |
| TUI_DEBUG_RENDER | RenderLogger | 渲染调试 |
| TUI_DEBUG_KEYS | KeyLogger | 键事件调试 |
| TUI_DEBUG_EVENTS | EventLogger | 事件调试 |
| TUI_DEBUG_UI | UILogger | UI 调试 |
| TUI_DEBUG_BUTTON | ButtonLogger | 按钮调试 |
| TUI_DEBUG_HITMAP | HitMapLogger | 点击映射调试 |
| TUI_INSPECTOR | InspectorLogger | Inspector 调试 |
| TUI_DEBUG_LAYOUT | LayoutLogger | 布局调试 |
| TUI_LAYER_DEBUG | LayerLogger | 层调试 |
| TUI_BORDER_DEBUG | BorderLogger | 边框调试 |
| TUI_PIPELINE_DEBUG | PipelineLogger | 管道调试 |
| TUI_PAINT_DEBUG | PaintLogger | 绘制调试 |
| TUI_WRAP_DEBUG | WrapLogger | 换行布局调试 |
| TUI_DEBUG_PUMP | PumpLogger | 事件泵调试 |
| TUI_FORM_DEBUG | FormLogger | 表单调试 |
| TUI_CURSOR_DEBUG | CursorLogger | 光标调试 |

## 注意事项

1. 所有环境变量支持多种值格式（`true`, `1`, `yes`, `on`），不区分大小写
2. `TUI_DEBUG` 环境变量会启用所有 Logger
3. 使用 `NewLoggerWithEnv()` 可以支持自定义环境变量名称
4. Logger 的 `Enabled()` 方法返回当前 Logger 是否启用
5. 直接调用 Logger 的 `Debug()` 方法会自动检查是否启用，无需手动检查
