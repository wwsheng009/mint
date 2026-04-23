package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propChangeIntent      = "changeIntent"
	propCloseOnOutside    = "closeOnOutside"
	propComponentID       = "componentID"
	propDisabled          = "disabled"
	propFilterOption      = "filterOption"
	propFilterPlaceholder = "filterPlaceholder"
	propFilterQuery       = "filterQuery"
	propFormID            = "formID"
	propHighlightedIndex  = "highlightedIndex"
	propKey               = "key"
	propMaxVisibleRows    = "maxVisibleRows"
	propOpen              = "open"
	propOptions           = "options"
	propOverlayPopup      = "overlayPopup"
	propOwnerID           = "ownerID"
	propPlaceholder       = "placeholder"
	propPortalRoot        = "portalRoot"
	propScrollOffset      = "scrollOffset"
	propSelectID          = "selectID"
	propSelectedIndex     = "selectedIndex"
	propSelectedIndices   = "selectedIndices"
	propSelectionMode     = "selectionMode"
	propStyle             = "style"
	propWidth             = "width"
)

// =============================================================================
// Option Type
// =============================================================================

// Option represents a single option in a select.
type Option struct {
	Value string
	Label string
	Group string
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
	options           []Option
	style             style.Style
	width             int
	placeholder       string
	filterOption      bool
	filterPlaceholder string
	maxVisibleRows    int
	overlayPopup      bool
	portalRoot        string
	closeOnOutside    bool

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent

	// === State Props (declarative, actual state managed by Instance) ===
	selectedIndex    int
	selectedIndices  []int
	selectionMode    SelectionMode
	disabled         bool
	formID           string // Form ID for Form integration (Phase 6)
	open             bool
	highlightedIndex int
	scrollOffset     int
	filterQuery      string
	selectID         string
	overlayCallbacks *overlayCallbacks

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
		ElementVNode:      rtui.NewElement("select"),
		options:           []Option{},
		selectedIndex:     -1,
		highlightedIndex:  -1,
		selectionMode:     SelectionSingle,
		maxVisibleRows:    6,
		placeholder:       "...",
		filterPlaceholder: "type to filter",
		overlayPopup:      false,
		portalRoot:        rtui.DefaultOverlayPortalRootID,
		closeOnOutside:    true,
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

// Children returns child nodes.
func (s *VNode) Children() []rtui.VNode {
	return nil
}

func (s *VNode) usesOverlayPopup() bool {
	return s.overlayPopup && s.ownerID() != ""
}

func (s *VNode) ownerID() string {
	return s.ID()
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
		propKey:               s.key,
		propComponentID:       s.componentID,
		propOptions:           s.options,
		propStyle:             s.style,
		propWidth:             s.width,
		propPlaceholder:       s.placeholder,
		propFilterOption:      s.filterOption,
		propFilterPlaceholder: s.filterPlaceholder,
		propFilterQuery:       s.filterQuery,
		propMaxVisibleRows:    s.maxVisibleRows,
		propOverlayPopup:      s.overlayPopup,
		popupPortalRootProp:   s.portalRoot,
		propCloseOnOutside:    s.closeOnOutside,
		propChangeIntent:      s.changeIntent,
		propSelectedIndex:     s.selectedIndex,
		propSelectedIndices:   s.resolvedSelectedIndices(),
		propSelectionMode:     s.selectionMode,
		propDisabled:          s.disabled,
		propFormID:            s.formID,
		propOpen:              s.open,
		propHighlightedIndex:  s.highlightedIndex,
		propScrollOffset:      s.scrollOffset,
		propSelectID:          s.selectID,
		overlayCallbacksProp:  s.overlayCallbacks,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		s.key = v
	}
	if v, ok := p[propComponentID].(string); ok {
		s.componentID = v
	}
	if v, ok := p[propOptions].([]Option); ok {
		s.options = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		s.style = v
	}
	if v, ok := p[propWidth].(int); ok {
		s.width = v
	}
	if v, ok := p[propPlaceholder].(string); ok {
		s.placeholder = v
	}
	if v, ok := p[propFilterOption].(bool); ok {
		s.filterOption = v
	}
	if v, ok := p[propFilterPlaceholder].(string); ok {
		s.filterPlaceholder = v
	}
	if v, ok := p[propFilterQuery].(string); ok {
		s.filterQuery = v
	}
	if v, ok := p[propMaxVisibleRows].(int); ok {
		s.maxVisibleRows = v
	}
	if v, ok := p[propOverlayPopup].(bool); ok {
		s.overlayPopup = v
	}
	if v, ok := p[popupPortalRootProp].(string); ok {
		s.portalRoot = v
	}
	if v, ok := p[propPortalRoot].(string); ok {
		s.portalRoot = v
	}
	if v, ok := p[propCloseOnOutside].(bool); ok {
		s.closeOnOutside = v
	}
	if v, ok := p[propChangeIntent].(intent.Intent); ok {
		s.changeIntent = v
	}
	if v, ok := p[propSelectedIndex].(int); ok {
		s.selectedIndex = v
	}
	if v, ok := p[propSelectedIndices].([]int); ok {
		s.selectedIndices = append([]int(nil), v...)
	}
	if v, ok := p[propSelectionMode].(SelectionMode); ok {
		s.selectionMode = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		s.disabled = v
	}
	if v, ok := p[propFormID].(string); ok {
		s.formID = v
	}
	if v, ok := p[propOpen].(bool); ok {
		s.open = v
	}
	if v, ok := p[propHighlightedIndex].(int); ok {
		s.highlightedIndex = v
	}
	if v, ok := p[propScrollOffset].(int); ok {
		s.scrollOffset = v
	}
	if v, ok := p[propSelectID].(string); ok {
		s.selectID = v
	}
	if v, ok := p[overlayCallbacksProp].(*overlayCallbacks); ok {
		s.overlayCallbacks = v
	}
	return s
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new SelectInstance from this VNode description.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		propKey:               s.key,
		propComponentID:       s.componentID,
		propOptions:           s.options,
		propStyle:             s.style,
		propWidth:             s.width,
		propPlaceholder:       s.placeholder,
		propFilterOption:      s.filterOption,
		propFilterPlaceholder: s.filterPlaceholder,
		propFilterQuery:       s.filterQuery,
		propMaxVisibleRows:    s.maxVisibleRows,
		propOverlayPopup:      s.usesOverlayPopup(),
		propOwnerID:           s.ownerID(),
		popupPortalRootProp:   s.portalRoot,
		propCloseOnOutside:    s.closeOnOutside,
		propChangeIntent:      s.changeIntent,
		propSelectedIndex:     s.selectedIndex,
		propSelectedIndices:   s.resolvedSelectedIndices(),
		propSelectionMode:     s.selectionMode,
		propDisabled:          s.disabled,
		propFormID:            s.formID,
		propOpen:              s.open,
		propHighlightedIndex:  s.highlightedIndex,
		propScrollOffset:      s.scrollOffset,
		propSelectID:          firstNonEmpty(s.selectID, s.ownerID()),
		overlayCallbacksProp:  s.overlayCallbacks,
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

// AddGroupedOption adds an option under an option group label.
func (s *VNode) AddGroupedOption(group, value, label string) *VNode {
	s.options = append(s.options, Option{Group: group, Value: value, Label: label})
	return s
}

// SetOptionGroups replaces options with a flattened set of grouped options.
func (s *VNode) SetOptionGroups(groups []OptionGroup) *VNode {
	s.options = flattenOptionGroups(groups)
	return s
}

// SetSelectedIndex sets the selected index.
func (s *VNode) SetSelectedIndex(idx int) *VNode {
	s.selectedIndex = idx
	if isMultiSelectionMode(s.selectionMode) {
		if idx >= 0 {
			s.selectedIndices = []int{idx}
		} else {
			s.selectedIndices = nil
		}
		return s
	}
	if idx >= 0 {
		s.selectedIndices = []int{idx}
	} else {
		s.selectedIndices = nil
	}
	return s
}

// SetSelectedIndices sets the selected indices for multi-select mode.
func (s *VNode) SetSelectedIndices(indices []int) *VNode {
	s.selectedIndices = append([]int(nil), indices...)
	if isMultiSelectionMode(s.selectionMode) {
		if len(s.selectedIndices) > 0 {
			s.selectedIndex = s.selectedIndices[len(s.selectedIndices)-1]
		} else {
			s.selectedIndex = -1
		}
		return s
	}
	if len(s.selectedIndices) > 0 {
		s.selectedIndex = s.selectedIndices[0]
		s.selectedIndices = []int{s.selectedIndex}
	} else {
		s.selectedIndex = -1
	}
	return s
}

// SetSelectionMode sets the selection mode.
func (s *VNode) SetSelectionMode(mode SelectionMode) *VNode {
	s.selectionMode = mode
	if isTagsSelectionMode(mode) {
		s.filterOption = true
	}
	switch mode {
	case SelectionMultiple, SelectionTags:
		if len(s.selectedIndices) == 0 && s.selectedIndex >= 0 {
			s.selectedIndices = []int{s.selectedIndex}
		}
	default:
		if s.selectedIndex >= 0 {
			s.selectedIndices = []int{s.selectedIndex}
		} else {
			s.selectedIndices = nil
		}
	}
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

// SetFilterOption enables search filtering in the popup.
func (s *VNode) SetFilterOption(enabled bool) *VNode {
	s.filterOption = enabled
	return s
}

// SetFilterPlaceholder configures the filter input hint text.
func (s *VNode) SetFilterPlaceholder(placeholder string) *VNode {
	s.filterPlaceholder = placeholder
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
	return s.resolvedSelectedIndices()
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

// FilterOption reports whether popup filtering is enabled.
func (s *VNode) FilterOption() bool {
	return s.filterOption
}

// FilterPlaceholder returns the filter input placeholder.
func (s *VNode) FilterPlaceholder() string {
	return s.filterPlaceholder
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

func (s *VNode) resolvedSelectedIndices() []int {
	if isMultiSelectionMode(s.selectionMode) {
		return append([]int(nil), s.selectedIndices...)
	}
	if s.selectedIndex >= 0 {
		return []int{s.selectedIndex}
	}
	return nil
}
