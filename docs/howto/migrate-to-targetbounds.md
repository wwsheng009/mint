# Migrate Components To `TargetBounds`

This guide explains how to migrate mouse-aware components away from stale local bounds checks and toward HitMap-derived target bounds.

## Current Event Model

Current mouse targeting is primarily carried by `runtime/msg.MouseMsg`.

The event pump fills mouse target data from the latest HitMap:

- `TargetID`
- `TargetFiber`
- `LocalX`
- `LocalY`
- `TargetBounds`

Compatibility `framework/event.MouseEvent` also has `TargetBounds`, but new code should prefer the runtime `msg` / action path where possible.

Relevant source:

- `../../framework/event/pump.go`: fills mouse target fields from HitMap.
- `../../framework/event/event.go`: legacy `MouseEvent.TargetBounds`.
- `../../runtime/msg/mouse_msg.go`: current `MouseMsg` target fields.
- `../../runtime/event/hitmap.go`: HitMap construction and hit testing.
- `../../internal/render/declarative_node.go`: stores/returns render HitMap.

## Problem

Local component bounds can become stale after transforms such as modal centering, Portal positioning, overlay clamping, or layer-specific layout.

Old style:

```go
if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) &&
    mouseEvent.Button == frameworkevent.MouseLeft {
    // Handle click...
}
```

This can fail when the component's internal bounds do not match the final rendered position.

## Preferred Pattern

Use target bounds from the HitMap-populated event/message when available:

```go
func containsTargetBounds(x, y int, bounds types.Rect) bool {
    if bounds.Width <= 0 || bounds.Height <= 0 {
        return false
    }
    return x >= bounds.X &&
        x < bounds.X+bounds.Width &&
        y >= bounds.Y &&
        y < bounds.Y+bounds.Height
}
```

For legacy `frameworkevent.MouseEvent`:

```go
case frameworkevent.EventMousePress:
    if mouseEvent.TargetBounds.Width > 0 && mouseEvent.TargetBounds.Height > 0 {
        inBounds := mouseEvent.X >= mouseEvent.TargetBounds.X &&
            mouseEvent.X < mouseEvent.TargetBounds.X+mouseEvent.TargetBounds.Width &&
            mouseEvent.Y >= mouseEvent.TargetBounds.Y &&
            mouseEvent.Y < mouseEvent.TargetBounds.Y+mouseEvent.TargetBounds.Height

        if inBounds && mouseEvent.Button == frameworkevent.MouseLeft {
            return c.handleClick()
        }
        return false
    }

    // Compatibility fallback for paths where HitMap is unavailable.
    if c.ContainsPoint(mouseEvent.X, mouseEvent.Y) &&
        mouseEvent.Button == frameworkevent.MouseLeft {
        return c.handleClick()
    }
```

For current `runtime/msg.MouseMsg` / Action payloads, prefer using the payload's `TargetBounds` and `TargetFiber` instead of converting back to framework events.

## When To Migrate

Prioritize components that:

- Handle mouse press/release/click/wheel.
- Perform `ContainsPoint`, `containsPoint`, or manual bounds checks.
- Render inside Modal, Popover, Popconfirm, Tooltip, Drawer, Select popup, Menu popup or Portal.
- Need correct click-outside behavior.

Do not migrate a component purely because it forwards events to children and does not own hit testing.

## Component Status

This document is a migration guide, not an authoritative status matrix. The codebase has evolved, and several components still intentionally use local geometry for component-internal subregions.

Before updating a status table, inspect the actual component:

```bash
rg "ContainsPoint|containsPoint|TargetBounds|TargetFiber|LocalX|LocalY" ui/components/<component>
```

Known areas that deserve careful review:

- `ui/components/button`
- `ui/components/modal`
- `ui/components/select`
- `ui/components/popover`
- `ui/components/popconfirm`
- `ui/components/menu`
- `ui/components/input`
- `ui/components/textarea`
- `ui/components/treeview`
- `ui/components/tabs`

## Click-Outside Components

Click-outside logic must distinguish the overlay content target from backdrop or outside targets. For overlay components, do not blindly treat the current event target bounds as the modal content bounds unless the HitMap target is known to be the modal content.

Safer approach:

- Use `TargetFiber` / target metadata when available.
- Keep local content rectangle checks when they refer to computed overlay content bounds.
- Add E2E coverage for centered modal, edge-clamped popup, and nested overlay cases.

## Testing Checklist

After migration, verify:

- Component responds to clicks in normal layout.
- Component responds to clicks inside Modal / Portal.
- Click-outside behavior still works.
- Hover and focus behavior still works.
- Keyboard interaction is unchanged.
- HitMap debug logs show the expected target.

Useful commands:

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true go run ./examples/modal
go test ./ui/e2e -run Modal -count=1
go test ./ui/e2e -run Overlay -count=1
go test ./ui/components/button ./ui/components/modal ./ui/components/select -count=1
```

## Common Pitfalls

### Do Not Trust Stale Bounds After Layout Transforms

```go
// Avoid using stale local geometry as the only source of truth.
if mouseEvent.X >= c.bounds[0] && mouseEvent.X < c.bounds[0]+c.bounds[2] {
    // ...
}
```

### Keep A Fallback For Legacy Paths

Some tests or compatibility event paths may not have a populated HitMap. Keep the fallback until the component is fully moved to the current Action/Msg path.

### Do Not Use The Wrong Target

For overlay click-outside behavior, the HitMap target may be a child component, backdrop, trigger, popup, or unrelated content. Confirm target semantics before replacing existing geometry logic.

## Related Docs

- Historical event architecture: `../../docsArchive/cleanup-2026-05-19/docs/event/long_term_event_architecture.md`
- Historical pressed-state solution: `../../docsArchive/cleanup-2026-05-19/docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md`
- [../features/focus/README.md](../features/focus/README.md)
