package intent

import (
	"errors"
	"testing"

	"github.com/wwsheng009/mint/runtime/priority"
)

// =============================================================================
// Test Intent Types
// =============================================================================

type TestIntent struct {
	Value string
}

func (TestIntent) IntentType() string {
	return "Test"
}

type HighPriorityIntent struct{}

func (HighPriorityIntent) IntentType() string {
	return "HighPriority"
}

func (HighPriorityIntent) Priority() ActionPriority {
	return PriorityImmediate
}

type TestTransitionIntent struct {
	URL string
}

func (TestTransitionIntent) IntentType() string {
	return "TestTransition"
}

func (TestTransitionIntent) IsTransition() bool {
	return true
}

// =============================================================================
// Types Tests
// =============================================================================

func TestActionPriority_String(t *testing.T) {
	tests := []struct {
		p        ActionPriority
		expected string
	}{
		{PriorityImmediate, "Immediate"},
		{PriorityUserBlocking, "UserBlocking"},
		{PriorityNormal, "Normal"},
		{PriorityTransition, "Transition"},
		{PriorityIdle, "Idle"},
		{ActionPriority(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.p.String(); got != tt.expected {
				t.Errorf("ActionPriority.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActionPriority_ToLane(t *testing.T) {
	tests := []struct {
		p        ActionPriority
		expected priority.DirtyLevel
	}{
		{PriorityImmediate, priority.DirtyHigh},
		{PriorityUserBlocking, priority.DirtyHigh},
		{PriorityNormal, priority.DirtyNormal},
		{PriorityTransition, priority.DirtyLow},
		{PriorityIdle, priority.DirtyLow},
		{ActionPriority(99), priority.DirtyNormal},
	}

	for _, tt := range tests {
		t.Run(tt.p.String(), func(t *testing.T) {
			if got := tt.p.ToLane(); got != tt.expected {
				t.Errorf("ActionPriority.ToLane() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIntentResult_Constructors(t *testing.T) {
	t.Run("HandledResult", func(t *testing.T) {
		r := HandledResult()
		if !r.Handled || r.Error != nil || r.Async {
			t.Errorf("HandledResult() = %+v, want {Handled:true}", r)
		}
	})

	t.Run("ErrorResult", func(t *testing.T) {
		err := errors.New("test error")
		r := ErrorResult(err)
		if r.Handled || r.Error != err {
			t.Errorf("ErrorResult() = %+v, want {Error:%v}", r, err)
		}
	})

	t.Run("AsyncResult", func(t *testing.T) {
		done := make(chan struct{})
		r := AsyncResult(done)
		if !r.Handled || !r.Async || r.Done != done {
			t.Errorf("AsyncResult() = %+v, want {Handled:true, Async:true}", r)
		}
	})
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	handler := HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		return HandledResult()
	})

	unregister := r.Register("Test", handler)

	if !r.HasHandler("Test") {
		t.Error("HasHandler() = false, want true")
	}

	unregister()

	if r.HasHandler("Test") {
		t.Error("HasHandler() = true after unregister, want false")
	}
}

func TestRegistry_RegisterTyped(t *testing.T) {
	r := NewRegistry()

	unregister := RegisterTyped(r, func(ctx *ActionContext, intent TestIntent) IntentResult {
		return HandledResult()
	})

	if !r.HasHandler("Test") {
		t.Error("HasHandler() = false, want true")
	}

	unregister()

	if r.HasHandler("Test") {
		t.Error("HasHandler() = true after unregister, want false")
	}
}

func TestRegistry_GetPriority(t *testing.T) {
	r := NewRegistry()

	t.Run("Default", func(t *testing.T) {
		intent := TestIntent{}
		if p := r.GetPriority(intent); p != PriorityNormal {
			t.Errorf("GetPriority() = %v, want Normal", p)
		}
	})

	t.Run("PriorityAware", func(t *testing.T) {
		intent := HighPriorityIntent{}
		if p := r.GetPriority(intent); p != PriorityImmediate {
			t.Errorf("GetPriority() = %v, want Immediate", p)
		}
	})

	t.Run("Explicit", func(t *testing.T) {
		r.RegisterPriority("Test", PriorityUserBlocking)
		intent := TestIntent{}
		if p := r.GetPriority(intent); p != PriorityUserBlocking {
			t.Errorf("GetPriority() = %v, want UserBlocking", p)
		}
	})
}

func TestRegistry_IsTransition(t *testing.T) {
	r := NewRegistry()

	t.Run("NotTransition", func(t *testing.T) {
		intent := TestIntent{}
		if r.IsTransition(intent) {
			t.Error("IsTransition() = true for non-transition intent")
		}
	})

	t.Run("Transition", func(t *testing.T) {
		intent := TestTransitionIntent{}
		if !r.IsTransition(intent) {
			t.Error("IsTransition() = false for transition intent")
		}
	})
}

func TestRegistry_Middleware(t *testing.T) {
	r := NewRegistry()
	callOrder := []string{}

	r.Use(func(next Handler) Handler {
		return HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
			callOrder = append(callOrder, "middleware1")
			return next.Handle(ctx, intent)
		})
	})

	r.Use(func(next Handler) Handler {
		return HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
			callOrder = append(callOrder, "middleware2")
			return next.Handle(ctx, intent)
		})
	})

	r.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		callOrder = append(callOrder, "handler")
		return HandledResult()
	}))

	handler, ok := r.GetHandler("Test")
	if !ok {
		t.Fatal("GetHandler() returned false")
	}

	handler.Handle(nil, TestIntent{})

	expected := []string{"middleware1", "middleware2", "handler"}
	if len(callOrder) != len(expected) {
		t.Errorf("Middleware call order = %v, want %v", callOrder, expected)
	}
}

// =============================================================================
// Dispatcher Tests
// =============================================================================

type mockStateSetter struct {
	state map[string]interface{}
	dirty bool
}

func newMockStateSetter() *mockStateSetter {
	return &mockStateSetter{state: make(map[string]interface{})}
}

func (m *mockStateSetter) SetState(key string, value interface{}) {
	m.state[key] = value
	m.dirty = true
}

func (m *mockStateSetter) GetState(key string) (interface{}, bool) {
	v, ok := m.state[key]
	return v, ok
}

func (m *mockStateSetter) ScheduleUpdate() {
	m.dirty = true
}

func TestDispatcher_Dispatch(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)
	store := newMockStateSetter()
	dispatcher.SetStateSetter(store)

	handled := false
	registry.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		handled = true
		if ctx == nil {
			t.Error("ctx is nil")
		}
		return HandledResult()
	}))

	result := dispatcher.Dispatch(TestIntent{Value: "hello"})

	if !handled {
		t.Error("Handler not called")
	}
	if !result.Handled {
		t.Error("Dispatch() returned Handled = false")
	}
}

func TestDispatcher_Dispatch_NoHandler(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)

	result := dispatcher.Dispatch(TestIntent{})

	if result.Handled {
		t.Error("Dispatch() should return Handled = false for unregistered intent")
	}
	if result.Error == nil {
		t.Error("Dispatch() should return error for unregistered intent")
	}
}

func TestDispatcher_Dispatch_WithPriority(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)

	var receivedPriority ActionPriority
	registry.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		receivedPriority = registry.GetPriority(intent)
		return HandledResult()
	}))

	result := dispatcher.DispatchWithPriority(TestIntent{}, "test-source", PriorityImmediate)

	if !result.Handled {
		t.Error("DispatchWithPriority() failed")
	}
	// Note: The priority is computed from the intent, not passed through
	_ = receivedPriority
}

func TestDispatcher_Scheduler(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)

	scheduled := false
	dispatcher.SetScheduler(func(lane priority.DirtyLevel) {
		scheduled = true
	})

	registry.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		return HandledResult()
	}))

	dispatcher.Dispatch(TestIntent{})

	if !scheduled {
		t.Error("Scheduler not called after dispatch")
	}
}

// =============================================================================
// SimpleStore Tests
// =============================================================================

func TestSimpleStore(t *testing.T) {
	store := NewSimpleStore()

	t.Run("SetGet", func(t *testing.T) {
		store.SetState("key", "value")
		v, ok := store.GetState("key")
		if !ok || v != "value" {
			t.Errorf("GetState() = (%v, %v), want (value, true)", v, ok)
		}
	})

	t.Run("IsDirty", func(t *testing.T) {
		store.ClearDirty()
		if store.IsDirty() {
			t.Error("IsDirty() = true after ClearDirty()")
		}
		store.SetState("key", "newvalue")
		if !store.IsDirty() {
			t.Error("IsDirty() = false after SetState()")
		}
	})
}

// =============================================================================
// Runtime Tests
// =============================================================================

func TestRuntime(t *testing.T) {
	rt := NewRuntime()

	handled := false
	RegisterTypedRuntime(rt, func(ctx *ActionContext, intent TestIntent) IntentResult {
		handled = true
		ctx.SetState("test", intent.Value)
		return HandledResult()
	})

	result := rt.Emit(TestIntent{Value: "hello"})

	if !handled {
		t.Error("Handler not called")
	}
	if !result.Handled {
		t.Error("Emit() returned Handled = false")
	}

	v, ok := rt.Store.GetState("test")
	if !ok || v != "hello" {
		t.Errorf("Store state = (%v, %v), want (hello, true)", v, ok)
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder(t *testing.T) {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)

	handled := false
	registry.Register("Test", HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
		handled = true
		return HandledResult()
	}))

	result := NewBuilder(TestIntent{Value: "test"}).
		WithPriority(PriorityImmediate).
		WithSource("test-source").
		Dispatch(dispatcher)

	if !handled || !result.Handled {
		t.Error("Builder dispatch failed")
	}
}

// =============================================================================
// Builtin Tests
// =============================================================================

func TestBuiltinIntents(t *testing.T) {
	t.Run("Navigate", func(t *testing.T) {
		intent := Navigate("/path")
		if intent.IntentType() != "Navigate" {
			t.Errorf("IntentType() = %v, want Navigate", intent.IntentType())
		}
		if intent.Priority() != PriorityUserBlocking {
			t.Errorf("Priority() = %v, want UserBlocking", intent.Priority())
		}
	})

	t.Run("OpenModal", func(t *testing.T) {
		intent := OpenModal("modal-1")
		if intent.IntentType() != "OpenModal" {
			t.Errorf("IntentType() = %v, want OpenModal", intent.IntentType())
		}
		if intent.Priority() != PriorityUserBlocking {
			t.Errorf("Priority() = %v, want UserBlocking", intent.Priority())
		}
	})

	t.Run("LoadData", func(t *testing.T) {
		intent := LoadData("/api/data", "dataKey")
		if intent.IntentType() != "LoadData" {
			t.Errorf("IntentType() = %v, want LoadData", intent.IntentType())
		}
		if !intent.IsTransition() {
			t.Error("IsTransition() = false, want true")
		}
	})

	t.Run("Focus", func(t *testing.T) {
		intent := Focus("element-1")
		if intent.IntentType() != "Focus" {
			t.Errorf("IntentType() = %v, want Focus", intent.IntentType())
		}
		if intent.Priority() != PriorityImmediate {
			t.Errorf("Priority() = %v, want Immediate", intent.Priority())
		}
	})
}

// =============================================================================
// Transition Tests
// =============================================================================

func TestTransitionWrapper(t *testing.T) {
	intent := TestIntent{Value: "test"}
	wrapped := Transition(intent)

	if !wrapped.IsTransition() {
		t.Error("Transition() wrapper did not mark intent as transition")
	}
	if wrapped.IntentType() != "Test" {
		t.Errorf("IntentType() = %v, want Test", wrapped.IntentType())
	}
}

func TestPriorityWrapper(t *testing.T) {
	intent := TestIntent{Value: "test"}
	wrapped := WithPriority(intent, PriorityIdle)

	if wrapped.Priority() != PriorityIdle {
		t.Errorf("WithPriority() did not override priority")
	}
	if wrapped.IntentType() != "Test" {
		t.Errorf("IntentType() = %v, want Test", wrapped.IntentType())
	}
}
