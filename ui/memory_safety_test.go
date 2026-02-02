package ui

import (
	"sync"
	"testing"
	"time"
)

// =============================================================================
// GoRoutine Tests
// =============================================================================

func TestGoRoutine_BasicUsage(t *testing.T) {
	gr := NewGoRoutine()
	count := 0
	var mu sync.Mutex

	gr.Go(func(ctx <-chan struct{}) {
		for i := 0; i < 3; i++ {
			select {
			case <-ctx:
				return
			case <-time.After(10 * time.Millisecond):
				mu.Lock()
				count++
				mu.Unlock()
			}
		}
	})

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	mu.Unlock()
}

func TestGoRoutine_Stop(t *testing.T) {
	gr := NewGoRoutine()
	count := 0
	var mu sync.Mutex

	gr.Go(func(ctx <-chan struct{}) {
		for {
			select {
			case <-ctx:
				mu.Lock()
				count++
				mu.Unlock()
				return
			case <-time.After(100 * time.Millisecond):
				// Should not reach here
				mu.Lock()
				count = 999
				mu.Unlock()
			}
		}
	})

	// Stop immediately
	gr.Stop()
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	if count != 1 {
		t.Errorf("goroutine should have stopped, count = %d", count)
	}
	mu.Unlock()
}

func TestGoRoutine_StopIdempotent(t *testing.T) {
	gr := NewGoRoutine()

	// Multiple stops should be safe
	gr.Stop()
	gr.Stop()
	gr.Stop()

	// Should not panic
}

func TestGoRoutine_WithRestart(t *testing.T) {
	gr := NewGoRoutine()
	count := 0
	var mu sync.Mutex

	// First run
	cleanup1 := gr.GoWithRestart(func(ctx <-chan struct{}) func() {
		go func() {
			for {
				select {
				case <-ctx:
					return
				case <-time.After(10 * time.Millisecond):
					mu.Lock()
					count++
					mu.Unlock()
					return
				}
			}
		}()
		return func() { mu.Lock(); count += 100; mu.Unlock() }
	})

	time.Sleep(30 * time.Millisecond)
	cleanup1() // Stop first run

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	if count != 101 { // 1 from goroutine, 100 from cleanup
		t.Errorf("count = %d, want 101", count)
	}
	mu.Unlock()
}

// =============================================================================
// Subscription Tests
// =============================================================================

func TestSubscription_Basic(t *testing.T) {
	called := false
	sub := NewSubscription(func() {
		called = true
	})

	sub.Unsubscribe()

	if !called {
		t.Error("cancel function should be called")
	}

	// Should be idempotent
	sub.Unsubscribe()
	sub.Unsubscribe()
}

func TestSubscription_DoneChannel(t *testing.T) {
	sub := NewSubscription(func() {})

	// Done channel should not be closed initially
	select {
	case <-sub.Done():
		t.Fatal("Done channel should not be closed initially")
	default:
	}

	sub.Unsubscribe()

	// Done channel should be closed after unsubscribe
	select {
	case <-sub.Done():
		// Expected
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Done channel should be closed after Unsubscribe")
	}
}

// =============================================================================
// MemStats Tests
// =============================================================================

func TestMemStats_CheckAlloc(t *testing.T) {
	ms := NewMemStats()

	// Allocate some memory
	data := make([]byte, 1024*1024) // 1MB

	delta := ms.CheckAlloc()
	if delta < 1024*1024 {
		t.Errorf("CheckAlloc() = %d, should be at least %d", delta, 1024*1024)
	}

	_ = data // Use the variable
}

func TestMemStats_TotalAlloc(t *testing.T) {
	ms := NewMemStats()

	// First allocation
	_ = make([]byte, 1000)
	ms.CheckAlloc()

	// Second allocation
	_ = make([]byte, 2000)
	ms.CheckAlloc()

	total := ms.TotalAlloc()
	if total < 3000 {
		t.Errorf("TotalAlloc() = %d, want at least 3000", total)
	}
}

func TestMemStats_Reset(t *testing.T) {
	ms := NewMemStats()

	_ = make([]byte, 1000)
	ms.CheckAlloc()

	ms.Reset()

	// After reset, total should be small
	_ = make([]byte, 500)
	ms.CheckAlloc()

	total := ms.TotalAlloc()
	if total > 1000 {
		t.Errorf("After Reset(), TotalAlloc() = %d, want < 1000", total)
	}
}

// =============================================================================
// GoroutineTracker Tests
// =============================================================================

func TestGoroutineTracker_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine tracking test in short mode")
	}

	gt := NewGoroutineTracker(10)

	initial := gt.Count()
	if initial < 1 {
		t.Errorf("Count() = %d, want at least 1", initial)
	}

	// Update should work
	gt.Update()
	current := gt.Count()
	if current < 1 {
		t.Errorf("After Update(), Count() = %d, want at least 1", current)
	}

	// Should not report leaks
	err := gt.CheckForLeaks()
	if err != nil {
		t.Errorf("CheckForLeaks() = %v, want nil", err)
	}
}

func TestGoroutineTracker_DetectLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping leak detection test in short mode")
	}

	gt := NewGoroutineTracker(0) // Zero threshold = any new goroutine is a "leak"

	// Spawn some goroutines
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		}()
	}

	gt.Update()

	// Should report leaks
	err := gt.CheckForLeaks()
	if err == nil {
		t.Error("CheckForLeaks() should detect goroutine increase")
	}

	wg.Wait()
}

// =============================================================================
// SafeTimer Tests
// =============================================================================

func TestSafeTimer_Basic(t *testing.T) {
	called := false
	st := NewSafeTimer(10*time.Millisecond, func() {
		called = true
	})

	st.Start() // Must explicitly start
	time.Sleep(50 * time.Millisecond)

	if !called {
		t.Error("timer callback should be called")
	}

	st.Stop() // Cleanup
}

func TestSafeTimer_Reset(t *testing.T) {
	count := 0
	st := NewSafeTimer(10*time.Millisecond, func() {
		count++
	})

	st.Start()

	// Reset before first fire
	time.Sleep(5 * time.Millisecond)
	st.Reset(10 * time.Millisecond)

	// Wait for original time (should not fire)
	time.Sleep(8 * time.Millisecond)
	if count != 0 {
		t.Errorf("count = %d, want 0 (timer should have been reset)", count)
	}

	// Wait for reset time
	time.Sleep(20 * time.Millisecond)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	st.Stop()
}

func TestSafeTimer_Stop(t *testing.T) {
	count := 0
	st := NewSafeTimer(10*time.Millisecond, func() {
		count++
	})

	// Don't start, just stop
	st.Stop()
	// And try stopping again
	st.Stop()

	time.Sleep(50 * time.Millisecond)

	if count != 0 {
		t.Errorf("count = %d, want 0 (timer should be stopped)", count)
	}
}

// =============================================================================
// SafeTicker Tests
// =============================================================================

func TestSafeTicker_Basic(t *testing.T) {
	st := NewSafeTicker(10 * time.Millisecond)
	defer st.Stop()

	count := 0
	done := make(chan struct{})

	go func() {
		for range st.Channel() {
			count++
			if count >= 3 {
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ticker")
	}
}

func TestSafeTicker_Stop(t *testing.T) {
	st := NewSafeTicker(10 * time.Millisecond)

	count := 0
	go func() {
		for range st.Channel() {
			count++
			if count >= 1 {
				st.Stop()
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	// Should have stopped at count = 1
	if count != 1 {
		t.Errorf("count = %d, want 1 (ticker should have stopped)", count)
	}
}

func TestSafeTicker_DoneChannel(t *testing.T) {
	st := NewSafeTicker(10 * time.Millisecond)

	// Done should not be closed initially
	select {
	case <-st.Done():
		t.Fatal("Done should not be closed initially")
	default:
	}

	st.Stop()

	// Done should be closed after Stop
	select {
	case <-st.Done():
		// Expected
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Done should be closed after Stop")
	}
}

// =============================================================================
// ResourcePool Tests
// =============================================================================

func TestResourcePool_Basic(t *testing.T) {
	pool := NewResourcePool(2)

	// Acquire resources
	if !pool.Acquire() {
		t.Error("First Acquire() should succeed")
	}
	if !pool.Acquire() {
		t.Error("Second Acquire() should succeed")
	}

	// Pool is full
	// This would need a goroutine to test blocking, but we skip that

	// Release resources
	pool.Release()
	pool.Release()

	// Should be able to acquire again
	if !pool.Acquire() {
		t.Error("Acquire() after Release should succeed")
	}

	pool.Release()
}

func TestResourcePool_Go(t *testing.T) {
	pool := NewResourcePool(2)
	var wg sync.WaitGroup
	count := 0
	var mu sync.Mutex

	// Launch 4 tasks but pool only allows 2 concurrent
	for i := 0; i < 4; i++ {
		wg.Add(1)
		if !pool.Go(func() {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
			mu.Lock()
			count++
			mu.Unlock()
		}) {
			wg.Done()
			t.Error("Go() should succeed for all 4 tasks (serialized)")
		}
	}

	wg.Wait()

	if count != 4 {
		t.Errorf("count = %d, want 4", count)
	}
}

func TestResourcePool_Close(t *testing.T) {
	pool := NewResourcePool(1)

	pool.Close()

	// Acquire should fail after close
	if pool.Acquire() {
		t.Error("Acquire() should fail after Close()")
	}

	// Close should be idempotent
	pool.Close()
	pool.Close()
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestUseGoRoutine_InComponent(t *testing.T) {
	ctx := NewComponentContextForRoot()
	SetCurrentContext(ctx)
	defer SetCurrentContext(nil)

	gr := UseGoRoutine()

	count := 0
	var mu sync.Mutex

	gr.Go(func(ctx <-chan struct{}) {
		for i := 0; i < 3; i++ {
			select {
			case <-ctx:
				return
			case <-time.After(5 * time.Millisecond):
				mu.Lock()
				count++
				mu.Unlock()
			}
		}
	})

	// Simulate component unmount
	time.Sleep(30 * time.Millisecond)
	gr.Stop()

	mu.Lock()
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	mu.Unlock()
}

func TestUseGoRoutine_HookPersistence(t *testing.T) {
	ctx := NewComponentContextForRoot()
	SetCurrentContext(ctx)
	defer SetCurrentContext(nil)

	// First render
	gr1 := UseGoRoutine()
	if gr1 == nil {
		t.Error("First UseGoRoutine() should return a GoRoutine")
	}

	// Re-render - should return same instance
	ctx.ResetContext()
	gr2 := UseGoRoutine()

	if gr1 != gr2 {
		t.Error("UseGoRoutine() should return same instance across renders")
	}
}

func TestUseSubscription_Basic(t *testing.T) {
	ctx := NewComponentContextForRoot()
	SetCurrentContext(ctx)
	defer SetCurrentContext(nil)

	var count int
	var mu sync.Mutex

	sub := UseSubscription(func() *Subscription {
		// Simulate a data source subscription
		go func() {
			for i := 0; i < 3; i++ {
				select {
				case <-time.After(5 * time.Millisecond):
					mu.Lock()
					count++
					mu.Unlock()
				}
			}
		}()

		return NewSubscription(func() {
			mu.Lock()
			count += 100 // Mark as cleaned up
			mu.Unlock()
		})
	})

	// Wait for some data
	time.Sleep(30 * time.Millisecond)

	// Cleanup happens via useEffect
	// The subscription should be cleaned up when context is cleaned
	ctx.CleanupAll()

	sub.Unsubscribe()

	mu.Lock()
	// We should have received data AND cleanup was called
	if count < 103 { // 3 from data + 100 from cleanup
		t.Errorf("count = %d, want at least 103", count)
	}
	mu.Unlock()
}
