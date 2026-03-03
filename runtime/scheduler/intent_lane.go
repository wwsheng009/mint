// Package scheduler provides IntentWithLane for priority-aware intent dispatching.
package scheduler

import "github.com/wwsheng009/mint/runtime/intent"

// =============================================================================
// IntentWithLane
// =============================================================================

// IntentWithLane wraps an Intent with a priority lane.
// This allows the dispatcher to schedule intents based on priority.
//
// Example:
//
//	// High-priority user input
//	dispatcher.Dispatch(scheduler.WithLane(intent, scheduler.InputLane))
//
//	// Low-priority background update
//	dispatcher.Dispatch(scheduler.WithLane(intent, scheduler.IdleLane))
type IntentWithLane struct {
	intent.Intent
	Lane Lane
}

// WithLane wraps an intent with a specific lane.
//
// Example:
//
//	wrapped := scheduler.WithLane(myIntent, scheduler.InputLane)
func WithLane(i intent.Intent, lane Lane) IntentWithLane {
	return IntentWithLane{
		Intent: i,
		Lane:   lane,
	}
}

// GetIntent returns the underlying intent.
func (i IntentWithLane) GetIntent() intent.Intent {
	return i.Intent
}

// GetLane returns the lane priority.
func (i IntentWithLane) GetLane() Lane {
	return i.Lane
}

// =============================================================================
// Priority Helpers
// =============================================================================

// HighPriority wraps an intent as high priority (InputLane).
func HighPriority(i intent.Intent) IntentWithLane {
	return WithLane(i, InputLane)
}

// NormalPriority wraps an intent as normal priority (DefaultLane).
func NormalPriority(i intent.Intent) IntentWithLane {
	return WithLane(i, DefaultLane)
}

// LowPriority wraps an intent as low priority (TransitionLane).
func LowPriority(i intent.Intent) IntentWithLane {
	return WithLane(i, TransitionLane)
}

// BackgroundPriority wraps an intent as background priority (IdleLane).
func BackgroundPriority(i intent.Intent) IntentWithLane {
	return WithLane(i, IdleLane)
}

// SyncPriority wraps an intent as synchronous (SyncLane).
func SyncPriority(i intent.Intent) IntentWithLane {
	return WithLane(i, SyncLane)
}

// =============================================================================
// Intent Lane Inference
// =============================================================================

// InferLane attempts to infer the appropriate lane for an intent.
// This provides sensible defaults for common intent types.
func InferLane(i intent.Intent) Lane {
	typeName := i.IntentType()

	switch typeName {
	case "FieldChange", "Submit", "Click", "KeyPress":
		return InputLane
	case "FetchData", "FetchComplete":
		return DefaultLane
	case "Navigate", "Transition":
		return TransitionLane
	case "Analytics", "Log", "Prefetch":
		return IdleLane
	default:
		return DefaultLane
	}
}

// AutoWrap wraps an intent with an inferred lane.
func AutoWrap(i intent.Intent) IntentWithLane {
	return WithLane(i, InferLane(i))
}

// =============================================================================
// Batch Intent Lane
// =============================================================================

// IntentBatch represents multiple intents with a shared lane.
type IntentBatch struct {
	Intents []intent.Intent
	Lane    Lane
}

// NewIntentBatch creates a new intent batch.
func NewIntentBatch(lane Lane, intents ...intent.Intent) IntentBatch {
	return IntentBatch{
		Intents: intents,
		Lane:    lane,
	}
}

// Add adds an intent to the batch.
func (b *IntentBatch) Add(i intent.Intent) {
	b.Intents = append(b.Intents, i)
}

// WithLane wraps all intents in the batch with the batch's lane.
func (b IntentBatch) WithLane() []IntentWithLane {
	result := make([]IntentWithLane, len(b.Intents))
	for i, intent := range b.Intents {
		result[i] = WithLane(intent, b.Lane)
	}
	return result
}
