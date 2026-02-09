# Inspector Overlay Not Showing - Analysis

## Problem Statement

- **With Inspector** (`TUI_INSPECTOR=true`): Interface does NOT show (blank screen)
- **Without Inspector** (`TUI_INSPECTOR=false`): Interface shows correctly

## What We've Fixed

1. ✅ **StripLayers logic**: Fixed to correctly preserve base content
2. ✅ **Position props**: Inspector overlay sets x=80, y=5
3. ✅ **Layer attribute**: Inspector overlay has LayerInspector
4. ✅ **Positioning**: LayerManager shifts Inspector to (80, 5)

## What We Know

From test output:
```
[CollectAndLayout] baseTree has 1 children (after stripping)
[CollectAndLayout]   child 0: layer=0 type=Element
[positionInspector] original=(0,0) target=(80,5)
[positionInspector] after shift: inspector=(80,5) size=80x5
```

This shows:
- ✅ baseTree has 1 child (correct)
- ✅ Inspector is positioned at (80, 5)
- ✅ Layer system is working

## The Real Problem

Looking at PaintEngine output:
```
[PaintEngine.Paint] START: layout.Root=*ui.BorderedNode, box=(0,0,80x1073741823)
```

Notice: **Height = 1073741823** - This is clearly WRONG!

This huge height value suggests:
1. Layout calculation is broken
2. OR height is uninitialized
3. OR there's integer overflow

## Root Cause Analysis

### When Inspector is DISABLED (works)

Flow:
```
RuntimeDemoWithInspectorOverlay()
  └─ return appContent  (NO overlay)
      └─ PipelineRenderer.Render()
          └─ LayoutEngine.Layout(appContent)
          └─ PaintEngine.Paint(layout)
              └─ NORMAL RENDERING ✅
```

### When Inspector is ENABLED (broken)

Flow:
```
RuntimeDemoWithInspectorOverlay()
  └─ return ui.VStack(appContent, inspectorOverlay)
      └─ DeclarativeNode.Paint()
          └─ PipelineRenderer.hasLayerNodes() = true
          └─ PipelineRenderer.RenderLayers()
              └─ LayerManager.CollectAndLayout()
                  ├─ StripLayers() → baseTree (appContent only)
                  ├─ Layout(baseTree) → baseLayout
                  └─ Layout(inspectorOverlay) → inspectorLayout
              └─ PaintEngine.PaintLayers({baseLayout, inspectorLayout})
                  ├─ Paint(baseLayout)  ← Renders at (0, 0, 80xHUGE_HEIGHT)
                  └─ Paint(inspectorLayout) ← Renders at (80, 5, 80x5)
```

## The Bug

The baseLayout has an **incorrect height** (1073741823).

This could be because:
1. **StripLayers creates a malformed tree**
2. **LayoutEngine calculates wrong height for stripped tree**
3. **Constraints are not properly passed**

## Investigation Steps

### Step 1: Check StripLayers Output

Create test to verify baseTree structure after stripping:

```go
// Test with actual app content structure
appContent := ui.VStack(
    HeaderPanel(),
    PipelineVisualization("idle"),
    StatisticsPanel(0, 0, 0),
    ControlPanel(...),
    ExplanationPanel("idle"),
)

inspectorOverlay := globalInspector.RenderOverlay()
root := ui.VStack(appContent, inspectorOverlay)

collector := NewCollector()
collector.Collect(root)
baseTree := collector.StripLayers(root)

// CHECK: Does baseTree equal appContent?
// CHECK: Does baseTree have correct structure?
```

### Step 2: Check Layout Calculation

```go
// Test layout on stripped tree
layout, err := engine.Layout(baseTree, constraints)

// CHECK: Is layout.Root.Box.Height correct?
// CHECK: Are all child boxes correct?
```

### Step 3: Compare With and Without Inspector

```go
// WITHOUT Inspector
root1 := appContent
layout1, _ := engine.Layout(root1, constraints)
// CHECK: layout1.Root.Box.Height

// WITH Inspector (after stripping)
root2 := ui.VStack(appContent, inspectorOverlay)
baseTree2 := collector.StripLayers(root2)
layout2, _ := engine.Layout(baseTree2, constraints)
// CHECK: layout2.Root.Box.Height

// COMPARE: Should be equal!
```

## Hypothesis

**StripLayers is modifying the tree structure**, causing LayoutEngine to calculate wrong dimensions.

Specifically, in `cloneWithoutLayers()`:
- We create NEW ElementVNode for LayoutNode, BorderedNode, etc.
- These new nodes might lose important layout properties
- Especially flex, stretch, or size properties

## Next Steps

1. ✅ Created tests for StripLayers (pass)
2. ❌ Need to test Layout calculation on stripped tree
3. ❌ Need to compare layout with/without Inspector
4. ❌ Need to verify baseTree preserves all layout properties

## Critical Code to Check

`runtime/layer/collector.go:cloneWithoutLayers()`

For LayoutNode:
```go
case *rtui.LayoutNode:
    // Creates NEW ElementVNode - loses LayoutNode properties!
    cloned := rtui.NewElement(n.Tag())
    cloned.SetProps(n.Props().Clone())
    cloned.SetStyle(n.Style())
    cloned.SetKey(n.Key())
    cloned.SetChildren(nonLayerChildren)
    return cloned
```

**This is the bug!** Converting LayoutNode to ElementVNode loses layout-specific properties.

## Solution

Preserve the original node type instead of converting to ElementVNode:

```go
case *rtui.LayoutNode:
    // Clone the LayoutNode properly
    cloned := n.Clone().(*rtui.LayoutNode)
    cloned.SetChildren(nonLayerChildren)
    return cloned
```

OR handle LayoutNode specially to preserve its layout properties.
