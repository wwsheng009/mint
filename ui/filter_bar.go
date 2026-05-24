package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/filterbar"
)

type FilterBarBuilder = filterbar.Builder
type FilterBarVNode = filterbar.VNode
type FilterBarField = filterbar.Field
type FilterBarFieldKind = filterbar.FieldKind
type FilterBarAction = filterbar.Action
type FilterBarOption = filterbar.Option

const (
	FilterBarFieldText   = filterbar.FieldText
	FilterBarFieldSearch = filterbar.FieldSearch
	FilterBarFieldSelect = filterbar.FieldSelect
	FilterBarFieldCustom = filterbar.FieldCustom
)

// NewFilterBarBuilder creates a FilterBar builder.
func NewFilterBarBuilder() *filterbar.Builder {
	return filterbar.NewBuilder()
}

// FilterBar creates a FilterBar from fields.
func FilterBar(fields []filterbar.Field) rtui.VNode {
	return filterbar.Of(fields)
}

// FilterBarSearch creates a search filter field.
func FilterBarSearch(key, label, value string) filterbar.Field {
	return filterbar.Search(key, label, value)
}

// FilterBarText creates a text filter field.
func FilterBarText(key, label, value string) filterbar.Field {
	return filterbar.Text(key, label, value)
}

// FilterBarSelect creates a select filter field.
func FilterBarSelect(key, label string, options []filterbar.Option) filterbar.Field {
	return filterbar.Select(key, label, options)
}

// FilterBarCustom creates a custom filter field.
func FilterBarCustom(key, label string, node rtui.VNode) filterbar.Field {
	return filterbar.Custom(key, label, node)
}

// FilterBarButton creates a command action for a FilterBar.
func FilterBarButton(key, label string, pressIntent intent.Intent) filterbar.Action {
	return filterbar.Button(key, label, pressIntent)
}
