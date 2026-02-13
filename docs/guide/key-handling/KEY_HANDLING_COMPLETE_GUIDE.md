# Key Handling - Complete Guide

## Overview

This document explains how keyboard input is handled in the Mint TUI framework, from the physical key press to the event handler execution.

## Data Flow

```
┌─────────────────┐
│ User presses    │
│ Ctrl+D          │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Windows Console (platform layer)       │
│ - Generates INPUT_RECORD               │
│ - Sets UChar=4 (control character)     │
│ - Sets VirtualKeyCode=0x44 (VK_D)      │
│ - Sets ControlKeyState=0x0008 (Ctrl)   │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ parseKeyEvent() (input_windows.go)     │
│ 1. Detects Shift (0x0010)              │
│ 2. Detects Ctrl (0x0004 | 0x0008)      │
│ 3. Detects Alt (0x0001 | 0x0002)      │
│ 4. Converts Ctrl+letter to letter      │
│ 5. Preserves case for Shift            │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Event Pump (framework/event/pump.go)   │
│ - Creates KeyEvent                     │
│ - Sets Key.Name, Key.Rune, Key.Alt    │
│ - Sets Key.Ctrl, Key.Shift             │
│ - Sets Modifiers bitmask               │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Framework App (framework/app.go)       │
│ 1. Check keyMap shortcuts (F12, etc.)  │
│ 2. Route to Inspector (if visible)     │
│ 3. Send to VNode tree (fallback)       │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Handler (Inspector, Component, etc.)   │
│ - Executes action                      │
│ - Returns true if handled              │
└─────────────────────────────────────────┘
```

## ControlKeyState Flags (Windows)

| Flag Value | Meaning                | Description |
|------------|------------------------|-------------|
| 0x0001     | RIGHT_ALT_PRESSED      | Right Alt key held |
| 0x0002     | LEFT_ALT_PRESSED       | Left Alt key held |
| 0x0004     | RIGHT_CTRL_PRESSED     | Right Ctrl key held |
| 0x0008     | LEFT_CTRL_PRESSED      | Left Ctrl key held |
| 0x0010     | SHIFT_PRESSED          | Shift key held |
| 0x0018     | SHIFT \| LEFT_CTRL     | Shift+Ctrl together |

**Combinations**:
- Ctrl alone: `0x0004` or `0x0008`
- Ctrl+Shift: `0x0018` (0x0010 | 0x0008)
- Alt alone: `0x0001` or `0x0002`
- Alt+Ctrl: `0x000A` (0x0002 | 0x0008)
- Alt+Shift: `0x0012` (0x0010 | 0x0002)

## Control Character Handling

### Windows Console Behavior

When you press Ctrl+Letter, Windows console sends a **control character** instead of the letter:

| Key Press | UChar (Decimal) | UChar (Hex) | Control Character Name |
|-----------|-----------------|-------------|------------------------|
| Ctrl+A    | 1               | 0x01        | SOH (Start of Heading) |
| Ctrl+B    | 2               | 0x02        | STX (Start of Text) |
| Ctrl+C    | 3               | 0x03        | ETX (End of Text) |
| Ctrl+D    | 4               | 0x04        | EOT (End of Transmission) |
| Ctrl+E    | 5               | 0x05        | ENQ (Enquiry) |
| Ctrl+F    | 6               | 0x06        | ACK (Acknowledge) |
| Ctrl+G    | 7               | 0x07        | BEL (Bell) |
| Ctrl+H    | 8               | 0x08        | BS (Backspace) |
| Ctrl+I    | 9               | 0x09        | HT (Horizontal Tab) |
| Ctrl+J    | 10              | 0x0A        | LF (Line Feed) |
| Ctrl+K    | 11              | 0x0B        | VT (Vertical Tab) |
| Ctrl+L    | 12              | 0x0C        | FF (Form Feed) |
| Ctrl+M    | 13              | 0x0D        | CR (Carriage Return) |
| Ctrl+N    | 14              | 0x0E        | SO (Shift Out) |
| Ctrl+O    | 15              | 0x0F        | SI (Shift In) |
| Ctrl+P    | 16              | 0x10        | DLE (Data Link Escape) |
| Ctrl+Q    | 17              | 0x11        | DC1 (Device Control 1) |
| Ctrl+R    | 18              | 0x12        | DC2 (Device Control 2) |
| Ctrl+S    | 19              | 0x13        | DC3 (Device Control 3) |
| Ctrl+T    | 20              | 0x14        | DC4 (Device Control 4) |
| Ctrl+U    | 21              | 0x15        | NAK (Negative Ack) |
| Ctrl+V    | 22              | 0x16        | SYN (Synchronous Idle) |
| Ctrl+W    | 23              | 0x17        | ETB (End of Trans Block) |
| Ctrl+X    | 24              | 0x18        | CAN (Cancel) |
| Ctrl+Y    | 25              | 0x19        | EM (End of Medium) |
| Ctrl+Z    | 26              | 0x1A        | SUB (Substitute) |

### Our Conversion Logic

```go
if uChar >= 1 && uChar <= 26 && virtualKeyCode >= 0x41 && virtualKeyCode <= 0x5A {
    // This is Ctrl+letter (A-Z)
    if controlKeyState&0x0010 != 0 {
        // Shift is pressed - use uppercase
        input.Key = rune(virtualKeyCode)  // 'A'-'Z'
    } else {
        // No shift - use lowercase
        input.Key = rune(virtualKeyCode + 32)  // 'a'-'z'
    }
    input.Modifiers |= ModCtrl
    input.Special = KeyUnknown
}
```

## Key Types and How They're Handled

### 1. Plain Letters (A-Z, a-z)

**Example**: Press 'd' or 'D'

**Windows sends**:
- UChar = 100 ('d') or 68 ('D')
- VirtualKeyCode = 0x44 (VK_D)
- ControlKeyState = 0 (no modifiers)

**Our handling**:
```go
input.Key = rune(UChar)  // Preserves original case
// No modifiers set
```

**Result**:
- 'd' → Key='d'
- 'D' → Key='D'

### 2. Ctrl+Letter (Lowercase)

**Example**: Press Ctrl+d

**Windows sends**:
- UChar = 4 (control character)
- VirtualKeyCode = 0x44 (VK_D)
- ControlKeyState = 0x0008 (Left Ctrl)

**Our handling**:
```go
// Detects control character (UChar 1-26)
input.Key = 'd'  // Lowercase (0x44 + 32 = 0x64)
input.Modifiers = ModCtrl
```

**Result**:
- Combo = "ctrl+d"
- Inspector shows: `🔍 Last key: 'd' (Ctrl+)`

### 3. Ctrl+Shift+Letter (Uppercase)

**Example**: Press Ctrl+Shift+D

**Windows sends**:
- UChar = 4 (control character)
- VirtualKeyCode = 0x44 (VK_D)
- ControlKeyState = 0x0018 (Shift | Left Ctrl)

**Our handling**:
```go
// Detects control character (UChar 1-26)
// Detects Shift flag (0x0010)
input.Key = 'D'  // Uppercase (0x44)
input.Modifiers = ModCtrl | ModShift
```

**Result**:
- Combo = "ctrl+shift+D"
- Inspector shows: `🔍 Last key: 'D' (Ctrl+Shift+)`

### 4. Alt+Letter

**Example**: Press Alt+k

**Windows sends**:
- UChar = 'k'
- VirtualKeyCode = 0x4B (VK_K)
- ControlKeyState = 0x0002 (Left Alt)

**Our handling**:
```go
// Not a control character (UChar > 26)
input.Key = 'k'  // Original character
input.Modifiers = ModAlt
```

**Result**:
- Combo = "alt+k"
- Inspector shows: `🔍 Last key: 'k' (Alt+)`

### 5. Special Keys (F12, Enter, Escape, Arrows)

**Example**: Press F12

**Windows sends**:
- VirtualKeyCode = 0x7B (VK_F12)
- ControlKeyState = 0

**Our handling**:
```go
input.Special = KeyF12
input.Key.Name = "f12"
```

**Result**:
- Combo = "f12"
- Inspector shows: `🔍 Last key: 'f12' (无)`

### 6. Alt+Special Keys

**Example**: Press Alt+Up

**Windows sends**:
- VirtualKeyCode = 0x26 (VK_UP)
- ControlKeyState = 0x0002 (Left Alt)

**Our handling**:
```go
input.Special = KeyUp
input.Key.Name = "up"
input.Modifiers = ModAlt
```

**Result**:
- Combo = "alt+up"
- Inspector shows: `🔍 Last key: 'up' (Alt+)`

## KeyMap Combo String Building

The `buildComboString()` method constructs combo strings like "ctrl+d" or "alt+shift+k":

```go
func (k *KeyMap) buildComboString(key Key, modifiers Modifier) string {
    if modifiers == 0 {
        return ""  // No modifiers
    }

    var combo string

    // Add modifier prefixes
    if modifiers&ModAlt != 0 {
        combo += "alt+"
    }
    if modifiers&ModCtrl != 0 {
        combo += "ctrl+"
    }
    if modifiers&ModShift != 0 {
        combo += "shift+"
    }

    // Add key name or rune
    if key.Rune > 0 {
        combo += string(key.Rune)  // Character key
    } else if key.Name != "" {
        combo += key.Name  // Special key
    }

    return combo
}
```

**Examples**:
| Input | Modifiers | Combo String |
|-------|-----------|--------------|
| Key='d', Rune='d', Ctrl=true | Ctrl | "ctrl+d" |
| Key='D', Rune='D', Ctrl\|Shift=true | Ctrl\|Shift | "ctrl+shift+D" |
| Key='k', Rune='k', Alt=true | Alt | "alt+k" |
| Key.Name='up', Alt=true | Alt | "alt+up" |
| Key.Name='f12' | none | "f12" |

## Registering Key Combos

### Using the Framework

```go
// In framework/app.go
a.OnKeyCombo("ctrl+d", func() {
    a.toggleInspector()
})

a.OnKeyCombo("f12", func() {
    a.toggleInspector()
})

a.OnKeyCombo("alt+k", func() {
    a.moveInspector(0, -1)  // Move up
})
```

**Important Rules**:
1. ✅ Use **lowercase** for Ctrl+letter: `"ctrl+d"`, `"ctrl+k"`
2. ✅ Use **lowercase** for Alt+letter: `"alt+k"`, `"alt+h"`
3. ✅ Use **lowercase** for special keys: `"f12"`, `"up"`, `"down"`
4. ❌ Don't use uppercase: `"ctrl+D"` won't match Ctrl+d

### Why Lowercase?

When you press Ctrl+d:
- Windows sends UChar=4, VirtualKeyCode=0x44
- We convert to: Key='d' (lowercase)
- buildComboString produces: `"ctrl+d"`

If you registered `"ctrl+D"` (uppercase):
- Your app receives: `"ctrl+d"` (lowercase)
- Looks up: `bindings["ctrl+d"]` ❌ Not found!
- **Result**: Key combo doesn't work!

## Inspector Key Debug Display

The Inspector shows the last pressed key in real-time:

```
╔═ INSPECTOR ═╗
F12:关闭 | Alt+H/J/K/L:移动 | Ctrl+D:按键调试
🔍 Last key: 'd' (Ctrl+)          ← This updates in real-time
[Elements] | Console | Performance | ...
```

**What it shows**:
- **Key name**: The character or special key name
- **Modifiers**: What modifiers were held
  - `(无)` - No modifiers (Chinese for "none")
  - `(Alt+)` - Alt was held
  - `(Ctrl+)` - Ctrl was held
  - `(Shift+)` - Shift was held
  - `(Ctrl+Shift+)` - Both Ctrl and Shift

## Common Key Combinations

### Inspector Movement
| Key | Combo | Action |
|-----|-------|--------|
| Alt+H | `"alt+h"` | Move panel left |
| Alt+J | `"alt+j"` | Move panel down |
| Alt+K | `"alt+k"` | Move panel up |
| Alt+L | `"alt+l"` | Move panel right |
| Alt+← | `"alt+left"` | Move panel left |
| Alt+→ | `"alt+right"` | Move panel right |
| Alt+↑ | `"alt+up"` | Move panel up |
| Alt+↓ | `"alt+down"` | Move panel down |

### Inspector Controls
| Key | Combo | Action |
|-----|-------|--------|
| F12 | `"f12"` | Toggle Inspector |
| Ctrl+D | `"ctrl+d"` | Toggle Inspector (alternate) |
| 1-5 | `"1"`, `"2"`, etc. | Switch tabs |
| Tab | `"tab"` | Cycle through tabs |

### Tree Navigation
| Key | Combo | Action |
|-----|-------|--------|
| ↑ | `"up"` | Navigate up |
| ↓ | `"down"` | Navigate down |
| PageUp | `"pageup"` | Scroll up |
| PageDown | `"pagedown"` | Scroll down |
| Home | `"home"` | Jump to top |
| End | `"end"` | Jump to bottom |
| E | `"e"` | Expand/collapse node |
| Enter | `"enter"` | Select node |

## Debugging Key Issues

### Enable Debug Output

```bash
# Platform input debug
set TUI_DEBUG_INPUT=true

# Framework debug
set TUI_DEBUG_UI=true

# Inspector debug
set TUI_DEBUG_INSPECTOR=true

# Run the app
go run main.go
```

### What You'll See

**Platform level**:
```
[WIN INPUT] VK=0x44 UChar=0x04 ControlKeyState=0x0008 Modifiers=Ctrl+
```

**Framework level**:
```
[KeyMap] Lookup: Rune=d Name="" Modifiers=2 Combo="ctrl+d"
[KeyMap] Found handler for combo 'ctrl+d'
```

**Inspector level**:
```
[Inspector] Key received: key='d' modifiers=Ctrl+ showKeyDebug=false
```

## Common Issues and Solutions

### Issue 1: Ctrl+D doesn't work

**Symptoms**: Pressing Ctrl+D does nothing

**Debug output shows**:
```
[KeyMap] Lookup: Rune= Name="" Modifiers=0 Combo=""
```

**Cause**: Control character not being converted

**Solution**: Make sure `parseKeyEvent()` converts UChar 1-26 to letters with Ctrl modifier

### Issue 2: Alt+K shows `(无)` instead of `(Alt+)`

**Symptoms**: Inspector shows "key='k' (无)" when pressing Alt+k

**Debug output shows**:
```
[WIN INPUT] Modifiers=none
```

**Cause**: Alt flag not detected in ControlKeyState

**Solution**: Check that both 0x0001 and 0x0002 are checked for Alt

### Issue 3: Shift not detected

**Symptoms**: Shift+key behaves like plain key

**Debug output shows**:
```
[WIN INPUT] ControlKeyState=0x0010 Modifiers=none
```

**Cause**: Shift flag checked against wrong mask (was checking 0x0008 which is LEFT_CTRL!)

**Solution**: Use 0x0010 for Shift detection

### Issue 4: Key combo registered but not found

**Symptoms**: Registered "ctrl+D" but lookup fails

**Debug output shows**:
```
[KeyMap] Lookup: Rune=D Modifiers=Ctrl Combo="ctrl+D"
[KeyMap] No handler found for combo 'ctrl+D'
```

**Cause**: Case mismatch - registered "ctrl+D" but receiving "ctrl+d"

**Solution**: Always register lowercase combos: `"ctrl+d"`, `"alt+k"`, etc.

## Complete Example: Ctrl+D Journey

### User Action
Press Ctrl+D while Inspector is visible

### Step 1: Windows Console
```
INPUT_RECORD generated:
- EventType: KEY_EVENT (1)
- KeyEvent.WChar: 4 (control character for Ctrl+D)
- KeyEvent.VirtualKeyCode: 0x44 (VK_D)
- KeyEvent.ControlKeyState: 0x0008 (LEFT_CTRL_PRESSED)
```

### Step 2: Platform Layer (input_windows.go)
```go
// Detects Ctrl flag
if ControlKeyState & 0x0008 != 0 {
    input.Modifiers |= ModCtrl  // ✅ Ctrl set
}

// Detects control character
if UChar >= 1 && UChar <= 26 && VK >= 0x41 && VK <= 0x5A {
    // Convert to lowercase (no Shift)
    input.Key = rune(0x44 + 32) = 'd'  // ✅ Lowercase
    input.Modifiers |= ModCtrl          // ✅ Ctrl set
    input.Special = KeyUnknown
}

// Result: input.Key='d', input.Modifiers=ModCtrl
```

### Step 3: Event Pump (pump.go)
```go
ev := &KeyEvent{
    BaseEvent: NewBaseEvent(EventKeyPress),
    Key: Key{
        Rune: 'd',           // ✅ Character
        Name: "",            // Empty for character keys
        Alt: false,
        Ctrl: true,          // ✅ Ctrl set
        Shift: false,
    },
    Modifiers: ModCtrl,      // ✅ Modifiers set
}
```

### Step 4: Framework App (app.go)
```go
// Try keyMap first
if handler, found := keyMap.Lookup(keyEv); found {
    // buildComboString produces "ctrl+d"
    // Finds registered handler for "ctrl+d"
    handler.HandleEvent(ev)
    return  // ✅ Handled!
}
```

### Step 5: Handler Execution
```go
func() {
    a.toggleInspector()  // ✅ Toggle Inspector visibility
}()
```

### Step 6: Inspector Update
```go
// HandleKeyEvent called
si.lastKey = "d"
si.lastCtrl = true
si.lastAlt = false
si.lastShift = false

// Render shows:
// 🔍 Last key: 'd' (Ctrl+)
```

## Summary

✅ **Ctrl combinations work** - Control characters converted to letters
✅ **Case preserved** - Ctrl+d → 'd', Ctrl+Shift+D → 'D'
✅ **All modifiers work** - Ctrl, Alt, Shift detected correctly
✅ **Debug output** - See exactly what keys are being received
✅ **Real-time display** - Inspector shows last key pressed

The key handling system is now complete and robust!
