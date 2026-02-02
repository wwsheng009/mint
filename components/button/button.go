package button

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// ButtonVariant represents button style variants
type ButtonVariant int

const (
	// ButtonVariantDefault is the default button style
	ButtonVariantDefault ButtonVariant = iota
	// ButtonVariantPrimary is the primary action button
	ButtonVariantPrimary
	// ButtonVariantSecondary is the secondary action button
	ButtonVariantSecondary
	// ButtonVariantDanger is the danger action button
	ButtonVariantDanger
	// ButtonVariantSuccess is the success action button
	ButtonVariantSuccess
)

// ButtonSize represents button sizes
type ButtonSize int

const (
	// ButtonSizeSmall is a small button
	ButtonSizeSmall ButtonSize = iota
	// ButtonSizeMedium is a medium button
	ButtonSizeMedium
	// ButtonSizeLarge is a large button
	ButtonSizeLarge
)

// ButtonVNode represents a button component
type ButtonVNode struct {
	*ui.ElementVNode
	label         string
	onClick       func()
	variant       ButtonVariant
	size          ButtonSize
	disabled      bool
	focusIndex    int // Index for focus management, set during collection
	// Focus state
	hasFocus      bool // Whether this button currently has keyboard focus
	// Mouse interaction state
	isHovered     bool
	onMouseEnter  func()
	onMouseLeave  func()
	onMousePress  func()
	onMouseRelease func()
	// Bounds for hit testing (x, y, width, height)
	bounds        [4]int
}

// NewButton creates a new button
func NewButton(label string) *ButtonVNode {
	return &ButtonVNode{
		ElementVNode: ui.NewElement("button"),
		label:        label,
		variant:      ButtonVariantDefault,
		size:         ButtonSizeMedium,
		disabled:     false,
	}
}

// Label returns the button label
func (b *ButtonVNode) Label() string {
	return b.label
}

// SetLabel sets the button label
func (b *ButtonVNode) SetLabel(label string) *ButtonVNode {
	b.label = label
	return b
}

// OnClick returns the click handler
func (b *ButtonVNode) OnClick() func() {
	return b.onClick
}

// SetOnClick sets the click handler
func (b *ButtonVNode) SetOnClick(fn func()) *ButtonVNode {
	b.onClick = fn
	b.SetProp("onClick", fn)
	return b
}

// Variant returns the button variant
func (b *ButtonVNode) Variant() ButtonVariant {
	return b.variant
}

// SetVariant sets the button variant
func (b *ButtonVNode) SetVariant(v ButtonVariant) *ButtonVNode {
	b.variant = v
	b.SetProp("variant", v)
	return b
}

// Size returns the button size
func (b *ButtonVNode) Size() ButtonSize {
	return b.size
}

// SetSize sets the button size
func (b *ButtonVNode) SetSize(s ButtonSize) *ButtonVNode {
	b.size = s
	b.SetProp("size", s)
	return b
}

// Disabled returns whether the button is disabled
func (b *ButtonVNode) Disabled() bool {
	return b.disabled
}

// SetDisabled sets the disabled state
func (b *ButtonVNode) SetDisabled(v bool) *ButtonVNode {
	b.disabled = v
	b.SetProp("disabled", v)
	return b
}

// Button creates a new button node
func Button(label string) ui.VNode {
	return NewButton(label)
}

// ButtonBuilder creates a button builder for chained calls
func ButtonBuilder(label string) *ButtonBuilderType {
	return &ButtonBuilderType{
		node: NewButton(label),
	}
}

// ButtonBuilderType provides fluent API for building buttons
type ButtonBuilderType struct {
	node *ButtonVNode
}

// Label sets the button label
func (b *ButtonBuilderType) Label(label string) *ButtonBuilderType {
	b.node.SetLabel(label)
	return b
}

// OnClick sets the click handler
func (b *ButtonBuilderType) OnClick(fn func()) *ButtonBuilderType {
	b.node.SetOnClick(fn)
	return b
}

// Variant sets the button variant
func (b *ButtonBuilderType) Variant(v ButtonVariant) *ButtonBuilderType {
	b.node.SetVariant(v)
	return b
}

// Size sets the button size
func (b *ButtonBuilderType) Size(s ButtonSize) *ButtonBuilderType {
	b.node.SetSize(s)
	return b
}

// Disabled sets the disabled state
func (b *ButtonBuilderType) Disabled(v bool) *ButtonBuilderType {
	b.node.SetDisabled(v)
	return b
}

// Key sets the key for diffing
func (b *ButtonBuilderType) Key(key string) *ButtonBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *ButtonBuilderType) Style(s style.Style) *ButtonBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *ButtonBuilderType) FgColor(c interface{}) *ButtonBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *ButtonBuilderType) BgColor(c interface{}) *ButtonBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Build returns the ui.VNode
func (b *ButtonBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the button is currently hovered
func (b *ButtonVNode) IsHovered() bool {
	return b.isHovered
}

// SetHovered sets the hover state
func (b *ButtonVNode) SetHovered(hovered bool) *ButtonVNode {
	b.isHovered = hovered
	return b
}

// SetBounds sets the button bounds for hit testing
func (b *ButtonVNode) SetBounds(x, y, width, height int) {
	b.bounds = [4]int{x, y, width, height}
}

// Bounds returns the button bounds
func (b *ButtonVNode) Bounds() [4]int {
	return b.bounds
}

// ContainsPoint checks if a point is within the button bounds
func (b *ButtonVNode) ContainsPoint(x, y int) bool {
	if b.bounds[2] <= 0 || b.bounds[3] <= 0 {
		return false
	}
	return x >= b.bounds[0] && x < b.bounds[0]+b.bounds[2] &&
		y >= b.bounds[1] && y < b.bounds[1]+b.bounds[3]
}

// SetOnMouseEnter sets the mouse enter handler
func (b *ButtonVNode) SetOnMouseEnter(fn func()) *ButtonVNode {
	b.onMouseEnter = fn
	return b
}

// SetOnMouseLeave sets the mouse leave handler
func (b *ButtonVNode) SetOnMouseLeave(fn func()) *ButtonVNode {
	b.onMouseLeave = fn
	return b
}

// SetOnMousePress sets the mouse press handler
func (b *ButtonVNode) SetOnMousePress(fn func()) *ButtonVNode {
	b.onMousePress = fn
	return b
}

// SetOnMouseRelease sets the mouse release handler
func (b *ButtonVNode) SetOnMouseRelease(fn func()) *ButtonVNode {
	b.onMouseRelease = fn
	return b
}

// HandleEvent processes mouse and keyboard events for the button
func (b *ButtonVNode) HandleEvent(e event.Event) bool {
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "Button HandleEvent called: label=%q, disabled=%v, hasFocus=%v, event type=%T\n",
			b.label, b.disabled, b.hasFocus, e)
	}

	if b.disabled {
		return false
	}

	// Handle keyboard events (Enter/Space to click)
	keyEvent, ok := e.(*event.KeyEvent)
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "Button HandleEvent: type assertion ok=%v, event type=%T\n", ok, e)
	}
	if ok {
		// Only respond to keyboard events when focused
		if !b.hasFocus {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "Button HandleEvent: ignoring key event, button not focused (label=%q)\n", b.label)
			}
			return false
		}

		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "Button HandleEvent: KeyEvent, Special=%d (%v), Rune=%c, KeyEnter=%d\n",
				keyEvent.Special, keyEvent.Special, keyEvent.Key.Rune, event.KeyEnter)
		}
		// Check for Enter key or Space key
		if keyEvent.Special == event.KeyEnter || keyEvent.Key.Rune == ' ' {
			// Space or Enter triggers click
			if b.onClick != nil {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "Button HandleEvent: triggering onClick for label=%q (Special=%d)\n", b.label, keyEvent.Special)
				}
				b.onClick()
				return true
			} else {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "Button HandleEvent: onClick is nil for label=%q\n", b.label)
				}
			}
		} else {
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "Button HandleEvent: key not matched (Special=%d vs KeyEnter=%d)\n",
					keyEvent.Special, event.KeyEnter)
			}
		}
		return false
	}

	mouseEvent, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case event.EventMouseEnter:
		if !b.isHovered {
			b.isHovered = true
			if b.onMouseEnter != nil {
				b.onMouseEnter()
			}
		}
		return true

	case event.EventMouseLeave:
		if b.isHovered {
			b.isHovered = false
			if b.onMouseLeave != nil {
				b.onMouseLeave()
			}
		}
		return true

	case event.EventMousePress:
		if b.isHovered && mouseEvent.Button == event.MouseLeft {
			if b.onMousePress != nil {
				b.onMousePress()
			}
			return true
		}

	case event.EventMouseRelease:
		if b.isHovered && mouseEvent.Button == event.MouseLeft {
			if b.onMouseRelease != nil {
				b.onMouseRelease()
			}
			// Trigger click on mouse release when still hovered
			if b.onClick != nil {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "Button HandleEvent: mouse click for label=%q\n", b.label)
				}
				b.onClick()
			}
			return true
		}

	case event.EventClick:
		if b.isHovered && mouseEvent.Button == event.MouseLeft {
			if b.onClick != nil {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "Button HandleEvent: click event for label=%q\n", b.label)
				}
				b.onClick()
			}
			return true
		}
	}

	return false
}

// =============================================================================
// Mouse Event Builder Methods
// =============================================================================

// OnMouseEnter sets the mouse enter handler (builder)
func (b *ButtonBuilderType) OnMouseEnter(fn func()) *ButtonBuilderType {
	b.node.SetOnMouseEnter(fn)
	return b
}

// OnMouseLeave sets the mouse leave handler (builder)
func (b *ButtonBuilderType) OnMouseLeave(fn func()) *ButtonBuilderType {
	b.node.SetOnMouseLeave(fn)
	return b
}

// OnMousePress sets the mouse press handler (builder)
func (b *ButtonBuilderType) OnMousePress(fn func()) *ButtonBuilderType {
	b.node.SetOnMousePress(fn)
	return b
}

// OnMouseRelease sets the mouse release handler (builder)
func (b *ButtonBuilderType) OnMouseRelease(fn func()) *ButtonBuilderType {
	b.node.SetOnMouseRelease(fn)
	return b
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the button based on label and constraints
func (b *ButtonVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if b == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate button width: label + padding (brackets)
	label := b.label
	if label == "" {
		label = " " // Empty button still has minimal width
	}

	// Width: label length + 2 for brackets "[]"
	width := utf8.RuneCountInString(label) + 2

	// Height is always 1 for single-line button
	height := 1

	// Apply size modifiers
	switch b.size {
	case ButtonSizeSmall:
		// Small button: no extra padding
	case ButtonSizeMedium:
		// Medium button: +1 padding on each side
		width += 2
	case ButtonSizeLarge:
		// Large button: +2 padding on each side
		width += 4
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	// Apply explicit style dimensions if set
	style := b.Style()
	if style.Width > 0 {
		width = style.Width
	}
	if style.Height > 0 {
		height = style.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this button component
func (b *ButtonVNode) Paint(x, y int) []paint.DrawCmd {
	if b == nil {
		return nil
	}

	// Get button style for rendering
	buttonStyle := b.Style()

	// Build button label with brackets
	displayLabel := b.label
	if displayLabel == "" {
		displayLabel = " "
	}

	// Format: [label]
	var labelText string
	switch b.size {
	case ButtonSizeSmall:
		labelText = "[" + displayLabel + "]"
	case ButtonSizeMedium:
		labelText = "[ " + displayLabel + " ]"
	case ButtonSizeLarge:
		labelText = "[  " + displayLabel + "  ]"
	default:
		labelText = "[" + displayLabel + "]"
	}

	// Apply variant-based styling if not explicitly set
	if buttonStyle.FG == "" && buttonStyle.BG == "" {
		switch b.variant {
		case ButtonVariantPrimary:
			buttonStyle = buttonStyle.Foreground(style.Color("white")).Background(style.Color("blue")).Bold(true)
		case ButtonVariantSecondary:
			buttonStyle = buttonStyle.Foreground(style.Color("black")).Background(style.Color("gray"))
		case ButtonVariantDanger:
			buttonStyle = buttonStyle.Foreground(style.Color("white")).Background(style.Color("red")).Bold(true)
		case ButtonVariantSuccess:
			buttonStyle = buttonStyle.Foreground(style.Color("white")).Background(style.Color("green")).Bold(true)
		case ButtonVariantDefault:
			buttonStyle = buttonStyle.Foreground(style.Color("black")).Background(style.Color("white"))
		}
	}

	// Apply disabled state
	if b.disabled {
		buttonStyle = buttonStyle.Foreground(style.Color("gray")).Background(style.Color("darkgray"))
	}

	// State priority: Focused > Hovered > Normal
	// Focus: distinct visual feedback (reversed colors)
	// Hover: subtle feedback (underline)
	if b.hasFocus && !b.disabled {
		// Focused state: blue background with white text for clear visibility
		buttonStyle = buttonStyle.Foreground(style.Color("white")).Background(style.Color("blue")).Bold(true)
	} else if b.isHovered && !b.disabled {
		// Hovered state: underline only
		buttonStyle = buttonStyle.Underline(true)
	}

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, labelText, buttonStyle),
	}
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// SetFocus sets the focus state of this button.
// When focused, the button will display visual feedback (e.g., underline).
func (b *ButtonVNode) SetFocus(hasFocus bool) {
	b.hasFocus = hasFocus
}

// HasFocus returns whether this button currently has focus.
func (b *ButtonVNode) HasFocus() bool {
	return b.hasFocus
}

// IsFocusable returns whether this button can receive focus.
// Disabled buttons cannot receive focus.
func (b *ButtonVNode) IsFocusable() bool {
	return !b.disabled
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the button's Key if set, otherwise generates a stable ID.
func (b *ButtonVNode) GetFocusID() string {
	if key := b.Key(); key != "" {
		return "button:" + key
	}
	// Generate stable ID based on label
	return "button:" + b.label
}
