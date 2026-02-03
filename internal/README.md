# Internal Packages

This directory contains internal implementation packages that are not part of the public API.

## Package Structure

### log/
**Purpose**: Structured logging for the mint framework

**Dependencies**: None (stdlib only)

**Exports**:
- `Logger` - Thread-safe structured logger
- Global loggers: `FocusLogger`, `ReconcilerLogger`, `RenderLogger`, `KeyLogger`, `UILogger`, `ButtonLogger`

**Usage**:
```go
log.FocusLogger.Debug("message %s", arg)
log.ReconcilerLogger.Debug("message %s", arg)
```

**Environment Variables**:
- `TUI_DEBUG` - Enable all debug logging
- `TUI_DEBUG_FOCUS` - Enable focus-related logging
- `TUI_DEBUG_RECONCILER` - Enable reconciler logging
- `TUI_DEBUG_RENDER` - Enable render logging
- `TUI_DEBUG_KEYS` - Enable key event logging
- `TUI_DEBUG_UI` - Enable UI component logging

---

### reconciler/
**Purpose**: Fiber reconciliation engine for incremental UI updates

**Dependencies**:
- `github.com/wwsheng009/mint/framework` - Framework integration
- `github.com/wwsheng009/mint/runtime` - Runtime types
- `github.com/wwsheng009/mint/internal/state` - Component instance management
- `github.com/wwsheng009/mint/internal/log` - Logging

**Exports**:
- `Reconciler` - Main reconciler struct
- `Fiber` - Fiber tree node (defined in runtime/ui, used here)
- `TreeWalker` - Fiber tree traversal abstraction
- `Lane`, `LaneSyncLane` - Update priority management

**Key Files**:
- `reconciler.go` - Main reconciler logic
- `tree_walker.go` - Fiber tree traversal utilities
- `begin_work.go` - BeginWork phase implementation
- `complete_work.go` - CompleteWork phase implementation
- `diff.go` - Diff algorithm for VNode comparison

---

### render/
**Purpose**: Declarative VNode to framework bridge

**Dependencies**:
- `github.com/wwsheng009/mint/runtime` - Runtime types
- `github.com/wwsheng009/mint/internal/reconciler` - Reconciler integration
- `github.com/wwsheng009/mint/components` - Component implementations

**Exports**:
- `DeclarativeNode` - VNode-based render node

---

### state/
**Purpose**: Component instance and state management

**Dependencies**:
- `github.com/wwsheng009/mint/runtime/ui` - VNode interfaces

**Sub-packages**:
- `instance_manager.go` - Component lifecycle management
- `interaction_state.go` - Focus and hover state tracking
- `key_validator.go` - Key validation for VNode trees

---

### scheduler/
**Purpose**: Update scheduling and priority lanes

**Dependencies**:
- Minimal external dependencies

---

## Dependency Rules

1. **No circular dependencies** - Internal packages should not import each other in cycles
2. **runtime/ is foundation** - Can be used by all packages
3. **internal/ packages** - Cannot be imported from outside the project
4. **Framework boundary** - `framework/` provides the public API

## Package Interaction Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                      │
│  (examples/, app/, cmd/)                                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Public API Layer                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐   │
│  │ framework/ │  │  runtime/  │  │   components/     │   │
│  └────────────┘  └────────────┘  └────────────────────┘   │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                  Internal Implementation                     │
│  ┌────────────┐  ┌──────────┐  ┌──────┐  ┌────────────┐   │
│  │reconciler/ │  │  state/  │  │ log/ │  │  render/   │   │
│  └────────────┘  └──────────┘  └──────┘  └────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Import Guidelines

When adding imports to internal packages:

1. **Prefer runtime/ over internal/** - If functionality can be moved to runtime/, do so
2. **Keep interfaces in runtime/** - Public interfaces belong in runtime/ui
3. **Implementation in internal/** - Complex internal logic goes here
4. **Document dependencies** - Add comments explaining why a dependency exists
