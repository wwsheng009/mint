# Components Documentation

This directory contains documentation for Mint TUI components.

## Available Components

### Layout Components
- **ScrollView** - Scrollable container with automatic scrollbars
  - [Scroll View Component](../layout/SCROLL_VIEW_COMPONENT.md) - ScrollView documentation

- **Virtual Scroll** - Virtual scrolling for large lists (performance optimization)
  - [Virtual List Component](VIRTUAL_LIST_COMPONENT.md) - Virtual scrolling component

- **Tabs** - Tab bar component for switching between panels
  - [Tabs Component](TABS_COMPONENT.md) - Tabs documentation

### Navigation Components
- **TreeView** - Hierarchical tree display with navigation
  - [TreeView Navigation Working](../examples/inspector/TREEVIEW_NAVIGATION_WORKING.md) - Navigation verification
  - [TreeView Navigation](TREEVIEW_NAVIGATION.md) - Navigation implementation

## Component Features

### Common Patterns
All components follow the **Builder pattern** for fluent API:

```go
component := ComponentBuilder().
    Property(value).
    AnotherProperty(value).
    Build()
```

### Layout Integration
Components integrate with the layout engine:
- Support for flex layouts (HStack, VStack)
- Automatic size calculation
- Proper hit testing for mouse events
- Focus management

### Event Handling
Components can handle keyboard and mouse events:
- Implement the `Component` interface
- Handle events through `HandleEvent(Event) bool`
- Return true to stop event propagation

## See Also

- [Layout Documentation](../layout/) - Layout system documentation
- [Component Summary](COMPONENTS_SUMMARY.md) - Overview of all components
- [Architecture Documentation](../architecture/) - System architecture
