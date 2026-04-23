# Inspector UniqueID: Final Solution

## Problem Evolution

### First Attempt (BROKEN): Counter-based IDs
```go
UniqueID: fmt.Sprintf("node-%d", tv.nextNodeID)
tv.nextNodeID++
```
**Problem**: IDs changed on every rebuild because VNodes are recreated

### Second Attempt (BROKEN): Path-only IDs
```go
uniqueID := nodePath  // e.g., "vstack.text"
```
**Problem**: Collisions when multiple siblings have same type:
- `VStack(Text("A"), Text("B"))` → both get ID `vstack.text`

### Third Attempt (BROKEN): Path + Content Hash
```go
contentHash := tv.generateContentHash(vnode)
uniqueID = fmt.Sprintf("%s@%s", nodePath, contentHash)
```
**Problem**: Hash changes when content changes, breaking state persistence

## Final Solution: Path + Index

### Implementation

**File**: `internal/inspector/tree_view.go:60-117`

```go
func (tv *TreeView) buildTree(vnode ui.VNode, parent *TreeNode, level int, path string, index int) *TreeNode {
    // ... (path generation)

    // Priority 1: User-defined key (most stable)
    var uniqueID string
    if keyer, ok := vnode.(interface{ Key() string }); ok {
        if key := keyer.Key(); key != "" {
            uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
        }
    }

    // Priority 2: Path + index (stable as long as structure doesn't change)
    if uniqueID == "" {
        uniqueID = fmt.Sprintf("%s[%d]", nodePath, index)
    }

    // ... (create node with uniqueID)
}
```

### How It Works

#### Case 1: User-Defined Keys (Best)
```go
VStack(
    Text("A").Key("first"),
    Text("B").Key("second"),
)
```
**IDs**: `vstack.text[first]`, `vstack.text[second]`
- ✅ Most stable
- ✅ Survives reordering
- ✅ Survives insertions/deletions elsewhere

#### Case 2: Index-Based (Default)
```go
VStack(
    Text("A"),  // index 0
    Text("B"),  // index 1
    Text("C"),  // index 2
)
```
**IDs**: `vstack.text[0]`, `vstack.text[1]`, `vstack.text[2]`
- ✅ No collisions
- ✅ Stable across rebuilds
- ⚠️ Index shifts if nodes added/removed before

## Why This Works

### Stability Matrix

| Approach | Stable Across Rebuilds | No Collisions | Survives Reordering | Survives Insertions |
|----------|----------------------|---------------|---------------------|---------------------|
| Counter | ❌ Changes every rebuild | ✅ Yes | ❌ No | ❌ No |
| Path-only | ✅ Yes | ❌ Collisions | ✅ Yes | ✅ Yes |
| Content Hash | ❌ Changes with content | ✅ Yes | ❌ No | ❌ No |
| **Path + Index** | ✅ **Yes** | ✅ **Yes** | ⚠️ **No** | ⚠️ **No** |
| **Path + User Key** | ✅ **Yes** | ✅ **Yes** | ✅ **Yes** | ✅ **Yes** |

### When to Use Each

1. **User provides keys**: Use them (most robust)
2. **No keys provided**: Use index-based (good enough)
   - Expand state persists as long as tree structure doesn't change
   - If structure changes, user can re-expand (acceptable)

## Test Results

### Unit Tests (All Pass)

```
TestTreeViewUniqueIDLookup      ✓ IDs match expected format
TestTreeViewExpandCollapse      ✓ Expand state persists
TestTreeViewPathConsistency     ✓ IDs stable across operations
TestTreeViewIndexBasedIDs       ✓ No collisions with same-type siblings
```

**Collision Test Output**:
```
Child 0: UniqueID=VStack.Text[0], Label=A
Child 1: UniqueID=VStack.Text[1], Label=B
Child 2: UniqueID=VStack.Text[2], Label=C
✓ All three Text nodes have different UniqueIDs
```

### Integration Test (Pass)

```
Status line: 🔍 Focus: #3 [ID: vstack.text[2]] → ElementVNode
```

## Key Insights

### 1. Index is Stable Across Rebuilds

When VNodes are recreated, they maintain the same structure:
```go
// Render 1
VStack(Text("A"), Text("B"))
// → vstack.text[0], vstack.text[1]

// Render 2 (VNodes recreated, but structure same)
VStack(Text("A"), Text("B"))
// → vstack.text[0], vstack.text[1]  ✓ Same IDs!
```

### 2. Index-Based IDs Work Like React's Key System

React uses index as fallback when no key provided:
```jsx
{/* React behavior */}
[
  <Text>A</Text>,  // React assigns index 0
  <Text>B</Text>,  // React assigns index 1
]
```

Our approach matches React's behavior!

### 3. User Keys Override Everything

```go
VStack(
    Text("A").Key("first"),   // vstack.text[first]
    Text("B").Key("second"),  // vstack.text[second]
)
```

User keys are always preferred (just like React).

## Comparison with Existing Systems

### Focus Management System

Mint's focus management uses similar approach:

```go
// Button.GetFocusID()
func (b *ButtonVNode) GetFocusID() string {
    if key := b.Key(); key != "" {
        return "button:" + key
    }
    return fmt.Sprintf("button:%s@%p", b.label, b)
}
```

**Key difference**: Focus uses pointer address (`@%p`) as fallback
- Works for focus (recomputed each frame)
- Doesn't work for inspector (needs persistence)

**Our solution**: Uses index instead of pointer address
- More stable across rebuilds
- No reliance on memory addresses

## Files Changed

| File | Changes |
|------|---------|
| `tree_view.go:60-117` | Index-based UniqueID generation |
| `tree_view.go:50-51` | Pass index to buildTree |
| `tree_expand_test.go` | Updated tests with index format |
| `INSPECTOR_UNIQUEID_FINAL_SOLUTION.md` | This document |

## Migration Guide

### Before
```
Status: 🔍 Focus: #3 [ID: node-136] → ElementVNode
Problem: ID changes every rebuild
```

### After
```
Status: 🔍 Focus: #3 [ID: vstack.text[2]] → ElementVNode
✓ ID is stable across rebuilds
✓ No collisions with siblings
```

## Best Practices for Users

### 1. Provide Keys When Structure Changes

```go
// Bad: Index shifts when items added/removed
VStack(
    Text(getItem(0)),
    Text(getItem(1)),
    Text(getItem(2)),
)

// Good: Keys make IDs stable
VStack(
    Text(getItem(0)).Key("item-0"),
    Text(getItem(1)).Key("item-1"),
    Text(getItem(2)).Key("item-2"),
)
```

### 2. Acceptable: Index-Based for Static Layouts

```go
// Fine: Static structure won't change
VStack(
    Text("Title"),
    Text("Subtitle"),
    Text("Body"),
)
```

### 3. Inspector Will Re-Expand If Needed

If tree structure changes and expand state is lost:
- User can press E to re-expand
- Inspector still functions correctly
- Minor inconvenience, not a blocker

## Conclusion

The **path + index** approach provides:
- ✅ Stability across rebuilds (same structure = same IDs)
- ✅ No collisions (index differentiates siblings)
- ✅ Simplicity (easy to understand and debug)
- ✅ Compatibility (works like React's key system)

For most use cases, this is **stable enough**. For cases where structure changes frequently, users can provide explicit keys for maximum stability.

**The Inspector's expand/collapse feature now works reliably!**
