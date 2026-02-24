package modal

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for modal components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	title      string
	modalStyle style.Style
	borderStyle string
	closeIntent intent.Intent

	// === Modal Props ===
	isOpen    bool
	centered  bool
	closeable bool

	// === Layout Props ===
	width  int
	height int

	// === Content ===
	content rtui.VNode
	footer  rtui.VNode

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new ModalInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:         getStringProp(props, "key", ""),
		title:       getStringProp(props, "title", ""),
		modalStyle:  getStyleProp(props),
		borderStyle: getStringProp(props, "borderStyle", "single"),
		closeIntent: getIntentProp(props),
		isOpen:      getBoolProp(props, "isOpen", false),
		centered:    getBoolProp(props, "centered", true),
		closeable:   getBoolProp(props, "closeable", true),
		width:       getIntProp(props, "width", 40),
		height:      getIntProp(props, "height", 15),
		content:     getChildProp(props, "content"),
		footer:      getChildProp(props, "footer"),
		dirty:       true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy() {
	inst.content = nil
	inst.footer = nil
}
func (inst *Instance) OnMount() {}
func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldOpen := inst.isOpen
	oldTitle := inst.title

	inst.title = getStringProp(props, "title", inst.title)
	inst.modalStyle = getStyleProp(props)
	inst.borderStyle = getStringProp(props, "borderStyle", inst.borderStyle)
	inst.closeIntent = getIntentProp(props)
	inst.isOpen = getBoolProp(props, "isOpen", inst.isOpen)
	inst.centered = getBoolProp(props, "centered", inst.centered)
	inst.closeable = getBoolProp(props, "closeable", inst.closeable)
	inst.width = getIntProp(props, "width", inst.width)
	inst.height = getIntProp(props, "height", inst.height)
	inst.content = getChildProp(props, "content")
	inst.footer = getChildProp(props, "footer")

	changed := oldOpen != inst.isOpen || oldTitle != inst.title
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":         inst.key,
		"title":       inst.title,
		"isOpen":      inst.isOpen,
		"centered":    inst.centered,
		"closeable":   inst.closeable,
		"width":       inst.width,
		"height":      inst.height,
		"borderStyle": inst.borderStyle,
	}
}

func (inst *Instance) MarkDirty()          { inst.dirty = true }
func (inst *Instance) IsDirty() bool       { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()         { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement.
// If modal is not open, it takes no space.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if !inst.isOpen {
		return layout.Size{Width: 0, Height: 0}
	}

	width := inst.width
	height := inst.height

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

	return layout.Size{Width: width, Height: height}
}

// GetBorder implements layout.Bordered interface.
// Provides border layout information for proper child node positioning.
func (inst *Instance) GetBorder() layout.Border {
	if !inst.isOpen {
		return layout.Border{Style: layout.BorderNone}
	}

	// Convert borderStyle string to layout.BorderStyle
	var borderStyle layout.BorderStyle
	switch inst.borderStyle {
	case "double":
		borderStyle = layout.BorderDouble
	case "rounded":
		borderStyle = layout.BorderRounded
	case "dashed":
		borderStyle = layout.BorderDashed
	default:
		borderStyle = layout.BorderSingle
	}

	// Modal uses a title row layout:
	// - Top border (1 row)
	// - Title row (1 row) -> only if title exists
	// - Horizontal separator (1 row) -> only if title exists
	// - Content area
	// - Bottom border (1 row)
	//
	// For layout purposes:
	// Horizontal: always 1 char per side (left/right border)
	// Vertical:
	//   Without title: 1 + 1 = 2 rows (top + bottom)
	//   With title: 1 + 1 + 1 + 1 = 4 rows (top + title + separator + bottom)

	border := layout.NewBorder(borderStyle)

	// Add label info (title) for layout awareness
	if inst.title != "" {
		border.Label = inst.title
	}

	return border
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the modal.
// Only draws the border and title. Content and footer children are rendered by the framework.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if !inst.isOpen {
		return nil
	}

	width := inst.width
	height := inst.height

	// Set bounds for hit testing
	inst.bounds = [4]int{x, y, width, height}

	var cmds []paint.DrawCmd
	modalStyle := inst.modalStyle

	// Get border characters
	borderChars := inst.getBorderChars()

	// Build top border
	topBorder := string(borderChars.topLeft) + strings.Repeat(string(borderChars.horizontal), width-2) + string(borderChars.topRight)
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, modalStyle))

	// Draw title row if present
	if inst.title != "" {
		titlePadding := 1
		maxTitleLen := width - 2 - 2*titlePadding
		title := inst.title
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen]
		}
		titlePadLeft := strings.Repeat(" ", titlePadding)
		titlePadRight := strings.Repeat(" ", width-2-titlePadding-len(title)-titlePadding)
		titleRow := string(borderChars.vertical) + titlePadLeft + title + titlePadRight + string(borderChars.vertical)
		cmds = append(cmds, paint.NewTextCmd(x, y+1, titleRow, modalStyle))

		// Separator after title
		separator := string(borderChars.leftT) + strings.Repeat(string(borderChars.horizontal), width-2) + string(borderChars.rightT)
		cmds = append(cmds, paint.NewTextCmd(x, y+2, separator, modalStyle))
	}

	// Draw bottom border
	bottomBorder := string(borderChars.bottomLeft) + strings.Repeat(string(borderChars.horizontal), width-2) + string(borderChars.bottomRight)
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, bottomBorder, modalStyle))

	// NOTE: We do NOT draw the side borders or content area here!
	// They are handled by the paint engine's border rendering logic.
	// Content and footer children are rendered by the framework's layout engine.

	// Draw side borders (left and right for each row from border to bottom-1)
	startRow := y + 1
	if inst.title != "" {
		startRow = y + 3
	}
	endRow := y + height - 1
	for i := startRow; i < endRow; i++ {
		// Left border
		cmds = append(cmds, paint.NewTextCmd(x, i, string(borderChars.vertical), modalStyle))
		// Right border
		cmds = append(cmds, paint.NewTextCmd(x+width-1, i, string(borderChars.vertical), modalStyle))
	}

	return cmds
}

// getBorderChars returns the border characters based on borderStyle.
func (inst *Instance) getBorderChars() borderChars {
	switch inst.borderStyle {
	case "double":
		return borderChars{
			horizontal: '═',
			vertical:   '║',
			topLeft:    '╔',
			topRight:   '╗',
			bottomLeft: '╚',
			bottomRight:'╝',
			leftT:      '╠',
			rightT:     '╣',
		}
	case "rounded":
		return borderChars{
			horizontal: '─',
			vertical:   '│',
			topLeft:    '╭',
			topRight:   '╮',
			bottomLeft: '╰',
			bottomRight:'╯',
			leftT:      '├',
			rightT:     '┤',
		}
	case "dashed":
		return borderChars{
			horizontal: '─',
			vertical:   '│',
			topLeft:    '┌',
			topRight:   '┐',
			bottomLeft: '└',
			bottomRight:'┘',
			leftT:      '├',
			rightT:     '┤',
		}
	default: // single
		return borderChars{
			horizontal: '─',
			vertical:   '│',
			topLeft:    '┌',
			topRight:   '┐',
			bottomLeft: '└',
			bottomRight:'┘',
			leftT:      '├',
			rightT:     '┤',
		}
	}
}

type borderChars struct {
	horizontal  rune
	vertical    rune
	topLeft     rune
	topRight    rune
	bottomLeft  rune
	bottomRight rune
	leftT       rune
	rightT      rune
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) CanHandleAction(actionType string) bool {
	if !inst.isOpen || !inst.closeable {
		return false
	}
	return actionType == "close" || actionType == "click_outside" || actionType == "escape"
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
	if !inst.isOpen || !inst.closeable {
		return false
	}

	switch actionType {
	case "close", "click_outside", "escape":
		inst.isOpen = false
		inst.dirty = true
		inst.emitCloseIntent()
		return true
	}

	return false
}

// HandleKeyMessage handles keyboard messages for ESC key.
func (inst *Instance) HandleKeyMessage(keyMsg *runtimemsg.KeyMsg) bool {
	if !inst.isOpen || !inst.closeable {
		return false
	}

	// ESC to close
	if keyMsg.Special == runtimeplatform.KeyEscape {
		inst.isOpen = false
		inst.dirty = true
		inst.emitCloseIntent()
		return true
	}

	return false
}

// HandleMouseMessage handles mouse messages for clicking outside modal.
func (inst *Instance) HandleMouseMessage(mouseMsg *runtimemsg.MouseMsg) bool {
	if !inst.isOpen || !inst.closeable {
		return false
	}

	if mouseMsg.Action == runtimemsg.MouseActionPress {
		// Check if click is outside modal bounds
		if !inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			inst.isOpen = false
			inst.dirty = true
			inst.emitCloseIntent()
			return true
		}
	}

	return false
}

// containsPoint checks if a point is within the modal bounds.
func (inst *Instance) containsPoint(x, y int) bool {
	if inst.bounds[2] <= 0 || inst.bounds[3] <= 0 {
		return false
	}
	return x >= inst.bounds[0] && x < inst.bounds[0]+inst.bounds[2] &&
		y >= inst.bounds[1] && y < inst.bounds[1]+inst.bounds[3]
}

// emitCloseIntent emits the close intent to the parent.
func (inst *Instance) emitCloseIntent() {
	if inst.closeIntent != nil && inst.intentEmitter != nil {
		inst.intentEmitter(inst.closeIntent)
	}
}

// =============================================================================
// Bounds Support
// =============================================================================

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// =============================================================================
// Intent Emitter Support
// =============================================================================

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["modalStyle"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getIntentProp(props rtui.Props) intent.Intent {
	if v, ok := props["closeIntent"]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}

func getChildProp(props rtui.Props, key string) rtui.VNode {
	if v, ok := props[key]; ok {
		if vn, ok := v.(rtui.VNode); ok {
			return vn
		}
	}
	return nil
}
