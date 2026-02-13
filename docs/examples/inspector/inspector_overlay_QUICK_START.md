# Inspector Overlay - Quick Start Guide ✅

## Run the Example

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go
```

## Keyboard Shortcuts (All Working!)

### Toggle Inspector
- **F12** - Show/hide Inspector ✅
- **Ctrl+D** - Show/hide Inspector ✅
- **[I] button** - Show/hide Inspector ✅

### Move Inspector Panel (when visible)
- **Alt+H** or **Alt+←** - Move left ✅
- **Alt+L** or **Alt+→** - Move right ✅
- **Alt+K** or **Alt+↑** - Move up ✅
- **Alt+J** or **Alt+↓** - Move down ✅

### Navigate Tree (Elements tab, when Inspector visible)
- **↑↓ Arrow keys** - Navigate tree nodes ✅
- **PageUp/PageDown** - Scroll tree view ✅
- **Home/End** - Jump to top/bottom ✅

### Switch Inspector Tabs (when Inspector visible)
- **1** - Elements tab ✅
- **2** - Console tab ✅
- **3** - Performance tab ✅
- **4** - Diagnostics tab ✅
- **5** - Network tab ✅

## What You'll See

```
┌─────────────────────────────────────────────────────────────┐
│ ╔═ INSPECTOR ═╗                                            │
│ F12:关闭 | Alt+H/J/K/L:移动                                │
│ [Elements] | Console | Performance | Diagnostics | Network  │
│ 📦 Layout Tree                                              │
│ └── 📦LayoutNode                                            │
│   ├── 📦ElementVNode(Root Node)                            │
│   ├── 📦ElementVNode(Node 1)                               │
└─────────────────────────────────────────────────────────────┘
```

## Test It Yourself

```bash
# Run tests
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay

# Test F12 toggle
go test -v -run TestInspectorOverlayWithF12

# Test movement keys
go test -v -run TestInspectorMovementKeys

# Test arrow movement
go test -v -run TestInspectorArrowMovement

# Test automatic routing
go test -v -run TestAutomaticEventRouting
```

## Debug Mode

```bash
export TUI_DEBUG_UI=true
export TUI_DEBUG_INSPECTOR=true
go run main.go
```

**Expected debug output**:
```
[APP] KeyMap found handler for key 'f12'
[APP] Inspector toggled: now visible=true
[Inspector] Moved up to y=4
[APP] Routing key 'down' to Inspector (visible=true)
[APP] Inspector handled key 'down'
```

## Common Issues

### Q: F12 doesn't toggle Inspector
**A**: Make sure you're using framework app directly:
```go
fwApp := framework.NewApp()
fwApp.SetInspector(inspector)        // ← REQUIRED
fwApp.SetupInspectorShortcut()       // ← REQUIRED
```

### Q: Arrow keys don't navigate tree
**A**: Make sure Inspector is **visible** first! Press F12 to show it.

### Q: Alt+K/J doesn't move panel
**A**: Works when Inspector is visible. The keys are handled by Inspector.HandleKeyEvent(), not keyMap.

## Summary

✅ All keyboard shortcuts working
✅ F12/Ctrl+D toggle Inspector
✅ Alt+H/J/K/L move panel
✅ Arrow keys navigate tree
✅ Automatic event routing functional
✅ Tests verify all features

**The Inspector overlay example is fully functional!** 🎉
