# 渲染系统文档

本目录包含 Mint 渲染系统的当前入口、历史设计和像素/图形渲染计划。当前默认声明式渲染路径是 Fiber-first。

## 当前源码主路径

```text
ui.Run
  -> internal/render.NewDeclarativeNodeFromFuncWithFiber
  -> internal/reconciler.Reconciler
  -> Fiber root
  -> internal/render.FiberToNodeAdapter
  -> runtime/layout.Engine
  -> optional PortalAwareLayoutEngine
  -> internal/render.FiberToPaintableConverter
  -> paint.PaintableBox / PaintablePlanes
  -> internal/render.PaintEngine
  -> runtime/paint.Buffer or SceneFrame
  -> framework.App renderer / terminal output
```

核心源码：

- `../../internal/render`
- `../../internal/reconciler`
- `../../runtime/paint`
- `../../runtime/render`
- `../../runtime/layout`
- `../../runtime/platform/*graphics*`
- `../../framework/app.go`

## 当前实现要点

- Fiber-first 是 `ui.Run` 默认路径。
- VNode 是声明输入；reconcile 后 layout/paint 主要基于 Fiber 和 ComponentInstance。
- Portal-aware layout 默认启用，可用 `MINT_PORTAL_LAYOUT=false` 或 `MINT_PORTAL_LAYOUT=0` 关闭。
- `framework.App` 默认启用 async renderer，可用 `MINT_ASYNC_RENDER=false` 关闭。
- `PaintEngine.PaintPaintablePlanes()` 负责按 layer 绘制 paint tree。
- HitMap 在 render 后传给事件泵，用于鼠标 target 填充。
- 图形/图片渲染仍属于实验/设计增强区域，受 `MINT_GRAPHICS`、`MINT_CELL_PIXELS` 等变量控制。

## 目录结构

当前实际目录：

```text
render/
  diff/
  hook/
  paint/
    optimized/
      refactor/
  pixel/
```

### `diff/`

- [diff/diff.md](diff/diff.md): buffer diff 与输出优化说明。

### `hook/`

- [hook/README.md](hook/README.md): render hook 说明。

### `paint/optimized/`

Fiber-first paint pipeline 当前说明：

- [paint/optimized/README.md](paint/optimized/README.md)
- [paint/optimized/FIBER_FIRST_RENDER_PIPELINE.md](paint/optimized/FIBER_FIRST_RENDER_PIPELINE.md)

历史迁移和实施指南已归档到 `../../docsArchive/cleanup-2026-05-19/docs/render/paint/optimized/`。

### `pixel/`

Pixel / image / graphics 当前保留文档：

- [pixel/README.md](pixel/README.md)
- [pixel/PIXEL_CHART_RENDERING_ARCHITECTURE.md](pixel/PIXEL_CHART_RENDERING_ARCHITECTURE.md)
Historical implementation sequence and PR planning notes were archived to `../../docsArchive/cleanup-2026-05-19/docs/render/pixel/`.

完整清单见 [pixel/README.md](pixel/README.md)。

## 调试变量

常用：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_RENDER=true go run ./examples/counter
TUI_LOG_OUTPUT=console TUI_DEBUG_PAINT=true go run ./examples/counter
TUI_LOG_OUTPUT=console TUI_DEBUG_PIPELINE=true go run ./examples/counter
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true go run ./examples/modal
MINT_ASYNC_RENDER=false TUI_DEBUG_RENDER=true go run ./examples/counter
```

图形/像素相关：

```bash
MINT_GRAPHICS=sixel MINT_CELL_PIXELS=8x16 go run ./examples/charts_linechart_image_prototype
MINT_GRAPHICS=inline-image go run ./examples/charts_linechart_image_prototype
```

完整变量见 [../debug/environment_variables.md](../debug/environment_variables.md)。

## 测试建议

```bash
go test ./internal/render ./internal/reconciler ./runtime/paint ./runtime/render -count=1
go test ./examples/charts_linechart_image_prototype/... -count=1
go test ./ui/e2e -run "Render|Charts|Overlay" -count=1
```

## 历史说明

旧文档中可能提到 `plan/`、`fixes/`、`tui/` 等目录；这些目录当前不在 `docs/render/` 下。历史记录通常已经移入 `docsArchive/` 或保留在其他专题目录中。

## 相关文档

- [../architecture/README.md](../architecture/README.md)
- [../fiber/fiber_first/consolidated/README.md](../fiber/fiber_first/consolidated/README.md)
- [../layout/README.md](../layout/README.md)
- [../layer/LAYER_SYSTEM_ARCHITECTURE.md](../layer/LAYER_SYSTEM_ARCHITECTURE.md)
- [../debug/environment_variables.md](../debug/environment_variables.md)
