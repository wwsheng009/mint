// Package panel provides Fiber-first Panel container component.
// This file contains the constraint tracing support for Panel.
package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
)

// =============================================================================
// PanelInstance - Wraps Stack Instance with constraint tracing
// =============================================================================

// PanelInstance is a runtime wrapper that delegates to the Stack Instance
// (which now has native border properties) and adds constraint tracing at the Panel level.
type PanelInstance struct {
	// Panel identification
	key string
	path string // 用于约束追踪的路径

	// Panel props
	panelVNode *VNode

	// Delegated Stack instance (which has native border properties)
	stackInstance *newstack.Instance

	// Context for component instance
	context *rtui.ComponentContext

	// Props
	props rtui.Props

	// Dirty flag
	dirty bool
}

// Ensure PanelInstance implements required interfaces
var (
	_ rtui.ComponentInstance = (*PanelInstance)(nil)
	_ rtui.PaintableInstance = (*PanelInstance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// newPanelInstance creates a new PanelInstance from a Panel VNode.
func newPanelInstance(vnode *VNode, path string) *PanelInstance {
	// Build the composed Stack (with native border properties)
	composed := vnode.getComposed()
	if composed == nil {
		return nil
	}

	// Create Stack instance
	var stackInst *newstack.Instance
	if factory, ok := composed.(rtui.InstanceFactory); ok {
		inst := factory.CreateInstance()
		if si, ok := inst.(*newstack.Instance); ok {
			stackInst = si
		}
	}

	if stackInst == nil {
		return nil
	}

	// Get border label (generate from title if not explicitly set)
	borderLabel := vnode.borderLabel
	if borderLabel == "" && vnode.title != "" {
		borderLabel = " " + vnode.title + " "
	}

	// Build props for tracking - use generated borderLabel
	props := rtui.Props{
		"key":         vnode.key,
		"title":       vnode.title,
		"width":       vnode.width,
		"height":      vnode.height,
		"flex":        vnode.flex,
		"padding":     vnode.padding,
		"borderStyle": vnode.borderStyle,
		"borderColor": vnode.borderColor,
		"borderLabel": borderLabel,
	}

	return &PanelInstance{
		key:           vnode.key,
		path:          path,
		panelVNode:    vnode,
		stackInstance: stackInst,
		context:       rtui.NewComponentContext(vnode.key),
		props:         props,
		dirty:         true,
	}
}

// =============================================================================
// ComponentInstance Interface - Delegates to Border Instance
// =============================================================================

// Key implements ComponentInstance.
func (inst *PanelInstance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *PanelInstance) SetKey(key string) {
	inst.key = key
	inst.stackInstance.SetKey(key)
	inst.context = rtui.NewComponentContext(key)
}

// Init implements ComponentInstance.
func (inst *PanelInstance) Init(props rtui.Props) {
	inst.stackInstance.Init(props)
	inst.props = props
	inst.dirty = true
}

// Destroy implements ComponentInstance.
func (inst *PanelInstance) Destroy() {
	inst.stackInstance.Destroy()
}

// OnMount implements ComponentInstance.
func (inst *PanelInstance) OnMount() {
	inst.stackInstance.OnMount()
}

// OnUnmount implements ComponentInstance.
func (inst *PanelInstance) OnUnmount() {
	inst.stackInstance.OnUnmount()
}

// SetProps implements ComponentInstance.
func (inst *PanelInstance) SetProps(props rtui.Props) bool {
	changed := inst.stackInstance.SetProps(props)
	// Update local props
	inst.props = props
	inst.dirty = true
	return changed
}

// GetProps implements ComponentInstance.
func (inst *PanelInstance) GetProps() rtui.Props {
	// Return local props instead of delegating
	if inst.props == nil {
		inst.props = make(rtui.Props)
	}
	return inst.props
}

// MarkDirty implements ComponentInstance.
func (inst *PanelInstance) MarkDirty() {
	inst.dirty = true
	inst.stackInstance.MarkDirty()
}

// IsDirty implements ComponentInstance.
func (inst *PanelInstance) IsDirty() bool {
	return inst.dirty || inst.stackInstance.IsDirty()
}

// GetContext implements ComponentInstance.
func (inst *PanelInstance) GetContext() *rtui.ComponentContext {
	return inst.context
}

// =============================================================================
// Custom Measure Method - Adds Panel-level constraint tracing
// =============================================================================

// Measure implements layout measurement with constraint tracing.
func (inst *PanelInstance) Measure(constraints layout.Constraints) layout.Size {
	// Delegate measurement to Stack instance (which handles native border internally)
	// No need to calculate border padding here - Stack's native border properties handle this
	size := inst.stackInstance.Measure(constraints)

	return size
}

// =============================================================================
// PaintableInstance Interface - Delegates to Stack Instance
// =============================================================================

// Paint implements PaintableInstance.
func (inst *PanelInstance) Paint(x, y int) []paint.DrawCmd {
	return inst.stackInstance.Paint(x, y)
}

// =============================================================================
// paint.BorderInfo Interface - Direct implementation
// =============================================================================

// GetBorderStyle returns the border style.
func (inst *PanelInstance) GetBorderStyle() paint.BorderStyle {
	return paint.BorderStyle(inst.panelVNode.borderStyle)
}

// GetBorderColor returns the border color.
func (inst *PanelInstance) GetBorderColor() string {
	return string(inst.panelVNode.borderColor)
}

// GetBorderLabel returns the border label.
func (inst *PanelInstance) GetBorderLabel() string {
	// Return label from Panel props
	borderLabel := inst.panelVNode.borderLabel
	if borderLabel == "" && inst.panelVNode.title != "" {
		borderLabel = " " + inst.panelVNode.title + " "
	}
	return borderLabel
}

// =============================================================================
// Helper Methods
// =============================================================================

// GetStackInstance returns the delegated stack instance.
func (inst *PanelInstance) GetStackInstance() *newstack.Instance {
	return inst.stackInstance
}

// GetBorderInstance returns the delegated border instance (deprecated, use GetStackInstance).
// This method is kept for backward compatibility but now returns the Stack instance.
func (inst *PanelInstance) GetBorderInstance() *newstack.Instance {
	return inst.stackInstance
}

// SetPath sets the constraint tracing path for this instance.
func (inst *PanelInstance) SetPath(path string) *PanelInstance {
	inst.path = path
	return inst
}

// ClearDirty clears the dirty flag.
func (inst *PanelInstance) ClearDirty() {
	inst.dirty = false
	inst.stackInstance.ClearDirty()
}