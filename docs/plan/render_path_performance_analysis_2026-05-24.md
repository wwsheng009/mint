# 渲染路径与性能优化分析报告

日期：2026-05-24

## 结论摘要

Mint 当前主渲染路径已经具备双缓冲、异步提交、行级 diff、事件合并、按需 tick 等基础优化，性能问题主要不在终端输出层，而在单帧内部的 Fiber-first 渲染流水线：

1. `PaintPaintablePlanes` 对每个 PaintableBox 都创建临时 `PaintableLayout` 并调用 `paintLayout`，而 `paintLayout` 会重复初始化缓存上下文、重复重建 parent background map，并可能触发重复的整帧状态处理。
2. Paint 阶段的 planes 收集会遍历整棵 PaintableBox 树，后续 frame box tracking 又再次递归遍历，完整绘制中还会递归绘制子节点，存在多次 O(n) 树遍历和潜在重复绘制。
3. `FiberToPaintableConverter.findFiber` 已经构建了 `fiberMap`，但 miss 时仍会线性扫描并 `fmt.Sprintf` 比较 NodeID，属于可消除的 O(n) fallback。
4. `PaintEngine` 的 cache 设计当前按 frame version 匹配，但 version 每次绘制递增，缓存命中率容易被压低；同时 cache key 与 tracking key 都使用格式化字符串，热路径分配较多。
5. `framework.App.render` 每帧都 `buf.Reset` 整个终端 buffer，后续虽然用 renderer diff 输出到终端，但内存写入仍是全屏级别。这个属于更大范围优化，需要和 dirty rect 语义一起处理。

本次优先实施低风险优化：让 `PaintPaintablePlanes` 在一个 frame 内只初始化一次 paint frame 状态，并直接调用 `paintBox`，移除每个 box 的临时 `PaintableLayout` 创建与重复 `paintLayout` 初始化。

## 主渲染路径

应用入口：

- `app.Run` 仅做薄封装，最终进入 `ui.Run`。
- `ui.Run` / `ui.RunApp` 创建 `framework.App`，初始化主题、Intent runtime、Fiber reconciler、默认 Portal roots，然后 `fwApp.SetRoot(declarativeRoot)` 并进入 `fwApp.Run()`。

主循环：

- `framework.App.Run` 使用 ticker 和 event pump 驱动。
- 输入事件会被 drain，键盘事件保留，鼠标 move 合并为最新一条。
- 用户输入会绕过普通 throttler 尽快渲染，非输入事件会受 `render.Throttler` 限制。
- tick 只有在 `activeTickables` 或 `dirty` 时才触发有效渲染。

单帧渲染：

- `framework.App.render`
- `renderer.GetBackBuffer()`
- `buf.Reset(terminalWidth, terminalHeight)`
- root 实现 `component.Paintable` 时进入 `DeclarativeNode.Paint`
- Fiber-first 模式进入 `DeclarativeNode.fiberFirstPaint`

Fiber-first 流水线：

- Reconcile：`n.reconciler.Render(..., renderFn)`
- Layout：标准路径 `n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)`；Portal 路径 `n.portalLayoutEngine.Layout(fiberRoot, constraints)`
- Convert：`NewFiberToPaintableConverter(fiberRoot).ConvertToLayout(innerResult.Root)`
- Planes：遍历 PaintableBox 树并按 layer 加入 `paint.PaintablePlanes`
- Paint：`n.paintEngine.PaintPaintablePlanes(planes, buf)`
- HitMap：`layoutResult.GetHitMap()` 转为 `runtime/event.HitMap` 并交给 pump
- Output：`renderTextFrame` 走 async renderer 或同步 `paint.Renderer.Render()` 行级 diff 输出

## 当前已有优化

- `framework.App.Run` 已合并鼠标移动事件，减少高频 move backlog。
- `framework.App` 已维护 `activeTickables`，无 tickable 且不 dirty 时跳过 ticker 渲染。
- `paint.Renderer` 使用 front/back buffer 和 line diff，终端输出不再默认整屏写。
- `paint.AsyncRenderer` 默认启用，可减少主循环阻塞。
- `runtime/paint.Buffer` 有 line hash、宽字符 continuation 语义和 optimized string 输出。
- `PaintEngine` 已移除一次明显的 `UpdateBufferCopy` 克隆开销注释，但 `NewPaintingContext` 仍会在初始化时克隆一次 buffer。

## 主要性能风险点

### 1. PaintPaintablePlanes 每个 box 重复初始化 paintLayout

位置：`internal/render/paint_engine.go`

当前逻辑：

- 遍历每一层的每个 box。
- 每个 box 都 `paint.NewPaintableLayout(box)`。
- 每个 box 都调用 `e.paintLayout(layout, buffer, false)`。
- `paintLayout` 内部会处理 cache 初始化、version 递增、parent background map 重建、force full render 判断。

影响：

- 对 n 个 PaintableBox 至少产生 n 次 layout wrapper 分配。
- cache context 在同一帧内反复初始化，cache version 反复递增，缓存命中语义被削弱。
- parent background map 反复重建，继承背景在一个 frame 中更难复用。
- 如果 `paintBox` 递归绘制子节点，则 planes 中的父子节点会造成潜在重复绘制。

本次优化：

- 新增 frame 级 paint 初始化方法。
- `PaintPaintablePlanes` 每帧只初始化一次 cache context / parent background / force full render。
- 每个 box 直接 `paintBox(box, buffer)`，不再创建临时 `PaintableLayout`。

### 2. FiberToPaintableConverter fallback 查找是 O(n)

位置：`internal/render/converter.go`

`NewFiberToPaintableConverter` 已经将 `DiffKey`、`Key`、`NodeID string` 放入 map，但 `findFiber` 在 map miss 后仍遍历整个 map，并对每个 Fiber 执行 `fmt.Sprintf("%d", f.NodeID)`。

影响：

- LayoutBox 数量多时，转换阶段可能退化为 O(n^2)。
- `fmt.Sprintf` 在热路径上产生额外分配和格式化成本。

建议：

- 保证 LayoutBox.ID 与 fiber map key 约定一致。
- 如需 fallback，构建独立 `nodeIDMap map[uint64]*Fiber` 或彻底移除线性扫描。

### 3. HitMap 转换逐项 FindFiberByID

位置：`internal/render/layout_switcher.go`

`convertLayoutHitMap` 对 layout hit entry 逐项调用 `rtui.FindFiberByID(fiberRoot, nodeID)`。如果 `FindFiberByID` 是树扫描，则 m 个 hit entry 会产生 O(m*n)。

建议：

- 在同一帧已有 Fiber -> Paintable 转换 map 的情况下复用 NodeID -> Fiber 索引。
- 或在 HitMap 转换前一次性构建 `map[uint64]*Fiber`。

### 4. Buffer Reset 是整屏级写入

位置：`framework/app.go`

每帧 `buf.Reset(a.terminalWidth, a.terminalHeight)` 会清空 back buffer 全部 cell。终端输出层虽然是 diff，但内存层仍按全屏重写。

建议：

- 中短期保留当前语义，因为 TUI 宽字符和 stale cell 清理依赖全覆盖。
- 后续可基于 `GetPaintDirtyRects` 和 renderer dirty hints，探索局部 reset 或分层 overlay buffer。

### 5. Paint cache 命中条件偏保守

位置：`internal/render/cache/paint.go`、`internal/render/paint_engine.go`

cache 使用 `boxID + version` 匹配，但 version 按绘制推进。若 version 每帧变化，上一帧缓存会自然 miss。并且部分节点在 cache 判定前调用一次 `box.Node.Paint` 检查 custom paint，随后正式绘制又调用一次 `Paint`。

建议：

- 将 cache version 改为节点内容版本或 dirty generation，而不是 frame generation。
- 为 `PaintableNode` 增加轻量 `HasCustomPaint` / `IsCacheable` 能力，避免为探测而调用 `Paint`。

## 已执行验证基线

命令：

```powershell
go test ./internal/render -bench Benchmark -benchmem -run '^$'
```

结果：

- 通过。
- 当前 benchmark 主要覆盖 VNode 创建、Pipeline Measure、focusable 收集和 Props/Style helper。
- 示例数据：`BenchmarkMeasure_Pipeline_Text` 约 `8258 ns/op`、`1040 B/op`、`14 allocs/op`；`BenchmarkMeasure_Pipeline_HStack` 约 `23360 ns/op`、`4516 B/op`、`39 allocs/op`。
- 现有 benchmark 不覆盖完整 `fiberFirstPaint -> PaintPaintablePlanes -> Renderer.Render` 路径，建议后续补充。

## 优化实施计划

本轮实施：

- 在 `PaintEngine` 中拆出 frame 级初始化，避免 planes 绘制时每个 box 重复初始化。
- `PaintPaintablePlanes` 在 planes 已包含所有子节点时只绘制当前 box 自身，避免父节点递归重复绘制整棵子树；对只传 root 的兼容路径保持递归绘制。
- `PaintPaintablePlanes` 移除每 box `paint.NewPaintableLayout` 分配。
- `FiberToPaintableConverter.findFiber` 移除冗余 O(n) fallback，NodeID 字符串格式化改为 `strconv.FormatUint`。
- `convertLayoutHitMap` 改为每帧先构建 `map[uint64]*Fiber`，再按 HitMap entry O(1) 查找目标 Fiber，避免每个 entry 扫描整棵 Fiber 树。
- 保持布局、HitMap、事件路由、Portal 逻辑不变。

后续建议：

- 为完整 Fiber-first 单帧渲染增加 benchmark，至少覆盖 100/500/1000 节点三档。
- 重新设计 Paint cache version，使缓存真正跨帧命中。
- 评估 `buf.Reset` 与 dirty rect 的协同，逐步减少全屏内存写入。
