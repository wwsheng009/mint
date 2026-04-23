# Fiber-First Layout Implementation Plan

## Overview

**Goal**: Implement layout that only uses Fiber tree, without depending on VNode.Children()

**Current State**: Phase 4.5 - Hybrid approach using both Fiber and VNode
**Target**: Phase 6 - Pure Fiber-based layout

---

## Architecture Comparison

### Current (Phase 4.5) - Hybrid

```
buildComputedBox(vnode, fiber, parent, constraints)
  ├── Uses vnode.Children() to get child list ❌
  ├── Uses vnode.(LayoutMeasurer) for flex algorithm ❌
  ├── Uses vnode.(Measurable) for content measurement ❌
  └── Matches Fiber nodes by DiffKey for NodeID propagation ✅
```

**Problems:**
- Requires VNode for almost everything
- Fiber tree structure (Child → Sibling) is not utilized
- Cannot work with Fiber-only trees

### Target (Phase 6) - Fiber-First

```
buildComputedBoxFromFiber(fiber, parent, constraints)
  ├── Uses fiber.Child → fiber.Sibling for traversal ✅
  ├── Uses fiber.MeasureLayout() for flex algorithm ✅
  ├── Uses fiber.MeasureContent() for content measurement ✅
  └── No VNode dependency ✅
```

---

## Implementation Plan

### Phase 5.1: Add Layout Methods to Fiber

**File**: `runtime/ui/fiber.go`

**Add Methods**:

```go
// MeasureLayout implements flex layout algorithm for this fiber
func (f *Fiber) MeasureLayout(measurer ChildMeasurer, constraints BoxConstraints) LayoutMeasurement {
    // Implementation moves from ElementVNode.MeasureLayout()
}

// MeasureContent measures natural size of this fiber's content
func (f *Fiber) MeasureContent(constraints BoxConstraints) Size {
    // Implementation moves from TextVNode.Measure(), ImageVNode.Measure(), etc.
}

// GetChildFibers returns all child fibers as a slice
// This converts Child→Sibling linked list to array for flex layout
func (f *Fiber) GetChildFibers() []*Fiber {
    var children []*Fiber
    for child := f.Child; child != nil; child = child.Sibling {
        children = append(children, child)
    }
    return children
}

// GetChildCount returns the number of children
func (f *Fiber) GetChildCount() int {
    count := 0
    for child := f.Child; child != nil; child = child.Sibling {
        count++
    }
    return count
}
```

### Phase 5.2: Move Layout Algorithm to Fiber

**File**: `runtime/ui/fiber_layout.go` (new file)

**Purpose**: Implement flex layout algorithm directly on Fiber

```go
package ui

import (
    "github.com/wwsheng009/mint/runtime"
)

// MeasureLayout implements flex layout for Fiber nodes
// This is the same algorithm as ElementVNode.MeasureLayout()
// but operates on Fiber tree instead of VNode tree
func (f *Fiber) MeasureLayout(measurer ChildMeasurer, constraints runtime.BoxConstraints) LayoutMeasurement {
    if f == nil {
        return NewLayoutMeasurement(runtime.Size{}, nil)
    }

    // Get children as array
    children := f.GetChildFibers()
    childCount := len(children)

    // Build child constraints and measure each child
    childConstraints := make([]runtime.BoxConstraints, childCount)
    for i, child := range children {
        childConstraints[i] = measurer.MeasureChild(child, constraints)
    }

    // Calculate own size based on flex props
    // ... (copy flex algorithm from ElementVNode.MeasureLayout)

    return NewLayoutMeasurement(size, childConstraints)
}
```

### Phase 5.3: Implement Content Measurement

**File**: `runtime/ui/fiber_content.go` (new file)

**Purpose**: Implement content measurement for different fiber types

```go
package ui

import (
    "github.com/wwsheng009/mint/runtime"
)

// MeasureContent measures natural size of fiber's content
// Based on VNode.Type, delegates to appropriate measurement
func (f *Fiber) MeasureContent(constraints runtime.BoxConstraints) runtime.Size {
    if f == nil {
        return runtime.Size{}
    }

    switch f.Type {
    case VNodeText:
        return f.measureTextContent(constraints)
    case VNodeElement:
        // For elements, size comes from layout
        return runtime.Size{Width: 0, Height: 0}
    case VNodeComponent:
        // For components, size comes from layout
        return runtime.Size{Width: 0, Height: 0}
    default:
        return runtime.Size{}
    }
}

// measureTextContent measures text content
func (f *Fiber) measureTextContent(constraints runtime.BoxConstraints) runtime.Size {
    // Get text from memoizedState (set by completeWorkText)
    text, ok := f.MemoizedState.(string)
    if !ok {
        return runtime.Size{}
    }

    // Measure text using rune width
    // ... (copy from TextVNode.Measure())
}
```

### Phase 5.4: Add buildComputedBoxFromFiber

**File**: `runtime/compute/engine.go`

**Purpose**: New method that only uses Fiber

```go
// buildComputedBoxFromFiber creates ComputedBox from Fiber tree only
// This is Phase 6 implementation that removes VNode dependency
func (e *Engine) buildComputedBoxFromFiber(
    fiber *rtui.Fiber,
    parent *ComputedBox,
    constraints runtime.BoxConstraints,
) *ComputedBox {
    if fiber == nil {
        return nil
    }

    box := &ComputedBox{
        VNode:        fiber.VNode, // Keep for now, remove later
        Parent:       parent,
        Box:          runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
        NodeID:       fiber.NodeID,
        Layer:        fiber.Layer,
    }

    // Measure content
    contentSize := fiber.MeasureContent(constraints)

    // Measure layout (flex)
    measurement := fiber.MeasureLayout(e, constraints)

    // Set box size
    box.Box.Width = measurement.Size.Width
    box.Box.Height = measurement.Size.Height

    // Build children
    children := fiber.GetChildFibers()
    box.Children = make([]*ComputedBox, 0, len(children))

    for i, childFiber := range children {
        childBox := e.buildComputedBoxFromFiber(childFiber, box, measurement.ChildConstraints[i])
        if childBox != nil {
            box.Children = append(box.Children, childBox)
        }
    }

    return box
}
```

### Phase 5.5: Update LayoutFiber

**File**: `runtime/compute/engine.go`

**Change**: Use buildComputedBoxFromFiber instead of buildComputedBox

```go
func (e *Engine) LayoutFiber(root *rtui.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    // ...

    // OLD: Uses VNode
    // rootBox := e.layoutFiber(root, constraints, 0)

    // NEW: Uses only Fiber
    rootBox := e.buildComputedBoxFromFiber(root, nil, constraints)

    // ...
}
```

---

## Migration Steps

1. ✅ **Phase 5.1**: Add GetChildFibers(), GetChildCount() to Fiber
2. **Phase 5.2**: Create fiber_layout.go with MeasureLayout()
3. **Phase 5.3**: Create fiber_content.go with MeasureContent()
4. **Phase 5.4**: Add buildComputedBoxFromFiber() to Engine
5. **Phase 5.5**: Update LayoutFiber() to use new method
6. **Phase 5.6**: Test Fiber-first layout
7. **Phase 5.7**: Remove old buildComputedBox() (deprecated)
8. **Phase 6**: Complete - VNode dependency removed from layout

---

## Testing Strategy

### Unit Tests

```go
func TestFiberGetChildFibers(t *testing.T) {
    root := &Fiber{}
    child1 := &Fiber{}
    child2 := &Fiber{}

    root.Child = child1
    child1.Sibling = child2

    children := root.GetChildFibers()
    if len(children) != 2 {
        t.Errorf("Expected 2 children, got %d", len(children))
    }
}

func TestFiberMeasureLayout(t *testing.T) {
    // Test flex layout on Fiber tree
    // Similar to TestElementMeasureLayout but uses Fiber
}
```

### Integration Tests

```go
func TestLayoutFiberOnly(t *testing.T) {
    fiber := createTestFiberTree()
    engine := NewEngine()

    layout, err := engine.LayoutFiber(fiber, constraints)
    if err != nil {
        t.Fatal(err)
    }

    // Verify layout worked without VNode
    if layout.Root == nil {
        t.Error("Layout root is nil")
    }
}
```

---

## Benefits

1. **Simpler Architecture**: Single source of truth (Fiber)
2. **Better Performance**: No VNode→Fiber matching during layout
3. **Clearer Separation**: Layout is purely about structure, not content
4. **Easier Testing**: Fiber trees are easier to create in tests
5. **Future-Proof**: Ready for async rendering (Fiber can be queued)

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Duplication of code | Maintenance burden | Share code between VNode and Fiber initially |
| Breaking changes | Tests fail | Keep old methods during transition |
| Performance regression | Layout slower | Benchmark before/after |
| Missed edge cases | Layout incorrect | Comprehensive test coverage |

---

## Timeline

- **Week 1**: Phase 5.1-5.2 (Fiber structure methods)
- **Week 2**: Phase 5.3-5.4 (Content measurement)
- **Week 3**: Phase 5.5-5.6 (Integration and testing)
- **Week 4**: Phase 5.7-6 (Cleanup and validation)

---

**Status**: Planning Phase
**Next Step**: Start Phase 5.1 - Add GetChildFibers() to Fiber
