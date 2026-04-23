# Inspector Path Display Fix

## Problem

The Inspector's Elements panel was displaying the old dot-notation path format instead of the new Fiber-generated slash-based path format:

**Before (old format):**
```
Path: vstack[0].bordered[0].hstack[0].text
```

**Expected (new format):**
```
Path: base[0]/vstack[0]/bordered[0]/hstack[0]/text[0]
```

## Root Cause

The `ElementInfo` struct has a `Path` field, but the `ExtractElementInfo()` function was **not populating it**. The sidebar displays the `info.Path` value, so it remained empty or showed the old fallback format.

## Solution

Added a `getPath()` function in `internal/inspector/element_info.go` that extracts the Fiber-generated hierarchical path from the VNode's Key:

```go
// getPath extracts the hierarchical path from a VNode's Fiber Key
// Fiber reconciliation sets VNode keys to path-based keys like /root/base[0]/vstack[0]/panel[0]
func getPath(vnode rtui.VNode) string {
    if keyer, ok := vnode.(interface{ Key() string }); ok {
        vnodeKey := keyer.Key()
        // Check if this is a path-based key (set by Fiber reconciliation)
        if vnodeKey != "" && strings.HasPrefix(vnodeKey, "/root/") {
            // Use the Fiber-generated path as our display path
            // Remove the "/root/" prefix for cleaner display
            // /root/base[0]/vstack[0]/panel[0] → base[0]/vstack[0]/panel[0]
            if len(vnodeKey) > 6 { // "/root/" is 6 characters
                return vnodeKey[6:] // Skip "/root/" prefix
            }
            return vnodeKey
        }
    }
    return ""
}
```

Then updated `ExtractElementInfo()` to call this new function:

```go
// Extract basic identification
info.Type = getTypeName(vnode)
info.Tag = getTag(vnode)
info.Key = getKey(vnode)
info.Label = getLabel(vnode)
info.Path = getPath(vnode)  // ✨ NEW: Extract Fiber path
```

## Changes Made

1. **`internal/inspector/element_info.go`**:
   - Added `strings` import
   - Added `getPath()` function to extract Fiber-generated paths
   - Updated `ExtractElementInfo()` to populate the `Path` field

2. **`internal/inspector/element_path_test.go`** (new file):
   - Added comprehensive tests for path extraction
   - Tests cover various scenarios:
     - Full hierarchical paths
     - Single layer paths
     - User-provided keys (should not extract as paths)
     - Empty keys
     - Invalid path formats

## Test Results

All new path extraction tests pass:

- ✅ `TestGetPath_FiberKey` - All 6 subtests pass
- ✅ `TestExtractElementInfo_WithPath` - Path field is populated correctly
- ✅ `TestExtractElementInfo_NoPath` - Empty path when no Fiber key
- ✅ `TestExtractElementInfo_UserKey` - User keys don't create paths
- ✅ `TestFormatSidebarWithPath` - Sidebar displays the path correctly

## Display Format

The Inspector sidebar now displays paths in the new slash-based format:

```
┌─ UI Inspector ──────────────────────┐
│ Element: Panel                      │
│ ├── Type                            │
│ │   VNode Type: ElementVNode        │
│ │   Tag: panel                      │
│ │   Key: /root/base[0]/vstack[0]/panel[0] │
│ ├── Path                            │
│ │   base[0]/vstack[0]/panel[0]       │
└──────────────────────────────────────┘
```

## Notes

- The `Key` field shows the full Fiber key (including `/root/` prefix)
- The `Path` field shows the display path (with `/root/` prefix removed for cleaner display)
- User-provided keys (like `my-button`) are still shown in the `Key` field but don't create a `Path`
- This matches the path extraction logic used in `TreeView.buildTree()`

## Related Files

- `internal/inspector/element_info.go` - Core fix
- `internal/inspector/element_path_test.go` - Test coverage
- `internal/inspector/sidebar.go` - Displays the path
- `internal/inspector/tree_view.go` - Also uses similar path extraction logic
