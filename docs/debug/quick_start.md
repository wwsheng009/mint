# 调试快速开始

本文档提供当前 Mint 调试变量的常用命令。完整参考见 [environment_variables.md](environment_variables.md)。

## 1. 输出到控制台

默认日志写入 `./logs/application.log`。排查时通常建议先输出到控制台：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG=true go run ./examples/counter
```

同时写文件和控制台：

```bash
TUI_LOG_OUTPUT=both TUI_DEBUG=true go run ./examples/counter
```

## 2. 渲染问题

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
```

Paint / pipeline：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_PAINT=true TUI_DEBUG_PIPELINE=true go run ./examples/counter
```

禁用异步渲染，排查输出顺序：

```bash
MINT_ASYNC_RENDER=false TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
```

## 3. 鼠标命中和 Overlay

适合排查 Modal、Popover、Select popup、Menu popup、Tooltip 等问题：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true go run ./examples/modal
```

运行相关 E2E：

```bash
go test ./ui/e2e -run "Overlay|Modal|Select|Popover|Tooltip" -count=1
```

## 4. Action / Intent

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_ACTION=true TUI_DEBUG_INTENT=true go run ./examples/store_reducer_demo
```

如果要同时看事件泵：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_PUMP=true TUI_DEBUG_ACTION=true TUI_DEBUG_INTENT=true go run ./examples/store_reducer_demo
```

## 5. Focus

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_FOCUS=true TUI_DEBUG_UI=true go run ./examples/ui_demos/demo1_full_featured
```

测试：

```bash
go test ./runtime/focus ./runtime/ui -run Focus -count=1
go test ./ui/e2e -run Focus -count=1
```

## 6. Layout / Layer

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_LAYOUT=true TUI_DEBUG_LAYER=true go run ./examples/layout_demo
```

Portal-aware layout 默认开启。需要排查兼容问题时可临时关闭：

```bash
MINT_PORTAL_LAYOUT=false TUI_LOG_OUTPUT=console TUI_DEBUG_LAYOUT=true go run ./examples/modal
```

## 7. 图形 / 图片渲染

```bash
MINT_GRAPHICS=sixel MINT_CELL_PIXELS=8x16 go run ./examples/charts_linechart_image_prototype
```

```bash
MINT_GRAPHICS=inline-image MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE=true go run ./examples/charts_linechart_image_prototype
```

## 8. 保留终端输出

便于复制和回看输出：

```bash
MINT_NO_ALTERNATE_SCREEN=true go run ./examples/counter
```

## 9. 历史变量提醒

早期文档中的 `TUI_UI_DEBUG_*`、`TUI_RENDER_DEBUG`、`TUI_OUTPUT_DEBUG` 等不再作为当前框架通用调试入口。当前通用入口请使用：

- `TUI_DEBUG_ALL`
- `TUI_DEBUG`
- `TUI_DEBUG_<CATEGORY>`
- `TUI_LOG_OUTPUT`

完整说明见 [environment_variables.md](environment_variables.md)。
