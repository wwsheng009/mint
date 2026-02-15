# Layout System Documentation

Welcome to the Mint TUI Layout System documentation. This directory contains comprehensive documentation about the layout system, organized by purpose and audience.

## 📚 Directory Structure

### [user_guide/](./user_guide/)
**Quick start guides and tutorials for users**

- **[README.md](./user_guide/README.md)** - Documentation index with navigation and quick links
- **[background_quick_reference.md](./user_guide/background_quick_reference.md)** - Quick reference for container backgrounds
- **[wrap_cheatsheet.md](./user_guide/wrap_cheatsheet.md)** - Wrap component quick reference
- **[flex_comparison.md](./user_guide/flex_comparison.md)** - Mint TUI vs CSS Flexbox comparison

**Best for**: New users getting started with Mint TUI layout

---

### [core_concepts/](./core_concepts/)
**In-depth explanations of layout system features**

- **[flex_layout.md](./core_concepts/flex_layout.md)** - Flex layout system implementation
- **[flex_examples.md](./core_concepts/flex_examples.md)** - Flex layout usage examples
- **[stretch_layout.md](./core_concepts/stretch_layout.md)** - Stretch layout system (FillWidth/FillHeight)
- **[wrap_component.md](./core_concepts/wrap_component.md)** - Wrap component for automatic line wrapping
- **[container_background_rendering.md](./core_concepts/container_background_rendering.md)** - Container background rendering system
- **[layer_system_guide.md](./core_concepts/layer_system_guide.md)** - Layer system for overlays and modals
- **[CONTAINER_BACKGROUND_IMPLEMENTATION_SUMMARY.md](./core_concepts/CONTAINER_BACKGROUND_IMPLEMENTATION_SUMMARY.md)** - Background implementation summary

**Best for**: Developers who want to understand how the layout system works

---

### [architecture/](./architecture/)
**System architecture and technical analysis**

- **[LAYER_LAYOUT_ARCHITECTURE_REVIEW.md](./architecture/LAYER_LAYOUT_ARCHITECTURE_REVIEW.md)** - Comprehensive Layer and Layout architecture review
- **[getChildconstraints_architecture_analysis.md](./architecture/getChildconstraints_architecture_analysis.md)** - Constraint propagation mechanism analysis
- **[layer_constraint_audit_report.md](./architecture/layer_constraint_audit_report.md)** - Layer system constraint issues
- **[rendering_pipeline_implementation.md](./architecture/rendering_pipeline_implementation.md)** - Rendering pipeline implementation details

**Best for**: Contributors working on the layout system internals

---

### [issues/](./issues/)
**Problem analysis and fix documentation**

- **[FLEX_LAYOUT_ISSUE_ANALYSIS.md](./issues/FLEX_LAYOUT_ISSUE_ANALYSIS.md)** - Flex layout issue analysis (elegant_api_demo)
- **[layout_fix_analysis_20260209.md](./issues/layout_fix_analysis_20260209.md)** - TreeView constraint propagation fix
- **[FILLWIDTH_FINAL_FIX.md](./issues/FILLWIDTH_FINAL_FIX.md)** - Wrap component FillWidth final fix
- **[FILLWIDTH_REAL_FIX.md](./issues/FILLWIDTH_REAL_FIX.md)** - FillWidth real fix attempt
- **[FILLWIDTH_TRUE_FIX.md](./issues/FILLWIDTH_TRUE_FIX.md)** - FillWidth true fix attempt
- **[WRAP_FILLWIDTH_FIX_SUMMARY.md](./issues/WRAP_FILLWIDTH_FIX_SUMMARY.md)** - Wrap FillWidth fix summary
- **[LAYOUT_FIX_PLAN.md](./issues/LAYOUT_FIX_PLAN.md)** - Layout fix plan
- **[border_overflow_explanation.md](./issues/border_overflow_explanation.md)** - Border overflow explanation
- **[flex_wrap_limitation.md](./issues/flex_wrap_limitation.md)** - Why flex-wrap is not supported
- **[layout_engine_fix_summary.md](./issues/layout_engine_fix_summary.md)** - Layout engine fix summary
- **[modal_centering_fix.md](./issues/modal_centering_fix.md)** - Modal centering fix

**Best for**: Debugging issues and understanding historical problems

---

### [refactor/](./refactor/)
**Refactoring design and history**

- **[layout_refactor.md](./refactor/layout_refactor.md)** - Layout system refactor history
- **[LAYOUT_RENDERING_REFACTOR.md](./refactor/LAYOUT_RENDERING_REFACTOR.md)** - Rendering pipeline refactor
- **[single_pass_layout_refactor_design.md](./refactor/single_pass_layout_refactor_design.md)** - Single-pass layout refactor design
- **[single_pass_refactor_summary.md](./refactor/single_pass_refactor_summary.md)** - Single-pass refactor summary

**Best for**: Understanding the evolution of the layout system

---

### [api_reference/](./api_reference/)
**API documentation** (placeholder - to be populated)

**Best for**: Looking up specific API methods and parameters

---

### [design/](./design/)
**Design documents** (placeholder - for future design docs)

**Best for**: Proposed features and design discussions

---

### [components/](./components/)
**Component-specific documentation** (placeholder - to be populated)

**Best for**: Documentation about specific layout components

---

## 🚀 Quick Navigation

### By Use Case

#### "I want to learn the basics"
→ Start with [user_guide/README.md](./user_guide/README.md)

#### "I need to implement automatic wrapping"
→ Read [core_concepts/wrap_component.md](./core_concepts/wrap_component.md)

#### "I need to make elements stretch"
→ Read [core_concepts/stretch_layout.md](./core_concepts/stretch_layout.md)

#### "I want to use flex layout"
→ Read [core_concepts/flex_layout.md](./core_concepts/flex_layout.md)

#### "I need container backgrounds"
→ Read [core_concepts/container_background_rendering.md](./core_concepts/container_background_rendering.md)

#### "I want to understand modals/overlays"
→ Read [core_concepts/layer_system_guide.md](./core_concepts/layer_system_guide.md)

#### "I'm debugging a layout issue"
→ Check [issues/](./issues/) for similar problems and solutions

#### "I want to contribute to the layout system"
→ Read [architecture/LAYER_LAYOUT_ARCHITECTURE_REVIEW.md](./architecture/LAYER_LAYOUT_ARCHITECTURE_REVIEW.md)

---

## 📖 Documentation Statistics

| Category | Documents | Focus |
|----------|-----------|-------|
| User Guides | 4 | Quick reference and tutorials |
| Core Concepts | 7 | Feature documentation |
| Architecture | 4 | System design and analysis |
| Issues | 11 | Problem analysis and fixes |
| Refactor | 4 | Refactoring history |
| **Total** | **30** | Complete layout system coverage |

---

## 🗂️ File Organization Philosophy

The documentation is organized by **purpose** and **audience**:

- **user_guide/** - For users who want to use the layout system
- **core_concepts/** - For developers who want to understand how it works
- **architecture/** - For contributors working on internals
- **issues/** - For debugging and historical reference
- **refactor/** - For understanding system evolution
- **api_reference/** - For API lookups (to be expanded)
- **design/** - For future design proposals
- **components/** - For component-specific docs (to be expanded)

---

## 🤝 Contributing

When adding new documentation:

1. **Choose the right directory** based on purpose and audience
2. **Use clear filenames** that describe the content
3. **Update this README** if adding a new category
4. **Follow the existing format** with clear sections and examples

---

## 🔗 Related Documentation

- [Main Project README](../../README.md)
- [Architecture Documentation](../../architecture/)
- [Component Documentation](../../components/)

---

**Last Updated**: 2025-02-15
**Maintained By**: Mint TUI Team
