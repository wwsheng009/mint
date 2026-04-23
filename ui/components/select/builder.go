package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building Select VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Select builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
}

// Options sets the options list.
func (b *Builder) Options(opts []Option) *Builder {
	b.node.SetOptions(opts)
	return b
}

// AddOption adds a single option.
func (b *Builder) AddOption(value, label string) *Builder {
	b.node.AddOption(value, label)
	return b
}

// AddGroupedOption adds an option under an option group label.
func (b *Builder) AddGroupedOption(group, value, label string) *Builder {
	b.node.AddGroupedOption(group, value, label)
	return b
}

// OptionGroups flattens grouped options into the select option list.
func (b *Builder) OptionGroups(groups []OptionGroup) *Builder {
	b.node.SetOptionGroups(groups)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Selected sets the selected index.
func (b *Builder) Selected(idx int) *Builder {
	b.node.SetSelectedIndex(idx)
	return b
}

// SelectedIndices sets the selected indices for multi-select mode.
func (b *Builder) SelectedIndices(indices ...int) *Builder {
	b.node.SetSelectedIndices(indices)
	return b
}

// SelectionMode sets the select selection mode.
func (b *Builder) SelectionMode(mode SelectionMode) *Builder {
	b.node.SetSelectionMode(mode)
	return b
}

// SingleSelect enables single-select behavior.
func (b *Builder) SingleSelect() *Builder {
	return b.SelectionMode(SelectionSingle)
}

// MultiSelect enables multi-select behavior.
func (b *Builder) MultiSelect() *Builder {
	return b.SelectionMode(SelectionMultiple)
}

// TagsMode enables tag creation + multi-select behavior.
func (b *Builder) TagsMode(enabled bool) *Builder {
	if enabled {
		b.node.SetSelectionMode(SelectionTags)
		return b
	}
	if b.node.SelectionMode() == SelectionTags {
		b.node.SetSelectionMode(SelectionMultiple)
	}
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// Width sets the explicit width.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// OverlayPopup enables portal-based popup rendering.
func (b *Builder) OverlayPopup(enabled bool) *Builder {
	b.node.SetOverlayPopup(enabled)
	return b
}

// PortalRoot sets the popup portal root.
func (b *Builder) PortalRoot(root string) *Builder {
	b.node.SetPopupPortalRoot(root)
	return b
}

// CloseOnOutside controls outside click dismissal for overlay popups.
func (b *Builder) CloseOnOutside(close bool) *Builder {
	b.node.SetCloseOnOutside(close)
	return b
}

// Placeholder sets the text shown when nothing is selected.
func (b *Builder) Placeholder(text string) *Builder {
	b.node.SetPlaceholder(text)
	return b
}

// FilterOption enables search filtering inside the popup.
func (b *Builder) FilterOption(enabled bool) *Builder {
	b.node.SetFilterOption(enabled)
	return b
}

// FilterPlaceholder sets the filter input hint text.
func (b *Builder) FilterPlaceholder(text string) *Builder {
	b.node.SetFilterPlaceholder(text)
	return b
}

// MaxVisibleRows sets the number of visible popup rows.
func (b *Builder) MaxVisibleRows(rows int) *Builder {
	b.node.SetMaxVisibleRows(rows)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// OnChange sets the change intent.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
	return b
}

// ForField binds the select to a state field using FieldBinding.
// When the selected option changes, it emits a FieldChangeIntent.
// Single-select emits the selected index, multi-select emits comma-separated indices.
// Example:
//
//	var country = intent.StateKey[int]("country")
//	selectcomp.NewBuilder().
//	    Options([]selectcomp.Option{
//	        {Value: "us", Label: "United States"},
//	        {Value: "cn", Label: "China"},
//	    }).
//	    ForField(intent.ForField(country)).
//	    Build()
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	// Set the FieldIntent as the changeIntentField for MVP mode
	b.node.SetProps(rtui.Props{
		"changeIntent": binding, // binding implements both FieldIntent and Intent
	})
	return b
}

// ForForm binds the select to a form using FormBinding (Phase 6).
// When the selected option changes, it emits a FormFieldChangeIntent.
// Example:
//
//	formBinding := form.ForForm("myForm")
//	selectcomp.NewBuilder().
//	    Options([]selectcomp.Option{
//	        {Value: "us", Label: "United States"},
//	        {Value: "cn", Label: "China"},
//	    }).
//	    ForField(intent.ForField(country)).
//	    ForForm(formBinding).
//	    Build()
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
	return b
}

// Build returns the VNode.
func (b *Builder) Build() rtui.VNode {
	if b.node != nil && b.node.overlayPopup {
		return newOverlayComponent(b.node)
	}
	return b.node
}

// BuildTyped returns the typed VNode.
func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Select creates a new Select VNode (for backward compatibility).
// This matches the old form.Select() API.
func Select() *VNode {
	return New()
}

// NewSelect creates a new Select VNode (alias for New, for backward compatibility).
// This matches the old form.NewSelect() API.
func NewSelect() *VNode {
	return New()
}

// SelectBuilder is an alias for Builder (for backward compatibility).
// This matches the old form.SelectBuilder type.
type SelectBuilder = Builder

// SelectOption is an alias for Option (for backward compatibility).
// This matches the old form.SelectOption type.
type SelectOption = Option
