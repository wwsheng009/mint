# Inspector Investigation & Analysis

This directory contains investigation reports and analysis documents from the development and debugging of the Inspector system.

## Investigation Reports

### [INSPECTOR_INVESTIGATION_COMPLETE.md](INSPECTOR_INVESTIGATION_COMPLETE.md)
Complete investigation report covering all major issues discovered and resolved during Inspector development.

**Topics Covered:**
- Inspector overlay not displaying
- App content disappearing when Inspector enabled
- TreeView expand/collapse failures
- UniqueID collision issues
- Layout and sizing problems

### [INSPECTOR_FINAL_INVESTIGATION_REPORT.md](INSPECTOR_FINAL_INVESTIGATION_REPORT.md)
Comprehensive final report with detailed analysis of all issues and their resolutions.

### [INSPECTOR_LAYER_INVESTIGATION_SUMMARY.md](INSPECTOR_LAYER_INVESTIGATION_SUMMARY.md)
Summary of layer system investigation and how multi-layer rendering works.

## Specific Issues

### UniqueID Collisions

#### [INSPECTOR_POINTER_ID_SUMMARY.md](INSPECTOR_POINTER_ID_SUMMARY.md)
Analysis of pointer-based UniqueID solution for preventing collisions.

**Problem:**
```
vstack.bordered.vstack[0]  ← LayoutNode
vstack.bordered.vstack[0]  ← BorderNode (same ID!)
```

**Solution:**
```
vstack.bordered.vstack[0]@0x1234567890  ← LayoutNode
vstack.bordered.vstack[0]@0x9876543210  ← BorderNode (different pointer!)
```

### TreeView Issues

#### [INSPECTOR_TREEVIEW_OVERFLOW.md](INSPECTOR_TREEVIEW_OVERFLOW.md)
Investigation of TreeView content overflow and scrolling issues.

**Findings:**
- TreeView height was fixed at 25 rows
- Content could exceed viewport
- No scrolling mechanism

**Solution:**
- Implemented flex-based auto-sizing
- Added virtual scrolling support
- Dynamic height based on inspector height

### Rendering Issues

#### [INSPECTOR_VSTACK_BUG_ANALYSIS.md](INSPECTOR_VSTACK_BUG_ANALYSIS.md)
Analysis of VStack rendering issues in Inspector.

### [TUI_INSPECTOR_ENVIRONMENT_IMPACT_REPORT.md](TUI_INSPECTOR_ENVIRONMENT_IMPACT_REPORT.md)
Report on how TUI_INSPECTOR environment variable affects the application.

**Key Findings:**
- Enabling Inspector changed rendering behavior
- App content disappeared with Inspector enabled
- Root cause: Fragment wrapping and Layer management

### Interactive Debugging

#### [INSPECTOR_THREE_QUESTIONS_ANSWER.md](INSPECTOR_THREE_QUESTIONS_ANSWER.md)
Answers to three key questions about Inspector behavior.

#### [FOUR_QUESTIONS_ANSWERED.md](FOUR_QUESTIONS_ANSWERED.md)
Additional investigation findings and question resolution.

## Investigation Process

### Phase 1: Problem Identification
1. Inspector overlay not visible
2. App content disappearing
3. TreeView not working (expand/collapse)

### Phase 2: Root Cause Analysis
1. **Inspector Not Showing:**
   - Checked TUI_INSPECTOR environment variable
   - Verified Inspector.IsVisible() state
   - Traced render pipeline

2. **App Content Disappearing:**
   - Found Fragment wrapping issue
   - Discovered SetProps/SetLayer bug
   - Identified layer rendering problem

3. **TreeView Failures:**
   - Analyzed UniqueID generation
   - Found collision bug
   - Tracked tree node selection

### Phase 3: Solution Development
1. Hook-based automatic injection
2. Pointer-based UniqueID fallback
3. Proper SetProps/SetLayer ordering
4. Flex layout for dynamic sizing

### Phase 4: Verification
1. Unit tests for each fix
2. Integration tests for full pipeline
3. Manual testing with demo2

## Key Findings

### Finding 1: Manual Fragment Wrapping is Error-Prone
**Issue:** Application code had to manually wrap content with Inspector overlay.

**Impact:**
- Easy to forget
- Inconsistent implementation
- Tight coupling between app and Inspector

**Solution:** Hook-based automatic injection

### Finding 2: SetProps/SetLayer Ordering Bug
**Issue:** `SetProps()` replaces entire props map, not merging.

**Impact:**
- Calling SetProps after SetLayer loses the layer
- Inspector overlay rendered in wrong layer
- Content not visible

**Solution:** Always call SetProps before SetLayer

### Finding 3: UniqueID Collisions
**Issue:** Index-based IDs collided when same path + index appeared multiple times.

**Impact:**
- TreeView couldn't distinguish nodes
- Wrong node expanded on 'E' press
- Selection state corrupted

**Solution:** Use VNode pointer address as fallback

### Finding 4: Fixed Layout Doesn't Adapt
**Issue:** Fixed heights and widths don't adapt to content or viewport.

**Impact:**
- Content overflow or underflow
- Poor user experience
- Wasted screen space

**Solution:** Flex layout with grow/shrink

## Debugging Techniques Used

### 1. Environment Variable Logging
```bash
TUI_DEBUG_UI=true              # Framework-level logging
TUI_DEBUG_INSPECTOR=true     # Inspector logging
```

### 2. Selective Test Execution
```bash
# Test specific components
go test -v ./internal/inspector -run TestTreeView
go test -v ./internal/inspector -run TestInspectorHook

# Test integration points
go test -v ./internal/render -run TestPipelineRenderer
```

### 3. Binary Search Debugging
- Isolated changes by reverting commits
- Identified breaking changes
- Verified fixes incrementally

### 4. Direct Execution Testing
```bash
# Build and run with specific conditions
go build -o demo.exe main.go
TUI_INSPECTOR=true ./demo.exe
```

## Lessons Learned

### 1. Pure Functions Are Easier to Debug
Hooks as pure functions (VNode → VNode) are:
- Easy to test in isolation
- Predictable behavior
- Simple to reason about

### 2. Separation of Concerns Prevents Bugs
- Framework shouldn't import render internals
- Application shouldn't know about Inspector
- Inspector shouldn't manage layers

### 3. Type System Limitations Require Workarounds
- Go's type assertions don't work across package boundaries
- Reflection needed for framework integration
- Interface{} requires careful handling

### 4. Small Issues Compound
- SetProps bug → Layer lost → Content invisible
- UniqueID collision → Wrong selection → TreeView broken
- Fixed height → Overflow → Poor UX

### 5. Debugging Requires Multiple Approaches
- Unit tests catch specific bugs
- Integration tests catch interaction bugs
- Manual testing catches UX issues
- Logging catches runtime issues

## Recommendations

### For Future Development

1. **Use Hooks for Cross-Cutting Concerns**
   - Automatic injection
   - Centralized logic
   - Clean application code

2. **Always Test Props Manipulation**
   - SetProps replaces, doesn't merge
   - Order matters (props before layer)
   - Verify after each call

3. **Provide Unique Identifiers**
   - User keys preferred
   - Pointer fallback for components
   - Index is not enough

4. **Use Flex Layout**
   - Adapts to content
   - Responsive to viewport
   - Better than fixed sizes

5. **Document Integration Points**
   - How layers work
   - Where hooks are applied
   - What each component does

## Related Documentation

- [Implementation Details](../implementation/)
- [Architecture Overview](../architecture/)
- [Integration Status](../integration/)
