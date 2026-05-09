# 调试环境变量参考

本文档按当前源码整理 Mint 调试、日志、渲染和图形相关环境变量。源码入口主要是 `internal/log`、`framework/app.go`、`internal/render`、`runtime/platform/*graphics*` 和部分组件调试文件。

## 通用日志开关

日志系统位于 `internal/log/logger.go`。

| 变量 | 作用 | 说明 |
|---|---|---|
| `TUI_DEBUG_ALL` | 启用所有分类 logger | 优先级最高，任何 truthy 值都会打开所有分类。 |
| `TUI_DEBUG` | 启用默认/全局调试 | 也会让各分类 logger 被视为启用。 |
| `TUI_DEBUG_<CATEGORY>` | 启用指定分类 | `<CATEGORY>` 来自源码中的 logger category，例如 `RENDER`、`HITMAP`。 |

truthy 值支持：

- `true`
- `1`
- `yes`
- `on`

大小写不敏感。

## 当前 Logger 分类

| 变量 | Logger | 典型用途 |
|---|---|---|
| `TUI_DEBUG_FOCUS` | `FocusLogger` | 焦点管理。 |
| `TUI_DEBUG_RENDER` | `RenderLogger` | render bridge、DeclarativeNode、App render。 |
| `TUI_DEBUG_KEY` | `KeyLogger` | 按键消息。 |
| `TUI_DEBUG_EVENT` | `EventLogger` | framework event logger。 |
| `TUI_DEBUG_WIN` | `WinLogger` | Windows 相关日志。 |
| `TUI_DEBUG_LINUX` | `LinuxLogger` | Linux 相关日志。 |
| `TUI_DEBUG_INSPECTOR` | `InspectorLogger` | Inspector。 |
| `TUI_DEBUG_LAYOUT` | `LayoutLogger` | 布局。 |
| `TUI_DEBUG_LAYER` | `LayerLogger` | Layer/Portal 相关。 |
| `TUI_DEBUG_ENGINE` | `EngineLogger` | 引擎内部信息。 |
| `TUI_DEBUG_UI` | `UILogger` | UI 层通用日志。 |
| `TUI_DEBUG_MESSAGE` | `MessageLogger` | 消息流。 |
| `TUI_DEBUG_FIBER` | `FiberLogger` | Fiber。 |
| `TUI_DEBUG_BUTTON` | `ButtonLogger` | Button 组件。 |
| `TUI_DEBUG_HITMAP` | `HitMapLogger` | 鼠标命中图和目标路由。 |
| `TUI_DEBUG_BORDER` | `BorderLogger` | 边框。 |
| `TUI_DEBUG_PIPELINE` | `PipelineLogger` | render pipeline。 |
| `TUI_DEBUG_PAINT` | `PaintLogger` | paint。 |
| `TUI_DEBUG_WRAP` | `WrapLogger` | Wrap 布局。 |
| `TUI_DEBUG_PUMP` | `PumpLogger` | 输入事件泵。 |
| `TUI_DEBUG_FORM` | `FormLogger` | Form。 |
| `TUI_DEBUG_CURSOR` | `CursorLogger` | Cursor。 |
| `TUI_DEBUG_INPUT` | `InputLogger` | 输入组件。 |
| `TUI_DEBUG_RENDERING` | `RenderingLogger` | 渲染细节。 |
| `TUI_DEBUG_VALIDATION` | `ValidationLogger` | 验证。 |
| `TUI_DEBUG_ACTION` | `ActionLogger` | Action 主路径。 |
| `TUI_DEBUG_INTENT` | `IntentLogger` | Intent 分发。 |
| `TUI_DEBUG_PLATFORM` | `PlatFormLogger` | 平台层。 |
| `TUI_DEBUG_TEMP` | `TempLogger` | 临时排查。 |

注意：`internal/log/logger.go` 中 category 是单数形式，例如 `EVENT`、`KEY`。历史文档里的 `TUI_DEBUG_EVENTS`、`TUI_DEBUG_KEYS` 不应再作为推荐变量。

## 日志输出位置

实现位于 `internal/log/file.go` 和 `internal/log/rotation.go`。

| 变量 | 值 | 作用 |
|---|---|---|
| `TUI_LOG_OUTPUT` | `file` 或空 | 默认模式，写入 `./logs/application.log`。 |
| `TUI_LOG_OUTPUT` | `console` | 输出到 stderr。 |
| `TUI_LOG_OUTPUT` | `both` | 同时写入文件和 stderr。 |
| `TUI_LOG_MAX_SIZE` | 如 `100M`、`50K` | 单个日志文件最大大小，默认 `100M`。 |
| `TUI_LOG_MAX_FILES` | 整数 | 保留日志文件数量，默认 10。 |
| `TUI_LOG_COMPRESS` | bool | 是否压缩旧日志，默认 true。 |

示例：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
TUI_LOG_OUTPUT=both TUI_DEBUG_HITMAP=true go run ./examples/modal
TUI_LOG_MAX_SIZE=20M TUI_LOG_MAX_FILES=5 TUI_DEBUG=true go run ./examples/menu_demo
```

## Framework 运行变量

| 变量 | 默认 | 作用 |
|---|---|---|
| `MINT_ASYNC_RENDER` | 开启 | 异步渲染开关；`false`、`0`、`no`、`off` 可关闭。 |
| `MINT_ASYNC_FPS` | 约 60 | 异步渲染帧率。 |
| `MINT_NO_ALTERNATE_SCREEN` | 关闭 | 为 true 时避免清屏，保留输出方便复制。 |
| `TUI_OUTPUT_MODE` | 源码默认路径 | 控制渲染输出模式，源码中可见 `direct` 分支。 |
| `MINT_DEBUG_TEST` | 关闭 | 部分 framework/app 渲染测试诊断输出。 |
| `ACTION_DEBUG` | 关闭 | `framework.NewApp` 选择 Action debug middleware chain。 |
| `ACTION_PROD` | 关闭 | `framework.NewApp` 选择 Action production middleware chain。 |

## Fiber 和 Portal

| 变量 | 默认 | 作用 |
|---|---|---|
| `MINT_PORTAL_LAYOUT` | 开启 | Portal-aware two-phase layout；`false` 或 `0` 可关闭。 |
| `MINT_WARN_LEGACY` | 关闭 | Legacy render path deprecation warning 相关变量。 |

源码注释中仍可看到历史 `MINT_FIBER_FIRST` 字样，但 `NewDeclarativeNodeFromFuncWithFiber` 当前初始化为 Fiber-first，`ui.Run` 也默认创建 Fiber reconciler。

## 图形和图片渲染

实现位于 `runtime/platform/graphics_env.go` 和相关 graphics 文件。

| 变量 | 作用 |
|---|---|
| `MINT_GRAPHICS` | 图形模式，支持源码定义的 phase 1 模式，例如 `auto`、`kitty`、`sixel`、`inline-image`、off/none。 |
| `MINT_CELL_PIXELS` | 终端 cell 像素尺寸，格式 `<width>x<height>`。 |
| `MINT_GRAPHICS_CELL_PIXELS` | `MINT_CELL_PIXELS` 的 legacy alias。 |
| `MINT_GRAPHICS_STRICT` | 严格模式。 |
| `MINT_GRAPHICS_ALLOW_TERMINAL_FRAME` | 允许 terminal-frame 图形呈现。 |
| `MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE` | 允许未验证 inline image。 |
| `MINT_DEBUG_KITTY_GRAPHICS` | Kitty graphics 调试。 |
| `MINT_DEBUG_SIXEL_GRAPHICS` | Sixel graphics 调试。 |
| `MINT_DEBUG_INLINE_IMAGE_GRAPHICS` | Inline image 调试。 |

示例：

```bash
MINT_GRAPHICS=sixel MINT_CELL_PIXELS=8x16 go run ./examples/charts_linechart_image_prototype
MINT_GRAPHICS=inline-image MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE=true go run ./examples/charts_linechart_image_prototype
```

## 组件专项变量

| 变量 | 作用 |
|---|---|
| `MINT_DEBUG_GRID` | Grid 组件调试。 |
| `MINT_LOG_LEVEL` | Grid 调试日志级别。 |
| `MINT_UPDATE_E2E_SNAPSHOTS` | 更新 `ui/e2e` 渲染快照。 |

## 历史变量说明

以下变量曾出现在旧文档或早期 demo 中，但不应作为当前框架通用调试入口：

- `TUI_UI_DEBUG_LAYOUT`
- `TUI_UI_DEBUG_PAINT`
- `TUI_UI_DEBUG_ENGINE`
- `TUI_UI_DEBUG_ALL`
- `TUI_UI_DEBUG_LEVEL`
- `TUI_UI_DEBUG_FILE`
- `TUI_UI_DEBUG_QUIET`
- `TUI_UI_DEBUG_COLOR`
- `TUI_RENDER_DEBUG`
- `TUI_OUTPUT_DEBUG`
- `TUI_DEBUG_LOG`

如果某个示例或历史文档仍提到这些变量，应先核对对应源码是否仍读取它们，再决定保留还是迁移到 `TUI_DEBUG_<CATEGORY>`。

## 常用组合

渲染问题：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true TUI_DEBUG_PAINT=true go run ./examples/counter
```

鼠标点击和 overlay 命中问题：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true go run ./examples/modal
```

事件、Action、Intent 联动问题：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_PUMP=true TUI_DEBUG_ACTION=true TUI_DEBUG_INTENT=true go run ./examples/store_reducer_demo
```

测试时保留输出：

```bash
MINT_NO_ALTERNATE_SCREEN=true TUI_LOG_OUTPUT=console TUI_DEBUG_UI=true go run ./examples/menu_demo
```

禁用异步渲染排查顺序：

```bash
MINT_ASYNC_RENDER=false TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
```
