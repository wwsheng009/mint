package intent

// componentIDRouting provides unified routing logic for components with ComponentID
// This addresses P1-4 from INTENT_BUBBLE_AUDIT_REPORT.md
//
// Usage: Components can embed this or use the helper functions to check if an Intent
// matches their componentID

// GetComponentID is the interface for intents that have a ComponentID field
type GetComponentID interface {
	GetComponentID() string
}

// ShouldHandleIntentWithID checks if a component should handle the given intent based on ComponentID
//
// Rules:
// - If component.componentID is empty, handle ALL intents (return true)
// - If intent.ComponentID is empty, handle the intent (backward compatibility)
// - If both componentID and intent.ComponentID are set, they must match to handle
//
// This provides a unified routing mechanism for all components to use
func ShouldHandleIntentWithID(componentID string, i Intent) bool {
	// Case 1: Component doesn't have componentID set -> handle all intents
	if componentID == "" {
		return true
	}

	// Case 2: Intent doesn't have componentID -> backward compatibility, handle it
	if id, ok := i.(GetComponentID); ok {
		intentID := id.GetComponentID()
		if intentID == "" {
			return true // Intent without ComponentID - handle for backward compatibility
		}

		// Case 3: Both are set - must match to handle
		if intentID == componentID {
			return true // Matches
		}

		// Case 4: Both are set but don't match
		return false // Not for this component
	}

	// Case 5: Intent doesn't implement GetComponentID -> handle it
	// (This is for backward compatibility with existing intents)
	return true
}

// GetComponentIDFromIntent extracts the ComponentID from an intent if available
// Returns empty string if the intent doesn't have ComponentID
func GetComponentIDFromIntent(i Intent) string {
	if id, ok := i.(GetComponentID); ok {
		return id.GetComponentID()
	}
	return ""
}
