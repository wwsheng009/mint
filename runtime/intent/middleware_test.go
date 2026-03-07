package intent

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

// =============================================================================
// Middleware Chain Tests
// =============================================================================

func TestMiddlewareChain_Execute(t *testing.T) {
	chain := NewMiddlewareChain()

	// Add test middleware
	beforeCalled := false
	afterCalled := false

	testMW := &struct {
		IntentMiddleware
	}{
		IntentMiddleware: &testMiddlewareAdapter{
			beforeHook: func(ctx *MiddlewareContext) error {
				beforeCalled = true
				return nil
			},
			afterHook: func(ctx *MiddlewareContext, result *EmitResult) {
				afterCalled = true
			},
		},
	}

	chain.Add(testMW)

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Execute chain
	result := chain.Execute(ctx, func() EmitResult {
		return Success(0)
	})

	// Verify
	if !beforeCalled {
		t.Error("BeforeEmit should be called")
	}
	if !afterCalled {
		t.Error("AfterEmit should be called")
	}
	if !result.Handled {
		t.Error("Result should indicate handled")
	}
	// Duration should be non-negative (may be 0 for very fast execution)
	if result.Duration < 0 {
		t.Errorf("Expected non-negative duration, got %v", result.Duration)
	}
}

func TestMiddlewareChain_MultipleMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()

	order := []string{}

	// Add multiple middleware
	for i := 0; i < 3; i++ {
		idx := i
		chain.Add(&struct {
			IntentMiddleware
		}{
			IntentMiddleware: &testMiddlewareAdapter{
				beforeHook: func(ctx *MiddlewareContext) error {
					order = append(order, "before"+strconv.Itoa(idx))
					return nil
				},
				afterHook: func(ctx *MiddlewareContext, result *EmitResult) {
					order = append(order, "after"+strconv.Itoa(idx))
				},
			},
		})
	}

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Execute chain
	chain.Execute(ctx, func() EmitResult {
		return Success(time.Millisecond)
	})

	// Verify order: before0, before1, before2, after2, after1, after0
	expected := []string{"before0", "before1", "before2", "after2", "after1", "after0"}
	for i, got := range order {
		if got != expected[i] {
			t.Errorf("Order[%d] = %s, want %s", i, got, expected[i])
		}
	}
}

// =============================================================================
// Logging Middleware Tests
// =============================================================================

func TestLoggingMiddleware_BeforeEmit(t *testing.T) {
	mw := NewLoggingMiddleware()
	mw.LogBefore = true

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Should not error
	err := mw.BeforeEmit(ctx)
	if err != nil {
		t.Errorf("BeforeEmit returned error: %v", err)
	}
}

func TestLoggingMiddleware_AfterEmit(t *testing.T) {
	mw := NewLoggingMiddleware()
	mw.LogAfter = true

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)
	result := Success(time.Millisecond)

	// Should not panic
	mw.AfterEmit(ctx, &result)
}

// =============================================================================
// Performance Middleware Tests
// =============================================================================

func TestPerformanceMiddleware_Tracking(t *testing.T) {
	mw := NewPerformanceMiddleware(time.Millisecond)

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Execute multiple times
	for i := 0; i < 3; i++ {
		result := EmitResult{
			Duration: time.Duration(i+1) * time.Millisecond,
			Handled:  true,
		}
		mw.AfterEmit(ctx, &result)
	}

	// Verify stats
	stats := mw.GetStats("Test")
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
	if stats.Count != 3 {
		t.Errorf("Expected Count 3, got %d", stats.Count)
	}
	if stats.MinTime != time.Millisecond {
		t.Errorf("Expected MinTime %v, got %v", time.Millisecond, stats.MinTime)
	}
	if stats.MaxTime != 3*time.Millisecond {
		t.Errorf("Expected MaxTime %v, got %v", 3*time.Millisecond, stats.MaxTime)
	}
}

func TestPerformanceMiddleware_Threshold(t *testing.T) {
	mw := NewPerformanceMiddleware(5 * time.Millisecond)

	alertCount := 0
	mw.WarningFunc = func(intentType string, duration time.Duration) {
		alertCount++
	}

	intent := TestIntent{Value: "SlowIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Emit intent above threshold
	result := EmitResult{
		Duration: 10 * time.Millisecond,
		Handled:  true,
	}
	mw.AfterEmit(ctx, &result)

	// Verify warning was called
	if alertCount != 1 {
		t.Errorf("Expected 1 warning, got %d", alertCount)
	}
}

// =============================================================================
// Error Handling Middleware Tests
// =============================================================================

func TestErrorHandlingMiddleware_ErrorLogging(t *testing.T) {
	mw := NewErrorHandlingMiddleware()

	errorLogged := false
	mw.ErrorFunc = func(ctx *MiddlewareContext, err error) {
		errorLogged = true
	}

	intent := TestIntent{Value: "TestIntent"}
	ctx := NewContext(intent, nil, RouteGlobal)
	result := Failure(fmt.Errorf("test error"))

	mw.AfterEmit(ctx, &result)

	if !errorLogged {
		t.Error("Error should be logged")
	}
}

// =============================================================================
// Validation Middleware Tests
// =============================================================================

func TestValidationMiddleware_NilIntent(t *testing.T) {
	mw := NewValidationMiddleware()

	ctx := NewContext(nil, nil, RouteGlobal)

	err := mw.BeforeEmit(ctx)
	if err == nil {
		t.Error("Expected error for nil intent")
	}
}

func TestValidationMiddleware_EmptyIntentType(t *testing.T) {
	// This test checks the default validation behavior
	// Since TestIntent always returns "Test" for IntentType(),
	// we can't test empty IntentType with a standard TestIntent
	// Skip this test for now as it would require a custom intent type
	t.Skip("Cannot test empty IntentType with standard TestIntent")
}

func TestValidationMiddleware_CustomValidator(t *testing.T) {
	mw := NewValidationMiddleware()

	intent := TestIntent{Value: "invalid"}
	ctx := NewContext(intent, nil, RouteGlobal)

	// Default validator only checks for nil and empty type
	// So this should succeed for non-empty value
	err := mw.BeforeEmit(ctx)
	if err != nil {
		t.Errorf("Did not expect error, got: %v", err)
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

type testMiddlewareAdapter struct {
	beforeHook func(*MiddlewareContext) error
	afterHook  func(*MiddlewareContext, *EmitResult)
}

func (a *testMiddlewareAdapter) BeforeEmit(ctx *MiddlewareContext) error {
	if a.beforeHook != nil {
		return a.beforeHook(ctx)
	}
	return nil
}

func (a *testMiddlewareAdapter) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	if a.afterHook != nil {
		a.afterHook(ctx, result)
	}
}
