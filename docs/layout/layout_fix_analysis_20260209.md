# Mint Layout Engine Repair Analysis Report
**Date:** 2026-02-09
**Module:** Layout Engine / Inspector
**Severity:** Critical (Layout correctness / Infinite rendering)

## 1. Executive Summary

This report details the investigation and resolution of a critical layout bug where `TreeView` components within the `Inspector` overlay failed to respect parent container height constraints. The issue manifested as the TreeView rendering all rows (e.g., 34 rows) instead of virtualizing based on the available viewport (e.g., 7 rows), causing UI overflow and rendering artifacts.

The root cause was a combination of **broken constraint propagation chains** in container components (`VStack`, `Bordered`) and a **logical race condition** within the core Layout Engine (`runtime/compute`).

## 2. Problem Description

### Observed Behavior
- **Context**: The `StandaloneInspector` uses a fixed-height overlay (25 lines).
- **Component Structure**: `Bordered` -> `VStack` (Outer) -> `Tabs` -> `VStack` (Inner) -> `TreeView`.
- **Symptom**: The `TreeView` received unbounded height constraints (`Infinity`) despite the parent `VStack` having a fixed height. Even when constraints were manually forced, the Engine continued to render stale children nodes.

### Expected Behavior
- The `Bordered` container should constrain its children to its set height.
- The `VStack` should distribute available space to `flex` children (`TreeView`).
- The `TreeView` should receive a specific `MaxHeight` (e.g., 7).
- The `TreeView` should only render 7 visible rows (Virtual Scrolling).

## 3. Root Cause Analysis

The investigation revealed three distinct layers of failure:

### A. Constraint Chain Breakage (The "Why it was Infinity")
Constraint propagation logic was missing in two key areas:
1.  **`VStack` / `HStack`**: In `runtime/ui/layout_measurement.go`, the measurement logic calculated size based on children but ignored explicit `Height(n)` or `Width(n)` properties set by the user. This prevented the "downward" flow of constraints.
2.  **`BorderedNode`**: In `runtime/ui/layout.go`, `BorderedNode.Measure` relied solely on incoming constraints from *its* parent, ignoring its own `Width`/`Height` props. Since the root overlay often starts with loose constraints, the Bordered node failed to clamp them before passing them down.

### B. The Engine Logical Race Condition (The "Why it still rendered 34 lines")
Even after fixing the constraints (TreeView correctly received `MaxHeight: 7`), the Engine still rendered all 34 lines.

**Mechanism of Failure**:
1.  **Engine Layout Start**: `Engine.buildComputedBoxWithSize` is called for `TreeView`.
2.  **Stale Cache**: The Engine captured `vnodeChildren := vnode.Children()` **before** measuring the node. At this point, `TreeView` contains 34 lines (from the previous frame or initial state).
3.  **Measurement**: The Engine calls `e.measureVNode()`, which calls `TreeView.Measure()`.
4.  **State Update**: Inside `TreeView.Measure()`, the component detects the new `MaxHeight: 7`, calculates `viewportHeight`, and calls `regenerateDisplay()`. This updates the internal children slice to contain only 7 items.
5.  **Stale Iteration**: The Engine resumes execution of `buildComputedBoxWithSize` but iterates over the **stale** `vnodeChildren` variable (size 34) instead of the updated children list.

### C. Robustness Issues (Interface Assertion)
During debugging, it was observed that `vnode.(runtime.Measurable)` assertions might fail in certain build/link contexts or when `Engine` logic paths (like `isBordered`) preempted the interface check.

## 4. Code Implementation Details

### Fix 1: Engine Logic Update (Critical)
**File**: `runtime/compute/engine.go`
**Change**: Moved the retrieval of children nodes to *after* the measurement phase.

```go
func (e *Engine) buildComputedBoxWithSize(...) {
    // 1. Measure first (this might mutate the VNode's children)
    size := e.measureVNode(vnode, constraints)
    
    // ... apply size ...

    // 2. REFRESH children list
    // Crucial for components like TreeView that implement virtual scrolling
    // during the Measure phase.
    vnodeChildren := vnode.Children() 

    // 3. Build children boxes
    box.Children = make([]*ComputedBox, 0, len(vnodeChildren))
    for _, child := range vnodeChildren {
        // ...
    }
}
```

### Fix 2: Constraint Propagation in Containers
**File**: `runtime/ui/layout_measurement.go` (VStack/HStack) and `runtime/ui/layout.go` (BorderedNode)
**Change**: Added explicit checks for `props["width"]` and `props["height"]`.

```go
// Example for VStack
if height, ok := props["height"].(int); ok && height > 0 {
    // Force the MaxHeight constraint to the explicit height
    constraints.MaxHeight = height
    // Ensure logical consistency
    if constraints.MinHeight > height {
        constraints.MinHeight = height
    }
}
```

### Fix 3: Robustness and Fallbacks
**File**: `runtime/compute/engine.go`
**Change**: Added a fallback mechanism using Reflection for the "treeview" tag. This ensures that even if interface assertions fail (due to subtle type issues), the `Measure` method is still invoked.

```go
case "treeview":
    // Try interface assertion
    if measurable, ok := vnode.(runtime.Measurable); ok {
        return measurable.Measure(constraints)
    }
    // Fallback to reflection
    if method := reflect.ValueOf(vnode).MethodByName("Measure"); method.IsValid() {
        // ... invoke via reflect ...
    }
```

## 5. Debugging Process & Tools Created

To isolate these issues, we developed a specialized debugging suite:

1.  **LayoutAnalyzer**: A tool in `internal/inspector/layout_analyzer.go` that can capture a snapshot of the VNode tree, including calculated Sizes and Constraints, and export it to text.
2.  **Real-time Tracing (`TUI_LAYOUT_DEBUG`)**:
    *   Enhanced `runtime/compute/engine.go` to support a `TUI_LAYOUT_DEBUG=true` environment variable.
    *   Prints an indented trace of `[Layout.ENTER]` (Constraints) and `[Layout.LEAVE]` (Size) for every node.
    *   This was crucial in proving that `TreeView` received the correct constraints but the Engine ignored the updated children.
3.  **Reproduction Scripts**:
    *   `inspect_self.go`: A headless script that constructs the Inspector UI tree and runs the Layout Engine against it without needing the full TUI terminal loop.

## 6. Validation Results

After applying the fixes:
1.  **Log Verification**: The trace log now shows `TreeView` receiving `H:7` constraints, followed immediately by only 7 `[Layout.ENTER]` calls for its children (instead of 34).
2.  **Visual Verification**: The Inspector overlay correctly fits within the 25-line window, and scrolling works as expected without overflow.

---
*Report generated by Opencode*
