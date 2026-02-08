# TreeView Navigation Support

## Overview

The TreeView component now supports full keyboard navigation for browsing and interacting with tree structures. This enables users to navigate through hierarchical data (like the Inspector's layout tree) using intuitive keyboard shortcuts.

## Features

### 1. Navigation Methods

The TreeView component provides the following navigation capabilities:

#### **Arrow Key Navigation**
- **Up Arrow** - Move focus to previous line
- **Down Arrow** - Move focus to next line

#### **Page Navigation**
- **Page Up** - Move up by one page (viewport height)
- **Page Down** - Move down by one page (viewport height)
- **Home** - Jump to first line (top of tree)
- **End** - Jump to last line (bottom of tree)

#### **Tree Interaction**
- **E key** - Toggle expand/collapse for focused node
- **Enter key** - Select the currently focused node

### 2. Visual Feedback

- **Focused Line** - Highlighted in bold yellow to show current position
- **Selected Line** - Highlighted with reverse video (inverted colors)
- **Scroll Position** - Automatically adjusted to keep focused line visible

### 3. State Management

The TreeView maintains the following state:

```go
type TreeView struct {
    lines       []TreeViewLine  // All tree lines
    focusIndex  int              // Currently focused line
    selectedIdx int              // Currently selected line
    scrollOffset int             // Current scroll position
    viewportHeight int           // Visible height
    expandState map[int]bool     // Expand/collapse state per node
    builder     *TreeViewBuilder // Reference for rebuilds
}
```

## API Reference

### Public Methods

#### Navigation
- `MoveUp()` - Move focus up one line
- `MoveDown()` - Move focus down one line
- `PageUp()` - Move focus up one page
- `PageDown()` - Move focus down one page
- `Home()` - Move focus to first line
- `End()` - Move focus to last line
- `HandleKey(key SpecialKey, r rune) bool` - Main keyboard event handler

#### Selection & Focus
- `SelectCurrent()` - Select the focused line
- `FocusLine(index int) bool` - Move focus to specific line
- `GetFocusIndex() int` - Get current focus index
- `SetFocusIndex(index int)` - Set focus index
- `GetFocusedLine() TreeViewLine` - Get focused line data
- `GetSelectedLine() TreeViewLine` - Get selected line data
- `ClearSelection()` - Clear selection
- `HasSelection() bool` - Check if has selection

#### Scrolling
- `GetScrollOffset() int` - Get scroll offset
- `SetScrollOffset(offset int)` - Set scroll offset
- `SetViewportHeight(height int)` - Set viewport height for calculations

#### Expand/Collapse
- `ToggleExpandCurrent()` - Toggle expand/collapse for focused node
- `IsExpanded(nodeID int) bool` - Check if node is expanded
- `SetExpanded(nodeID int, expanded bool)` - Set expand state

#### Rebuilding
- `Rebuild() VNode` - Rebuild the tree view with current state

## Integration with Inspector

The StandaloneInspector now uses the TreeView component for navigation:

```go
type StandaloneInspector struct {
    treeView       *TreeView
    treeViewComponent *display.TreeView // New navigation component
    // ...
}
```

### Keyboard Event Handling

When a keyboard event occurs in the Inspector with the Elements tab active:

1. Event is routed to `HandleKeyEvent()`
2. If TreeView component exists, keys are delegated to it:
   ```go
   if si.treeViewComponent != nil {
       handled := si.treeViewComponent.HandleKey(platformKey, r)
       if handled {
           // Sync scroll offset back to Inspector
           si.treeScrollOffset = si.treeViewComponent.GetScrollOffset()
           return true
       }
   }
   ```

3. Fallback to original scrolling behavior if TreeView not available

### Rendering

The Inspector renders visible tree lines with highlighting:

```go
for i := startLine; i < endLine; i++ {
    line := allLines[i]

    if i == selectedIdx {
        // Selected line - reverse video
        lineNodes = append(lineNodes, app.NewTextBuilder(line).
            Style(style.NewStyle().Reverse(true)).
            Build())
    } else if i == focusIndex {
        // Focused line - bold yellow
        lineNodes = append(lineNodes, app.NewTextBuilder(line).
            Style(style.NewStyle().Bold(true).Foreground(style.Yellow)).
            Build())
    } else {
        // Normal line
        lineNodes = append(lineNodes, ui.Text(line))
    }
}
```

## Usage Example

### Creating a TreeView with Navigation

```go
// Create TreeView from pre-formatted lines
treeView := display.NewTreeView().
    FromLines(treeLines).
    ExpandLevel(1).           // Expand first level by default
    ShowIcons(true).          // Show type icons
    Compact(false).           // Use full display
    Build().(*display.TreeView)

// Set viewport height for scrolling calculations
treeView.SetViewportHeight(20)

// Handle keyboard events
func handleKeyEvent(key string) {
    var platformKey platform.SpecialKey
    var r rune

    switch key {
    case "up":
        platformKey = platform.KeyUp
    case "down":
        platformKey = platform.KeyDown
    case "e":
        r = 'e'
    }

    if treeView.HandleKey(platformKey, r) {
        // Key was handled, refresh display
        scrollOffset := treeView.GetScrollOffset()
        focusIndex := treeView.GetFocusIndex()
        // ... update UI
    }
}
```

## Implementation Details

### Auto-Scroll

The `ensureVisible()` method automatically scrolls the tree when the focused line moves outside the viewport:

```go
func (t *TreeView) ensureVisible() {
    if t.viewportHeight <= 0 {
        return
    }

    // Scroll down if focus is below viewport
    if t.focusIndex >= t.scrollOffset+t.viewportHeight {
        t.scrollOffset = t.focusIndex - t.viewportHeight + 1
    }

    // Scroll up if focus is above viewport
    if t.focusIndex < t.scrollOffset {
        t.scrollOffset = t.focusIndex
    }
}
```

### Builder Pattern

The TreeView uses a builder pattern for configuration:

```go
builder := &TreeViewBuilder{
    node: &TreeView{
        // Initialize with defaults
        focusIndex:   0,
        selectedIdx:  -1,
        expandState:   make(map[int]bool),
    },
}

// Configure builder
builder.
    ExpandLevel(1).
    ShowIcons(true).
    ShowLineNumbers(false).
    Compact(false)

// Build and return
return builder.Build()
```

## Testing

Run the Inspector standalone test to verify navigation:

```bash
go test -v ./examples/ui_demos/demo2_runtime_internals -run TestInspectorStandalone
```

Expected behavior:
- Tree displays with proper formatting
- Arrow keys navigate between nodes
- Page Up/Down scrolls by page
- Home/End jump to top/bottom
- E key toggles expand/collapse (future enhancement)
- Enter selects nodes (future enhancement)

## Future Enhancements

1. **Expand/Collapse Integration** - Connect ToggleExpand to tree structure
2. **Node Selection Callbacks** - Add event handlers for selection changes
3. **Mouse Support** - Add click-to-select and double-click-to-expand
4. **Search/Filter** - Add ability to search and filter tree nodes
5. **Multi-Selection** - Support selecting multiple nodes with Ctrl+Click

## Files Modified

- `components/display/treeview.go` - Added navigation methods and state management
- `internal/inspector/standalone_inspector.go` - Integrated TreeView navigation

## Files Created

- `TREEVIEW_NAVIGATION.md` - This documentation
