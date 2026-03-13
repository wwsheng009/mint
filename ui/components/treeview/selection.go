package treeview

import "github.com/wwsheng009/mint/ui/components/internal/selection"

// SelectionMode defines how nodes can be selected in the treeview.
type SelectionMode = selection.SelectionMode

const (
	SelectionNone     = selection.SelectionNone
	SelectionSingle   = selection.SelectionSingle
	SelectionMultiple = selection.SelectionMultiple
)
