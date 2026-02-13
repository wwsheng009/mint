# Inspector Integration Documentation

This directory contains documents about the integration of Inspector with the framework, render layer, and hook system.

## Integration Status

### [INSPECTOR_FIX_COMPLETE.md](INSPECTOR_FIX_COMPLETE.md)
Summary of completed fixes and current status of the Inspector integration.

### [INSPECTOR_FINAL_STATUS.md](INSPECTOR_FINAL_STATUS.md)
Final status report with all issues resolved and system working correctly.

### [INSPECTOR_FIXES_SUMMARY.md](INSPECTOR_FIXES_SUMMARY.md)
Comprehensive summary of all fixes applied to the Inspector system.

## Integration Architecture

### [CORRECT_ARCHITECTURE.md](CORRECT_ARCHITECTURE.md)
Document describing the correct architectural approach for Inspector integration.

**Key Principles:**
- Application layer: Business logic only
- Framework layer: Cross-cutting integration
- Render layer: Automatic transformation
- Clear separation of concerns

### [RENDER_HOOK_DESIGN.md](RENDER_HOOK_DESIGN.md)
Design document for the hook-based rendering system.

**Design Goals:**
- Automatic Inspector injection
- No manual Fragment wrapping
- Centralized layer management
- Avoid import cycles

## Layer System Integration

### [LAYER_SYSTEM_ARCHITECTURE_ANALYSIS.md](LAYER_SYSTEM_ARCHITECTURE_ANALYSIS.md)
Analysis of the layer system architecture and how Inspector integrates with it.

**Layer Types:**
- LayerBase (0): Normal application content
- LayerModal (1): Modal dialogs
- LayerTooltip (2): Tooltips and popups
- LayerInspector (4): Inspector overlay

**Rendering Flow:**
```
PipelineRenderer.Render()
    ↓
ApplyVNodeHooks() → wraps content in Fragment
    ↓
LayerManager.CollectLayers() → separates by layer
    ↓
LayerManager.Render() → renders each layer to buffer
    ↓
Composite → combine all layers
```

### [LAYER_BUG_ANALYSIS.md](LAYER_BUG_ANALYSIS.md)
Analysis of layer-related bugs and their fixes.

### [LAYER_ARCHITECTURE_REFACTOR.md](LAYER_ARCHITECTURE_REFACTOR.md)
Document describing the refactoring of layer architecture.

### [INSPECTOR_LAYER_FIX_IMPLEMENTATION.md](INSPECTOR_LAYER_FIX_IMPLEMENTATION.md)
Implementation details of the layer fix for Inspector.

## Rendering Integration

### [RENDER_PATH_COMPARISON.md](RENDER_PATH_COMPARISON.md)
Comparison of different rendering approaches and why the hook-based approach was chosen.

**Approaches Considered:**
1. Manual Fragment wrapping (rejected)
2. Direct layer manipulation (rejected)
3. Hook-based automatic injection (chosen)

### [INSPECTOR_RENDERING_FINAL_REPORT.md](INSPECTOR_RENDERING_FINAL_REPORT.md)
Final report on Inspector rendering integration.

## Demo Integration

### [DEMO_STARTUP_FIX.md](DEMO_STARTUP_FIX.md)
Fixes applied to demo2 startup process.

### [DEMO_VERIFICATION.md](DEMO_VERIFICATION.md)
Verification of demo2 functionality after integration.

### [INSPECTOR_OVERLAY_BUG_FIX.md](INSPECTOR_OVERLAY_BUG_FIX.md)
Fix for Inspector overlay bug in demo.

### [INSPECTOR_COMPONENTID_SOLUTION.md](INSPECTOR_COMPONENTID_SOLUTION.md)
Solution for ComponentID-related issues in Inspector.

## Proposal Documents

### [LAYER_ARCHITECTURE_PROPOSAL.go](LAYER_ARCHITECTURE_PROPOSAL.go)
Initial proposal for layer architecture changes (saved as .go to avoid conflicts).

## Integration Timeline

### Phase 1: Initial Integration (Completed)
- [x] Basic Inspector overlay rendering
- [x] Manual Fragment wrapping in demo
- [x] SetLayer calls in application code

### Phase 2: Bug Discovery (Completed)
- [x] Identified SetProps/SetLayer bug
- [x] Found layer rendering issues
- [x] Discovered UniqueID collisions
- [x] Identified architectural problems

### Phase 3: Architecture Refactoring (Completed)
- [x] Created hook system
- [x] Implemented Inspector hook
- [x] Framework integration via reflection
- [x] Removed manual Fragment wrapping
- [x] Centralized SetLayer call

### Phase 4: Bug Fixes (Completed)
- [x] Fixed SetProps/SetLayer ordering
- [x] Implemented pointer-based UniqueIDs
- [x] Fixed Inspector positioning
- [x] Implemented flex layout for TreeView

### Phase 5: Testing & Verification (Completed)
- [x] Unit tests for all components
- [x] Integration tests for full pipeline
- [x] Manual testing with demo2
- [x] Documentation complete

## Current Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Application Code (demo2/main.go)                             │
│                                                              │
│ func AppRender() ui.VNode {                                  │
│     return buildApp()  // Just app content!                 │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Framework (framework/app.go)                                 │
│                                                              │
│ fwApp.SetRoot(declarativeRoot)                               │
│ fwApp.SetInspector(inspector)  ← Registers hook             │
│                                                              │
│ func registerInspectorHook() {                               │
│     hookManager := getHookManager()  ← Reflection!          │
│     inspector.RegisterWithHookManager(hookManager)          │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Hook System (internal/inspector/hook.go)                     │
│                                                              │
│ func CreateInspectorHook() VNodeHook {                       │
│     return func(vnode) VNode {                               │
│         if !inspector.IsVisible() {                          │
│             return vnode  // No-op                          │
│         }                                                    │
│         overlay := inspector.RenderContent()                 │
│         overlay.SetProps({x, y, width, height})              │
│         overlay.SetLayer(LayerInspector)  ← ONLY place!     │
│         return Fragment(vnode, overlay)                      │
│     }                                                        │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ PipelineRenderer (internal/render/pipeline_renderer.go)      │
│                                                              │
│ func Render(vnode, x, y, buffer) {                           │
│     vnode = hooks.ApplyVNodeHooks(vnode)  ← Apply hooks     │
│     // ... rest of rendering                                 │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ LayerManager (runtime/layer/manager.go)                      │
│                                                              │
│ func CollectLayers(vnode) {                                  │
│     // Separate Fragment children by layer                   │
│     baseLayer = children with LayerBase                      │
│     inspectorLayer = children with LayerInspector            │
│ }                                                            │
│                                                              │
│ func Render() {                                              │
│     renderToBuffer(baseLayer)                                │
│     renderToBuffer(inspectorLayer)  ← Overlay               │
│ }                                                            │
└─────────────────────────────────────────────────────────────┘
```

## Key Integration Points

### 1. Framework → Inspector
**Method:** `App.SetInspector(inspector)`
- Takes `interface{}` to avoid import cycle
- Calls `registerInspectorHook()`
- Uses reflection to get HookManager

### 2. Inspector → Hook System
**Method:** `StandaloneInspector.RegisterWithHookManager(hookManager)`
- Implements `HookRegistrar` interface
- Creates Inspector hook
- Registers with HookManager

### 3. Hook → PipelineRenderer
**Method:** `PipelineRenderer.Render()`
- Applies VNodeHooks at start of render
- Transforms app content → Fragment(app, Inspector)
- Sets LayerInspector on overlay

### 4. PipelineRenderer → LayerManager
**Method:** `RenderingPipeline.renderNode()`
- Detects Fragment children
- Separates by layer property
- Calls LayerManager for multi-layer rendering

## Testing Integration

### Unit Tests
```bash
# Test hook registration
go test -v ./internal/inspector -run TestHookRegistrar

# Test hook application
go test -v ./runtime/render -run TestVNodeHook

# Test framework integration
go test -v ./internal/inspector -run TestFrameworkIntegration
```

### Integration Tests
```bash
# Test full pipeline
go test -v ./internal/render -run TestPipelineRenderer

# Test Inspector overlay
go test -v ./internal/inspector -run TestInspectorOverlay
```

### Manual Testing
```bash
# Run demo with Inspector
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
TUI_INSPECTOR=true go run main.go
```

## Verification Checklist

- [x] Inspector overlay displays when visible
- [x] App content renders correctly
- [x] TreeView expand/collapse works
- [x] UniqueIDs are unique
- [x] No manual Fragment wrapping needed
- [x] SetLayer called in only one place
- [x] No import cycles
- [x] Framework doesn't import internal/render
- [x] Application code is clean
- [x] All tests pass

## Environment Variables

### TUI_INSPECTOR
Auto-show Inspector on startup
```bash
TUI_INSPECTOR=true go run main.go
```

### TUI_DEBUG_INSPECTOR
Enable verbose Inspector logging
```bash
TUI_DEBUG_INSPECTOR=true go run main.go
```

### TUI_DEBUG_UI
Enable framework-level debug logging
```bash
TUI_DEBUG_UI=true go run main.go
```

## Related Documentation

- [Architecture Overview](../architecture/)
- [Implementation Details](../implementation/)
- [Investigation Reports](../investigation/)
- [Hook System](../../render/hook/README.md)
