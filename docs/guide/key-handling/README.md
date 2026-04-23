# Key Handling Guide

This directory contains comprehensive documentation about keyboard input handling in the Mint TUI framework.

## Overview

The Mint framework has a sophisticated keyboard input handling system that:
- ✅ Correctly handles Ctrl+letter combinations (Ctrl+A through Ctrl+Z)
- ✅ Properly detects all modifier keys (Ctrl, Alt, Shift)
- ✅ Preserves case for Shift combinations (Ctrl+D vs Ctrl+Shift+D)
- ✅ Works with special keys (F12, arrows, PageUp/Down, etc.)
- ✅ Provides real-time key debugging in Inspector

## Documentation Files

### Complete Guide
- [Key Handling Complete Guide](KEY_HANDLING_COMPLETE_GUIDE.md) ⭐ **START HERE**
  - Complete explanation of the key handling system
  - Data flow from keypress to handler
  - Windows ControlKeyState flags
  - Control character handling
  - Key types and how they're processed
  - Debugging key issues
  - Common issues and solutions

### Alt+K Movement Keys
- [Alt+K Debug Guide](ALT_K_DEBUG_GUIDE.md) - Debugging Alt+K movement key issues
  - Why Alt+K might not work
  - How to diagnose problems
  - Platform-specific notes
  - Workarounds and alternatives

### Ctrl Key Fixes
- [Ctrl Key Fix Summary](/docsArchive/guide/key-handling/CTRL_KEY_FIX_SUMMARY.md) - Summary of Ctrl detection fixes (archived)
  - Root cause analysis
  - What was fixed
  - Test results

- [Ctrl+Shift Case Preservation](CTRL_SHIFT_CASE_PRESERVATION.md) - Case preservation for Shift
  - Implementation details
  - Before vs After comparison
  - Why it matters
  - Backward compatibility

### Key Debug Feature
- [Key Debug Feature Summary](/docsArchive/guide/key-handling/KEY_DEBUG_FEATURE_SUMMARY.md) - Real-time key debugging (archived)
  - What was implemented
  - How to use it
  - Display format
  - Troubleshooting guide

- [Key Debug Usage](KEY_DEBUG_USAGE.md) - How to use the debug feature
  - Step-by-step usage
  - What you'll see
  - Example sessions

## Quick Reference

### Modifier Keys

| Modifier | Windows Flags | Detected By |
|----------|---------------|-------------|
| **Shift** | 0x0010 | `ControlKeyState & 0x0010` |
| **Ctrl** | 0x0004 (right), 0x0008 (left) | `ControlKeyState & 0x0004 != 0 \|\| ControlKeyState & 0x0008 != 0` |
| **Alt** | 0x0001 (right), 0x0002 (left) | `ControlKeyState & 0x0001 != 0 \|\| ControlKeyState & 0x0002 != 0` |

### Control Character Handling

| Key Press | UChar | VirtualKeyCode | Result Key | Result Combo |
|-----------|-------|----------------|------------|--------------|
| **Ctrl+d** | 4 | 0x44 | 'd' (lowercase) | `"ctrl+d"` |
| **Ctrl+Shift+D** | 4 | 0x44 | 'D' (uppercase) | `"ctrl+shift+d"` |
| **Ctrl+K** | 11 | 0x4B | 'k' (lowercase) | `"ctrl+k"` |
| **Ctrl+Shift+K** | 11 | 0x4B | 'K' (uppercase) | `"ctrl+shift+k"` |

### Key Registration Rules

✅ **Always use lowercase for registrations**:
```go
a.OnKeyCombo("ctrl+d", func() { ... })   // ✅ Correct
a.OnKeyCombo("alt+k", func() { ... })   // ✅ Correct
a.OnKeyCombo("f12", func() { ... })     // ✅ Correct
```

❌ **Don't use uppercase**:
```go
a.OnKeyCombo("ctrl+D", func() { ... })   // ❌ Won't match Ctrl+d!
a.OnKeyCombo("ALT+K", func() { ... })   // ❌ Won't match Alt+k!
```

## Debugging Key Issues

### Enable Debug Output

```bash
# Platform input debug (shows what Windows sends)
set TUI_DEBUG_INPUT=true

# Framework debug (shows key routing)
set TUI_DEBUG_UI=true

# Inspector debug (shows key handling)
set TUI_DEBUG_INSPECTOR=true

go run main.go
```

### Common Issues

1. **Ctrl key not detected**
   - Check if `parseKeyEvent()` converts UChar 1-26 correctly
   - Verify ControlKeyState flags (0x0004, 0x0008)

2. **Shift key detected as Ctrl**
   - Fixed! Was checking wrong flag (0x0008 is LEFT_CTRL, not Shift)
   - Shift is 0x0010

3. **Key combo not found**
   - Make sure registration is lowercase: `"ctrl+d"` not `"ctrl+D"`
   - Check if case matches: Ctrl+d → 'd', Ctrl+Shift+D → 'D'

4. **Alt+letter not working**
   - Some terminals don't handle Alt+letter well
   - Try Alt+arrows instead: Alt+↑, Alt+↓, Alt+←, Alt+→

## Testing

Run the key handling tests:
```bash
# Unit tests for control character conversion
cd runtime/platform
go test -v -run TestCtrlCharacterConversion

# Integration tests for all modifiers
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go test -v -run TestCtrlModifierDetection
```

## See Also

- [Windows Platform Documentation](/runtime/platform/) - Platform input layer
- [Framework Event System](/framework/event/) - Event handling
- [Inspector Documentation](/docs/inspector/) - Inspector key handling
