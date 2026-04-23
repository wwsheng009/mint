package list

import "github.com/wwsheng009/mint/ui/components/internal/selection"

// SelectionMode defines how items can be selected in the list.
type SelectionMode = selection.SelectionMode

const (
	SelectionNone     = selection.SelectionNone
	SelectionSingle   = selection.SelectionSingle
	SelectionMultiple = selection.SelectionMultiple
)
