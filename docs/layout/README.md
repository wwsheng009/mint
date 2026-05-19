# Layout System Documentation

Welcome to the Mint TUI Layout System documentation. This directory contains comprehensive documentation about the layout system, organized by purpose and audience.

> Current source note: the layout engine is now reached from the Fiber-first render path through `internal/render.FiberToNodeAdapter` and `runtime/layout`. Many older layout notes still contain pre-`ui/components` snippets such as `app.WrapBuilder(...)` or `app.NewTextBuilder(...)`; treat those snippets as historical unless the page explicitly says it was updated for the current `ui.NewXBuilder` API.

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
- **[CONTAINER_BACKGROUND_IMPLEMENTATION_SUMMARY.md](/docsArchive/layout/core_concepts/CONTAINER_BACKGROUND_IMPLEMENTATION_SUMMARY.md)** - Background implementation summary (archived)

**Best for**: Developers who want to understand how the layout system works

---

### Architecture Notes
**System architecture and technical analysis**

Historical layout architecture and rendering implementation notes are archived under `../../docsArchive/cleanup-2026-05-19/docs/layout/` and `../../docsArchive/`.

**Best for**: Contributors working on the layout system internals

---

### [issues/](/docsArchive/issues/)
**Problem analysis and fix documentation**

- **[FLEX_LAYOUT_ISSUE_ANALYSIS.md](/docsArchive/issues/FLEX_LAYOUT_ISSUE_ANALYSIS.md)** - Flex layout issue analysis (elegant_api_demo)
- **[layout_fix_analysis_20260209.md](/docsArchive/issues/layout_fix_analysis_20260209.md)** - TreeView constraint propagation fix
- **[FILLWIDTH_FINAL_FIX.md](/docsArchive/issues/FILLWIDTH_FINAL_FIX.md)** - Wrap component FillWidth final fix
- **[FILLWIDTH_REAL_FIX.md](/docsArchive/issues/FILLWIDTH_REAL_FIX.md)** - FillWidth real fix attempt
- **[FILLWIDTH_TRUE_FIX.md](/docsArchive/issues/FILLWIDTH_TRUE_FIX.md)** - FillWidth true fix attempt
- **[WRAP_FILLWIDTH_FIX_SUMMARY.md](/docsArchive/issues/WRAP_FILLWIDTH_FIX_SUMMARY.md)** - Wrap FillWidth fix summary
- **[LAYOUT_FIX_PLAN.md](/docsArchive/issues/LAYOUT_FIX_PLAN.md)** - Layout fix plan
- **[border_overflow_explanation.md](/docsArchive/issues/border_overflow_explanation.md)** - Border overflow explanation
- **[flex_wrap_limitation.md](/docsArchive/issues/flex_wrap_limitation.md)** - Why flex-wrap is not supported
- **[layout_engine_fix_summary.md](/docsArchive/issues/layout_engine_fix_summary.md)** - Layout engine fix summary
- **[modal_centering_fix.md](/docsArchive/issues/modal_centering_fix.md)** - Modal centering fix

**Best for**: Debugging issues and understanding historical problems

---

### Refactor History
**Refactoring design and history**

Refactor history is archived under `../../docsArchive/cleanup-2026-05-19/docs/layout/refactor/` and older `../../docsArchive/` records.

**Best for**: Understanding the evolution of the layout system

---

### api_reference/
**API documentation** (planned - not yet available)

**Best for**: Looking up specific API methods and parameters

---

### design/
**Design documents** (planned - not yet available)

**Best for**: Proposed features and design discussions

---

### components/
**Component-specific documentation** (planned - not yet available)

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
→ Check [issues/](/docsArchive/issues/) for similar problems and solutions

#### "I want to contribute to the layout system"
→ Read [architecture/LAYER_LAYOUT_ARCHITECTURE_REVIEW.md](/docsArchive/LAYER_LAYOUT_ARCHITECTURE_REVIEW.md)

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
- [Architecture Documentation](/docs/architecture/)
- [Component Documentation](/docs/components/)

---

**Last Updated**: 2025-02-15
**Maintained By**: Mint TUI Team
