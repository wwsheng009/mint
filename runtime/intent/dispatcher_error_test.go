package intent

import (
	"errors"
	"testing"
	"time"

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

// =============================================================================
// ErrorLogRetry Tests
// =============================================================================

func TestErrorHandlingStrategy_ErrorLogRetry(t *testing.T) {
	logger := mintlog.NewLogger("RetryTest", "RETRY")
	logger.SetEnabled(true)

	// Test 1: Retry with SetMaxRetry
	t.Run("WithCustomMaxRetry", func(t *testing.T) {
		r := NewRegistry()
		d := NewDispatcher(r)
		store := newMockStateSetter()
		d.SetStateSetter(store)
		d.SetErrorStrategy(ErrorLogRetry)
		d.SetMaxRetry(2) // Retry 2 times (total 3 attempts)
		d.SetLogger(logger)

		// Register a handler that always fails but allows testing retry behavior
		attemptCount := 0
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
			attemptCount++
			return ErrorResult(errors.New("test error"))
		}))

		// Should attempt 1 initial + 2 retries = 3 total
		startTime := time.Now()
		_ = d.Dispatch(TestIntent{})
		duration := time.Since(startTime)

		// At least 2 retries with ~50ms delay each => ~150ms minimum
		// Allow some tolerance
		if duration < 100*time.Millisecond {
			t.Errorf("Expected at least ~100ms for retries, got %v", duration)
		}
		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}
	})

	// Test 2: Verify SetMaxRetry method
	t.Run("SetMaxRetry", func(t *testing.T) {
		r := NewRegistry()
		d := NewDispatcher(r)
		d.SetMaxRetry(5)
		// Verify it doesn't panic
	})

	// Test 3: Retry with handler that eventually succeeds
	t.Run("RetryUntilSuccess", func(t *testing.T) {
		r := NewRegistry()
		d := NewDispatcher(r)
		store := newMockStateSetter()
		d.SetStateSetter(store)
		d.SetErrorStrategy(ErrorLogRetry)
		d.SetMaxRetry(5)
		d.SetLogger(logger)

		attemptCount := 0
		r.Register("RetrySuccess", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
			attemptCount++
			if attemptCount < 3 {
				return ErrorResult(errors.New("not yet"))
			}
			return HandledResult()
		}))

		// Dispatch will return first failure result, but retry will succeed
		result := d.Dispatch(RetryIntent{})

		// Original dispatch returns error (first attempt failed)
		if result.Error == nil {
			t.Errorf("Expected first attempt to fail")
		}

		// But 3 attempts should have been made (1 initial + 2 retries before success)
		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}

		// Verify scheduler was called (indicating handler succeeded after retry)
		// We can check this by ensuring attemptCount reached 3 and test didn't fail
	})

	// Test 4: Default retry count (3 retries)
	t.Run("DefaultRetryCount", func(t *testing.T) {
		r := NewRegistry()
		d := NewDispatcher(r)
		store := newMockStateSetter()
		d.SetStateSetter(store)
		d.SetErrorStrategy(ErrorLogRetry)
		// Don't set MaxRetry, should use default of 3
		d.SetLogger(logger)

		attemptCount := 0
		r.Register("DefaultRetry", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
			attemptCount++
			return ErrorResult(errors.New("always fails"))
		}))

		_ = d.Dispatch(DefaultRetryIntent{})
		// Should attempt 1 initial + 3 retries = 4 total
		if attemptCount != 4 {
			t.Errorf("Expected 4 attempts with default retry count, got %d", attemptCount)
		}
	})
}

// RetryIntent is used for retry tests
type RetryIntent struct{}

func (RetryIntent) IntentType() string { return "RetrySuccess" }

// DefaultRetryIntent is used to test default retry behavior
type DefaultRetryIntent struct{}

func (DefaultRetryIntent) IntentType() string { return "DefaultRetry" }
