# Inspector Overlay - Quick Reference ✅

## Run It

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go
```

## What You'll See

**Inspector Hidden (Default)**:
- Background demo2 app is active
- Buttons are focusable
- All keys work normally

**Inspector Visible (F12 pressed)**:
- Inspector overlay appears
- Background app is still visible but not interactive
- Inspector captures all keyboard input

## Keyboard Shortcuts

### Toggle Inspector
| Key | Action |
|-----|--------|
| **F12** | Show/hide Inspector ✅ |
| **Ctrl+D** | Show/hide Inspector ✅ |
| **[I] button** | Show/hide Inspector ✅ |

### Move Inspector Panel (when visible)
| Key | Action |
|-----|--------|
| **Alt+H** | Move left ✅ |
| **Alt+L** | Move right ✅ |
| **Alt+K** | Move up ✅ |
| **Alt+J** | Move down ✅ |
| **Alt+←** | Move left ✅ |
| **Alt+→** | Move right ✅ |
| **Alt+↑** | Move up ✅ |
| **Alt+↓** | Move down ✅ |

### Navigate Tree (Elements tab, when Inspector visible)
| Key | Action |
|-----|--------|
| **↑** | Navigate up ✅ |
| **↓** | Navigate down ✅ |
| **PageUp** | Scroll up ✅ |
| **PageDown** | Scroll down ✅ |
| **Home** | Jump to top ✅ |
| **End** | Jump to bottom ✅ |
| **E** | Expand/collapse node |
| **Enter** | Select node |

### Switch Tabs (when Inspector visible)
| Key | Tab |
|-----|-----|
| **1** | Elements ✅ |
| **2** | Console ✅ |
| **3** | Performance ✅ |
| **4** | Diagnostics ✅ |
| **5** | Network ✅ |

### Debug Tools (when Inspector visible)
| Key | Action |
|-----|--------|
| **Ctrl+D** | Toggle key debug mode ✅ NEW |
| Press any key | See what key was detected ✅ NEW |
| **4** | Diagnostics ✅ |
| **5** | Network ✅ |

## Modal Behavior

### When Inspector is HIDDEN:
- ✅ Background buttons work normally
- ✅ Tab/Enter keys work normally
- ✅ All keys go to background app

### When Inspector is VISIBLE:
- ✅ Arrow keys navigate tree (not buttons)
- ✅ Tab is captured by Inspector
- ✅ Enter is captured by Inspector
- ✅ Background app can't steal focus
- ✅ Inspector is truly modal

## Debug Mode

### Built-in Key Debug (Recommended) ✅ NEW

Press **Ctrl+D** while Inspector is visible to enable key debug mode.

**What you'll see**:
```
🔍 按键调试: key='k' Alt+
```

This shows exactly what key was received, making it easy to diagnose issues.

### Verbose Logging

See what's happening under the hood:
```bash
export TUI_DEBUG_UI=true
export TUI_INSPECTOR_VERBOSE=true
go run main.go
```

**Expected output**:
```
[APP] KeyMap found handler for key 'f12'
[APP] Inspector toggled: now visible=true
[APP] Routing key 'down' to Inspector (visible=true)
[Inspector] Visible mode: capturing key 'down'
```

### Platform Input Debug

See what the platform layer receives:
```bash
export TUI_DEBUG_INPUT=true
go run main.go
```

**Expected output when pressing Alt+K**:
```
[WIN INPUT] VK=0x4B UChar=k Special=0 Key=k Modifiers=0x2(Alt=true,Ctrl=false,Shift=false)
```

## Test the Fix

Run all tests to verify everything works:
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay

go test -v -run TestInspectorOverlayWithF12      # Test F12 toggle
go test -v -run TestInspectorMovementKeys        # Test Alt+H/J/K/L
go test -v -run TestInspectorArrowMovement        # Test Alt+Arrows
go test -v -run TestAutomaticEventRouting        # Test automatic routing
```

All tests pass! ✅

## Summary

✅ **F12 toggle works**
✅ **Movement keys work**
✅ **Tree navigation works**
✅ **Inspector is modal when visible**
✅ **Background app can't interfere**

**Everything is working correctly!** 🎉
