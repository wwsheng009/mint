# Modal Button Click Investigation and Refactoring Plan

## Executive Summary

This document investigates why modal button clicks were not working and proposes a comprehensive refactoring plan to address the architectural issues revealed during the investigation.

**Status**: Fixed ✅ (Temporary Solution)
**Refactoring**: Pending 🔄 (Long-term Architecture Improvement)

## Problem Timeline

### Initial Symptoms
- Modal buttons did not respond to mouse clicks
- HitTest correctly found buttons at `(43, 11, 11x1)`
- Button.HandleEvent was called but `ContainsPoint` returned `false`
- Button.bounds was `[5, 5, 15x1]` (incorrect)

### Investigation Process

#### Phase 1: Event Routing Analysis
**Question**: Why is Button.HandleEvent returning false?

**Findings**:
1. Pump correctly identified button via HitTest: `Bounds=(43,11,11x1)`
2. Direct routing failed: `Component not found in registry: button`
   - Multiple buttons have the same ID="button"
3. Event fallback distributed to all buttons via `distributeEventToVNode`
4. Each button's `ContainsPoint` check failed because bounds were wrong

**Conclusion**: Event routing was working correctly. The issue was with button bounds.

#### Phase 2: Bounds Synchronization Analysis
**Question**: Why are Button.bounds different from HitMap positions?

**Findings**:
```
Timeline of Button.bounds value:
┌────────────────────────────────────────────────────┐
│ 1. Initial Layout                                  │
│    calculatePositions(5, 5)                        │
│    → SetBounds(5, 5, 15, 1)                       │
│    Button.bounds = [5,5,15x1] ✅                   │
│    ComputedBox.Box = (5,5,15x1) ✅                │
└────────────────────────────────────────────────────┘
                     ↓
┌────────────────────────────────────────────────────┐
│ 2. Modal Centering                                 │
│    centerModal() → shiftPositions(+20, +6)         │
│    → ComputedBox.Box = (43,11,11x1) ✅             │
│    → Button.bounds = [5,5,15x1] ❌ STALE!         │
└────────────────────────────────────────────────────┘
```

**Root Cause**: `shiftPositions()` updated `ComputedBox.Box` but did NOT call `SetBounds()` to update `Button.bounds`.

**Fix Applied**: Added `SetBounds()` call in `shiftPositions()` to keep `Button.bounds` synchronized with `ComputedBox.Box`.

### Architectural Issues Revealed

The investigation revealed several fundamental architectural problems:

#### Issue 1: Dual State Management (Violates Single Source of Truth)
```
Layout information is stored in TWO places:
┌──────────────────────────────────────────┐
│ ComputedBox.Box (correct: 43,11,11x1)   │
│ - Managed by LayoutEngine                │
│ - Updated during modal centering         │
│ - Used by HitTest ✅                     │
└──────────────────────────────────────────┘
                    ↓
            Should be the same but...
                    ↓
┌──────────────────────────────────────────┐
│ Button.bounds (stale: 5,5,15x1)         │
│ - Managed by component                   │
│ - Set during initial layout              │
│ - NOT updated during modal centering ❌  │
│ - Used by Button.HandleEvent ❌          │
└──────────────────────────────────────────┘
```

**Problem**:
- Violates Single Source of Truth principle
- State synchronization burden on developers
- Easy to forget updating one of them (like in `shiftPositions`)

#### Issue 2: Coordinate System Confusion
```
Mouse event has THREE coordinate systems:
┌──────────────────────────────────────────┐
│ Global: (48, 11)                         │
│ - Screen coordinates                     │
│ - Used by HitTest ✅                     │
└──────────────────────────────────────────┘
                    ↓ HitTest calculates
┌──────────────────────────────────────────┐
│ Local: (5, 0)                            │
│ - Relative to button                     │
│ - Calculated but NOT used ❌             │
└──────────────────────────────────────────┘
                    ↓
┌──────────────────────────────────────────┐
│ Component: bounds[4]int                  │
│ - [5, 5, 15, 1]                         │
│ - Used by ContainsPoint() ❌             │
│ - Stale/incorrect ❌                     │
└──────────────────────────────────────────┘
```

**Problem**:
- Unclear which coordinate system should be used where
- `LocalX/LocalY` calculated but never used
- `ContainsPoint` uses global coordinates against local bounds

#### Issue 3: Responsibility Confusion

**Question**: Who is responsible for managing component bounds?

```
Current architecture has THREE different paths:

Path A: LayoutEngine.calculatePositions()
├─ Calculates final positions
├─ Calls SetBounds(x, y, w, h)
└─ Used during initial layout

Path B: LayerManager.shiftPositions()
├─ Transforms positions (modal centering)
├─ NOW calls SetBounds() (after fix)
└─ Post-layout transformation

Path C: DeclarativeNode.PaintVNode() (legacy)
├─ Has SetBounds call but was skipped
├─ Comment: "TODO: We need to get calculated layout size"
└─ Unused in new pipeline
```

**Problem**:
- Unclear when SetBounds should be called
- Different code paths have different behaviors
- Easy to miss synchronization points

#### Issue 4: Interface Fragmentation

```go
// Paint interface needs x,y but doesn't document where they come from
type Paintable interface {
    Paint(x, y int) []DrawCmd
}

// SetBounds interface doesn't document when to call it
type SetBoundsProvider interface {
    SetBounds(x, y, width, height int)
}

// Component doesn't know if it should use internal bounds or trust the event
func (b *Button) HandleEvent(ev Event) bool {
    // Should I use b.bounds.ContainsPoint(mouseEv.X, mouseEv.Y)?
    // Or should I trust that HandleEvent was only called if I'm the target?
}
```

## Refactoring Plan

### Goal
Eliminate dual state management and establish a clear single source of truth for layout information.

### Proposed Solutions

#### Option A: Remove Component-Internal Bounds (Recommended)

**Principle**: Components should NOT store layout state. Layout is external.

```go
// BEFORE (current - problematic)
type ButtonVNode struct {
    bounds [4]int  // ← Redundant state!
    label string
    onClick func()
}

func (b *Button) HandleEvent(ev Event) bool {
    if mouseEv, ok := ev.(*MouseEvent); ok {
        if b.ContainsPoint(mouseEv.X, mouseEv.Y) {  // Uses internal bounds
            // handle click
        }
    }
}

// AFTER (proposed)
type ButtonVNode struct {
    label string
    onClick func()
    // NO bounds field!
}

func (b *Button) HandleEvent(ev Event) bool {
    if mouseEv, ok := ev.(*MouseEvent); ok {
        // Option 1: Trust that HandleEvent is only called for target
        if mouseEv.Button == MouseLeft {
            b.onClick()
            return true
        }

        // Option 2: Use TargetBounds from event
        if mouseEv.TargetBounds.ContainsPoint(mouseEv.X, mouseEv.Y) {
            b.onClick()
            return true
        }
    }
}
```

**Benefits**:
- ✅ Single source of truth: ComputedBox
- ✅ No synchronization burden
- ✅ Components focus on behavior, not layout
- ✅ Simpler component model

**Costs**:
- ❌ Breaking change to all components
- ❌ Need to enhance MouseEvent with TargetBounds
- ❌ Need to verify event routing always targets correct component

**Migration Path**:
1. Add `TargetBounds runtime.Box` to `MouseEvent`
2. Modify `HandleEvent` to use `ev.TargetBounds` instead of `b.bounds`
3. Remove `bounds` field from components
4. Remove `SetBounds` interface
5. Update tests

#### Option B: Enforce Synchronization (Conservative)

**Principle**: Keep component bounds but enforce strict synchronization.

```go
// Add assertion to detect inconsistency
func (b *Button) SetBounds(x, y, w, h int) {
    b.bounds[0] = x
    b.bounds[1] = y
    b.bounds[2] = w
    b.bounds[3] = h

    // Add validation in debug mode
    if os.Getenv("TUI_DEBUG_BOUNDS") == "true" {
        // This would need access to ComputedBox, which creates circular dependency
        // So this option is limited
    }
}

// Wrapper to ensure SetBounds is always called after Box changes
func (box *ComputedBox) SetPosition(x, y int) {
    box.Box.X = x
    box.Box.Y = y

    // Automatically sync to VNode
    if boundsAware, ok := box.VNode.(interface{ SetBounds(int, int, int, int) }); ok {
        boundsAware.SetBounds(x, y, box.Box.Width, box.Box.Height)
    }
}
```

**Benefits**:
- ✅ Less breaking change
- ✅ Maintains current component model
- ✅ Easier to migrate incrementally

**Costs**:
- ❌ Still has dual state
- ❌ Performance overhead of always syncing
- ❌ Easy to forget using SetPosition() instead of direct assignment

#### Option C: Event-Centric Architecture (Most Radical)

**Principle**: Events carry all information needed, components don't need bounds.

```go
// MouseEvent carries complete targeting information
type MouseEvent struct {
    Type      EventType
    X, Y      int           // Global coordinates
    LocalX, LocalY int     // Relative to target
    TargetID  string       // From HitTest
    TargetBounds Box      // From HitTest (final position)
    Target    VNode       // Direct reference (optional)
}

// Event distribution is precise
func (n *DeclarativeNode) distributeEventToVNode(vnode VNode, ev Event) bool {
    if mouseEv, ok := ev.(*MouseEvent); ok {
        // Check if this VNode is the actual target
        if mouseEv.TargetID != getVNodeID(vnode) {
            return false  // Skip non-targets
        }

        // For the actual target, call HandleEvent without bounds checking
        if component, ok := vnode.(Component); ok {
            return component.HandleEvent(ev)
        }
    }
    // ... rest of distribution
}

// Component trusts it's the target
func (b *Button) HandleEvent(ev Event) bool {
    if mouseEv, ok := ev.(*MouseEvent); ok {
        if mouseEv.Button == MouseLeft {
            b.onClick()  // No bounds check needed!
            return true
        }
    }
    return false
}
```

**Benefits**:
- ✅ Single source of truth: HitMap/ComputedBox
- ✅ Events are self-contained
- ✅ Components don't need layout state
- ✅ Clean separation: Layout = external, Behavior = internal

**Costs**:
- ❌ Most breaking changes
- ❌ Requires unique IDs for all VNodes
- ❌ Complex event routing changes

## Recommended Refactoring Approach

### Phase 1: Immediate Improvements (Low Risk)
**Timeline**: Week 1-2

1. **Add Bounds Validation** (Testing infrastructure)
   ```go
   // In debug mode, verify ComputedBox.Box == Component.bounds
   func ValidateBoundsConsistency(box *ComputedBox) error {
       if os.Getenv("TUI_DEBUG_VALIDATION") != "true" {
           return nil
       }
       if boundsAware, ok := box.VNode.(interface{ GetBounds() [4]int }); ok {
           actualBounds := boundsAware.GetBounds()
           expected := [4]int{box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height}
           if actualBounds != expected {
               return fmt.Errorf("bounds mismatch: got %v want %v", actualBounds, expected)
           }
       }
       return nil
   }
   ```

2. **Document SetBounds Contract**
   ```go
   // SetBounds sets the component's position and size.
   // This MUST be called whenever the ComputedBox.Box changes.
   //
   // When to call:
   // - In LayoutEngine.calculatePositions() after setting Box.X/Y
   // - In LayerManager.shiftPositions() after transforming positions
   // - Any custom code that modifies ComputedBox.Box
   //
   // DO NOT:
   // - Call this in Paint() (too late, should have been called in Layout)
   // - Assume this is the only place bounds are set (dual state issue)
   func (b *Button) SetBounds(x, y, width, height int)
   ```

3. **Add MouseEvent.TargetBounds**
   ```go
   type MouseEvent struct {
       // ... existing fields
       TargetBounds Box  // Final bounds from HitMap (post-transform)
   }

   // In HitTest
   mouseMsg.TargetBounds = runtime.Box{
       X: entry.Bounds.X,
       Y: entry.Bounds.Y,
       Width: entry.Bounds.Width,
       Height: entry.Bounds.Height,
   }
   ```

### Phase 2: Gradual Migration (Medium Risk)
**Timeline**: Week 3-4

1. **Migrate Button to use MouseEvent.TargetBounds**
   ```go
   func (b *Button) HandleEvent(ev Event) bool {
       if mouseEv, ok := ev.(*MouseEvent); ok {
           // NEW: Use TargetBounds from event
           if mouseEv.TargetBounds.ContainsPoint(mouseEv.X, mouseEv.Y) {
               if mouseEv.Button == MouseLeft && b.onClick != nil {
                   b.onClick()
                   return true
               }
           }
       }
       return false
   }
   ```

2. **Deprecate Component.bounds Field**
   ```go
   type ButtonVNode struct {
       boundsDeprecated [4]int  // TODO: Remove in v2.0
       label string
       onClick func()
   }

   func (b *Button) SetBounds(x, y, w, h int) {
       if os.Getenv("TUI_WARN_DEPRECATED") == "true" {
           log.Warn("Button.bounds is deprecated. Use MouseEvent.TargetBounds instead.")
       }
       b.boundsDeprecated = [4]int{x, y, w, h}
   }
   ```

3. **Add Unique IDs to VNodes** (for Option C)
   ```go
   // Generate stable IDs for all VNodes
   func generateVNodeID(vnode VNode) string {
       return fmt.Sprintf("%s-%p", vnode.Type(), vnode)
   }
   ```

### Phase 3: Full Refactoring (High Risk)
**Timeline**: Week 5-8

1. **Implement Event-Centric Distribution** (Option C)
   - Modify `distributeEventToVNode` to check `TargetID`
   - Remove bounds checking from components
   - Components trust they are the target

2. **Remove Component.bounds**
   - Delete `bounds` field from all components
   - Remove `SetBounds` interface
   - Update all component tests

3. **Cleanup**
   - Remove dual-state synchronization code
   - Simplify event routing
   - Update documentation

## Success Criteria

### Phase 1 Success Criteria
- ✅ Bounds validation detects all inconsistencies
- ✅ MouseEvent includes TargetBounds
- ✅ All SetBounds call sites are documented

### Phase 2 Success Criteria
- ✅ Button uses MouseEvent.TargetBounds
- ✅ All existing tests pass
- ✅ No performance regression

### Phase 3 Success Criteria
- ✅ Component.bounds removed from all components
- ✅ SetBounds interface removed
- ✅ Event routing uses TargetID exclusively
- ✅ 100% test coverage for event handling

## Rollback Plan

If Phase 2 or Phase 3 introduces critical issues:

1. **Revert to Phase 1**
   ```bash
   git revert <phase-2-commits>
   git revert <phase-3-commits>
   ```

2. **Keep bounds validation**
   - Use validation to catch issues early
   - Document known limitations

3. **Document technical debt**
   - Create issue tracking architectural debt
   - Link to this investigation document

## Lessons Learned

1. **Single Source of Truth**: Layout state should live in ONE place
2. **Encapsulation**: Components shouldn't manage their own layout
3. **Event-Driven**: Events should carry all context needed
4. **Validation**: Add assertions to detect state inconsistencies early
5. **Documentation**: Document when and where to call synchronization methods

## Related Issues

- Modal button click handling (this investigation)
- Event routing architecture
- Component lifecycle management
- Layout-Paint separation

## References

- Original Issue: Modal buttons not responding to clicks
- HitMap Implementation: `runtime/event/hitmap.go`
- Layout Engine: `runtime/compute/engine.go`
- Layer Manager: `runtime/layer/manager.go`
- Event Routing: `framework/event/` and `internal/render/declarative_node.go`

---

**Document Version**: 1.0
**Last Updated**: 2025-02-11
**Status**: Ready for Review
**Owner**: Architecture Team
