package intent

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/priority"
)

// =============================================================================
// Instance Tree Intent Bubble System (Phase 3)
// =============================================================================
// This is a local, Instance Tree-based intent bubble system.
// It works side-by-side with the global intent system (Registry/Dispatcher).
//
// Use this for:
//   - Parent-child communication via intent bubble
//   - Local event handling (e.g., Option -> OptionGroup)
//   - Declarative component interactions
//
// For global/state management, use the global intent system (Runtime/Emit).
// =============================================================================//

// TreeComponent is the interface for components in the instance tree.
// This interface only requires a Parent() method for bubbling.
// Parent() returns interface{} to avoid import cycles between ui and intent packages.
type TreeComponent interface {
	// Parent returns the parent component, or nil if this is root.
	Parent() interface{}
}

// IntentHandler is the interface for components that can handle intents.
// This is for local, instance-tree-based intent handling.
type IntentHandler interface {
	// HandleIntent handles an intent.
	// Returns true if the intent was handled (stop bubbling).
	// Returns false to continue bubbling up the tree.
	HandleIntent(i Intent) bool
}

// IntentHandlerProvider is an interface for components that provide an IntentHandler.
// This allows components to optionally provide intent handling without directly implementing the interface.
type IntentHandlerProvider interface {
	// GetIntentHandler returns the intent handler for this instance.
	GetIntentHandler() IntentHandler
}

// IntentEmitter is the interface for components that can emit bubble intents.
// This is typically implemented by components that want to send intents up the tree.
type IntentEmitter interface {
	// EmitIntent emits an intent that will bubble up the instance tree.
	EmitIntent(i Intent)
}

// IntentEmitterFunc is a function that emits intents via bubble.
// Components can store this as a field to emit intents.
type IntentEmitterFunc func(i Intent)

// NewIntentEmitter creates an emitter function that bubbles intents.
func NewIntentEmitter(emitFunc func(i Intent)) IntentEmitterFunc {
	return emitFunc
}

// Emit performs intent bubble on the instance tree.
// The component parameter is expected to implement:
//   - Parent() method returning the parent component
//   - Either IntentHandler or IntentHandlerProvider for handling
//
// This implements a local "event bubble" pattern similar to DOM event propagation.
func Emit(component TreeComponent, i Intent) {
	if component == nil {
		return
	}

	if i == nil {
		return
	}

	// Track depth to prevent infinite loops
	depth := 0
	maxDepth := 100

	// Start from the current component
	current := component

	// Bubble up through the instance tree
	for current != nil && depth < maxDepth {
		var handler IntentHandler = nil

		// Check if current component implements IntentHandler directly
		if h, ok := current.(IntentHandler); ok {
			handler = h
		} else if hProvider, ok := current.(IntentHandlerProvider); ok {
			// Check if component provides an IntentHandler via GetIntentHandler method
			handler = hProvider.GetIntentHandler()
		}

		if handler != nil {
			handled := handler.HandleIntent(i)
			if handled {
				// Intent was handled, stop bubbling
				return
			}
		}

		// Move to parent
		parent := current.Parent()
		if parent == nil {
			break
		}
		// Type assert the parent to TreeComponent to continue bubbling
		var ok bool
		current, ok = parent.(TreeComponent)
		if !ok {
			// Parent doesn't have Parent() method, can't bubble further
			break
		}

		depth++
	}

	// If we get here, the intent bubbled to the root without being handled
	// Optional: Forward to global intent system in future versions
}

// intentKey returns a string key for an intent.
// Some intents may have a Type() method (legacy), others have IntentType().
func intentKey(i Intent) string {
	// Try Type() first (for legacy compatibility)
	if t, ok := i.(interface{ Type() string }); ok {
		return t.Type()
	}
	// Fall back to IntentType()
	if i != nil {
		return i.IntentType()
	}
	return ""
}

// String returns a string representation of an intent for debugging.
func String(i Intent) string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Intent{Type: %s, Key: %s}", i.IntentType(), intentKey(i))
}

// IsTransition checks if an intent is a transition (async) intent.
func IsTransition(i Intent) bool {
	if i == nil {
		return false
	}
	if ti, ok := i.(interface{ IsTransition() bool }); ok {
		return ti.IsTransition()
	}
	return false
}

// GetPriority gets the priority of an intent.
func GetPriority(i Intent) ActionPriority {
	if i == nil {
		return PriorityNormal
	}
	if pa, ok := i.(interface{ Priority() ActionPriority }); ok {
		return pa.Priority()
	}
	return PriorityNormal
}

// ToLane converts intent priority to a priority.DirtyLevel.
func ToLane(i Intent) priority.DirtyLevel {
	return GetPriority(i).ToLane()
}
