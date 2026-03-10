# Box Model Interface Design

**Concept**: Define an interface for box model properties that all components can implement

---

## Problem with Current Implementation

Current implementation uses **props** to store box model information:

```go
// ❌ Current: Props-based approach
props["padding"] = [4]int{top, right, bottom, left}
props["margin"] = [4]int{top, right, bottom, left}
props["textAlign"] = int(AlignCenter)
```

**Issues**:
- Type safety: props are `interface{}`, no compile-time checking
- No contract: components don't declare they support box model
- Runtime errors: typos in prop names only caught at runtime

---

## Solution: BoxModel Interface

```go
// runtime/ui/box_model.go

// BoxModel defines the CSS box model contract
// Any component implementing this interface automatically supports
// padding, margin, and text alignment
type BoxModel interface {
    VNode

    // Padding returns the inner spacing [top, right, bottom, left]
    Padding() [4]int

    // Margin returns the outer spacing [top, right, bottom, left]
    Margin() [4]int

    // TextAlign returns the text alignment
    TextAlign() Align
}

// =============================================================================
// Universal Box Model Mixin
// =============================================================================

// BoxModelMixin provides default implementation for components
// Embed this in your component to automatically support box model
type BoxModelMixin struct {
    padding   [4]int
    margin    [4]int
    textAlign Align
}

// Padding returns the padding
func (b *BoxModelMixin) Padding() [4]int {
    return b.padding
}

// Margin returns the margin
func (b *BoxModelMixin) Margin() [4]int {
    return b.margin
}

// TextAlign returns the text alignment
func (b *BoxModelMixin) TextAlign() Align {
    return b.textAlign
}

// SetPadding sets the padding
func (b *BoxModelMixin) SetPadding(top, right, bottom, left int) {
    b.padding = [4]int{top, right, bottom, left}
}

// SetMargin sets the margin
func (b *BoxModelMixin) SetMargin(top, right, bottom, left int) {
    b.margin = [4]int{top, right, bottom, left}
}

// SetTextAlign sets the text alignment
func (b *BoxModelMixin) SetTextAlign(align Align) {
    b.textAlign = align
}
```

---

## Component Implementation

### Option 1: Embed BoxModelMixin (Recommended)

```go
// components/button/button.go

type ButtonVNode struct {
    *ui.ElementVNode
    label    string
    // ... other fields ...

    // Embed BoxModelMixin
    ui.BoxModelMixin
}

// Now ButtonVNode automatically implements BoxModel interface!
```

**Benefits**:
- ✅ Automatic interface satisfaction
- ✅ Minimal code
- ✅ Type-safe getters/setters
- ✅ Clear documentation of capabilities

### Option 2: Direct Implementation

```go
// components/button/button.go

type ButtonVNode struct {
    *ui.ElementVNode
    label    string
    // ... other fields ...

    // Box model fields
    padding   [4]int
    margin    [4]int
    textAlign ui.Align
}

// Implement BoxModel interface
func (b *ButtonVNode) Padding() [4]int {
    return b.padding
}

func (b *ButtonVNode) Margin() [4]int {
    return b.margin
}

func (b *ButtonVNode) TextAlign() ui.Align {
    return b.textAlign
}
```

---

## GetLayoutInfo Update

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

    // ⭐ NEW: Check if VNode implements BoxModel interface
    if boxModel, ok := vnode.(BoxModel); ok {
        info.Padding = boxModel.Padding()
        info.Margin = boxModel.Margin()
        info.TextAlign = boxModel.TextAlign()
    }

    // ... rest of existing logic
}
```

---

## Universal Helpers Update

```go
// ui/box_model.go

// Padding adds padding to any VNode implementing BoxModel
// Returns the same VNode for chaining
func Padding(vnode VNode, top, right, bottom, left int) VNode {
    if boxModel, ok := vnode.(BoxModel); ok {
        boxModel.SetPadding(top, right, bottom, left)
    } else {
        // Fallback: store in props (for components not yet migrated)
        props := vnode.Props()
        if props == nil {
            props = make(Props)
            vnode.SetProps(props)
        }
        props["padding"] = [4]int{top, right, bottom, left}
    }
    return vnode
}

// Similar for Margin, SetTextAlign...
```

---

## Component Migration

### Step 1: Embed Mixin

```go
type ButtonVNode struct {
    *ui.ElementVNode
    label    string
    onClick  func()

    // Embed box model support
    ui.BoxModelMixin
}
```

### Step 2: Update Measure()

```go
func (b *ButtonVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // ... calculate content size ...

    // ⭐ Use box model mixin
    horizontalPadding := b.Padding()[1] + b.Padding()[3]
    verticalPadding := b.Padding()[0] + b.Padding()[2]

    width := contentWidth + horizontalPadding
    height := contentHeight + verticalPadding

    // ... apply constraints ...
}
```

### Step 3: Update Paint()

```go
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
    // ... build content text ...

    // ⭐ Use box model mixin
    paddingLeft := b.Padding()[3]
    paddingRight := b.Padding()[1]
    textAlign := b.TextAlign()

    // Apply padding and alignment...
}
```

---

## Interface Contract

```go
// runtime/ui/box_model.go

// BoxModel defines the CSS box model contract
//
// Components implementing this interface automatically support:
// - Padding: inner spacing
// - Margin: outer spacing
// - TextAlign: text alignment within component
//
// Example:
//   type ButtonVNode struct {
//       *ui.ElementVNode
//       ui.BoxModelMixin  // Embed to automatically implement interface
//       // ... other fields ...
//   }
type BoxModel interface {
    VNode

    // Padding returns the inner spacing [top, right, bottom, left]
    Padding() [4]int

    // Margin returns the outer spacing [top, right, bottom, left]
    Margin() [4]int

    // TextAlign returns the text alignment within the component
    TextAlign() Align
}

// =============================================================================
// Helper Functions
// =============================================================================

// HasBoxModel checks if a VNode implements BoxModel interface
func HasBoxModel(vnode VNode) bool {
    _, ok := vnode.(BoxModel)
    return ok
}

// GetBoxModel safely gets BoxModel interface, returns nil if not implemented
func GetBoxModel(vnode VNode) BoxModel {
    if boxModel, ok := vnode.(BoxModel); ok {
        return boxModel
    }
    return nil
}
```

---

## Benefits

### 1. Type Safety ✅

```go
// Compile-time checking
var _ BoxModel = (*ButtonVNode)(nil)  // Button implements BoxModel

// Compiler error if component doesn't implement
var _ BoxModel = (*SomeVNode)(nil)  // ❌ Error!
```

### 2. Interface Contract ✅

```go
// Clear documentation of capabilities
func UseBoxModel(vnode VNode) {
    if boxModel, ok := vnode.(BoxModel); ok {
        // Guaranteed to have Padding(), Margin(), TextAlign()
        padding := boxModel.Padding()
    }
}
```

### 3. Documentation ✅

```go
// Component immediately shows it supports box model
type ButtonVNode struct {
    *ui.ElementVNode
    ui.BoxModelMixin  // ← Ah, this component supports padding/margin!
}
```

### 4. Refactoring ✅

```go
// Easy to find all components with box model support
// Search: "BoxModelMixin"

// Easy to migrate components to use box model
// Just embed BoxModelMixin
```

---

## Migration Path

### Phase 1: Create Interface (Foundation)
1. Create `BoxModel` interface in `runtime/ui/box_model.go`
2. Create `BoxModelMixin` struct
3. Update universal helpers to check interface

### Phase 2: Migrate Button (Proof of Concept)
1. Embed `BoxModelMixin` in `ButtonVNode`
2. Update `Measure()` to use `b.Padding()`
3. Update `Paint()` to use `b.Padding()`, `b.TextAlign()`
4. Remove internal `padding`, `textAlign` fields

### Phase 3: Migrate Other Components
1. Text component
2. Input component
3. Any custom components

### Phase 4: Update Layout Engine
1. `GetLayoutInfo()` checks interface first, then props
2. Prefer type-safe interface over props

---

## Implementation Timeline

| Phase | Tasks | Time |
|-------|-------|------|
| 1 | Create BoxModel interface and mixin | 1 hour |
| 2 | Migrate Button component | 1 hour |
| 3 | Update GetLayoutInfo and helpers | 30 min |
| 4 | Migrate Text and Input | 1 hour |
| 5 | Documentation and examples | 30 min |
| **Total** | | **4 hours** |

---

**This is the RIGHT approach!**

Interface-based design:
- ✅ Type-safe
- ✅ Self-documenting
- ✅ Go idiomatic
- ✅ Compile-time checking
- ✅ Clear contract

Props-based approach:
- ❌ Runtime-only checking
- ❌ No contract
- ❌ Type-unsafe
- ❌ Easy to make typos

---

**Next Step**: Implement BoxModel interface and migrate Button to use it.
