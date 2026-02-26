// Package intent provides transition/intent support for async operations.
package intent

// =============================================================================
// Pending Intent
// =============================================================================

// ShowPendingIntent displays a pending/progress state for a transition.
// This is emitted to show the UI that an async operation is in progress.
//
// Example:
//
//	// Emit before async operation
//	ctx.Emit(ShowPendingIntent{
//	    Name: "loadUserData",
//	    Label: "Loading user data...",
//	})
type ShowPendingIntent struct {
	Name  string // Transition name (used to match with Complete)
	Label string // Human-readable label for display
}

func (ShowPendingIntent) IntentType() string {
	return "ShowPending"
}

// Priority - Pending states have medium priority
func (ShowPendingIntent) Priority() ActionPriority {
	return PriorityNormal
}

// =============================================================================
// Complete Intent
// =============================================================================

// CompleteTransitionIntent marks a transition as complete and provides the result.
// This is emitted after an async operation finishes to update the UI with the result.
//
// Example:
//
//	// Emit after async operation
//	ctx.Emit(CompleteTransitionIntent{
//	    Name: "loadUserData",
//	    Result: userData,
//	})
type CompleteTransitionIntent struct {
	Name   string       // Transition name (must match ShowPendingIntent)
	Result interface{}  // Operation result (any type)
	Error  error        // Optional error (if operation failed)
}

func (CompleteTransitionIntent) IntentType() string {
	return "CompleteTransition"
}

// Priority - Completion has normal priority
func (CompleteTransitionIntent) Priority() ActionPriority {
	return PriorityNormal
}

// =============================================================================
// Predefined Transition Names
// =============================================================================

// Common transition names for async operations.
const (
	TransitionLoadData     = "loadData"
	TransitionFetchAPI     = "fetchAPI"
	TransitionSearch       = "search"
	TransitionFilter       = "filter"
	TransitionSort         = "sort"
	TransitionValidate     = "validate"
	TransitionSubmit       = "submit"
	TransitionExport       = "export"
	TransitionImport       = "import"
)

// =============================================================================
// Builder Helpers
// =============================================================================

// NewShowPending creates a ShowPendingIntent for a given name.
func NewShowPending(name, label string) ShowPendingIntent {
	return ShowPendingIntent{
		Name:  name,
		Label: label,
	}
}

// NewCompleteTransition creates a CompleteTransitionIntent.
func NewCompleteTransition(name string, result interface{}) CompleteTransitionIntent {
	return CompleteTransitionIntent{
		Name:   name,
		Result: result,
	}
}

// NewCompleteTransitionWithError creates a CompleteTransitionIntent with an error.
func NewCompleteTransitionWithError(name string, err error) CompleteTransitionIntent {
	return CompleteTransitionIntent{
		Name:  name,
		Error: err,
	}
}
