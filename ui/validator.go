package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Hook Validator (re-exported from runtime/types)
// =============================================================================

// HookValidator validates hook call order and consistency
type HookValidator = rtui.HookValidator

// HookOrderError represents a hook order violation
type HookOrderError = rtui.HookOrderError

// NewHookValidator creates a new hook validator
func NewHookValidator(componentID string) *HookValidator {
	return rtui.NewHookValidator(componentID)
}
