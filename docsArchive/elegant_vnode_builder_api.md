# Elegant VNode Builder API

**Problem**: `SetProp` is not elegant. Users want fluent chaining like CSS.

**Solution**: Universal VNode builder wrapper.

---

## Current Problem

```go
// ❌ Ugly: Can't chain after universal helpers
btn := app.ButtonBuilder("Click")
    .Build()
btn.SetProp("flex", 1)  // Can't chain

// ❌ Also ugly: Universal helpers return VNode
Button("Click").
    Padding(2)  // Returns ui.VNode, can't chain Button-specific methods
```

---

## Proposed Solution: Universal VNode Builder

```go
// ui/vnode_builder.go

// VNodeBuilder wraps any VNode and provides fluent chainable methods
type VNodeBuilder struct {
    vnode VNode
}

// Build returns the wrapped VNode
func (b *VNodeBuilder) Build() VNode {
    return b.vnode
}

// =============================================================================
// Universal Box Model Methods (work on ANY component)
// =============================================================================

// Padding adds padding - returns builder for chaining
func (b *VNodeBuilder) Padding(top, right, bottom, left int) *VNodeBuilder {
    ui.Padding(b.vnode, top, right, bottom, left)
    return b
}

func (b *VNodeBuilder) PaddingH(left, right int) *VNodeBuilder {
    ui.PaddingH(b.vnode, left, right)
    return b
}

func (b *VNodeBuilder) PaddingV(top, bottom int) *VNodeBuilder {
    ui.PaddingV(b.vnode, top, bottom)
    return b
}

func (b *VNodeBuilder) PaddingAll(p int) *VNodeBuilder {
    ui.PaddingAll(b.vnode, p)
    return b
}

// Margin adds margin - returns builder for chaining
func (b *VNodeBuilder) Margin(top, right, bottom, left int) *VNodeBuilder {
    ui.Margin(b.vnode, top, right, bottom, left)
    return b
}

func (b *VNodeBuilder) MarginH(left, right int) *VNodeBuilder {
    ui.MarginH(b.vnode, left, right)
    return b
}

func (b *VNodeBuilder) MarginV(top, bottom int) *VNodeBuilder {
    ui.MarginV(b.vnode, top, bottom)
    return b
}

func (b *VNodeBuilder) MarginAll(m int) *VNodeBuilder {
    ui.MarginAll(b.vnode, m)
    return b
}

// =============================================================================
// Layout Properties
// =============================================================================

func (b *VNodeBuilder) Flex(f int) *VNodeBuilder {
    props := b.vnode.Props()
    if props == nil {
        props = make(ui.Props)
        b.vnode.SetProps(props)
    }
    props["flex"] = f
    return b
}

func (b *VNodeBuilder) FillWidth() *VNodeBuilder {
    props := b.vnode.Props()
    if props == nil {
        props = make(ui.Props)
        b.vnode.SetProps(props)
    }
    props["fillWidth"] = true
    return b
}

func (b *VNodeBuilder) FillHeight() *VNodeBuilder {
    props := b.vnode.Props()
    if props == nil {
        props = make(ui.Props)
        b.vnode.SetProps(props)
    }
    props["fillHeight"] = true
    return b
}

// =============================================================================
// Universal Constructor
// =============================================================================

// Wrap creates a VNodeBuilder from any VNode
func Wrap(vnode VNode) *VNodeBuilder {
    return &VNodeBuilder{vnode: vnode}
}
```

---

## Usage Examples

### Before (Ugly)

```go
// ❌ Can't chain after universal methods
app.ButtonBuilder("Click").
    Padding(2).  // Returns ui.VNode
    Build()

// ❌ Need SetProp for layout properties
btn := app.ButtonBuilder("Click").Build()
btn.SetProp("flex", 1)
```

### After (Elegant)

```go
// ✅ Everything is chainable
app.ButtonBuilder("Click").
    Wrap().         // Wrap in universal builder
    Padding(2).     // Chain padding
    Flex(1).        // Chain flex
    Build().        // Get VNode
    Build()         // Get final VNode
```

### Even Better: Auto-Wrap in ButtonBuilder

```go
// ButtonBuilder auto-wraps internal node
func (b *ButtonBuilderType) Padding(p int) *ButtonBuilderType {
    ui.PaddingAll(b.node, p)
    return b
}

func (b *ButtonBuilderType) Flex(f int) *ButtonBuilderType {
    b.node.SetProp("flex", f)
    return b
}

// Usage:
app.ButtonBuilder("Click").
    Padding(2).   // Still returns *ButtonBuilderType
    Flex(1).      // Still returns *ButtonBuilderType
    Build()       // Perfect!
```

---

## Implementation Priority

1. **High**: Add `Flex()`, `FillWidth()`, `FillHeight()` to ButtonBuilder
2. **High**: Ensure all layout properties have chainable methods
3. **Medium**: Create universal VNodeBuilder wrapper
4. **Low**: Deprecate SetProp (keep for backward compatibility)

---

**Key Insight**:
The best API is where users never see `SetProp`. Every property should have a chainable method.

**CSS Comparison**:
```css
/* CSS is elegant */
button {
    padding: 10px;
    flex: 1;
}
```

```go
// Mint TUI should be equally elegant
app.ButtonBuilder("Click").
    Padding(2).   // Not SetProp("padding", 2)
    Flex(1).      // Not SetProp("flex", 1)
    Build()
```

---

**Next Step**: Add Flex/FillWidth/FillHeight methods to ButtonBuilder
