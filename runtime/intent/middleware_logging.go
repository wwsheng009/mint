package intent

import (
	"fmt"
)

// =============================================================================
// Logging Middleware
// =============================================================================

// LoggingMiddleware logs all intent emissions.
type LoggingMiddleware struct {
	// Logger is the logging function.
	// If nil, uses fmt.Printf.
	Logger func(string, ...interface{})

	// LogBefore enables logging before emission.
	LogBefore bool

	// LogAfter enables logging after emission.
	LogAfter bool

	// LogDetails enables detailed logging (IntentType, RouteType).
	LogDetails bool
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger:     nil, // Use fmt.Printf
		LogBefore:  true,
		LogAfter:   false,
		LogDetails: true,
	}
}

// BeforeEmit logs intent before emission.
func (m *LoggingMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	if !m.LogBefore {
		return nil
	}

	logger := m.getLogger()
	msg := fmt.Sprintf("[IntentEmit] Before: Type=%s, Route=%s", ctx.Intent.IntentType(), ctx.Route)

	if m.LogDetails {
		if ctx.Source != nil {
			msg += fmt.Sprintf(", Source=%T", ctx.Source)
		}
		msg += fmt.Sprintf(", TraceID=%s", ctx.TraceID)
	}

	logger("%s\n", msg)
	return nil
}

// AfterEmit logs intent after emission.
func (m *LoggingMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	if !m.LogAfter {
		return
	}

	logger := m.getLogger()
	msg := fmt.Sprintf("[IntentEmit] After: Type=%s, Route=%s", ctx.Intent.IntentType(), ctx.Route)

	if m.LogDetails {
		msg += fmt.Sprintf(", Duration=%v", result.Duration)
		if result.Error != nil {
			msg += fmt.Sprintf(", Error=%v", result.Error)
		}
		msg += fmt.Sprintf(", Handled=%v", result.Handled)
	}

	logger("%s\n", msg)
}

// getLogger returns the logger function.
func (m *LoggingMiddleware) getLogger() func(string, ...any) {
	if m.Logger != nil {
		// The Logger is of type func(string, ...interface{})
		// We need to adapt it to func(string, ...any)
		return func(format string, a ...any) {
			// Convert []any to []interface{}
			args := make([]interface{}, len(a))
			for i, v := range a {
				args[i] = v
			}
			m.Logger(format, args...)
		}
	}
	// Wrap fmt.Printf to match expected signature
	return func(format string, a ...any) {
		fmt.Printf(format, a...)
	}
}

// =============================================================================
// Console Logging Middleware (Simple)
// =============================================================================

// ConsoleLogger is a simple console-only logger.
type ConsoleLogger struct {
	// Enable enables console logging.
	Enable bool
}

// NewConsoleLogger creates a new console logger.
func NewConsoleLogger(enable bool) *ConsoleLogger {
	return &ConsoleLogger{Enable: enable}
}

// BeforeEmit logs to console if enabled.
func (m *ConsoleLogger) BeforeEmit(ctx *MiddlewareContext) error {
	if !m.Enable {
		return nil
	}

	fmt.Printf("[ConsoleLogger] %s → %s\n", ctx.Intent.IntentType(), ctx.Route)
	return nil
}

// AfterEmit is a no-op.
func (m *ConsoleLogger) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	// No-op
}

// =============================================================================
// Detailed Logging Middleware (JSON-like)
// =============================================================================

// DetailedLogger logs detailed intent information.
type DetailedLogger struct {
	// Enable enables detailed logging.
	Enable bool
}

// NewDetailedLogger creates a new detailed logger.
func NewDetailedLogger(enable bool) *DetailedLogger {
	return &DetailedLogger{Enable: enable}
}

// BeforeEmit logs detailed information.
func (m *DetailedLogger) BeforeEmit(ctx *MiddlewareContext) error {
	if !m.Enable {
		return nil
	}

	fmt.Printf("[DetailedLogger] {\n")
	fmt.Printf("  TraceID: %s\n", ctx.TraceID)
	fmt.Printf("  IntentType: %s\n", ctx.Intent.IntentType())
	fmt.Printf("  Route: %s\n", ctx.Route)
	if ctx.Source != nil {
		fmt.Printf("  Source: %T\n", ctx.Source)
	}
	fmt.Printf("}\n")
	return nil
}

// AfterEmit logs detailed result information.
func (m *DetailedLogger) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	if !m.Enable {
		return
	}

	fmt.Printf("[DetailedLogger Result] {\n")
	fmt.Printf("  TraceID: %s\n", ctx.TraceID)
	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Handled: %v\n", result.Handled)
	if result.Error != nil {
		fmt.Printf("  Error: %v\n", result.Error)
	}
	fmt.Printf("}\n")
}
