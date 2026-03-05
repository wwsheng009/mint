package store

import (
	"sync"
	"testing"
)

func TestNewStore(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})
	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	state := store.Get()
	if state.Count != 0 {
		t.Errorf("expected Count=0, got %d", state.Count)
	}
}

func TestGetSet(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	// Test Set
	store.Set(TestState{Count: 42})
	state := store.Get()
	if state.Count != 42 {
		t.Errorf("expected Count=42, got %d", state.Count)
	}
}

func TestUpdate(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	// Test Update
	store.Update(func(s TestState) TestState {
		s.Count++
		return s
	})

	state := store.Get()
	if state.Count != 1 {
		t.Errorf("expected Count=1, got %d", state.Count)
	}
}

func TestSubscribe(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	called := false
	var mu sync.Mutex

	unsubscribe := store.Subscribe(func(state TestState) {
		mu.Lock()
		defer mu.Unlock()
		called = true
	})

	store.Set(TestState{Count: 1})

	mu.Lock()
	calledResult := called
	mu.Unlock()

	if !calledResult {
		t.Error("Subscribe callback was not called")
	}

	// Test unsubscribe
	mu.Lock()
	called = false
	mu.Unlock()

	unsubscribe()
	store.Set(TestState{Count: 2})

	mu.Lock()
	calledResult = called
	mu.Unlock()

	if calledResult {
		t.Error("Subscribe callback was called after unsubscribe")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	calls := make([]bool, 3)
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		idx := i
		store.Subscribe(func(state TestState) {
			mu.Lock()
			defer mu.Unlock()
			calls[idx] = true
		})
	}

	store.Set(TestState{Count: 1})

	mu.Lock()
	defer mu.Unlock()

	for i, called := range calls {
		if !called {
			t.Errorf("Subscriber %d was not called", i)
		}
	}
}

func TestComputed(t *testing.T) {
	type TestState struct {
		Values []int
	}

	store := NewStore(TestState{Values: []int{1, 2, 3}})

	// Simple computed
	sumComputed := NewComputed(store, func(s TestState) int {
		sum := 0
		for _, v := range s.Values {
			sum += v
		}
		return sum
	})

	sum := sumComputed.Get()
	if sum != 6 {
		t.Errorf("expected sum=6, got %d", sum)
	}

	// Second call should use cached value
	sum2 := sumComputed.Get()
	if sum2 != 6 {
		t.Errorf("expected sum=6, got %d", sum2)
	}

	// Update state should invalidate cache
	store.Set(TestState{Values: []int{4, 5, 6}})
	sum3 := sumComputed.Get()
	if sum3 != 15 {
		t.Errorf("expected sum=15, got %d", sum3)
	}
}

func TestComputedWithInvalidator(t *testing.T) {
	type TestState struct {
		Count int
		Name  string
	}

	store := NewStore(TestState{Count: 0, Name: "test"})

	// Computed with invalidator - only invalidate when Count changes
	countComputed := NewComputedWithInvalidator(store,
		func(s TestState) int {
			return s.Count
		},
		func(oldState, newState TestState) bool {
			// Only invalidate if Count changed
			return oldState.Count != newState.Count
		},
	)

	// Initial value
	count := countComputed.Get()
	if count != 0 {
		t.Errorf("expected count=0, got %d", count)
	}

	// Update Name (should not invalidate)
	store.Set(TestState{Count: 0, Name: "updated"})
	count2 := countComputed.Get()
	if count2 != 0 {
		t.Errorf("expected count=0, got %d", count2)
	}

	// Update Count (should invalidate)
	store.Set(TestState{Count: 42, Name: "updated"})
	count3 := countComputed.Get()
	if count3 != 42 {
		t.Errorf("expected count=42, got %d", count3)
	}
}

func TestComputedInvalidate(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	callCount := 0
	computed := NewComputed(store, func(s TestState) int {
		callCount++
		return s.Count
	})

	// First call computes
	_ = computed.Get()
	if callCount != 1 {
		t.Errorf("expected callCount=1, got %d", callCount)
	}

	// Invalidate the cache
	computed.Invalidate()

	// Second call should recompute after invalidation
	_ = computed.Get()
	if callCount != 2 {
		t.Errorf("expected callCount=2, got %d", callCount)
	}
}

func TestComputedDispose(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	callCount := 0
	mu := &sync.Mutex{}

	computed := NewComputed(store, func(s TestState) int {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		return s.Count
	})

	// Get initial value to ensure computed is subscribed
	_ = computed.Get()

	mu.Lock()
	initialCalls := callCount
	mu.Unlock()

	// Dispose the computed
	computed.Dispose()

	// Update store - computed should not be called
	store.Set(TestState{Count: 1})

	mu.Lock()
	defer mu.Unlock()

	if callCount != initialCalls {
		t.Errorf("expected callCount to remain %d, got %d", initialCalls, callCount)
	}
}

func TestSelectCache(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	callCount := 0
	cache := NewSelectCache(store, func(s TestState) int {
		callCount++
		return s.Count * 2
	})

	// First call computes
	result := cache.Get()
	if result != 0 {
		t.Errorf("expected result=0, got %d", result)
	}
	if callCount != 1 {
		t.Errorf("expected callCount=1, got %d", callCount)
	}

	// Second call uses cache
	result2 := cache.Get()
	if result2 != 0 {
		t.Errorf("expected result=0, got %d", result2)
	}
	if callCount != 1 { // Should still be 1 (using cache)
		t.Errorf("expected callCount=1, got %d", callCount)
	}

	// Update state - should invalidate
	store.Set(TestState{Count: 5})
	result3 := cache.Get()
	if result3 != 10 {
		t.Errorf("expected result=10, got %d", result3)
	}
	if callCount != 2 {
		t.Errorf("expected callCount=2, got %d", callCount)
	}
}

func TestSelectCacheWithInvalidator(t *testing.T) {
	type TestState struct {
		Count int
		Name  string
	}

	store := NewStore(TestState{Count: 0, Name: "test"})

	callCount := 0
	cache := NewSelectCacheWithInvalidator(store,
		func(s TestState) int {
			callCount++
			return s.Count
		},
		func(old, new TestState) bool {
			// Only invalidate when Count changes
			return old.Count != new.Count
		},
	)

	// Initial call
	_ = cache.Get()
	if callCount != 1 {
		t.Errorf("expected callCount=1, got %d", callCount)
	}

	// Update Name (should not invalidate)
	store.Set(TestState{Count: 0, Name: "updated"})
	_ = cache.Get()
	if callCount != 1 {
		t.Errorf("expected callCount=1, got %d", callCount)
	}

	// Update Count (should invalidate)
	store.Set(TestState{Count: 5, Name: "updated"})
	_ = cache.Get()
	if callCount != 2 {
		t.Errorf("expected callCount=2, got %d", callCount)
	}
}

func TestSelectCacheInvalidateNow(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 5})

	cache := NewSelectCache(store, func(s TestState) int {
		return s.Count * 2
	})

	// Initial value
	result := cache.Get()
	if result != 10 {
		t.Errorf("expected result=10, got %d", result)
	}

	// Update store
	store.Set(TestState{Count: 10})

	// InvalidateNow should invalidate and return new value
	result2 := cache.InvalidateNow()
	if result2 != 20 {
		t.Errorf("expected result=20, got %d", result2)
	}
}

func TestConcurrentAccess(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			store.Update(func(s TestState) TestState {
				s.Count++
				return s
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Get()
		}()
	}

	wg.Wait()

	finalCount := store.Get().Count
	if finalCount != 100 {
		t.Errorf("expected final count=100, got %d", finalCount)
	}
}

func TestListenerCount(t *testing.T) {
	type TestState struct {
		Count int
	}

	store := NewStore(TestState{Count: 0})

	if store.ListenerCount() != 0 {
		t.Errorf("expected 0 listeners, got %d", store.ListenerCount())
	}

	unsubscribe1 := store.Subscribe(func(s TestState) {})
	if store.ListenerCount() != 1 {
		t.Errorf("expected 1 listener, got %d", store.ListenerCount())
	}

	unsubscribe2 := store.Subscribe(func(s TestState) {})
	if store.ListenerCount() != 2 {
		t.Errorf("expected 2 listeners, got %d", store.ListenerCount())
	}

	unsubscribe1()
	if store.ListenerCount() != 1 {
		t.Errorf("expected 1 listener, got %d", store.ListenerCount())
	}

	unsubscribe2()
	if store.ListenerCount() != 0 {
		t.Errorf("expected 0 listeners, got %d", store.ListenerCount())
	}
}
