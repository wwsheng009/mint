# Inspector Event Integration Issue & Solution

## Problem

The Inspector's `HandleKeyEvent()` method is **NOT automatically called** when keyboard events occur in the testable app (sandbox). This requires manual intervention in tests:

```go
// ❌ Current approach - Manual event routing required
testApp.InjectSpecialKey(platform.KeyDown)
insp.HandleKeyEvent("down", false, false)  // MUST call this manually!
overlay = insp.RenderOverlay()
testApp.ForceRender()
```

This is **wrong** - the Inspector should automatically receive keyboard events from the platform input system, just like buttons and other interactive components do.

## Root Cause

### 1. Inspector is Outside the VNode Tree

The Inspector is a **global overlay** that exists **outside** the normal VNode component tree:

```
App Root
├── VStack
│   ├── Header
│   ├── Content
│   └── Footer
└── Inspector Overlay (separate, not in tree!)
```

Since it's not in the tree, normal event propagation doesn't reach it.

### 2. Framework Doesn't Route Events to Inspector ❌ **PRIMARY BUG**

The framework's `handleEvent()` method in `framework/app.go` routes keyboard events to:
1. keyMap shortcuts (F12, Ctrl+D, Alt+H/J/K/L)
2. VNode tree

**But NOT to the Inspector**, so regular navigation keys never reach it!

```go
// Current broken flow in framework/app.go:handleEvent()
if ev.Type() == frameworkevent.EventKeyPress {
    // Check keyMap shortcuts
    if handler, found := a.keyMap.Lookup(keyEv); found {
        if handler.HandleEvent(ev) {
            return  // Shortcut handled
        }
    }

    // Send to VNode tree ← Inspector never sees events!
    if a.root != nil {
        a.root.HandleEvent(ev)
    }
    return
}
```

**Missing**: Inspector routing step between keyMap check and VNode tree.

### 3. PageUp/PageDown String Mismatch ❌ **SECONDARY BUG**

- Framework `KeyPageUp.String()` returns `"pageup"`
- Inspector expects `"pgup"`

- Framework `KeyPageDown.String()` returns `"pagedown"`
- Inspector expects `"pgdn"`

This means even after fixing event routing, PageUp/PageDown won't work!

## Solution: Add Inspector Event Routing in Framework

### Architecture

Add Inspector routing in the framework's event handler:

```
Platform Input
    ↓
[Event Dispatch - framework/app.go:handleEvent()]
    ↓
    ├──→ 1. Check keyMap shortcuts (F12, Ctrl+D)
    │       └─ If handled: stop
    │
    ├──→ 2. Route to Inspector if visible ← **NEW!**
    │       ├─ Arrow keys for navigation
    │       ├─ PageUp/PageDown/Home/End
    │       └─ Number keys for tab switching
    │       └─ If handled: stop
    │
    └──→ 3. Send to App VNode Tree (gets remaining events)
        └── [Normal component event handling]
```

### Implementation Plan

#### Step 1: Add Inspector Routing in handleEvent()

**File**: `framework/app.go`

**Location**: In `handleEvent()` method, after keyMap check (around line 792)

**Add this code**:

```go
// Route to Inspector if visible (NEW!)
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

**Add helper method**:

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

#### Step 2: Fix PageUp/PageDown String Mismatch

**File**: `internal/inspector/standalone_inspector.go`

**Location**: In `HandleKeyEvent()` method (around line 1158)

**Change from**:
```go
case "pgup":
    platformKey = platform.KeyPageUp
    handled = si.treeViewComponent.HandleKey(platformKey, r)
case "pgdn":
    platformKey = platform.KeyPageDown
    handled = si.treeViewComponent.HandleKey(platformKey, r)
```

**Change to**:
```go
case "pageup", "pgup":  // Accept both for compatibility
    platformKey = platform.KeyPageUp
    handled = si.treeViewComponent.HandleKey(platformKey, r)
case "pagedown", "pgdn":  // Accept both for compatibility
    platformKey = platform.KeyPageDown
    handled = si.treeViewComponent.HandleKey(platformKey, r)
```

Also update the fallback scrolling code (around line 1200):
```go
case "pageup", "pgup":  // Accept both
    // Scroll up by one page
    si.treeScrollOffset -= treeViewHeight
    // ... rest of code
case "pagedown", "pgdn":  // Accept both
    // Scroll down by one page
    si.treeScrollOffset += treeViewHeight
    // ... rest of code
```

## Benefits

### 1. **Automatic Event Routing** ✅
- No manual `HandleKeyEvent` calls needed
- Inspector receives events through framework event routing
- Works for BOTH production and tests automatically

### 2. **Proper Event Priority** ✅
- keyMap shortcuts (F12, Ctrl+D) checked first
- Inspector gets second chance at events when visible
- VNode tree only gets events Inspector doesn't handle
- No conflicts between Inspector and app components

### 3. **Simpler Tests** ✅
```go
// ✅ After fix - Just inject keys, no manual routing
testApp.InjectSpecialKey(platform.KeyDown)
time.Sleep(100 * time.Millisecond)
render := testApp.GetRenderString()
// Inspector automatically received the event!
```

### 4. **Works for Both Input Sources** ✅
- Real platform input in production apps
- Injected events from test sandbox
- Both converge at `handleEvent()`, same behavior

### 5. **PageUp/PageDown Fixed** ✅
- Inspector now accepts both "pageup"/"pgup" and "pagedown"/"pgdn"
- Compatible with framework's SpecialKey.String() output
- Page navigation works correctly

## Current Workaround

Until this is implemented, tests **must** manually call HandleKeyEvent:

```go
// ❌ Workaround for now
testApp.InjectSpecialKey(platform.KeyDown)
insp.HandleKeyEvent("down", false, false)  // Required until fixed
overlay = insp.RenderOverlay()
testApp.ForceRender()
```

## Implementation Checklist

### ✅ Completed
1. ✅ Analyzed complete event flow from test injection
2. ✅ Identified two bugs: missing routing + string mismatch
3. ✅ Designed simple fix for framework/app.go
4. ✅ Documented PageUp/PageDown compatibility fix

### 🔄 In Progress
1. **Add Inspector routing to framework/app.go** - PRIMARY FIX
   - Add `isInspectorVisible()` helper method
   - Update `handleEvent()` to route keyboard events to Inspector

2. **Fix PageUp/PageDown in inspector** - SECONDARY FIX
   - Accept both "pageup"/"pgup" strings
   - Accept both "pagedown"/"pgdn" strings

3. **Update tests** - Remove manual HandleKeyEvent calls
   - Update `treeview_navigation_test.go`
   - Verify automatic event routing works

4. **Fix demo2 main.go** - Add Inspector setup (optional)
   - Create Inspector instance
   - Call SetInspector() and SetupInspectorShortcut()

### 📋 Future Enhancements
5. **Add mouse event routing** for click-outside-to-close
6. **Add focus management** for modal overlays
7. **Consider auto-creating Inspector in ui.Run()** for convenience

## References

- **Runtime Event System**: `runtime/event/`
- **Platform Input**: `runtime/platform/input.go`
- **Framework App**: `framework/app.go`
- **Layer System**: `runtime/layer/` - Already supports layered rendering
- **Inspector**: `internal/inspector/standalone_inspector.go`

## Conclusion

The Inspector should receive keyboard events automatically from the framework's event routing system. The current manual routing is a **workaround** for a missing integration step.

This fix makes the Inspector:
- ✅ More consistent with the component model
- ✅ Easier to use (no manual event routing in tests)
- ✅ More reliable (automatic event handling)
- ✅ Simpler to test (just inject keys)

**The fix is simple**: Add one routing step in `framework/app.go:handleEvent()` to check if Inspector is visible and wants the event, BEFORE sending to the VNode tree.

This works for both production apps and tests because both input paths converge at the same `handleEvent()` method.
