# Layout Fix Plan for Demo1

## Problem Statement

The current HStack/VStack layout system cannot achieve the expected layout design from `framework/docs/ui/demo/demo1.md`.

### Expected Layout:
```
+--------------------------------------------------+
| TUI Engine Demo        [Open Modal] Clicks: 2 |
+-----------+--------------------------------------+
| Menu      | [ Input box....................... ] |
| Add Count |--------------------------------------|
| Quit      | Log line #0                          |
|           | Log line #1                          |
+-----------+--------------------------------------+
```

### Actual Layout (current):
The Sidebar and ContentArea render independently. HStack paints all children at the same Y position, creating a fundamental mismatch.

## Root Cause

**HStack behavior**: Paints all children at the same Y coordinate, side by side horizontally.

**VStack behavior**: Paints children at different Y coordinates, stacked vertically.

**The issue**: The expected layout requires **row-based rendering** where each row has different content in different columns:
- Row 1: Sidebar="Menu", ContentArea="Input box"
- Row 2: Sidebar="Add Count", ContentArea="separator"
- Row 3: Sidebar="Quit", ContentArea="Log line #0"

Current HStack/VStack cannot express this because:
1. HStack renders all its children at once (same Y)
2. When VBox contains HStack(Sidebar, ContentArea), each renders independently
3. There's no coordination between Sidebar's row N and ContentArea's row N

## Possible Solutions

### Option 1: Pre-formatted Text Blocks (Quick Win)
**Approach**: Use pre-built text strings for complex layouts instead of component composition.

**Pros**:
- Simplest implementation
- Works immediately with existing system
- No architectural changes

**Cons**:
- Loses component reusability
- Hard to maintain
- Not scalable for dynamic content

**Effort**: Low
**Recommendation**: Use as temporary workaround

---

### Option 2: Row-Based Layout System
**Approach**: Implement a `Table` or `Grid` component that renders row-by-row.

```go
ui.Table(
    ui.Row(
        ui.Cell("Menu"),
        ui.Cell("[Input box]"),
    ),
    ui.Row(
        ui.Cell("Add Count"),
        ui.Cell("----------"),
    ),
    ui.Row(
        ui.Cell("Quit"),
        ui.Cell("Log line #0"),
    ),
)
```

**Pros**:
- Matches expected layout pattern
- Declarative and composable
- Scalable for dynamic content

**Cons**:
- New component to implement
- Needs row height synchronization
- Complex measurement logic

**Effort**: Medium
**Recommendation**: **Primary solution for long-term**

---

### Option 3: Layout Rows Abstraction
**Approach**: Add `LayoutRows` function that coordinates column rendering.

```go
ui.LayoutRows(
    []LayoutRow{
        {Height: 1, Content: func(y int) VNode {
            return ui.HStack(SidebarRow(y), ContentRow(y))
        }},
        // ...
    },
)
```

**Pros**:
- More flexible than Table
- Can handle variable row heights

**Cons**:
- More complex API
- Harder to use than Table

**Effort**: Medium-High
**Recommendation**: Consider if Table is insufficient

---

### Option 4: Constrain-Based Layout
**Approach**: Implement constraints like "Sidebar.bottom == ContentArea.bottom"

**Pros**:
- Most flexible
- Industry standard (AutoLayout, CSS Grid)

**Cons**:
- Highest complexity
- Overkill for current needs
- Long implementation time

**Effort**: High
**Recommendation**: Future consideration

## Recommended Action Plan

### Phase 1: Immediate (Workaround)
1. Use pre-formatted text blocks for demo1
2. Document the limitation

### Phase 2: Short-term (2-3 days)
1. Implement `Table` component in `runtime/ui/layout.go`
2. Add `TableRow`, `TableCell` types
3. Implement row-by-row rendering logic

### Phase 3: Refactor
1. Migrate demo1 to use Table component
2. Update other demos as needed
3. Add tests for Table layout

## Files to Modify

1. **New**: `runtime/ui/table.go` - Table component implementation
2. **Modify**: `runtime/ui/layout.go` - Export Table types
3. **Modify**: `internal/render/declarative_node.go` - Add Table rendering logic
4. **Modify**: `examples/ui_demos/demo1_full_featured/main.go` - Use Table for main layout
