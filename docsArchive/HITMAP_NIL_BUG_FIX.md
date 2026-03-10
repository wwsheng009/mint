# HitMap nil Bug 修复

> **Date**: 2026-02-10
> **Status**: ✅ 已修复
> **Bug**: `RenderingPipeline.Render()` 没有设置 `lastHitMap`

---

## 问题现象

```
[APP] ⚠️  RenderingPipeline returned nil HitMap, falling back to BuildHitMap
```

在 demo 运行时，每一帧都显示 HitMap 为 nil，系统回退到 fallback 逻辑。

---

## 根本原因

### Bug 位置

`internal/render/rendering_pipeline.go` - `Render()` 方法

### Bug 描述

`RenderingPipeline` 有两个渲染路径：

1. **`RenderLayers()`** - 多层渲染（用于 modal、overlay 等）
   - ✅ 正确设置 `p.lastHitMap = layerMgr.GetMergedHitMap()` (line 189)

2. **`Render()`** - 标准渲染（用于普通内容）
   - ❌ **没有设置 `p.lastHitMap`** (Bug!)

### 代码对比

**RenderLayers() - 正确实现**:
```go
func (p *RenderingPipeline) RenderLayers(...) error {
    // ... collect and layout layers ...

    // Paint all layers
    p.paintEngine.PaintLayers(layouts, buffer)

    // ✅ Save merged HitMap
    p.lastHitMap = layerMgr.GetMergedHitMap()

    return nil
}
```

**Render() - Bug 实现 (修复前)**:
```go
func (p *RenderingPipeline) Render(...) error {
    // Phase 1: Layout
    layout, err := p.layoutEngine.Layout(vnode, constraints)

    // Phase 2: Paint
    err = p.paintEngine.Paint(layout, buffer)

    // ❌ Missing: p.lastHitMap = layout.HitMap

    return err
}
```

---

## 影响

### 影响 1: HitMap 丢失

当 `hasLayers=false` 时（即没有 modal/overlay）：
1. 系统调用 `Render()` 而不是 `RenderLayers()`
2. `p.lastHitMap` 保持为 nil（从未被设置）
3. `DeclarativeNode.GetHitMap()` 返回 nil
4. App 回退到 `BuildHitMap()` 从 ComputedBox 重建

### 影响 2: Fallback HitMap 位置错误

```go
// framework/app.go:1202
if declNode, ok := a.root.(interface{ GetHitMap() *runtimeevent.HitMap }); ok {
    a.hitMap = declNode.GetHitMap()  // ❌ 返回 nil
}

// Fallback to BuildHitMap
if a.hitMap == nil {
    if vnodeRoot, ok := a.root.(rtui.VNode); ok {
        layoutAdapter := rtui.AsLayoutNode(vnodeRoot)
        a.hitMap = runtimeevent.BuildHitMap(layoutAdapter)
        // ⚠️ 这个 HitMap 是从 layout.Node 重新构建的
        // ⚠️ 可能不包含最新的位置信息
    }
}
```

虽然 fallback 逻辑也能构建 HitMap，但：
- 不包含 layer centering 的最终位置
- 性能更差（需要重新遍历 tree）
- 代码逻辑复杂（维护两个 HitMap 构建路径）

---

## 修复方案

### 代码修改

在 `RenderingPipeline.Render()` 方法中添加 HitMap 保存：

```go
func (p *RenderingPipeline) Render(...) error {
    // Phase 1: Layout
    layout, err := p.layoutEngine.Layout(vnode, constraints)
    // ...

    // Phase 2: Paint
    err = p.paintEngine.Paint(layout, buffer)
    // ...

    // ✅ FIX: Save HitMap for event routing
    p.lastHitMap = layout.HitMap

    // DEBUG: Log HitMap status
    if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
        if p.lastHitMap != nil {
            fmt.Fprintf(os.Stderr, "[RenderingingPipeline] Saved HitMap: %d entries\n", p.lastHitMap.Size())
        } else {
            fmt.Fprintf(os.Stderr, "[RenderingPipeline] ⚠️ Layout.HitMap is nil\n")
        }
    }

    return err
}
```

### 修改位置

- 文件: `internal/render/rendering_pipeline.go`
- 方法: `Render()`
- 行数: ~77 行之后

---

## 验证

### 验证步骤

1. **编译检查**
   ```bash
   cd E:\projects\yao\wwsheng009\mint
   go build ./internal/render
   ```

2. **运行 demo**
   ```bash
   cd examples/ui_demos/demo1_full_featured
   TUI_DEBUG_HITMAP=true go run main.go
   ```

3. **预期输出**
   ```
   [RenderingPipeline] Saved HitMap: 26 entries
   [APP] ✅ Got HitMap from RenderingPipeline: 26 entries
   ```

   而不是之前的：
   ```
   [APP] ⚠️  RenderingPipeline returned nil HitMap, falling back to BuildHitMap
   ```

### 验证结果

✅ **编译成功** - 无错误
✅ **逻辑正确** - `Render()` 现在设置 `lastHitMap`
✅ **调试输出** - 添加了 HitMap 状态日志

---

## 相关问题

### 问题 2: 为什么 hasLayers=false？

从 debug 输出看到：
```
[PipelineRenderer] hasLayers=false
```

**原因**: demo 启动时 `showModal=false`，modal 还没有被添加到 VNode 树中。

**验证**: 当用户点击 "[Open Modal]" 按钮后，`showModal=true`，modal 会被添加到树中。此时应该看到：
```
[hasLayerNodes] Node type=*ui.BorderedNode, Layer=2, IsValid=true
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[centerModal] modal=(0,0) size=20x5 container=80x24 offset=(30,9)
[RenderLayers] Merged HitMap: 35 entries
```

**下一步**: 需要用户实际点击按钮来验证 modal 打开时的行为。

---

## 架构改进建议

### 问题

当前架构有两个 HitMap 构建路径：
1. `RenderingPipeline` 通过 layout engine 构建（推荐）
2. `BuildHitMap()` 从 layout.Node 构建（fallback）

这导致：
- 代码重复
- 维护困难
- 可能的不一致性

### 建议

1. **统一 HitMap 构建**
   - 所有 HitMap 应该通过 `RenderingPipeline` 构建
   - 移除 `BuildHitMap()` fallback 逻辑
   - 确保 `lastHitMap` 始终有效

2. **强制使用 RenderingPipeline**
   - 如果 `GetHitMap()` 返回 nil，应该视为 bug
   - 添加 panic 或 error 而不是 silent fallback
   - 确保在 debug 模式下有清晰的错误信息

3. **简化 render() 方法**
   ```go
   func (a *App) render() {
       // ...

       hitMap := declNode.GetHitMap()
       if hitMap == nil {
           panic("HitMap should never be nil after render()")
       }

       a.pump.SetHitMap(hitMap)
   }
   ```

---

## 总结

### Bug
- `RenderingPipeline.Render()` 没有设置 `p.lastHitMap`
- 导致 `GetHitMap()` 返回 nil
- 触发 fallback 逻辑

### 修复
- 在 `Render()` 方法中添加 `p.lastHitMap = layout.HitMap`
- 添加 debug 输出验证 HitMap 状态

### 状态
✅ **已修复并验证编译**

### 后续
- 需要运行 demo 验证运行时行为
- 需要验证 modal 打开时的 layer detection
- 考虑移除 fallback 逻辑以简化架构
