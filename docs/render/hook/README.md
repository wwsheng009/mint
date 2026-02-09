# Hook System Documentation

## Overview

The Hook System is a VNode transformation mechanism that allows automatic modification of the render tree before rendering. It provides a clean way to inject cross-cutting concerns (like Inspector overlays) without requiring application code to manually handle Fragment wrapping or Layer management.

## Architecture

```
Application Layer
    ↓ Returns: appContent VNode
Framework Layer (App.SetInspector)
    ↓ Registers: Inspector hook via HookRegistrar
Render Layer (PipelineRenderer.Render)
    ↓ Applies: VNodeHook transformation
Result: Fragment(appContent, InspectorOverlay)
    ↓
Multi-Layer Rendering (LayerManager)
    ↓ Renders: Base layer + Inspector layer
```

## Core Components

### 1. Hook System (`runtime/render/hook.go`)

**VNodeHook Type:**
```go
type VNodeHook func(rtui.VNode) rtui.VNode
```

A hook is a pure function that transforms a VNode into another VNode. It can:
- Wrap the original VNode in a Fragment
- Modify props, styles, or structure
- Return the VNode unchanged (no-op)
- Return nil to suppress rendering

**HookManager:**
```go
type HookManager struct {
    vnodeHooks []VNodeHook
}

// Apply hooks in LIFO order (last registered runs first)
func (hm *HookManager) ApplyVNodeHooks(vnode rtui.VNode) rtui.VNode
```

### 2. Inspector Hook (`internal/inspector/hook.go`)

**CreateInspectorHook:**
```go
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook
```

The Inspector hook:
1. Checks if Inspector is visible
2. Gets Inspector UI content via `RenderContent()`
3. Sets positioning props (x, y, width, height)
4. Sets `LayerInspector` on overlay
5. Wraps original VNode and Inspector in Fragment

**Key Design Decision:**
- The hook is the **ONLY place** that calls `SetLayer(LayerInspector)`
- Application code and Inspector itself don't need to know about Layers
- Clean separation of concerns

### 3. Framework Integration (`framework/inspector_integration.go`)

**Challenge:** Framework cannot import `internal/render` (import cycle)

**Solution:** Reflection-based access to HookManager

```go
func (a *App) getHookManager() interface{} {
    rootValue := reflect.ValueOf(a.root)
    getHooksMethod := rootValue.MethodByName("GetHooks")

    if !getHooksMethod.IsValid() {
        return nil
    }

    results := getHooksMethod.Call(nil)
    return results[0].Interface()
}
```

**HookRegistrar Interface:**
```go
type HookRegistrar interface {
    RegisterWithHookManager(interface{})
}
```

Inspector implements this interface, allowing framework to register it without direct import.

### 4. Pipeline Integration (`internal/render/pipeline_renderer.go`)

```go
func (r *PipelineRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) error {
    // Apply VNode hooks BEFORE rendering
    vnode = r.hooks.ApplyVNodeHooks(vnode)

    // ... rest of rendering pipeline
}
```

Hooks are applied at the start of every render cycle.

## LIFO Order

Hooks are applied in **reverse order** (Last In, First Out):

```go
hooks.RegisterVNodeHook(hook1)
hooks.RegisterVNodeHook(hook2)
hooks.RegisterVNodeHook(hook3)

// Application order: hook3 → hook2 → hook1
```

**Why LIFO?** Allows later hooks to wrap earlier ones:

```go
// Hook 1: Wrap with border
func(vnode) {
    return ui.Bordered().Child(vnode).Build()
}

// Hook 2: Wrap with Inspector (runs FIRST due to LIFO)
func(vnode) {
    inspector := inspector.RenderContent()
    return rtui.Fragment(vnode, inspector)
}

// Result: Fragment(Bordered(app), Inspector)
// Border is applied to app content, NOT to Inspector
```

## Example: Inspector Integration

### Before (Manual Fragment Wrapping)

```go
// Application code must manually handle Inspector
func AppRender() ui.VNode {
    appContent := buildApp()

    if globalInspector.IsVisible() {
        inspectorOverlay := globalInspector.RenderOverlay()
        inspectorOverlay.SetLayer(ui.LayerInspector)
        inspectorOverlay.SetProps(ui.Props{
            "x": 80, "y": 5, "width": 80, "height": 25,
        })
        return ui.Fragment(appContent, inspectorOverlay)
    }

    return appContent
}
```

**Problems:**
- ❌ Application code knows about Inspector
- ❌ Application code knows about Layers
- ❌ Application code must remember to wrap
- ❌ Easy to forget or make mistakes

### After (Automatic Hook Injection)

```go
// Application code returns app content only
func AppRender() ui.VNode {
    return buildApp()
}

// Framework registers Inspector hook
fwApp.SetInspector(globalInspector)

// Hook automatically injects Inspector when visible
```

**Benefits:**
- ✅ Application code is clean and focused
- ✅ Inspector injection is automatic
- ✅ No manual Fragment wrapping
- ✅ No manual Layer management
- ✅ Centralized injection logic

## Advantages

### 1. Separation of Concerns
- Application layer: Business logic only
- Framework layer: Cross-cutting integration
- Render layer: Automatic transformation

### 2. No Import Cycles
- Framework doesn't need to import internal packages
- Reflection provides dynamic method invocation
- Interface-based registration avoids coupling

### 3. Composable
- Multiple hooks can be registered
- LIFO order allows wrapping
- Each hook is independent and testable

### 4. Transparent
- Hooks are invisible to application code
- Automatic application on every render
- No developer intervention needed

### 5. Testable
- Each hook is a pure function
- Easy to unit test in isolation
- Integration tests verify full pipeline

## Disadvantages

### 1. Hidden Complexity
- Hook execution is not obvious from reading app code
- Debugging requires understanding hook system
- May surprise developers unfamiliar with the pattern

### 2. Performance Overhead
- Every render applies all hooks
- Hook creation (if using closures) has allocation cost
- LIFO traversal adds function call overhead

**Mitigation:**
- Keep hooks fast and lightweight
- Avoid expensive operations in hot path
- Cache hook results where appropriate

### 3. Debugging Challenges
- Stack traces show hook wrapper functions
- Harder to trace VNode transformations
- Requires specialized debugging tools

**Mitigation:**
- Add verbose logging in development
- Use `TUI_DEBUG_HOOK` environment variable
- Provide hook inspection tools

### 4. Reflection Dependency
- Framework integration relies on reflection
- No compile-time type safety for GetHooks()
- Method name is string (typos possible)

**Mitigation:**
- Add debug logging to verify reflection works
- Unit tests cover reflection paths
- Consider alternative if performance critical

## Best Practices

### 1. Keep Hooks Pure
```go
// Good: Pure transformation
func(vnode rtui.VNode) rtui.VNode {
    return ui.Bordered().Child(vnode).Build()
}

// Bad: Side effects
func(vnode rtui.VNode) rtui.VNode {
    globalState.modified = true
    return vnode
}
```

### 2. Check Conditions Early
```go
// Good: Early return
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        if !inspector.IsVisible() {
            return vnode  // No-op when not visible
        }
        // ... expensive inspector logic
    }
}

// Bad: Always does work
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        overlay := inspector.RenderContent()  // Expensive
        if !inspector.IsVisible() {
            return vnode
        }
        // ...
    }
}
```

### 3. Use LIFO Order Intentionally
```go
// Register hooks in reverse order of desired application
hooks.RegisterVNodeHook(loggingHook)        // Runs LAST (wraps everything)
hooks.RegisterVNodeHook(inspectorHook)      // Runs MIDDLE
hooks.RegisterVNodeHook(perfHook)           // Runs FIRST (closest to vnode)
```

### 4. Document Hook Behavior
```go
// CreateInspectorHook creates a hook that injects Inspector overlay.
//
// The hook:
// 1. Checks Inspector.IsVisible()
// 2. Wraps vnode in Fragment with Inspector overlay
// 3. Sets LayerInspector on overlay (ONLY place SetLayer is called)
// 4. Positions Inspector at configured (x,y) coordinates
//
// IMPORTANT: This is the ONLY place that should call SetLayer(LayerInspector).
// Application code and Inspector itself should not set LayerInspector.
```

### 5. Test Hooks in Isolation
```go
func TestInspectorHook(t *testing.T) {
    inspector := NewStandaloneInspector()
    inspector.Enable()
    inspector.ToggleVisibility()

    hook := CreateInspectorHook(inspector)
    testVNode := ui.Text("test")

    result := hook(testVNode)

    fragment, ok := result.(*ui.FragmentVNode)
    if !ok {
        t.Fatal("Expected Fragment")
    }

    // Verify structure
    if len(fragment.Children()) != 2 {
        t.Errorf("Expected 2 children")
    }
}
```

## Environment Variables

### TUI_DEBUG_UI
Enable verbose framework-level logging:
```
export TUI_DEBUG_UI=true
```

Shows:
- Hook registration status
- HookManager retrieval via reflection
- Type assertion results

### TUI_INSPECTOR_VERBOSE
Enable Inspector hook logging:
```
export TUI_INSPECTOR_VERBOSE=true
```

Shows:
- When hook is applied
- Inspector visibility status
- Overlay positioning details

## Debugging

### Enable All Debug Output
```bash
TUI_DEBUG_UI=true TUI_INSPECTOR_VERBOSE=true go run main.go
```

### Check Hook Registration
Look for:
```
[APP] getHookManager: root type=*render.DeclarativeNode
[APP] getHookManager: found GetHooks() method via reflection
[APP] getHookManager: got hooks type=*render.HookManager
[APP] ✅ Inspector hook registered via HookRegistrar interface
```

### Verify Hook Execution
Look for:
```
[InspectorHook] Injecting Inspector overlay
[InspectorHook] Inspector overlay: layer=4, pos=(0,0), size=80x25
```

### Common Issues

**Issue:** Hook not registered
```
[APP] Cannot register Inspector hook: no HookManager found
```
**Cause:** SetInspector() called before SetRoot()
**Fix:** Call SetRoot() first, then SetInspector()

**Issue:** Hook executes but Inspector not visible
```
[InspectorHook] Inspector not visible, skipping injection
```
**Cause:** Inspector.IsVisible() returns false
**Fix:** Call inspector.ToggleVisibility() or press F12

**Issue:** Reflection fails
```
[APP] getHookManager: root does not have GetHooks() method
```
**Cause:** Root node doesn't implement GetHooks()
**Fix:** Ensure root is a DeclarativeNode or implements interface

## Testing

### Unit Tests
```bash
# Hook system tests
go test -v ./runtime/render -run TestHook

# Inspector hook tests
go test -v ./internal/inspector -run TestHook

# Framework integration tests
go test -v ./internal/inspector -run TestFramework
```

### Integration Tests
```bash
# Full pipeline with hooks
go test -v ./internal/render -run TestPipelineRenderer

# Inspector overlay rendering
go test -v ./internal/inspector -run TestInspectorOverlay
```

## Future Enhancements

### 1. Hook Metadata
```go
type HookMetadata struct {
    Name        string
    Priority    int  // Override LIFO order
    Enabled     bool
    Description string
}

func (hm *HookManager) RegisterVNodeHookWithMeta(hook VNodeHook, meta HookMetadata)
```

### 2. Conditional Hooks
```go
func (hm *HookManager) RegisterConditionalHook(
    condition func() bool,
    hook VNodeHook,
)
```

### 3. Hook Composition
```go
func CombineHooks(hooks ...VNodeHook) VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        for _, h := range hooks {
            vnode = h(vnode)
        }
        return vnode
    }
}
```

### 4. Middleware Pattern
```go
type HookMiddleware func(VNodeHook) VNodeHook

func LoggingMiddleware(next VNodeHook) VNodeHook {
    return func(vnode rtui.VNode) rtui.VNode {
        log.Printf("Before: %T", vnode)
        result := next(vnode)
        log.Printf("After: %T", result)
        return result
    }
}
```

## References

- **Implementation:** `runtime/render/hook.go`
- **Inspector Hook:** `internal/inspector/hook.go`
- **Framework Integration:** `framework/inspector_integration.go`
- **Pipeline Usage:** `internal/render/pipeline_renderer.go`
- **Tests:** `runtime/render/hook_test.go`, `internal/inspector/hook_test.go`

## Summary

The Hook System provides a clean, automatic way to inject cross-cutting concerns into the render tree. By using reflection for framework integration and LIFO ordering for composable transformations, it achieves separation of concerns without import cycles.

**Key Takeaways:**
- Hooks are pure VNode transformation functions
- LIFO order allows wrapping and composition
- Reflection enables framework integration without import cycles
- Inspector hook is the ONLY place that calls SetLayer(LayerInspector)
- Application code remains clean and focused on business logic
