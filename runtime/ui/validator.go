package ui

import (
	"fmt"
	"runtime"
)

// HookValidator validates hook call order and consistency
type HookValidator struct {
	componentID   string
	expectedOrder []HookType
	currentIndex  int
	isFirstRender bool
}

// HookOrderError represents a hook order violation
type HookOrderError struct {
	Component    string
	Message      string
	Position     int
	Expected     int
	Actual       int
	ExpectedType HookType
	ActualType   HookType
}

func (e *HookOrderError) Error() string {
	baseMsg := fmt.Sprintf("[Hooks Error] %s in component '%s'", e.Message, e.Component)
	if e.Position > 0 {
		baseMsg += fmt.Sprintf(" at position %d", e.Position)
	}
	if e.ExpectedType != 0 && e.ActualType != 0 {
		baseMsg += fmt.Sprintf(": expected %s, got %s", e.ExpectedType, e.ActualType)
	} else if e.Expected > 0 && e.Actual > 0 {
		baseMsg += fmt.Sprintf(": expected %d hooks, got %d", e.Expected, e.Actual)
	}
	baseMsg += "\n\nHint: Hooks must be called in the same order on every render.\n" +
		"Make sure hooks are not called inside conditions, loops, or nested functions."
	return baseMsg
}

// NewHookValidator creates a new hook validator
func NewHookValidator(componentID string) *HookValidator {
	return &HookValidator{
		componentID:   componentID,
		expectedOrder: make([]HookType, 0),
		currentIndex:  0,
		isFirstRender: true,
	}
}

// ValidateHookCall validates a hook call
func (v *HookValidator) ValidateHookCall(hookType HookType) error {
	if v.isFirstRender {
		// First render - record the hook order
		v.expectedOrder = append(v.expectedOrder, hookType)
		v.currentIndex++
		return nil
	}

	// Subsequent render - validate the hook order
	if v.currentIndex >= len(v.expectedOrder) {
		return &HookOrderError{
			Component: v.componentID,
			Message:   "Hook call count exceeds first render",
			Expected:  len(v.expectedOrder),
			Actual:    v.currentIndex + 1,
		}
	}

	expected := v.expectedOrder[v.currentIndex]
	if hookType != expected {
		// Get caller location for better error message
		_, file, line, _ := runtime.Caller(2)
		return &HookOrderError{
			Component:    v.componentID,
			Message:      fmt.Sprintf("Hook call order mismatch at %s:%d", file, line),
			Position:     v.currentIndex,
			ExpectedType: expected,
			ActualType:   hookType,
		}
	}

	v.currentIndex++
	return nil
}

// FinishRender finishes the render and validates hook count
func (v *HookValidator) FinishRender() error {
	if !v.isFirstRender && v.currentIndex != len(v.expectedOrder) {
		return &HookOrderError{
			Component: v.componentID,
			Message:   "Hook call count is less than first render",
			Expected:  len(v.expectedOrder),
			Actual:    v.currentIndex,
		}
	}

	v.isFirstRender = false
	v.currentIndex = 0
	return nil
}

// Reset resets the validator for a new component
func (v *HookValidator) Reset() {
	v.expectedOrder = make([]HookType, 0)
	v.currentIndex = 0
	v.isFirstRender = true
}
