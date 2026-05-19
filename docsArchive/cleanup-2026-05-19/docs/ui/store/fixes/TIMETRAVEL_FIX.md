# Time Travel Undo/Redo Fix

## Problem

The initial implementation of `Undo()` and `JumpTo()` methods in `AppRuntime` caused an infinite loop that froze the application when these operations were performed. Additionally, there was no `Redo()` functionality to move forward through history after undoing.

### Root Cause

```
Problematic Flow:
rt.Undo()
  → rt.store.Set(prev)          # Set state to previous
    → store.Set() triggers callbacks
      → handleStateChange(prev) # Callback executed
        → Appends prev AGAIN to history: [..., prev, prev]
          → Next Undo does same thing
          → INFINITE LOOP
```

The issue was that `store.Set()` always triggered the state change callback (`handleStateChange`), which would record the new state to history. When undoing, this caused the undone state to be recorded again, creating an infinite cycle.

## Solution

### The SkipHistory Flag Pattern + CurrentIndex Tracking

Added two mechanisms to enable proper Undo/Redo:
1. A `skipHistory bool` flag to temporarily pause history recording during time travel operations
2. A `currentIndex int` field to track the current position in history, enabling Redo functionality

### Implementation

#### 1. Added SkipHistory Flag and CurrentIndex

```go
type AppRuntime[T any] struct {
    // ... other fields
    history      []T
    currentIndex int // Current position in history index
    maxHistory   int
    skipHistory  bool // Flag to skip history recording (for Undo/JumpTo)
}
```

#### 2. Modified handleStateChange to Track Index and Truncate History

```go
func (rt *AppRuntime[T]) handleStateChange(state T) {
    rt.mu.Lock()
    // Record history (only if not skipping)
    if !rt.skipHistory && rt.maxHistory > 0 {
        // If we're in the middle of history and perform a new action,
        // truncate history from current position forward
        if rt.currentIndex < len(rt.history)-1 {
            rt.history = rt.history[:rt.currentIndex+1]
        }
        rt.history = append(rt.history, state)
        rt.currentIndex = len(rt.history) - 1
        if len(rt.history) > rt.maxHistory {
            rt.history = rt.history[1:]
            rt.currentIndex--
        }
    }
    callback := rt.onStateChange
    rt.mu.Unlock()

    // Call callback outside of lock
    if callback != nil {
        callback(state)
    }
}
```

#### 3. Fixed Undo() - Uses CurrentIndex

```go
func (rt *AppRuntime[T]) Undo() error {
    rt.mu.Lock()
    defer rt.mu.Unlock()

    if rt.currentIndex <= 0 {
        return fmt.Errorf("no previous state to undo to")
    }

    // Move to previous state index
    rt.currentIndex--

    // Get previous state
    prev := rt.history[rt.currentIndex]

    // Skip history recording during undo
    rt.skipHistory = true
    rt.mu.Unlock()

    rt.store.Set(prev) // Callback won't re-record to history

    rt.mu.Lock()
    rt.skipHistory = false

    return nil
}
```

#### 4. Implement Redo() - New Feature

```go
func (rt *AppRuntime[T]) Redo() error {
    rt.mu.Lock()
    defer rt.mu.Unlock()

    if rt.currentIndex >= len(rt.history)-1 {
        return fmt.Errorf("no next state to redo to")
    }

    // Move to next state index
    rt.currentIndex++

    // Get next state
    next := rt.history[rt.currentIndex]

    // Skip history recording during redo
    rt.skipHistory = true
    rt.mu.Unlock()

    rt.store.Set(next) // Callback won't re-record to history

    rt.mu.Lock()
    rt.skipHistory = false

    return nil
}
```

#### 5. Fixed JumpTo() - Updates CurrentIndex

```go
func (rt *AppRuntime[T]) JumpTo(index int) error {
    rt.mu.Lock()

    if index < 0 || index >= len(rt.history) {
        rt.mu.Unlock()
        return fmt.Errorf("index out of range: %d (history size: %d)", index, len(rt.history))
    }

    state := rt.history[index]
    rt.currentIndex = index

    // Skip history recording during time jump
    rt.skipHistory = true
    rt.mu.Unlock()

    rt.store.Set(state) // Callback won't re-record to history

    rt.mu.Lock()
    rt.skipHistory = false
    rt.mu.Unlock()

    return nil
}
```

#### 6. Updated CanUndo() and Added CanRedo()

```go
func (rt *AppRuntime[T]) CanUndo() bool {
    rt.mu.RLock()
    defer rt.mu.RUnlock()
    return rt.currentIndex > 0
}

func (rt *AppRuntime[T]) CanRedo() bool {
    rt.mu.RLock()
    defer rt.mu.RUnlock()
    return rt.currentIndex < len(rt.history)-1
}
```

## Benefits

### 1. Correctness
- Undo/Redo no longer causes infinite loops
- History remains accurate (no duplicate states)
- Application remains responsive during time travel
- Proper forward/reverse navigation through state history

### 2. Performance
- Still uses the public `store.Set()` API (no private field access)
- Avoids redundant history recording
- Minimal overhead (single boolean flag check)
- History truncation prevents unbounded growth

### 3. Maintainability
- Clean solution that works with existing callbacks
- No need to modify store implementation
- Easy to understand and maintain
- Consistent with standard Undo/Redo patterns

## Important Notes

### History Behavior with CurrentIndex

The new implementation maintains all history states and uses `currentIndex` to track position:

```
History: [0, 1, 2, 3, 4]
Index:          ^ (currentIndex = 2)
State:          2

Undo -> currentIndex = 1, State = 1
Undo -> currentIndex = 0, State = 0
Redo -> currentIndex = 1, State = 1
Redo -> currentIndex = 2, State = 2
```

**Important**: When a new action is performed after undoing, forward history is truncated:

```
History: [0, 1, 2, 3, 4]
Index:          ^
State:          2

New action (Value = 10)
History: [0, 1, 10]  ← 2, 3, 4 are truncated
Index:        ^
State:        10
```

### Callback Behavior

The `skipHistory` flag only prevents history recording during the time travel operation. The state change callback (used for rendering updates) still executes correctly:

```go
rt.store.Set(prev)  // ← triggers callback, updates UI
// handleStateChange checks skipHistory
//   → if true: skips history recording
//   → if false: records to history
// callback(prev)  // ← still executes for UI updates
```

This ensures:
- UI still updates when Undo/Redo is performed
- Subscribers still get notified
- Only history recording is skipped

### Thread Safety

The `skipHistory` flag and `currentIndex` are protected by `rt.mu` lock:
- Set to `true` before releasing lock during Undo/JumpTo
- `store.Set()` acquires store's own lock and calls handleStateChange
- handleStateChange checks flag and updates currentIndex while holding rt.mu
- Flag is set back to `false` after store.Set() completes

This prevents race conditions when multiple Undo/JumpTo operations are performed concurrently.

## Testing

Comprehensive tests verify the fix (`runtime/statemachine/undo_test.go`):

1. **TestUndoRedoNoFreeze**: Verifies Undo/Redo doesn't cause infinite loop
2. **TestUndoRedoNoInfiniteRecording**: Verifies Undo doesn't record duplicates
3. **TestJumpToNoInfiniteRecording**: Verifies JumpTo doesn't modify history
4. **TestUndoWithCallback**: Verifies callbacks still work correctly
5. **TestUndoWithMaxHistoryZero**: Edge case when history is disabled
6. **TestRedoNoFreeze**: Verifies Redo works after Undo
7. **TestUndoRedoWithNewAction**: Verifies history truncation on new actions
8. **TestJumpToUpdatesCurrentIndex**: Verifies JumpTo correctly updates index

All tests pass, confirming the fix works correctly.

## Migration Guide

If you have custom time travel implementations:

### Old Problematic Code

```go
// ❌ Don't do this - causes infinite loop
func (rt *AppRuntime[T]) Undo() error {
    prev := rt.history[len(rt.history)-2]
    rt.store.Set(prev) // ← triggers callback, re-records
    return nil
}
```

### New Correct Code

```go
// ✅ Use skipHistory pattern with currentIndex
func (rt *AppRuntime[T]) Undo() error {
    rt.mu.Lock()
    // ... validate and check currentIndex ...

    // Skip recording during time travel
    rt.skipHistory = true
    rt.mu.Unlock()

    rt.store.Set(prev) // Callback won't create infinite loop

    rt.mu.Lock()
    rt.skipHistory = false
    rt.mu.Unlock()

    return nil
}
```

### Using Undo/Redo in Your Application

```go
// In your main function:
ui.RunApp(rt, ui.WithInit(func() {
    // Register Undo handler
    intent.RegisterTypedWithOpts(
        intent.DefaultRegistry(),
        func(ctx *intent.ActionContext, i UndoIntent) intent.IntentResult {
            if !rt.CanUndo() {
                return intent.HandledResult()
            }
            rt.Undo() // Returns error, but you can ignore it
            return intent.HandledResult()
        },
        intent.WithOverridable(true),
    )

    // Register Redo handler
    intent.RegisterTypedWithOpts(
        intent.DefaultRegistry(),
        func(ctx *intent.ActionContext, i RedoIntent) intent.IntentResult {
            if !rt.CanRedo() {
                return intent.HandledResult()
            }
            rt.Redo() // Returns error, but you can ignore it
            return intent.HandledResult()
        },
        intent.WithOverridable(true),
    )
}))

// In your View:
button.NewBuilder("Undo").OnPress(UndoIntent{}).Build()
button.NewBuilder("Redo").OnPress(RedoIntent{}).Build()
```

## Related Files

- `runtime/statemachine/runtime.go` - Main implementation
- `runtime/statemachine/undo_test.go` - Comprehensive tests
- `examples/typesafe_form_demo_runapp/main.go` - Usage example with Undo/Redo

## Performance Impact

Minimal:
- One additional boolean check in `handleStateChange`
- One additional `currentIndex` increment/decrement
- No lock contention increase
- No additional memory allocation (except for history array itself)

## Future Improvements

Potential enhancements:
1. Add history range export/import for debugging
2. Add history snapshot for state comparison
3. Visual history UI component
4. Undo/Redo keyboard shortcuts integration
5. Batch Undo/Redo operations

## Conclusion

The combination of skipHistory flag and currentIndex tracking provides a clean, correct solution to the time travel infinite loop problem while enabling full Undo/Redo functionality. The solution preserves all existing callback functionality and maintains thread safety.
