package intent

import (
	"testing"

	mintlog "github.com/wwsheng009/mint/internal/log"
)

// =============================================================================
// Default Logger Tests
// =============================================================================

func TestDispatcherDefaultLogger(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)

	// Verify defaults
	if d.GetLogger() == nil {
		t.Error("Dispatcher should have a default logger")
	}

	// Verify it's the IntentLogger
	if d.GetLogger() != mintlog.IntentLogger {
		t.Error("Dispatcher should use IntentLogger by default")
	}
}

func TestRegistryDefaultLogger(t *testing.T) {
	r := NewRegistry()

	// Verify defaults
	if r.GetLogger() == nil {
		t.Error("Registry should have a default logger")
	}

	// Verify it's the IntentLogger
	if r.GetLogger() != mintlog.IntentLogger {
		t.Error("Registry should use IntentLogger by default")
	}
}

func TestRuntimeDefaultLogger(t *testing.T) {
	rt := NewRuntime()

	// Verify defaults
	if rt.GetLogger() == nil {
		t.Error("Runtime should have a default logger")
	}

	// Verify it's the IntentLogger
	if rt.GetLogger() != mintlog.IntentLogger {
		t.Error("Runtime should use IntentLogger by default")
	}
}

func TestRuntimeNewRegistryDefaultLogger(t *testing.T) {
	rt := NewRuntimeWithNewRegistry()

	// Verify defaults (uses same IntentLogger globally)
	if rt.GetLogger() == nil {
		t.Error("Runtime should have a default logger")
	}

	if rt.GetLogger() != mintlog.IntentLogger {
		t.Error("Runtime should use IntentLogger by default")
	}
}

// =============================================================================
// Custom Logger Tests
// =============================================================================

func TestDispatcherCustomLogger(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)

	// Create custom logger
	customLogger := mintlog.NewLogger("Custom", "CUSTOM")

	// Set custom logger
	d.SetLogger(customLogger)

	// Verify custom logger is now used
	if d.GetLogger() != customLogger {
		t.Error("Dispatcher should use custom logger after SetLogger")
	}

	// Should not be the IntentLogger anymore
	if d.GetLogger() == mintlog.IntentLogger {
		t.Error("Dispatcher should not use IntentLogger after SetLogger is called")
	}
}

func TestRegistryCustomLogger(t *testing.T) {
	r := NewRegistry()

	// Create custom logger
	customLogger := mintlog.NewLogger("Custom", "CUSTOM")

	// Set custom logger
	r.SetLogger(customLogger)

	// Verify custom logger is now used
	if r.GetLogger() != customLogger {
		t.Error("Registry should use custom logger after SetLogger")
	}

	// Should not be the IntentLogger anymore
	if r.GetLogger() == mintlog.IntentLogger {
		t.Error("Registry should not use IntentLogger after SetLogger is called")
	}
}

func TestRuntimeCustomLogger(t *testing.T) {
	rt := NewRuntime()

	// Create custom logger
	customLogger := mintlog.NewLogger("Custom", "CUSTOM")

	// Set custom logger via Runtime
	rt.SetLogger(customLogger)

	// Verify custom logger is now used
	if rt.GetLogger() != customLogger {
		t.Error("Runtime should use custom logger after SetLogger")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestDispatcherUsesLogger(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)

	// Enable IntentLogger
	mintlog.IntentLogger.SetEnabled(true)

	// Register handler
	called := false
	r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
		called = true
		return HandledResult()
	}))

	// Dispatch will trigger logging
	result := d.Dispatch(TestIntent{})

	if !result.Handled {
		t.Error("Dispatch should succeed")
	}

	if !called {
		t.Error("Handler should be called")
	}

	// Reset
	mintlog.IntentLogger.SetEnabled(false)
}
