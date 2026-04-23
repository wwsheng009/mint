# Inspector Rendering Flow - Complete Analysis

## Problem Statement

**启用 TUI_INSPECTOR=true 后，原界面和 Inspector 都无法显示**

## Complete Rendering Flow

### 1. Application Entry Point (`main.go`)

```go
// Line 46-50: Inspector 环境变量检查
if os.Getenv("TUI_INSPECTOR") == "true" {
    globalInspector.ToggleVisibility()  // 设置 visible=true
    fmt.Println("UI Inspector auto-enabled")
}
```

**关键点**:
- `TUI_INSPECTOR=true` → 调用 `ToggleVisibility()` 设置 `visible=true`
- `TUI_INSPECTOR=false` → 不调用，`visible=false` (默认)

### 2. Render Function (`RuntimeDemoWithInspectorOverlay`)

```go
// Line 125: 检查 Inspector 可见性
inspectorVisible := globalInspector.IsVisible()

// Line 134-157: 根据 visible 状态返回不同的 VNode
if inspectorVisible {
    // TUI_INSPECTOR=true 走这里
    inspectorOverlay := globalInspector.RenderOverlay()  // LayerInspector layer
    return ui.VStack(appContent, inspectorOverlay)
} else {
    // TUI_INSPECTOR=false 走这里
    return appContent  // 只有应用内容，没有 layer
}
```

**关键差异**:
- **With Inspector**: 返回 `VStack(appContent, inspectorOverlay)` - 包含 LayerInspector 节点
- **Without Inspector**: 返回 `appContent` - 没有 layer 节点

### 3. PipelineRenderer Detection (`pipeline_renderer.go:63`)

```go
// 检测是否有 layer 节点
hasLayers := r.hasLayerNodes(vnode)

// Line 69-82: 根据检测结果选择渲染路径
if hasLayers {
    // TUI_INSPECTOR=true: 有 LayerInspector 节点
    err = r.pipeline.RenderLayers(vnode, constraints, buf)
} else {
    // TUI_INSPECTOR=false: 没有 layer 节点
    err = r.pipeline.Render(vnode, constraints, buf)
}
```

**渲染路径分支**:
- **With Inspector**: `RenderLayers()` → 多层渲染
- **Without Inspector**: `Render()` → 单层渲染 ✅ 正常工作

### 4. RenderLayers Flow (`rendering_pipeline.go:145`)

```go
func (p *RenderingPipeline) RenderLayers(...) error {
    // Step 1: Collect and layout all layers
    layerMgr := layer.NewManager()
    layerMgr.CollectAndLayout(vnode, constraints, p.layoutEngine)

    // Step 2: Get all layer layouts
    layouts := layerMgr.GetLayouts()

    // Step 3: Paint all layers
    p.paintEngine.PaintLayers(layouts, buffer)
}
```

### 5. LayerManager.CollectAndLayout (`runtime/layer/manager.go`)

```go
func (m *Manager) CollectAndLayout(...) error {
    // Step 1: Collect layer nodes from VNode tree
    collector.Collect(vnode)
    // → Finds: LayerBase nodes (appContent)
    // → Finds: LayerInspector nodes (inspectorOverlay)

    // Step 2: Strip layers from base tree
    baseTree := collector.StripLayers(vnode)
    // → Returns: appContent without inspector overlay

    // Step 3: Layout base tree
    baseLayout := engine.Layout(baseTree, constraints)

    // Step 4: Layout each layer independently
    for layer, nodes := range collector.GetLayers() {
        for _, node := range nodes {
            layerLayout := engine.Layout(node.Content, constraints)

            // Step 5: Position Inspector overlay
            if layer == rtui.LayerInspector {
                m.positionInspector(node, layerLayout.Root)
                // → Sets inspector position to (80, 5)
            }

            layouts[layer] = layerLayout
        }
    }
}
```

**Debug Output (from status report)**:
```
[PipelineRenderer] hasLayers=true ✅
[CollectAndLayout] baseTree has 1 children (after stripping) ✅
[positionInspector] inspector=(80,5) size=80x5 ✅
```

### 6. PaintEngine.PaintLayers (`paint_engine.go:329`)

```go
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer *paint.Buffer) error {
    renderOrder := []rtui.Layer{
        rtui.LayerBase,      // 0
        rtui.LayerOverlay,   // 1
        rtui.LayerModal,     // 2
        rtui.LayerTooltip,   // 3
        rtui.LayerInspector, // 4
    }

    for _, l := range renderOrder {
        layout, ok := layouts[l]
        if !ok || layout.Root == nil {
            continue
        }

        // Paint this layer to buffer
        e.Paint(layout, buffer)
    }
}
```

**关键**: 所有层都绘制到同一个 buffer，从底层到高层逐层覆盖

### 7. The Problem: Where Does It Fail?

#### Path 1: Without Inspector (Works ✅)

```
RuntimeDemoWithInspectorOverlay()
  └─ return appContent  (NO overlay)
      └─ PipelineRenderer.Render()
          └─ hasLayerNodes() = false
          └─ RenderingPipeline.Render()
              └─ LayoutEngine.Layout(appContent)
              └─ PaintEngine.Paint(layout)
                  └─ paintNode(layout.Root, buffer)
                      └─ buffer.DrawString(...)
                          └─ Terminal displays content ✅
```

#### Path 2: With Inspector (Broken ❌)

```
RuntimeDemoWithInspectorOverlay()
  └─ return ui.VStack(appContent, inspectorOverlay)
      └─ PipelineRenderer.Render()
          └─ hasLayerNodes() = true  (has LayerInspector)
          └─ RenderingPipeline.RenderLayers()
              ├─ LayerManager.CollectAndLayout()
              │   ├─ StripLayers() → baseTree ✅
              │   ├─ Layout(baseTree) → baseLayout ✅
              │   └─ Layout(inspectorOverlay) → inspectorLayout ✅
              │       └─ positionInspector() → (80, 5) ✅
              └─ PaintEngine.PaintLayers({baseLayout, inspectorLayout})
                  ├─ Paint(baseLayout)  → 绘制到 (0, 0)
                  └─ Paint(inspectorLayout) → 绘制到 (80, 5)
                      └─ Buffer should have content
                          └─ ??? But nothing displays ❌
```

## Root Cause Analysis

### What We Know

1. ✅ **Layer detection works**: `hasLayerNodes()` correctly detects LayerInspector
2. ✅ **StripLayers works**: baseTree has 1 child (appContent)
3. ✅ **Layout works**: Both layers have valid layout with computed positions
4. ✅ **Positioning works**: Inspector positioned at (80, 5) with size 80x5
5. ❌ **Display fails**: Nothing shown on terminal

### What Could Be Wrong

#### Hypothesis 1: PaintLayers Not Called

**Check**: Add debug log at start of `PaintLayers()`
```go
func (e *PaintEngine) PaintLayers(...) {
    fmt.Fprintf(os.Stderr, "[PaintLayers] Called with %d layers\n", len(layouts))
}
```

#### Hypothesis 2: Buffer Not Output to Terminal

**Check**: After `PaintLayers()`, verify buffer has content
```go
// In RenderLayers(), after PaintLayers()
hasContent := false
for y := 0; y < buffer.Height; y++ {
    for x := 0; x < buffer.Width; x++ {
        if buffer.Cells[y][x].Rune != 0 || buffer.Cells[y][x].Cluster != "" {
            hasContent = true
            break
        }
    }
}
fmt.Fprintf(os.Stderr, "[RenderLayers] Buffer has content: %v\n", hasContent)
```

#### Hypothesis 3: Base Layout Has Wrong Size

**From status report**:
```
[PaintEngine.Paint] box=(0,0,80x1073741823)
```

**Problem**: Height = 1073741823 (uninitialized/overflow) instead of normal height ~19

**Root Cause**: This was fixed by preserving LayoutNode/BorderedNode types in StripLayers

**But Wait**: The fix was applied, so why still doesn't work?

#### Hypothesis 4: VStack Changes Rendering Behavior

**Key insight**: When inspector is enabled, we return `ui.VStack(appContent, inspectorOverlay)`

This VStack:
- Contains LayerBase children (appContent)
- Contains LayerInspector children (inspectorOverlay)

**Question**: Does VStack with layer children break the layout?

**Check**: In `cloneWithoutLayers()`:
```go
for _, child := range vnode.Children() {
    if child.GetLayer() != rtui.LayerBase {
        changed = true
        continue  // Skip layer children
    }
    filteredChild := c.cloneWithoutLayers(child)
    nonLayerChildren = append(nonLayerChildren, filteredChild)
}
```

**This should work correctly** - it filters out the inspector overlay.

#### Hypothesis 5: Inspector Position Outside Screen

**From status report**:
```
[positionInspector] inspector=(80,5) size=80x5
```

**Problem**: Inspector positioned at x=80, which is at the right edge of the screen (screen width = 120)

- Screen width: 120
- Inspector x: 80
- Inspector width: 80
- Inspector right edge: 80 + 80 = 160 > 120 ❌ **OFF SCREEN!**

**This is the bug!** The inspector is positioned beyond the screen width.

## The Real Bug: Inspector Position Calculation

### Current Positioning Logic

```go
// Inspector sets default position
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    content.SetProps(ui.Props{
        "x": si.floatX,  // Default: 80
        "y": si.floatY,  // Default: 5
    })
}
```

### Screen Size

```go
// main.go line 61
fwApp.Resize(120, 40)
```

### Position Analysis

```
Screen:  0 ....................................................... 119
         |<---------------- 120 pixels ----------------------->|

Inspector:                            [80 ............ 159]
                                          ^         ^
                                        x=80    x+80=160
                                                   |
                                        OFF SCREEN! ❌
```

**Inspector right edge at 160 > screen width 120**

## Solution

The inspector position needs to be calculated based on screen size:

```go
// Option 1: Position inspector at right edge, ensure it fits on screen
screenWidth := 120
inspectorWidth := 80
inspectorX := screenWidth - inspectorWidth  // 120 - 80 = 40

// Option 2: Position inspector partially off-screen (scrollable)
inspectorX := screenWidth - 40  // 120 - 40 = 80 (current)

// Option 3: Center inspector on right side
inspectorX := (screenWidth + inspectorWidth) / 2 - inspectorWidth
// (120 + 80) / 2 - 80 = 20
```

## Recommended Fix

Change inspector default position to be screen-aware:

```go
// internal/inspector/standalone_inspector.go
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    // Get screen size from framework app
    screenWidth, screenHeight := si.getScreenSize()

    // Calculate position to fit on screen
    inspectorWidth := 80
    inspectorHeight := 30

    // Position at right edge, ensure it fits
    x := screenWidth - inspectorWidth
    if x < 0 { x = 0 }

    y := 5  // Top margin

    content.SetProps(ui.Props{
        "x": x,
        "y": y,
    })
}
```

## Testing

After fix, verify with:
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

Expected output:
- Main app content visible on left side
- Inspector overlay visible on right side
- Both fully interactive
