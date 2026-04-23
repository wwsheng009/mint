// Package selection provides a shared SelectionMode type used by list, table, and treeview components.
package selection

// SelectionMode defines how items can be selected in a component.
type SelectionMode int

const (
	SelectionNone     SelectionMode = iota // No selection allowed
	SelectionSingle                        // Only one item can be selected at a time
	SelectionMultiple                      // Multiple items can be selected
)
