# Inspector Elements Panel Path Display Fix Report

## Overview

This report documents the comprehensive fix for the Inspector's Elements panel path display issues. The problems involved incorrect path formats, missing hierarchical information, and mismatched selected item display.

## Problem Statement

The Inspector's Elements panel had multiple issues:

1. **User keys not showing full path**: When a component had a user key (e.g., `key="btn-event"`), the Inspector showed only the original key instead of the full hierarchical path.

2. **Missing layer/base information**: Paths displayed as `/vstack[0]/bordered[0]` instead of `/root/base[0]/vstack[0]/bordered[0]`.

3. **Selected Path mismatch**: The "Path:" field in the selected item details showed a different path than what was displayed in the tree, and often used the old dot-notation format (e.g., `vstack[1].bordered`) instead of the new slash-based format.

## Root Cause Analysis

### Issue 1: Root Fiber Missing Path

**Location**: `internal/reconciler/reconciler.go` - `prepareFreshStack()`

**Problem**: The root ComponentVNode Fiber was created with `Key="root"` but `Path` was not set. When child nodes generated their paths using `parent.Path + "/" + segment`, they got paths like `/vstack[0]` instead of `/root/base[0]/vstack[0]`.

**Root Cause Chain**:
```
prepareFreshStack() creates root Fiber
  └── root.Path = "" (not set)
       └── createChildFiber() for root's children
            └── PathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
                 └── return parent.Path + "/" + segment
                      └── Result: "" + "/" + "vstack[0]" = "/vstack[0]" (WRONG!)
```

### Issue 2: Root Child Not Using Layer Information

**Location**: `internal/reconciler/diff.go` - `createChildFiberWithIndex()`

**Problem**: When the parent is the root ComponentVNode (Key="root", Path="/root"), the child should get a layer-based path (e.g., `/root/base[0]`). But the code used standard path generation which doesn't include layer info.

**Expected**: `/root/base[0]` (base is the layer name)
**Actual**: `/root/vstack[0]` (vstack is the first child type)

### Issue 3: expandVNodeTree Using Wrong Children Source

**Location**: `internal/reconciler/reconciler.go` - `expandVNodeTree()`

**Problem**: The function used `vnode.Children()` to get child VNodes, but these are the original VNodes before Fiber reconciliation. The Fiber tree has the correct VNodes with properly set Keys.

**Code before fix**:
```go
originalChildren := vnode.Children()  // Original VNodes - Keys not updated
for _, child := range originalChildren {
    expandedChild := r.expandVNodeTree(child, childFiber)
    // child.Key() returns OLD key, not the Fiber path
}
```

### Issue 4: flattenRecursive Not Checking Expand State

**Location**: `internal/inspector/tree_view.go` - `flattenRecursive()`

**Problem**: The function included ALL children regardless of expand state, but `formatNode()` only showed children of expanded nodes. This caused index mismatch between `GetTreeLines()` and `GetFlatList()`.

**Code before fix**:
```go
func (tv *TreeView) flattenRecursive(node *TreeNode, nodes *[]*TreeNode) {
    *nodes = append(*nodes, node)
    for _, child := range node.Children {  // Always adds ALL children
        tv.flattenRecursive(child, nodes)
    }
}
```

### Issue 5: Focus Index to Flat Nodes Mapping Error

**Location**: `internal/inspector/standalone_inspector.go` - render logic

**Problem**: The code used `flatNodes[focusIndex]` but the correct mapping is `flatNodes[focusIndex - 1]` because:
- `treeLines[0]` = header (no corresponding node)
- `treeLines[1..n-2]` = nodes → `flatNodes[0..n-3]`
- `treeLines[n-1]` = footer (no corresponding node)

## Solutions Implemented

### Fix 1: Set Root Fiber Path

**File**: `internal/reconciler/reconciler.go`

```go
func (r *Reconciler) prepareFreshStack(renderFunc func() rtui.VNode) {
    // ... existing code ...
    if r.root == nil {
        r.root = CreateFiberFromVNode(rootComponentVNode)
        // ✨ Set root Fiber's Path for proper path generation in children
        r.root.Path = "/root"
        r.root.Key = "root"
        r.workInProgress = r.root
    } else {
        r.workInProgress = r.createWorkInProgress(r.root, rootComponentVNode)
        // Ensure workInProgress also has the correct Path
        r.workInProgress.Path = "/root"
    }
}
```

### Fix 2: Root Child Gets Layer-Based Path

**File**: `internal/reconciler/diff.go`

```go
func createChildFiberWithIndex(returnFiber *Fiber, vnode rtui.VNode, lanes Lane, siblingIndex int, typeIndex int) *Fiber {
    // ... existing code ...

    // ✨ Special case: If parent is the root ComponentVNode (Key="root"),
    // this child is the actual app content and should get a layer-based path
    isRootChild := returnFiber != nil && returnFiber.Key == "root" && returnFiber.Path == "/root"

    if userKey != "" {
        // ... existing code ...
        var typePath string
        if isRootChild {
            // Root's child gets layer-based path (e.g., /root/base[0])
            typePath = pathGenerator.generateRootPath(vnode)
        } else {
            typePath = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
        }
        // ...
    } else {
        // Priority 3: Static UI → auto-generate path key
        if isRootChild {
            fiber.Path = pathGenerator.generateRootPath(vnode)
        } else if typeIndex >= 0 {
            fiber.Path = pathGenerator.GeneratePathWithIndex(returnFiber, vnode, siblingIndex, typeIndex)
        } else {
            fiber.Path = pathGenerator.GeneratePath(returnFiber, vnode, siblingIndex)
        }
        // ...
    }
}
```

### Fix 3: Use Fiber Children in expandVNodeTree

**File**: `internal/reconciler/reconciler.go`

```go
func (r *Reconciler) expandVNodeTree(vnode rtui.VNode, fiber *Fiber) rtui.VNode {
    // ... existing code ...

    // For other VNodes, recursively expand children
    if vnode.Type() == rtui.VNodeElement || vnode.Type() == rtui.VNodeFragment {
        originalChildren := vnode.Children()
        if len(originalChildren) == 0 {
            return vnode
        }

        if fiber == nil {
            return vnode
        }

        // ✨ IMPORTANT: Build children from Fiber tree, not from VNode.Children()
        // Fiber.VNode has the correct Key set by reconciliation, but VNode.Children()
        // returns the original children which may have outdated keys.
        expandedChildren := r.buildVNodeList(fiber.Child)
        if len(expandedChildren) == 0 {
            return vnode
        }

        cloned := r.cloneVNodeWithChildren(vnode, expandedChildren)
        return cloned
    }
    // ...
}
```

### Fix 4: flattenRecursive Respects Expand State

**File**: `internal/inspector/tree_view.go`

```go
func (tv *TreeView) flattenRecursive(node *TreeNode, nodes *[]*TreeNode) {
    if node == nil {
        return
    }

    *nodes = append(*nodes, node)

    // Only add children if node is expanded (matches formatNode behavior)
    if node.Expanded {
        for _, child := range node.Children {
            tv.flattenRecursive(child, nodes)
        }
    }
}
```

### Fix 5: Correct Focus Index Mapping

**File**: `internal/inspector/standalone_inspector.go`

```go
// In render logic:
} else if focusIndex >= 0 {
    // ✨ Map focusIndex to flatNodes index
    // treeLines structure: [0]=header, [1..n-1]=nodes, [n]=footer
    // flatNodes structure: [0..m-1]=nodes (same nodes as treeLines[1..n-1])
    flatNodes := si.treeView.GetFlatList()

    // Adjust for header line offset
    nodeIndex := focusIndex - 1

    // Verify we're within valid range (not header, not footer)
    if nodeIndex >= 0 && nodeIndex < len(flatNodes) && focusIndex < len(si.treeLines)-1 {
        node := flatNodes[nodeIndex]
        targetVNode = node.VNode
        targetPath = node.Path
        displayType = "Focused"
    }
}
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         USER CODE                                        │
│   ui.VStack(...), ui.Button(...).Key("btn-event"), etc.                  │
│   VNode.Key() = "" or user key                                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    FIBER RECONCILIATION (diff.go)                        │
│                                                                          │
│   prepareFreshStack():                                                   │
│     root = CreateFiberFromVNode(rootComponentVNode)                      │
│     root.Path = "/root"  ← FIX 1                                         │
│                                                                          │
│   createChildFiberWithIndex():                                           │
│     if isRootChild:                                                      │
│       fiber.Path = generateRootPath(vnode)  ← FIX 2                      │
│       // Returns "/root/base[0]" with layer info                         │
│     else:                                                                │
│       fiber.Path = parent.Path + "/" + segment                           │
│       // e.g., "/root/base[0]/vstack[0]"                                 │
│                                                                          │
│     vnode.SetKey(fiber.Path)  // Sync to VNode for Inspector             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    VNODE TREE REBUILD (reconciler.go)                    │
│                                                                          │
│   buildVNodeTree():                                                      │
│     For ComponentVNode: return children from Fiber                       │
│     For ElementVNode: return fiber.VNode (with correct Key)              │
│                                                                          │
│   expandVNodeTree():                                                     │
│     expandedChildren = buildVNodeList(fiber.Child)  ← FIX 3              │
│     // Uses Fiber children, not VNode.Children()                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    INSPECTOR RECEIVE (standalone_inspector.go)           │
│                                                                          │
│   App.GetRenderedRoot() → reconciler.GetRenderedRoot()                   │
│   Inspector.AttachToApp(renderedRoot)                                    │
│   Inspector.TreeView.SetRoot(renderedRoot)                               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    TREE BUILDING (tree_view.go)                          │
│                                                                          │
│   buildTree():                                                           │
│     if vnode.Key() starts with "/root/":                                 │
│       nodePath = vnode.Key()[6:]  // Remove "/root/" prefix              │
│       // e.g., "base[0]/vstack[0]/bordered[0]"                           │
│     else:                                                                │
│       nodePath = fallback format (dot notation)                          │
│                                                                          │
│   flattenRecursive():  ← FIX 4                                           │
│     if node.Expanded:                                                    │
│       for child in node.Children:                                        │
│         flattenRecursive(child)                                          │
│     // Matches formatNode() behavior for consistent indexing             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    DISPLAY & SELECTION (standalone_inspector.go)         │
│                                                                          │
│   GetTreeLines() → [header, node1, node2, ..., footer]                   │
│   GetFlatList() → [node1, node2, ...]  (no header/footer)                │
│                                                                          │
│   focusIndex mapping:  ← FIX 5                                           │
│     flatNodes[focusIndex - 1] = selected node                            │
│                                                                          │
│   Result:                                                                │
│     Tree: "└── bordered key:/root/base[0]/vstack[0]/bordered[0]"         │
│     Selected Path: "base[0]/vstack[0]/bordered[0]"  ← MATCHES!           │
└─────────────────────────────────────────────────────────────────────────┘
```

## Index Mapping Verification

The following unit test verifies the correct index mapping:

```go
// TestTreeLinesAndFlatListConsistency verifies:
// - treeLines[0] = header (no node)
// - treeLines[1..n-2] = nodes → flatNodes[0..n-3]
// - treeLines[n-1] = footer (no node)

// TestFocusIndexToFlatNodeMapping verifies:
// - focusIndex=0 (header) → no node
// - focusIndex=1 → flatNodes[0]
// - focusIndex=2 → flatNodes[1]
// - ...
// - focusIndex=n-1 (footer) → no node
```

## Test Results

All relevant tests pass:

```
=== RUN   TestTreeLinesAndFlatListConsistency
    ✓ flatNodes[0] (Path=base[0]) matches treeLines[1]
    ✓ flatNodes[1] (Path=base[0]/vstack[0]) matches treeLines[2]
    ✓ flatNodes[2] (Path=base[0]/vstack[0]/text[0]) matches treeLines[3]
    ...
--- PASS: TestTreeLinesAndFlatListConsistency

=== RUN   TestFocusIndexToFlatNodeMapping
    ✓ focusIndex=0 (header) correctly maps to no node
    ✓ focusIndex=1 maps to flatNodes[0]: base[0]
    ✓ focusIndex=2 maps to flatNodes[1]: base[0]/child1[0]
    ✓ focusIndex=3 maps to flatNodes[2]: base[0]/child2[1]
    ✓ focusIndex=4 (footer) correctly maps to no node
--- PASS: TestFocusIndexToFlatNodeMapping

=== RUN   TestDisplayTreeViewIndexConsistency
--- PASS: TestDisplayTreeViewIndexConsistency

=== RUN   TestPathBasedKeyStrategy_Integration
    Panel1: Key=/root/base[0]/vstack[0]/panel[0]
    Panel2: Key=/root/base[0]/vstack[0]/panel[1]
    Button: Key=/root/base[0]/vstack[0]/button[0]
--- PASS: TestPathBasedKeyStrategy_Integration

=== RUN   TestUserKeyPriority_Integration
    Button1: Key=save-btn, Path=/root/base[0]/vstack[0]/button[0]/key[save-btn]
    Button2: Key=cancel-btn, Path=/root/base[0]/vstack[0]/button[0]/key[cancel-btn]
--- PASS: TestUserKeyPriority_Integration
```

## Files Modified

| File | Changes |
|------|---------|
| `internal/reconciler/reconciler.go` | Set root Fiber Path, use Fiber children in expandVNodeTree |
| `internal/reconciler/diff.go` | Root child gets layer-based path |
| `internal/inspector/tree_view.go` | flattenRecursive respects expand state |
| `internal/inspector/standalone_inspector.go` | Correct focus index mapping |
| `internal/inspector/index_mapping_test.go` | NEW: Unit tests for index mapping |
| `internal/reconciler/user_key_path_test.go` | NEW: Unit tests for user key paths |
| `components/display/treeview.go` | Improved getNodeDescription for better display |

## Expected Output After Fix

### Tree Display
```
┌─ Layout Tree ─────────────────────────────────
└── 📦 ElementVNode key:/root/base[0]
│  ├── 🎨 BorderedNode key:/root/base[0]/bordered[0]
│  │  └── 📦 LayoutNode key:/root/base[0]/bordered[0]/hstack[0]
│  │  │  └── 📝 TextVNode key:/root/base[0]/bordered[0]/hstack[0]/text[0]
│  ├── 🎨 BorderedNode key:/root/base[0]/bordered[1]
│  │  └── 📦 LayoutNode key:/root/base[0]/bordered[1]/vstack[0]
│  │  │  └── 📝 TextVNode key:/root/base[0]/bordered[1]/vstack[0]/text[0]
└─────────────────────────────────────────────┘
```

### Selected Item Display
```
Focused: Text
Path: base[0]/bordered[0]/hstack[0]/text[0]
```

The Path now matches the key displayed in the tree (minus `/root/` prefix), providing consistent and accurate element identification.

## Summary

This fix addresses five interconnected issues in the Inspector's Elements panel:

1. **Root Fiber Path** - Ensures path hierarchy starts correctly from `/root`
2. **Layer Information** - Includes layer name (e.g., `base`) in paths
3. **Fiber Children** - Uses Fiber tree's VNodes with correct Keys
4. **Expand State** - Keeps flat list consistent with displayed tree
5. **Index Mapping** - Correctly maps focus index to node index

The result is a consistent, accurate path display that helps developers identify and debug UI component hierarchies.
