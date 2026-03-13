package table

import "github.com/wwsheng009/mint/ui/components/internal/selection"

// SelectionMode defines how items can be selected in the table.
type SelectionMode = selection.SelectionMode

const (
	SelectionNone     = selection.SelectionNone
	SelectionSingle   = selection.SelectionSingle
	SelectionMultiple = selection.SelectionMultiple
)
