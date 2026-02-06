# Single-Pass Layout Pipeline Refactor Design

## Problem Statement

Current implementation has dual measurement:
1. `measureVNode()` calls `LayoutNode.Measure()` (layout.go) to measure parent
2. `buildComputedBox()` calls `getChildConstraints()` + `buildComputedBox()` on children

These two paths use different constraint logic, causing inconsistencies.

## Root Cause

```go
// In buildComputedBox():
size := e.measureVNode(vnode, constraints)  // ← First measurement

for _, child := range vnode.Children() {
    childConstraints := e.getChildConstraints(...)  // ← Different logic!
    childBox := e.buildComputedBox(child, ...)  // ← Second measurement
}
```

`measureVNode()` for LayoutNode calls `LayoutNode.Measure()` which has its own constraint logic.
`getChildConstraints()` has separate, inconsistent constraint logic.

## Refactored Design

### Key Principles

1. **Single Source of Truth**: All layout measurement logic in one place
2. **Integrated Constraint Propagation**: Constraints are part of measurement, not separate
3. **No Dual Measurement**: Each node measured exactly once per layout pass

### New Architecture

```
buildComputedBox(vnode, parent, constraints)
│
├─► If VNode implements LayoutMeasurer interface:
│   └─► vnode.MeasureLayout(e, constraints, parent) → LayoutMeasurement
│       ├── Returns: size, childConstraints[], childSizes[]
│       └─► No recursive measurement in MeasureLayout
│
└─► For each child:
    └─► buildComputedBox(child, vnode, childConstraints[i])
```

### Interface Definition

```go
// LayoutMeasurer is implemented by nodes that want custom layout logic
type LayoutMeasurer interface {
    VNode

    // MeasureLayout measures this node and returns layout information
    // The engine uses this information to build the ComputedBox tree
    MeasureLayout(
        engine *Engine,
        constraints BoxConstraints,
        parent *ComputedBox,
    ) LayoutMeasurement
}

// LayoutMeasurement contains the result of measuring a layout node
type LayoutMeasurement struct {
    Size            Size
    ChildConstraints []BoxConstraints  // One per child
    ChildSizes      []Size            // Optional: pre-measured child sizes
}
```

### Implementation Strategy

#### Phase 1: Create New Interface (Non-Breaking)

```go
// Add to runtime/compute/engine.go

// MeasureLayout measures a node using the new single-pass approach
func (e *Engine) MeasureLayout(vnode VNode, constraints BoxConstraints, parent *ComputedBox) *LayoutMeasurement {
    if measurer, ok := vnode.(LayoutMeasurer); ok {
        return measurer.MeasureLayout(e, constraints, parent)
    }

    // Fallback: Use existing measureVNode + getChildConstraints
    size := e.measureVNode(vnode, constraints)
    children := vnode.Children()
    childConstraints := make([]BoxConstraints, len(children))

    for i, child := range children {
        childConstraints[i] = e.getChildConstraints(vnode, child, constraints, size)
    }

    return &LayoutMeasurement{
        Size:            size,
        ChildConstraints: childConstraints,
    }
}
```

#### Phase 2: Implement LayoutMeasurer for LayoutNode

```go
// In runtime/ui/layout.go

func (l *LayoutNode) MeasureLayout(
    engine *compute.Engine,
    constraints runtime.BoxConstraints,
    parent *compute.ComputedBox,
) compute.LayoutMeasurement {
    children := l.Children()
    if len(children) == 0 {
        return compute.LayoutMeasurement{
            Size: l.measureEmpty(constraints),
        }
    }

    // Get layout properties
    layoutInfo := GetLayoutInfo(l)
    gap := layoutInfo.Gap
    padding := layoutInfo.Padding

    if l.direction == DirectionRow {
        return l.measureHStackLayout(engine, constraints, layoutInfo, gap, padding)
    }
    return l.measureVStackLayout(engine, constraints, layoutInfo, gap, padding)
}

func (l *LayoutNode) measureHStackLayout(
    engine *compute.Engine,
    constraints runtime.BoxConstraints,
    layoutInfo *LayoutInfo,
    gap int,
    padding [4]int,
) compute.LayoutMeasurement {
    // Implementation that:
    // 1. Measures children with appropriate constraints
    // 2. Calculates flex distribution
    // 3. Returns: total size AND child constraints used
    //    (so buildComputedBox can reuse them)
}
```

#### Phase 3: Update buildComputedBox

```go
func (e *Engine) buildComputedBox(vnode VNode, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
    // ...

    // NEW: Try single-pass measurement first
    if measurer, ok := vnode.(LayoutMeasurer); ok {
        measurement := measurer.MeasureLayout(e, constraints, parent)
        box.Box.Width = measurement.Size.Width
        box.Box.Height = measurement.Size.Height

        // Build children using pre-calculated constraints
        for i, child := range vnode.Children() {
            childConstraints := measurement.ChildConstraints[i]
            childBox := e.buildComputedBox(child, box, childConstraints)
            if childBox != nil {
                box.Children = append(box.Children, childBox)
            }
        }
    } else {
        // FALLBACK: Use existing two-pass approach
        size := e.measureVNode(vnode, constraints)
        box.Box.Width = size.Width
        box.Box.Height = size.Height

        for _, child := range vnode.Children() {
            childConstraints := e.getChildConstraints(vnode, child, constraints, size)
            childBox := e.buildComputedBox(child, box, childConstraints)
            // ...
        }
    }

    // ...
}
```

### Cross-Axis Constraint Resolution

For the HStack-in-VStack alignment issue:

```go
func (l *LayoutNode) measureVStackLayout(...) compute.LayoutMeasurement {
    // ...

    childConstraints := make([]compute.BoxConstraints, len(children))

    for i, child := range children {
        childInfo := GetLayoutInfo(child)

        // NEW: Check if child needs cross-axis tight constraints
        childMinWidth := 0
        if innerMaxWidth != runtime.Infinity && isHStack(child) {
            // HStack in VStack fills width for main-axis alignment
            childMinWidth = innerMaxWidth
        }

        childConstraints[i] = runtime.BoxConstraints{
            MinWidth:  childMinWidth,
            MaxWidth:  innerMaxWidth,
            MinHeight: 0,
            MaxHeight: runtime.Infinity,
        }

        childSize := engine.MeasureChild(child, childConstraints[i])
        // ... accumulate totals
    }

    return compute.LayoutMeasurement{
        Size:            runtime.Size{Width: maxWidth, Height: totalHeight},
        ChildConstraints: childConstraints,
    }
}
```

## Benefits

1. **Single Source of Truth**: Layout logic lives in one place (LayoutNode.MeasureLayout)
2. **Consistent Constraints**: Child constraints determined once, reused everywhere
3. **Better Testability**: Can test MeasureLayout independently
4. **Clearer Separation**: Engine handles tree building, nodes handle measurement
5. **No Hidden Duplication**: All constraint logic visible in one method

## Migration Path

1. Add `LayoutMeasurer` interface and `LayoutMeasurement` struct
2. Implement `MeasureLayout()` for `LayoutNode`
3. Update `buildComputedBox()` to use new interface when available
4. Remove old `Measure()` implementation from `LayoutNode` (after validation)
5. Remove `getChildConstraints()` special cases
6. Add tests validating single-pass behavior

## Backward Compatibility

- Old `Measurable` interface still supported
- Nodes not implementing `LayoutMeasurer` use fallback path
- Gradual migration possible
