package wrap

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Wrap components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Layout Props ===
	gap        int
	rowGap     int
	align      Align
	width      int
	padding    [4]int
	fillWidth  bool
	fillHeight bool

	// === Children ===
	children     []rtui.VNode
	childMeasure []layout.Size // cached child measurements

	// === Calculated Layout ===
	rows         [][]int       // each row contains child indices
	rowHeights   []int         // height of each row
	childBounds  [][4]int      // bounds for each child

	// === Style ===
	instStyle style.Style

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new WrapInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:       getStringProp(props, "key", ""),
		gap:       getIntProp(props, "gap", 1),
		rowGap:    getIntProp(props, "rowGap", 0),
		align:     getAlignProp(props, AlignStart),
		width:     getIntProp(props, "width", 80),
		padding:   getPaddingProp(props),
		fillWidth: getBoolProp(props, "fillWidth", false),
		fillHeight: getBoolProp(props, "fillHeight", false),
		children:  getChildrenProp(props),
		instStyle: getStyleProp(props),
		dirty:     true,
	}

	return inst
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
	inst.children = nil
	inst.childMeasure = nil
	inst.rows = nil
	inst.rowHeights = nil
	inst.childBounds = nil
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldGap := inst.gap
	oldWidth := inst.width
	oldChildren := inst.children

	inst.key = getStringProp(props, "key", inst.key)
	inst.gap = getIntProp(props, "gap", inst.gap)
	inst.rowGap = getIntProp(props, "rowGap", inst.rowGap)
	inst.align = getAlignPropWithDefault(props, inst.align)
	inst.width = getIntProp(props, "width", inst.width)
	inst.padding = getPaddingPropWithDefault(props, inst.padding)
	inst.fillWidth = getBoolProp(props, "fillWidth", inst.fillWidth)
	inst.fillHeight = getBoolProp(props, "fillHeight", inst.fillHeight)
	inst.children = getChildrenPropWithDefault(props, inst.children)
	inst.instStyle = getStyleProp(props)

	changed := oldGap != inst.gap ||
		oldWidth != inst.width ||
		len(oldChildren) != len(inst.children)

	if changed {
		inst.dirty = true
		inst.childMeasure = nil
		inst.rows = nil
		inst.rowHeights = nil
		inst.childBounds = nil
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":        inst.key,
		"gap":        inst.gap,
		"rowGap":     inst.rowGap,
		"align":      inst.align,
		"width":      inst.width,
		"padding":    inst.padding,
		"fillWidth":  inst.fillWidth,
		"fillHeight": inst.fillHeight,
		"children":   inst.children,
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

// GetContext implements ComponentInstance (no hooks for Wrap).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// Bounds Management
// =============================================================================

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// SetChildBounds sets bounds for a specific child.
func (inst *Instance) SetChildBounds(index int, x, y, w, h int) {
	if index < 0 || index >= len(inst.children) {
		return
	}
	if inst.childBounds == nil {
		inst.childBounds = make([][4]int, len(inst.children))
	}
	inst.childBounds[index] = [4]int{x, y, w, h}
}

// GetChildBounds returns bounds for a specific child.
func (inst *Instance) GetChildBounds(index int) (x, y, w, h int) {
	if index < 0 || index >= len(inst.childBounds) {
		return 0, 0, 0, 0
	}
	b := inst.childBounds[index]
	return b[0], b[1], b[2], b[3]
}

// GetRows returns the calculated rows (child indices per row).
func (inst *Instance) GetRows() [][]int {
	return inst.rows
}

// GetRowHeights returns the height of each row.
func (inst *Instance) GetRowHeights() []int {
	return inst.rowHeights
}

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the wrap's ideal size given the constraints.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	children := inst.children
	if len(children) == 0 {
		return layout.Size{
			Width:  constraints.ConstrainWidth(inst.padding[1] + inst.padding[3]),
			Height: constraints.ConstrainHeight(inst.padding[0] + inst.padding[2]),
		}
	}

	// Calculate padding
	paddingW := inst.padding[1] + inst.padding[3] // right + left
	paddingH := inst.padding[0] + inst.padding[2] // top + bottom

	// Available width for content
	availableWidth := inst.width - paddingW
	if availableWidth <= 0 {
		availableWidth = constraints.MaxWidth - paddingW
	}
	if availableWidth < 0 {
		availableWidth = 0
	}

	// Measure all children and estimate widths
	inst.childMeasure = make([]layout.Size, len(children))
	childWidths := make([]int, len(children))
	for i, child := range children {
		size := inst.measureChild(child, constraints)
		inst.childMeasure[i] = size
		childWidths[i] = size.Width
	}

	// Calculate rows
	inst.rows = inst.calculateRows(childWidths, availableWidth)

	// Calculate row heights
	inst.rowHeights = make([]int, len(inst.rows))
	for rowIdx, row := range inst.rows {
		maxH := 1
		for _, childIdx := range row {
			if inst.childMeasure[childIdx].Height > maxH {
				maxH = inst.childMeasure[childIdx].Height
			}
		}
		inst.rowHeights[rowIdx] = maxH
	}

	// Calculate total height
	totalH := paddingH
	rowGap := inst.rowGap
	if rowGap == 0 {
		rowGap = inst.gap
	}
	for i, h := range inst.rowHeights {
		totalH += h
		if i < len(inst.rowHeights)-1 {
			totalH += rowGap
		}
	}

	// Width is the container width
	totalW := inst.width

	// Apply constraints
	totalW = constraints.ConstrainWidth(totalW)
	totalH = constraints.ConstrainHeight(totalH)

	return layout.Size{Width: totalW, Height: totalH}
}

// calculateRows divides children into rows based on available width.
func (inst *Instance) calculateRows(childWidths []int, availableWidth int) [][]int {
	var rows [][]int
	currentRow := []int{}
	currentWidth := 0

	for i, childWidth := range childWidths {
		// Check if we need to wrap
		shouldWrap := len(currentRow) > 0 &&
			(currentWidth+childWidth+inst.gap > availableWidth)

		if shouldWrap {
			// Finish current row
			rows = append(rows, currentRow)
			// Start new row with this child
			currentRow = []int{i}
			currentWidth = childWidth
		} else {
			// Add to current row
			currentRow = append(currentRow, i)
			if len(currentRow) > 1 {
				currentWidth += inst.gap
			}
			currentWidth += childWidth
		}
	}

	// Don't forget the last row
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return rows
}

// measureChild measures a single child.
func (inst *Instance) measureChild(child rtui.VNode, constraints layout.Constraints) layout.Size {
	if child == nil {
		return layout.Size{}
	}

	// Try InstanceFactory -> Measure
	if factory, ok := child.(rtui.InstanceFactory); ok {
		cinst := factory.CreateInstance()
		if measurable, ok := cinst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			return measurable.Measure(constraints)
		}
	}

	// Try direct Measurable
	if measurable, ok := child.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(constraints)
	}

	// Fallback: estimate from content
	return inst.estimateChildSize(child, constraints)
}

// estimateChildSize estimates child size when Measure is not available.
func (inst *Instance) estimateChildSize(child rtui.VNode, constraints layout.Constraints) layout.Size {
	w := 10 // default width
	h := 1  // default height

	if props := child.Props(); props != nil {
		if pw := props.GetInt("width"); pw > 0 {
			w = pw
		}
		if ph := props.GetInt("height"); ph > 0 {
			h = ph
		}
		if content := props.GetString("content"); content != "" {
			w = utf8.RuneCountInString(content)
		}
		if label := props.GetString("label"); label != "" {
			// Button: label + brackets + focus indicator
			w = utf8.RuneCountInString(label) + 4
		}
	}

	// Also check for Measurable interface on VNode directly
	if measurable, ok := child.(runtime.Measurable); ok {
		size := measurable.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		})
		w = size.Width
		h = size.Height
	}

	return layout.Size{
		Width:  constraints.ConstrainWidth(w),
		Height: constraints.ConstrainHeight(h),
	}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
// Wrap is a pure layout container - layout is handled by layout engine.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Pure layout container has no content to paint
	return nil
}

// =============================================================================
// Style Management
// =============================================================================

// GetStyle returns the instance style.
func (inst *Instance) GetStyle() style.Style {
	return inst.instStyle
}

// SetStyle sets the instance style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.instStyle = s
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
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

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
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

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getAlignProp(props rtui.Props, def Align) Align {
	if v, ok := props["align"]; ok {
		if a, ok := v.(Align); ok {
			return a
		}
	}
	return def
}

func getAlignPropWithDefault(props rtui.Props, def Align) Align {
	if v, ok := props["align"]; ok {
		if a, ok := v.(Align); ok {
			return a
		}
	}
	return def
}

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{0, 0, 0, 0}
}

func getPaddingPropWithDefault(props rtui.Props, def [4]int) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return def
}

func getChildrenProp(props rtui.Props) []rtui.VNode {
	if v, ok := props["children"]; ok {
		if c, ok := v.([]rtui.VNode); ok {
			return c
		}
	}
	return nil
}

func getChildrenPropWithDefault(props rtui.Props, def []rtui.VNode) []rtui.VNode {
	if v, ok := props["children"]; ok {
		if c, ok := v.([]rtui.VNode); ok {
			return c
		}
	}
	return def
}
