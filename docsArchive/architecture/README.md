# Inspector Architecture Documentation

This directory contains documents describing the architectural design and analysis of the Inspector system.

## Documents

### [INSPECTOR_RENDERING_FLOW_ANALYSIS.md](INSPECTOR_RENDERING_FLOW_ANALYSIS.md)
Detailed analysis of how the Inspector rendering flow works, including the interaction between application code, framework, and render layer.

**Key Topics:**
- Rendering pipeline flow
- Layer system integration
- VNode transformation
- Fragment wrapping mechanism

### [INSPECTOR_LAYER_SOLUTION_ANALYSIS.md](INSPECTOR_LAYER_SOLUTION_ANALYSIS.md)
Analysis of the multi-layer rendering solution and how Inspector integrates with the layer system.

**Key Topics:**
- Layer architecture (Base, Modal, Inspector, Tooltip)
- LayerManager responsibilities
- Multi-layer rendering pipeline
- Layer-based positioning and sizing

### [FRAMEWORK_LAYER_MANAGEMENT.md](FRAMEWORK_LAYER_MANAGEMENT.md)
How the framework manages layers and integrates Inspector without requiring application code to handle layer management.

**Key Topics:**
- Framework-level layer management
- Hook-based automatic injection
- Separation of concerns
- Import cycle prevention

### [INSPECTOR_ARCHITECTURE_ISSUES.md](INSPECTOR_ARCHITECTURE_ISSUES.md)
Historical document describing architectural issues encountered during development and their solutions.

**Key Topics:**
- Initial architecture problems
- Refactoring to hook-based approach
- Framework integration challenges

## Related Documentation

- [Implementation Details](../implementation/)
- [Investigation Reports](../investigation/)
- [Integration Status](../integration/)
- [Hook System](../../render/hook/README.md)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  (demo2, user applications)                                  │
│  - Returns: appContent VNode only                            │
│  - No knowledge of Inspector or Layers                       │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Framework Layer                          │
│  (App, SetInspector, HookRegistrar)                          │
│  - Registers Inspector hook via reflection                  │
│  - Manages layer lifecycle                                  │
│  - No import of internal/render packages                    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Render Layer                             │
│  (PipelineRenderer, HookManager)                             │
│  - Applies VNodeHook transformations                         │
│  - Wraps appContent with Inspector overlay                   │
│  - Sets LayerInspector on overlay                           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Layer System                             │
│  (LayerManager, Multi-layer rendering)                       │
│  - Renders Base layer (appContent)                          │
│  - Renders Inspector layer (overlay)                        │
│  - Handles positioning, sizing, and compositing             │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Hook-Based Injection
**Decision:** Use hooks instead of manual Fragment wrapping in application code.

**Rationale:**
- Separation of concerns
- Automatic injection
- Centralized logic
- Easier to maintain

### 2. Reflection-Based Integration
**Decision:** Framework uses reflection to access HookManager instead of importing internal packages.

**Rationale:**
- Avoid import cycles
- Clean package boundaries
- Framework remains application-layer focused

### 3. Single SetLayer Call Point
**Decision:** The Inspector hook is the ONLY place that calls `SetLayer(LayerInspector)`.

**Rationale:**
- Centralized layer management
- No scattered layer logic
- Easier to debug and maintain

### 4. LIFO Hook Order
**Decision:** Hooks are applied in Last-In-First-Out order.

**Rationale:**
- Allows hook composition and wrapping
- Later hooks wrap earlier ones
- More intuitive for middleware-style patterns
