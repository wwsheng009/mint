# Demo2 Inspector Not Working - Root Cause Analysis

## Problem Statement

The Inspector in demo2 does not respond to keyboard navigation events (arrow keys, PageUp/Down, etc.). User reported: "the demo2 application is not working yet"

## Investigation Summary

After examining demo1, demo2, and the framework code, I've identified **THREE separate issues**:

---

## Issue #1: Inspector Not Created in demo2/main.go ❌

### Location: `examples/ui_demos/demo2_runtime_internals/main.go`

### Current Code:
```go
func main() {
    _ = theme.SetTheme("nord")

    err := ui.Run(RuntimeDemo,
        ui.WithWidth(100),
        ui.WithHeight(35),
        ui.WithTitle("Mint TUI - Runtime Internals"),
    )
    // NO Inspector setup here!
}
```

### What's Missing:
```go
// This code does NOT exist in demo2/main.go:
func main() {
    _ = theme.SetTheme("nord")

    // MISSING: Create Inspector instance
    globalInspector := inspector.NewStandaloneInspector()
    globalInspector.Enable()

    // MISSING: Use framework app instead of ui.Run
    fwApp := framework.NewApp()
    fwApp.Resize(100, 35)

    // MISSING: Register Inspector with framework
    fwApp.SetInspector(globalInspector)
    fwApp.SetupInspectorShortcut()

    // MISSING: Set root and run
    declarativeRoot := render.NewDeclarativeNodeFromFunc(RuntimeDemo)
    declarativeRoot.SetFrameworkApp(fwApp)
    fwApp.SetRoot(declarativeRoot)
    fwApp.Run()
}
```

### Comparison with Working Example:

**✅ inspector_overlay/main.go (WORKS):**
```go
func main() {
    globalInspector = inspector.NewStandaloneInspector()
    globalInspector.Enable()

    fwApp := framework.NewApp()
    fwApp.SetInspector(globalInspector)          // ← CRITICAL
    fwApp.SetupInspectorShortcut()                // ← CRITICAL

    declarativeRoot := render.NewDeclarativeNodeFromFunc(RuntimeDemoWithInspectorOverlay)
    declarativeRoot.SetFrameworkApp(fwApp)
    fwApp.SetRoot(declarativeRoot)
    fwApp.Run()
}
```

**❌ demo1/main.go (NO Inspector):**
```go
func main() {
    _ = theme.SetTheme("nord")
    err := ui.Run(App, ...)  // Just ui.Run(), no Inspector
}
```

**❌ demo2/main.go (NO Inspector):**
```go
func main() {
    _ = theme.SetTheme("nord")
    err := ui.Run(RuntimeDemo, ...)  // Just ui.Run(), no Inspector
}
```

---

## Issue #2: ui.Run() Doesn't Auto-Create Inspector ❌

### Location: `ui/app.go`

### Current Behavior:
```go
func Run(app ComponentFunc, opts ...Option) error {
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)

    // NO Inspector creation here!
    // NO fwApp.SetInspector() call!

    declarativeNode := render.NewDeclarativeNodeFromFunc(app)
    fwApp.SetRoot(declarativeNode)
    return fwApp.Run()
}
```

### Why This is a Problem:
- Every app using `ui.Run()` must manually create framework app
- Every app must manually create Inspector
- Every app must manually call `SetInspector()` and `SetupInspectorShortcut()`
- This is error-prone and inconvenient

---

## Issue #3: Keyboard Events NOT Routed to Inspector ❌ **(CRITICAL)**

### Location: `framework/app.go` - `handleEvent()` method

### Current Event Flow:
```go
func (a *App) handleEvent(ev frameworkevent.Event) {
    // 1. Check event filter
    if !a.eventFilter(ev) {
        return
    }

    // 2. Route to registered subscribers
    a.router.Route(ev)

    // 3. Handle resize events
    if ev.Type() == frameworkevent.EventResize {
        // ...
        return
    }

    // 4. Handle keyboard events
    if ev.Type() == frameworkevent.EventKeyPress {
        // Check keyMap shortcuts (F12, Ctrl+D, Alt+H/J/K/L)
        if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
            if handler, found := a.keyMap.Lookup(keyEv); found {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                    return  // ← Shortcut handled, stop here
                }
            }
        }

        // Send to root component (VNode tree)
        if a.root != nil {
            if handler, ok := a.root.(frameworkevent.Component); ok {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                }
            }
        }
        return  // ← Event sent to VNode tree, Inspector never sees it!
    }

    // 5. Handle mouse events
    if ev.Type().IsMouse() {
        if a.root != nil {
            if handler, ok := a.root.(frameworkevent.Component); ok {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                }
            }
        }
        return
    }
}
```

### What's Missing:

**NO CODE ROUTES REGULAR KEYBOARD EVENTS TO INSPECTOR!**

The event flow should be:
```
KeyPress Event
    ↓
[Check keyMap shortcuts]
    ↓ (if not a shortcut)
[Check if Inspector is visible]
    ↓ (if visible)
[Inspector.HandleKeyEvent()]
    ↓ (if not handled by Inspector)
[Send to VNode tree]
```

But the actual flow is:
```
KeyPress Event
    ↓
[Check keyMap shortcuts] ← Only F12, Ctrl+D, Alt+H/J/K/L
    ↓ (if not a shortcut)
[Send to VNode tree] ← Arrow keys, PageUp/Down go here!
    ↓
[Inspector never sees them] ❌
```

### Why Arrow Keys Don't Work:

When Inspector is visible and user presses **Down Arrow**:

1. ✅ Platform receives key input
2. ✅ Event pump creates KeyEvent
3. ✅ handleEvent() receives KeyEvent
4. ❌ keyMap.Lookup() returns false (Down Arrow not registered)
5. ❌ Event goes to root VNode tree (not Inspector!)
6. ❌ Inspector's HandleKeyEvent("down", ...) is NEVER called
7. ❌ TreeView's HandleKey(KeyDown) is NEVER called
8. ❌ Focus doesn't move

### What DOES Work:

Only registered shortcuts work via keyMap:
- ✅ F12 - Toggles Inspector visibility
- ✅ Ctrl+D - Toggles Inspector visibility
- ✅ Alt+H/J/K/L - Moves Inspector panel
- ✅ 1-9 keys - Switch Inspector tabs (via switchInspectorTab)

But regular navigation keys don't work:
- ❌ Arrow keys (↑↓←→)
- ❌ PageUp/PageDown
- ❌ Home/End
- ❌ E key (expand/collapse)
- ❌ Enter (select)

---

## Root Cause Summary

### Architecture Issue:

The Inspector is designed as a **global overlay** that exists **outside** the VNode component tree:

```
App Root (VNode Tree)
├── VStack
│   ├── Header
│   ├── Content
│   └── Footer
└── Inspector Overlay ← Separate, not in tree!
    ├── TreeView (needs arrow keys)
    ├── Props Panel
    └── State Panel
```

### Event Routing Problem:

The platform input system routes events to **focused VNodes** in the component tree. The Inspector overlay is **NOT** in the tree, so it never receives keyboard events.

### Working Example:

In `inspector_overlay/main.go`, Inspector works because:
1. Inspector is manually created and registered
2. F12/Ctrl+D shortcuts are registered via SetupInspectorShortcut()
3. Tab keys (1-9) are registered via OnKeyCombo() calls
4. **BUT arrow keys still don't work!** (same architecture issue)

---

## Solution

The fix requires **automatic event routing** to the Inspector when it's visible. This is documented in `INSPECTOR_EVENT_INTEGRATION.md`.

### Implementation:

Add an event dispatcher that checks if Inspector is visible and wants the event:

```go
func (a *App) handleEvent(ev frameworkevent.Event) {
    // ... existing code ...

    if ev.Type() == frameworkevent.EventKeyPress {
        // 1. Check keyMap shortcuts first
        if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
            if handler, found := a.keyMap.Lookup(keyEv); found {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                    return
                }
            }
        }

        // 2. NEW: Route to Inspector if visible
        if a.inspector != nil && a.isInspectorVisible() {
            if inspectorObj, ok := a.inspector.(interface {
                HandleKeyEvent(key string, alt, ctrl bool) bool
            }); ok {
                keyName := a.getKeyName(ev)
                if inspectorObj.HandleKeyEvent(keyName, false, false) {
                    a.dirty = true
                    return  // Inspector handled it, don't send to VNode tree
                }
            }
        }

        // 3. Send to root component if Inspector didn't handle it
        if a.root != nil {
            if handler, ok := a.root.(frameworkevent.Component); ok {
                if handler.HandleEvent(ev) {
                    a.dirty = true
                }
            }
        }
        return
    }

    // ... rest of event handling ...
}
```

---

## Next Steps

### Option 1: Fix demo2 Only (Quick Fix)

Add Inspector setup to demo2/main.go following inspector_overlay pattern.

### Option 2: Fix Framework (Proper Solution)

1. Add automatic Inspector event routing in `framework/app.go`
2. Optionally auto-create Inspector in `ui.Run()` for convenience

### Option 3: Both (Recommended)

1. Fix framework to route events to Inspector properly
2. Fix demo2 to create Inspector instance
3. Optionally make ui.Run() auto-create Inspector

---

## Files That Need Changes

### Critical (Must Fix):

1. **`framework/app.go`** - Add Inspector event routing in handleEvent()
2. **`examples/ui_demos/demo2_runtime_internals/main.go`** - Add Inspector setup

### Optional (Nice to Have):

3. **`ui/app.go`** - Auto-create Inspector in Run()
4. **`ui/test.go`** - Auto-create Inspector in RunTest()

---

## Verification Steps

After fixing:

1. Run demo2 and press F12 to toggle Inspector
2. Inspector should appear with TreeView
3. Press arrow keys (↑↓) - TreeView focus should move
4. Press PageUp/PageDown - TreeView should scroll
5. Press Home/End - TreeView should jump to top/bottom
6. Tab between Inspector panels with 1-9 keys
7. All navigation should work without manual HandleKeyEvent calls

---

## References

- **Inspector Event Integration Issue**: `INSPECTOR_EVENT_INTEGRATION.md`
- **TreeView Navigation**: `TREEVIEW_NAVIGATION_WORKING.md`
- **Working Example**: `examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go`
- **Framework App**: `framework/app.go`
- **Event System**: `framework/event/`
