# Inspector TreeView Enhancements

## Issues Fixed

This document describes four critical fixes to the Inspector TreeView and tab switching functionality.

---

## 4. Tree Arrow Navigation Not Working ✅ FIXED

**Problem**: Arrow keys (↑↓) and PageUp/PageDown keys didn't visually update the TreeView focus.

**Root Cause**: The TreeView component had two issues:
1. The Inspector was creating a new TreeView component on every render, losing navigation state
2. The TreeView's `SetFocusIndex()` and navigation methods (`MoveUp()`, `MoveDown()`, etc.) updated internal state but didn't regenerate the display

**Detailed Problem**:
```go
// In standalone_inspector.go buildElementsTabContent():
// Every frame, this code ran:
si.treeViewComponent = display.NewTreeView().
    FromLines(allLines).
    Build().(*display.TreeView)
si.treeViewComponent.SetFocusIndex(oldFocusIndex) // ❌ Sets state but doesn't update display

// In treeview.go Build():
func (b *TreeViewBuilder) Build() ui.VNode {
    // Creates text nodes with CURRENT state
    if i == b.node.focusIndex {
        // Create yellow highlighted node
    }
    // ...returns static VNodes
}

// In treeview.go SetFocusIndex():
func (t *TreeView) SetFocusIndex(index int) {
    t.focusIndex = index
    // ❌ Doesn't regenerate text nodes!
}
```

The result was that `Build()` created text nodes based on the state at build time, but subsequent `SetFocusIndex()` calls only updated the internal state without regenerating the visual display.

**Fix**: Added `regenerateDisplay()` method that recreates the text nodes whenever the state changes:

```go
// New method in treeview.go:
func (t *TreeView) regenerateDisplay() {
    // Recreate text nodes for each line with CURRENT focus/selection state
    var lineNodes []ui.VNode
    for i, line := range t.lines {
        if i == t.selectedIdx {
            // Cyan background
        } else if i == t.focusIndex {
            // Yellow text
        } else {
            // White text
        }
    }
    result := ui.VStack(lineNodes...)
    t.SetChildren([]ui.VNode{result}) // Update the display
}

// Updated SetFocusIndex() to call regenerateDisplay():
func (t *TreeView) SetFocusIndex(index int) {
    if index >= 0 && index < len(t.lines) {
        t.focusIndex = index
        t.ensureVisible()
        t.regenerateDisplay() // ✅ Now updates the display!
    }
}

// Similarly updated all navigation methods:
func (t *TreeView) MoveUp() {
    if t.focusIndex > 0 {
        t.focusIndex--
        t.ensureVisible()
        t.regenerateDisplay() // ✅ Updates display
    }
}
// ... MoveDown(), PageUp(), PageDown(), Home(), End(), SelectCurrent(), etc.
```

Also fixed the Inspector to properly preserve state across rebuilds:

```go
// In standalone_inspector.go buildElementsTabContent():
// Preserve current navigation state BEFORE creating new component
currentFocusIndex := si.treeViewComponent.GetFocusIndex()
currentSelectedIdx := si.treeViewComponent.GetSelectedLine().NodeID
currentScrollOffset := si.treeScrollOffset

// Create new component
si.treeViewComponent = display.NewTreeView().
    FromLines(allLines).
    Build().(*display.TreeView)

// Restore state (this now triggers regenerateDisplay())
si.treeViewComponent.SetFocusIndex(currentFocusIndex) // ✅ Updates display
si.treeViewComponent.SelectLine(currentSelectedIdx)     // ✅ Updates display
si.treeViewComponent.SetScrollOffset(currentScrollOffset)
```

**Result**: Arrow keys now properly update the visual display:
- ✅ Up/Down arrows move focus with yellow highlighting
- ✅ PageUp/PageDown scroll by pages
- ✅ Home/End jump to top/bottom
- ✅ Enter selects with cyan background
- ✅ All navigation updates the display in real-time

---

### 1. TreeView Focus Feedback ✅ FIXED

**Problem**: TreeView items had no visual feedback when focused or selected.

**Root Cause**: The TreeView component was using markdown-style style tags `[reverse]` and `[bold]` which were not being parsed or rendered by the styling system.

**Fix**: Updated `components/display/treeview.go` to use actual style objects instead of markdown tags.

**Before**:
```go
if i == b.node.selectedIdx {
    lineNodes = append(lineNodes, ui.Text(fmt.Sprintf("[reverse]%s[/reverse]", fullLine)))
} else if i == b.node.focusIndex {
    lineNodes = append(lineNodes, ui.Text(fmt.Sprintf("[bold]%s[/bold]", fullLine)))
}
```

**After**:
```go
if i == b.node.selectedIdx {
    // Selected line - cyan background with black text
    lineNodes = append(lineNodes, app.NewTextBuilder(fullLine).
        Style(style.NewStyle().
            Foreground(style.Black).
            Background(style.Cyan).
            Bold(true)).
        Build())
} else if i == b.node.focusIndex {
    // Focused line - yellow text
    lineNodes = append(lineNodes, app.NewTextBuilder(fullLine).
        Style(style.NewStyle().
            Foreground(style.Yellow).
            Bold(true)).
        Build())
}
```

**Result**: Clear visual distinction between:
- **Normal items**: White text on default background
- **Focused items**: Yellow bold text (distinct highlight)
- **Selected items**: Black text on cyan background (clear selection state)

---

### 2. Background Color Conflict ✅ FIXED

**Problem**: Selected TreeView items used reverse video (`[reverse]`), which swapped foreground/background colors. This caused visual conflicts with other controls that also use reverse video (like focused buttons).

**Root Cause**: The reverse video style is context-dependent - it swaps whatever the current foreground/background colors are. When TreeView is rendered over other components, the result is unpredictable.

**Fix**: Use explicit colors instead of reverse video.

**Selection Color Scheme**:
- **Background**: Cyan (distinct from most UI elements)
- **Foreground**: Black (high contrast on cyan)
- **Bold**: Yes (for additional emphasis)

This color scheme was chosen because:
- ✅ Cyan is rarely used by other controls
- ✅ Black-on-cyan has excellent contrast
- ✅ Distinct from button focus styles (usually reverse video)
- ✅ Works well on both light and dark terminal backgrounds

---

### 3. Number Keys (1-5) Not Detected ✅ FIXED

**Problem**: Pressing number keys 1-5 did not switch Inspector tabs as documented.

**Root Cause**: Method signature mismatch between framework and Inspector:

**Framework** (framework/app.go:580-582):
```go
if inspectorObj, ok := a.inspector.(interface {
    HandleKeyEvent(key string, alt, ctrl bool) bool  // ❌ 3 parameters
}); ok {
    if inspectorObj.HandleKeyEvent(key, false, false) {
        // ...
    }
}
```

**Inspector** (internal/inspector/standalone_inspector.go:1112):
```go
func (si *StandaloneInspector) HandleKeyEvent(key string, alt bool, ctrl bool, shift bool) bool {
    // ... ✅ 4 parameters
}
```

The type assertion `inspector.(interface { ... })` was **failing** because the interface didn't match the actual method signature. This meant:
- `switchInspectorTab()` was never calling `HandleKeyEvent()`
- Number keys 1-5 were registered but never routed to the Inspector
- Tab switching was completely broken

**Fix**: Updated framework to use correct 4-parameter signature:

```go
if inspectorObj, ok := a.inspector.(interface {
    HandleKeyEvent(key string, alt, ctrl, shift bool) bool  // ✅ 4 parameters
}); ok {
    if inspectorObj.HandleKeyEvent(key, false, false, false) {
        // ...
    }
}
```

**Result**: Number keys 1-5 now correctly switch tabs:
- `1` → Elements tab
- `2` → Console tab
- `3` → Performance tab
- `4` → Diagnostics tab
- `5` → Network tab

---

## Testing

### Test Focus Feedback

1. Run the Inspector demo:
   ```bash
   cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
   go run main.go
   ```

2. Press F12 to show Inspector

3. Verify TreeView appearance:
   - ✅ Normal items: White text
   - ✅ Navigate with ↑/↓ - focused item shows yellow bold text
   - ✅ Press Enter - selected item shows black-on-cyan background

### Test Tab Switching

1. Run the number key test:
   ```bash
   cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
   go run number_key_test.go
   ```

2. Press F12 to show Inspector

3. Press keys 1-5:
   - ✅ `1` switches to Elements tab
   - ✅ `2` switches to Console tab
   - ✅ `3` switches to Performance tab
   - ✅ `4` switches to Diagnostics tab
   - ✅ `5` switches to Network tab

4. Enable debug mode:
   ```bash
   TUI_DEBUG_UI=true go run number_key_test.go
   ```

   You should see:
   ```
   [APP] Key received: '1' routing to Inspector
   [Inspector] Key received: key='1' switching to tab 0
   [APP] Inspector switched to tab 1
   ```

---

## Files Modified

1. **`components/display/treeview.go`**
   - Fixed focus/selection rendering
   - Added proper imports (app, style packages)
   - Changed from markdown tags to style objects

2. **`framework/app.go`**
   - Fixed HandleKeyEvent signature mismatch
   - Updated `switchInspectorTab()` to use 4-parameter call

3. **`examples/ui_demos/demo2_runtime_internals/inspector_overlay/number_key_test.go`**
   - Added test program for number key detection

---

## Design Decisions

### Color Choice for Selection

**Why cyan background?**
- Distinct from default terminal colors
- High contrast with black text (WCAG AAA compliant)
- Not commonly used by other components
- Works on both light and dark terminals
- Professional appearance

**Why not reverse video?**
- Context-dependent (swaps current fg/bg)
- Conflicts with other controls using reverse
- Unpredictable results when rendered over other elements
- Less accessible for color-blind users

### Focus vs Selection

The TreeView now distinguishes between:
- **Focus**: Yellow text (temporary navigation target)
- **Selection**: Cyan background (persistent user choice)

This follows standard UI conventions where:
- Focus follows keyboard navigation (↑/↓ arrows)
- Selection requires user action (Enter key)
- Both can be independent (item selected ≠ item focused)

---

## Related Issues

- [Inspector Quick Reference](INSPECTOR_QUICK_REFERENCE.md) - Tab switching shortcuts
- [TreeView Navigation](TREEVIEW_NAVIGATION_WORKING.md) - Navigation implementation
- [Key Handling Guide](../../guide/key-handling/KEY_HANDLING_COMPLETE_GUIDE.md) - Complete key handling reference

---

## Summary

All four issues have been resolved:

1. ✅ **TreeView focus feedback** - Clear visual distinction for focused/selected items
2. ✅ **Background color conflict** - Cyan background eliminates conflicts
3. ✅ **Number key detection** - Fixed signature mismatch, tabs 1-5 now work
4. ✅ **Tree arrow navigation** - All navigation keys now properly update the display

The Inspector is now fully functional with:
- Clear visual feedback for all interactions
- Proper tab switching via number keys
- Working arrow key navigation
- Real-time visual updates
- No visual conflicts with other controls

---

## Files Modified (Summary)

1. **`components/display/treeview.go`** (~80 lines changed)
   - Fixed focus/selection rendering with proper style objects
   - Added `regenerateDisplay()` method to update visuals when state changes
   - Updated all navigation methods to call `regenerateDisplay()`
   - Fixed pre-existing bug: `string(node.Type())` → `fmt.Sprintf("%d", node.Type())`

2. **`framework/app.go`** (5 lines changed)
   - Fixed `HandleKeyEvent` signature from 3 to 4 parameters
   - Updated `switchInspectorTab()` to call with correct parameters

3. **`internal/inspector/standalone_inspector.go`** (~30 lines changed)
   - Fixed TreeView component state management
   - Properly preserve navigation state across rebuilds
   - Use component's children directly instead of manual rendering

4. **`docs/examples/inspector/INSPECTOR_TREEVIEW_FIX.md`** (updated)
   - Complete documentation of all four fixes
   - Testing instructions
   - Design decisions explained
