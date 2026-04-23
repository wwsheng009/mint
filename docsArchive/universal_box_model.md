# Universal Box Model Implementation

**Date**: 2025-01-07
**Concept**: CSS-like padding/margin for ALL components

---

## Architecture Principle

**CSS Box Model applies to ALL elements**, not just specific components:

```css
/* In CSS, ANY element can have padding/margin */
div { padding: 10px; margin: 20px; }
button { padding: 10px; margin: 20px; }
input { padding: 10px; margin: 20px; }
span { padding: 10px; margin: 20px; }
```

**Mint TUI should follow the same pattern:**

```go
// ALL VNodes should support padding/margin
ui.Text("Hello").Padding(1).Margin(2)
Button("Click").Padding(1).Margin(2)
Input("...").Padding(1).Margin(2)
```

---

## Implementation Strategy

### Phase 1: Universal Helper Methods (Generic)

Create **chainable helper functions** that work on any VNode:

```go
// ui/box_model.go

// Padding adds padding to any VNode
func Padding(vnode VNode, top, right, bottom, left int) VNode {
    setPadding(vnode, top, right, bottom, left)
    return vnode
}

// PaddingH sets horizontal padding
func PaddingH(vnode VNode, left, right int) VNode {
    return Padding(vnode, 0, right, 0, left)
}

// PaddingV sets vertical padding
func PaddingV(vnode VNode, top, bottom int) VNode {
    return Padding(vnode, top, 0, bottom, 0)
}

// PaddingAll sets same padding on all sides
func PaddingAll(vnode VNode, p int) VNode {
    return Padding(vnode, p, p, p, p)
}

// Margin adds margin to any VNode
func Margin(vnode VNode, top, right, bottom, left int) VNode {
    setMargin(vnode, top, right, bottom, left)
    return vnode
}

// MarginH sets horizontal margin
func MarginH(vnode VNode, left, right int) VNode {
    return Margin(vnode, 0, right, 0, left)
}

// MarginV sets vertical margin
func MarginV(vnode VNode, top, bottom int) VNode {
    return Margin(vnode, 0, right, 0, left)
}

// MarginAll sets same margin on all sides
func MarginAll(vnode VNode, m int) VNode {
    return Margin(vnode, m, m, m, m)
}
```

### Phase 2: Component-Specific Builders

Each component provides convenience methods that delegate to universal helpers:

```go
// components/button/button.go

func (b *ButtonBuilderType) Padding(top, right, bottom, left int) *ButtonBuilderType {
    ui.Padding(b.node, top, right, bottom, left)
    return b
}

func (b *ButtonBuilderType) PaddingH(left, right int) *ButtonBuilderType {
    ui.PaddingH(b.node, left, right)
    return b
}

func (b *ButtonBuilderType) PaddingV(top, bottom int) *ButtonBuilderType {
    ui.PaddingV(b.node, top, bottom)
    return b
}

func (b *ButtonBuilderType) PaddingAll(p int) *ButtonBuilderType {
    ui.PaddingAll(b.node, p)
    return b
}

func (b *ButtonBuilderType) Margin(top, right, bottom, left int) *ButtonBuilderType {
    ui.Margin(b.node, top, right, bottom, left)
    return b
}

func (b *ButtonBuilderType) MarginH(left, right int) *ButtonBuilderType {
    ui.MarginH(b.node, left, right)
    return b
}

func (b *ButtonBuilderType) MarginV(top, bottom int) *ButtonBuilderType {
    ui.MarginV(b.node, top, bottom)
    return b
}

func (b *ButtonBuilderType) MarginAll(m int) *ButtonBuilderType {
    ui.MarginAll(b.node, m)
    return b
}
```

### Phase 3: Layout Engine Integration

Layout engine automatically accounts for padding and margin:

```go
// runtime/compute/engine.go

func (e *Engine) layoutHStack(box *ComputedBox, x, y int) {
    for _, child := range box.Children {
        childInfo := rtui.GetLayoutInfo(child.VNode)

        // Get margin
        marginLeft := childInfo.Margin[3]   // [top, right, bottom, left]
        marginRight := childInfo.Margin[1]

        // Get padding
        paddingLeft := childInfo.Padding[3]
        paddingRight := childInfo.Padding[1]

        // Position child with margin
        childX := x + marginLeft

        // Calculate content width (excluding padding)
        contentWidth := child.Box.Width - paddingLeft - paddingRight

        // Layout child content
        e.calculatePositions(child, childX, y)

        // Next position: current + width + margins + gap
        x = childX + child.Box.Width + marginRight + layoutInfo.Gap
    }
}
```

### Phase 4: Component Rendering

Components apply padding when rendering:

```go
// components/button/button.go

func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // Get padding from LayoutInfo
    layoutInfo := rtui.GetLayoutInfo(b)
    paddingLeft := layoutInfo.Padding[3]
    paddingRight := layoutInfo.Padding[1]

    // Build content text
    contentText := focusIndicator + labelText

    // Apply padding
    buttonText := strings.Repeat(" ", paddingLeft) + contentText +
                  strings.Repeat(" ", paddingRight)

    return []paint.DrawCmd{
        paint.NewTextCmd(x, y, buttonText, buttonStyle),
    }
}
```

---

## Universal Helper Implementation

```go
// ui/box_model.go

package ui

// Internal helpers to set padding/margin props
func setPadding(vnode VNode, top, right, bottom, left int) {
    props := vnode.Props()
    if props == nil {
        props = make(Props)
        vnode.SetProps(props)
    }
    props["padding"] = [4]int{top, right, bottom, left}
}

func setMargin(vnode VNode, top, right, bottom, left int) {
    props := vnode.Props()
    if props == nil {
        props = make(Props)
        vnode.SetProps(props)
    }
    props["margin"] = [4]int{top, right, bottom, left}
}

// Chained helper functions (fluent API)

// Padding adds padding to any VNode
// Returns the same VNode for chaining
func Padding(vnode VNode, top, right, bottom, left int) VNode {
    setPadding(vnode, top, right, bottom, left)
    return vnode
}

// PaddingH sets horizontal padding (left, right)
func PaddingH(vnode VNode, left, right int) VNode {
    return Padding(vnode, 0, right, 0, left)
}

// PaddingV sets vertical padding (top, bottom)
func PaddingV(vnode VNode, top, bottom int) VNode {
    return Padding(vnode, top, 0, bottom, 0)
}

// PaddingAll sets same padding on all sides
func PaddingAll(vnode VNode, p int) VNode {
    return Padding(vnode, p, p, p, p)
}

// Margin adds margin to any VNode
// Returns the same VNode for chaining
func Margin(vnode VNode, top, right, bottom, left int) VNode {
    setMargin(vnode, top, right, bottom, left)
    return vnode
}

// MarginH sets horizontal margin (left, right)
func MarginH(vnode VNode, left, right int) VNode {
    return Margin(vnode, 0, right, 0, left)
}

// MarginV sets vertical margin (top, bottom)
func MarginV(vnode VNode, top, bottom int) VNode {
    return Margin(vnode, 0, right, 0, left)
}

// MarginAll sets same margin on all sides
func MarginAll(vnode VNode, m int) VNode {
    return Margin(vnode, m, m, m, m)
}
```

---

## Updated GetLayoutInfo

```go
// runtime/ui/layout_util.go

func GetLayoutInfo(vnode VNode) LayoutInfo {
    info := LayoutInfo{
        IsHorizontal: false,
        Gap:          0,
    }

    if vnode == nil {
        return info
    }

    // Check props for universal padding/margin (works for ANY VNode)
    if props := vnode.Props(); props != nil {
        // Read padding
        if p, ok := props["padding"].([4]int); ok {
            info.Padding = p
        }
        // Read margin
        if m, ok := props["margin"].([4]int); ok {
            info.Margin = m
        }
        // Read textAlign
        if ta, ok := props["textAlign"].(int); ok {
            info.TextAlign = Align(ta)
        }
    }

    // ... rest of existing logic
}
```

---

## Usage Examples

### Example 1: Universal padding on different components

```go
// All components can use padding
HStack(
    ui.PaddingAll(ui.Text("Hello"), 1),
    ui.PaddingAll(Button("Click"), 1),
    ui.PaddingAll(Input("..."), 1),
).
    Gap(1).
    Build()
```

### Example 2: Builder pattern (preferred)

```go
HStack(
    ui.Text("Hello").PaddingAll(1),
    Button("Click").PaddingAll(1),
    Input("...").PaddingAll(1),
).
    Gap(1).
    Build()
```

### Example 3: Complex layout with margin

```go
VStack(
    Button("Button1").MarginV(0, 1).PaddingH(2, 2),
    Button("Button2").MarginV(0, 1).PaddingH(2, 2),
    Button("Button3").MarginV(0, 1).PaddingH(2, 2),
).
    Gap(0).
    Build()
```

---

## Implementation Checklist

### Core Universal Methods
- [ ] Create `ui/box_model.go` with universal helpers
- [ ] Implement `Padding()`, `PaddingH()`, `PaddingV()`, `PaddingAll()`
- [ ] Implement `Margin()`, `MarginH()`, `MarginV()`, `MarginAll()`
- [ ] Update `GetLayoutInfo()` to read from props

### Layout Engine
- [ ] Update `layoutHStack()` to account for margins
- [ ] Update `layoutVStack()` to account for margins
- [ ] Ensure padding doesn't affect layout positioning (only rendering)

### Component Updates
- [ ] Button: Remove padding fields, use LayoutInfo instead
- [ ] Text: Apply padding when rendering
- [ ] Input: Apply padding when rendering
- [ ] (Future) All components: Support padding/margin

### Documentation
- [ ] Universal box model API guide
- [ ] Component integration examples
- [ ] Migration guide

---

## Benefits

1. ✅ **Universal** - Works on ALL components
2. ✅ **Consistent** - Same API everywhere
3. ✅ **CSS-like** - Familiar to web developers
4. ✅ **Minimal Code** - No duplication in each component
5. ✅ **Composable** - Easy to chain with other methods

---

## File Structure

```
mint/
├── ui/
│   └── box_model.go          # NEW: Universal helpers
├── runtime/ui/
│   └── layout_util.go        # MODIFY: Read padding/margin from props
├── runtime/compute/
│   └── engine.go             # MODIFY: Account for margins in positioning
└── components/
    ├── button/
    │   └── button.go         # MODIFY: Use LayoutInfo, apply padding in Paint
    ├── text/
    │   └── text.go           # MODIFY: Apply padding in Paint
    └── input/
        └── input.go          # MODIFY: Apply padding in Paint
```

---

**This is the CORRECT architectural approach!**

Padding and margin are **layout concepts**, not component-specific features.

---

**Next Steps:**
1. Implement universal helpers in `ui/box_model.go`
2. Update `GetLayoutInfo()` to read from props
3. Update layout engine to account for margins
4. Update components to use LayoutInfo instead of internal fields
