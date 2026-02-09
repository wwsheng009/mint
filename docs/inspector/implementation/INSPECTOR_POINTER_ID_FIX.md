# Inspector Pointer-Based ID Fix

## Problem Statement

The Inspector TreeView's expand/collapse feature was broken because UniqueIDs were colliding, causing the wrong nodes to expand when pressing E.

**Specific Issue**: `vstack.bordered.vstack[0]` appeared multiple times - LayoutNode and BorderNode had the same ID, causing the Inspector to toggle the wrong node.

## Root Cause

The collision happened because:

1. **ComponentID was shared among siblings** - All nodes in the same component context got the same ComponentID
2. **Path + index not unique enough** - Nested structures had the same path at different levels
3. **Example collision**:
   ```
   ComponentID: comp-3
   - LayoutNode path: comp-3.vstack.bordered[0]
   - BorderNode path: comp-3.vstack.bordered[0]  // COLLISION!
   ```

## Solution: Pointer-Based IDs

Following React's philosophy, we now use:

1. **User-defined key** (preferred, like React's `key` prop)
2. **VNode pointer address** (guaranteed unique, like React's index fallback)

### Implementation

**File**: `internal/inspector/tree_view.go:76-88`

```go
// Generate UniqueID following React's philosophy:
// 1. User-defined key (preferred, like React's key prop)
// 2. VNode pointer address (guaranteed unique, like React's index fallback)
var uniqueID string

// Priority 1: User-defined key (most stable)
if keyer, ok := vnode.(interface{ Key() string }); ok {
    if key := keyer.Key(); key != "" {
        uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
    }
}

// Priority 2: VNode pointer (prevents collisions, based on VNode itself)
if uniqueID == "" {
    uniqueID = fmt.Sprintf("%s[%d]@%p", nodePath, index, vnode)
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

#### Case 2: Pointer-Based (Default)
```go
VStack(
    Text("A"),  // VNode at 0x1234567890
    Text("B"),  // VNode at 0x12345678a0
)
```
**IDs**: `vstack.text[0]@0x1234567890`, `vstack.text[1]@0x12345678a0`
- ✅ No collisions (pointer addresses are unique)
- ✅ Stable across rebuilds (same VNode = same pointer)
- ✅ Works for any structure

## Comparison with React

| Aspect | React | Mint TUI (After Fix) |
|--------|-------|---------------------|
| **Primary ID** | User-provided key | User-provided key |
| **Fallback** | Array index (with warning) | Pointer address (unique) |
| **Collision handling** | Warns about missing keys | Uses pointer (no collision) |
| **Uniqueness** | Key + index | Key + pointer |

## Why This Matches React

React's reconciliation algorithm:
```javascript
// React's matching logic
if (oldFiber.key !== newVNode.key) return false;  // Key first
if (oldFiber.type !== newVNode.type) return false; // Type second
```

Mint TUI's approach:
```go
// Mint TUI's matching logic
if key := vnode.Key(); key != "" {
    uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)  // Key first
} else {
    uniqueID = fmt.Sprintf("%s[%d]@%p", nodePath, index, vnode)  // Pointer fallback
}
```

Both prioritize user keys, but Mint TUI uses pointer addresses instead of array index for better uniqueness.

## Verification

### Automated Tests
```bash
cd internal/inspector
go test -v -run "TestTreeView"
```

All tests pass with pointer-based IDs:
- ✅ `TestTreeViewPointerBasedIDs` - Verifies pointer addresses create unique IDs
- ✅ `TestTreeViewNestedPointerIDs` - Verifies nested structures have unique IDs
- ✅ `TestTreeViewExpandCollapse` - Verifies expand/collapse works correctly

### Manual Test
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

Expected behavior:
- Status line shows: `[ID: vstack.text[0]@0x1234567890]`
- No duplicate IDs in tree
- Pressing E expands the correct node

## Files Modified

1. **`internal/inspector/tree_view.go`**
   - Changed UniqueID generation to use pointer address when no key
   - Removed ComponentID from UniqueID generation
   - Removed ComponentID field from TreeNode struct
   - Removed dependency on `runtime/ui` package

2. **`internal/inspector/tree_pointerid_test.go`** (renamed from `tree_componentid_test.go`)
   - Updated tests to verify pointer-based IDs
   - Tests verify `@` prefix in pointer-based IDs
   - Tests verify no collisions occur

## Success Criteria

1. ✅ No UniqueID collisions
2. ✅ Correct node expands when pressing E
3. ✅ IDs based on VNode (pointer address when no key)
4. ✅ Matches React's philosophy: user keys preferred
5. ✅ Tests pass

## Migration Guide

If you have code that relied on ComponentID-based IDs:

### Before
```go
// Old ID format: "comp-3.vstack.text[0]"
uniqueID := fmt.Sprintf("%s.%s[%d]", componentID, nodePath, index)
```

### After
```go
// New ID format: "vstack.text[0]@0x1234567890"
if key := vnode.Key(); key != "" {
    uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
} else {
    uniqueID = fmt.Sprintf("%s[%d]@%p", nodePath, index, vnode)
}
```

### Best Practice
Add keys to your components for the most stable IDs:
```go
VStack(
    Text("Item 1").Key("item-1"),
    Text("Item 2").Key("item-2"),
)
```

## Future Improvements

1. **Enable key warnings by default** - Like React, warn developers when lists don't have keys
2. **Auto-generate keys in development** - Helper pattern for automatic key generation
3. **Provide key-setting helpers** - Better API for key management

## References

- React's reconciliation: https://react.dev/learn/render-and-commit#react-only-changes-whats-necessary
- React key documentation: https://react.dev/learn/rendering-lists#keeping-list-items-in-order-with-key
- Plan document: `INSPECTOR_UNIQUEID_FIX_PLAN.md`
