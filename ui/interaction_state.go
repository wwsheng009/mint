// Package ui provides interaction state management for UI components.
package ui

import (
	state "github.com/wwsheng009/mint/internal/state"
)

// =============================================================================
// Type Aliases to internal/state
// =============================================================================
// These types are now defined in internal/state and re-exported here for compatibility.

// InteractionType is an alias to state.InteractionType
type InteractionType = state.InteractionType

// InteractionStateManager is an alias to state.InteractionStateManager
// Use state.InteractionStateManager directly for new code.
type InteractionStateManager = state.InteractionStateManager

// KeyValidator is an alias to state.KeyValidator
type KeyValidator = state.KeyValidator

// =============================================================================
// Re-exported Constants
// =============================================================================

const (
	InteractionHovered  = state.InteractionHovered
	InteractionFocused  = state.InteractionFocused
	InteractionPressed  = state.InteractionPressed
	InteractionSelected = state.InteractionSelected
)

// =============================================================================
// Re-exported Functions
// =============================================================================

// NewInteractionStateManager creates a new interaction state manager
func NewInteractionStateManager() *InteractionStateManager {
	return state.NewInteractionStateManager()
}

// NewKeyValidator creates a new key validator
func NewKeyValidator() *KeyValidator {
	return state.NewKeyValidator()
}
