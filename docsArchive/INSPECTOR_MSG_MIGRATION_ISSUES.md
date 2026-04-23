# Inspector Msg/Cmd Migration Issues

> **Date**: 2026-02-10
> **Status**: 🔍 Analysis in progress
> **Related Commit**: 5afab0a8 (feat: Complete Msg/Cmd architecture migration)

---

## Problem Description

After the Msg/Cmd migration (commit 5afab0a8), the following issues have been identified:

### Inspector Issues:
1. **First-time display issue**: Inspector doesn't show interface on first open (F12), requires pressing Enter to display
2. **Incomplete rendering**: Inspector rendering is incomplete
3. **Full screen issue**: Interface cannot display in full screen properly

### Modal Issues:
4. **Modal dialogs don't respond to mouse clicks**: Modal component has NO event handling (no `HandleEvent`, no `Update(Msg)`)

### Border Issues:
5. **Borders not rendering completely**: BorderedNode borders are incomplete

### Resize Events Issue:
6. **Screen resize events not working**: After Msg/Cmd migration, resize events were not being handled because Pump was creating generic BaseMsg without width/height information

**FIXED** (2026-02-10):
- Created `runtime/msg/resize_msg.go` with `ResizeMsg` type containing OldWidth, OldHeight, NewWidth, NewHeight
- Updated `Pump.convertToResizeMsg()` to use `runtimemsg.NewResizeMsg()` with actual dimensions
- Updated `MsgToEvent()` adapter to properly convert `ResizeMsg` to `ResizeEvent` with dimension data
- Resize events now work correctly - App.Resize() is called with proper width/height values

### Modal Internal Buttons Issue:
7. **Modal buttons don't respond to clicks**: Modal's content and footer weren't being included in HitMap, so buttons inside modal couldn't receive mouse events

**FIXED** (2026-02-10):
- Added `Children()` method to ModalVNode that returns content and footer
- This allows HitMap to see and register modal's child components
- Buttons inside modal now correctly receive mouse events via direct Msg routing
- Modal only handles ESC key and clicks outside its bounds

---

## Critical Finding: Modal Component Has No Event Handling

**File**: `components/overlay/modal.go`

The Modal component is purely visual - it has:
- ✅ `Measure()` method for layout
- ✅ `Paint()` method for rendering
- ❌ NO `HandleEvent()` method
- ❌ NO `Update(Msg)` method
- ❌ NO event handling of any kind

**Impact**: After Msg/Cmd migration, mouse clicks on modal dialogs are NOT routed to the modal because:
1. Modal doesn't implement `component.Updater` interface
2. Modal doesn't have a NodeID for HitMap routing
3. MouseMsg with TargetID will skip the modal
4. Fallback to Event system won't help (Modal has no HandleEvent)

**Result**: Modal buttons, close buttons, and interactive elements DON'T WORK.

---

## Root Cause Analysis

### Event Flow After Msg/Cmd Migration

The event flow has changed significantly:

#### Before (Event System):
```
User Input → Pump → KeyEvent/MouseEvent → KeyMap → Handler → Action
```

#### After (Msg/Cmd System):
```
User Input → Pump → KeyMsg/MouseMsg → App.handleMsg()
                                      ↓
                            Try direct routing (TargetID)
                                      ↓
                            Fallback: MsgToEvent() → KeyEvent
                                      ↓
                            KeyMap → Handler → Action
```

### Key Changes in Commit 5afab0a8:

1. **Pump** now outputs `Msg` instead of `Event`
2. **App.handleMsg()** added for direct routing
3. **Component Registry** for O(1) lookups by TargetID
4. **Msg→Event Adapter** (`MsgToEvent`) for backward compatibility

### Inspector Integration Points:

1. **Hook Registration**: ✅ Working
   ```
   [APP] ✅ Inspector hook registered via HookRegistrar interface
   ```

2. **Event Handling**: ⚠️ Needs Investigation
   - F12 key press → KeyMsg → (fallback) → KeyEvent → KeyMap → toggleInspector()
   - Inspector visibility state toggles correctly
   - `a.dirty = true` is set

3. **Rendering**: ⚠️ Needs Investigation
   - Hook should inject inspector overlay via Fragment
   - LayerInspector should be set on overlay
   - PipelineRenderer should render multi-layer output

---

## Potential Issues

### Issue 1: Inspector State Initialization

**Hypothesis**: The inspector's internal state may not be properly initialized when first toggled.

**Evidence**:
- `ToggleVisibility()` only sets `visible = true`
- No explicit state initialization happens on first open
- Inspector content generation depends on `AttachToApp()` being called

**Code**:
```go
// standalone_inspector.go:198
func (si *StandaloneInspector) ToggleVisibility() {
    si.mu.Lock()
    defer si.mu.Unlock()

    if !si.enabled {
        return
    }

    si.visible = !si.visible  // Just toggles, no init
    // ...
}
```

### Issue 2: Hook Timing

**Hypothesis**: The hook may be checking `IsVisible()` BEFORE the toggle completes.

**Flow**:
1. F12 pressed → KeyEvent → KeyMap handler
2. Handler calls `toggleInspector()`
3. `toggleInspector()` → `ToggleVisibility()` → sets `visible = true`
4. `a.dirty = true` set
5. **Next tick** → `render()` called
6. During render → hooks applied → `IsVisible()` checked

If there's any timing issue where the hook runs before the state is fully committed, the overlay won't be injected.

### Issue 3: Inspector Content Generation

**Hypothesis**: `RenderContent()` may return nil or incomplete content on first call.

**Code**:
```go
// standalone_inspector.go:325
func (si *StandaloneInspector) RenderContent() rtui.VNode {
    si.mu.RLock()
    defer si.mu.RUnlock()

    if !si.visible {
        return nil  // ❌ Returns nil if not visible
    }

    // Build overlay content (UI only, no Layer set)
    return si.buildOverlayContent()
}
```

If `AttachToApp()` hasn't been called yet, or if the tree view isn't initialized, `buildOverlayContent()` may fail or return incomplete data.

### Issue 4: Layer Rendering Order

**Hypothesis**: The LayerInspector may not be rendered in the correct order.

**Code**:
```go
// hook.go:61
inspectorContent.SetLayer(rtui.LayerInspector)
```

The layer should be rendered LAST (on top of everything). If the layer order is incorrect or if LayerInspector nodes aren't being detected, the overlay won't show.

---

## Investigation Steps

### Step 1: Enable Inspector Verbose Logging

```bash
TUI_DEBUG_INSPECTOR=true TUI_DEBUG_UI=true go run main.go
```

Look for:
- `[InspectorHook]` messages
- `ToggleVisibility` calls
- `IsVisible()` return values
- `RenderContent()` results

### Step 2: Trace F12 Key Press

Add logging to track:
1. F12 key reception
2. ToggleVisibility() execution
3. dirty flag setting
4. render() call timing
5. Hook execution
6. Layer detection and rendering

### Step 3: Verify Hook Registration

Confirm that:
1. Hook is registered before first render
2. Hook has correct inspector reference
3. `ApplyVNodeHooks()` is being called during render

### Step 4: Check Inspector State

Verify that when inspector is toggled:
1. `enabled = true`
2. `visible = true`
3. `appRoot` is set (from `AttachToApp()`)
4. `treeView` is initialized
5. `treeLines` are populated

---

## Proposed Fixes

### Fix 1: Ensure Inspector Initialization on First Toggle

**File**: `internal/inspector/standalone_inspector.go`

```go
func (si *StandaloneInspector) ToggleVisibility() {
    si.mu.Lock()
    defer si.mu.Unlock()

    if !si.enabled {
        return
    }

    si.visible = !si.visible

    // NEW: Initialize inspector state on first show
    if si.visible && si.appRoot == nil {
        if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
            fmt.Fprintf(os.Stderr, "[Inspector] First show - initializing\n")
        }
        // Trigger initialization via AttachToApp if app root available
        // This ensures treeView, treeLines are ready
    }

    // ...
}
```

### Fix 2: Force Immediate Render After Toggle

**File**: `framework/app.go`

```go
func (a *App) toggleInspector() {
    // ... existing code ...

    inspectorObj.ToggleVisibility()
    a.inspectorVisible = inspectorObj.IsVisible()

    // NEW: Force immediate render instead of waiting for tick
    a.dirty = true
    a.throttler.ForceRender()  // Force render on next tick

    // ...
}
```

### Fix 3: Add Inspector State Verification

**File**: `internal/inspector/hook.go`

```go
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        if !inspector.IsVisible() {
            return vnode
        }

        // NEW: Verify inspector is ready to render
        inspector.mu.RLock()
        hasRoot := inspector.appRoot != nil
        inspector.mu.RUnlock()

        if !hasRoot {
            if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
                fmt.Fprintf(os.Stderr, "[InspectorHook] Inspector not ready (no appRoot)\n")
            }
            return vnode
        }

        // ... rest of hook
    }
}
```

### Fix 4: Add Event Handling to Modal Component

**File**: `components/overlay/modal.go`

The modal component needs event handling to support:
- Click outside modal to close
- ESC key to close
- Mouse interactions with modal content

```go
package overlay

import (
    frameworkevent "github.com/wwsheng009/mint/framework/event"
    "github.com/wwsheng009/mint/framework/cmd"
    "github.com/wwsheng009/mint/framework/component"
    runtimemsg "github.com/wwsheng009/mint/runtime/msg"
    runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// Interface implementation assertions
var _ frameworkevent.Component = (*ModalVNode)(nil)
var _ component.Updater = (*ModalVNode)(nil)

// HandleEvent processes events (legacy, for backward compatibility)
func (m *ModalVNode) HandleEvent(e frameworkevent.Event) bool {
    if !m.isOpen {
        return false
    }

    // Handle ESC key to close
    if keyEvent, ok := e.(*frameworkevent.KeyEvent); ok {
        if keyEvent.Special == frameworkevent.KeyEscape {
            m.isOpen = false
            if m.onClose != nil {
                m.onClose()
            }
            return true
        }
    }

    // Handle click outside modal
    if mouseEvent, ok := e.(*frameworkevent.MouseEvent); ok {
        if mouseEvent.Type() == frameworkevent.EventMousePress {
            // Check if click is outside modal bounds
            if !m.containsPoint(mouseEvent.X, mouseEvent.Y) {
                m.isOpen = false
                if m.onClose != nil {
                    m.onClose()
                }
                return true
            }
        }
    }

    return false
}

// Update implements component.Updater interface for Msg/Cmd architecture
func (m *ModalVNode) Update(message runtimemsg.Msg) cmd.Cmd {
    if !m.isOpen {
        return nil
    }

    switch msg := message.(type) {
    case *runtimemsg.KeyMsg:
        return m.updateKey(msg)
    case *runtimemsg.MouseMsg:
        return m.updateMouse(msg)
    }

    return nil
}

// updateKey handles keyboard messages
func (m *ModalVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
    // ESC to close
    if keyMsg.Special == runtimeplatform.KeyEscape {
        m.isOpen = false
        if m.onClose != nil {
            m.onClose()
        }
        return nil
    }

    return nil
}

// updateMouse handles mouse messages
func (m *ModalVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
    if mouseMsg.Action == runtimemsg.MouseActionPress {
        // Check if click is outside modal bounds
        if !m.containsPoint(mouseMsg.X, mouseMsg.Y) {
            m.isOpen = false
            if m.onClose != nil {
                m.onClose()
            }
            return nil
        }
    }

    return nil
}

// containsPoint checks if a point is within the modal bounds
func (m *ModalVNode) containsPoint(x, y int) bool {
    // Modal bounds should be tracked during Paint
    // This is a simplified check
    return x >= m.bounds[0] && x < m.bounds[0]+m.bounds[2] &&
           y >= m.bounds[1] && y < m.bounds[1]+m.bounds[3]
}
```

### Fix 5: Fix BorderedNode Border Rendering

**File**: `runtime/ui/layout.go` or appropriate border rendering file

Investigate and fix incomplete border rendering. Common causes:
1. Width calculation doesn't account for wide characters
2. Border characters are being clipped
3. Layer rendering order is incorrect

---

## Summary of Issues

| Issue | Component | Root Cause | Fix Priority | Status |
|-------|-----------|------------|-------------|--------|
| Inspector not showing on first open | Inspector | State initialization timing | HIGH | ⏳ Pending |
| Inspector incomplete rendering | Inspector | Hook timing or content generation | HIGH | ⏳ Pending |
| Modal doesn't respond to clicks | Modal | No Update(Msg) implementation | CRITICAL | ✅ FIXED |
| Modal buttons don't work | Modal | Children() not exposing content/footer | CRITICAL | ✅ FIXED |
| Borders not rendering completely | BorderedNode | Rendering issue | MEDIUM | ⏳ Pending |
| Screen resize events not working | Pump/MsgAdapter | ResizeMsg missing width/height | CRITICAL | ✅ FIXED |

---

## Action Plan

1. ✅ **CRITICAL** - Fix Modal event handling
   - Modal is completely broken for user interaction
   - Add `Update(Msg)` implementation ✅
   - Add bounds tracking for click detection ✅
   - **Status**: FIXED - Modal now responds to clicks and ESC key

2. ✅ **CRITICAL** - Fix Modal internal button clicks
   - Modal buttons were not receiving mouse events
   - Add `Children()` method to expose content/footer ✅
   - HitMap now includes modal's child components ✅
   - **Status**: FIXED - Modal buttons now work correctly

3. ✅ **CRITICAL** - Fix resize event handling
   - Resize events were not working after Msg/Cmd migration
   - Create `ResizeMsg` type with width/height ✅
   - Update Pump to use new ResizeMsg ✅
   - Update MsgToEvent adapter ✅
   - **Status**: FIXED - Resize events now work correctly

4. **HIGH** - Fix Inspector first-open issue (Fixes 1-3)
   - Ensure proper state initialization
   - Force render after toggle
   - Verify hook timing
   - **Status**: PENDING

5. **MEDIUM** - Fix border rendering (Fix 5)
   - Investigate BorderedNode rendering
   - Fix width calculations
   - Ensure proper layer order
   - **Status**: PENDING

---

## Next Steps

1. ✅ **Confirm issues exist** - User confirmed: first open doesn't show, needs Enter, incomplete rendering
2. ✅ **Identify root cause** - Event flow changed, inspector state initialization may be incomplete
3. ⏳ **Implement fixes** - Apply proposed fixes based on investigation
4. ⏳ **Test thoroughly** - Verify inspector works correctly on first open
5. ⏳ **Document changes** - Update migration docs with inspector integration notes

---

## Related Documentation

- [MSG_CMD_MIGRATION_100_COMPLETE.md](./MSG_CMD_MIGRATION_100_COMPLETE.md) - Migration completion report
- [INSPECTOR_MSG_CMD_REFACTOR.md](./INSPECTOR_MSG_CMD_REFACTOR.md) - Inspector-specific refactoring notes
- `framework/app.go` - Main event loop and render logic
- `internal/inspector/standalone_inspector.go` - Inspector implementation
- `internal/inspector/hook.go` - Hook integration
