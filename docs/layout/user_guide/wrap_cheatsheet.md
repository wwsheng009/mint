# Wrap Component - Quick Reference

## Import

```go
import (
    "github.com/wwsheng009/mint/app"
    ui "github.com/wwsheng009/mint/ui"
)
```

## Basic Syntax

```go
app.WrapBuilder(children...).
    Gap(1).
    ScreenWidth(80).
    Build()
```

## Common Patterns

### 1. Simple Button Grid

```go
buttons := []ui.VNode{
    app.ButtonBuilder("Btn1").Build(),
    app.ButtonBuilder("Btn2").Build(),
    app.ButtonBuilder("Btn3").Build(),
}

return app.WrapBuilder(buttons...).
    Gap(1).
    ScreenWidth(76).
    Build()
```

### 2. With Row Gap

```go
app.WrapBuilder(items...).
    Gap(1).      // Horizontal spacing
    RowGap(2).   // Vertical spacing
    ScreenWidth(80).
    Build()
```

### 3. Centered Rows

```go
app.WrapBuilder(items...).
    Gap(1).
    Align(ui.AlignCenter).
    ScreenWidth(80).
    Build()
```

### 4. With Border

```go
ui.Bordered().
    Style(string(theme.Border())).
    Child(app.WrapBuilder(items...).
        Gap(1).
        ScreenWidth(98).  // 100 - border width
        Build()).
    Build()
```

### 5. With FillWidth (Stretched Rows)

```go
ui.Bordered().
    Style(string(theme.Border())).
    Child(app.WrapBuilder(buttons...).
        Gap(1).
        ScreenWidth(98).
        FillWidth().  // Each row stretches to full width
        Build()).
    Build()
```

### 6. With SpaceBetween (Even Distribution)

```go
app.WrapBuilder(buttons...).
    Gap(1).
    ScreenWidth(98).
    FillWidth().
    Align(ui.AlignSpaceBetween).  // Items evenly distributed
    Build()
```

## API Reference

| Method | Type | Default | Description |
|--------|------|---------|-------------|
| `Gap(n)` | int | 1 | Spacing between items in row |
| `RowGap(n)` | int | 0 | Spacing between rows (0 = use Gap) |
| `Align(a)` | ui.Align | AlignStart | Alignment for each row |
| `ScreenWidth(w)` | int | 80 | Container width for wrapping |
| `FillWidth()` | - | false | Make each row stretch to full width |
| `FillHeight()` | - | false | Make container stretch to parent height |
| `Style(s)` | style.Style | - | Visual style |
| `Key(k)` | string | - | Key for diffing |

## Alignment Options

```go
ui.AlignStart         // Items align to start
ui.AlignCenter        // Items centered
ui.AlignSpaceBetween  // Space between items
ui.AlignSpaceAround   // Space around items
```

## Width Calculation

```
Available Width = Total Width - Borders - Padding

// Example:
Terminal Width: 100
Border Width:   2
ScreenWidth:    98
```

## Common Use Cases

### Control Panel

```go
func ControlPanel() ui.VNode {
    buttons := []ui.VNode{
        app.ButtonBuilder("[1] Event").Build(),
        app.ButtonBuilder("[2] State").Build(),
        app.ButtonBuilder("[3] Render").Build(),
        app.ButtonBuilder("[4] Paint").Build(),
    }

    return ui.Bordered().
        Child(app.WrapBuilder(buttons...).
            Gap(1).
            RowGap(0).
            ScreenWidth(98).
            Build()).
        FillWidth().
        Build()
}
```

### Tag Cloud

```go
func TagCloud(tags []string) ui.VNode {
    var nodes []ui.VNode
    for _, tag := range tags {
        nodes = append(nodes,
            app.TextBuilder(tag).
                FgColor("blue").
                Bold(true).
                Build(),
        )
    }

    return app.WrapBuilder(nodes...).
        Gap(2).
        Align(ui.AlignCenter).
        ScreenWidth(76).
        Build()
}
```

## Tips

✅ **Always account for borders** when setting ScreenWidth
✅ **Use Gap(0)** for compact layouts
✅ **Use AlignCenter** for visual balance
✅ **Provide explicit widths** for custom components

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Items not wrapping | Reduce ScreenWidth |
| Uneven rows | Expected - use AlignCenter |
| Wrong wrap point | Add explicit width to child |

## Migration from HStack

**Before:**
```go
row1 := ui.HStackBuilder(items[0:4]).Gap(1).Build()
row2 := ui.HStackBuilder(items[4:8]).Gap(1).Build()
return ui.VStack(row1, row2)
```

**After:**
```go
return app.WrapBuilder(items...).
    Gap(1).
    ScreenWidth(80).
    Build()
```

## Related

- **HStack**: Single-row layout
- **VStack**: Vertical layout
- **Grid**: Two-dimensional layout

See full documentation: [wrap_component.md](../core_concepts/wrap_component.md)
