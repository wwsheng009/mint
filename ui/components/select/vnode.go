package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Option Type
// =============================================================================

// Option represents a single option in a select.
type Option struct {
	Value string
	Label string
}

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the select description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key         string
	componentID string // Component ID for Intent routing (Phase 10)

	// === Visual Props ===
	options        []Option
	style          style.Style
	width          int
	placeholder    string
	maxVisibleRows int
	overlayPopup   bool
	portalRoot     string
	closeOnOutside bool

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent

	// === State Props (declarative, actual state managed by Instance) ===
	selectedIndex   int
	selectedIndices []int
	selectionMode   SelectionMode
	disabled        bool
	formID          string // Form ID for Form integration (Phase 6)

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Select VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:   rtui.NewElement("select"),
		options:        []Option{},
		selectedIndex:  -1,
		selectionMode:  SelectionSingle,
		maxVisibleRows: 6,
		placeholder:    "...",
		overlayPopup:   false,
		portalRoot:     rtui.DefaultOverlayPortalRootID,
		closeOnOutside: true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (s *VNode) Key() string {
	return s.key
}

// SetKey sets the component key - returns VNode for chaining.
func (s *VNode) SetKey(key string) rtui.VNode {
	s.key = key
	return s
}

// ID returns the business identifier used for anchoring.
// Falls back to componentID/key so overlay popup can anchor without an explicit SetID call.
func (s *VNode) ID() string {
	if id := s.ElementVNode.ID(); id != "" {
		return id
	}
	if s.componentID != "" {
		return s.componentID
	}
	return s.key
}

// SetID sets the business identifier.
func (s *VNode) SetID(id string) rtui.VNode {
	s.ElementVNode.SetID(id)
	return s
}

// Tag returns the tag name.
func (s *VNode) Tag() string {
	return "select"
}

// Style returns the visual style.
func (s *VNode) Style() style.Style {
	return s.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (s *VNode) SetStyle(st style.Style) rtui.VNode {
	s.style = st
	return s
}

// Children returns the optional overlay popup child.
func (s *VNode) Children() []rtui.VNode {
	if !s.usesOverlayPopup() {
		return nil
	}
	return []rtui.VNode{newPopupVNode(s)}
}

func (s *VNode) usesOverlayPopup() bool {
	return s.overlayPopup && s.ownerID() != ""
}

func (s *VNode) ownerID() string {
	return s.ID()
}

func newPopupVNode(owner *VNode) rtui.VNode {
	ownerID := owner.ownerID()
	surface := &popupVNode{ElementVNode: rtui.NewElement("select-popup")}
	surface.SetKey(ownerID + "-popup")
	surface.SetID(ownerID + "-popup")
	surface.SetProps(rtui.Props{
		"ownerID":        ownerID,
		"closeOnOutside": owner.closeOnOutside,
	})

	portal := rtui.NewElement("box")
	portal.SetKey(ownerID + "-popup-portal")
	portal.SetID(ownerID + "-popup-portal")
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     0,
		"top":      0,
		"width":    1,
		"height":   1,
	})
	if owner.portalRoot != "" {
		portal.SetPortalRoot(owner.portalRoot)
	}
	portal.SetAnchorTo(ownerID, rttypes.AnchorBottomLeft)
	portal.SetPortalPosition(rttypes.PositionAbsolute)
	portal.SetProp("left", 0)
	portal.SetProp("top", 1)
	portal.SetChildren([]rtui.VNode{surface})
	return portal
}

// SetChildren is a no-op for select - returns VNode for chaining.
func (s *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return s
}

// GetLayer returns the rendering layer.
func (s *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (s *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return s
}

// Props returns the node properties.
func (s *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":             s.key,
		"componentID":     s.componentID,
		"options":         s.options,
		"style":           s.style,
		"width":           s.width,
		"placeholder":     s.placeholder,
		"maxVisibleRows":  s.maxVisibleRows,
		"overlayPopup":    s.overlayPopup,
		"portalRoot":      s.portalRoot,
		"closeOnOutside":  s.closeOnOutside,
		"changeIntent":    s.changeIntent,
		"selectedIndex":   s.selectedIndex,
		"selectedIndices": append([]int(nil), s.selectedIndices...),
		"selectionMode":   s.selectionMode,
		"disabled":        s.disabled,
		"formID":          s.formID,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		s.key = v
	}
	if v, ok := p["componentID"].(string); ok {
		s.componentID = v
	}
	if v, ok := p["options"].([]Option); ok {
		s.options = v
	}
	if v, ok := p["style"].(style.Style); ok {
		s.style = v
	}
	if v, ok := p["width"].(int); ok {
		s.width = v
	}
	if v, ok := p["placeholder"].(string); ok {
		s.placeholder = v
	}
	if v, ok := p["maxVisibleRows"].(int); ok {
		s.maxVisibleRows = v
	}
	if v, ok := p["overlayPopup"].(bool); ok {
		s.overlayPopup = v
	}
	if v, ok := p["portalRoot"].(string); ok {
		s.portalRoot = v
	}
	if v, ok := p["closeOnOutside"].(bool); ok {
		s.closeOnOutside = v
	}
	if v, ok := p["changeIntent"].(intent.Intent); ok {
		s.changeIntent = v
	}
	if v, ok := p["selectedIndex"].(int); ok {
		s.selectedIndex = v
	}
	if v, ok := p["selectedIndices"].([]int); ok {
		s.selectedIndices = append([]int(nil), v...)
	}
	if v, ok := p["selectionMode"].(SelectionMode); ok {
		s.selectionMode = v
	}
	if v, ok := p["disabled"].(bool); ok {
		s.disabled = v
	}
	if v, ok := p["formID"].(string); ok {
		s.formID = v
	}
	return s
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new SelectInstance from this VNode description.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":             s.key,
		"componentID":     s.componentID,
		"options":         s.options,
		"style":           s.style,
		"width":           s.width,
		"placeholder":     s.placeholder,
		"maxVisibleRows":  s.maxVisibleRows,
		"overlayPopup":    s.usesOverlayPopup(),
		"ownerID":         s.ownerID(),
		"portalRoot":      s.portalRoot,
		"closeOnOutside":  s.closeOnOutside,
		"changeIntent":    s.changeIntent,
		"selectedIndex":   s.selectedIndex,
		"selectedIndices": append([]int(nil), s.selectedIndices...),
		"selectionMode":   s.selectionMode,
		"disabled":        s.disabled,
		"formID":          s.formID,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetOptions sets the options list.
func (s *VNode) SetOptions(opts []Option) *VNode {
	s.options = opts
	return s
}

// AddOption adds a single option.
func (s *VNode) AddOption(value, label string) *VNode {
	s.options = append(s.options, Option{Value: value, Label: label})
	return s
}

// SetSelectedIndex sets the selected index.
func (s *VNode) SetSelectedIndex(idx int) *VNode {
	s.selectedIndex = idx
	return s
}

// SetSelectedIndices sets the selected indices for multi-select mode.
func (s *VNode) SetSelectedIndices(indices []int) *VNode {
	s.selectedIndices = append([]int(nil), indices...)
	return s
}

// SetSelectionMode sets the selection mode.
func (s *VNode) SetSelectionMode(mode SelectionMode) *VNode {
	s.selectionMode = mode
	return s
}

// SetDisabled sets the disabled state.
func (s *VNode) SetDisabled(disabled bool) *VNode {
	s.disabled = disabled
	return s
}

// SetComponentID sets the component ID for Intent routing.
func (s *VNode) SetComponentID(componentID string) *VNode {
	s.componentID = componentID
	return s
}

// SetWidth sets the explicit width.
func (s *VNode) SetWidth(width int) *VNode {
	s.width = width
	return s
}

// SetPlaceholder sets the text shown when nothing is selected.
func (s *VNode) SetPlaceholder(placeholder string) *VNode {
	s.placeholder = placeholder
	return s
}

// SetMaxVisibleRows sets the number of visible rows in the dropdown popup.
func (s *VNode) SetMaxVisibleRows(rows int) *VNode {
	s.maxVisibleRows = rows
	return s
}

// SetOverlayPopup enables/disables overlay portal popup rendering.
func (s *VNode) SetOverlayPopup(enabled bool) *VNode {
	s.overlayPopup = enabled
	return s
}

// SetPopupPortalRoot sets the portal root used by the overlay popup.
func (s *VNode) SetPopupPortalRoot(root string) *VNode {
	s.portalRoot = root
	return s
}

// SetCloseOnOutside controls outside click dismissal for overlay popup.
func (s *VNode) SetCloseOnOutside(close bool) *VNode {
	s.closeOnOutside = close
	return s
}

// SetChangeIntent sets the change intent.
func (s *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	s.changeIntent = changeIntent
	return s
}

// SetStyleProps sets the visual style.
func (s *VNode) SetStyleProps(st style.Style) *VNode {
	s.style = st
	return s
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Options returns the options list.
func (s *VNode) Options() []Option {
	return s.options
}

// SelectedIndex returns the selected index.
func (s *VNode) SelectedIndex() int {
	return s.selectedIndex
}

// SelectedIndices returns the selected indices for multi-select mode.
func (s *VNode) SelectedIndices() []int {
	return append([]int(nil), s.selectedIndices...)
}

// SelectionMode returns the current selection mode.
func (s *VNode) SelectionMode() SelectionMode {
	return s.selectionMode
}

// Disabled returns the disabled state.
func (s *VNode) Disabled() bool {
	return s.disabled
}

// Width returns the explicit width.
func (s *VNode) Width() int {
	return s.width
}

// Placeholder returns the empty-state label.
func (s *VNode) Placeholder() string {
	return s.placeholder
}

// MaxVisibleRows returns the number of visible popup rows.
func (s *VNode) MaxVisibleRows() int {
	return s.maxVisibleRows
}

// OverlayPopup reports whether overlay popup mode is enabled.
func (s *VNode) OverlayPopup() bool {
	return s.overlayPopup
}

// PortalRoot returns the popup portal root.
func (s *VNode) PortalRoot() string {
	return s.portalRoot
}

// CloseOnOutside reports whether outside clicks dismiss the popup.
func (s *VNode) CloseOnOutside() bool {
	return s.closeOnOutside
}

// ChangeIntent returns the change intent.
func (s *VNode) ChangeIntent() intent.Intent {
	return s.changeIntent
}

// GetComponentID returns the component ID.
func (s *VNode) GetComponentID() string {
	return s.componentID
}

// SetFormID sets the form ID for Form integration (Phase 6).
func (s *VNode) SetFormID(formID string) *VNode {
	s.formID = formID
	return s
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Select VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Select uses BoxModelMixin for padding/margin, and has no border.
func (s *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   s.BoxModelMixin.Padding()[3],
			Right:  s.BoxModelMixin.Padding()[1],
			Top:    s.BoxModelMixin.Padding()[0],
			Bottom: s.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   s.BoxModelMixin.Margin()[3],
			Right:  s.BoxModelMixin.Margin()[1],
			Top:    s.BoxModelMixin.Margin()[0],
			Bottom: s.BoxModelMixin.Margin()[2],
		},
		// Select typically doesn't have a border
		Border: layout.Border{Style: layout.BorderNone},
	}
}
