# Inspector Documentation Index

This directory contains comprehensive documentation about the Mint TUI Inspector implementation, from initial investigations to final hook-based integration.

## Quick Links

### 🚀 Getting Started
- [Hook-Based Integration](../render/hook/README.md) - How the hook system automatically injects Inspector
- [Inspector Usage Guide](usage/) - How to use the Inspector in your application

### 🏗️ Architecture
- [Rendering Flow Analysis](architecture/INSPECTOR_RENDERING_FLOW_ANALYSIS.md) - How Inspector rendering works
- [Layer System Integration](architecture/INSPECTOR_LAYER_SOLUTION_ANALYSIS.md) - Multi-layer rendering architecture
- [Framework Layer Management](architecture/FRAMEWORK_LAYER_MANAGEMENT.md) - Layer management in the framework

### 🔍 Investigation & Analysis
- [Initial Investigation](investigation/INSPECTOR_INVESTIGATION_COMPLETE.md) - Complete investigation of Inspector issues
- [Final Investigation Report](investigation/INSPECTOR_FINAL_INVESTIGATION_REPORT.md) - Detailed findings
- [TreeView Issues](investigation/INSPECTOR_TREEVIEW_OVERFLOW.md) - TreeView overflow and scrolling
- [UniqueID Problem](investigation/INSPECTOR_UNIQUEID_FINAL_SOLUTION.md) - UniqueID collision issues and solutions

### 🔧 Implementation Details
- [Flex Layout Implementation](implementation/INSPECTOR_FLEX_LAYOUT_IMPLEMENTATION.md) - Flex layout for TreeView
- [AutoSize Implementation](implementation/INSPECTOR_FLEX_AUTOSIZE_IMPLEMENTATION.md) - Auto-sizing components
- [Pointer ID Fix](implementation/INSPECTOR_POINTER_ID_FIX.md) - Pointer-based UniqueID solution
- [Border Fix](implementation/INSPECTOR_HARDCODED_BORDER_FIX.md) - Border rendering fixes

### 🎯 Key Solutions
- [UniqueID Final Solution](implementation/INSPECTOR_UNIQUEID_FINAL_SOLUTION.md) - How UniqueID collisions were resolved
- [Position Fix](implementation/INSPECTOR_POSITION_FIX.md) - Inspector positioning solution
- [SetLayer Bug Fix](implementation/INSPECTOR_SETLAYER_BUG_FIX.md) - SetProps/SetLayer order fix

## Project Structure

```
docs/inspector/
├── README.md (this file)
├── architecture/        # Architecture and design documents
├── investigation/       # Investigation and analysis reports
├── implementation/      # Implementation details and fixes
└── integration/         # Integration with framework and hook system
```

## Development Timeline

1. **Initial Implementation** - Basic Inspector with manual Fragment wrapping
2. **Investigation Phase** - Identified UniqueID collisions, layout issues, rendering problems
3. **Architecture Refactoring** - Moved to hook-based automatic injection
4. **Flex Layout Integration** - Implemented proper flex layout for TreeView
5. **Final Polish** - Position fixes, SetLayer bug fix, comprehensive testing

## Key Files

- `internal/inspector/standalone_inspector.go` - Main Inspector implementation
- `internal/inspector/hook.go` - Hook-based automatic injection
- `internal/inspector/tree_view.go` - TreeView component with flex layout
- `framework/inspector_integration.go` - Framework integration via reflection
- `runtime/render/hook.go` - Core hook system

## Environment Variables

- `TUI_INSPECTOR=true` - Auto-show Inspector on startup
- `TUI_DEBUG_INSPECTOR=true` - Enable verbose Inspector logging
- `TUI_DEBUG_UI=true` - Enable framework-level debug logging

## Testing

```bash
# Run all Inspector tests
go test -v ./internal/inspector

# Run hook system tests
go test -v ./runtime/render -run Hook

# Run integration tests
go test -v ./internal/render -run Pipeline
```

## Related Documentation

- [Hook System Documentation](../render/hook/README.md)
- [Framework Architecture](../../framework/README.md)
- [Rendering Pipeline](../../internal/render/README.md)
