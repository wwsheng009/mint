# Layout Optimization Plan

> **Version**: 2.0
> **Created**: 2025-02-06
> **Status**: ✅ Completed (Priority 1 & 2)
> **Completed**: 2025-02-06

## Overview

This document outlines the optimization plan for the TUI layout system based on the evaluation completed in `docs/layout/layout_refactor.md` and `docs/layout/stretch_layout.md`.

## Summary

**All Priority 1 and 2 optimizations have been completed successfully!**

| ID | Title | Priority | Status |
|----|-------|----------|--------|
| OPT-1 | Eliminate Repeated Flex Calculation | P1 | ✅ Completed |
| OPT-2 | Optimize Space Padding | P1 | ✅ Completed |
| OPT-3 | Add Text Alignment Options | P2 | ✅ Completed |
| OPT-4 | Add VStackBuilder API | P2 | ✅ Already existed |

## Baseline

Before optimization, baseline output has been saved to:
- **File**: `docs/plan/baseline_output.txt`
- **Demo**: `./examples/ui_demos/demo1_full_featured/main.go`

---

## Completed Optimizations

### OPT-1: Eliminate Repeated Flex Calculation ✅

**Problem**: In `getChildConstraints()`, when calculating flex width for a child, the code re-measures all sibling elements. For N flex children, this results in O(N²) measurements.

**Solution**: Added `FlexDistributionInfo` cache and `getFlexDistribution()` method.

**Files Modified**:
- `runtime/compute/types.go` - Added `FlexDistributionInfo` struct
- `runtime/compute/engine.go` - Added cache field, `getFlexDistribution()` method, updated `getChildConstraints()`

**Verification**:
- Debug output now shows `(cached)` marker
- Measure calls reduced from O(N²) to O(N)

### OPT-2: Optimize Space Padding ✅

**Problem**: Text padding uses character-by-character loop instead of efficient string building.

**Solution**: Replaced loop with `strings.Repeat()`.

**Files Modified**:
- `runtime/compute/engine.go` - Added `strings` import, replaced loop

**Verification**:
- Demo runs successfully
- Text stretching works identically

### OPT-3: Add Text Alignment Options ✅

**Problem**: When text stretches, it's always left-aligned. No option for center or right alignment.

**Solution**: Added `TextAlign` type and alignment support in `layoutVStack()`.

**Files Modified**:
- `runtime/types.go` - Added `TextAlign` type (`Left`, `Center`, `Right`)
- `runtime/compute/engine.go` - Updated `layoutVStack()` to apply alignment
- `ui/shortcuts.go` - Added `TextAlign()`, `TextCenter()`, `TextRight()` functions

**API**:
```go
ui.Text("Hello")                      // Left aligned (default)
ui.TextAlign("Hello", "center")       // Center aligned
ui.TextCenter("Hello")                // Center aligned (shortcut)
ui.TextRight("Hello")                 // Right aligned (shortcut)
```

### OPT-4: Add VStackBuilder API ✅

**Status**: Already existed in codebase - no changes needed.

Both `HStackBuilder` and `VStackBuilder` were already implemented with:
- `.Gap(int)` - Set spacing between children
- `.Stretch()` - Make children fill cross-axis space
- `.Align(Align)` - Set main axis alignment
- `.Padding(top, right, bottom, left)` - Set padding

---

## Pending Optimizations (Priority 3)

### OPT-5: Clarify Stretch Behavior in Documentation

**Status**: Pending

**Action**: Add diagram showing propagation behavior in `docs/layout/stretch_layout.md`

### OPT-6: Separate Measure and Layout Phase Sizes

**Status**: Pending

**Action**: Add separate debug output for Layout phase in `runtime/compute/engine.go`

---

## Implementation Order (Actual)

1. ✅ **OPT-4** (VStackBuilder) - Already existed
2. ✅ **OPT-2** (Space padding) - `strings.Repeat()`
3. ✅ **OPT-1** (Flex cache) - Added `FlexDistributionInfo` cache
4. ✅ **OPT-3** (Text alignment) - Added `TextAlign` type and helpers
5. ⏳ **OPT-5** (Documentation) - Pending
6. ⏳ **OPT-6** (Debug output) - Pending

---

## Tracking

| ID | Title | Priority | Status | Date Completed |
|----|-------|----------|--------|----------------|
| OPT-1 | Eliminate Repeated Flex Calculation | P1 | ✅ Completed | 2025-02-06 |
| OPT-2 | Optimize Space Padding | P1 | ✅ Completed | 2025-02-06 |
| OPT-3 | Add Text Alignment Options | P2 | ✅ Completed | 2025-02-06 |
| OPT-4 | Add VStackBuilder API | P2 | ✅ Already Existed | 2025-02-06 |
| OPT-5 | Clarify Stretch Documentation | P3 | ⏳ Pending | - |
| OPT-6 | Separate Debug Output | P3 | ⏳ Pending | - |

---

## Change Log

| Date | Version | Changes |
|------|---------|---------|
| 2025-02-06 | 1.0 | Initial plan creation |
| 2025-02-06 | 2.0 | Completed OPT-1, OPT-2, OPT-3; verified OPT-4 already existed |
