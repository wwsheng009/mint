package modal

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"strings"
	"sync"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Global Modal Registry
// =============================================================================

// modalRegistry stores all open modal instances globally
// This allows GlobalActionHandler to access modals without InstanceManager dependency
type modalRegistry struct {
	mu     sync.RWMutex
	modals map[*Instance]bool
	order  []*Instance
}

var (
	globalRegistry = &modalRegistry{
		modals: make(map[*Instance]bool),
	}
)

// register adds a modal to the global registry
func (r *modalRegistry) register(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeLocked(inst)
	r.modals[inst] = true
	r.order = append(r.order, inst)
}

// unregister removes a modal from the global registry
func (r *modalRegistry) unregister(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeLocked(inst)
}

// getOpenModals returns all open modals
// The returned slice is ordered from oldest to newest.
func (r *modalRegistry) getOpenModals() []*Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Instance, 0, len(r.order))
	for _, inst := range r.order {
		if inst != nil && inst.isOpen {
			result = append(result, inst)
		}
	}
	return result
}

// hasOpenModal checks if there are any open modals
func (r *modalRegistry) hasOpenModal() bool {
	return r.getTopmostOpenModal() != nil
}

func (r *modalRegistry) getTopmostOpenModal() *Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := len(r.order) - 1; i >= 0; i-- {
		inst := r.order[i]
		if inst != nil && inst.isOpen {
			return inst
		}
	}
	return nil
}

func (r *modalRegistry) removeLocked(inst *Instance) {
	delete(r.modals, inst)
	for i, existing := range r.order {
		if existing == inst {
			r.order = append(r.order[:i], r.order[i+1:]...)
			return
		}
	}
}

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for modal components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	title       string
	modalStyle  style.Style
	borderStyle string
	closeIntent intent.Intent

	// === Modal Props ===
	isOpen          bool
	centered        bool
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool

	// === Layout Props ===
	width   int
	height  int
	padding int

	// === Content ===
	content rtui.VNode
	footer  rtui.VNode

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool

	// === Registration State ===
	registered bool // Tracks if this instance is in the global registry

	// === Visual State ===
	shadowStyle style.Style
	showShadow  bool

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
		key:             proputil.GetString(props, "key", ""),
		title:           proputil.GetString(props, "title", ""),
		modalStyle:      proputil.GetStyle(props, "style", style.Style{}),
		shadowStyle:     getShadowStyleProp(props),
		borderStyle:     proputil.GetString(props, "borderStyle", "single"),
		closeIntent:     proputil.GetIntent(props, "closeIntent", nil),
		isOpen:          proputil.GetBool(props, "isOpen", false),
		centered:        proputil.GetBool(props, "centered", true),
		closeable:       proputil.GetBool(props, "closeable", true),
		closeOnEsc:      proputil.GetBool(props, "closeOnEsc", true),
		closeOnBackdrop: proputil.GetBool(props, "closeOnBackdrop", true),
		width:           proputil.GetInt(props, "width", 40),
		height:          proputil.GetInt(props, "height", 15),
		padding:         proputil.GetInt(props, "padding", 0),
		content:         getChildProp(props, "content"),
		footer:          getChildProp(props, "footer"),
		dirty:           true,
		registered:      false, // Not registered yet
		showShadow:      proputil.GetBool(props, "showShadow", true),
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy() {
	inst.cleanup()
}
func (inst *Instance) OnMount() {}
func (inst *Instance) OnUnmount() {
	inst.cleanup()
}

func (inst *Instance) cleanup() {
	if inst.registered {
		globalRegistry.unregister(inst)
		inst.registered = false
	}
	inst.isOpen = false
	inst.content = nil
	inst.footer = nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldOpen := inst.isOpen
	oldTitle := inst.title
	oldBorderStyle := inst.borderStyle
	oldModalStyle := inst.modalStyle
	oldShadowStyle := inst.shadowStyle
	oldPadding := inst.padding
	oldShowShadow := inst.showShadow
	oldCloseOnEsc := inst.closeOnEsc
	oldCloseOnBackdrop := inst.closeOnBackdrop
	oldHasFooter := inst.footer != nil

	inst.title = proputil.GetString(props, "title", inst.title)
	inst.modalStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.shadowStyle = getShadowStyleProp(props)
	inst.borderStyle = proputil.GetString(props, "borderStyle", inst.borderStyle)
	inst.closeIntent = proputil.GetIntent(props, "closeIntent", nil)
	inst.isOpen = proputil.GetBool(props, "isOpen", inst.isOpen)
	inst.centered = proputil.GetBool(props, "centered", inst.centered)
	inst.closeable = proputil.GetBool(props, "closeable", inst.closeable)
	inst.closeOnEsc = proputil.GetBool(props, "closeOnEsc", inst.closeOnEsc)
	inst.closeOnBackdrop = proputil.GetBool(props, "closeOnBackdrop", inst.closeOnBackdrop)
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.height = proputil.GetInt(props, "height", inst.height)
	inst.padding = proputil.GetInt(props, "padding", inst.padding)
	inst.content = getChildProp(props, "content")
	inst.footer = getChildProp(props, "footer")
	inst.showShadow = proputil.GetBool(props, "showShadow", inst.showShadow)
	// Track modal open/close state in global registry for GlobalActionHandler
	// 1. If state changed from closed to open, register
	if !oldOpen && inst.isOpen {
		globalRegistry.register(inst)
		inst.registered = true
	} else if oldOpen && !inst.isOpen { // 2. If state changed from open to closed, unregister
		globalRegistry.unregister(inst)
		inst.registered = false
	} else if inst.isOpen && !inst.registered {
		// 3. If initially open and not yet registered, register (case: modal created with isOpen=true)
		globalRegistry.register(inst)
		inst.registered = true
	}

	changed := oldOpen != inst.isOpen ||
		oldTitle != inst.title ||
		oldBorderStyle != inst.borderStyle ||
		oldModalStyle != inst.modalStyle ||
		oldShadowStyle != inst.shadowStyle ||
		oldPadding != inst.padding ||
		oldShowShadow != inst.showShadow ||
		oldCloseOnEsc != inst.closeOnEsc ||
		oldCloseOnBackdrop != inst.closeOnBackdrop ||
		oldHasFooter != (inst.footer != nil)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":             inst.key,
		"title":           inst.title,
		"isOpen":          inst.isOpen,
		"centered":        inst.centered,
		"closeable":       inst.closeable,
		"closeOnEsc":      inst.closeOnEsc,
		"closeOnBackdrop": inst.closeOnBackdrop,
		"width":           inst.width,
		"height":          inst.height,
		"padding":         inst.padding,
		"borderStyle":     inst.borderStyle,
		"showShadow":      inst.showShadow,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

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
func (inst *Instance) ShouldCenter() bool {
	// ✨ Phase 1.4: Modal 居中控制
	// 如果 Modal 被显式地设置了 centered=false，或者没有设置（默认 true），
	// 则返回是否使用居中布局
	return inst.centered
}

func (inst *Instance) GetBorder() layout.Border {
	if !inst.isOpen {
		return layout.Border{Style: layout.BorderNone}
	}
	return layout.NewBorder(inst.layoutBorderStyle())
}

func (inst *Instance) GetBoxModel() layout.BoxModel {
	if !inst.isOpen {
		return layout.BoxModel{}
	}

	topPadding := inst.padding
	if inst.title != "" {
		topPadding += 2
	}

	return layout.BoxModel{
		Padding: layout.Padding{
			Top:    topPadding,
			Right:  inst.padding,
			Bottom: inst.padding,
			Left:   inst.padding,
		},
		Border: layout.NewBorder(inst.layoutBorderStyle()),
	}
}

func (inst *Instance) GetFlexStyle() *layout.FlexStyle {
	if !inst.isOpen {
		return nil
	}
	return &layout.FlexStyle{
		Direction: layout.FlexColumn,
		MainAxis:  layout.MainStart,
		CrossAxis: layout.CrossStart,
		Gap:       0,
	}
}

func (inst *Instance) GetSize() (int, int) {
	return inst.width, inst.height
}

func (inst *Instance) SetSize(width, height int) {
	if width > 0 {
		inst.width = width
	}
	if height > 0 {
		inst.height = height
	}
	inst.bounds[2] = width
	inst.bounds[3] = height
}

func (inst *Instance) layoutBorderStyle() layout.BorderStyle {
	switch inst.borderStyle {
	case "double":
		return layout.BorderDouble
	case "rounded":
		return layout.BorderRounded
	case "dashed":
		return layout.BorderDashed
	case "none":
		return layout.BorderNone
	default:
		return layout.BorderSingle
	}
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

	width := maxInt(inst.width, 2)
	height := maxInt(inst.height, 2)

	// Set bounds for hit testing
	inst.bounds = [4]int{x, y, width, height}

	var cmds []paint.DrawCmd
	bodyStyle := inst.resolveBodyStyle()
	chromeStyle := inst.resolveChromeStyle()

	// Get border characters
	borderChars := inst.getBorderChars()

	if inst.showShadow {
		cmds = append(cmds, inst.paintShadow(x, y, width, height)...)
	}

	// Build top border
	topBorder := string(borderChars.topLeft) + strings.Repeat(string(borderChars.horizontal), maxInt(0, width-2)) + string(borderChars.topRight)
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, chromeStyle))

	// Draw title row if present
	bodyStartY := y + 1
	if inst.title != "" && height >= 4 {
		titleRow := string(borderChars.vertical) + inst.composeHeaderRow(width-2) + string(borderChars.vertical)
		cmds = append(cmds, paint.NewTextCmd(x, y+1, titleRow, chromeStyle))
		if titleText := inst.headerTitle(width - 2); titleText != "" {
			cmds = append(cmds, paint.NewTextCmd(x+1, y+1, titleText, inst.resolveTitleStyle()))
		}
		if hintText, hintX := inst.headerHint(width - 2); hintText != "" {
			cmds = append(cmds, paint.NewTextCmd(x+1+hintX, y+1, hintText, inst.resolveHintStyle()))
		}

		// Separator after title
		separator := string(borderChars.leftT) + strings.Repeat(string(borderChars.horizontal), maxInt(0, width-2)) + string(borderChars.rightT)
		cmds = append(cmds, paint.NewTextCmd(x, y+2, separator, chromeStyle))
		bodyStartY = y + 3
	}

	// Draw bottom border
	bottomBorder := string(borderChars.bottomLeft) + strings.Repeat(string(borderChars.horizontal), maxInt(0, width-2)) + string(borderChars.bottomRight)
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, bottomBorder, chromeStyle))

	// Draw side borders and fill the body background so child content sits on a consistent surface.
	for row := bodyStartY; row < y+height-1; row++ {
		cmds = append(cmds, paint.NewTextCmd(x, row, string(borderChars.vertical), chromeStyle))
		if width > 2 {
			cmds = append(cmds, paint.NewTextCmd(x+1, row, strings.Repeat(" ", width-2), bodyStyle))
		}
		cmds = append(cmds, paint.NewTextCmd(x+width-1, row, string(borderChars.vertical), chromeStyle))
	}

	return cmds
}

func (inst *Instance) paintShadow(x, y, width, height int) []paint.DrawCmd {
	if width <= 1 || height <= 1 {
		return nil
	}
	shadowStyle := inst.resolveShadowStyle()
	var cmds []paint.DrawCmd
	for row := y + 1; row < y+height; row++ {
		cmds = append(cmds, paint.NewTextCmd(x+width, row, " ", shadowStyle))
	}
	if width > 2 {
		cmds = append(cmds, paint.NewTextCmd(x+1, y+height, strings.Repeat(" ", width), shadowStyle))
	}
	return cmds
}

func (inst *Instance) composeHeaderRow(innerWidth int) string {
	if innerWidth <= 0 {
		return ""
	}
	return strings.Repeat(" ", innerWidth)
}

func (inst *Instance) headerTitle(innerWidth int) string {
	if innerWidth <= 0 {
		return ""
	}

	available := innerWidth
	if hint := inst.closeHint(); hint != "" {
		available -= paint.StringWidth(hint) + 1
	}
	if available <= 0 {
		return ""
	}

	title := inst.fitDisplayWidth(" "+inst.title, available)
	return inst.padRightDisplayWidth(title, available)
}

func (inst *Instance) headerHint(innerWidth int) (string, int) {
	hint := inst.closeHint()
	if innerWidth <= 0 || hint == "" {
		return "", 0
	}
	hint = inst.fitDisplayWidth(hint, innerWidth)
	return hint, maxInt(0, innerWidth-paint.StringWidth(hint))
}

func (inst *Instance) closeHint() string {
	if inst.closeable && inst.closeOnEsc {
		return "ESC "
	}
	return ""
}

func (inst *Instance) resolveBodyStyle() style.Style {
	s := inst.modalStyle
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	return s
}

func (inst *Instance) resolveChromeStyle() style.Style {
	bodyStyle := inst.resolveBodyStyle()
	chrome := style.NewStyle().Foreground(theme.Border()).Background(bodyStyle.BG)
	if inst.modalStyle.FG != "" {
		chrome.FG = inst.modalStyle.FG
	}
	chrome = chrome.Merge(inst.modalStyle)
	if chrome.BG == "" {
		chrome.BG = bodyStyle.BG
	}
	return chrome
}

func (inst *Instance) resolveTitleStyle() style.Style {
	return inst.resolveBodyStyle().Bold(true)
}

func (inst *Instance) resolveHintStyle() style.Style {
	hintStyle := style.NewStyle().Foreground(theme.Muted()).Background(inst.resolveBodyStyle().BG)
	if inst.modalStyle.BG != "" {
		hintStyle.BG = inst.modalStyle.BG
	}
	return hintStyle
}

func (inst *Instance) resolveShadowStyle() style.Style {
	s := inst.shadowStyle
	if s.BG == "" {
		s = s.Background(theme.Shadow())
	}
	return s
}

func (inst *Instance) fitDisplayWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= width {
		return text
	}

	var b strings.Builder
	current := 0
	for _, r := range text {
		rw := paint.StringWidth(string(r))
		if current+rw > width {
			break
		}
		current += rw
		b.WriteRune(r)
	}
	return b.String()
}

func (inst *Instance) padRightDisplayWidth(text string, width int) string {
	padding := width - paint.StringWidth(text)
	if padding <= 0 {
		return text
	}
	return text + strings.Repeat(" ", padding)
}

// getBorderChars returns the border characters based on borderStyle.
func (inst *Instance) getBorderChars() borderChars {
	switch inst.borderStyle {
	case "none":
		return borderChars{
			horizontal:  ' ',
			vertical:    ' ',
			topLeft:     ' ',
			topRight:    ' ',
			bottomLeft:  ' ',
			bottomRight: ' ',
			leftT:       ' ',
			rightT:      ' ',
		}
	case "double":
		return borderChars{
			horizontal:  '═',
			vertical:    '║',
			topLeft:     '╔',
			topRight:    '╗',
			bottomLeft:  '╚',
			bottomRight: '╝',
			leftT:       '╠',
			rightT:      '╣',
		}
	case "rounded":
		return borderChars{
			horizontal:  '─',
			vertical:    '│',
			topLeft:     '╭',
			topRight:    '╮',
			bottomLeft:  '╰',
			bottomRight: '╯',
			leftT:       '├',
			rightT:      '┤',
		}
	case "dashed":
		return borderChars{
			horizontal:  '─',
			vertical:    '│',
			topLeft:     '┌',
			topRight:    '┐',
			bottomLeft:  '└',
			bottomRight: '┘',
			leftT:       '├',
			rightT:      '┤',
		}
	default: // single
		return borderChars{
			horizontal:  '─',
			vertical:    '│',
			topLeft:     '┌',
			topRight:    '┐',
			bottomLeft:  '└',
			bottomRight: '┘',
			leftT:       '├',
			rightT:      '┤',
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

func (inst *Instance) HandleAction(act *action.Action) bool {
	if !inst.isOpen {
		return false
	}

	switch act.Type {
	case action.ActionCancel, action.ActionQuit:
		if inst.closeable && inst.closeOnEsc {
			inst.requestClose()
		}
		return true
	case action.ActionScroll, action.ActionNavigateUp, action.ActionNavigateDown,
		action.ActionNavigatePageUp, action.ActionNavigatePageDown,
		action.ActionNavigateHome, action.ActionNavigateEnd:
		// Modal doesn't handle navigation - return false to let parent handle it
		return false
	}

	return false
}

// HandleKeyMessage handles keyboard messages for ESC key.
func (inst *Instance) HandleKeyMessage(keyMsg *runtimemsg.KeyMsg) bool {
	if !inst.isOpen {
		return false
	}

	// ESC to close
	if keyMsg.Special == runtimeplatform.KeyEscape {
		if inst.closeable && inst.closeOnEsc {
			inst.requestClose()
		}
		return true
	}

	return false
}

// HandleMouseMessage handles mouse messages for clicking outside modal.
func (inst *Instance) HandleMouseMessage(mouseMsg *runtimemsg.MouseMsg) bool {
	if !inst.isOpen {
		return false
	}

	if mouseMsg.Action == runtimemsg.MouseActionPress {
		// Check if click is outside modal bounds
		if !inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			if inst.closeable && inst.closeOnBackdrop {
				inst.requestClose()
			}
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

func (inst *Instance) requestClose() {
	if !inst.isOpen {
		return
	}
	inst.isOpen = false
	inst.dirty = true
	if inst.registered {
		globalRegistry.unregister(inst)
		inst.registered = false
	}
	inst.emitCloseIntent()
}

// =============================================================================
// Bounds Support
// =============================================================================

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetPosition(x, y int) {
	inst.bounds[0] = x
	inst.bounds[1] = y
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

func getShadowStyleProp(props rtui.Props) style.Style {
	if v, ok := props["shadowStyle"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getChildProp(props rtui.Props, key string) rtui.VNode {
	if v, ok := props[key]; ok {
		if vn, ok := v.(rtui.VNode); ok {
			return vn
		}
	}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
