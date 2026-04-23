# Inspector Path Display Fix - Complete Solution

## Problem

The Inspector's Elements panel was displaying the old dot-notation path format instead of the new Fiber-generated slash-based path format:

**Before (old format):**
```
Path: vstack[2].bordered[0].hstack[2].text
```

**Expected (new format):**
```
Path: base[0]/vstack[0]/bordered[0]/hstack[0]/text[0]
```

## Root Causes

### 1. `ElementInfo.Path` Was Not Populated
The `ExtractElementInfo()` function in `element_info.go` was not populating the `Path` field, even though the struct had it.

### 2. VNodes Lost Fiber Keys After Re-render
When VNodes were re-rendered (new instances created), they lost the Fiber-generated path keys that were set during initial render. The `cloneExistingFiber()` function preserved the Fiber's Path and Key fields but didn't sync them back to the new VNode instance.

## Solution

### Fix 1: Populate `ElementInfo.Path` Field
**File:** `internal/inspector/element_info.go`

Added a `getPath()` function to extract the Fiber-generated path from VNode's Key:

```go
// getPath extracts the hierarchical path from a VNode's Fiber Key
func getPath(vnode rtui.VNode) string {
    if keyer, ok := vnode.(interface{ Key() string }); ok {
        vnodeKey := keyer.Key()
        // Check if this is a path-based key (set by Fiber reconciliation)
        if vnodeKey != "" && strings.HasPrefix(vnodeKey, "/root/") {
            // Remove "/root/" prefix for cleaner display
            if len(vnodeKey) > 6 {
                return vnodeKey[6:]
            }
            return vnodeKey
        }
    }
    return ""
}
```

Then updated `ExtractElementInfo()` to call it:

```go
info.Path = getPath(vnode)  // ✨ Extract Fiber path
```

### Fix 2: Sync Fiber Path to VNode After Re-render
**File:** `internal/reconciler/diff.go`

When cloning existing Fibers (re-render scenario), the new VNode instance needs to have the Fiber path key set:

```go
// In cloneExistingFiber():
if current.Path != "" && strings.HasPrefix(current.Path, "/root/") {
    // Only sync if it's an auto-generated path key (not a user key)
    // User keys don't start with "/root/"
    vnode.SetKey(current.Path)
}
```

**Important Design Decision:**
- **Auto-generated path keys** (Priority 3, start with `/root/`) are synced to VNode
- **User-provided keys** (Priority 1, like `"my-button"`) are preserved as-is

This ensures:
1. Reconciliation works correctly (uses user keys when provided)
2. Inspector can display hierarchical paths (uses auto-generated path keys)

## Changes Made

### 1. `internal/inspector/element_info.go`
- Added `strings` import
- Added `getPath()` function to extract Fiber-generated paths
- Updated `ExtractElementInfo()` to populate `Path` field

### 2. `internal/reconciler/diff.go`
- Modified `cloneExistingFiber()` to sync auto-generated path keys to new VNode instances
- Only syncs keys that start with `/root/` (auto-generated), preserving user keys

### 3. `internal/inspector/element_path_test.go` (new file)
- Comprehensive tests for path extraction
- Tests cover various scenarios:
  - Full hierarchical paths
  - Single layer paths
  - User-provided keys (should not extract as paths)
  - Empty keys
  - Invalid path formats

## Test Results

All tests pass:
- ✅ `TestGetPath_FiberKey` - All 6 subtests
- ✅ `TestExtractElementInfo_WithPath` - Path field is populated correctly
- ✅ `TestExtractElementInfo_NoPath` - Empty path when no Fiber key
- ✅ `TestExtractElementInfo_UserKey` - User keys don't create paths
- ✅ `TestFormatSidebarWithPath` - Sidebar displays the path correctly
- ✅ All reconciler tests pass (including `TestFiberSync_SiblingChainIntegrity`)

## Display Format

The Inspector sidebar now displays paths in the new slash-based format:

```
┌─ UI Inspector ──────────────────────┐
│ Element: TextVNode                  │
│ ├── Type                            │
│ │   VNode Type: TextVNode           │
│ │   Tag: text                       │
│ ├── Path                            │
│ │   base[0]/vstack[0]/text[0]       │
└──────────────────────────────────────┘
```

## Key Insights

### Mixed Key Strategy Integration
The fix properly integrates with the three-tier priority key strategy:

1. **Priority 1: User-provided keys**
   - Example: `.Key("my-button")`
   - VNode.Key() returns: `"my-button"`
   - Fiber.Key: `"my-button"`
   - Fiber.Path: `"/root/.../.../my-button"`
   - Not synced to VNode (user key is preserved)

2. **Priority 2: Dynamic list keys**
   - Required for dynamic children
   - Same behavior as user keys

3. **Priority 3: Auto-generated path keys**
   - Example: `/root/base[0]/vstack[0]/panel[0]`
   - VNode.Key() returns: `/root/base[0]/vstack[0]/panel[0]`
   - Fiber.Key: `/root/base[0]/vstack[0]/panel[0]`
   - Fiber.Path: `/root/base[0]/vstack[0]/panel[0]`
   - Synced to VNode (for Inspector display)

### Why Distinguish User Keys vs Auto-generated Keys?

**User keys** must be preserved exactly as provided for reconciliation to work correctly:
```go
// First render
vnode.SetKey("item-1")

// Re-render (new VNode instance)
// We MUST use "item-1", not "/root/.../item-1"
// Otherwise reconciliation will think it's a different item
```

**Auto-generated path keys** are safe to sync because they're deterministic:
- Same tree structure = same path
- Used for debugging and Inspector display
- Not involved in user-provided key logic
