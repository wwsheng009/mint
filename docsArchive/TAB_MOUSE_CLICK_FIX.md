# Tab/Button Mouse Click Not Working After Msg/Cmd Migration

> **Date**: 2026-02-10
> **Status**: 🔍 Root cause identified

---

## Problem

After Msg/Cmd migration (commit 5afab0a8):
- ✅ Tab switching works with keyboard (arrows, Enter)
- ✅ Modal buttons work with keyboard (Tab, Enter)
- ❌ Tab clicking with mouse DOES NOT work
- ❌ Modal buttons clicking with mouse DOES NOT work

---

## Root Cause

**Component Registry + HitMap Routing Issue**

The event flow is:
1. User clicks → Pump generates `MouseMsg` with `TargetID` (from HitMap)
2. `App.handleMsg()` looks up component by `TargetID` in `ComponentRegistry`
3. Calls `component.Update(MouseMsg)`

**The problem**: Components need:
1. ✅ Valid `ID` for registration
2. ✅ Implement `component.Updater` interface
3. ❌ **Be discoverable in layout tree traversal**

---

## Investigation

### Tab Component
```go
// components/navigation/tabs.go
func NewTabs() *TabsVNode {
    return &TabsVNode{
        ElementVNode: ui.NewElement("tabs"),  // ❌ No ID set!
        ...
    }
}
```

**Problem**: `NewElement("tabs")` creates an ElementVNode without an ID. Without an ID, `buildComponentRegistry()` skips it (line 865: `if nodeID != ""`).

### Modal Buttons
Modal uses `ui.Modal()` which wraps content in a `BorderedNode`. The buttons inside may not have proper IDs or may not be traversed correctly by the layout tree.

---

## Solution

### Fix 1: Ensure Tab Component Has ID

```go
// components/navigation/tabs.go
func NewTabs() *TabsVNode {
    node := &TabsVNode{
        ElementVNode: ui.NewElement("tabs"),
        ...
    }
    node.SetID("tabs-" + generateUniqueID())  // Set unique ID
    return node
}
```

### Fix 2: Ensure All Interactive Components Have IDs

Every component that implements `component.Updater` MUST have a unique ID:
- Button
- Input
- Checkbox
- Select
- TextArea
- TreeView
- Tabs
- Modal

### Fix 3: Debug Component Registry

Enable debug logging to see what's being registered:

```bash
TUI_DEBUG_UI=true go run main.go
```

Look for:
```
[APP] Registered component: <id>
[APP] Component registry built: <count> components
```

If your component is NOT listed, it's not being registered.

---

## Testing

After fix, verify:
1. Run with `TUI_DEBUG_UI=true`
2. Click on a tab/button
3. Should see:
   ```
   [APP] Direct routing: MouseMsg → <component-id>
   [APP] Component returned Cmd: ...
   ```
4. UI should update immediately

---

## Related Files

- `framework/app.go` - buildComponentRegistry() (line 848)
- `components/navigation/tabs.go` - NewTabs() (line 64)
- `components/button/button.go` - NewButton()
- `framework/component/registry.go` - ComponentRegistry

---

## Next Steps

1. ✅ Identify root cause (missing IDs)
2. ⏳ Add unique IDs to all components
3. ⏳ Test mouse click functionality
4. ⏳ Verify dirty flag triggers re-render
