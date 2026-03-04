package intent

import (
	"sync/atomic"
	"testing"

	mintlog "github.com/wwsheng009/mint/internal/log"
)

// =============================================================================
// WithOverridable Tests
// =============================================================================

func TestRegistry_WithOverridable(t *testing.T) {
	t.Run("Overridable handler can be replaced", func(t *testing.T) {
		r := NewRegistry()
		logger := mintlog.NewLogger("Test", "TEST")
		r.SetLogger(logger)
		logger.SetEnabled(true)

		firstCalled := int32(0)
		secondCalled := int32(0)

		// Register first handler as overridable
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			atomic.AddInt32(&firstCalled, 1)
			return HandledResult()
		}), WithOverridable(true))

		// Register second handler - should replace first
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			atomic.AddInt32(&secondCalled, 1)
			return HandledResult()
		}))

		// Get and call handler
		handler, ok := r.GetHandler("Test")
		if !ok {
			t.Fatal("GetHandler() returned false")
		}
		handler.Handle(nil, TestIntent{})

		// Only second handler should be called
		if atomic.LoadInt32(&firstCalled) != 0 {
			t.Error("First handler should not be called")
		}
		if atomic.LoadInt32(&secondCalled) != 1 {
			t.Error("Second handler should be called once")
		}
	})

	t.Run("Protected handler cannot be replaced", func(t *testing.T) {
		r := NewRegistry()
		logger := mintlog.NewLogger("Test", "TEST")
		r.SetLogger(logger)
		logger.SetEnabled(true)

		firstCalled := int32(0)
		secondCalled := int32(0)

		// Register first handler as NOT overridable (default)
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			atomic.AddInt32(&firstCalled, 1)
			return HandledResult()
		}))

		// Try to register second handler - should be ignored
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			atomic.AddInt32(&secondCalled, 1)
			return HandledResult()
		}))

		// Get and call handler
		handler, ok := r.GetHandler("Test")
		if !ok {
			t.Fatal("GetHandler() returned false")
		}
		handler.Handle(nil, TestIntent{})

		// Only first handler should be called
		if atomic.LoadInt32(&firstCalled) != 1 {
			t.Error("First handler should be called once")
		}
		if atomic.LoadInt32(&secondCalled) != 0 {
			t.Error("Second handler should not be called")
		}
	})

	t.Run("New registration with overridable flag", func(t *testing.T) {
		r := NewRegistry()
		logger := mintlog.NewLogger("Test", "TEST")
		r.SetLogger(logger)
		logger.SetEnabled(true)

		callCount := int32(0)

		// Register as non-overridable
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			atomic.AddInt32(&callCount, 1)
			return HandledResult()
		}), WithOverridable(false))

		// Try to replace
		r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
			return HandledResult()
		}))

		handler, _ := r.GetHandler("Test")
		handler.Handle(nil, TestIntent{})

		if atomic.LoadInt32(&callCount) != 1 {
			t.Error("Original handler should still be called")
		}
	})
}

func TestRegistry_WithHandlerPriority(t *testing.T) {
	r := NewRegistry()

	r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
		return HandledResult()
	}), WithHandlerPriority(PriorityImmediate))

	// Verify handler exists
	if !r.HasHandler("Test") {
		t.Error("Handler not registered")
	}

	// Priority is stored internally, can't easily verify from public API
	// The fact it compiled successfully is the main test
}

func TestRegistry_UnregisterWithLogger(t *testing.T) {
	r := NewRegistry()
	logger := mintlog.NewLogger("Test", "TEST")
	r.SetLogger(logger)
	logger.SetEnabled(true)

	// Register first handler
	r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
		return HandledResult()
	}))

	// Verify handler exists
	if !r.HasHandler("Test") {
		t.Fatal("Handler should be registered")
	}

	// Unregister
	unregister := r.Register("Test", HandlerFunc(func(ctx *ActionContext, i Intent) IntentResult {
		return HandledResult()
	}))
	unregister()

	// After unregister, the first handler should still be there (protected by default)
	if !r.HasHandler("Test") {
		t.Error("First handler should still exist after unregistering second handler")
	}
}

// =============================================================================
// RegisterTypedWithOpts Tests
// =============================================================================

func TestRegisterTypedWithOpts(t *testing.T) {
	r := NewRegistry()
	logger := mintlog.NewLogger("Test", "TEST")
	r.SetLogger(logger)
	logger.SetEnabled(true)

	firstCalled := int32(0)
	secondCalled := int32(0)

	// Register typed handler with options
	unregister1 := RegisterTypedWithOpts(r, func(ctx *ActionContext, i TestIntent) IntentResult {
		atomic.AddInt32(&firstCalled, 1)
		return HandledResult()
	}, WithOverridable(true))

	// Replace with new handler
	unregister2 := RegisterTypedWithOpts(r, func(ctx *ActionContext, i TestIntent) IntentResult {
		atomic.AddInt32(&secondCalled, 1)
		return HandledResult()
	})

	handler, _ := r.GetHandler("Test")
	handler.Handle(nil, TestIntent{})

	if atomic.LoadInt32(&firstCalled) != 0 {
		t.Error("First handler should not be called")
	}
	if atomic.LoadInt32(&secondCalled) != 1 {
		t.Error("Second handler should be called")
	}

	unregister1()
	unregister2()
}
