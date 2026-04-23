package drawer

import (
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
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

// =============================================================================
// Global Drawer Registry
// =============================================================================

type drawerRegistry struct {
	mu      sync.RWMutex
	drawers map[*Instance]bool
	order   []*Instance
}

var (
	globalRegistry = &drawerRegistry{
		drawers: make(map[*Instance]bool),
	}
)

func (r *drawerRegistry) register(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeLocked(inst)
	r.drawers[inst] = true
	r.order = append(r.order, inst)
}

func (r *drawerRegistry) unregister(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeLocked(inst)
}

func (r *drawerRegistry) removeLocked(inst *Instance) {
	delete(r.drawers, inst)
	for i, existing := range r.order {
		if existing == inst {
			r.order = append(r.order[:i], r.order[i+1:]...)
			return
		}
	}
}

// =============================================================================
// Instance — Runtime Entity
// =============================================================================

// Instance is the runtime entity for drawer components.
type Instance struct {
	// === Identification ===
	key string

	// === Props ===
	placement       Placement
	title           string
	drawerStyle     style.Style
	borderStyle     string
	closeIntent     intent.Intent
	isOpen          bool
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool

	// === Layout ===
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
	registered bool

	// === Visual State ===
	shadowStyle style.Style
	showShadow  bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)
}

// Compile-time interface checks.
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

// NewInstance creates a new Instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:             proputil.GetString(props, propKey, ""),
		placement:       Placement(proputil.GetInt(props, propPlacement, int(PlacementRight))),
		title:           proputil.GetString(props, propTitle, ""),
		drawerStyle:     proputil.GetStyle(props, propDrawerStyle, style.Style{}),
		shadowStyle:     getShadowStyleProp(props),
		borderStyle:     proputil.GetString(props, propBorderStyle, "single"),
		closeIntent:     proputil.GetIntent(props, propCloseIntent, nil),
		isOpen:          proputil.GetBool(props, propIsOpen, false),
		closeable:       proputil.GetBool(props, propCloseable, true),
		closeOnEsc:      proputil.GetBool(props, propCloseOnEsc, true),
		closeOnBackdrop: proputil.GetBool(props, propCloseOnBackdrop, true),
		width:           proputil.GetInt(props, propWidth, 30),
		height:          proputil.GetInt(props, propHeight, 15),
		padding:         proputil.GetInt(props, propPadding, 0),
		content:         getChildProp(props, propContent),
		footer:          getChildProp(props, propFooter),
		dirty:           true,
		registered:      false,
		showShadow:      proputil.GetBool(props, propShowShadow, true),
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
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() { inst.cleanup() }

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
	oldDrawerStyle := inst.drawerStyle
	oldShadowStyle := inst.shadowStyle
	oldPadding := inst.padding
	oldShowShadow := inst.showShadow
	oldCloseOnEsc := inst.closeOnEsc
	oldCloseOnBackdrop := inst.closeOnBackdrop
	oldHasFooter := inst.footer != nil
	oldPlacement := inst.placement

	inst.placement = Placement(proputil.GetInt(props, propPlacement, int(inst.placement)))
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.drawerStyle = proputil.GetStyle(props, propDrawerStyle, style.Style{})
	inst.shadowStyle = getShadowStyleProp(props)
	inst.borderStyle = proputil.GetString(props, propBorderStyle, inst.borderStyle)
	inst.closeIntent = proputil.GetIntent(props, propCloseIntent, nil)
	inst.isOpen = proputil.GetBool(props, propIsOpen, inst.isOpen)
	inst.closeable = proputil.GetBool(props, propCloseable, inst.closeable)
	inst.closeOnEsc = proputil.GetBool(props, propCloseOnEsc, inst.closeOnEsc)
	inst.closeOnBackdrop = proputil.GetBool(props, propCloseOnBackdrop, inst.closeOnBackdrop)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.height = proputil.GetInt(props, propHeight, inst.height)
	inst.padding = proputil.GetInt(props, propPadding, inst.padding)
	inst.content = getChildProp(props, propContent)
	inst.footer = getChildProp(props, propFooter)
	inst.showShadow = proputil.GetBool(props, propShowShadow, inst.showShadow)

	// Track open/close state in global registry
	if !oldOpen && inst.isOpen {
		globalRegistry.register(inst)
		inst.registered = true
	} else if oldOpen && !inst.isOpen {
		globalRegistry.unregister(inst)
		inst.registered = false
	} else if inst.isOpen && !inst.registered {
		globalRegistry.register(inst)
		inst.registered = true
	}

	changed := oldOpen != inst.isOpen ||
		oldTitle != inst.title ||
		oldBorderStyle != inst.borderStyle ||
		oldDrawerStyle != inst.drawerStyle ||
		oldShadowStyle != inst.shadowStyle ||
		oldPadding != inst.padding ||
		oldShowShadow != inst.showShadow ||
		oldCloseOnEsc != inst.closeOnEsc ||
		oldCloseOnBackdrop != inst.closeOnBackdrop ||
		oldHasFooter != (inst.footer != nil) ||
		oldPlacement != inst.placement
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:             inst.key,
		propPlacement:       int(inst.placement),
		propTitle:           inst.title,
		propIsOpen:          inst.isOpen,
		propCloseable:       inst.closeable,
		propCloseOnEsc:      inst.closeOnEsc,
		propCloseOnBackdrop: inst.closeOnBackdrop,
		propWidth:           inst.width,
		propHeight:          inst.height,
		propPadding:         inst.padding,
		propBorderStyle:     inst.borderStyle,
		propShowShadow:      inst.showShadow,
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
// If the drawer is not open, it takes no space.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if !inst.isOpen {
		return layout.Size{Width: 0, Height: 0}
	}

	width := inst.width
	height := inst.height

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
	case propDouble:
		return layout.BorderDouble
	case propRounded:
		return layout.BorderRounded
	case propDashed:
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

// Paint draws the drawer panel at the given position.
// The position is the top-left corner of the drawer area.
// The drawer draws its border, title row, and body background.
// Content children are rendered by the framework.
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
	borderChars := inst.getBorderChars()

	if inst.showShadow {
		cmds = append(cmds, inst.paintShadow(x, y, width, height)...)
	}

	// Top border
	topBorder := string(borderChars.topLeft) + strings.Repeat(string(borderChars.horizontal), maxInt(0, width-2)) + string(borderChars.topRight)
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, chromeStyle))

	// Title row
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

	// Bottom border
	bottomBorder := string(borderChars.bottomLeft) + strings.Repeat(string(borderChars.horizontal), maxInt(0, width-2)) + string(borderChars.bottomRight)
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, bottomBorder, chromeStyle))

	// Side borders and body background
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
	// Shadow on the right edge
	for row := y + 1; row < y+height; row++ {
		cmds = append(cmds, paint.NewTextCmd(x+width, row, " ", shadowStyle))
	}
	// Shadow on the bottom edge
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

// =============================================================================
// Style Resolution
// =============================================================================

func (inst *Instance) resolveBodyStyle() style.Style {
	s := inst.drawerStyle
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
	if inst.drawerStyle.FG != "" {
		chrome.FG = inst.drawerStyle.FG
	}
	chrome = chrome.Merge(inst.drawerStyle)
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
	if inst.drawerStyle.BG != "" {
		hintStyle.BG = inst.drawerStyle.BG
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
	case propDouble:
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
	case propRounded:
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
	case propDashed:
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
	}

	return false
}

// HandleKeyMessage handles keyboard messages for ESC key.
func (inst *Instance) HandleKeyMessage(keyMsg *runtimemsg.KeyMsg) bool {
	if !inst.isOpen {
		return false
	}

	if keyMsg.Special == runtimeplatform.KeyEscape {
		if inst.closeable && inst.closeOnEsc {
			inst.requestClose()
		}
		return true
	}

	return false
}

// HandleMouseMessage handles mouse messages for clicking outside drawer.
func (inst *Instance) HandleMouseMessage(mouseMsg *runtimemsg.MouseMsg) bool {
	if !inst.isOpen {
		return false
	}

	if mouseMsg.Action == runtimemsg.MouseActionPress {
		if !inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			if inst.closeable && inst.closeOnBackdrop {
				inst.requestClose()
			}
			return true
		}
	}

	return false
}

func (inst *Instance) containsPoint(x, y int) bool {
	if inst.bounds[2] <= 0 || inst.bounds[3] <= 0 {
		return false
	}
	return x >= inst.bounds[0] && x < inst.bounds[0]+inst.bounds[2] &&
		y >= inst.bounds[1] && y < inst.bounds[1]+inst.bounds[3]
}

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
	if v, ok := props[propShadowStyle]; ok {
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
