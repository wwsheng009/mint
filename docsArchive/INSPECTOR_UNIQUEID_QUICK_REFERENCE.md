# Inspector UniqueID: Before vs After

## The Problem

**Counter-based IDs (BROKEN)**:
```
Build 1: node-1, node-2, node-3, node-4
Rebuild: node-1, node-2, node-3, node-4  ← Different nodes!
```

When VNodes are recreated, the counter restarts, so:
- Old `node-3` = "Text component"
- New `node-3` = "Button component" (different!)

## The Solution

**Path-based IDs (STABLE)**:
```
Build 1: vstack, vstack.text, vstack.button
Rebuild: vstack, vstack.text, vstack.button  ← Same nodes!
```

The path is deterministic:
- Same node structure → same path → same UniqueID
- Expand state persists across rebuilds

## Code Comparison

### Before (Unstable)

```go
type TreeView struct {
    nextNodeID   int64  // Problem: resets on rebuild
}

func (tv *TreeView) buildTree(...) {
    node := &TreeNode{
        UniqueID: fmt.Sprintf("node-%d", tv.nextNodeID),
    }
    tv.nextNodeID++  // Different rebuild = different IDs
}
```

### After (Stable)

```go
type TreeView struct {
    // No counter needed!
}

func (tv *TreeView) buildTree(vnode ui.VNode, parent *TreeNode, level int, path string) {
    // Generate stable path
    nodePath := path + "." + getSimpleType(vnode)

    // Use path as UniqueID
    uniqueID := nodePath
    if keyer, ok := vnode.(interface{ Key() string }); ok {
        if key := keyer.Key(); key != "" {
            uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
        }
    }

    node := &TreeNode{
        UniqueID: uniqueID,  // Same path = same ID
    }
}
```

## Status Line Comparison

### Before
```
🔍 Focus: #3 [ID: node-136] → ElementVNode
```
- What is "node-136"? → Need to look it up
- Changes every rebuild → Confusing

### After
```
🔍 Focus: #3 [ID: vstack.text] → ElementVNode
```
- It's a Text node inside a VStack → Clear!
- Stable across rebuilds → Reliable

## Expand State Persistence

### Before (BROKEN)
```go
// First render
expanded["node-5"] = true  // User expanded this node

// Second render (VNode recreated)
expanded["node-5"] = ???   // node-5 now refers to different node!
// User's expand state is lost!
```

### After (WORKING)
```go
// First render
expanded["vstack.layoutnode"] = true  // User expanded this node

// Second render (VNode recreated)
expanded["vstack.layoutnode"] = true  // Same path = same node
// Expand state persists! ✓
```

## Test Results

**Unit Tests** (All Pass):
```
TestTreeViewUniqueIDLookup      ✓ Verifies path-based IDs work
TestTreeViewExpandCollapse      ✓ Verifies expand state persists
TestTreeViewPathConsistency     ✓ Verifies IDs stable across operations
```

**Integration Test** (Pass):
```
Status line shows: [ID: vstack.text]  ← Path-based, not node-136
```

## Why Not Use State Functions?

**User Question**: "use the stateFunction()?"

**Answer**: Inspector can't use `UseState()` because:
1. Inspector runs **outside** component lifecycle
2. `UseState()` requires component context
3. Inspector needs **cross-tree** state, not per-component

**Path-based IDs are better**:
- ✓ Stable across rebuilds
- ✓ No component context needed
- ✓ Readable and debuggable
- ✓ Works with any VNode structure

## Key Files Changed

| File | Change |
|------|--------|
| `tree_view.go:67-96` | Path-based UniqueID generation |
| `tree_view.go:24-34` | Removed `nextNodeID` counter |
| `tree_expand_test.go` | Updated tests to use path-based IDs |
| `standalone_inspector.go:546` | Status line shows path-based IDs |

## Verification

Run the demo to see stable IDs:
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

**What you'll see**:
- Navigate with ↑/↓ → Status shows `[ID: vstack.text]`
- Press E to expand → Correct node expands
- Navigate away and back → Same ID shows
- Tree rebuilds → IDs remain stable!

## Summary

✅ **Problem**: Counter-based IDs changed every rebuild
✅ **Solution**: Use path as UniqueID (deterministic, stable)
✅ **Result**: Expand state persists, correct node expands

The Inspector now properly tracks expand/collapse state across VNode rebuilds!
