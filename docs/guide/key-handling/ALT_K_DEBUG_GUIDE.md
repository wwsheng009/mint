# Alt+K Movement Keys Debug Guide

## Summary of Fixes

Three issues were fixed to support Alt+J/K/H/L movement keys:

### 1. Shift Modifier Support (framework/event/pump.go)
**Issue**: Shift modifier wasn't being set in KeyEvent.Key.Shift field
**Fix**: Added `ev.Key.Shift = true` when ModShift is detected

```go
if raw.Modifiers&platform.ModShift != 0 {
    ev.Key.Shift = true  // NEW: This was missing!
    ev.Modifiers |= ModShift
}
```

### 2. Character Key Routing to Inspector (framework/app.go)
**Issue**: When routing to Inspector, only `keyEv.Key.Name` was used, which is empty for character keys
**Fix**: Check both Name (for special keys) and Rune (for character keys)

```go
// Use Name for special keys, Rune for character keys
var keyName string
if keyEv.Key.Name != "" {
    keyName = keyEv.Key.Name
} else if keyEv.Key.Rune > 0 {
    keyName = string(keyEv.Key.Rune)  // NEW: This handles Alt+K
}
```

### 3. Debug Output Added
Added debug logging to help diagnose issues:
- framework/event/handler.go: Shows combo string being built
- runtime/platform/input_windows.go: Shows what Windows receives

## How Alt+K Works

When you press Alt+K on Windows console:

1. **Windows Input Layer** (input_windows.go):
   - VirtualKeyCode = 0x4B (VK_K)
   - UChar = 'k'
   - ControlKeyState includes Alt modifier (0x0002)
   - `virtualKeyToSpecial(0x4B)` returns KeyUnknown (no case for K)
   - Falls through to: `input.Key = rune('k')`
   - Sets `input.Modifiers |= ModAlt`

2. **Event Pump** (pump.go):
   - Creates KeyEvent with:
     - Key.Rune = 'k'
     - Key.Name = ""
     - Key.Alt = true
     - Modifiers = ModAlt

3. **KeyMap Lookup** (handler.go):
   - `buildComboString()` creates "alt+k" from Rune='k' + Alt modifier
   - Looks up bindings["alt+k"]
   - Finds the registered handler

4. **Handler Execution** (app.go):
   - Calls `moveInspector(0, -1)` to move panel up
   - Panel position changes: Y 5 → 4

## Test Results

```
✅ TestInspectorMovementKeys PASS
   - Alt+K moved panel up: Y 5 → 4
   - Alt+J moved panel down: Y 4 → 5
   - Alt+H moved panel left: X 80 → 78
   - Alt+L moved panel right: X 78 → 80
```

## How to Debug Your Issue

If Alt+K doesn't work in your terminal:

### Step 1: Run with Debug Output
```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
set TUI_DEBUG_INPUT=true
set TUI_DEBUG_UI=true
set TUI_INSPECTOR_VERBOSE=true
go run main.go
```

### Step 2: Press Alt+K and Look For:
```
[WIN INPUT] VK=0x4B UChar=k Special=0 Key=k Modifiers=0x2(Alt=true,Ctrl=false,Shift=false)
[KeyMap] Lookup: Rune=k Name="" Modifiers=2 Combo="alt+k"
[KeyMap] Found handler for combo 'alt+k'
```

### Step 3: What You Should See:
- ✅ `VK=0x4B` (K key virtual key code)
- ✅ `UChar=k` (character)
- ✅ `Modifiers=0x2` (Alt modifier)
- ✅ `Combo="alt+k"` (combo string built correctly)
- ✅ `Found handler for combo 'alt+k'` (handler registered)
- ✅ Inspector panel moves up

### Step 4: What If It Doesn't Work?

**Scenario A: No debug output when pressing Alt+K**
- **Cause**: Your terminal isn't sending Alt+K events
- **Solution**: Try a different terminal (Windows Terminal vs CMD vs PowerShell)

**Scenario B: Debug shows Modifiers=0 (no Alt)**
- **Cause**: Alt modifier not detected by Windows console
- **Solution**: Use Ctrl+D instead, or arrow keys with Alt

**Scenario C: Combo string is empty or wrong**
- **Cause**: buildComboString() not working correctly
- **Check**: Verify keyEv.Key.Rune contains 'k'

**Scenario D: "No handler found for combo"**
- **Cause**: Handler not registered
- **Check**: Verify `SetupInspectorShortcut()` was called

## Workarounds

If Alt+K doesn't work in your terminal, use these alternatives:

### Arrow Keys (usually work better):
- **Alt+↑** (Alt+Up) - Move up
- **Alt+↓** (Alt+Down) - Move down
- **Alt+←** (Alt+Left) - Move left
- **Alt+→** (Alt+Right) - Move right

### F12 for Toggle:
- **F12** - Show/hide Inspector

### Number Keys for Tabs:
- **1-5** - Switch to different tabs

## Platform-Specific Notes

### Windows Console
- Alt+letter combinations work in most cases
- Some terminals may filter or modify these events
- Windows Terminal recommended over legacy CMD

### Alternative Terminals
If Alt+K doesn't work:
- Try Windows Terminal (modern)
- Try PowerShell Core
- Try Git Bash
- Try VS Code integrated terminal

## Code Changes Summary

**Files Modified**:
1. `framework/event/pump.go` - Added Shift field setting
2. `framework/app.go` - Fixed character key routing to Inspector
3. `framework/event/handler.go` - Added debug output
4. `runtime/platform/input_windows.go` - Added debug output

**Test Status**: ✅ All tests pass
