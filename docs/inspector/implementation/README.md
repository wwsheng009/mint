# Inspector Implementation Documentation

This directory contains documents detailing specific implementations, fixes, and technical solutions for Inspector components.

## Core Components

### TreeView Implementation

#### [INSPECTOR_FLEX_LAYOUT_IMPLEMENTATION.md](INSPECTOR_FLEX_LAYOUT_IMPLEMENTATION.md)
Implementation of flex layout for the TreeView component to support proper dynamic sizing.

**Key Changes:**
- Replaced fixed-height TextVNode with flex-based layout
- Implemented grow/shrink behavior
- Auto-sizing based on content

#### [INSPECTOR_FLEX_AUTOSIZE_IMPLEMENTATION.md](INSPECTOR_FLEX_AUTOSIZE_IMPLEMENTATION.md)
Details on auto-sizing implementation for Inspector components.

**Key Changes:**
- Dynamic height calculation
- Flex-based constraints
- Content-aware sizing

#### [TREEVIEW_VIRTUAL_SCROLL_IMPLEMENTATION.md](TREEVIEW_VIRTUAL_SCROLL_IMPLEMENTATION.md)
Virtual scrolling implementation for handling large trees efficiently.

**Features:**
- Viewport-based rendering
- Scroll position tracking
- Performance optimization

### UniqueID System

#### [INSPECTOR_UNIQUEID_FINAL_SOLUTION.md](/docsArchive/INSPECTOR_UNIQUEID_FINAL_SOLUTION.md)
Complete solution for UniqueID collision issues in the TreeView.

**Problem:**
- Multiple nodes had same ID (e.g., `vstack.bordered.vstack[0]`)
- TreeView couldn't distinguish between LayoutNode and BorderNode
- Expand/collapse failed due to ID collisions

**Solution:**
- Priority 1: User-defined key (if provided)
- Priority 2: VNode pointer address (unique per instance)
- Format: `path[key]` or `path[index]@0x...`

#### [INSPECTOR_UNIQUEID_IMPLEMENTATION.md](/docsArchive/INSPECTOR_UNIQUEID_IMPLEMENTATION.md)
Initial implementation of the UniqueID system.

#### [INSPECTOR_UNIQUEID_QUICK_REFERENCE.md](/docsArchive/INSPECTOR_UNIQUEID_QUICK_REFERENCE.md)
Quick reference guide for UniqueID format and usage.

### Bug Fixes

#### [INSPECTOR_POINTER_ID_FIX.md](/docsArchive/INSPECTOR_POINTER_ID_FIX.md)
Fix for using VNode pointer addresses as fallback IDs.

**Before:**
```go
uniqueID = fmt.Sprintf("%s[%d]", nodePath, index)
// Collisions: same path + same index
```

**After:**
```go
uniqueID = fmt.Sprintf("%s[%d]@%p", nodePath, index, vnode)
// Unique: pointer address differs per instance
```

#### [INSPECTOR_SETLAYER_BUG_FIX.md](/docsArchive/INSPECTOR_SETLAYER_BUG_FIX.md)
Critical bug fix: SetProps() replaces entire props map, not merging.

**Problem:**
```go
// WRONG: Layer is lost!
overlay.SetLayer(rtui.LayerInspector)
overlay.SetProps(rtui.Props{"x": 80, "y": 5})
// SetProps() REPLACES props, layer property is gone
```

**Solution:**
```go
// CORRECT: Set props BEFORE setting layer
overlay.SetProps(rtui.Props{"x": 80, "y": 5, "width": 80, "height": 25})
overlay.SetLayer(rtui.LayerInspector)
```

#### [INSPECTOR_POSITION_FIX.md](/docsArchive/INSPECTOR_POSITION_FIX.md)
Fix for Inspector default positioning being off-screen.

**Problem:**
```go
floatX: 80  // Off-screen for 80-column terminal!
```

**Solution:**
```go
floatX: 0   // Left edge, always visible
floatY: 0   // Top edge
```

#### [INSPECTOR_HARDCODED_BORDER_FIX.md](/docsArchive/INSPECTOR_HARDCODED_BORDER_FIX.md)
Fix for hardcoded border characters in components.

#### [INSPECTOR_PATH_INDEX_FIX.md](/docsArchive/INSPECTOR_PATH_INDEX_FIX.md)
Fix for path index calculation in UniqueID generation.

### Diagnostics

#### [TREEVIEW_DISPLAY_DIAGNOSIS.md](/docsArchive/TREEVIEW_DISPLAY_DIAGNOSIS.md)
Diagnostic analysis of TreeView display issues.

## Implementation Timeline

### Phase 1: Initial Implementation
- Basic TreeView with text-based display
- Fixed-height components
- Index-based UniqueIDs

### Phase 2: Bug Discovery
- Identified UniqueID collisions
- Found SetProps/SetLayer bug
- Discovered positioning issues

### Phase 3: Flex Layout
- Implemented flex-based sizing
- Dynamic height calculation
- Auto-grow behavior

### Phase 4: UniqueID Overhaul
- Added pointer-based fallback
- User key priority system
- Collision resolution

### Phase 5: Polish
- Position fixes
- Border rendering fixes
- Performance optimization

## Code Examples

### Creating a UniqueID

```go
func (tv *TreeView) buildTree(vnode ui.VNode, parent *TreeNode, level int, path string, index int) *TreeNode {
    // Build path
    var nodePath string
    if parent == nil {
        nodePath = "root"
    } else {
        nodePath = fmt.Sprintf("%s.%s", parent.Path, vnode.Type())
    }

    // Generate UniqueID
    var uniqueID string

    // Priority 1: User-defined key
    if keyer, ok := vnode.(interface{ Key() string }); ok {
        if key := keyer.Key(); key != "" {
            uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
        }
    }

    // Priority 2: Pointer address (fallback)
    if uniqueID == "" {
        uniqueID = fmt.Sprintf("%s[%d]@%p", nodePath, index, vnode)
    }

    // Create tree node
    return &TreeNode{
        ID:       uniqueID,
        Path:     nodePath,
        VNode:    vnode,
        Children: buildChildren(vnode, level+1, nodePath),
    }
}
```

### Proper Props and Layer Setting

```go
// Step 1: Create inspector content
inspectorContent := inspector.RenderContent()

// Step 2: Set ALL props (including positioning)
inspectorContent.SetProps(rtui.Props{
    "x":      x,
    "y":      y,
    "width":  width,
    "height": height,
})

// Step 3: Set layer AFTER props
inspectorContent.SetLayer(rtui.LayerInspector)

// Step 4: Wrap in Fragment
return rtui.Fragment(appContent, inspectorContent)
```

## Testing

### Unit Tests
```bash
# Test UniqueID generation
go test -v ./internal/inspector -run TestTreeView

# Test flex layout
go test -v ./internal/inspector -run TestFlex

# Test pointer-based IDs
go test -v ./internal/inspector -run TestPointerID
```

### Integration Tests
```bash
# Test TreeView rendering
go test -v ./internal/inspector -run TestTreeViewRender

# Test Inspector overlay
go test -v ./internal/inspector -run TestInspectorOverlay
```

## Performance Considerations

### UniqueID Generation
- Pointer addresses are fast to obtain
- No allocation for user keys (string reuse)
- Path building is O(depth) where depth is tree depth

### TreeView Rendering
- Virtual scrolling reduces render cost
- Only visible nodes are rendered
- Flex layout is O(n) where n is visible nodes

### Memory Usage
- TreeView maintains full tree in memory
- Each TreeNode stores VNode reference (not copy)
- Pointer addresses don't allocate

## Best Practices

### 1. Always Set Props Before Layer
```go
// Good
vnode.SetProps(props)
vnode.SetLayer(layer)

// Bad
vnode.SetLayer(layer)
vnode.SetProps(props)  // Loses layer!
```

### 2. Use Pointer-Based IDs as Fallback
```go
// Good: User key preferred, pointer fallback
if key := vnode.Key(); key != "" {
    uniqueID = fmt.Sprintf("%s[%s]", path, key)
} else {
    uniqueID = fmt.Sprintf("%s[%d]@%p", path, index, vnode)
}

// Bad: Index only (collisions)
uniqueID = fmt.Sprintf("%s[%d]", path, index)
```

### 3. Leverage Flex Layout for Sizing
```go
// Good: Auto-sized to content
ui.VStack(
    content,
).Grow(1).Shrink(1)

// Bad: Fixed height
ui.VStack(
    content,
).Height(25)  // May overflow or be too small
```

## Related Documentation

- [Architecture Overview](/docsArchive/architecture/)
- [Investigation Analysis](/docsArchive/investigation/)
- [Hook System](../../render/hook/README.md)
