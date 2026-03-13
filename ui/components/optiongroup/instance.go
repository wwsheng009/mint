package optiongroup

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"strings"

	fcontext "github.com/wwsheng009/mint/runtime/context"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

// Context Key for OptionGroup (Phase 2)
const OptionGroupContext fcontext.ContextKey = "github.com/wwsheng009/mint/ui/components/optiongroup:group"

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for OptionGroup components.
// It persists across renders and holds all state.
//
// Architecture (after refactoring):
// - Options are now separate child Fiber nodes (OptionInstance)
// - This Instance manages the selection state and option callbacks
// - Paint only renders the group label (if any)
// - Options are painted by child OptionInstance nodes
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label        string
	optionStyle  style.Style
	selectIntent intent.Intent

	// === Runtime State (managed by instance) ===
	state       control.InteractionState
	mode        SelectMode
	options     []Option
	orientation Orientation // from vnode.go: OrientationVertical/Horizontal
	spacing     int        // gap between options
	selected    string     // For ModeSingle
	selecteds   []string   // For ModeMultiple
	bounds      [4]int     // x, y, w, h
	dirty       bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	// === VNode Reference (for updating child callbacks) ===
	vnode *VNode

	// === Child Instances (for direct callback setup) ===
	childInstances []*OptionInstance

	// === Behaviors ===
	behaviors *control.BehaviorList
}

// ===== Instance Tree Methods (Mint Runtime 2.0 - Phase 1) =====

// Parent implements rtui.TreeNode/intent.TreeComponent interface (for intent bubble).
// Returns nil since OptionGroup is typically a root component.
func (inst *Instance) Parent() interface{} {
	return nil
}

// Children implements rtui.TreeNode/TreeContainer interface
func (inst *Instance) Children() []rtui.ComponentInstance {
	children := make([]rtui.ComponentInstance, len(inst.childInstances))
	for i, child := range inst.childInstances {
		children[i] = child
	}
	return children
}

// AddChild implements TreeContainer interface
func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	if optInst, ok := child.(*OptionInstance); ok {
		// Check if already added (compare by key, not pointer)
		// This prevents duplicate children with the same value from Fiber diffing
		for i, existing := range inst.childInstances {
			if existing.Key() == optInst.Key() {
				// Already exists - replace with new instance
				// This fixes memory leak where Fiber diffing creates new instances
				// but old instances remain in childInstances
				oldOpt := existing
				// Clear parent reference on old instance to prevent memory leak
				oldOpt.parent = nil
				// Replace with new instance
				inst.childInstances[i] = optInst
				optInst.parent = inst
				return
			}
		}
		// Not found, add as new child
		inst.childInstances = append(inst.childInstances, optInst)
		// Set parent reference on child (Instance Tree)
		optInst.parent = inst
	}
}

// RemoveChild implements TreeContainer interface
func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	if optInst, ok := child.(*OptionInstance); ok {
		for i, existing := range inst.childInstances {
			if existing == optInst {
				inst.childInstances = append(inst.childInstances[:i], inst.childInstances[i+1:]...)
				// Clear parent reference on child (Instance Tree)
				optInst.parent = nil
				break
			}
		}
	}
}

// ClearChildren implements TreeContainer interface
func (inst *Instance) ClearChildren() {
	// Clear parent references on all children
	for _, child := range inst.childInstances {
		child.parent = nil
	}
	inst.childInstances = inst.childInstances[:0]
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ control.Instance           = (*Instance)(nil)
	_ rtui.TreeNode              = (*Instance)(nil)      // Instance Tree Phase 1
	_ rtui.TreeContainer         = (*Instance)(nil)      // Instance Tree Phase 1
	_ intent.IntentHandler       = (*Instance)(nil)      // Intent Bubble
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new OptionGroupInstance from props.
func NewInstance(props rtui.Props) *Instance {
	selected := proputil.GetString(props, "selected", "")
	selecteds := getStringsProp(props, "selecteds", []string{})

	// If selected is set but selecteds is empty, initialize it
	if selected != "" && len(selecteds) == 0 {
		selecteds = []string{selected}
	}

	inst := &Instance{
		key:          proputil.GetString(props, "key", ""),
		label:        proputil.GetString(props, "label", ""),
		optionStyle:  proputil.GetStyle(props, "style", style.Style{}),
		selectIntent: proputil.GetIntent(props, "selectIntent", nil),
		mode:         getSelectModeProp(props, ModeSingle),
		options:      getOptionsProp(props, []Option{}),
		selected:     selected,
		selecteds:    selecteds,
		dirty:        true,
	}

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, "disabled", false),
	}

	// Initialize child Option instances from options list
	for i, opt := range inst.options {
		childProps := rtui.Props{
			propKey:   inst.key + "-opt-" + opt.Value,
			"value": opt.Value,
			propLabel: opt.Label,
			"idx":   i,
			propMode:  inst.mode,
		}
		childInst := NewOptionInstance(childProps)
		inst.childInstances = append(inst.childInstances, childInst)
		// Set parent reference
		childInst.parent = inst
	}

	// Initialize behaviors
	inst.initBehaviors()

	return inst
}

// initBehaviors initializes the behavior composition.
func (inst *Instance) initBehaviors() {
	// Compose behaviors
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *Instance) Destroy() {
	inst.behaviors.OnUnmount(inst)
}

// OnMount implements ComponentInstance.
// After refactoring: Update child option callbacks to ensure they can select values.
func (inst *Instance) OnMount() {
	inst.behaviors.OnMount(inst)

	// Update child option callbacks AND set parent references (Phase 2)
	for _, child := range inst.childInstances {
		// Set direct parent reference for backward compatibility
		child.parent = inst
	}

	// OptionGroup Context is injected at the Fiber level during provider creation
	// See fiber_util.go for context injection logic
}

// updateOptionCallbacks updates the selectFunc on all child option instances.
// This is called during OnMount to ensure children have access to the parent's SelectOption method.
func (inst *Instance) updateOptionCallbacks() {
	// Access children through the associated VNode if available
	// In a production system, this would be done via a parent-child reference
	// For now, we'll defer this to another approach

	// Note: The actual implementation would need access to child instances.
	// Since the Fiber system doesn't provide direct parent-to-child instance references,
	// we'll use a different mechanism: update the VNode so future children get the callback.

	// The VNode.optionSelectFunc is already set in CreateInstance,
	// so future renders will have the callback.
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldLabel := inst.label
	oldDisabled := inst.state.Disabled
	oldSelected := inst.selected
	oldSelecteds := inst.selecteds
	oldIntent := inst.selectIntent
	oldOptions := make([]Option, len(inst.options))
	copy(oldOptions, inst.options)

	inst.label = proputil.GetString(props, "label", inst.label)
	inst.optionStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.selectIntent = proputil.GetIntent(props, "selectIntent", nil)
	inst.mode = getSelectModeProp(props, inst.mode)
	inst.options = getOptionsProp(props, inst.options)
	inst.orientation = getOrientationProp(props, inst.orientation)
	inst.spacing = proputil.GetInt(props, "spacing", inst.spacing)
	inst.selected = proputil.GetString(props, "selected", inst.selected)
	inst.selecteds = getStringsProp(props, "selecteds", inst.selecteds)

	newDisabled := proputil.GetBool(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	// Check if options changed - if so, rebuild children
	optionsChanged := !equalOptions(oldOptions, inst.options)
	if optionsChanged {
		inst.rebuildChildInstances()
	}

	// Check if props changed
	changed := oldLabel != inst.label ||
		oldDisabled != inst.state.Disabled ||
		oldSelected != inst.selected ||
		oldIntent != inst.selectIntent ||
		optionsChanged

	if changed {
		inst.dirty = true
	}

	// Check if selecteds changed (slice comparison)
	if len(oldSelecteds) != len(inst.selecteds) {
		inst.dirty = true
		changed = true
	} else {
		for i := range oldSelecteds {
			if oldSelecteds[i] != inst.selecteds[i] {
				inst.dirty = true
				changed = true
				break
			}
		}
	}

	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:       inst.key,
		propLabel:     inst.label,
		propDisabled:  inst.state.Disabled,
		propMode:      inst.mode,
		propSelected:  inst.selected,
		propSelecteds: inst.selecteds,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for OptionGroup).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
//
// Note: OptionGroup does not render anything directly.
// - Label is rendered as a separate text child node (see VNode.Children())
// - Options are rendered by child OptionInstance nodes
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// No direct rendering - all content is rendered by child nodes
	return []paint.DrawCmd{}
}

// resolveStyle resolves the visual style based on state for the group label.
func (inst *Instance) resolveStyle() style.Style {
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
		s = s.Foreground(theme.Focus()).Bold(true)
	}

	return s
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

// SetFocus implements FocusableInstance.
// Note: After refactoring, individual options capture focus, not the group.
// This method can still be used for container-level focus (e.g., label highlighting).
func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		inst.state.Focused = focused
		inst.dirty = true
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
}

// HasFocus implements FocusableInstance.
func (inst *Instance) HasFocus() bool {
	return inst.state.Focused
}

// IsDisabled implements FocusableInstance.
func (inst *Instance) IsDisabled() bool {
	return inst.state.Disabled
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

// HandleAction implements ActionHandlerInstance.
//
// After refactoring: Navigation actions are handled by FocusManager at the option level.
// This handler can be used for group-level actions if needed.
func (inst *Instance) HandleAction(act *action.Action) bool {
	// Let behaviors process first
	if inst.behaviors.OnAction(inst, act) {
		return true
	}

	if inst.state.Disabled {
		return false
	}

	// Currently no group-level actions, options handle their own clicks/enters
	return false
}

// =============================================================================
// OptionGroup-specific Methods
// =============================================================================

// SelectOption selects or deselects an option based on mode.
func (inst *Instance) SelectOption(value string) {
	inst.dirty = true

	if inst.mode == ModeSingle {
		// Single-select: set the selected value
		inst.selected = value
		inst.selecteds = []string{value}

		// Update child instances' selected state
		inst.updateChildInstances()

		// Emit FieldChangeIntent with runtime value
		inst.emitFieldChange(value)
	} else {
		// Multi-select: toggle the option
		idx := -1
		for i, v := range inst.selecteds {
			if v == value {
				idx = i
				break
			}
		}

		if idx >= 0 {
			// Remove from selection
			inst.selecteds = append(inst.selecteds[:idx], inst.selecteds[idx+1:]...)
		} else {
			// Add to selection
			inst.selecteds = append(inst.selecteds, value)
		}

		// Update child instances' selected state
		inst.updateChildInstances()

		// Emit FieldChangeIntent with comma-separated values
		valueStr := strings.Join(inst.selecteds, ",")
		inst.emitFieldChange(valueStr)
	}
}

// updateChildInstances updates all child Option instances with the current selected state.
// This ensures Paint() renders the correct selected indicator.
func (inst *Instance) updateChildInstances() {
	for _, child := range inst.childInstances {
		var isSelected bool
		if inst.mode == ModeSingle {
			isSelected = (inst.selected == child.value)
		} else {
			for _, v := range inst.selecteds {
				if v == child.value {
					isSelected = true
					break
				}
			}
		}

		// Update child's selected state through Props
		child.SetProps(rtui.Props{propSelected: isSelected})
	}
}

// DeselectOption removes an option from the selection (for ModeMultiple).
// Phase 5: Added to support Intent Bubble with toggle semantics.
func (inst *Instance) DeselectOption(value string) {
	inst.dirty = true

	if inst.mode == ModeSingle {
		// Single-select: clear selection
		inst.selected = ""
		inst.selecteds = []string{}

		// Update child instances' selected state
		inst.updateChildInstances()

		inst.emitFieldChange("")
	} else {
		// Multi-select: remove the option
		idx := -1
		for i, v := range inst.selecteds {
			if v == value {
				idx = i
				break
			}
		}

		if idx >= 0 {
			inst.selecteds = append(inst.selecteds[:idx], inst.selecteds[idx+1:]...)

			// Update child instances' selected state
			inst.updateChildInstances()

			// Emit FieldChangeIntent with comma-separated values
			valueStr := strings.Join(inst.selecteds, ",")
			inst.emitFieldChange(valueStr)
		}
	}
}

// isOptionSelected checks if an option is currently selected.
func (inst *Instance) isOptionSelected(value string) bool {
	if inst.mode == ModeSingle {
		return inst.selected == value
	}
	for _, v := range inst.selecteds {
		if v == value {
			return true
		}
	}
	return false
}

// emitFieldChange emits a FieldChangeIntent if a FieldBinding is set.
func (inst *Instance) emitFieldChange(value string) {
	if inst.intentEmitter != nil {
		if fieldIntent, ok := inst.selectIntent.(intent.FieldIntent); ok {
			changeIntent := intent.FieldChangeIntent{
				Field: fieldIntent.GetField(),
				Value: value,
			}
			inst.intentEmitter(changeIntent)
		} else if inst.selectIntent != nil {
			// Fallback: emit the original intent
			inst.intentEmitter(inst.selectIntent)
		}
	}
}

// SetSelected sets the selected value (for ModeSingle).
func (inst *Instance) SetSelected(selected string) {
	inst.selected = selected
	inst.selecteds = []string{selected}
	inst.dirty = true
}

// SetSelecteds sets the selected values (for ModeMultiple).
func (inst *Instance) SetSelecteds(selecteds []string) {
	inst.selecteds = selecteds
	if len(selecteds) > 0 {
		inst.selected = selecteds[0] // For backwards compatibility
	}
	inst.dirty = true
}

// GetSelected returns the selected value (for ModeSingle).
func (inst *Instance) GetSelected() string {
	return inst.selected
}

// GetSelecteds returns the selected values (for ModeMultiple).
func (inst *Instance) GetSelecteds() []string {
	return inst.selecteds
}

// Mode returns the selection mode.
func (inst *Instance) Mode() SelectMode {
	return inst.mode
}

// =============================================================================
// control.Instance Interface (for Behaviors)
// =============================================================================

// GetState returns the interaction state.
func (inst *Instance) GetState() *control.InteractionState {
	return &inst.state
}

// SetState sets the interaction state.
func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

// EmitIntent emits an intent.
func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// GetStyle returns the visual style.
func (inst *Instance) GetStyle() style.Style {
	return inst.optionStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.optionStyle = s
}

func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propMode:
		return inst.mode, true
	case propSelected:
		return inst.selected, true
	case propSelecteds:
		return inst.selecteds, true
	case propLabel:
		return inst.label, true
	case propSelectIntent:
		return inst.selectIntent, true
	default:
		return nil, false
	}
}

// SetProp sets a prop value.
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case propDisabled:
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case propSelected:
		if v, ok := value.(string); ok {
			inst.SetSelected(v)
		}
	case propSelecteds:
		if v, ok := value.([]string); ok {
			inst.SetSelecteds(v)
		}
	case propMode:
		if v, ok := value.(SelectMode); ok {
			inst.mode = v
			inst.dirty = true
		}
	}
}

// SetIntentEmitter sets the intent emitter function.
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// =============================================================================
// IntentHandler Interface (Phase 5: Intent Bubble Migration)
// =============================================================================

// HandleIntent implements the intent.IntentHandler interface.
// This method is called when intents bubble up the instance tree.
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	switch v := i.(type) {
	case OptionSelectIntent:
		// Ensure this intent is for the correct group
		if v.GroupKey != inst.key {
			return false // Not for this group, continue bubbling
		}

		// Handle the selection based on mode
		if v.Mode == ModeSingle {
			// Single select: replace current selection
			inst.SelectOption(v.Value)
		} else {
			// Multi-select: toggle selection
			if v.IsSelected {
				inst.SelectOption(v.Value) // Add to selection
			} else {
				inst.DeselectOption(v.Value) // Remove from selection
			}
		}
		return true // Intent handled, stop bubbling
	}
	return false // Intent not handled, continue bubbling
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the optiongroup's ideal size given the constraints.
//
// Calculates total size based on children (options) + label + spacing.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate dimensions based on label
	var width, height int

	if inst.label != "" {
		height = 1
		width = len(inst.label)
	}

	// Sum up child option dimensions
	numChildren := len(inst.childInstances)
	if numChildren > 0 {
		var totalOptionWidth, totalOptionHeight int

		// Calculate max single option size (needed for horizontal layout)
		maxOptionWidth := 0
		maxOptionHeight := 0

		for _, child := range inst.childInstances {
			// Measure child with same constraints
			childSize := child.Measure(constraints)
			if childSize.Width > maxOptionWidth {
				maxOptionWidth = childSize.Width
			}
			if childSize.Height > maxOptionHeight {
				maxOptionHeight = childSize.Height
			}
		}

		// Calculate based on orientation
		if inst.orientation == OrientationVertical {
			// Vertical: Stack options, width = max, height = sum
			// Note: spacing is handled by layout system, not by Measure
			totalOptionWidth = maxOptionWidth
			totalOptionHeight = numChildren * maxOptionHeight
		} else {
			// Horizontal: Options side by side, width = sum, height = max
			// Note: spacing is handled by layout system, not by Measure
			totalOptionWidth = numChildren * maxOptionWidth
			totalOptionHeight = maxOptionHeight
		}

		// Update container size to fit options
		if totalOptionWidth > width {
			width = totalOptionWidth
		}
		height += totalOptionHeight
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	// Apply explicit style dimensions if set
	if inst.optionStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.optionStyle.Width)
	}
	if inst.optionStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.optionStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getSelectModeProp(props rtui.Props, def SelectMode) SelectMode {
	if v, ok := props[propMode]; ok {
		if m, ok := v.(SelectMode); ok {
			return m
		}
	}
	return def
}

func getOptionsProp(props rtui.Props, def []Option) []Option {
	if v, ok := props[propOptions]; ok {
		if opts, ok := v.([]Option); ok {
			return opts
		}
	}
	return def
}

func getStringsProp(props rtui.Props, key string, def []string) []string {
	if v, ok := props[key]; ok {
		if strs, ok := v.([]string); ok {
			return strs
		}
	}
	return def
}

func getOrientationProp(props rtui.Props, def Orientation) Orientation {
	if v, ok := props[propOrientation]; ok {
		if o, ok := v.(Orientation); ok {
			return o
		}
	}
	return def
}

// equalOptions compares two slices of Option for equality
func equalOptions(a, b []Option) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Value != b[i].Value || a[i].Label != b[i].Label {
			return false
		}
	}
	return true
}

// rebuildChildInstances rebuilds child instances from current options
func (inst *Instance) rebuildChildInstances() {
	// Clear parent references on old children
	for _, child := range inst.childInstances {
		child.parent = nil
	}

	// Rebuild child instances
	inst.childInstances = inst.childInstances[:0]
	for i, opt := range inst.options {
		childProps := rtui.Props{
			propKey:   inst.key + "-opt-" + opt.Value,
			"value": opt.Value,
			propLabel: opt.Label,
			"idx":   i,
			propMode:  inst.mode,
		}
		childInst := NewOptionInstance(childProps)
		inst.childInstances = append(inst.childInstances, childInst)
		// Set parent reference
		childInst.parent = inst
	}
}

// =============================================================================
// IntentHandler Implementation (Phase 3 - Infrastructure Ready)
// =============================================================================//

// Phase 3 Intent Bubble infrastructure is in place in runtime/intent/bubble.go.
// OptionGroup can implement intent.IntentHandlerProvider to handle intents from its children.
//
// To use intent bubble for OptionGroup in future:
// 1. Implement GetIntentHandler() method returning intent.IntentHandler
// 2. Add OptionIntentHandler struct wrapper similar to Phase 1's TreeNode pattern
// 3. The intent will automatically bubble from Option to OptionGroup via Parent() references
//
// For now, we maintain backward compatibility with direct parent references.
