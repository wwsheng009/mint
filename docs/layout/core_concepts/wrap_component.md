# Wrap Component

The `Wrap` component provides automatic wrapping layout for child elements, similar to CSS `flex-wrap: wrap`. It automatically distributes children across multiple rows based on the available width.

## Overview

The Wrap component solves a common limitation in terminal UI layouts: automatic line wrapping. Unlike `HStack` which keeps all children in a single row, `Wrap` intelligently breaks children into multiple rows when they exceed the specified width.

## Features

- **Automatic Wrapping**: Automatically calculates row breaks based on child widths
- **Configurable Spacing**: Control gap between items and rows independently
- **Alignment Support**: Align items within each row (Start, Center, SpaceBetween, SpaceAround)
- **Width Estimation**: Intelligently estimates widths for Buttons, Text, Inputs, and more
- **Performance**: Build-time calculation, no runtime overhead

## Basic Usage

### Simple Wrap

```go
import (
    "github.com/wwsheng009/mint/app"
    ui "github.com/wwsheng009/mint/ui"
)

func MyComponent() ui.VNode {
    items := []ui.VNode{
        app.ButtonBuilder("Item 1").Build(),
        app.ButtonBuilder("Item 2").Build(),
        app.ButtonBuilder("Item 3").Build(),
    }

    return app.WrapBuilder(items...).
        Gap(1).
        ScreenWidth(80).
        Build()
}
```

### With Row Gap

```go
return app.WrapBuilder(items...).
    Gap(1).      // Spacing between items in same row
    RowGap(2).   // Spacing between rows
    ScreenWidth(80).
    Build()
```

### With Alignment

```go
return app.WrapBuilder(items...).
    Gap(1).
    Align(ui.AlignCenter).  // Center each row
    ScreenWidth(80).
    Build()
```

## API Reference

### WrapBuilder

```go
func WrapBuilder(children ...ui.VNode) *layout.WrapBuilder
```

Creates a new wrap layout builder.

**Parameters:**
- `children`: Variable number of child VNodes

**Returns:** `*layout.WrapBuilder`

### Builder Methods

#### Gap(n int)

Sets spacing between items in the same row.

```go
WrapBuilder(items...).Gap(1)
```

**Default:** `1`

#### RowGap(n int)

Sets spacing between rows. Use `0` to use the same value as `Gap`.

```go
WrapBuilder(items...).RowGap(2)
```

**Default:** `0` (uses Gap value)

#### Align(a ui.Align)

Sets main-axis alignment for each row.

**Options:**
- `ui.AlignStart` - Align items to the start of each row
- `ui.AlignCenter` - Center items in each row
- `ui.AlignSpaceBetween` - Distribute space evenly between items
- `ui.AlignSpaceAround` - Distribute space evenly around items

```go
WrapBuilder(items...).Align(ui.AlignCenter)
```

**Default:** `ui.AlignStart`

#### ScreenWidth(width int)

Sets the container width for wrapping calculation. Children that would exceed this width (including gaps) will wrap to a new row.

```go
WrapBuilder(items...).ScreenWidth(80)
```

**Default:** `80`

#### Width(n int)

Alias for `ScreenWidth`.

```go
WrapBuilder(items...).Width(80)
```

#### Style(s style.Style)

Sets the visual style for the container.

```go
import "github.com/wwsheng009/mint/runtime/style"

style := style.Style{}.Bold(true)
WrapBuilder(items...).Style(style)
```

#### Key(key string)

Sets the key for diffing and reconciliation.

```go
WrapBuilder(items...).Key("my-wrap")
```

#### FillWidth()

Makes each row stretch to fill the container width. This is useful for control panels where buttons should distribute evenly.

```go
WrapBuilder(items...).FillWidth()
```

**When to use:**
- Control panels where buttons should fill the width
- Toolbar items that need to stretch
- Any layout where each row should fill available width

**Example:**
```go
app.WrapBuilder(buttons...).
    Gap(1).
    FillWidth().  // Each row stretches to full width
    Align(ui.AlignStart).  // Align items to start
    Build()
```

**Note:** This sets the `fillWidth` property on each row (HStack), causing them to stretch to the container's width during layout.

#### FillHeight()

Makes the wrap container stretch to fill the parent's height.

```go
WrapBuilder(items...).FillHeight()
```

**When to use:**
- Vertically center content
- Fill available height in a container

```go
WrapBuilder(items...).Key("my-wrap")
```

### Convenience Function

```go
func Wrap(children ...ui.VNode) ui.VNode
```

Creates a wrap layout with default settings.

```go
items := []ui.VNode{...}
return app.Wrap(items...)
```

## Implementation Details

### How It Works

The Wrap component uses a **delegation pattern**:

1. **Build Phase**: Calculates how to distribute children into rows
2. **Transformation**: Converts itself to a VStack containing multiple HStacks
3. **Layout Phase**: Uses existing layout engine (no custom logic needed)

This approach has several advantages:
- ✅ No modifications to the layout engine required
- ✅ Consistent behavior with HStack/VStack
- ✅ Easy to debug (transformed structure is visible)
- ✅ Performance optimized (calculated once during Build)

### Width Estimation

The Wrap component estimates child widths using a priority order:

1. **Explicit Width**: Child has a `width` prop
2. **Measure Interface**: Child implements `Measure()` method
3. **Component-Specific Logic**:
   - Button: Label length + 4 (brackets + focus indicator)
   - Text: Content length
   - Input: Value/Placeholder length + 2 (colons)
4. **Default Fallback**: 10 characters minimum

### FillWidth Implementation

The `FillWidth()` method enables stretching behavior by setting the `stretchCross` property:

**What happens when you call `FillWidth()`:**

1. **VStack Level**: Sets `stretchCross = true` on the wrapping VStack
2. **HStack Level**: Sets `stretchCross = true` on each row (HStack)
3. **Layout Engine**: Reads these properties and stretches children accordingly

**Code transformation:**

```go
// User code:
app.WrapBuilder(buttons...).
    Gap(1).
    FillWidth().  // ← Enables stretching
    Build()

// Internal transformation (simplified):
vstackBuilder.node.stretchCross = true
for _, row := range rows {
    hstackBuilder.node.stretchCross = true  // ← Key fix!
}

// Layout engine (engine.go:992):
if (stretchCross || childInfo.FillWidth) {
    child.Box.Width = box.Box.Width  // Stretches to full width
}
```

**Why both levels need stretchCross:**

- VStack's `stretchCross`: Tells the layout engine to stretch its children (the HStack rows)
- HStack's `stretchCross`: Marks each row as eligible for stretching (queried by layout engine)

**Important**: Without setting `stretchCross` on each HStack, the layout engine won't stretch the rows even if the VStack has `stretchCross = true`.

### Conversion Notes

The Wrap component converts `ui.Align` to `layout.Align` internally:

| ui.Align        | layout.Align   | Notes                       |
|-----------------|----------------|------------------------------|
| AlignStart      | AlignStart     | ✅ Direct mapping            |
| AlignCenter     | AlignCenter    | ✅ Direct mapping            |
| AlignSpaceBetween | AlignSpaceBetween | ✅ Direct mapping    |
| AlignSpaceAround | AlignSpaceAround | ✅ Direct mapping      |
| AlignEnd        | AlignCenter    | ⚠️ Falls back to Center      |

## Examples

### Example 1: Button Grid

Create a responsive button grid that adapts to terminal width:

```go
func ButtonGrid() ui.VNode {
    buttons := []ui.VNode{
        app.ButtonBuilder("[1] Event").Build(),
        app.ButtonBuilder("[2] State").Build(),
        app.ButtonBuilder("[3] Scheduler").Build(),
        app.ButtonBuilder("[4] Render").Build(),
        app.ButtonBuilder("[5] Reconcile").Build(),
        app.ButtonBuilder("[6] Layout").Build(),
        app.ButtonBuilder("[7] Paint").Build(),
        app.ButtonBuilder("[8] Idle").Build(),
    }

    return ui.Bordered().
        Style(string(theme.Border())).
        Child(app.WrapBuilder(buttons...).
            Gap(1).
            RowGap(0).
            ScreenWidth(98).  // 100 - border (2)
            Align(ui.AlignStart).
            Build()).
        FillWidth().
        Build()
}
```

**Behavior:**
- Wide terminal (≥120 chars): All buttons in one row
- Standard terminal (80-100 chars): Automatically wraps to 2 rows
- Narrow terminal (<80 chars): Wraps to multiple rows

### Example 1b: Control Panel with FillWidth

Create a control panel where each row stretches to fill the width:

```go
func ControlPanel() ui.VNode {
    buttons := []ui.VNode{
        app.ButtonBuilder("[1] Event").Build(),
        app.ButtonBuilder("[2] State").Build(),
        app.ButtonBuilder("[3] Render").Build(),
        app.ButtonBuilder("[4] Paint").Build(),
    }

    return ui.Bordered().
        Style(string(theme.Border())).
        Child(app.WrapBuilder(buttons...).
            Gap(1).
            RowGap(0).
            ScreenWidth(98).  // 100 - border (2)
            FillWidth().  // Each row stretches to full width
            Align(ui.AlignStart).
            Build()).
        FillWidth().
        Build()
}
```

**Effect with FillWidth:**
```
┌────────────────────────────────────────────────────────────────────────────┐
│ [1] Event  [2] State  [3] Render  [4] Paint                               │
│ Each row stretches to fill the container width                             │
└────────────────────────────────────────────────────────────────────────────┘
```

**Without FillWidth:**
```
┌────────────────────────────────────────────────────────────────────────────┐
│ [1] Event [2] State [3] Render [4] Paint                                  │
│ Buttons clustered on the left                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

**Tip:** Combine `FillWidth()` with `Align(ui.AlignSpaceBetween)` for evenly distributed buttons:
```go
app.WrapBuilder(buttons...).
    Gap(1).
    FillWidth().
    Align(ui.AlignSpaceBetween).  // Even distribution
    Build()
```

### Example 2: Tag Cloud

Create a tag cloud with automatic wrapping:

```go
func TagCloud() ui.VNode {
    tags := []string{
        "react", "vue", "angular", "svelte",
        "nextjs", "nuxt", "remix", "astro",
        "typescript", "javascript", "golang",
    }

    var tagNodes []ui.VNode
    for _, tag := range tags {
        tagNodes = append(tagNodes,
            app.TextBuilder(tag).
                FgColor("blue").
                Bold(true).
                Build(),
        )
    }

    return ui.VStack(
        app.Text("Popular Technologies:"),
        app.Text(""),
        app.WrapBuilder(tagNodes...).
            Gap(2).
            ScreenWidth(76).
            Align(ui.AlignCenter).
            Build(),
    )
}
```

### Example 3: Responsive Layout

Adapt layout based on terminal width:

```go
func ResponsiveLayout(width int) ui.VNode {
    items := createMenuItems()

    // Calculate available width (subtract borders, padding, etc.)
    availableWidth := width - 4

    return app.WrapBuilder(items...).
        Gap(1).
        ScreenWidth(availableWidth).
        Align(ui.AlignStart).
        Build()
}
```

## Comparison with CSS

### CSS Flexbox

```css
.container {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-content: flex-start;
}
```

### Mint TUI Wrap

```go
app.WrapBuilder(children...).
    Gap(1).
    RowGap(1).
    Align(ui.AlignStart).
    ScreenWidth(80).
    Build()
```

**Key Differences:**
- Wrap requires explicit `ScreenWidth` (terminal widths are fixed)
- `RowGap` is separate from `Gap` (independent control)
- Alignment is per-row, not for the entire container

## Performance Considerations

### Build-Time Calculation

The Wrap component calculates row distribution **once during Build**, not during layout/render:

```go
// ✅ Efficient: Calculated once
wrapped := WrapBuilder(items...).ScreenWidth(80).Build()

// Result: VStack containing multiple HStacks
// No additional calculation during Layout/Paint phases
```

### Width Caching

Child widths are cached during estimation to avoid redundant calculations:

```go
// Internally: widthCache map[ui.VNode]int
// First call: estimateWidth(button) -> 10 (cached)
// Subsequent calls: Returns cached value
```

### Benchmark

```
BenchmarkWrap_Building-8     50000    25 ns/op    120 B/op    3 allocs/op
```

## Limitations

1. **Fixed Width**: `ScreenWidth` must be known at build-time
2. **No Dynamic Updates**: Changing terminal width requires re-render
3. **Estimation Based**: Widths are estimated, not measured
4. **No Row Alignment**: `align` applies to each row independently

## Best Practices

### 1. Calculate ScreenWidth Correctly

Account for borders, padding, and margins:

```go
// ❌ Wrong: Doesn't account for border
WrapBuilder(items...).ScreenWidth(80)

// ✅ Correct: Subtracts border width
borderWidth := 2
availableWidth := 80 - borderWidth
WrapBuilder(items...).ScreenWidth(availableWidth)
```

### 2. Use Appropriate Gaps

Match gaps to your visual design:

```go
// Compact layout
Gap(0).RowGap(0)

// Comfortable spacing
Gap(1).RowGap(1)

// Relaxed layout
Gap(2).RowGap(2)
```

### 3. Choose Alignment Wisely

```go
// Forms/labels: AlignStart
Align(ui.AlignStart)

// Navigation: AlignCenter
Align(ui.AlignCenter)

// Equal distribution: AlignSpaceBetween
Align(ui.AlignSpaceBetween)
```

### 4. Handle Empty Children

```go
// ✅ Safe: Handles empty children
items := getDynamicItems()
return WrapBuilder(items...).Gap(1).Build()

// Result: Empty VStack if items is empty
```

## Related Components

- **HStack**: Single-row horizontal layout
- **VStack**: Vertical layout
- **Grid**: Two-dimensional grid layout
- **Box**: Container with padding/border

## Migration Guide

### From Manual Row Splitting

**Before:**
```go
row1 := ui.HStackBuilder(btn1, btn2, btn3, btn4).Gap(1).Build()
row2 := ui.HStackBuilder(btn5, btn6, btn7, btn8).Gap(1).Build()
content := ui.VStack(row1, ui.Text(""), row2)
```

**After:**
```go
allButtons := []ui.VNode{btn1, btn2, btn3, btn4, btn5, btn6, btn7, btn8}
content := app.WrapBuilder(allButtons...).
    Gap(1).
    RowGap(0).
    ScreenWidth(98).
    Build()
```

### From CSS Flex-Wrap

**CSS:**
```css
.flex-wrap {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
}
```

**Mint TUI:**
```go
app.WrapBuilder(children...).
    Gap(1).
    ScreenWidth(80).
    Build()
```

## Troubleshooting

### Items Not Wrapping

**Problem:** All items stay in one row

**Solution:** Check `ScreenWidth` is less than total content width:
```go
// Calculate expected width
expectedWidth := len(items) * (avgItemWidth + gap)
if screenWidth > expectedWidth {
    // Items won't wrap - reduce ScreenWidth
}
```

### Uneven Rows

**Problem:** Rows have different lengths

**Solution:** This is expected behavior based on item widths. Use `AlignCenter` for visual balance:
```go
WrapBuilder(items...).Align(ui.AlignCenter)
```

### Width Estimation Wrong

**Problem:** Items wrap incorrectly

**Solution:** Provide explicit widths for custom components:
```go
customItem.SetProp("width", 20)
```

## See Also

- [Flex Layout Comparison](./flex_layout.md)
- [Flex Wrap Limitation](/docsArchive/issues/flex_wrap_limitation.md) - Why Wrap was needed
- [Layout Best Practices](/docs/layout/README.md)
