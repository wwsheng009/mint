# Debug: Click Event Not Triggering Issue

**Date**: 2026-02-17
**Status**: Investigating
**Priority**: High

## Summary

Mouse click and Enter key events are not triggering Button handlers in demo1 tests. The counter stays at 0 after clicking "Add Count" button.

## Test Failures

```
=== RUN   TestActionSystemEnter
    action_system_test.go:98: ActionEnter did not increment counter. State: Clicks: 0
--- FAIL: TestActionSystemEnter (0.47s)

=== RUN   TestClickCount
    layer_test.go:297: Click count not working properly: Clicks: 0
--- FAIL: TestClickCount (1.33s)

=== RUN   TestModalOpenClick
    layer_test.go:75: Modal should be visible after clicking button: render does not contain expected text
--- FAIL: TestModalOpenClick (0.83s)
```

## Architecture Overview

### Fiber-first Action Architecture

According to design documents (`fiber_event.md`, `fiber_action.md`):

```
Event Flow:
Input → HitTest → Fiber.ActionTargetID → ActionBridge → Router → HandleAction
```

### Two Processing Paths in App.runLoop

```go
// framework/app.go:828-835
if a.actionRouter != nil && a.inputProcessor != nil {
    a.processMsg(msg)    // Path 1: Action path (takes priority)
    continue
}
handled := a.handleMsg(msg)  // Path 2: Legacy path
```

### Key Components

1. **InputProcessor** (`framework/action/processor.go`)
   - Converts `Msg` to `Action`
   - **CRITICAL**: Returns `nil` if `mouseMsg.TargetID == 0` (line 77-79)

2. **ActionBridge** (`runtime/bridge/actionbridge/bridge.go`)
   - Bridges Fiber tree and Action Router
   - Supports both modes:
     - Semantic Action: `Fiber.ActionTargetID` → Router → registered handler
     - Closure mode: `FocusableVNode.HandleAction()` → onClick callback

3. **HitMap** (`runtime/event/hitmap.go`)
   - Maps screen coordinates to Fiber nodes
   - `HitMapEntry.TargetFiber` should reference the Fiber

4. **Pump** (`framework/event/pump.go`)
   - Fills `MouseMsg.TargetFiber` from HitMap

## Root Cause Analysis

### Problem 1: InputProcessor Returns nil for Closure Mode

**File**: `framework/action/processor.go:75-79`

```go
func (p *InputProcessor) processMouseMsg(mouseMsg *runtimemsg.MouseMsg) *Action {
    // If no target, don't process
    if mouseMsg.TargetID == 0 {
        return nil  // ❌ PROBLEM: TargetID is 0 in closure mode!
    }
    // ...
}
```

In closure mode, Button doesn't set explicit `ActionTargetID`, so `TargetID` is 0. This causes `InputProcessor.ProcessMsg()` to return `nil`.

### Problem 2: processMsg Doesn't Fallback to handleMsg

**File**: `framework/app.go:990-997`

```go
func (a *App) processMsg(msg runtimemsg.Msg) {
    act := a.inputProcessor.ProcessMsg(msg)
    
    if act == nil {
        a.handleSystemMsg(msg)  // ❌ Mouse events not handled here
        return
    }
    // ...
}
```

When `InputProcessor` returns `nil`, the code falls through to `handleSystemMsg()` which only handles Resize and Quit events. Mouse events are lost!

### Fix Attempted (Partial)

Added fallback to `handleMsg` for mouse events:

```go
if act == nil {
    // 鼠标事件：尝试通过 TargetFiber 路由（闭包模式）
    if _, ok := msg.(*runtimemsg.MouseMsg); ok {
        if a.handleMsg(msg) {
            return
        }
    }
    a.handleSystemMsg(msg)
    return
}
```

**Result**: Tests still failing. The fix doesn't fully resolve the issue.

## Remaining Issues to Investigate

### Issue A: Is TargetFiber Being Set?

**Question**: Is `HitMapEntry.TargetFiber` properly populated?

**Check Points**:
1. `runtime/ui/fiber_util.go:445` - Sets `TargetFiber: fiber` ✓
2. `framework/event/pump.go:232` - Copies `targetFiber = entry.TargetFiber` ✓
3. `framework/event/pump.go:276` - Sets `mouseMsg.TargetFiber = targetFiber` ✓

**Verify**: Add debug logging to confirm `TargetFiber` is non-nil when clicking.

### Issue B: Does ActionBridge Find FocusableVNode?

**File**: `runtime/bridge/actionbridge/bridge.go:64-71`

```go
// Mode 2: Closure mode (FocusableVNode implements ActionTarget)
if f.FocusableVNode != nil {
    if target, ok := f.FocusableVNode.(action.ActionTarget); ok {
        a := action.NewAction(actionType).WithPayload(payload)
        if target.HandleAction(a) {
            return true
        }
    }
}
```

**Question**: Is `Fiber.FocusableVNode` set for Button components?

**Check Point**: `runtime/ui/fiber_util.go:117-119`
```go
var focusableVNode FocusableVNode
if f, ok := vnode.(FocusableVNode); ok && f.IsFocusable() {
    focusableVNode = f
}
```

**Verify**: Button implements `FocusableVNode` interface. Check if it's being set during Fiber creation.

### Issue C: MouseMsg.Type() Check

In `handleMsg`, we check:
```go
if mouseMsg, ok := message.(*runtimemsg.MouseMsg); ok {
    if mouseMsg.TargetFiber != nil {
        if fiber, ok := mouseMsg.TargetFiber.(*rtui.Fiber); ok {
```

**Question**: Is the type assertion `mouseMsg.TargetFiber.(*rtui.Fiber)` succeeding?

The `TargetFiber` field type is:
```go
TargetFiber interface {
    GetActionTargetID() string
}
```

But in `fiber_util.go`, we set `TargetFiber: fiber` where `fiber` is `*Fiber`.

**Verify**: Check if `*Fiber` implements `GetActionTargetID() string`.

### Issue D: Enter Key Event Path

For keyboard events (Enter key):
```go
if keyMsg, ok := message.(*runtimemsg.KeyMsg); ok {
    if a.focusManager != nil {
        focused := a.focusManager.GetCurrent()
        if focused != nil {
            actionType := a.keyMsgToActionType(keyMsg)
            if a.actionBridge.DispatchFromFiber(focused, actionType, keyMsg) {
```

**Question**: Is `focusManager.GetCurrent()` returning the correct focused Fiber?

**Verify**: Check FocusManager state when Enter is pressed.

## Debug Steps

### Step 1: Add Logging to Pump

```go
// framework/event/pump.go - convertToMouseMsg
log.UILogger.Debug("HitTest: TargetFiber=%v, FocusableVNode=%v", 
    entry.TargetFiber != nil,
    entry.TargetFiber.(*rtui.Fiber).FocusableVNode != nil)
```

### Step 2: Add Logging to ActionBridge

```go
// runtime/bridge/actionbridge/bridge.go - DispatchFromFiber
log.UILogger.Debug("DispatchFromFiber: ActionTargetID=%s, FocusableVNode=%v",
    f.ActionTargetID,
    f.FocusableVNode != nil)
```

### Step 3: Add Logging to Button.HandleAction

```go
// components/button/button.go - HandleAction
log.UILogger.Debug("Button.HandleAction: type=%v, onClick=%v", 
    act.Type, 
    b.onClick != nil)
```

## Hypotheses

### Hypothesis 1: TargetFiber is nil
- HitMap is not being populated correctly
- Pump.SetHitMap is not called with the latest HitMap

### Hypothesis 2: FocusableVNode is nil on Fiber
- Button VNode is not implementing FocusableVNode correctly
- Fiber creation doesn't preserve FocusableVNode reference

### Hypothesis 3: Type Assertion Fails
- `TargetFiber` interface type doesn't match `*Fiber`
- Need to verify `*Fiber` implements `GetActionTargetID()`

### Hypothesis 4: onClick callback is nil
- Button's `onClick` is not being set during component creation
- Closure reference is lost during Fiber reconciliation

## Code References

| Component | File | Key Lines |
|-----------|------|-----------|
| InputProcessor | `framework/action/processor.go` | 75-79 |
| App.processMsg | `framework/app.go` | 985-1023 |
| App.handleMsg | `framework/app.go` | 905-938 |
| ActionBridge | `runtime/bridge/actionbridge/bridge.go` | 44-74 |
| HitMapEntry | `runtime/event/hitmap.go` | 54-78 |
| BuildHitMapFromFiber | `runtime/ui/fiber_util.go` | 385-460 |
| Pump.convertToMouseMsg | `framework/event/pump.go` | 174-282 |
| Button.HandleAction | `components/button/button.go` | 842-861 |

## Next Steps

1. [ ] Add debug logging to trace event flow
2. [ ] Run tests with verbose logging enabled
3. [ ] Verify TargetFiber is populated in HitMap
4. [ ] Verify FocusableVNode is set on Fiber
5. [ ] Check if Button.onClick callback is preserved
6. [ ] Investigate Enter key path through FocusManager
