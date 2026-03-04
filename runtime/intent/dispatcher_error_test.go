package intent

import (
	"errors"
	"testing"

	mintlog "github.com/wwsheng009/mint/internal/log"
)

// =============================================================================
// ErrorHandlingStrategy Tests
// =============================================================================

func TestErrorHandlingStrategy_ErrorLogIgnore(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)
	store := newMockStateSetter()
	d.SetStateSetter(store)

	d.SetErrorStrategy(ErrorLogIgnore)
	logger := mintlog.NewLogger("Test", "TEST")
	d.SetLogger(logger)
	logger.SetEnabled(true)

	// Dispatch without handler
	result := d.Dispatch(TestIntent{})

	if result.Handled {
		t.Error("Should return unhandled result")
	}
	if result.Error == nil {
		t.Error("Should return error")
	}
	// Should not panic
}

func TestErrorHandlingStrategy_ErrorLogPanic(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)
	store := newMockStateSetter()
	d.SetStateSetter(store)

	d.SetErrorStrategy(ErrorLogPanic)

	// Should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Should have panicked")
		}
	}()

	result := d.Dispatch(TestIntent{})

	// Should reach here if panic didn't happen
	_ = result.Error
}

func TestErrorHandlingStrategy_CustomCallback(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)
	store := newMockStateSetter()
	d.SetStateSetter(store)

	d.SetErrorStrategy(ErrorCustomCallback)

	capturedIntent := Intent(nil)
	capturedError := error(nil)
	d.SetErrorHandler(func(intent Intent, err error) {
		capturedIntent = intent
		capturedError = err
	})

	result := d.Dispatch(TestIntent{})

	if result.Handled {
		t.Error("Should return unhandled result")
	}
	if result.Error == nil {
		t.Error("Should return error")
	}
	if capturedIntent == nil {
		t.Error("Error handler should be called with intent")
	}
	if capturedError == nil {
		t.Error("Error handler should be called with error")
	}
}

func TestErrorHandlingStrategy_HandlerFailure(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)
	store := newMockStateSetter()
	d.SetStateSetter(store)

	// Register a handler that returns an error
	errorIntent := TestIntent{}
	r.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		return ErrorResult(errors.New("handler error"))
	}))

	testCases := []struct {
		name     string
		strategy ErrorHandlingStrategy
		wantPanic bool
	}{
		{"LogIgnore", ErrorLogIgnore, false},
		{"LogPanic", ErrorLogPanic, true},
		{"CustomCallback", ErrorCustomCallback, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDispatcher(r)
			d.SetStateSetter(store)
			d.SetErrorStrategy(tc.strategy)

			if tc.strategy == ErrorCustomCallback {
				d.SetErrorHandler(func(intent Intent, err error) {
					// Just capture, don't test for simplicity
				})
			}

			if tc.wantPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("Should have panicked")
					}
				}()
			}

			result := d.Dispatch(errorIntent)

			if result.Error == nil {
				t.Error("Should return error")
			}
		})
	}
}

func TestDispatcher_Logger(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)
	store := newMockStateSetter()
	d.SetStateSetter(store)

	logger := mintlog.NewLogger("DispatcherTest", "DISPATCHER")
	logger.SetEnabled(true)
	d.SetLogger(logger)

	r.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		return HandledResult()
	}))

	// Dispatch should trigger logging (but we can't verify output easily)
	result := d.Dispatch(TestIntent{})

	if !result.Handled {
		t.Error("Dispatch should succeed")
	}
}

func TestDispatcher_SetLogger(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)

	logger := mintlog.NewLogger("Test", "TEST")
	d.SetLogger(logger)
}

func TestDispatcher_EnableLog(t *testing.T) {
	r := NewRegistry()
	d := NewDispatcher(r)

	// Test EnableLog (deprecated but should still work)
	d.EnableLog(true)
	d.EnableLog(false)

	// Just verify it compiles
}
