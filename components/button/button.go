package button

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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

// ButtonFocusStyle defines how a button displays focus state
type ButtonFocusStyle int

const (
	// FocusStyleReverse uses reversed colors (default)
	FocusStyleReverse ButtonFocusStyle = iota
	// FocusStyleUnderline uses underline only (preserves background)
	FocusStyleUnderline
	// FocusStyleBracket uses brackets around the label (preserves background)
	FocusStyleBracket
	// FocusStyleBold uses bold text only (preserves background)
	FocusStyleBold
)

// ButtonVNode represents a button component
type ButtonVNode struct {
	*ui.ElementVNode
	label         string
	onClick       func()
	variant       ButtonVariant
	size          ButtonSize
	disabled      bool
	// Focus state
	hasFocus      bool // Whether this button currently has keyboard focus
	focusStyle    ButtonFocusStyle // How to display focus state
	// Mouse interaction state
	isHovered     bool
	onMouseEnter  func()
	onMouseLeave  func()
	onMousePress  func()
	onMouseRelease func()
	// Bounds for hit testing (x, y, width, height)
	bounds        [4]int
	// ⭐ NEW: Embed BoxModelMixin for automatic box model support
	rtui.BoxModelMixin
}

// NewButton creates a new button
func NewButton(label string) *ButtonVNode {
	return &ButtonVNode{
		ElementVNode: ui.NewElement("button"),
		label:        label,
		variant:      ButtonVariantDefault,
		size:         ButtonSizeMedium,
		disabled:     false,
		focusStyle:   FocusStyleReverse, // Default: reverse colors
		// BoxModelMixin fields are zero-initialized: padding=[0,0,0,0], margin=[0,0,0,0], textAlign=AlignStart
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

// FocusStyle returns the button focus style
func (b *ButtonVNode) FocusStyle() ButtonFocusStyle {
	return b.focusStyle
}

// SetFocusStyle sets the button focus style
func (b *ButtonVNode) SetFocusStyle(s ButtonFocusStyle) *ButtonVNode {
	b.focusStyle = s
	b.SetProp("focusStyle", s)
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

// FocusStyle sets the focus style
func (b *ButtonBuilderType) FocusStyle(s ButtonFocusStyle) *ButtonBuilderType {
	b.node.SetFocusStyle(s)
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

// SetTextAlign sets the text alignment within the button
// Uses BoxModelMixin method - type-safe!
func (b *ButtonBuilderType) SetTextAlign(align rtui.Align) *ButtonBuilderType {
	b.node.SetTextAlign(align)
	return b
}

// =============================================================================
// Padding Methods - Universal box model (same API for all components)
// =============================================================================

// Padding adds padding on all sides (top, right, bottom, left)
func (b *ButtonBuilderType) Padding(top, right, bottom, left int) *ButtonBuilderType {
	ui.Padding(b.node, top, right, bottom, left)
	return b
}

// PaddingH sets horizontal padding (left, right)
func (b *ButtonBuilderType) PaddingH(left, right int) *ButtonBuilderType {
	ui.PaddingH(b.node, left, right)
	return b
}

// PaddingV sets vertical padding (top, bottom)
func (b *ButtonBuilderType) PaddingV(top, bottom int) *ButtonBuilderType {
	ui.PaddingV(b.node, top, bottom)
	return b
}

// PaddingAll sets same padding on all sides
func (b *ButtonBuilderType) PaddingAll(p int) *ButtonBuilderType {
	ui.PaddingAll(b.node, p)
	return b
}

// PaddingTop sets top padding
func (b *ButtonBuilderType) PaddingTop(p int) *ButtonBuilderType {
	ui.Padding(b.node, p, 0, 0, 0)
	return b
}

// PaddingRight sets right padding
func (b *ButtonBuilderType) PaddingRight(p int) *ButtonBuilderType {
	ui.Padding(b.node, 0, p, 0, 0)
	return b
}

// PaddingBottom sets bottom padding
func (b *ButtonBuilderType) PaddingBottom(p int) *ButtonBuilderType {
	ui.Padding(b.node, 0, 0, p, 0)
	return b
}

// PaddingLeft sets left padding
func (b *ButtonBuilderType) PaddingLeft(p int) *ButtonBuilderType {
	ui.Padding(b.node, 0, 0, 0, p)
	return b
}

// =============================================================================
// Margin Methods - Universal box model (same API for all components)
// =============================================================================

// Margin adds margin on all sides (top, right, bottom, left)
func (b *ButtonBuilderType) Margin(top, right, bottom, left int) *ButtonBuilderType {
	ui.Margin(b.node, top, right, bottom, left)
	return b
}

// MarginH sets horizontal margin (left, right)
func (b *ButtonBuilderType) MarginH(left, right int) *ButtonBuilderType {
	ui.MarginH(b.node, left, right)
	return b
}

// MarginV sets vertical margin (top, bottom)
func (b *ButtonBuilderType) MarginV(top, bottom int) *ButtonBuilderType {
	ui.MarginV(b.node, top, bottom)
	return b
}

// MarginAll sets same margin on all sides
func (b *ButtonBuilderType) MarginAll(m int) *ButtonBuilderType {
	ui.MarginAll(b.node, m)
	return b
}

// MarginTop sets top margin
func (b *ButtonBuilderType) MarginTop(m int) *ButtonBuilderType {
	ui.Margin(b.node, m, 0, 0, 0)
	return b
}

// MarginRight sets right margin
func (b *ButtonBuilderType) MarginRight(m int) *ButtonBuilderType {
	ui.Margin(b.node, 0, m, 0, 0)
	return b
}

// MarginBottom sets bottom margin
func (b *ButtonBuilderType) MarginBottom(m int) *ButtonBuilderType {
	ui.Margin(b.node, 0, 0, m, 0)
	return b
}

// MarginLeft sets left margin
func (b *ButtonBuilderType) MarginLeft(m int) *ButtonBuilderType {
	ui.Margin(b.node, 0, 0, 0, m)
	return b
}

// =============================================================================
// Layout Properties - Elegant chainable methods (no SetProp needed!)
// =============================================================================

// Flex sets the flex factor for this button
// CSS equivalent: flex: value
func (b *ButtonBuilderType) Flex(f int) *ButtonBuilderType {
	b.node.SetProp("flex", f)
	return b
}

// FillWidth makes this button stretch to fill parent's width
// CSS equivalent: width: 100%
func (b *ButtonBuilderType) FillWidth() *ButtonBuilderType {
	b.node.SetProp("fillWidth", true)
	return b
}

// FillHeight makes this button stretch to fill parent's height
func (b *ButtonBuilderType) FillHeight() *ButtonBuilderType {
	b.node.SetProp("fillHeight", true)
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
	log.UILogger.Debug("Button HandleEvent called: label=%q, disabled=%v, hasFocus=%v, event type=%T",
		b.label, b.disabled, b.hasFocus, e)

	if b.disabled {
		return false
	}

	// Handle keyboard events (Enter/Space to click)
	keyEvent, ok := e.(*event.KeyEvent)
	log.UILogger.Debug("Button HandleEvent: type assertion ok=%v, event type=%T", ok, e)
	if ok {
		// Only respond to keyboard events when focused
		if !b.hasFocus {
			log.UILogger.Debug("Button HandleEvent: ignoring key event, button not focused (label=%q)", b.label)
			return false
		}

		log.UILogger.Debug("Button HandleEvent: KeyEvent, Special=%d (%v), Rune=%c, KeyEnter=%d",
			keyEvent.Special, keyEvent.Special, keyEvent.Key.Rune, event.KeyEnter)
		// Check for Enter key or Space key
		if keyEvent.Special == event.KeyEnter || keyEvent.Key.Rune == ' ' {
			// Space or Enter triggers click
			if b.onClick != nil {
				log.UILogger.Debug("Button HandleEvent: triggering onClick for label=%q (Special=%d)", b.label, keyEvent.Special)
				b.onClick()
				return true
			} else {
				log.UILogger.Debug("Button HandleEvent: onClick is nil for label=%q", b.label)
			}
		} else {
			log.UILogger.Debug("Button HandleEvent: key not matched (Special=%d vs KeyEnter=%d)",
				keyEvent.Special, event.KeyEnter)
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
		// Check if mouse is within button bounds
		if b.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == event.MouseLeft {
			log.UILogger.Debug("Button HandleEvent: mouse press within bounds for label=%q, x=%d, y=%d, bounds=%v",
				b.label, mouseEvent.X, mouseEvent.Y, b.bounds)
			if b.onMousePress != nil {
				b.onMousePress()
			}
			// Trigger click on press for better responsiveness
			if b.onClick != nil {
				log.UILogger.Debug("Button HandleEvent: triggering onClick for label=%q", b.label)
				b.onClick()
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
				log.UILogger.Debug("Button HandleEvent: mouse click for label=%q", b.label)
				b.onClick()
			}
			return true
		}

	case event.EventClick:
		if b.isHovered && mouseEvent.Button == event.MouseLeft {
			if b.onClick != nil {
				log.UILogger.Debug("Button HandleEvent: click event for label=%q", b.label)
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

	// Calculate button content width: label + brackets + focus indicator
	label := b.label
	if label == "" {
		label = " " // Empty button still has minimal width
	}

	// Width: label length + 2 for brackets "[]" + 1 for focus indicator
	contentWidth := utf8.RuneCountInString(label) + 3

	// Height is always 1 for single-line button
	contentHeight := 1

	// Apply size modifiers
	switch b.size {
	case ButtonSizeSmall:
		// Small button: no extra padding
	case ButtonSizeMedium:
		// Medium button: +1 padding on each side
		contentWidth += 2
	case ButtonSizeLarge:
		// Large button: +2 padding on each side
		contentWidth += 4
	}

	// ⭐ NEW: Apply user-specified padding from BoxModelMixin
	// Use interface method instead of props/LayoutInfo
	padding := b.Padding()
	horizontalPadding := padding[1] + padding[3] // right + left
	verticalPadding := padding[0] + padding[2]   // top + bottom

	width := contentWidth + horizontalPadding
	height := contentHeight + verticalPadding

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

	// Debug: log button paint state with bounds
	focusMarker := " "
	if b.hasFocus {
		focusMarker = "*"
	}
	log.ButtonLogger.Debug("ButtonPaint label=%q, hasFocus=%v, focusMarker=%s", b.label, b.hasFocus, focusMarker)

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
	// Based on component spec: comp_1.md and comp_2.md
	if buttonStyle.FG == "" && buttonStyle.BG == "" {
		switch b.variant {
		case ButtonVariantPrimary:
			// Primary: BG=PRIMARY, FG=BG
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
		case ButtonVariantSecondary:
			// Secondary: BG=SURFACE, FG=TEXT
			buttonStyle = buttonStyle.Foreground(theme.Text()).Background(theme.Surface())
		case ButtonVariantDanger:
			// Danger: BG=ERROR, FG=BG
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Error()).Bold(true)
		case ButtonVariantSuccess:
			// Success: BG=SUCCESS, FG=BG
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Success()).Bold(true)
		case ButtonVariantDefault:
			// Default: BG=SURFACE, FG=TEXT
			buttonStyle = buttonStyle.Foreground(theme.Text()).Background(theme.Surface())
		}
	}

	// Apply disabled state
	// Disabled: FG=DISABLED_FG, BG=DISABLED_BG
	if b.disabled {
		buttonStyle = buttonStyle.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}

	// State priority: Focused > Hovered > Normal
	// Focus style is controlled by b.focusStyle
	if b.hasFocus && !b.disabled {
		// Apply focus style based on setting
		switch b.focusStyle {
		case FocusStyleUnderline:
			// Bright underline with contrasting color for visibility
			buttonStyle = buttonStyle.
				Foreground(theme.FocusBright()).
				Underline(true).
				Bold(true)
		case FocusStyleBracket:
			// Brackets with bright color for visibility
			// IMPORTANT: Preserve background color, but ALWAYS use bright foreground for focus
			customBG := buttonStyle.BG
			// Force bright foreground color for focus visibility
			buttonStyle = buttonStyle.
				Foreground(theme.FocusBright()).
				Bold(true)
			// Ensure custom background is not lost
			if customBG != "" {
				buttonStyle = buttonStyle.Background(customBG)
			}
		case FocusStyleBold:
			// Bold only (preserves background color)
			buttonStyle = buttonStyle.Bold(true)
		case FocusStyleReverse:
			// Default: theme focus background with foreground text
			buttonStyle = buttonStyle.Foreground(theme.Foreground()).Background(theme.Focus()).Bold(true)
		}
	} else if b.isHovered && !b.disabled {
		// Hovered state: underline only (no background)
		buttonStyle = buttonStyle.Underline(true)
	}

	// Add focus indicator: * before focused button (only for Reverse style)
	var focusIndicator string
	if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleReverse {
		focusIndicator = "*"
	} else if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleUnderline {
		focusIndicator = ">" // Visible indicator for underline style
	} else if b.hasFocus && !b.disabled && b.focusStyle == FocusStyleBracket {
		focusIndicator = ">" // Visible indicator for bracket style
	} else {
		focusIndicator = " "
	}

	// Build the button text with natural width
	buttonText := focusIndicator + labelText
	contentWidth := len(buttonText)

	// ⭐ NEW: Get padding from BoxModelMixin (type-safe, no props needed!)
	padding := b.Padding()
	paddingLeft := padding[3]
	paddingRight := padding[1]

	// Natural width includes content and user-specified padding
	naturalWidth := contentWidth + paddingLeft + paddingRight

	// Get layout width from bounds (set by layout engine)
	layoutWidth := naturalWidth
	if b.bounds[2] > 0 {
		layoutWidth = b.bounds[2]  // Allocated width from flex
	}

	// Apply text alignment if button is stretched by flex layout
	if layoutWidth > naturalWidth {
		textAlign := b.TextAlign()  // ⭐ Use interface method
		availableSpace := layoutWidth - naturalWidth

		switch textAlign {
		case rtui.AlignCenter:
			// Center text: distribute space on both sides
			leftPadding := paddingLeft + availableSpace/2
			rightPadding := paddingRight + (availableSpace - availableSpace/2)
			buttonText = strings.Repeat(" ", leftPadding) + buttonText +
				strings.Repeat(" ", rightPadding)

		case rtui.AlignEnd:
			// Right-align: add all space to left
			leftPadding := paddingLeft + availableSpace
			buttonText = strings.Repeat(" ", leftPadding) + buttonText

		case rtui.AlignStart:
			// Left-align: add all space to right (default)
			buttonText = buttonText + strings.Repeat(" ", paddingRight + availableSpace)

		case rtui.AlignSpaceBetween, rtui.AlignSpaceAround:
			// These don't make sense for single button text alignment
			// Fall back to left-align
			buttonText = buttonText + strings.Repeat(" ", paddingRight + availableSpace)
		}
	} else {
		// Not stretched: just apply user-specified padding
		if paddingLeft > 0 {
			buttonText = strings.Repeat(" ", paddingLeft) + buttonText
		}
		if paddingRight > 0 {
			buttonText = buttonText + strings.Repeat(" ", paddingRight)
		}
	}

	// Return draw commands: focus indicator + button label
	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, buttonText, buttonStyle),
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
// Uses the button's Key if set, otherwise generates a unique ID using
// the button's memory address. This ensures uniqueness without requiring
// a separate focusIndex field that would need to be managed during collection.
func (b *ButtonVNode) GetFocusID() string {
	if key := b.Key(); key != "" {
		return "button:" + key
	}
	// Use the button's pointer address for uniqueness when no key is set.
	// This works because each button component is a distinct instance.
	// The %p format gives us a stable unique identifier for the lifetime of the button.
	return fmt.Sprintf("button:%s@%p", b.label, b)
}
