package reducer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
)

// Test Intent types
type TestIncIntent struct{}

func (TestIncIntent) IntentType() string { return "TestInc" }

type TestDecIntent struct{}

func (TestDecIntent) IntentType() string { return "TestDec" }

// Test Store
type testStore struct {
	state int
}

func (s *testStore) Get() int         { return s.state }
func (s *testStore) Set(state int)    { s.state = state }

// =============================================================================
// BuildAndRegister Tests
// =============================================================================

func TestBuildAndRegister(t *testing.T) {
	registry := intent.NewRegistry()
	store := &testStore{state: 0}

	// Build reducer and register handlers
	rd := NewBuilder[int]().
		On(TestIncIntent{}, func(s int, i intent.Intent) int {
			return s + 1
		}).
		On(TestDecIntent{}, func(s int, i intent.Intent) int {
			return s - 1
		}).
		BuildAndRegister(registry, store)

	if rd == nil {
		t.Error("Reducer should not be nil")
	}

	// Verify handlers are registered
	if !registry.HasHandler("TestInc") {
		t.Error("TestInc handler should be registered")
	}
	if !registry.HasHandler("TestDec") {
		t.Error("TestDec handler should be registered")
	}

	// Test that handlers work through registry
	handler, ok := registry.GetHandler("TestInc")
	if !ok {
		t.Fatal("TestInc handler not found")
	}

	ctx := intent.NewActionContext(nil, "test", nil)
	result := handler.Handle(ctx, TestIncIntent{})
	if !result.Handled {
		t.Error("Handler should return handled result")
	}
	if store.Get() != 1 {
		t.Errorf("Store state = %d, want 1", store.Get())
	}

	// Test decrement
	handler, _ = registry.GetHandler("TestDec")
	handler.Handle(ctx, TestDecIntent{})
	if store.Get() != 0 {
		t.Errorf("Store state = %d, want 0", store.Get())
	}
}

func TestRegisterToGlobal(t *testing.T) {
	// Reset global registry for test
	registry := intent.DefaultRegistry()
	store := &testStore{state: 10}

	// Register to global
	rd := NewBuilder[int]().
		On(TestIncIntent{}, func(s int, i intent.Intent) int {
			return s + 5
		}).
		RegisterToGlobal(store)

	if rd == nil {
		t.Error("Reducer should not be nil")
	}

	// Verify handler works
	handler, ok := registry.GetHandler("TestInc")
	if !ok {
		t.Fatal("TestInc handler not found in global registry")
	}

	ctx := intent.NewActionContext(nil, "test", nil)
	handler.Handle(ctx, TestIncIntent{})
	if store.Get() != 15 {
		t.Errorf("Store state = %d, want 15", store.Get())
	}
}

func TestBuilderWithStore(t *testing.T) {
	store := &testStore{state: 5}

	rd, update := NewBuilder[int]().
		On(TestIncIntent{}, func(s int, i intent.Intent) int {
			return s + 1
		}).
		BuildWithStore(store)

	if rd == nil {
		t.Error("Reducer should not be nil")
	}

	// Test update function
	update(TestIncIntent{})
	if store.Get() != 6 {
		t.Errorf("Store state = %d, want 6", store.Get())
	}
}
