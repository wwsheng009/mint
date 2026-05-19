# Sandbox Event Injection Analysis - Complete Event Flow

## Overview

This document traces the complete event flow from test injection through the framework to identify where Inspector event routing breaks down.

---

## Event Flow for Test Injection

### Step-by-Step Trace

```
1. TEST INJECTION
   └─> testApp.InjectSpecialKey(platform.KeyDown)
       File: ui/test.go:321

2. CREATE RAW INPUT
   └─> platform.RawInput{
           Type:    platform.InputKeyPress,
           Special: platform.KeyDown,
       }
       File: ui/test.go:322-326

3. INJECT INTO FRAMEWORK APP
   └─> fwApp.InjectEvent(raw)
       File: ui/test.go:326

4. FRAMEWORK INJECT TO PUMP
   └─> a.pump.Inject(raw)
       File: framework/app.go:1252

5. PUMP CONVERTS TO EVENT
   └─> p.convertToEvent(raw)
       └─> p.convertKeyEvent(raw)
           File: framework/event/pump.go:112-160

       Creates KeyEvent:
       └─> &KeyEvent{
               BaseEvent: NewBaseEvent(EventKeyPress),
               Special:   KeyDown,
               Key: Key{
                   Name: "down",  // ← From SpecialKey.String()
               },
           }
           File: framework/event/pump.go:129-160

6. PUMP SENDS TO EVENT CHANNEL
   └─> p.events <- ev
       File: framework/event/pump.go:102

7. FRAMEWORK MAIN LOOP RECEIVES EVENT
   └─> case ev := <-eventChan:
       File: framework/app.go:701

8. FRAMEWORK HANDLES EVENT
   └─> a.handleEvent(ev)
       File: framework/app.go:709
```

---

## Framework handleEvent() - Current Behavior

### Location: `framework/app.go:746-815`

```go
func (a *App) handleEvent(ev frameworkevent.Event) {
    // 1. Event filter
    if !a.eventFilter(ev) {
        return
    }

    // 2. Route to router subscribers
    if a.router != nil {
        a.router.Route(ev)
    }

    // 3. Handle resize
    if ev.Type() == frameworkevent.EventResize {
        // ... handle resize
        return
    }

    // 4. Handle keyboard events
    if ev.Type() == frameworkevent.EventKeyPress {
        // 4a. Check keyMap shortcuts (F12, Ctrl+D, Alt+H/J/K/L)
        if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
            if handler, found := a.keyMap.Lookup(keyEv); found {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                    return  // ← Shortcut handled, stop
                }
            }
        }

        // 4b. Send to root component (VNode tree)
        if a.root != nil {
            if handler, ok := a.root.(frameworkevent.Component); ok {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                }
            }
        }
        return  // ← Inspector never sees the event! ❌
    }

    // 5. Handle mouse events
    if ev.Type().IsMouse() {
        // ... send to root for hit testing
        return
    }
}
```

---

## The Problem: Missing Inspector Routing Step

### Current Event Flow (BROKEN):

```
KeyPress Event (Key="down")
    ↓
[Check keyMap shortcuts]
    ↓ (not found: "down" is not registered)
[Send to VNode tree]
    ↓
[Inspector NEVER sees it] ❌
```

### Correct Event Flow (NEEDED):

```
KeyPress Event (Key="down")
    ↓
[Check keyMap shortcuts]
    ↓ (not found)
[Check if Inspector visible?]
    ↓ (yes, Inspector is visible)
[Inspector.HandleKeyEvent("down", false, false)]
    ↓ (returns true - handled)
[STOP - don't send to VNode tree] ✅
```

---

## Key String Mapping

### From framework/event/keyboard.go - SpecialKey.String()

| SpecialKey    | String Returns | Inspector Expects | Match? |
|---------------|----------------|-------------------|--------|
| KeyUp         | "up"           | "up"              | ✅    |
| KeyDown       | "down"         | "down"            | ✅    |
| KeyLeft       | "left"         | -                 | ✅    |
| KeyRight      | "right"        | -                 | ✅    |
| KeyHome       | "home"         | "home"            | ✅    |
| KeyEnd        | "end"          | "end"             | ✅    |
| KeyPageUp     | "pageup"       | "pgup"            | ❌    |
| KeyPageDown   | "pagedown"     | "pgdn"            | ❌    |
| KeyEnter      | "enter"        | "enter"           | ✅    |
| KeyF12        | "f12"          | -                 | ✅    |

### Issue Found: PageUp/PageDown String Mismatch

**Framework**: `KeyPageUp.String()` → `"pageup"`
**Inspector**: `case "pgup":`

**Framework**: `KeyPageDown.String()` → `"pagedown"`
**Inspector**: `case "pgdn":`

This means PageUp/PageDown won't work even after fixing event routing!

---

## Inspector's HandleKeyEvent() Method

### Location: `internal/inspector/standalone_inspector.go:1073`

#### Signature:
```go
func (si *StandaloneInspector) HandleKeyEvent(key string, alt bool, ctrl bool) bool
```

#### Key String Expectations:

**Navigation Keys (Elements tab only):**
- `"up"` - Navigate up
- `"down"` - Navigate down
- `"pgup"` - Page up (❌ framework sends "pageup")
- `"pgdn"` - Page down (❌ framework sends "pagedown")
- `"home"` - Jump to top
- `"end"` - Jump to bottom
- `"e"` - Toggle expand/collapse
- `"enter"` - Select node

**Panel Movement (Alt modifier):**
- `"h"` or `"left"` - Move panel left
- `"l"` or `"right"` - Move panel right
- `"k"` or `"up"` - Move panel up
- `"j"` or `"down"` - Move panel down

**Tab Switching:**
- `"1"` - Elements tab
- `"2"` - Console tab
- `"3"` - Performance tab
- `"4"` - Diagnostics tab
- `"5"` - Network tab

---

## Two Fixes Required

### Fix #1: Add Inspector Event Routing in framework/app.go

**Location**: `framework/app.go:handleEvent()` method (around line 778)

**Insert after keyMap check (line 792)**:

```go
// Then route to Inspector if visible
if a.inspector != nil && a.isInspectorVisible() {
    if inspectorObj, ok := a.inspector.(interface {
        HandleKeyEvent(key string, alt, ctrl bool) bool
    }); ok {
        if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
            keyName := keyEv.Key.Name
            alt := keyEv.Key.Alt
            ctrl := keyEv.Key.Ctrl

            if inspectorObj.HandleKeyEvent(keyName, alt, ctrl) {
                a.dirty = true
                return  // Inspector handled it, don't send to VNode tree
            }
        }
    }
}
```

**This ensures:**
- ✅ Inspector receives ALL keyboard events when visible
- ✅ Inspector can choose to handle or ignore each event
- ✅ Works for both production (real platform input) and tests (injected events)
- ✅ VNode tree only gets events Inspector doesn't handle

### Fix #2: Fix PageUp/PageDown String Mismatch

**Two options:**

#### Option A: Change Inspector to match framework (RECOMMENDED)

**File**: `internal/inspector/standalone_inspector.go`

Change lines 1158-1163:
```go
// OLD:
case "pgup":
    platformKey = platform.KeyPageUp
case "pgdn":
    platformKey = platform.KeyPageDown

// NEW:
case "pageup", "pgup":  // Accept both for compatibility
    platformKey = platform.KeyPageUp
case "pagedown", "pgdn":  // Accept both for compatibility
    platformKey = platform.KeyPageDown
```

#### Option B: Change framework to match Inspector

**File**: `framework/event/keyboard.go`

Change lines 72-73:
```go
// OLD:
KeyPageUp:    "pageup",
KeyPageDown:  "pagedown",

// NEW:
KeyPageUp:    "pgup",
KeyPageDown:  "pgdn",
```

**Recommendation**: Use Option A because:
1. Inspector accepting both strings is more flexible
2. Doesn't break other code that might depend on "pageup"/"pagedown"
3. Aligns with common naming conventions

---

## isInspectorVisible() Helper Method

### Needs to be added to framework/app.go:

```go
// isInspectorVisible checks if the Inspector overlay is currently visible
func (a *App) isInspectorVisible() bool {
    if a.inspector == nil {
        return false
    }
    if inspector, ok := a.inspector.(interface{ IsVisible() bool }); ok {
        return inspector.IsVisible()
    }
    return false
}
```

---

## Complete Fixed handleEvent() Flow

```go
func (a *App) handleEvent(ev frameworkevent.Event) {
    // 1. Event filter
    if !a.eventFilter(ev) {
        return
    }

    // 2. Route to router subscribers
    a.router.Route(ev)

    // 3. Handle resize
    if ev.Type() == frameworkevent.EventResize {
        // ... handle resize
        return
    }

    // 4. Handle keyboard events
    if ev.Type() == frameworkevent.EventKeyPress {
        // 4a. Check keyMap shortcuts FIRST (F12, Ctrl+D)
        if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
            if handler, found := a.keyMap.Lookup(keyEv); found {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                    return  // Shortcut handled
                }
            }
        }

        // 4b. Route to Inspector if visible (NEW!)
        if a.inspector != nil && a.isInspectorVisible() {
            if inspectorObj, ok := a.inspector.(interface {
                HandleKeyEvent(key string, alt, ctrl bool) bool
            }); ok {
                if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
                    keyName := keyEv.Key.Name
                    alt := keyEv.Key.Alt
                    ctrl := keyEv.Key.Ctrl

                    if inspectorObj.HandleKeyEvent(keyName, alt, ctrl) {
                        a.dirty = true
                        return  // Inspector handled it, stop here
                    }
                }
            }
        }

        // 4c. Send to root component if Inspector didn't handle it
        if a.root != nil {
            if handler, ok := a.root.(frameworkevent.Component); ok {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                }
            }
        }
        return
    }

    // 5. Handle mouse events
    if ev.Type().IsMouse() {
        // ... send to root for hit testing
        return
    }
}
```

---

## Why This Fix Works for Both Production and Tests

### Production App (Real Platform Input):
```
Physical Key Press
    ↓
Platform Input Reader
    ↓
Pump.convertLoop()
    ↓
Pump.convertToEvent() → KeyEvent
    ↓
Pump.events channel
    ↓
Framework.handleEvent()
    ↓
[NEW] Inspector.HandleKeyEvent()
    ↓
TreeView navigation works! ✅
```

### Test App (MockSandbox Injection):
```
testApp.InjectSpecialKey()
    ↓
fwApp.InjectEvent()
    ↓
Pump.Inject() (bypasses platform input)
    ↓
Pump.convertToEvent() → KeyEvent
    ↓
Pump.events channel
    ↓
Framework.handleEvent()
    ↓
[NEW] Inspector.HandleKeyEvent()
    ↓
TreeView navigation works! ✅
```

**Both paths converge at Framework.handleEvent()**, so the fix works for both!

---

## Testing the Fix

### After implementing, this test should pass WITHOUT manual HandleKeyEvent calls:

```go
func TestTreeViewNavigationAutomatic(t *testing.T) {
    // Create Inspector
    insp := inspector.NewStandaloneInspector()
    insp.Enable()
    insp.ToggleVisibility()

    // Create testable app
    testApp, err := ui.RunTest(func() ui.VNode {
        return insp.RenderOverlay()
    })
    defer testApp.Close()

    // Inject keys - NO MANUAL HandleKeyEvent CALLS NEEDED!
    testApp.InjectSpecialKey(platform.KeyDown)
    time.Sleep(100 * time.Millisecond)

    // Verify focus moved
    focusIdx := insp.GetTreeViewComponent().GetFocusIndex()
    if focusIdx == 0 {
        t.Error("Focus should have moved down")
    }
}
```

### Current Test (WORKAROUND):
```go
testApp.InjectSpecialKey(platform.KeyDown)
insp.HandleKeyEvent("down", false, false)  // ← REQUIRED until fix
overlay = insp.RenderOverlay()
testApp.ForceRender()
```

### After Fix (CLEAN):
```go
testApp.InjectSpecialKey(platform.KeyDown)
time.Sleep(100 * time.Millisecond)
// That's it! Inspector automatically received the event
```

---

## Files to Modify

1. **`framework/app.go`**
   - Add `isInspectorVisible()` helper method
   - Modify `handleEvent()` to route keyboard events to Inspector

2. **`internal/inspector/standalone_inspector.go`**
   - Fix PageUp/PageDown key string handling (accept both "pageup"/"pgup")

3. **`examples/ui_demos/demo2_runtime_internals/main.go`** (OPTIONAL)
   - Add Inspector setup (currently doesn't create Inspector at all)

---

## Summary

**Root Cause**: Framework's `handleEvent()` doesn't route keyboard events to Inspector.

**Solution**: Add Inspector routing step after keyMap check but before VNode tree.

**Result**:
- ✅ Inspector receives all keyboard events when visible
- ✅ TreeView navigation works with arrow keys
- ✅ Works for both production apps and test injection
- ✅ No manual `HandleKeyEvent()` calls needed in tests
- ✅ VNode tree only gets events Inspector doesn't handle

**Two bugs to fix**:
1. Missing event routing in framework/app.go ❌
2. PageUp/PageDown string mismatch ❌
