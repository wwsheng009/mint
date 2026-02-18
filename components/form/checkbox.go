package form

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/dimension"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ frameworkevent.Component = (*CheckboxVNode)(nil)
var _ component.Updater = (*CheckboxVNode)(nil) // Phase 3: Msg/Cmd support
var _ action.ActionTarget = (*CheckboxVNode)(nil)
var _ action.FocusableActionTarget = (*CheckboxVNode)(nil)

// CheckboxVNode represents a checkbox component
type CheckboxVNode struct {
	*ui.ElementVNode
	checked   bool
	disabled  bool
	label     string
	onChange  func(bool)
	isFocused bool // Internal focus state
	// Mouse interaction state
	isHovered    bool
	onMouseEnter func()
	onMouseLeave func()
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
}

// NewCheckbox creates a new checkbox
func NewCheckbox() *CheckboxVNode {
	return &CheckboxVNode{
		ElementVNode: ui.NewElement("checkbox"),
		checked:      false,
		disabled:     false,
		label:        "",
		isFocused:    false,
	}
}

// Checkbox creates a new checkbox node
func Checkbox() ui.VNode {
	return NewCheckbox()
}

// CheckboxBuilder creates a checkbox builder for chained calls
func CheckboxBuilder() *CheckboxBuilderType {
	return &CheckboxBuilderType{
		node: NewCheckbox(),
	}
}

// =============================================================================
// CheckboxVNode methods
// =============================================================================

// Checked returns whether the checkbox is checked
func (c *CheckboxVNode) Checked() bool {
	return c.checked
}

// SetChecked sets the checked state
func (c *CheckboxVNode) SetChecked(v bool) *CheckboxVNode {
	c.checked = v
	c.SetProp("checked", v)
	return c
}

// Toggle toggles the checked state
func (c *CheckboxVNode) Toggle() bool {
	c.checked = !c.checked
	c.SetProp("checked", c.checked)
	return c.checked
}

// Disabled returns whether the checkbox is disabled
func (c *CheckboxVNode) Disabled() bool {
	return c.disabled
}

// SetDisabled sets the disabled state
func (c *CheckboxVNode) SetDisabled(v bool) *CheckboxVNode {
	c.disabled = v
	c.SetProp("disabled", v)
	return c
}

// Label returns the label text
func (c *CheckboxVNode) Label() string {
	return c.label
}

// SetLabel sets the label text
func (c *CheckboxVNode) SetLabel(text string) *CheckboxVNode {
	c.label = text
	c.SetProp("label", text)
	return c
}

// OnChange returns the change handler
func (c *CheckboxVNode) OnChange() func(bool) {
	return c.onChange
}

// GetOnChange returns the change handler (implements ChangeHandlerProvider)
func (c *CheckboxVNode) GetOnChange() interface{} {
	return c.onChange
}

// SetOnChange sets the change handler
func (c *CheckboxVNode) SetOnChange(fn func(bool)) *CheckboxVNode {
	c.onChange = fn
	c.SetProp("onChange", fn)
	return c
}

// IsFocused returns whether the checkbox is focused
func (c *CheckboxVNode) IsFocused() bool {
	return c.isFocused
}

// SetFocus sets the focused state (implements FocusableVNode)
func (c *CheckboxVNode) SetFocus(focused bool) {
	c.isFocused = focused
}

// =============================================================================
// CheckboxBuilderType provides fluent API for building checkboxes
// =============================================================================

// CheckboxBuilderType is the builder for Checkbox
type CheckboxBuilderType struct {
	node *CheckboxVNode
}

// Checked sets the initial checked state
func (b *CheckboxBuilderType) Checked(v bool) *CheckboxBuilderType {
	b.node.SetChecked(v)
	return b
}

// Disabled sets the disabled state
func (b *CheckboxBuilderType) Disabled(v bool) *CheckboxBuilderType {
	b.node.SetDisabled(v)
	return b
}

// Label sets the label text
func (b *CheckboxBuilderType) Label(text string) *CheckboxBuilderType {
	b.node.SetLabel(text)
	return b
}

// OnChange sets the change handler
func (b *CheckboxBuilderType) OnChange(fn func(bool)) *CheckboxBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// Key sets the key for diffing
func (b *CheckboxBuilderType) Key(key string) *CheckboxBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *CheckboxBuilderType) Style(s style.Style) *CheckboxBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *CheckboxBuilderType) FgColor(c interface{}) *CheckboxBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *CheckboxBuilderType) BgColor(c interface{}) *CheckboxBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyle(s)
	}
	return b
}

// Bold sets the bold attribute
func (b *CheckboxBuilderType) Bold(v bool) *CheckboxBuilderType {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// Build returns the ui.VNode
func (b *CheckboxBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the checkbox is currently hovered
func (c *CheckboxVNode) IsHovered() bool {
	return c.isHovered
}

// SetHovered sets the hover state
func (c *CheckboxVNode) SetHovered(hovered bool) *CheckboxVNode {
	c.isHovered = hovered
	return c
}

// SetBounds sets the checkbox bounds for hit testing
func (c *CheckboxVNode) SetBounds(x, y, width, height int) {
	c.bounds = [4]int{x, y, width, height}
}

// Bounds returns the checkbox bounds
func (c *CheckboxVNode) Bounds() [4]int {
	return c.bounds
}

// ContainsPoint checks if a point is within the checkbox bounds
func (c *CheckboxVNode) ContainsPoint(x, y int) bool {
	if c.bounds[2] <= 0 || c.bounds[3] <= 0 {
		return false
	}
	return x >= c.bounds[0] && x < c.bounds[0]+c.bounds[2] &&
		y >= c.bounds[1] && y < c.bounds[1]+c.bounds[3]
}

// SetOnMouseEnter sets the mouse enter handler
func (c *CheckboxVNode) SetOnMouseEnter(fn func()) *CheckboxVNode {
	c.onMouseEnter = fn
	return c
}

// GetOnMouseEnter returns the mouse enter handler (implements MouseHandlerProvider)
func (c *CheckboxVNode) GetOnMouseEnter() func() {
	return c.onMouseEnter
}

// SetOnMouseLeave sets the mouse leave handler
func (c *CheckboxVNode) SetOnMouseLeave(fn func()) *CheckboxVNode {
	c.onMouseLeave = fn
	return c
}

// GetOnMouseLeave returns the mouse leave handler (implements MouseHandlerProvider)
func (c *CheckboxVNode) GetOnMouseLeave() func() {
	return c.onMouseLeave
}

// HandleEvent processes mouse and keyboard events for the checkbox
func (c *CheckboxVNode) HandleEvent(e frameworkevent.Event) bool {
	if c.disabled {
		return false
	}

	// Handle keyboard events (Space to toggle)
	keyEvent, ok := e.(*frameworkevent.KeyEvent)
	if ok {
		// Only respond to keyboard events when focused
		if !c.isFocused {
			return false
		}

		// Space toggles the checkbox
		if keyEvent.Key.Rune == ' ' {
			newState := c.Toggle()
			if c.onChange != nil {
				c.onChange(newState)
			}
			return true
		}
		return false
	}

	mouseEvent, ok := e.(*frameworkevent.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case frameworkevent.EventMouseEnter:
		if !c.isHovered {
			c.isHovered = true
			if c.onMouseEnter != nil {
				c.onMouseEnter()
			}
		}
		return true

	case frameworkevent.EventMouseLeave:
		if c.isHovered {
			c.isHovered = false
			if c.onMouseLeave != nil {
				c.onMouseLeave()
			}
		}
		return true

	case frameworkevent.EventMousePress, frameworkevent.EventClick:
		if c.isHovered && mouseEvent.Button == frameworkevent.MouseLeft {
			// Toggle the checkbox
			newState := c.Toggle()
			if c.onChange != nil {
				c.onChange(newState)
			}
			return true
		}
	}

	return false
}

// =============================================================================
// Msg/Cmd Architecture Support (Phase 3)
// =============================================================================

// Update implements component.Updater interface for Msg/Cmd architecture
//
// Handles:
// - KeyMsg: Space to toggle (when focused)
// - MouseMsg: Click to toggle
func (c *CheckboxVNode) Update(message runtimemsg.Msg) cmd.Cmd {
	if c.disabled {
		return nil
	}

	switch msg := message.(type) {
	case *runtimemsg.KeyMsg:
		// Only respond to keyboard events when focused
		if !c.isFocused {
			return nil
		}
		return c.updateKey(msg)

	case *runtimemsg.MouseMsg:
		return c.updateMouse(msg)
	}

	return nil
}

// updateKey handles keyboard messages (Space to toggle)
func (c *CheckboxVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	// Space toggles the checkbox
	if keyMsg.Rune == ' ' {
		newState := c.Toggle()
		if c.onChange != nil {
			c.onChange(newState)
		}
		return nil
	}

	return nil
}

// updateMouse handles mouse messages (click to toggle, hover tracking)
func (c *CheckboxVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	switch mouseMsg.Action {
	case runtimemsg.MouseActionMove:
		// Update hover state
		if !c.isHovered {
			c.isHovered = true
			if c.onMouseEnter != nil {
				c.onMouseEnter()
			}
		}
		return nil

	case runtimemsg.MouseActionPress:
		if mouseMsg.Button == runtimemsg.MouseLeft {
			// Toggle the checkbox
			newState := c.Toggle()
			if c.onChange != nil {
				c.onChange(newState)
			}
			return nil
		}
	}

	return nil
}

// =============================================================================
// Mouse Event Builder Methods
// =============================================================================

// OnMouseEnter sets the mouse enter handler (builder)
func (b *CheckboxBuilderType) OnMouseEnter(fn func()) *CheckboxBuilderType {
	b.node.SetOnMouseEnter(fn)
	return b
}

// OnMouseLeave sets the mouse leave handler (builder)
func (b *CheckboxBuilderType) OnMouseLeave(fn func()) *CheckboxBuilderType {
	b.node.SetOnMouseLeave(fn)
	return b
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the checkbox based on label and constraints
// Per Ant Design spec: box-width=3, gap=1
func (c *CheckboxVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if c == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Width per dimension spec: CheckBoxWidth "[X]" + CheckBoxGap + label length
	width := dimension.CheckBoxWidth + dimension.CheckBoxGap + utf8.RuneCountInString(c.label)
	height := 1

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
	elemStyle := c.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this checkbox component
func (c *CheckboxVNode) Paint(x, y int) []paint.DrawCmd {
	if c == nil {
		return nil
	}

	checkboxStyle := c.Style()

	// Checkbox indicator: [X] or [ ]
	var indicator string
	if c.checked {
		indicator = "[X]"
	} else {
		indicator = "[ ]"
	}

	// Build checkbox display: indicator + label
	var displayText string
	if c.label != "" {
		displayText = indicator + " " + c.label
	} else {
		displayText = indicator
	}

	// Apply default colors based on component spec
	// Normal: BG=SURFACE, FG=TEXT
	if checkboxStyle.FG == "" {
		checkboxStyle = checkboxStyle.Foreground(theme.Text())
	}
	if checkboxStyle.BG == "" {
		checkboxStyle = checkboxStyle.Background(theme.Surface())
	}

	// Checked state: BG=PRIMARY, FG=TEXT for readability
	if c.checked && !c.disabled {
		checkboxStyle = checkboxStyle.Foreground(theme.Text()).Bold(true)
	}

	// State priority: Focused > Hovered > Normal > Checked
	// Focus: FOCUS border (outline effect)
	if c.isFocused && !c.disabled {
		checkboxStyle = checkboxStyle.Foreground(theme.Focus()).Bold(true)
	} else if c.isHovered && !c.disabled {
		checkboxStyle = checkboxStyle.Underline(true)
	}

	// Apply disabled state: DISABLED_BG, DISABLED_FG
	if c.disabled {
		checkboxStyle = checkboxStyle.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}

	// CRITICAL: Set bounds for mouse hit testing
	checkboxWidth := utf8.RuneCountInString(displayText)
	checkboxHeight := 1
	c.SetBounds(x, y, checkboxWidth, checkboxHeight)

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, displayText, checkboxStyle),
	}
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// IsFocusable returns whether this checkbox can receive focus.
// Disabled checkboxes cannot receive focus.
func (c *CheckboxVNode) IsFocusable() bool {
	return !c.disabled
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the checkbox's Key if set, otherwise generates a stable ID.
func (c *CheckboxVNode) GetFocusID() string {
	if key := c.Key(); key != "" {
		return "checkbox:" + key
	}
	// Generate stable ID based on label
	id := c.label
	if id == "" {
		id = "checkbox"
	}
	return "checkbox:" + id
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction implements ActionTarget interface
func (c *CheckboxVNode) HandleAction(act *action.Action) bool {
	if act == nil || c.disabled {
		return false
	}

	switch act.Type {
	case action.ActionToggle, action.ActionClick, action.ActionEnter, action.ActionSelect:
		// Toggle the checkbox
		newState := c.Toggle()
		if c.onChange != nil {
			c.onChange(newState)
		}
		return true
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (c *CheckboxVNode) GetSupportedActions() []action.ActionType {
	return []action.ActionType{
		action.ActionToggle,
		action.ActionClick,
		action.ActionEnter,
		action.ActionSelect,
	}
}

// CanHandleAction implements ActionTarget interface
func (c *CheckboxVNode) CanHandleAction(act *action.Action) bool {
	if act == nil || c.disabled {
		return false
	}

	switch act.Type {
	case action.ActionToggle, action.ActionClick, action.ActionEnter, action.ActionSelect:
		return true
	}

	return false
}

// ============================================================================
// FocusableActionTarget 接口实现
// ============================================================================

// Focus implements FocusableActionTarget interface
func (c *CheckboxVNode) Focus() bool {
	if c.disabled {
		return false
	}
	c.SetFocus(true)
	return true
}

// Blur implements FocusableActionTarget interface
func (c *CheckboxVNode) Blur() {
	c.SetFocus(false)
}
