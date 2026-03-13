# Migrate Components to MouseEvent.TargetBounds

## Overview

This guide explains how to migrate components to use `MouseEvent.TargetBounds` for accurate hit testing, fixing issues where components don't respond to clicks after layout transforms (e.g., modal centering).

### The Problem

Components previously used internal `bounds` state for hit testing, which could become stale after layout transforms like modal centering:

```go
// OLD: Uses internal bounds (can be stale)
if b.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == framewarkevent.MouseLeft {
    // Handle click...
}
```

### The Solution

Use `MouseEvent.TargetBounds` which contains the final rendered position from the HitMap:

```go
// NEW: Uses TargetBounds (always accurate)
if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
    inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
        mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
        mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
        mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height

    if inBounds && mouseEvent.Button == framewarkevent.MouseLeft {
        // Handle click...
    }
}
```

## When to Migrate

Migrate components that:

1. ✅ **Handle mouse clicks** (`EventMousePress`, `EventClick`)
2. ✅ **Have bounds-based hit testing** (use `ContainsPoint()` or similar)
3. ✅ **Are used inside modals/overlays** (where centering transforms happen)

### Components That Need Migration

Based on codebase analysis:

| Component | Has HandleEvent | Mouse Events | Priority |
|-----------|----------------|--------------|----------|
| Button ✅ | ✅ | EventMousePress, EventMouseRelease | ✅ Done |
| Checkbox | ✅ | EventMousePress, EventClick | High |
| Modal | ✅ | EventMousePress (click-outside) | High |
| Input | ✅ | EventMousePress (focus handling) | Medium |
| Select | ✅ | EventMousePress (dropdown toggle) | Medium |
| Textarea | ✅ | EventMousePress (focus handling) | Medium |
| TreeView | ✅ | EventMousePress (node selection) | Medium |
| Tabs | ✅ | EventMousePress (tab selection) | Low |
| Panel | ✅ | EventMousePress (forwarding only) | Low |

## Step-by-Step Migration

### Phase 1: Identify the Pattern

Look for code like this in your component's `HandleEvent` method:

```go
case frameworkevent.EventMousePress:
    // OLD: Uses internal bounds
    if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == frameworkevent.MouseLeft {
        // Handle click...
    }
```

### Phase 2: Apply the TargetBounds Pattern

Replace with:

```go
case frameworkevent.EventMousePress:
    // Phase 3: Use TargetBounds from event for accurate hit testing
    // TargetBounds contains the final rendered position after all transforms (modal centering, etc.)
    if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
        // NEW: Use TargetBounds for hit testing (post-transform position from HitMap)
        inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
            mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
            mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
            mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height

        if inBounds && mouseEvent.Button == frameworkevent.MouseLeft {
            // Handle click...
        }
    } else {
        // FALLBACK: Use internal bounds if TargetBounds not available (legacy path)
        // This should only happen if HitMap is not properly populated
        log.UILogger.Debug("Component EventMousePress: TargetBounds empty, using legacy bounds check")
        if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == frameworkevent.MouseLeft {
            // Handle click...
        }
    }
```

### Phase 3: Remove Old Helper Methods (Optional)

If your component has a `ContainsPoint` or `containsPoint` method that's no longer needed:

```go
// OLD: Can be removed if no longer used
func (c *Component) ContainsPoint(x, y int) bool {
    if c.bounds[2] <= 0 || c.bounds[3] <= 0 {
        return false
    }
    return x >= c.bounds[0] && x < c.bounds[0]+c.bounds[2] &&
        y >= c.bounds[1] && y < c.bounds[1]+c.bounds[3]
}
```

**Note:** Keep the method if it's still used elsewhere in the component.

## Component-Specific Patterns

### Pattern 1: Click Targets (Button, Checkbox, Input)

These components respond to clicks on their own bounds.

**Before (Button):**
```go
case frameworkevent.EventMousePress:
    if b.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == frameworkevent.MouseLeft {
        if b.onClick != nil {
            b.onClick()
        }
        return true
    }
```

**After:**
```go
case frameworkevent.EventMousePress:
    if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
        inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
            mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
            mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
            mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height

        if inBounds && mouseEvent.Button == frameworkevent.MouseLeft {
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }
    } else {
        // Fallback...
    }
```

### Pattern 2: Click-Outside Detection (Modal)

These components detect clicks outside their bounds (e.g., modal close on click-outside).

**Before (Modal):**
```go
case frameworkevent.EventMousePress:
    // Check if click is outside modal bounds
    if !m.containsPoint(mouseEvent.X, mouseEvent.Y) {
        m.isOpen = false
        if m.onClose != nil {
            m.onClose()
        }
        return true
    }
```

**After:**
```go
case frameworkevent.EventMousePress:
    // Check if click is outside modal bounds
    if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
        // NEW: Use TargetBounds for hit testing
        inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
            mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
            mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
            mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height

        if !inBounds {
            m.isOpen = false
            if m.onClose != nil {
                m.onClose()
            }
            return true
        }
    } else {
        // Fallback: Use legacy bounds check
        if !m.containsPoint(mouseEvent.X, mouseEvent.Y) {
            m.isOpen = false
            if m.onClose != nil {
                m.onClose()
            }
            return true
        }
    }
```

### Pattern 3: Container Components (Panel)

These components forward events to children. They typically don't need bounds checking.

**No migration needed:**
```go
func (p *Panel) HandleEvent(ev frameworkevent.Event) bool {
    // Only forward mouse events
    if ev.Type() != frameworkevent.EventMousePress {
        return false
    }

    // Forward to child component
    if contentComponent, ok := p.content.(frameworkevent.Component); ok {
        return contentComponent.HandleEvent(ev)
    }

    return false
}
```

## Testing Checklist

After migration, verify:

- [ ] Component responds to clicks in normal layout
- [ ] Component responds to clicks inside modal (if applicable)
- [ ] Component responds to clicks after window resize
- [ ] Hover states work correctly
- [ ] Focus handling still works
- [ ] No regression in keyboard interaction
- [ ] Fallback path logs warning (check logs with `TUI_DEBUG_UI=true`)

### Test Case: Modal Button Click

```bash
# Enable debug logging
export TUI_DEBUG_UI=true

# Run the app with modal buttons
go run ./examples/modal

# Click on modal buttons and verify:
# 1. No "legacy bounds" warnings in logs
# 2. TargetBounds logs show correct coordinates
# 3. Buttons respond to clicks
```

## Common Pitfalls

### ❌ Don't Use Internal Bounds After Layout Transforms

```go
// WRONG: This will fail after modal centering
if mouseEvent.X >= b.bounds[0] && mouseEvent.X < b.bounds[0]+b.bounds[2] {
    // Handle click...
}
```

### ❌ Don't Forget the Fallback Path

```go
// INCOMPLETE: No fallback if TargetBounds is empty
if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
    // Handle with TargetBounds...
}
// Missing: What if TargetBounds is empty?
```

### ✅ Always Include Fallback

```go
if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
    // NEW: Use TargetBounds
} else {
    // FALLBACK: Use legacy bounds
    log.UILogger.Debug("TargetBounds empty, using legacy check")
    // Legacy handling...
}
```

## Migration Status

| Component | Status | Notes |
|-----------|--------|-------|
| Button | ✅ Complete | Uses TargetBounds with fallback |
| Checkbox | ⏳ Pending | High priority - commonly used in forms |
| Modal | ⏳ Pending | High priority - click-outside detection |
| Input | ⏳ Pending | Medium priority - focus handling |
| Select | ⏳ Pending | Medium priority - dropdown toggle |
| Textarea | ⏳ Pending | Medium priority - focus handling |
| TreeView | ⏳ Pending | Medium priority - node selection |
| Tabs | ⏳ Pending | Low priority - mostly keyboard nav |
| Panel | ✅ No migration needed | Container, forwards events only |

## Related Documentation

- **Investigation:** `docsArchive/issue/modal-button-click-investigation.md`
- **Architecture:** Runtime event system, HitMap-based routing
- **Implementation:**
  - `framework/event/event.go` - MouseEvent.TargetBounds definition
  - `framework/event/pump.go` - TargetBounds population from HitMap
  - `components/button/button.go` - Reference implementation

## Summary

**Key Points:**

1. **TargetBounds is authoritative** - Contains final rendered position from HitMap
2. **Always include fallback** - For cases where HitMap isn't populated
3. **Log debug info** - Helps troubleshoot if issues arise
4. **Test with modals** - That's where the bug manifests

**Migration Formula:**

```go
// Replace this:
if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) {

// With this:
if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
    inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
        mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
        mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
        mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height
    if inBounds {

// Plus fallback:
} else {
    log.UILogger.Debug("TargetBounds empty, using legacy")
    if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) {
```

Questions? Refer to the Button implementation in `components/button/button.go:537` as a reference.
