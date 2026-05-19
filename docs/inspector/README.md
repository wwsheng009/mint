# Inspector Documentation Index

This directory contains comprehensive documentation about the Mint TUI Inspector implementation, from initial investigations to final hook-based integration.

## Quick Links

### 🚀 Getting Started
- [Hook-Based Integration](../render/hook/README.md) - How the hook system automatically injects Inspector
- [Inspector Usage Guide](QUICK_START.md) - How to use the Inspector in your application

### 🏗️ Architecture
- [Rendering Flow Analysis](/docsArchive/architecture/INSPECTOR_RENDERING_FLOW_ANALYSIS.md) - How Inspector rendering works (archived)
- [Layer System Integration](/docsArchive/architecture/INSPECTOR_LAYER_SOLUTION_ANALYSIS.md) - Multi-layer rendering architecture (archived)
- [Framework Layer Management](/docsArchive/architecture/FRAMEWORK_LAYER_MANAGEMENT.md) - Layer management in the framework (archived)

### 🔍 Investigation & Analysis
- [Initial Investigation](/docsArchive/investigation/INSPECTOR_INVESTIGATION_COMPLETE.md) - Complete investigation of Inspector issues (archived)
- [Final Investigation Report](/docsArchive/investigation/INSPECTOR_FINAL_INVESTIGATION_REPORT.md) - Detailed findings (archived)
- [TreeView Issues](/docsArchive/investigation/INSPECTOR_TREEVIEW_OVERFLOW.md) - TreeView overflow and scrolling (archived)
- [UniqueID Problem](/docsArchive/INSPECTOR_UNIQUEID_FINAL_SOLUTION.md) - UniqueID collision issues and solutions (archived)

### 🔧 Implementation Details
- Flex layout and AutoSize implementation records were archived under `../../docsArchive/cleanup-2026-05-19/docs/inspector/implementation/`.
- [Pointer ID Fix](/docsArchive/INSPECTOR_POINTER_ID_FIX.md) - Pointer-based UniqueID solution (archived)
- [Border Fix](/docsArchive/INSPECTOR_HARDCODED_BORDER_FIX.md) - Border rendering fixes (archived)

### 🎯 Key Solutions
- [UniqueID Final Solution](/docsArchive/INSPECTOR_UNIQUEID_FINAL_SOLUTION.md) - How UniqueID collisions were resolved (archived)
- [Position Fix](/docsArchive/INSPECTOR_POSITION_FIX.md) - Inspector positioning solution (archived)
- [SetLayer Bug Fix](/docsArchive/INSPECTOR_SETLAYER_BUG_FIX.md) - SetProps/SetLayer order fix (archived)

## Project Structure

```
docs/inspector/
├── README.md (this file)
├── architecture/        # Architecture and design documents (archived: /docsArchive/architecture/)
├── investigation/       # Investigation and analysis reports (archived: /docsArchive/investigation/)
├── implementation/      # archived: ../../docsArchive/cleanup-2026-05-19/docs/inspector/implementation/
└── integration/         # Integration with framework and hook system (archived: /docsArchive/integration/)
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
- [Current Architecture](../architecture/README.md)
- [Rendering Pipeline](/runtime/render/README.md)
