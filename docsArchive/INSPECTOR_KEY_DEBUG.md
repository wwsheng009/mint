# Inspector Key Debug Feature

## What It Does

The Inspector now has a **Key Debug Mode** that shows exactly what keys you're pressing. This helps diagnose why certain key combinations might not work.

## How to Use

### Step 1: Run the Inspector Overlay
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go
```

### Step 2: Press F12 to Show Inspector
The Inspector overlay will appear.

### Step 3: Enable Key Debug Mode
Press **Ctrl+D** while the Inspector is visible.

A new line will appear in the Inspector title bar showing:
```
🔍 按键调试: key='k' Alt+
```

### Step 4: Press Keys and Watch the Display

Try pressing different key combinations:

| Key Press | Display Shows | What It Means |
|----------|---------------|---------------|
| **k** | `key='k' 无` | Plain "k" key, no modifiers |
| **Alt+k** | `key='k' Alt+` | "k" key with Alt modifier |
| **Alt+K** | `key='k' Alt+` | Same as Alt+k (character key) |
| **Alt+↑** | `key='up' Alt+` | Arrow up with Alt modifier |
| **Ctrl+d** | `key='d' Ctrl+` | "d" key with Ctrl modifier |
| **Shift+a** | `key='A' Shift+` | "A" key with Shift modifier |
| **Alt+Ctrl+x** | `key='x' Alt+Ctrl+` | "x" with Alt and Ctrl |

### Step 5: Disable Key Debug Mode
Press **Ctrl+D** again to toggle off.

## What the Display Means

```
🔍 按键调试: key='k' Alt+
```

- **key='k'** - The name or character of the key
  - For character keys: shows the letter (e.g., 'k', 'a', '1')
  - For special keys: shows the name (e.g., 'up', 'down', 'f12', 'enter')

- **Alt+** - Alt modifier was pressed
- **Ctrl+** - Ctrl modifier was pressed
- **Shift+** - Shift modifier was pressed
- **无** - No modifiers (Chinese for "none")

## Example Debugging Session

### Problem: Alt+K Doesn't Move Panel

**Step 1**: Enable key debug (Ctrl+D)
**Step 2**: Press Alt+K
**Step 3**: Check what it shows

**Scenario A: Shows `key='k' Alt+`**
- ✅ Correct! The key is being detected properly.
- Problem might be in keyMap routing or handler registration.
- Check framework logs for "KeyMap found handler" messages.

**Scenario B: Shows `key='k' 无` (no Alt)**
- ❌ Alt modifier not detected!
- This is a terminal/platform issue.
- Try using arrow keys instead (Alt+↑/↓/←/→)

**Scenario C: Shows nothing**
- ❌ Key event not reaching Inspector!
- Check if Inspector is visible.
- Check if another component is capturing the event.

**Scenario D: Shows `key='' Alt+` (empty key name)**
- ❌ Character key routing issue!
- This means the framework is sending empty string for character keys.
- Check framework/app.go routing code.

## Tips for Different Terminals

### Windows Terminal (Recommended)
- ✅ Best support for Alt+letter combinations
- ✅ Arrow keys work well with modifiers

### CMD / Legacy Console
- ⚠️ Some Alt+letter combinations may not work
- ✅ Arrow keys usually work

### PowerShell
- ✅ Generally good support
- ⚠️ May vary by version

### VS Code Integrated Terminal
- ⚠️ Some key combinations may be intercepted by VS Code
- Try using external terminal

## Keyboard Shortcuts Reference

### Movement Keys (when Inspector visible)
| Key | Action |
|-----|--------|
| **Alt+H** or **Alt+←** | Move panel left |
| **Alt+L** or **Alt+→** | Move panel right |
| **Alt+K** or **Alt+↑** | Move panel up |
| **Alt+J** or **Alt+↓** | Move panel down |

### Navigation Keys (Elements tab)
| Key | Action |
|-----|--------|
| **↑↓** | Navigate tree |
| **PageUp/PageDown** | Scroll tree |
| **Home/End** | Jump to top/bottom |
| **E** | Expand/collapse node |
| **Enter** | Select node |

### Inspector Controls
| Key | Action |
|-----|--------|
| **F12** | Show/hide Inspector |
| **Ctrl+D** | Toggle key debug mode |
| **1-5** | Switch tabs |
| **Tab** | Cycle through tabs (if implemented) |

## Implementation Details

### Fields Added to StandaloneInspector
```go
// Key debug info (for displaying what keys are being pressed)
lastKey      string  // Last key name received
lastAlt      bool    // Last Alt modifier state
lastCtrl     bool    // Last Ctrl modifier state
lastShift    bool    // Last Shift modifier state
showKeyDebug bool    // Show key debug info in UI
```

### Code Changes
1. **HandleKeyEvent()** - Stores key info on every key press
2. **buildOverlayContent()** - Adds debug line to title bar when enabled
3. **Ctrl+D** - Toggles `showKeyDebug` flag

### Why This Helps

Before this feature, if a key didn't work, you had to:
- Add debug logging to multiple files
- Recompile
- Run with environment variables
- Scan through verbose logs

Now you can:
- Press Ctrl+D
- Press the problematic key
- Immediately see what's being received

## Troubleshooting

**Q: Ctrl+D toggles Inspector off instead of enabling debug!**
A: Ctrl+D is also registered as a shortcut to toggle Inspector. Press it again to re-enable Inspector, then the debug mode will be active.

**Q: The debug display doesn't update!**
A: Make sure the Inspector is visible (F12). The debug display only appears when Inspector is shown.

**Q: It shows the wrong key!**
A: This might be a terminal encoding issue. Try a different terminal or use arrow keys instead.

**Q: Can I change the toggle key?**
A: Currently Ctrl+D is hardcoded. You can modify `standalone_inspector.go:1103` if needed.

## See Also

- `ALT_K_DEBUG_GUIDE.md` - Detailed guide for Alt+K movement keys
- `INSPECTOR_QUICK_REFERENCE.md` - Quick reference for all Inspector features
- `INSPECTOR_ALL_FIXES_SUMMARY.md` - Summary of all event routing fixes
