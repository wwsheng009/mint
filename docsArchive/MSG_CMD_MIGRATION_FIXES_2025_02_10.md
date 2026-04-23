# Msg/Cmd Migration Fixes - 2025-02-10

> **Date**: 2026-02-10
> **Commit**: Work after 5afab0a8 (feat: Complete Msg/Cmd architecture migration)

---

## Summary

After the Msg/Cmd migration (commit 5afab0a8), several critical issues were discovered that prevented the application from working correctly. This document summarizes the issues found and fixes applied.

---

## Issues Fixed

### 1. ✅ Modal Component Event Handling

**Problem**: Modal dialogs were completely broken for user interaction after Msg/Cmd migration.

**Root Cause**: Modal had no event handling - no `HandleEvent()` and no `Update(Msg)` methods.

**Fix** (`components/overlay/modal.go`):
- Added interface assertions for `frameworkevent.Component` and `component.Updater`
- Implemented `HandleEvent()` for backward compatibility
- Implemented `Update(Msg)` for Msg/Cmd architecture
- Added bounds tracking during `Paint()` for hit testing
- Implemented ESC key to close modal
- Implemented click-outside-to-close functionality
- Added `onClose` callback field and setter

**Code Added**:
```go
// HandleEvent processes events (legacy, for backward compatibility)
func (m *ModalVNode) HandleEvent(e frameworkevent.Event) bool {
    // Handle ESC key to close
    // Handle click outside modal
    return false for internal clicks (let them route to children)
}

// Update implements component.Updater interface for Msg/Cmd architecture
func (m *ModalVNode) Update(message runtimemsg.Msg) cmd.Cmd {
    // Handle ESC and click-outside-to-close
}

// Added bounds tracking
type ModalVNode struct {
    ...
    bounds [4]int // x, y, width, height
    onClose func()
}
```

---

### 2. ✅ Modal Internal Button Clicks

**Problem**: Buttons inside modal dialogs were not responding to mouse clicks.

**Root Cause**: Modal's `content` and `footer` fields were not exposed via `Children()` method, so HitMap couldn't see or register modal's child components.

**Fix** (`components/overlay/modal.go`):
- Implemented `Children()` method that returns content and footer
- This allows HitMap to discover and register modal's child components
- Buttons now receive mouse events via direct Msg routing

**Code Added**:
```go
// Children returns all child nodes for HitMap building
func (m *ModalVNode) Children() []ui.VNode {
    var children []ui.VNode

    if m.content != nil {
        children = append(children, m.content)
    }

    if m.footer != nil {
        children = append(children, m.footer)
    }

    // Also include any children set via ElementVNode
    baseChildren := m.ElementVNode.Children()
    if len(baseChildren) > 0 {
        children = append(children, baseChildren...)
    }

    return children
}
```

---

### 3. ✅ Screen Resize Events

**Problem**: Screen resize events were not working after Msg/Cmd migration.

**Root Cause**: Pump was creating generic `BaseMsg` for resize events without width/height information.

**Fix** (`runtime/msg/resize_msg.go`):
- Created new `ResizeMsg` type with proper dimension fields
- Updated `Pump.convertToResizeMsg()` to use new type
- Updated `MsgToEvent()` adapter to properly convert `ResizeMsg` to `ResizeEvent`

**Code Added**:
```go
// resize_msg.go
type ResizeMsg struct {
    BaseMsg
    OldWidth, OldHeight int
    NewWidth, NewHeight int
}

func NewResizeMsg(oldW, oldH, newW, newH int) *ResizeMsg {
    return &ResizeMsg{
        BaseMsg: BaseMsg{
            TypeValue:      MsgTypeResize,
            TimestampValue: time.Now(),
        },
        OldWidth:  oldW,
        OldHeight: oldH,
        NewWidth:  newW,
        NewHeight: newH,
    }
}
```

**Code Updated**:
```go
// pump.go
func (p *Pump) convertToResizeMsg(raw platform.RawInput) runtimemsg.Msg {
    oldWidth := 0
    oldHeight := 0
    return runtimemsg.NewResizeMsg(oldWidth, oldHeight, raw.Width, raw.Height)
}

// msg_adapter.go
case *runtimemsg.ResizeMsg:
    return &ResizeEvent{
        BaseEvent: NewBaseEvent(EventResize),
        OldWidth:  m.OldWidth,
        OldHeight: m.OldHeight,
        NewWidth:  m.NewWidth,
        NewHeight: m.NewHeight,
    }
```

---

### 4. ✅ Modal Click-Outside Logic

**Problem**: Modal was intercepting all mouse clicks, preventing internal buttons from receiving events.

**Fix**: Modified `HandleEvent()` and `updateMouse()` to only handle clicks outside modal bounds, and return `false` for internal clicks to allow routing to child components.

---

## Remaining Issues

### ⏳ Inspector Display Issues

**Problems**:
1. Inspector doesn't show interface on first open (F12)
2. Requires pressing Enter to display
3. Rendering is incomplete
4. Cannot display in full screen properly

**Suspected Root Causes**:
- Inspector state initialization timing
- Hook may be checking `IsVisible()` before toggle completes
- `RenderContent()` may return nil or incomplete on first call
- Layer rendering order may be incorrect

**Status**: PENDING investigation

---

### ⏳ Border Rendering Issues

**Problem**: Borders not rendering completely

**Suspected Root Causes**:
- Width calculation doesn't account for wide characters
- Border characters being clipped
- Layer rendering order issues

**Status**: PENDING investigation

---

## Testing Results

All core packages compile successfully:
- ✅ `./components/...`
- ✅ `./framework/...`
- ✅ `./runtime/msg/...`

Demo application builds:
- ✅ `examples/ui_demos/demo2_runtime_internals/inspector_overlay`

---

## Files Modified

### New Files:
1. `runtime/msg/resize_msg.go` - Resize message type with dimensions
2. `docs/plan/event/INSPECTOR_MSG_MIGRATION_ISSUES.md` - Issue tracking document

### Modified Files:
1. `components/overlay/modal.go` - Added event handling, bounds tracking, Children() method
2. `framework/event/pump.go` - Updated convertToResizeMsg() to use ResizeMsg
3. `framework/event/msg_adapter.go` - Added ResizeMsg case to MsgToEvent()
4. `framework/testing/integration_test.go` - Fixed import (msg → runtimemsg)

---

## Related Documentation

- [MSG_CMD_MIGRATION_100_COMPLETE.md](./MSG_CMD_MIGRATION_100_COMPLETE.md) - Original migration completion report
- [INSPECTOR_MSG_MIGRATION_ISSUES.md](./INSPECTOR_MSG_MIGRATION_ISSUES.md) - Detailed issue tracking

---

## Next Steps

1. **HIGH PRIORITY** - Fix Inspector display issues
   - Debug first-open problem
   - Investigate hook timing
   - Verify RenderContent() initialization

2. **MEDIUM PRIORITY** - Fix border rendering
   - Investigate BorderedNode rendering
   - Fix width calculations

3. **LOW PRIORITY** - Fix legacy code
   - Fix `ev.Click` errors in runtime/selection and runtime/engine
   - Update examples and tests to match new API
