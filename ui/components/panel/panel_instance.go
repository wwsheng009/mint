// Package panel provides Fiber-first Panel container component.
// This file contains the constraint tracing support for Panel.
package panel

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
)

// =============================================================================
// PanelInstance - Wraps Border Instance with constraint tracing
// =============================================================================

// PanelInstance is a runtime wrapper that delegates to the Border Instance
// and adds constraint tracing at the Panel level.
type PanelInstance struct {
	// Panel identification
	key string
	path string // 用于约束追踪的路径

	// Panel props
	panelVNode *VNode

	// Delegated Border instance
	borderInstance *newborder.Instance

	// Border label for tracing
	borderLabel string

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
	// Build the composed Border
	composed := vnode.getComposed()
	if composed == nil {
		return nil
	}

	// Create Border instance
	var borderInst *newborder.Instance
	if factory, ok := composed.(rtui.InstanceFactory); ok {
		inst := factory.CreateInstance()
		if bi, ok := inst.(*newborder.Instance); ok {
			borderInst = bi
			// Set path for Border's internal tracing
			panelPath := path
			if vnode.key != "" {
				panelPath = fmt.Sprintf("%s/panel(%s)", path, vnode.key)
			} else {
				panelPath = fmt.Sprintf("%s/panel", path)
			}
			borderInst.SetPath(panelPath)
		}
	}

	if borderInst == nil {
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
		key:            vnode.key,
		path:           path,
		panelVNode:     vnode,
		borderInstance: borderInst,
		borderLabel:    borderLabel,
		context:        rtui.NewComponentContext(vnode.key),
		props:          props,
		dirty:          true,
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
	inst.borderInstance.SetKey(key)
	inst.context = rtui.NewComponentContext(key)
}

// Init implements ComponentInstance.
func (inst *PanelInstance) Init(props rtui.Props) {
	inst.borderInstance.Init(props)
	inst.props = props
	inst.dirty = true
}

// Destroy implements ComponentInstance.
func (inst *PanelInstance) Destroy() {
	inst.borderInstance.Destroy()
}

// OnMount implements ComponentInstance.
func (inst *PanelInstance) OnMount() {
	inst.borderInstance.OnMount()
}

// OnUnmount implements ComponentInstance.
func (inst *PanelInstance) OnUnmount() {
	inst.borderInstance.OnUnmount()
}

// SetProps implements ComponentInstance.
func (inst *PanelInstance) SetProps(props rtui.Props) bool {
	changed := inst.borderInstance.SetProps(props)
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
	inst.borderInstance.MarkDirty()
}

// IsDirty implements ComponentInstance.
func (inst *PanelInstance) IsDirty() bool {
	return inst.dirty || inst.borderInstance.IsDirty()
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
	// Build panel ID for tracing
	panelID := "panel"
	if inst.key != "" {
		panelID = fmt.Sprintf("panel(%s)", inst.key)
	}
	panelPath := inst.path

	// Build border ID for tracing
	borderID := "border"
	if inst.borderInstance.Key() != "" {
		borderID = fmt.Sprintf("border(%s)", inst.borderInstance.Key())
	}

	// Calculate inner constraints (accounting for border padding)
	borderWidth := newborder.GetBorderWidth(inst.panelVNode.borderStyle)
	borderPadding := borderWidth * 2

	var innerConstraints layout.Constraints

	// Use explicit dimensions if set
	if inst.panelVNode.width > 0 {
		innerWidth := inst.panelVNode.width - borderPadding
		if innerWidth > 0 {
			innerConstraints.MinWidth = innerWidth
			innerConstraints.MaxWidth = innerWidth
		}
	}
	if inst.panelVNode.height > 0 {
		innerHeight := inst.panelVNode.height - borderPadding
		if innerHeight > 0 {
			innerConstraints.MinHeight = innerHeight
			innerConstraints.MaxHeight = innerHeight
		}
	}

	// Use parent constraints for auto dimensions
	if innerConstraints.MinWidth == 0 && innerConstraints.MaxWidth == 0 {
		innerConstraints.MinWidth = max(0, constraints.MinWidth-borderPadding)
		innerConstraints.MaxWidth = max(0, constraints.MaxWidth-borderPadding)
	}
	if innerConstraints.MinHeight == 0 && innerConstraints.MaxHeight == 0 {
		innerConstraints.MinHeight = max(0, constraints.MinHeight-borderPadding)
		innerConstraints.MaxHeight = max(0, constraints.MaxHeight-borderPadding)
	}

	// Trace constraint propagation from Panel to Border
	borderPath := fmt.Sprintf("%s/%s", panelPath, "border")
	layout.TraceMeasuring(
		panelID,
		borderID,
		borderPath,
		constraints,        // Panel's input constraints
		innerConstraints,   // Constraints passed to Border (inner dimensions)
		layout.Size{},      // Size will be updated after measurement
		fmt.Sprintf("Panel: Applied border padding (%dx%d), explicit width=%d, height=%d, flex=%d",
			borderWidth, borderWidth, inst.panelVNode.width, inst.panelVNode.height, inst.panelVNode.flex),
	)

	// Delegate measurement to Border instance
	size := inst.borderInstance.Measure(constraints)

	return size
}

// =============================================================================
// PaintableInstance Interface - Delegates to Border Instance
// =============================================================================

// Paint implements PaintableInstance.
func (inst *PanelInstance) Paint(x, y int) []paint.DrawCmd {
	return inst.borderInstance.Paint(x, y)
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
	return inst.borderLabel
}

// =============================================================================
// Helper Methods
// =============================================================================

// GetBorderInstance returns the delegated border instance.
func (inst *PanelInstance) GetBorderInstance() *newborder.Instance {
	return inst.borderInstance
}

// SetPath sets the constraint tracing path for this instance.
func (inst *PanelInstance) SetPath(path string) *PanelInstance {
	inst.path = path
	return inst
}

// ClearDirty clears the dirty flag.
func (inst *PanelInstance) ClearDirty() {
	inst.dirty = false
	inst.borderInstance.ClearDirty()
}

// =============================================================================
// max helper function
// =============================================================================

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
