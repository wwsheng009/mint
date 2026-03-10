# Inspector UniqueID Fix: Path + Index Solution

## Problem Statement

The Inspector TreeView's expand/collapse feature was broken because UniqueIDs were colliding, causing the wrong nodes to expand when pressing E.

**Specific Issue**: `vstack.bordered.hstack[0]` appeared multiple times - nodes with the same structure but at different positions had the same ID.

**Critical Bug in Initial Attempt**: Using pointer addresses (`@0x...`) doesn't work because VNodes are recreated on every render, so the addresses change every time!

## Root Cause

The collision happened because:

1. **Path didn't include parent index** - Both `vstack[0]` and `vstack[1]` had children with path `vstack.text`
2. **Example collision**:
   ```
   First VStack child:  vstack.text[0]
   Second VStack child: vstack.text[0]  // COLLISION!
   ```

## Solution: Path with Index

Include the **parent's index** in the path, making every path unique:

### Implementation

**File**: `internal/inspector/tree_view.go:65-71`

```go
// Generate path for this node
// Include parent's index in path to ensure uniqueness
var nodePath string
if path == "" {
    nodePath = getSimpleType(vnode)
} else {
    nodePath = fmt.Sprintf("%s[%d].%s", path, index, getSimpleType(vnode))
}
```

**File**: `internal/inspector/tree_view.go:76-93`

```go
// Generate UniqueID following React's philosophy:
// 1. User-defined key (preferred, like React's key prop)
// 2. Path + index (stable across rebuilds, like React's index fallback)
var uniqueID string

// Priority 1: User-defined key (most stable)
if keyer, ok := vnode.(interface{ Key() string }); ok {
    if key := keyer.Key(); key != "" {
        uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
    }
}

// Priority 2: Path + index (stable across rebuilds)
if uniqueID == "" {
    uniqueID = fmt.Sprintf("%s[%d]", nodePath, index)
}
```

### How It Works

#### Example Structure
```go
VStack(
    VStack(
        Text("A"),  // index 0
        Text("B"),  // index 1
    ),
    VStack(
        Text("C"),  // index 0
        Text("D"),  // index 1
    ),
)
```

#### Generated Paths
```
Root:                    vstack[0]
  ├─ Child 1:            vstack[0].vstack[0]
  │   ├─ Text "A":       vstack[0].vstack[0].text[0]
  │   └─ Text "B":       vstack[0].vstack[0].text[1]
  └─ Child 2:            vstack[1].vstack[1]
      ├─ Text "C":       vstack[1].vstack[0].text[0]
      └─ Text "D":       vstack[1].vstack[1].text[1]
```

All paths are unique! ✅

## Why This Works

### Stability Across Rebuilds

| Approach | Stable Across Rebuilds | Unique |
|----------|----------------------|--------|
| Counter | ❌ Changes every rebuild | ✅ Yes |
| Path-only | ✅ Yes | ❌ Collisions |
| Pointer address | ❌ Changes (VNodes recreated) | ✅ Yes |
| **Path + index** | ✅ **Yes** | ✅ **Yes** |

### Comparison with React

| Aspect | React | Mint TUI (This Fix) |
|--------|-------|---------------------|
| **Primary ID** | User-provided key | User-provided key |
| **Fallback** | Array index | Path with parent indices |
| **Collision handling** | Warns about missing keys | No collisions (path is unique) |
| **Stability** | Index shifts on insert/delete | Stable as long as structure doesn't change |

## Test Results

All tests pass ✅:
```
✓ TestTreeViewUniqueIDLookup
✓ TestTreeViewExpandCollapse
✓ TestTreeViewPathConsistency
✓ TestTreeViewIndexBasedIDs
✓ TestTreeViewPathIndexBasedIDs
✓ TestTreeViewNestedUniqueness
✓ TestTreeViewIDStability
```

### Example Test Output
```
Node: vstack[0]
Node: vstack[0].vstack[0]
Node: vstack[0].vstack[0].text[0]
Node: vstack[0].vstack[1].text[1]
Node: vstack[1].vstack[1]
Node: vstack[1].vstack[0].text[0]
Node: vstack[1].vstack[1].text[1]

✓ All 7 nodes have unique IDs
```

## Files Modified

1. **`internal/inspector/tree_view.go`**
   - Changed path generation to include parent index: `path[index].type`
   - Changed UniqueID generation to use path + index
   - Removed ComponentID field (no longer needed)

2. **`internal/inspector/tree_pathindex_test.go`** (renamed from `tree_pointerid_test.go`)
   - Tests verify path + index creates unique IDs
   - Tests verify IDs are stable across rebuilds
   - Tests verify no collisions occur

3. **`internal/inspector/tree_view_test.go`**
   - Updated `TestToggleNode` to use `UniqueID` instead of `Path`

## Success Criteria

1. ✅ No UniqueID collisions
2. ✅ Correct node expands when pressing E
3. ✅ IDs stable across rebuilds (no pointers!)
4. ✅ Matches React's philosophy: user keys preferred
5. ✅ Tests pass

## Migration Guide

If you have code that relied on old ID formats:

### Before (ComponentID-based - BROKEN)
```go
// "comp-3.vstack.text[0]" - Collisions!
uniqueID = fmt.Sprintf("%s.%s[%d]", componentID, nodePath, index)
```

### After (Path + Index - WORKS)
```go
// "vstack[0].vstack[0].text[0]" - Unique!
nodePath = fmt.Sprintf("%s[%d].%s", parentPath, index, type)
uniqueID = fmt.Sprintf("%s[%d]", nodePath, index)
```

### Best Practice
Add keys to your components for the most stable IDs:
```go
VStack(
    Text("Item 1").Key("item-1"),
    Text("Item 2").Key("item-2"),
)
// IDs: "vstack[0].text[item-1]", "vstack[1].text[item-2]"
```

## Key Insight

**The fix is simple: Include the parent's index in the path.**

This ensures:
- Every path is unique (no collisions)
- Paths are stable across rebuilds (no pointers)
- Matches React's approach (index-based fallback)

The previous attempt using **pointer addresses was fundamentally broken** because VNodes are recreated on every render, so pointers change. The correct approach is to use **path + index**, which is stable as long as the structure doesn't change.

## References

- React's reconciliation: https://react.dev/learn/render-and-commit
- React key documentation: https://react.dev/learn/rendering-lists#keeping-list-items-in-order-with-key
