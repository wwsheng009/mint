# Mint 调试指南

本目录记录当前 Mint 调试入口和环境变量。旧版 `TUI_UI_DEBUG_*` 文档曾用于早期 demo/runtime internals 调试；当前框架源码的通用日志系统以 `internal/log` 为准，主要使用 `TUI_DEBUG*` 和 `TUI_LOG_*`。

## 快速入口

| 场景 | 推荐变量 |
|---|---|
| 打开全部调试日志 | `TUI_DEBUG_ALL=true` |
| 打开所有默认 logger | `TUI_DEBUG=true` |
| 只看渲染 | `TUI_DEBUG_RENDER=true` |
| 只看事件/输入泵 | `TUI_DEBUG_EVENTS=true`, `TUI_DEBUG_PUMP=true` |
| 只看 HitMap 命中 | `TUI_DEBUG_HITMAP=true` |
| 只看 UI 层日志 | `TUI_DEBUG_UI=true` |
| 日志输出到控制台 | `TUI_LOG_OUTPUT=console` |
| 日志同时写文件和控制台 | `TUI_LOG_OUTPUT=both` |
| 禁用异步渲染以便排查顺序问题 | `MINT_ASYNC_RENDER=false` |
| 保留屏幕输出，便于复制 | `MINT_NO_ALTERNATE_SCREEN=true` |

完整变量清单见 [environment_variables.md](environment_variables.md)。

## 基本用法

默认日志输出到 `./logs/application.log`：

```bash
TUI_DEBUG_RENDER=true go run ./examples/counter
```

输出到控制台：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
```

同时输出到文件和控制台：

```bash
TUI_LOG_OUTPUT=both TUI_DEBUG_EVENTS=true TUI_DEBUG_PUMP=true go run ./examples/menu_demo
```

排查鼠标命中、Portal、Overlay 和 Modal 点击问题：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_RENDER=true go run ./examples/modal
```

排查 Action 主路径：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_ACTION=true TUI_DEBUG_INTENT=true go run ./examples/store_reducer_demo
```

## 当前日志系统

当前通用日志系统实现位于：

- `../../internal/log/logger.go`: logger 分类和 `TUI_DEBUG*` 开关。
- `../../internal/log/file.go`: `TUI_LOG_OUTPUT` 和默认日志文件。
- `../../internal/log/rotation.go`: 日志轮转配置。

支持的布尔真值包括 `true`、`1`、`yes`、`on`，大小写不敏感。

## 常用分类

`internal/log` 当前定义了以下常用 logger：

- `TUI_DEBUG_FOCUS`
- `TUI_DEBUG_RENDER`
- `TUI_DEBUG_KEY`
- `TUI_DEBUG_EVENT`
- `TUI_DEBUG_WIN`
- `TUI_DEBUG_LINUX`
- `TUI_DEBUG_INSPECTOR`
- `TUI_DEBUG_LAYOUT`
- `TUI_DEBUG_LAYER`
- `TUI_DEBUG_ENGINE`
- `TUI_DEBUG_UI`
- `TUI_DEBUG_MESSAGE`
- `TUI_DEBUG_FIBER`
- `TUI_DEBUG_BUTTON`
- `TUI_DEBUG_HITMAP`
- `TUI_DEBUG_BORDER`
- `TUI_DEBUG_PIPELINE`
- `TUI_DEBUG_PAINT`
- `TUI_DEBUG_WRAP`
- `TUI_DEBUG_PUMP`
- `TUI_DEBUG_FORM`
- `TUI_DEBUG_CURSOR`
- `TUI_DEBUG_INPUT`
- `TUI_DEBUG_RENDERING`
- `TUI_DEBUG_VALIDATION`
- `TUI_DEBUG_ACTION`
- `TUI_DEBUG_INTENT`
- `TUI_DEBUG_PLATFORM`
- `TUI_DEBUG_TEMP`

注意：部分历史文档或旧代码可能写作 `TUI_DEBUG_EVENTS`、`TUI_DEBUG_KEYS`、`TUI_RENDER_DEBUG`、`TUI_UI_DEBUG_*`。更新文档时应以当前源码读取的变量为准。

## 渲染与运行时变量

框架运行时还读取若干 `MINT_*` 变量：

- `MINT_ASYNC_RENDER`: 异步渲染开关，默认开启；`false`、`0`、`no`、`off` 表示关闭。
- `MINT_ASYNC_FPS`: 异步渲染 FPS，默认约 60。
- `MINT_NO_ALTERNATE_SCREEN`: 保留终端输出，避免退出时清屏。
- `MINT_PORTAL_LAYOUT`: Portal-aware layout 开关，默认开启；`false` 或 `0` 可关闭。
- `MINT_GRAPHICS`: 图形输出模式，支持 `auto`、`off`/none、`kitty`、`sixel`、`inline-image` 等源码支持值。
- `MINT_CELL_PIXELS`: 终端 cell 像素尺寸，格式 `<width>x<height>`。
- `MINT_GRAPHICS_STRICT`
- `MINT_GRAPHICS_ALLOW_TERMINAL_FRAME`
- `MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE`
- `MINT_DEBUG_GRID`, `MINT_LOG_LEVEL`: Grid 组件调试相关。

## `cmd/mint-debugger`

`cmd/mint-debugger` 是独立 CLI，默认读取 `~/.mint/devtools/logs/session_*.log` 这类 DevTools JSONL 事件日志。它不是 `internal/log` 的 `./logs/application.log` 文本日志查看器。

常用参数：

```bash
go run ./cmd/mint-debugger -analyze-only
go run ./cmd/mint-debugger -watch
go run ./cmd/mint-debugger -log path/to/session.jsonl -report report.md
```

如果没有 DevTools session 日志，`mint-debugger` 不能直接分析普通 `TUI_DEBUG_*` 文本日志。

## 相关文档

- [environment_variables.md](environment_variables.md)
- [../log/LOGGER_ENV_VAR_STANDARD.md](../log/LOGGER_ENV_VAR_STANDARD.md)
- [../sandbox/SANDBOX_DEBUG_GUIDE.md](../sandbox/SANDBOX_DEBUG_GUIDE.md)

旧版 `docs/debugging/DEBUG_ENVIRONMENT_VARIABLES.md` 已归档到 `../../docsArchive/cleanup-2026-05-19/docs/debugging/`。
