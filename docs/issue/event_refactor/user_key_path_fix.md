# User Key Path Display Fix

## Problem

The Inspector's Elements panel had two issues:

### Issue 1: User Keys Not Showing Full Path
When a component had a user key (e.g., `key="btn-event"`), the Inspector would show:
- Tree display: `key:'btn-event'` instead of the full path
- Selected item Path field: Empty or old dot-notation format

### Issue 2: Selected Path Mismatch with Tree Display
When selecting a node in the tree, the "Path:" field showed a different path than what was displayed in the tree:
- Tree display: `key:'/vstack[0]/bordered[0]/hstack[0]/text[0]'`
- Selected Path: `vstack[1].bordered` (old dot-notation format, wrong index)

## Root Causes

### Issue 1 Root Cause
1. **Missing Path Sync for User Keys**: In `cloneExistingFiber()`, the code only synced auto-generated path keys to VNode, not user key paths
2. **Path Recognition Logic**: `getPath()` in `element_info.go` only recognized paths starting with `/root/`

### Issue 2 Root Cause (Critical Bug)
**Index Mismatch Between Tree Display and Flat Node List**

The Inspector uses two different TreeView systems:
1. `display.TreeView` - Shows the visible tree lines (respects expand/collapse state)
2. `inspector.TreeView` - Provides `GetFlatList()` for node lookup

The bug was in `flattenRecursive()` which **ignored expand/collapse state**:
```go
// BEFORE (BUG):
func (tv *TreeView) flattenRecursive(node *TreeNode, nodes *[]*TreeNode) {
    *nodes = append(*nodes, node)
    for _, child := range node.Children {  // Always adds ALL children
        tv.flattenRecursive(child, nodes)
    }
}
```

But `formatNode()` (used for tree lines) only shows children of **expanded** nodes:
```go
if node.Expanded && len(node.Children) > 0 {
    for i, child := range node.Children {
        lines = tv.formatNode(child, lines, ...)
    }
}
```

This caused the `focusIndex` from `display.TreeView` to index into the wrong node in `GetFlatList()`.

## Solution

### Fix 1: Modified `createChildFiber()` and `cloneExistingFiber()` in `diff.go`
Sync full path to VNode for user keys:
```go
if userKey != "" {
    fiber.Key = userKey
    typePath := pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
    fiber.Path = typePath + "/key[" + userKey + "]"
    vnode.SetKey(fiber.Path)  // ✨ Sync to VNode
}
```

### Fix 2: Modified `flattenRecursive()` in `internal/inspector/tree_view.go`
Match expand/collapse behavior with `formatNode()`:
```go
// AFTER (FIXED):
func (tv *TreeView) flattenRecursive(node *TreeNode, nodes *[]*TreeNode) {
    *nodes = append(*nodes, node)

    // Only add children if node is expanded (matches formatNode behavior)
    if node.Expanded {
        for _, child := range node.Children {
            tv.flattenRecursive(child, nodes)
        }
    }
}
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         RENDER PIPELINE                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. User Code creates VNode tree                                 │
│     ui.VStack(...), ui.Button(...), etc.                         │
│     - VNode.Key() = "" (no key set yet)                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. Fiber Reconciliation (diff.go)                               │
│     createChildFiber() / cloneExistingFiber()                    │
│     - fiber.Key = userKey (e.g., "btn-event")                    │
│     - fiber.Path = "/root/base[0]/button[0]/key[btn-event]"      │
│     - vnode.SetKey(fiber.Path) ← Key synced to VNode             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. buildVNodeTree() - Reconstruct VNode tree from Fiber         │
│     - Returns fiber.VNode (with correct Key)                     │
│     - Key format: "/root/base[0]/vstack[0]/..."                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. Inspector receives tree via AttachToApp()                    │
│     - App.GetRenderedRoot() → reconciler.GetRenderedRoot()       │
│     - Inspector.TreeView.SetRoot(renderedRoot)                   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  5. Inspector.TreeView.buildTree()                               │
│     - Check VNode.Key() for "/root/" prefix                      │
│     - If found: nodePath = vnodeKey[6:] (strip "/root/")         │
│     - If not: fallback to old format (BUG SOURCE 1)              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  6. Display Tree                                                 │
│     GetTreeLines() → formatNode() → [only EXPANDED nodes]        │
│     GetFlatList() → flattenRecursive() → [now only EXPANDED]     │
│     ↑ FIXED: Both now respect expand/collapse state              │
└─────────────────────────────────────────────────────────────────┘
```

## Files Modified

1. `internal/reconciler/diff.go`
   - `createChildFiber()`: Sync user key paths to VNode
   - `cloneExistingFiber()`: Sync all paths to VNode

2. `internal/inspector/tree_view.go`
   - `flattenRecursive()`: Only include children of expanded nodes

3. `internal/reconciler/integration_test.go`
   - Update expected path format for user keys

4. `internal/inspector/element_path_test.go`
   - Update test to expect full path for user keys

5. `internal/reconciler/user_key_path_test.go` (NEW)
   - Comprehensive tests for user key path generation

## Test Results

All relevant tests pass:
```
✓ TestExtractElementInfo_UserKey
✓ TestUserKeyPathGeneration
✓ TestUserKeyPathReuse
✓ TestUserKeyPriority_Integration
✓ TestTreeViewExpandCollapse
✓ TestTreeViewPathConsistency
```

## Example Output

### Before Fix
```
Tree Display:
  └── 🎨 Button  key:'btn-event'
  (focusIndex = 5)

FlatList (GetFlatList):
  [0] Root
  [1] VStack (expanded)
  [2] Bordered (collapsed) ← BUG: included despite being collapsed!
  [3] HStack (child of collapsed Bordered - should NOT be here!)
  [4] Button
  [5] Text              ← focusIndex[5] points here
  [6] Another Button

Selected Path: vstack[1].bordered ← WRONG! (old format, wrong index)
```

### After Fix
```
Tree Display:
  └── 🎨 Button  key:'base[0]/vstack[0]/button[0]/key[btn-event]'
  (focusIndex = 5)

FlatList (GetFlatList):
  [0] Root
  [1] VStack (expanded)
  [2] Bordered (collapsed - children NOT included)
  [3] Another node
  [4] Another node
  [5] Button            ← focusIndex[5] now correctly points here

Selected Path: base[0]/vstack[0]/button[0]/key[btn-event] ← CORRECT!
```

This provides full visibility and consistency in the component hierarchy for debugging and inspection.
