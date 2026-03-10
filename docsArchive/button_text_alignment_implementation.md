# Button Text Alignment Implementation

**Date**: 2025-01-07
**Goal**: Implement CSS flex-like text alignment for buttons

---

## CSS Flex Behavior

### justify-content Effects

When children are stretched by flex (e.g., `flex: 1`):

```
┌─────────────────────────────────────────────────────────┐
│  justify-content: flex-start (左对齐)                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ Button 1 │  │ Button 2 │  │ Button 3 │             │
│  └──────────┘  └──────────┘  └──────────┘             │
│  ↑ Text aligned to start (left)                        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  justify-content: center (居中对齐)                     │
│      ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│      │ Button 1 │  │ Button 2 │  │ Button 3 │         │
│      └──────────┘  └──────────┘  └──────────┘         │
│      ↑ Text centered within allocated space            │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  justify-content: flex-end (右对齐)                     │
│             ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│             │ Button 1 │  │ Button 2 │  │ Button 3 │  │
│             └──────────┘  └──────────┘  └──────────┘  │
│                          ↑ Text aligned to end (right) │
└─────────────────────────────────────────────────────────┘
```

---

## Implementation Strategy

### 1. Button Component

Add `textAlign` property to control text alignment within button:

```go
// ButtonVNode properties
// - align: ui.Align (text alignment: Start/Center/End)
// - Default: AlignStart (left-aligned)
```

### 2. Button.Paint() Logic

```go
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // Calculate natural width
    buttonText := focusIndicator + labelText
    naturalWidth := len(buttonText)

    // Get layout width (from bounds set by layout engine)
    layoutWidth := naturalWidth  // Default
    if b.bounds[2] > 0 {
        layoutWidth = b.bounds[2]  // Allocated width from flex
    }

    // Get text alignment from props
    text_align := b.getTextAlign()  // Default: AlignStart

    // Add padding based on alignment
    if layoutWidth > naturalWidth {
        padding := layoutWidth - naturalWidth

        switch text_align {
        case ui.AlignStart:
            // Left-aligned: add padding to right
            buttonText += strings.Repeat(" ", padding)

        case ui.AlignCenter:
            // Centered: distribute padding on both sides
            leftPadding := padding / 2
            rightPadding := padding - leftPadding
            buttonText = strings.Repeat(" ", leftPadding) + buttonText +
                         strings.Repeat(" ", rightPadding)

        case ui.AlignEnd:
            // Right-aligned: add padding to left
            buttonText = strings.Repeat(" ", padding) + buttonText
        }
    }

    return []paint.DrawCmd{
        paint.NewTextCmd(x, y, buttonText, buttonStyle),
    }
}
```

### 3. Layout Engine

**No changes needed** - layout engine already:
1. Calculates flex allocation
2. Sets bounds with allocated width
3. Applies container-level alignment (space-between, space-around, etc.)

---

## API Usage

### Example 1: Left-aligned buttons (default)

```go
WrapBuilder(
    Button("Button1"),
    Button("Button2"),
    Button("Button3"),
).
    Gap(1).
    Align(ui.AlignStart).  // Container alignment
    FillWidth().
    Build()
```

**Result**:
```
│>[Button1]   >[Button2]   >[Button3]   │
│↑left        ↑left        ↑left        │
```

### Example 2: Centered text buttons

```go
WrapBuilder(
    Button("Button1").TextAlign(ui.AlignCenter),
    Button("Button2").TextAlign(ui.AlignCenter),
    Button("Button3").TextAlign(ui.AlignCenter),
).
    Gap(1).
    Align(ui.AlignCenter).  // Container centers children
    FillWidth().
    Build()
```

**Result**:
```
│  >[Button1]   >[Button2]   >[Button3]  │
│ ↑centered     ↑centered    ↑centered   │
```

### Example 3: Right-aligned text buttons

```go
WrapBuilder(
    Button("Button1").TextAlign(ui.AlignEnd),
    Button("Button2").TextAlign(ui.AlignEnd),
    Button("Button3").TextAlign(ui.AlignEnd),
).
    Gap(1).
    Align(ui.AlignEnd).  // Container aligns to end
    FillWidth().
    Build()
```

**Result**:
```
│   >[Button1]   >[Button2]   >[Button3]│
│        ↑right        ↑right       ↑right│
```

---

## Implementation Steps

### Step 1: Add TextAlign method to ButtonBuilder

**File**: `components/button/button.go`

```go
// TextAlign sets the text alignment within the button
func (bb *ButtonBuilder) TextAlign(align ui.Align) *ButtonBuilder {
    bb.node.SetProp("textAlign", int(align))
    return bb
}
```

### Step 2: Add getTextAlign helper to ButtonVNode

**File**: `components/button/button.go`

```go
// getTextAlign returns the text alignment, defaulting to AlignStart
func (b *ButtonVNode) getTextAlign() ui.Align {
    if props := b.Props(); props != nil {
        if align, ok := props["textAlign"].(int); ok {
            return ui.Align(align)
        }
    }
    return ui.AlignStart  // Default: left-aligned
}
```

### Step 3: Update Paint() with alignment logic

**File**: `components/button/button.go`

```go
// Build the button text
buttonText := focusIndicator + labelText
naturalWidth := len(buttonText)

// Get layout width from bounds
layoutWidth := naturalWidth
if b.bounds[2] > 0 {
    layoutWidth = b.bounds[2]
}

// Apply text alignment if button is stretched
if layoutWidth > naturalWidth {
    textAlign := b.getTextAlign()
    padding := layoutWidth - naturalWidth

    switch textAlign {
    case ui.AlignCenter:
        // Center text: distribute padding on both sides
        leftPadding := padding / 2
        rightPadding := padding - leftPadding
        buttonText = strings.Repeat(" ", leftPadding) + buttonText +
                     strings.Repeat(" ", rightPadding)

    case ui.AlignEnd:
        // Right-align: add padding to left
        buttonText = strings.Repeat(" ", padding) + buttonText

    case ui.AlignStart:
        // Left-align: add padding to right (default)
        buttonText += strings.Repeat(" ", padding)
    }
}
```

### Step 4: Import strings package

**File**: `components/button/button.go`

```go
import (
    "fmt"
    "strings"  // ← Add this
    "unicode/utf8"
    ...
)
```

---

## Visual Examples

### Left-aligned (AlignStart)

```
Allocated: 19 chars, Natural: 14 chars
┌───────────────────┐
│>* [1] Event       │  ← 5 spaces on right
└───────────────────┘
```

### Centered (AlignCenter)

```
Allocated: 19 chars, Natural: 14 chars
┌───────────────────┐
│  >* [1] Event     │  ← 2 spaces left, 3 spaces right
└───────────────────┘
```

### Right-aligned (AlignEnd)

```
Allocated: 19 chars, Natural: 14 chars
┌───────────────────┐
│       >* [1] Event│  ← 5 spaces on left
└───────────────────┘
```

---

## Benefits

1. ✅ **Matches CSS flex behavior** - Developers familiar with CSS will recognize the behavior
2. ✅ **Flexible text alignment** - Each button can have different alignment
3. ✅ **Container alignment works independently** - space-between, space-around, etc.
4. ✅ **Backward compatible** - Default behavior (AlignStart) matches current implementation
5. ✅ **Simple API** - ButtonBuilder.TextAlign(align)

---

## Testing

```go
func TestButtonTextAlignment(t *testing.T) {
    // Test left-aligned (default)
    btn1 := Button("Test").FillWidth()
    // Expected: "[Test]     " (padding on right)

    // Test centered
    btn2 := Button("Test").TextAlign(ui.AlignCenter).FillWidth()
    // Expected: "  [Test]   " (padding on both sides)

    // Test right-aligned
    btn3 := Button("Test").TextAlign(ui.AlignEnd).FillWidth()
    // Expected: "     [Test]" (padding on left)
}
```

---

**Implementation Time**: ~30 minutes
**Files to Modify**:
- `components/button/button.go` (add TextAlign, update Paint)
**API Changes**:
- New: `ButtonBuilder.TextAlign(align ui.Align)`
**Breaking Changes**: None
