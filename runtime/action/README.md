# Unified Action System

> **Merged from `framework/action` and `runtime/action`**
>
> This is the result of merging two action systems into a unified, feature-rich implementation.

## Overview

The Unified Action System provides:
- **Semantic user actions** (navigation, editing, forms, mouse, etc.)
- **Three-phase routing** (Capture → Target → Bubble) with middleware support
- **Composite actions** (concurrent/sequential execution with worker pools)
- **Input processing** with configurable KeyMap
- **Scope-based dispatchers** for action isolation
- **8 built-in middleware types** (logging, throttle, validation, metrics, recovery, profiling, caching, audit)
- **Backward compatibility adapters** for smooth migration

## Design Principles

1. **Semantic over Raw**: Actions represent user intent, not raw input
2. **Immutable Payloads**: Action payloads must be value types (string, int, struct)
3. **Traceable**: All state changes traceable to actions for debugging and replay
4. **Dual ID Support**: Supports both string (semantic) and uint64 (fast) target IDs
5. **Extensible**: Middleware system for cross-cutting concerns

## Core Components

| Component | Description |
|-----------|-------------|
| `Action` | Core action struct with type, payload, target (string & uint64) |
| `Target` | Interface for receiving actions |
| `Router` | Three-phase action routing (Capture/Target/Bubble) |
| `Dispatcher` | Simple action distribution with global handlers |
| `InputProcessor` | Converts low-level events to semantic actions |
| `KeyMap` | Keyboard to action mappings (global and context-aware) |
| `ScopeDispatcher` | Isolated action handling with parent bubbling |
| `Middleware` | 8 middleware types for cross-cutting concerns |
| `Errors` | Structured error handling |

## Action Structure

```go
type Action struct {
    // Core (from runtime/action)
    Type      ActionType
    Payload   interface{}
    Source    string
    Target    string      // Semantic target ID
    TargetID  uint64      // Fast lookup ID (auto-computed)

    // Enhanced (from framework/action)
    ID        uint64
    Timestamp time.Time
    Meta      map[string]interface{}

    // Propagation control
    stopped   bool
}
```

### Action Types (50+)

| Category | Actions |
|----------|---------|
| **Navigation** | `NavigateFirst`, `NavigateLast`, `NavigateNext`, `NavigatePrev`, `NavigateUp`, `NavigateDown`, `NavigateLeft`, `NavigateRight`, `NavigatePageUp`, `NavigatePageDown`, `NavigateHome`, `NavigateEnd` |
| **Editing** | `InputChar`, `InputText`, `DeleteChar`, `DeleteWord`, `DeleteLine`, `Backspace`, `CursorHome`, `CursorEnd`, `CursorLeft`, `CursorRight`, `CursorWordLeft`, `CursorWordRight`, `SelectAll`, `SelectWord`, `SelectLine`, `Enter` |
| **Form** | `Submit`, `Cancel`, `Validate`, `Reset`, `Clear` |
| **Selection** | `SelectItem`, `DeselectItem`, `ToggleSelect`, `SelectRange`, `Select`, `Toggle`, `Expand`, `Collapse` |
| **Mouse** | `Click`, `DoubleClick`, `TripleClick`, `MousePress`, `MouseRelease`, `MouseMotion`, `MouseDrag`, `MouseWheel`, `RightClick`, `MiddleClick`, `Hover`, `DragStart`, `DragMove`, `DragEnd` |
| **View** | `Scroll`, `ScrollUp`, `ScrollDown`, `ScrollLeft`, `ScrollRight`, `ZoomIn`, `ZoomOut`, `ZoomReset`, `Resize` |
| **Window** | `Quit`, `Close`, `Maximize`, `Minimize`, `Fullscreen` |
| **System** | `Copy`, `Cut`, `Paste`, `Undo`, `Redo`, `Search`, `Help`, `Refresh`, `Inspect`, `Focus`, `Blur` |
| **Focus** | `FocusGained`, `FocusLost`, `FocusNext`, `FocusPrev` |
| **Data** | `DataLoad`, `DataUpdate`, `DataError` |
| **AI** | `AIInspect`, `AIFind`, `AIQuery`, `AIDispatch`, `AIWait`, `AIWatch` |
| **Lifecycle** | `Init`, `Mount`, `Unmount` |

## Three-Phase Routing

```
Event → InputProcessor → Action → Router
                                   ↓
                          ┌─────────────────────────────┐
                          │  Phase 1: Capture (root→target) │
                          │  - Global capture handlers    │
                          └─────────────────────────────┘
                                   ↓
                          ┌─────────────────────────────┐
                          │     Phase 2: Target            │
                          │  - Target.HandleAction()      │
                          └─────────────────────────────┘
                                   ↓
                          ┌─────────────────────────────┐
                          │  Phase 3: Bubble (target→root)  │
                          │  - Global bubble handlers    │
                          └─────────────────────────────┘
```

### Router Usage

```go
router := action.NewRouter(rootNode)

// Add capture phase handler (with priority)
router.AddCaptureHandler(myInspector, 100)

// Add bubble phase handler
router.AddBubbleHandler(myLogger, "logger")

// Set middleware chain
router.SetMiddleware(action.DefaultMiddlewareChain())

// Build target registry
router.BuildTargetRegistry()

// Dispatch action
result := router.Dispatch(myAction)
```

## Target Capabilities

Components can implement optional capability interfaces:

```go
type Focusable interface {     // Can receive/be focused
    Focus() bool
    Blur()
    IsFocused() bool
    IsFocusable() bool
}

type Scrollable interface {    // Supports scrolling
    CanScroll(delta int) bool
    Scroll(delta int) bool
    GetScrollPosition() (current, total, visible int)
}

type Editable interface {      // Supports text editing
    InsertText(text string) bool
    DeleteText(direction int) bool
    GetText() string
    SetCursorPosition(pos int) bool
    GetCursorPosition() int
}

type Selectable interface {   // Supports selection
    Select() bool
    ToggleSelection() bool
    IsSelected() bool
    GetSelectedCount() int
}

type Expandable interface {   // Supports expand/collapse
    Expand() bool
    Collapse() bool
    Toggle() bool
    IsExpanded() bool
}

type Draggable interface {    // Supports dragging
    StartDrag(act *Action) bool
    Drag(act *Action) bool
    EndDrag(act *Action) bool
    IsDragging() bool
}
```

## Dispatcher (Simple Distribution)

```go
// Create dispatcher
d := action.NewDispatcher()

// Register targets
d.Register(myComponent)

// Subscribe global handlers
quitUnsub := d.Subscribe(action.ActionQuit, func(a *action.Action) bool {
    // Handle quit
    return true
})
defer quitUnsub()

// Set default handler
d.SetDefaultHandler(func(a *action.Action) bool {
    fmt.Printf("Unhandled: %s\n", a.Type)
    return false
})

// Enable logging
d.EnableLog(true)

// Dispatch
d.Dispatch(action.NewAction(action.ActionSubmit).WithTarget("form1"))
```

## Input Processing

```go
processor := action.NewInputProcessor()

// Set custom keymap
processor.SetKeyMap(action.DefaultKeyMap())

// Add custom bindings
processor.AddKeyMapping(action.KeyMapping{
    KeySpec: "ctrl+f",
    Action:  action.NewAction(action.ActionFind),
    Context: "",
})

// Process raw messages
action := processor.ProcessMsg(keyMsg)
if action != nil {
    dispatcher.Dispatch(action)
}
```

## KeyMap

```go
km := action.NewKeyMap()

// Global bindings
km.Bind("ctrl+c", action.NewAction(action.ActionCopy))
km.Bind("ctrl+v", action.NewAction(action.ActionPaste))

// Context-specific bindings
km.BindWithContext("editor", "ctrl+f", action.NewAction(action.ActionFind))
km.PushContext("editor")

// Lookup
action := km.LookupKeyMsg(keyMsg)
```

## Middleware System

### Built-in Middleware (8 Types)

| Middleware | Purpose |
|------------|---------|
| **Logging** | Records action dispatch and results |
| **Throttle** | Prevents action spam (configurable interval) |
| **Validation** | Validates actions before dispatch |
| **Metrics** | Collects statistics (counts, durations, errors) |
| **Recovery** | Catches panics during action processing |
| **Profiling** | Performance profiling with hotspots |
| **Caching** | Caches handler results |
| **Audit** | Logs all actions for compliance |

### Middleware Chains

```go
// Default middleware
chain := action.DefaultMiddlewareChain()
// Contains: Recovery, Throttle, Validation, Metrics, Logging

// Full middleware (all 8 types)
chain := action.FullMiddlewareChain()

// Production middleware
chain := action.ProductionMiddlewareChain()

// Debug middleware
chain := action.DebugMiddlewareChain()

// Custom chain
chain := action.NewMiddlewareChain(
    action.NewRecoveryMiddleware(),
    action.NewThrottleMiddleware(16*time.Millisecond),
    action.NewValidationMiddleware(),
)

// Add custom middleware
chain.Add(customMiddleware)
```

### Custom Middleware

```go
type CustomMiddleware struct{}

func (m *CustomMiddleware) Name() string {
    return "custom"
}

func (m *CustomMiddleware) Before(action *action.Action) *action.Action {
    // Pre-process action
    return action
}

func (m *CustomMiddleware) After(action *action.Action, result *action.RouterResult) {
    // Post-process result
}
```

## Scope Dispatcher

```go
// Create root scope
root := action.NewScopeDispatcher(nil)
root.SetScopeName("root")

// Create child scope
child := root.CreateChildScope("component1")

// Register handlers
actionID := action.GenerateScopeActionID()
child.Register(actionID, func(act *action.Action) bool {
    // Handle action in this scope
    return true
})

// Dispatch (auto-bubbles to parent)
child.Dispatch(action.NewAction(action.ActionClick).WithTargetID(actionID))
```

## Composite Actions

```go
// Batch (concurrent)
batch := action.NewBatchAction(
    action.NewAction(action.ActionFetch),
    action.NewAction(action.ActionLoad),
    action.NewAction(action.ActionRender),
)
results := batch.Execute(dispatcher)

// Sequential
seq := action.NewSequentialActions(
    action.NewAction(action.ActionLoad),
    action.NewAction(action.ActionProcess),
    action.NewAction(action.ActionSave),
)
for _, a := range seq.actions {
    dispatcher.Dispatch(a)
}

// Throttled action
throttled := action.NewThrottledAction(myAction, 16 /* min interval ms */)
if throttled.IsNowAllowed(now) {
    throttled.UpdateLastExecution(now)
    dispatcher.Dispatch(throttled.action)
}
```

## Error Handling

```go
// Create error
err := action.NewError(
    action.ErrTargetDisabled,
    "component is disabled",
    myAction,
).WithTarget("button1").WithComponentType("Button")

// Use with validation
if s, validateErr := action.ValidateStringPayload(myAction); validateErr != nil {
    fmt.Println(validateErr.Error())
} else {
    fmt.Println("Payload:", s)
}

// Check error types
if action.IsErrTarget(err) {
    // Handle target error
}
if action.IsErrPayload(err) {
    // Handle payload error
}
```

## Backward Compatibility Adapters

```go
// Adapter for old Updater interface
updaterAdapter := action.NewUpdaterAdapter(oldComponent, "comp1")
dispatcher.Register(updaterAdapter)

// Adapter for old EventHandler
eventAdapter := action.NewEventHandlerAdapter(oldHandler, "comp2")
dispatcher.Register(eventAdapter)

// Adapter for func-based handlers
funcAdapter := action.NewFuncAdapter("comp3", func(a *action.Action) bool {
    // Handle action
    return true
})
dispatcher.Register(funcAdapter)

// Auto-wrap component
target := action.WrapComponent(someComponent, "comp4")
dispatcher.Register(target)
```

## Migration from framework/action

1. **Update imports**: Replace `framework/action` with `runtime/action`
2. **Update API**:
   - `action.TargetID` (uint64) → Use `action.WithTarget(string)` for semantic IDs
   - Target still supports uint64 via `WithTargetID(uint64)`
3. **Use adapters**: Wrap old components with adapters during transition
4. **Apply middleware**: Enable desired middleware types

```go
// Before (framework/action)
func MyComponent() Target {
    return &MyComponentImpl{}
}

// After (runtime/action)
func MyComponent() action.Target {
    // Direct implementation
    return &MyComponentImpl{}
}

// Or use adapter for old interface
func MyComponent() action.Target {
    return action.NewUpdaterAdapter(oldComponent, "my-comp")
}
```

## File Structure

| File | Source | Description |
|------|--------|-------------|
| `action.go` | Merged | Action types, constants (50+), helper methods |
| `target.go` | Merged | Target interface, capabilities, adapters |
| `payload.go` | Framework | Immutable payload structures |
| `router.go` | Framework | Three-phase router with middleware |
| `dispatcher.go` | Runtime | Simple dispatcher with global handlers |
| `processor.go` | Framework | Input to action conversion |
| `keymap.go` | Framework | Keyboard mappings (global + context) |
| `scope.go` | Framework | Scope-based dispatchers |
| `middleware.go` | Framework + Enhanced | 8 middleware types |
| `errors.go` | Runtime | Structured error handling |
| `adapters.go` | Framework | Migration adapters |
| `composite.go` | Runtime | Composite actions |

## Pure Go Constraint

This package must remain pure Go without dependencies on:
- Bubble Tea
- DSL parsers
- Component implementations
- UI libraries (lipgloss, etc.)

## Best Practices

1. **Use semantic Actions**: `ActionNavigateNext` instead of `ActionTabKey`
2. **Keep Payloads simple**: Use primitive types (string, int, struct)
3. **Use middleware**: Enable logging/metrics in production
4. **Implement capabilities**: Add Selectable, Scrollable, etc. as needed
5. **Handle errors appropriately**: Use structured error types
6. **Use scope dispatchers**: Isolate component actions

## Common Patterns

### Form Validation

```go
// Validation middleware
validator := action.NewValidationMiddleware()
validator.RegisterValidator(action.ActionSubmit, func(act *action.Action) error {
    payload, ok := act.Payload.(*action.SubmitPayload)
    if !ok {
        return errors.New("invalid payload")
    }
    if validateForm(payload.Values) {
        return nil
    }
    return errors.New("form validation failed")
})
```

### Focus Management

```go
// Register focusable component
if focusable, ok := component.(action.FocusableTarget); ok {
    dispatcher.Register(focusable)
}

// Handle focus navigation
router.AddCaptureHandler(NewFocusNavigator(), 50)
```

### Batch Updates

```go
// Batch multiple updates
batch := action.NewBatchAction()
for _, item := range items {
    batch.AddAction(action.NewAction(action.ActionUpdate).WithPayload(item))
}
results := batch.Execute(dispatcher)
```

### Context-Sensitive KeyBindings

```go
km := action.NewKeyMap()
km.BindWithContext("editor", "ctrl+s", saveAction)
km.BindWithContext("editor", "ctrl+f", findAction)

km.PushContext("editor") // Enter editor context
```

## Performance Notes

- **Target Lookup**: String IDs are hashed to uint64 for O(1) lookup
- **Middleware**: Minimal overhead when disabled
- **Caching**: Optional caching middleware reduces redundant processing
- **Profiling**: Built-in profiling to identify hotspots
- **Throttling**: Prevents action spam at configurable intervals

## See Also

- `docs/plan/action/MERGE_EXECUTION_PLAN.md` - Detailed merge execution plan
- `runtime/action/composite.go` - Composite actions for batch/parallel execution
- `runtime/event` - Low-level event system
