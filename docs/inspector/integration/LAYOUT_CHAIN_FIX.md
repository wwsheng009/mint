# Inspector Layout Chain Fix

## Problem

After adding `Flex(1)` to `treePreview` in the Inspector, the entire interface stopped displaying anything. The layout system was failing silently.

## Root Cause Analysis

### Layout Constraint Chain Breakdown

The layout constraint propagation was broken at the Tabs component:

```
Bordered(Height=25)
  → VStack(content) - bounded height (23) ✓
    → TabsComponent(Height=21 prop)
      → tabBar (height=1)
      → content (from buildElementsTabContent) - SHOULD have bounded height (20) ✗
        → VStack
          → header
          → selectedInfo
          → treePreview(Flex=1) - requires bounded height ✗
          → instructions
```

### Key Issue: Tabs Component Ignored Height Prop

**File**: `components/navigation/tabs.go`

The `TabsVNode.Measure()` method did not check or use its own `height` prop:

```go
// BEFORE - Missing prop handling
func (t *TabsVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // ...
    // ❌ Never checked t.Props()["height"]
    // ❌ Never used explicit height to constrain content
}
```

When `tabsBuilder.Height(21)` was set, the Tabs component:
1. ✅ Stored the height in props
2. ❌ Never used it in `Measure()`
3. ❌ Passed unbounded constraints to content
4. ❌ `treePreview.Flex(1)` failed - no bounded height from parent

## Solution

### 1. Fix Tabs Component to Use Height Prop

**File**: `components/navigation/tabs.go`

```go
// AFTER - Uses height prop
func (t *TabsVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // Check for explicit width/height props first
    props := t.Props()
    explicitHeight := 0
    hasHeightProp := false

    if props != nil {
        if h, ok := props["height"].(int); ok && h > 0 {
            explicitHeight = h
            hasHeightProp = true
        }
    }

    // ... measure content ...

    // Determine max height for content
    maxContentHeight := runtime.Infinity
    if hasHeightProp {
        // Use explicit height prop (subtract tab bar height)
        maxContentHeight = explicitHeight - 1
    } else if constraints.HasBoundedHeight() {
        // Use parent constraints
        maxContentHeight = constraints.MaxHeight - 1
    }

    // Create bounded constraints for content
    contentConstraints := runtime.BoxConstraints{
        MaxHeight: maxContentHeight,  // ✓ Now bounded!
    }

    // Apply explicit height to final result
    if hasHeightProp {
        height = explicitHeight
    }

    return runtime.Size{Width: totalWidth, Height: height}
}
```

### 2. Add Defensive Code to Layout Engine

**File**: `runtime/compute/engine.go`

Added warnings and minimum height guarantees:

```go
// DEFENSIVE: Warn if flex children can't be distributed
if len(flexChildren) > 0 && !constraints.HasBoundedHeight() {
    if e.debug || os.Getenv("TUI_LAYOUT_WARNINGS") == "true" {
        fmt.Fprintf(os.Stderr, "[Layout WARNING] VStack has flex children but no bounded height constraint.\n")
        fmt.Fprintf(os.Stderr, "  → Flex children will be measured with natural height.\n")
        fmt.Fprintf(os.Stderr, "  → To fix: Add .Height(n) to parent VStack or use .FillHeight()\n")
    }
}

// DEFENSIVE: Ensure minimum height to prevent invisible components
if totalHeight == 0 && len(children) > 0 {
    if e.debug || os.Getenv("TUI_LAYOUT_WARNINGS") == "true" {
        fmt.Fprintf(os.Stderr, "[Layout WARNING] VStack with %d children computed height=0.\n", len(children))
        fmt.Fprintf(os.Stderr, "  → Using minimum height of 1 to ensure visibility.\n")
    }
    totalHeight = 1
}
```

## Verification

### Test Command
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

### Expected Results
✅ Inspector overlay displays correctly
✅ Tabs component shows all tabs
✅ Tree view is visible with navigation
✅ Selected/focused info displays above tree
✅ Instructions display at bottom
✅ No content overflow
✅ No empty/blank areas

## Layout Constraint Propagation (After Fix)

```
Bordered(Height=25)
  → measureBordered creates innerConstraints.MaxHeight = 23
    → VStack(content) receives bounded height (23) ✓
      → tabsComponent with Height(21) prop
        → Tabs.Measure() checks props ✓
        → Creates contentConstraints.MaxHeight = 21 - 1 = 20 ✓
          → VStack (from buildElementsTabContent) receives bounded height (20) ✓
            → header (measured naturally)
            → selectedInfo (measured naturally)
            → treePreview(Flex=1) - parent has bounded height! ✓
              → Uses remaining space: 20 - 3 - 4 - 6 = 7 lines ✓
            → instructions (measured naturally: 6 lines)
```

## Lessons Learned

### 1. Props Must Be Used in Measure()
Setting a prop (like `Height()`) is useless unless the component's `Measure()` method checks and uses it.

### 2. Constraint Propagation is Critical
Every component in the chain must correctly propagate constraints:
- Parent → bounded constraint → Child
- Child uses bounded constraint to calculate size
- Child passes constraints to its children

### 3. Flex Layout Requires Bounded Height
`Flex(1)` only works when the parent container has bounded height. The layout engine checks:
```go
if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
    // Distribute space
}
```

### 4. Defensive Programming Helps
Adding warnings and minimum guarantees prevents silent failures and helps debugging.

## Related Files

- `components/navigation/tabs.go` - Fixed to use height prop
- `runtime/compute/engine.go` - Added defensive warnings
- `internal/inspector/standalone_inspector.go` - Uses Flex(1) correctly

## Environment Variables

### TUI_LAYOUT_WARNINGS
Show layout warnings during development:
```bash
TUI_LAYOUT_WARNINGS=true go run main.go
```

### TUI_LAYOUT_DEBUG
Enable detailed layout debug output:
```bash
TUI_LAYOUT_DEBUG=true go run main.go
```
