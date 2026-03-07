package intent

import (
	"fmt"
)

// =============================================================================
// Error Handling Middleware
// =============================================================================

// ErrorHandlingMiddleware handles intent emission errors.
type ErrorHandlingMiddleware struct {
	// ErrorFunc is called when an error occurs.
	ErrorFunc func(ctx *MiddlewareContext, err error)

	// LogErrors enables error logging.
	LogErrors bool

	// Recover enables panic recovery.
	Recover bool
}

// NewErrorHandlingMiddleware creates a new error handling middleware.
func NewErrorHandlingMiddleware() *ErrorHandlingMiddleware {
	return &ErrorHandlingMiddleware{
		ErrorFunc: func(ctx *MiddlewareContext, err error) {
			fmt.Printf("[IntentError] Type: %s, Error: %v\n",
				ctx.Intent.IntentType(), err)
		},
		LogErrors: true,
		Recover:   false,
	}
}

// BeforeEmit is a no-op.
func (m *ErrorHandlingMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	return nil
}

// AfterEmit handles errors.
func (m *ErrorHandlingMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	if result.Error == nil {
		return
	}

	if m.LogErrors && m.ErrorFunc != nil {
		m.ErrorFunc(ctx, result.Error)
	}
}

// WithErrorFunc sets a custom error handler.
func (m *ErrorHandlingMiddleware) WithErrorFunc(fn func(ctx *MiddlewareContext, err error)) *ErrorHandlingMiddleware {
	m.ErrorFunc = fn
	return m
}

// WithRecover enables or disables panic recovery.
func (m *ErrorHandlingMiddleware) WithRecover(enable bool) *ErrorHandlingMiddleware {
	m.Recover = enable
	return m
}

// =============================================================================
// Recovery Middleware (Panic Recovery)
// =============================================================================

// RecoveryMiddleware recovers from panics during intent emission.
type RecoveryMiddleware struct {
	// RecoverFunc is called after panic recovery.
	RecoverFunc func(ctx *MiddlewareContext, v interface{})

	// LogPanics enables panic logging.
	LogPanics bool
}

// NewRecoveryMiddleware creates a new recovery middleware.
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{
		RecoverFunc: func(ctx *MiddlewareContext, v interface{}) {
			fmt.Printf("[IntentPanic] Type: %s, Panic: %v\n",
				ctx.Intent.IntentType(), v)
		},
		LogPanics: true,
	}
}

// BeforeEmit is a no-op.
func (m *RecoveryMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	return nil
}

// AfterEmit checks if there was a panic and recovers.
func (m *RecoveryMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	// Check if the emitFunc panicked
	if result.Error != nil && result.Error.Error() == "panic recovered" {
		if m.LogPanics && m.RecoverFunc != nil {
			m.RecoverFunc(ctx, result.Error)
		}
	}
}

// WithRecoverFunc sets a custom recovery handler.
func (m *RecoveryMiddleware) WithRecoverFunc(fn func(ctx *MiddlewareContext, v interface{})) *RecoveryMiddleware {
	m.RecoverFunc = fn
	return m
}

// =============================================================================
// Validation Middleware
// =============================================================================

// ValidationMiddleware validates intents before emission.
type ValidationMiddleware struct {
	// Validator is a custom validation function.
	Validator func(Intent) error
}

// NewValidationMiddleware creates a new validation middleware.
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		Validator: func(intent Intent) error {
			// Default: check for nil intents
			if intent == nil {
				return fmt.Errorf("intent cannot be nil")
			}
			// Default: require valid IntentType
			if intent.IntentType() == "" {
				return fmt.Errorf("intent must have a valid IntentType")
			}
			return nil
		},
	}
}

// BeforeEmit validates the intent.
func (m *ValidationMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	if m.Validator != nil {
		return m.Validator(ctx.Intent)
	}
	return nil
}

// AfterEmit is a no-op.
func (m *ValidationMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	// No-op
}

// WithValidator sets a custom validation function.
func (m *ValidationMiddleware) WithValidator(fn func(Intent) error) *ValidationMiddleware {
	m.Validator = fn
	return m
}
