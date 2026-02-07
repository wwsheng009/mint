# CSS-like Layout System Design

**Date**: 2025-01-07
**Goal**: Implement padding, margin, and text alignment similar to CSS box model

---

## CSS Box Model

```
┌─────────────────────────────────────────────┐
│                margin (外边距)               │
│  ┌───────────────────────────────────────┐  │
│  │           border (边框)               │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │        padding (内边距)          │  │
│  │  │  ┌───────────────────────────┐  │  │  │
│  │  │  │                           │  │  │  │
│  │  │  │       Content             │  │  │  │
│  │  │  │      (文本/组件)          │  │  │  │
│  │  │  │                           │  │  │  │
│  │  │  └───────────────────────────┘  │  │  │
│  │  │                                  │  │  │
│  │  └─────────────────────────────────┘  │  │
│  │                                         │  │
│  └───────────────────────────────────────┘  │
│                                             │
└─────────────────────────────────────────────┘
```

---

## Data Structures

### 1. Extended LayoutInfo

```go
// runtime/ui/layout_info.go
type LayoutInfo struct {
    // Existing fields
    Flex        int
    Align       Align
    CrossAlign  Align
    Gap         int
    FillHeight  bool
    StretchCross bool

    // ⭐ NEW: Padding (inner spacing)
    PaddingTop    int
    PaddingRight  int
    PaddingBottom int
    PaddingLeft   int

    // ⭐ NEW: Margin (outer spacing)
    MarginTop    int
    MarginRight  int
    MarginBottom int
    MarginLeft   int

    // ⭐ NEW: Text alignment for content
    TextAlign Align  // AlignStart, AlignCenter, AlignEnd
}
```

### 2. BoxConstraints Update

```go
// runtime/box_constraints.go
type BoxConstraints struct {
    MinWidth  int
    MaxWidth  int
    MinHeight int
    MaxHeight int

    // ⭐ NEW: Padding information
    InnerPadding struct {
        Left   int
        Right  int
        Top    int
        Bottom int
    }
}
```

---

## Layout Engine Changes

### 1. Calculate Positions with Padding and Margin

```go
// runtime/compute/engine.go

func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    mainAlign := layoutInfo.Align

    for i, child := range box.Children {
        childInfo := rtui.GetLayoutInfo(child.VNode)

        // Get margin
        marginLeft := childInfo.MarginLeft
        marginRight := childInfo.MarginRight

        // Calculate child X with margin
        childX := x + marginLeft

        // Apply main-axis alignment (if flex-stretched)
        alignedChildX := childX
        if child.NaturalWidth < child.Box.Width {
            switch mainAlign {
            case rtui.AlignCenter:
                padding := (child.Box.Width - child.NaturalWidth) / 2
                alignedChildX = childX + padding
            case rtui.AlignEnd:
                padding := child.Box.Width - child.NaturalWidth
                alignedChildX = childX + padding
            }
        }

        // Store text alignment for child to use
        if boundsAware, ok := child.VNode.(interface {
            SetBounds(int, int, int, int)
            SetTextAlign(rtui.Align)
        }); ok {
            boundsAware.SetBounds(alignedChildX, y, child.Box.Width, child.Box.Height)
            boundsAware.SetTextAlign(childInfo.TextAlign)  // ⭐ Pass text alignment
        }

        // Next child position: current + width + margins + gap
        x = alignedChildX + child.Box.Width + marginRight + layoutInfo.Gap
    }
}
```

### 2. Measure with Padding

```go
// components/button/button.go

func (b *ButtonVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // Get padding
    padding := b.getPadding()
    horizontalPadding := padding.Left + padding.Right
    verticalPadding := padding.Top + padding.Bottom

    // Calculate content size (text only)
    labelWidth := utf8.RuneCountInString(b.label)
    contentWidth := labelWidth + 3  // brackets + focus indicator
    contentHeight := 1

    // Add padding to get total size
    totalWidth := contentWidth + horizontalPadding
    totalHeight := contentHeight + verticalPadding

    // Apply constraints
    if totalWidth < constraints.MinWidth {
        totalWidth = constraints.MinWidth
    }
    if totalWidth > constraints.MaxWidth {
        totalWidth = constraints.MaxWidth
    }

    return runtime.Size{Width: totalWidth, Height: totalHeight}
}
```

---

## Component Changes

### 1. Button with Padding and Text Alignment

```go
// components/button/button.go

type ButtonVNode struct {
    *ui.ElementVNode
    label    string
    // ... existing fields ...

    // ⭐ NEW: Padding and text alignment
    paddingTop    int
    paddingRight  int
    paddingBottom int
    paddingLeft   int
    textAlign     rtui.Align
}

// Paint with padding and text alignment
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // Get padding
    paddingLeft := b.paddingLeft
    paddingRight := b.paddingRight

    // Build content text
    buttonText := focusIndicator + labelText
    contentWidth := len(buttonText)

    // Get layout width
    layoutWidth := contentWidth + paddingLeft + paddingRight
    if b.bounds[2] > 0 {
        layoutWidth = b.bounds[2]
    }

    // Apply text alignment if stretched
    if layoutWidth > contentWidth + paddingLeft + paddingRight {
        availableSpace := layoutWidth - paddingLeft - paddingRight - contentWidth

        switch b.textAlign {
        case rtui.AlignCenter:
            // Center text in available space
            leftPadding := paddingLeft + availableSpace/2
            rightPadding := paddingRight + (availableSpace - availableSpace/2)
            buttonText = strings.Repeat(" ", leftPadding) + buttonText +
                         strings.Repeat(" ", rightPadding)

        case rtui.AlignEnd:
            // Right-align: put all space on left
            leftPadding := paddingLeft + availableSpace
            buttonText = strings.Repeat(" ", leftPadding) + buttonText

        case rtui.AlignStart:
            // Left-align: put all space on right
            buttonText = buttonText + strings.Repeat(" ", paddingRight + availableSpace)
        }
    } else {
        // Not stretched: just apply padding
        if paddingLeft > 0 {
            buttonText = strings.Repeat(" ", paddingLeft) + buttonText
        }
        if paddingRight > 0 {
            buttonText = buttonText + strings.Repeat(" ", paddingRight)
        }
    }

    return []paint.DrawCmd{
        paint.NewTextCmd(x, y, buttonText, buttonStyle),
    }
}
```

---

## API Design

### 1. Padding Methods

```go
// Padding sets all sides
func (b *ButtonBuilderType) Padding(p int) *ButtonBuilderType {
    b.node.paddingTop = p
    b.node.paddingRight = p
    b.node.paddingBottom = p
    b.node.paddingLeft = p
    return b
}

// PaddingV sets vertical padding
func (b *ButtonBuilderType) PaddingV(top, bottom int) *ButtonBuilderType {
    b.node.paddingTop = top
    b.node.paddingBottom = bottom
    return b
}

// PaddingH sets horizontal padding
func (b *ButtonBuilderType) PaddingH(left, right int) *ButtonBuilderType {
    b.node.paddingLeft = left
    b.node.paddingRight = right
    return b
}

// PaddingTop, PaddingRight, PaddingBottom, PaddingLeft
func (b *ButtonBuilderType) PaddingTop(p int) *ButtonBuilderType {
    b.node.paddingTop = p
    return b
}

func (b *ButtonBuilderType) PaddingRight(p int) *ButtonBuilderType {
    b.node.paddingRight = p
    return b
}

func (b *ButtonBuilderType) PaddingBottom(p int) *ButtonBuilderType {
    b.node.paddingBottom = p
    return b
}

func (b *ButtonBuilderType) PaddingLeft(p int) *ButtonBuilderType {
    b.node.paddingLeft = p
    return b
}
```

### 2. Margin Methods

```go
// Margin sets all sides
func (b *ButtonBuilderType) Margin(m int) *ButtonBuilderType {
    b.node.SetProp("marginTop", m)
    b.node.SetProp("marginRight", m)
    b.node.SetProp("marginBottom", m)
    b.node.SetProp("marginLeft", m)
    return b
}

// MarginV sets vertical margin
func (b *ButtonBuilderType) MarginV(top, bottom int) *ButtonBuilderType {
    b.node.SetProp("marginTop", top)
    b.node.SetProp("marginBottom", bottom)
    return b
}

// MarginH sets horizontal margin
func (b *ButtonBuilderType) MarginH(left, right int) *ButtonBuilderType {
    b.node.SetProp("marginLeft", left)
    b.node.SetProp("marginRight", right)
    return b
}

// Individual margin methods
func (b *ButtonBuilderType) MarginTop(m int) *ButtonBuilderType {
    b.node.SetProp("marginTop", m)
    return b
}

func (b *ButtonBuilderType) MarginRight(m int) *ButtonBuilderType {
    b.node.SetProp("marginRight", m)
    return b
}

func (b *ButtonBuilderType) MarginBottom(m int) *ButtonBuilderType {
    b.node.SetProp("marginBottom", m)
    return b
}

func (b *ButtonBuilderType) MarginLeft(m int) *ButtonBuilderType {
    b.node.SetProp("marginLeft", m)
    return b
}
```

### 3. Text Alignment

```go
// TextAlign sets text alignment within the button
func (b *ButtonBuilderType) TextAlign(align rtui.Align) *ButtonBuilderType {
    b.node.textAlign = align
    return b
}
```

---

## Usage Examples

### Example 1: Button with Padding

```go
// CSS: .btn { padding: 10px; }
Button("Click Me").
    Padding(2).  // 2 spaces padding on all sides
    Build()

// Result: "  [ Click Me ]  "
```

### Example 2: Button with Different Padding

```go
// CSS: .btn { padding: 2px 10px; }
Button("Click Me").
    PaddingV(1, 1).   // top/bottom: 1 space
    PaddingH(3, 3).   // left/right: 3 spaces
    Build()

// Result: "   [ Click Me ]   "
```

### Example 3: Button with Margin and Centered Text

```go
// CSS: .btn { margin: 10px; text-align: center; }
Button("Click Me").
    Margin(2).              // 2 spaces margin
    TextAlign(rtui.AlignCenter).
    FillWidth().            // Stretch to fill
    Build()
```

### Example 4: Complex Layout

```go
// CSS flex-like layout
HStack(
    Button("Left").
        Padding(1).
        TextAlign(rtui.AlignStart).
        Flex(1),

    Button("Center").
        Padding(2).
        TextAlign(rtui.AlignCenter).
        Flex(1),

    Button("Right").
        Padding(1).
        TextAlign(rtui.AlignEnd).
        Flex(1),
).
    Gap(1).
    Align(rtui.AlignCenter).  // Container alignment
    Build()
```

**Visual Result:**
```
│ [Left]     [  Center  ]      [Right]│
│ ↑left       ↑center           ↑right │
```

---

## Implementation Steps

### Phase 1: Data Structures (1 hour)
1. ✅ Extend `LayoutInfo` with padding/margin fields
2. ✅ Update `GetLayoutInfo()` to read padding/margin from props
3. ✅ Add padding fields to `ButtonVNode`

### Phase 2: Layout Engine (2 hours)
1. ✅ Modify `layoutHStack()` to account for margins
2. ✅ Modify `layoutVStack()` to account for margins
3. ✅ Update `Measure()` to include padding in size calculation
4. ✅ Pass text alignment through `SetTextAlign()` (if interface exists)

### Phase 3: Component Implementation (2 hours)
1. ✅ Add padding methods to `ButtonBuilderType`
2. ✅ Add margin methods to `ButtonBuilderType`
3. ✅ Update `Button.Measure()` to include padding
4. ✅ Update `Button.Paint()` to apply padding and text alignment

### Phase 4: Testing (1 hour)
1. ✅ Test padding-only buttons
2. ✅ Test margin-only buttons
3. ✅ Test padding + margin + text alignment
4. ✅ Test flex layout with padding/margin

### Phase 5: Documentation (1 hour)
1. ✅ Write usage guide
2. ✅ Create examples
3. ✅ Update API documentation

---

## CSS Comparison Table

| CSS Property            | Mint TUI API                         | Effect                              |
|-------------------------|--------------------------------------|-------------------------------------|
| `padding: 10px`         | `.Padding(2)`                        | 2 spaces on all sides               |
| `padding: 5px 10px`     | `.PaddingV(1, 1).PaddingH(2, 2)`     | Vertical 1, horizontal 2            |
| `padding-left: 10px`    | `.PaddingLeft(2)`                    | 2 spaces on left                    |
| `margin: 10px`          | `.Margin(2)`                         | 2 spaces margin on all sides        |
| `margin-left: 10px`     | `.MarginLeft(2)`                     | 2 spaces margin on left             |
| `text-align: center`    | `.TextAlign(ui.AlignCenter)`         | Center text within button           |
| `text-align: right`     | `.TextAlign(ui.AlignEnd)`            | Right-align text                    |
| `flex: 1`               | `.Flex(1)` (via props)               | Grow to fill space                  |
| `justify-content: center` | `HStack().Align(ui.AlignCenter)`   | Center children in container        |

---

## Benefits

1. ✅ **Familiar API** - Developers familiar with CSS will recognize the pattern
2. ✅ **Flexible** - Support for complex layouts
3. ✅ **Consistent** - Same pattern applies to all components
4. ✅ **Backward Compatible** - Default values (0) don't break existing code
5. ✅ **Composable** - Can combine padding, margin, and text alignment

---

## Migration Guide

### Before (No Padding)
```go
Button("Click Me").Build()
// Result: "[Click Me]"
```

### After (With Padding)
```go
Button("Click Me").
    Padding(1).
    Build()
// Result: " [ Click Me ] "
```

---

**Implementation Time**: 6-7 hours
**Breaking Changes**: None (all new features)
**Backward Compatibility**: 100%
