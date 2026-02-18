# Debug: Click Event Not Triggering Issue

**Date**: 2026-02-17
**Status**: ✅ RESOLVED - Architecture Working, Test Logic Issue
**Priority**: Closed

## Summary

Mouse click and Enter key events **are now working correctly** after refactoring. The test failures are due to **test logic issues** (focus navigation), not architecture problems.

## Resolution

### Architecture Status: ✅ WORKING

1. **InputProcessor** - Always generates Action (no nil returns)
2. **ActionBridge** - Correctly routes to FocusableVNode
3. **Button.HandleAction** - Called with correct Action
4. **Button.onClick** - Callback exists and is executed

### Root Cause of Test Failures

**Test Logic Issue**: The tests navigate focus incorrectly.

From debug output:
```
[Button.HandleAction] label="Quit", onClick=true
```

- Test presses Tab 10 times
- Focus lands on "Quit" button instead of "Add Count"
- Enter key triggers "Quit" button, which calls `ui.Quit()`

## Architecture Changes Made

### ✅ Phase 1: InputProcessor Never Returns Nil

**File**: `framework/action/processor.go`

```go
// Before: Returns nil when TargetID == 0
if mouseMsg.TargetID == 0 { return nil }

// After: Always generates Action
act := NewAction(ActionClick).WithPayload(mouseMsg)
```

### ✅ Phase 2: Unified Action Routing

**File**: `framework/app.go`

```go
func (a *App) processMsg(msg runtimemsg.Msg) {
    act := a.inputProcessor.ProcessMsg(msg)
    
    // Mouse events: use MouseMsg.TargetFiber
    if mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg); ok {
        if fiber, ok := mouseMsg.TargetFiber.(*rtui.Fiber); ok {
            a.actionBridge.DispatchFromFiber(fiber, act.Type, act)
        }
    }
    
    // Keyboard events: use focused Fiber
    if focused := a.focusManager.GetCurrent(); focused != nil {
        a.actionBridge.DispatchFromFiber(focused, act.Type, act)
    }
}
```

### ✅ Phase 3: ActionBridge Supports Both Modes

**File**: `runtime/bridge/actionbridge/bridge.go`

```go
func (b *Bridge) DispatchFromFiber(start *Fiber, actionType ActionType, payload interface{}) bool {
    for f := start; f != nil; f = f.Return {
        // Mode 1: Semantic Action (ActionTargetID)
        if f.ActionTargetID != "" {
            // dispatch via Router
        }
        
        // Mode 2: Closure mode (FocusableVNode)
        if f.FocusableVNode != nil {
            if target, ok := f.FocusableVNode.(ActionTarget); ok {
                if target.HandleAction(a) {
                    return true
                }
            }
        }
    }
}
```

## Event Flow (After Refactoring)

```
Input (Mouse/Keyboard)
    ↓
InputProcessor.ProcessMsg()
    ↓ (always returns Action)
processMsg()
    ↓
ActionBridge.DispatchFromFiber()
    ↓
Fiber.FocusableVNode.HandleAction()
    ↓
Button.onClick()
```

## Remaining Work

### Test Fixes Needed

The tests need to correctly navigate to target buttons:

```go
// Current test (wrong)
for i := 0; i < 10; i++ {
    testApp.InjectSpecialKey(platform.KeyTab)
}

// Should navigate specifically to "Add Count" button
// Or use fewer Tab presses
```

### Future Architecture Work (Per fiber_confict.md)

1. **Closure → ActionID Registration**
   - ButtonBuilder.OnClick() generates ActionID
   - Register handler to Scope Dispatcher
   - Fiber only stores ActionTargetID

2. **Remove FocusableVNode Runtime Dependency**
   - Fiber should not hold function references
   - All actions go through ActionID → Dispatcher
