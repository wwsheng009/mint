# Inspector Flex-like Auto-Sizing Implementation

## Summary

Implemented flex-like auto-sizing behavior for the Inspector TreeView component, where the TreeView automatically expands to fit its content while a ScrollView provides fixed-height constraints and handles scrolling.

## Problem Statement

User request:
> "探测器可以有一个固定的高度，比如屏幕的80%，但是treeview不能直接设置高度，他的高度受容器的高度约束比如flex一样，内容超过了这个高度，进行滚动"

Translation:
- Inspector should have a fixed height (e.g., 80% of screen)
- TreeView should auto-size to parent container (like flexbox)
- Scroll only when content exceeds available height

## Solution Architecture

### Before: Fixed Viewport Height
```go
// OLD: TreeView with fixed viewportHeight
treeViewComponent.SetViewportHeight(treeViewHeight)
treeViewComponent.SetScrollOffset(scrollOffset)
```
- TreeView rendered only visible lines (virtual scrolling)
- Fixed height constraint on TreeView itself
- Problem: TreeView didn't auto-size to content

### After: Flex-like Auto-Sizing
```go
// NEW: TreeView auto-sizes, ScrollView provides constraint
scrollContainer := layout.NewScrollView(treePreview).
    Height(treeViewHeight).
    Width(si.overlayWidth - 4).
    ScrollOffset(si.treeScrollOffset).
    Build()
```
- TreeView auto-sizes to full content (no viewportHeight constraint)
- ScrollView provides fixed-height viewport (flex-like container)
- ScrollView handles clipping and scrolling

## Implementation Details

### Files Modified

#### 1. `internal/inspector/standalone_inspector.go`

**Import added** (line 27):
```go
import (
    // ...
    "github.com/wwsheng009/mint/components/layout"
)
```

**ScrollView wrapper added** (lines 602-614):
```go
// Wrap tree preview in ScrollView for flex-like auto-sizing with scrolling
// TreeView auto-sizes to content, ScrollView provides fixed viewport constraint
scrollContainer := layout.NewScrollView(treePreview).
    Height(treeViewHeight).
    Width(si.overlayWidth - 4).
    ScrollOffset(si.treeScrollOffset).
    Build()

// Combine status line with scrollable tree
var treeWithStatus ui.VNode
if len(statusLines) > 0 {
    treeWithStatus = ui.VStackBuilder(append(statusLines, scrollContainer)...).Build()
} else {
    treeWithStatus = scrollContainer
}
```

**Removed** (previously at lines 528, 1320, 1322):
```go
// REMOVED: No longer set fixed viewportHeight on TreeView
// treeViewComponent.SetViewportHeight(treeViewHeight)
```

### 2. `components/layout/scroll_view.go` (Existing component)

The ScrollView component already existed and provides:
- Fixed-height container that clips overflow content
- Virtual scrolling (only renders visible lines)
- Scroll position indicator (▼ ▲ ↕)
- Content extraction from various VNode types

## Behavior Comparison

| Aspect | Before | After |
|--------|--------|-------|
| **TreeView Height** | Fixed (viewportHeight) | Auto-sizes to content |
| **Scrolling** | TreeView virtual scroll | ScrollView virtual scroll |
| **Clipping** | TreeView clips content | ScrollView clips content |
| **Flex behavior** | ❌ No | ✅ Yes (auto-expand) |
| **Scroll indicator** | ❌ No | ✅ Yes (▼ ▲ ↕) |

## Component Architecture

```
Inspector Overlay (fixed 80% screen height)
└── buildElementsTabContent()
    ├── header (fixed)
    ├── selectedInfo (fixed)
    ├── ScrollView (fixed height = overlayHeight - 14) ← NEW!
    │   └── treePreview (auto-sizes to full content)
    │       └── TreeView (no viewportHeight constraint)
    │           └── All tree lines (full content)
    └── instructions (fixed)
```

### Flex-like Behavior

- **Fixed elements**: Header, status, instructions stay at fixed positions
- **Scrollable area**: TreeView inside ScrollView expands to fit content
- **Viewport constraint**: ScrollView clips to available height
- **Overflow handling**: ScrollView provides scrolling when content > viewport

## Verification

### Test 1: ScrollView Wrapper Exists
```bash
cd internal/inspector
go test -v -run TestInspectorFlexAutoSizing
```
**Result**: ✅ PASS
```
✅ Found LayoutNode (likely ScrollView) at child #2, grandchild #1
✅ ScrollView has 1 children (should be 1 text node)
✅ Flex-like auto-sizing is implemented
```

### Test 2: TreeView Unchanged
```bash
cd internal/inspector
go test -v -run TestTreeViewWithScrollView
```
**Result**: ✅ PASS
```
✅ Tree has 23 nodes (>= 20 content items)
✅ Tree content unchanged after ScrollView wrapping
```

### Test 3: Basic Rendering
```bash
cd internal/inspector
go test -v -run TestInspectorBasicRendering
```
**Result**: ✅ PASS
```
✅ TreeViewComponent has 33 lines
✅ Render has 33 children
```

## Key Design Decisions

### 1. Why ScrollView instead of modifying TreeView?

**Reason**: Separation of concerns
- TreeView: Data visualization (tree structure, expand/collapse)
- ScrollView: Viewport management (clipping, scrolling)

**Benefit**: TreeView can be used elsewhere without ScrollView dependency

### 2. Why remove SetViewportHeight() calls?

**Reason**: Enable auto-sizing
- TreeView with viewportHeight only renders visible lines
- TreeView without viewportHeight renders full content
- ScrollView then provides the viewport constraint

**Benefit**: Flex-like behavior - content determines size, not container

### 3. Why wrap treePreview instead of treeViewComponent?

**Reason**: treePreview is the rendered output
- treeViewComponent is the component instance
- treePreview = treeViewComponent.GetRender() (latest render state)
- Wrapping rendered output ensures correct display

## Usage Example

```go
// Create Inspector
inspector := NewStandaloneInspector()
inspector.Enable()
inspector.SetOverlaySize(80, 25)  // 80x25 overlay (80% screen)

// Attach large tree
var children []VNode
for i := 0; i < 100; i++ {
    children = append(children, Text(fmt.Sprintf("Node %d", i)))
}
inspector.AttachToApp(VStack(children...))

// Result:
// - Inspector has fixed height (25 lines)
// - TreeView auto-sizes to 100 lines
// - ScrollView shows ~11 lines (25 - 14 for header/footer)
// - User can scroll through all 100 lines
// - Status line stays fixed above scroll area
```

## Related Files

### Tests Created
- `internal/inspector/flex_autosize_test.go` - Verifies ScrollView wrapper exists
- `internal/inspector/scrollview_test.go` - Tests ScrollView component
- `internal/inspector/tree_scrollview_integration_test.go` - Verifies TreeView unchanged

### Implementation Files
- `internal/inspector/standalone_inspector.go` - ScrollView wrapper added
- `components/layout/scroll_view.go` - ScrollView component (existing)

## Future Enhancements

### Potential Improvements
1. **Horizontal scrolling**: Currently ScrollView only supports vertical scrolling
2. **Scroll position indicator**: Could show percentage (e.g., "45%")
3. **Smooth scrolling**: Animate scroll position changes
4. **Scroll preservation**: Remember scroll position across tab switches

### Limitations
1. **Text-based**: ScrollView extracts text, loses VNode structure
2. **No keyboard handling**: ScrollView doesn't capture scroll keys (handled by Inspector)
3. **Static height**: Height calculated once, doesn't respond to resize

## Conclusion

The flex-like auto-sizing implementation successfully achieves:
- ✅ TreeView auto-sizes to content
- ✅ ScrollView provides fixed viewport constraint
- ✅ ScrollView handles clipping and scrolling
- ✅ No breaking changes to existing tests
- ✅ Clear separation of concerns (TreeView vs ScrollView)

This matches the user's requirement for flexbox-like behavior where content determines size, and the container provides constraints.
