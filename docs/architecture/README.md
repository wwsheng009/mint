# Architecture Documentation

This directory contains high-level architecture documentation for the Mint TUI framework.

## Documents

### Layer System
- [Layer System Architecture](LAYER_SYSTEM_ARCHITECTURE.md) - Overview of the layer system
  - Layer types (Normal, Modal, Overlay, Inspector)
  - Layer ordering and rendering
  - Event routing through layers
  - Integration with framework

- [Layer System Implementation Summary](LAYER_SYSTEM_IMPLEMENTATION_SUMMARY.md) - Implementation details
  - How layers are integrated
  - Rendering pipeline
  - Event handling
  - Performance considerations

### Agent System
- [Agents](AGENTS.md) - AI Agent integration for the framework
  - Agent types and capabilities
  - Integration points
  - Use cases

### Rendering Systems
- [Two Rendering Systems Explained](TWO_RENDERING_SYSTEMS_EXPLAINED.md) - Comparison of rendering approaches
  - Legacy vs new rendering
  - Migration path
  - Performance differences

## Related Documentation

- [Component Documentation](../components/) - Component architecture
- [Layout Documentation](../layout/) - Layout system architecture
- [Feature Documentation](../features/) - Feature implementation details

## Architecture Overview

The Mint framework follows a layered architecture:

```
┌─────────────────────────────────────┐
│         Application Layer           │
│  (User code, VNode tree)            │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Framework Layer             │
│  (App, Event routing, Themes)       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Runtime Layer               │
│  (Layout engine, Paint, Components) │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Platform Layer              │
│  (Input, Terminal abstraction)     │
└─────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Separation of Concerns**
   - Platform layer abstracts OS differences
   - Runtime provides cross-platform UI
   - Framework adds app-level features

2. **Event-Driven**
   - All input generates events
   - Events flow through layers
   - Handlers can intercept or allow propagation

3. **Declarative UI**
   - Components describe UI as VNode tree
   - Framework handles rendering and updates
   - Automatic diffing and reconciliation

4. **Layered Rendering**
   - Multiple independent layers
   - Z-ordering for compositing
   - Inspector runs in overlay layer

## Data Flow

### Event Flow
```
Platform Input → Event Pump → Framework App → Layer Manager → Component Handlers
```

### Render Flow
```
Component Tree → Layout Engine → Paint Engine → Layer Compositor → Terminal
```

## See Also

- [Runtime Documentation](../../runtime/) - Runtime layer implementation
- [Framework Documentation](../../framework/) - Framework layer implementation
