package optiongroup

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

// =============================================================================
// Parent Instance Registry
// =============================================================================

// parentRegistry stores mapping from parent key to parent Instance.
var parentRegistry = struct {
	sync.RWMutex
	registry map[string]*Instance
}{
	registry: make(map[string]*Instance),
}

// registerParent registers an OptionGroup Instance by its key.
func registerParent(key string, inst *Instance) {
	if key == "" || inst == nil {
		return
	}
	parentRegistry.Lock()
	parentRegistry.registry[key] = inst
	parentRegistry.Unlock()
}

// unregisterParent unregisters an OptionGroup Instance.
func unregisterParent(key string) {
	if key == "" {
		return
	}
	parentRegistry.Lock()
	delete(parentRegistry.registry, key)
	parentRegistry.Unlock()
}

// lookupParent looks up an OptionGroup Instance by parent key.
func lookupParent(key string) *Instance {
	if key == "" {
		return nil
	}
	parentRegistry.RLock()
	defer parentRegistry.RUnlock()
	return parentRegistry.registry[key]
}

// =============================================================================
// Option - Single Option Component (Child of OptionGroup)
// =============================================================================

// SelectOptionFunc is the callback type for when an option is selected.
type SelectOptionFunc func(value string)

// =============================================================================
// OptionVNode - Description Only
// =============================================================================

// OptionVNode represents a single option within an OptionGroup.
// Each option is an independent FocusableInstance, allowing precise mouse targeting.
type OptionVNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string
	idx int // Index in the option list (for keyboard navigation reference)

	// === Visual Props ===
	value string
	label string
	style style.Style

	// === State Props (declarative, synced from parent) ===
	selected  bool
	disabled  bool
	mode      SelectMode

	// === Parent Reference ===
	// selectFunc is called when this option is clicked/selected
	// Can be set via SetSelectFunc() method (deferred binding)
	selectFunc SelectOptionFunc

	// === Box Model ===
	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*OptionVNode)(nil)
	_ rtui.InstanceFactory = (*OptionVNode)(nil)
	_ rtui.BoxModel        = (*OptionVNode)(nil)
)

// NewOptionVNode creates a new OptionVNode.
func NewOptionVNode(value, label string, idx int, mode SelectMode, selectFunc SelectOptionFunc) *OptionVNode {
	return &OptionVNode{
		ElementVNode: rtui.NewElement("option"),
		key:          "opt-" + value,
		value:        value,
		label:        label,
		idx:          idx,
		mode:         mode,
		selectFunc:   selectFunc,
	}
}

// NewOptionVNodeDeferred creates a new OptionVNode without the selectFunc.
// The selectFunc can be set later via SetSelectFunc() method.
// This is used when creating OptionVNodes from the parent VNode before the parent Instance is available.
func NewOptionVNodeDeferred(value, label string, idx int, mode SelectMode) *OptionVNode {
	return &OptionVNode{
		ElementVNode: rtui.NewElement("option"),
		key:          "opt-" + value,
		value:        value,
		label:        label,
		idx:          idx,
		mode:         mode,
		selectFunc:   nil, // Will be set later by parent
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (o *OptionVNode) Key() string        { return o.key }
func (o *OptionVNode) SetKey(key string) rtui.VNode {
	o.key = key
	return o
}
func (o *OptionVNode) Tag() string        { return "option" }
func (o *OptionVNode) Style() style.Style { return o.style }
func (o *OptionVNode) SetStyle(s style.Style) rtui.VNode {
	o.style = s
	return o
}
func (o *OptionVNode) Children() []rtui.VNode { return nil }
func (o *OptionVNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return o
}
func (o *OptionVNode) GetLayer() rtui.Layer { return rtui.LayerBase }
func (o *OptionVNode) SetLayer(l rtui.Layer) rtui.VNode { return o }

func (o *OptionVNode) Props() rtui.Props {
	return rtui.Props{
		"key":        o.key,
		"value":      o.value,
		"label":      o.label,
		"style":      o.style,
		"selected":   o.selected,
		"disabled":   o.disabled,
		"mode":       o.mode,
		"selectFunc": o.selectFunc,
		"idx":        o.idx,
	}
}

func (o *OptionVNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		o.key = v
	}
	if v, ok := p["value"].(string); ok {
		o.value = v
	}
	if v, ok := p["label"].(string); ok {
		o.label = v
	}
	if v, ok := p["style"].(style.Style); ok {
		o.style = v
	}
	if v, ok := p["selected"].(bool); ok {
		o.selected = v
	}
	if v, ok := p["disabled"].(bool); ok {
		o.disabled = v
	}
	if v, ok := p["mode"].(SelectMode); ok {
		o.mode = v
	}
	if v, ok := p["selectFunc"].(SelectOptionFunc); ok {
		o.selectFunc = v
	}
	if v, ok := p["idx"].(int); ok {
		o.idx = v
	}
	return o
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (o *OptionVNode) CreateInstance() rtui.ComponentInstance {
	return NewOptionInstance(o.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (o *OptionVNode) SetSelected(selected bool) *OptionVNode {
	o.selected = selected
	return o
}

func (o *OptionVNode) SetDisabled(disabled bool) *OptionVNode {
	o.disabled = disabled
	return o
}

func (o *OptionVNode) SetStyleProps(s style.Style) *OptionVNode {
	o.style = s
	return o
}

// SetSelectFunc sets the parent callback for selection.
// This allows deferred binding when the parent Instance is available.
func (o *OptionVNode) SetSelectFunc(fn SelectOptionFunc) *OptionVNode {
	o.selectFunc = fn
	return o
}

// GetSelectFunc returns the selection callback.
func (o *OptionVNode) GetSelectFunc() SelectOptionFunc {
	return o.selectFunc
}

// =============================================================================
// OptionInstance - Runtime Entity
// =============================================================================

// OptionInstance is the runtime entity for a single option.
type OptionInstance struct {
	// === Identification ===
	key  string
	idx  int // Index in parent's option list
	value string
	label string

	// === State (synced from parent) ===
	state control.InteractionState

	// === Mode ===
	mode SelectMode

	// === Parent Reference ===
	selectFunc SelectOptionFunc
	parentKey  string // Key of parent OptionGroup Instance (for callback lookup)

	// === Style/Base ===
	optionStyle style.Style
	bounds      [4]int // x, y, w, h
	dirty       bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	// === Behaviors ===
	behaviors *control.BehaviorList
}

var (
	_ rtui.ComponentInstance     = (*OptionInstance)(nil)
	_ rtui.PaintableInstance     = (*OptionInstance)(nil)
	_ rtui.FocusableInstance     = (*OptionInstance)(nil)
	_ rtui.ActionHandlerInstance = (*OptionInstance)(nil)
	_ control.Instance           = (*OptionInstance)(nil)
)

// NewOptionInstance creates a new OptionInstance from props.
func NewOptionInstance(props rtui.Props) *OptionInstance {
	value := getStringProp(props, "value", "")
	label := getStringProp(props, "label", value)

	inst := &OptionInstance{
		key:         getStringProp(props, "key", "opt-"+value),
		idx:         getIntProp(props, "idx", 0),
		value:       value,
		label:       label,
		optionStyle: getStyleProp(props),
		mode:        getSelectModeProp(props, ModeSingle),
		selectFunc:  getSelectFuncProp(props),
		dirty:       true,
	}

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: getBoolProp(props, "disabled", false),
	}

	// Initialize behaviors
	inst.initBehaviors()

	return inst
}

func (inst *OptionInstance) initBehaviors() {
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *OptionInstance) Key() string          { return inst.key }
func (inst *OptionInstance) SetKey(key string)    { inst.key = key }

func (inst *OptionInstance) Init(props rtui.Props) {
	inst.SetProps(props)
}

func (inst *OptionInstance) Destroy() {
	inst.behaviors.OnUnmount(inst)
}

func (inst *OptionInstance) OnMount() {
	inst.behaviors.OnMount(inst)
}

func (inst *OptionInstance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

func (inst *OptionInstance) SetProps(props rtui.Props) bool {
	oldDisabled := inst.state.Disabled
	oldSelected := inst.state.Focused // Useocused as proxy for dirty check

	inst.value = getStringProp(props, "value", inst.value)
	inst.label = getStringProp(props, "label", inst.label)
	inst.optionStyle = getStyleProp(props)
	inst.mode = getSelectModeProp(props, inst.mode)
	inst.selectFunc = getSelectFuncProp(props)

	if v, ok := props["disabled"].(bool); ok {
		inst.state.Disabled = v
	}
	if _, ok := props["selected"].(bool); ok {
		// Selected is stored externally by parent, but we track focus
	}

	changed := oldDisabled != inst.state.Disabled || oldSelected != inst.state.Focused

	if changed {
		inst.dirty = true
	}

	return changed
}

func (inst *OptionInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":      inst.key,
		"value":    inst.value,
		"label":    inst.label,
		"disabled": inst.state.Disabled,
		"idx":      inst.idx,
	}
}

func (inst *OptionInstance) MarkDirty() { inst.dirty = true }
func (inst *OptionInstance) IsDirty() bool  { return inst.dirty }
func (inst *OptionInstance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// PaintableInstance Interface
// =============================================================================

func (inst *OptionInstance) Paint(x, y int) []paint.DrawCmd {
	s := inst.resolveStyle()

	// Build indicator based on mode
	var indicator string
	if inst.mode == ModeSingle {
		// Radio style
		if inst.state.Focused {
			indicator = "(•)"
		} else {
			indicator = "( )"
		}
	} else {
		// Checkbox style
		if inst.state.Focused {
			indicator = "[X]"
		} else {
			indicator = "[ ]"
		}
	}

	// Build option text: indicator + space + label
	var optionText string
	if inst.label != "" {
		optionText = indicator + " " + inst.label
	} else {
		optionText = indicator + " " + inst.value
	}

	return []paint.DrawCmd{
		{
			X:     x,
			Y:     y,
			Text:  optionText,
			Style: s,
		},
	}
}

func (inst *OptionInstance) resolveStyle() style.Style {
	s := inst.optionStyle

	// Apply default colors if not set
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	// Disabled state
	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		// Focused state: highlight
		s = s.Foreground(theme.Highlight()).Underline(true)
	} else if inst.state.Hovered {
		// Hovered state
		s = s.Bold(true)
	}

	return s
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *OptionInstance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		inst.state.Focused = focused
		inst.dirty = true
	}
}

func (inst *OptionInstance) HasFocus() bool { return inst.state.Focused }
func (inst *OptionInstance) IsDisabled() bool { return inst.state.Disabled }

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *OptionInstance) HandleAction(act *action.Action) bool {
	// Let behaviors process first
	if inst.behaviors.OnAction(inst, act) {
		return true
	}

	if inst.state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionClick, action.ActionEnter, action.ActionToggle:
		// Select this option
		if inst.selectFunc != nil {
			inst.selectFunc(inst.value)
		}
		return true
	}

	return false
}

// =============================================================================
// control.Instance Interface
// =============================================================================

func (inst *OptionInstance) GetState() *control.InteractionState {
	return &inst.state
}

func (inst *OptionInstance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

func (inst *OptionInstance) EmitIntent(i intent.Intent) {
	// Option doesn't emit intents directly, it calls selectFunc
}

func (inst *OptionInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *OptionInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *OptionInstance) GetStyle() style.Style {
	return inst.optionStyle
}

func (inst *OptionInstance) SetStyle(s style.Style) {
	inst.optionStyle = s
}

func (inst *OptionInstance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "idx":
		return inst.idx, true
	case "value":
		return inst.value, true
	default:
		return nil, false
	}
}

func (inst *OptionInstance) SetProp(key string, value interface{}) {
	switch key {
	case "disabled":
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	}
}

func (inst *OptionInstance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *OptionInstance) ClearDirty() {
	inst.dirty = false
}

// =============================================================================
// Helper Props Extraction
// =============================================================================

func getSelectFuncProp(props rtui.Props) SelectOptionFunc {
	if v, ok := props["selectFunc"].(SelectOptionFunc); ok {
		return v
	}
	return nil
}
