package stack

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Stack components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Layout Props ===
	direction    Direction
	align        Align
	crossAlign   Align
	gap          int
	padding      [4]int
	stretchCross bool

	// === Sizing Props ===
	width  int
	height int
	flex   int

	// === Children ===
	children     []rtui.VNode
	childMeasure []layout.Size // cached child measurements

	// === Style ===
	instStyle style.Style

	// === Runtime State ===
	bounds      [4]int // x, y, w, h
	childBounds [][4]int
	dirty       bool
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

// NewInstance creates a new StackInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          getStringProp(props, "key", ""),
		direction:    getDirectionProp(props, Row),
		align:        getAlignProp(props, "align", AlignStart),
		crossAlign:   getAlignProp(props, "crossAlign", AlignStart),
		gap:          getIntProp(props, "gap", 0),
		padding:      getPaddingProp(props),
		stretchCross: getBoolProp(props, "stretchCross", false),
		width:        getIntProp(props, "width", 0),
		height:       getIntProp(props, "height", 0),
		flex:         getIntProp(props, "flex", 0),
		children:     getChildrenProp(props),
		instStyle:    getStyleProp(props),
		dirty:        true,
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
	inst.childBounds = nil
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldDirection := inst.direction
	oldGap := inst.gap
	oldChildren := inst.children

	inst.key = getStringProp(props, "key", inst.key)
	inst.direction = getDirectionProp(props, inst.direction)
	inst.align = getAlignProp(props, "align", inst.align)
	inst.crossAlign = getAlignProp(props, "crossAlign", inst.crossAlign)
	inst.gap = getIntProp(props, "gap", inst.gap)
	inst.padding = getPaddingPropWithDefault(props, inst.padding)
	inst.stretchCross = getBoolProp(props, "stretchCross", inst.stretchCross)
	inst.width = getIntProp(props, "width", inst.width)
	inst.height = getIntProp(props, "height", inst.height)
	inst.flex = getIntProp(props, "flex", inst.flex)
	inst.children = getChildrenPropWithDefault(props, inst.children)
	inst.instStyle = getStyleProp(props)

	changed := oldDirection != inst.direction ||
		oldGap != inst.gap ||
		len(oldChildren) != len(inst.children)

	if changed {
		inst.dirty = true
		inst.childMeasure = nil
		inst.childBounds = nil
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"direction":    inst.direction,
		"align":        inst.align,
		"crossAlign":   inst.crossAlign,
		"gap":          inst.gap,
		"padding":      inst.padding,
		"stretchCross": inst.stretchCross,
		"width":        inst.width,
		"height":       inst.height,
		"flex":         inst.flex,
		"children":     inst.children,
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

// GetContext implements ComponentInstance (no hooks for Stack).
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

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the stack's ideal size given the constraints.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	children := inst.children
	if len(children) == 0 {
		return layout.Size{
			Width:  constraints.ConstrainWidth(inst.width),
			Height: constraints.ConstrainHeight(inst.height),
		}
	}

	// Calculate padding
	paddingW := inst.padding[1] + inst.padding[3] // right + left
	paddingH := inst.padding[0] + inst.padding[2] // top + bottom

	// Inner constraints
	innerMaxW := constraints.MaxWidth - paddingW
	innerMaxH := constraints.MaxHeight - paddingH
	if innerMaxW < 0 {
		innerMaxW = 0
	}
	if innerMaxH < 0 {
		innerMaxH = 0
	}

	var totalW, totalH int

	// Initialize child measurements cache
	inst.childMeasure = make([]layout.Size, len(children))

	if inst.direction == Row {
		// HStack: measure total width and max height
		maxChildH := 0

		// First pass: identify flex children and measure non-flex
		type flexChild struct {
			index  int
			factor int
		}
		var flexChildren []flexChild
		fixedW := 0
		flexTotal := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, flexChild{
					index:  i,
					factor: childInfo.Flex,
				})
				flexTotal += childInfo.Flex
			} else {
				// Measure non-flex child
				// When Stack itself is auto-height (no explicit height), use MaxInt to allow children
				// to measure their natural height without constraint
				childMaxH := innerMaxH
				if inst.height == 0 && childMaxH == layout.MaxInt {
					// Pass MaxInt to allow unhindered measurement
					childMaxH = layout.MaxInt
				}

				cc := layout.Constraints{
					MinWidth:  0,
					MaxWidth:  layout.MaxInt,
					MinHeight: 0,
					MaxHeight: childMaxH,
				}
				size := inst.measureChild(child, cc)
				inst.childMeasure[i] = size
				fixedW += size.Width
				if size.Height > maxChildH {
					maxChildH = size.Height
				}
			}
		}

		totalW = fixedW

		// Check if width is bounded
		isBoundedWidth := constraints.MaxWidth < layout.MaxInt

		// Second pass: distribute space to flex children
		if len(flexChildren) > 0 && isBoundedWidth {
			availableW := constraints.MaxWidth - paddingW - (len(children)-1)*inst.gap
			remaining := availableW - fixedW
			if remaining < 0 {
				remaining = 0
			}

			baseFlexW := remaining / flexTotal
			remainder := remaining % flexTotal

			for _, fc := range flexChildren {
				flexW := baseFlexW * fc.factor
				if remainder > 0 {
					flexW++
					remainder--
				}
				if flexW < 0 {
					flexW = 0
				}

				cc := layout.Constraints{
					MinWidth:  flexW,
					MaxWidth:  flexW,
					MinHeight: 0,
					MaxHeight: innerMaxH,
				}
				size := inst.measureChild(children[fc.index], cc)
				inst.childMeasure[fc.index] = size
				totalW += flexW
				if size.Height > maxChildH {
					maxChildH = size.Height
				}
			}
		} else {
			// No bounded width or no flex: measure flex children naturally
			for _, fc := range flexChildren {
				cc := layout.Constraints{
					MinWidth:  0,
					MaxWidth:  layout.MaxInt,
					MinHeight: 0,
					MaxHeight: innerMaxH,
				}
				size := inst.measureChild(children[fc.index], cc)
				inst.childMeasure[fc.index] = size
				totalW += size.Width
				if size.Height > maxChildH {
					maxChildH = size.Height
				}
			}
		}

		// Add gaps
		totalW += (len(children) - 1) * inst.gap
		totalH = maxChildH
	} else {
		// VStack: measure max width and total height
		maxChildW := 0

		for i, child := range children {
			cc := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  innerMaxW,
				MinHeight: 0,
				MaxHeight: layout.MaxInt,
			}
			size := inst.measureChild(child, cc)
			inst.childMeasure[i] = size
			if size.Width > maxChildW {
				maxChildW = size.Width
			}
			totalH += size.Height
			if i < len(children)-1 {
				totalH += inst.gap
			}
		}

		totalW = maxChildW
	}

	// Add padding
	totalW += paddingW
	totalH += paddingH

	// Apply explicit dimensions
	if inst.width > 0 {
		totalW = inst.width
	}
	if inst.height > 0 {
		totalH = inst.height
	}

	// Apply constraints
	totalW = constraints.ConstrainWidth(totalW)
	totalH = constraints.ConstrainHeight(totalH)

	return layout.Size{Width: totalW, Height: totalH}
}

// measureChild measures a single child.
func (inst *Instance) measureChild(child rtui.VNode, constraints layout.Constraints) layout.Size {
	if child == nil {
		return layout.Size{}
	}

	// Debug log child measurement
	debugMeas := os.Getenv("MINT_DEBUG_STACK_MEASURE") == "true"
	var childTag string
	if child != nil {
		childTag = child.Tag()
	}
	if debugMeas {
		fmt.Printf("[Stack.measureChild] child=%s MaxW=%d MaxH=%d\n", childTag, constraints.MaxWidth, constraints.MaxHeight)
	}

	// Try InstanceFactory -> Measure
	if factory, ok := child.(rtui.InstanceFactory); ok {
		inst := factory.CreateInstance()
		if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			size := measurable.Measure(constraints)
			if debugMeas {
				fmt.Printf("[Stack.measureChild]   -> measured: %dx%d\n", size.Width, size.Height)
			}
			return size
		}
	}

	// Try direct Measurable
	if measurable, ok := child.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints)
		if debugMeas {
			fmt.Printf("[Stack.measureChild]   -> measured: %dx%d\n", size.Width, size.Height)
		}
		return size
	}

	// Fallback: estimate from content
	size := inst.estimateChildSize(child, constraints)
	if debugMeas {
		fmt.Printf("[Stack.measureChild]   -> estimated: %dx%d\n", size.Width, size.Height)
	}
	return size
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
			w = len([]rune(content))
		}
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
// Stack 是纯布局容器，布局由 LayoutBox 处理，子元素渲染由 PaintEngine 处理。
// 此方法仅处理容器自身的绘制（如背景色），目前 Stack 没有需要绘制的内容。
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// 纯布局容器没有自身需要绘制的内容
	// 布局由 LayoutBox (FlexLayout) 处理
	// 子元素渲染由 PaintEngine 使用 LayoutBox 坐标处理
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

func getDirectionProp(props rtui.Props, def Direction) Direction {
	if v, ok := props["direction"]; ok {
		if d, ok := v.(Direction); ok {
			return d
		}
	}
	return def
}

func getAlignProp(props rtui.Props, key string, def Align) Align {
	if v, ok := props[key]; ok {
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
