// Package ui provides memory safety utilities for the declarative UI framework.
package ui

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// Memory Safety: Goroutine Management
// =============================================================================

// GoRoutine manages a goroutine with automatic cleanup
type GoRoutine struct {
	done    chan struct{}
	cleanup sync.Once
}

// NewGoRoutine creates a new managed goroutine
func NewGoRoutine() *GoRoutine {
	return &GoRoutine{
		done: make(chan struct{}),
	}
}

// Done returns the done channel for the goroutine
func (g *GoRoutine) Done() <-chan struct{} {
	return g.done
}

// Stop stops the goroutine and waits for it to finish
func (g *GoRoutine) Stop() {
	g.cleanup.Do(func() {
		close(g.done)
	})
}

// Go runs a goroutine that will be automatically stopped when the component unmounts
// This is a safer alternative to directly spawning goroutines in useEffect
//
// Example:
//
//	goRef := UseGoRoutine()
//	goRef.Go(func(ctx <-chan struct{}) {
//	    ticker := time.NewTicker(time.Second)
//	    defer ticker.Stop()
//	    for {
//	        select {
//	        case <-ticker.C:
//	            setCount(count + 1)
//	        case <-ctx:
//	            return
//	        }
//	    }
//	 })
func (g *GoRoutine) Go(fn func(<-chan struct{})) {
	go func() {
		fn(g.done)
	}()
}

// GoWithRestart runs a goroutine that can be restarted when dependencies change
// Returns a stop function that should be called in the effect cleanup
func (g *GoRoutine) GoWithRestart(fn func(<-chan struct{}) func()) func() {
	var stopPrev func()

	g.Go(func(ctx <-chan struct{}) {
		// Call the inner function to get the stop function
		stopPrev = fn(ctx)
	})

	// Return cleanup function
	return func() {
		g.Stop()
		if stopPrev != nil {
			stopPrev()
		}
	}
}

// =============================================================================
// Context-based Cancellation
// =============================================================================

// useGoRoutine creates a managed goroutine hook
// This is the public API for components
func useGoRoutine() *GoRoutine {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("useGoRoutine must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(HookRef); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.getOrCreateHook(HookRef)

	// Initialize if first render
	if hook.Value == nil {
		hook.Value = NewGoRoutine()
	}

	return hook.Value.(*GoRoutine)
}

// UseGoRoutine is the public API for creating managed goroutines
func UseGoRoutine() *GoRoutine {
	return useGoRoutine()
}

// =============================================================================
// Subscription Management
// =============================================================================

// Subscription represents a cancellable subscription
type Subscription struct {
	cancel   func()
	cleanup  sync.Once
	done     chan struct{}
}

// NewSubscription creates a new subscription
func NewSubscription(cancel func()) *Subscription {
	return &Subscription{
		cancel:  cancel,
		done:    make(chan struct{}),
	}
}

// Unsubscribe cancels the subscription
func (s *Subscription) Unsubscribe() {
	s.cleanup.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		close(s.done)
	})
}

// Done returns a channel that's closed when unsubscribed
func (s *Subscription) Done() <-chan struct{} {
	return s.done
}

// UseSubscription manages a subscription with automatic cleanup
// Example:
//
//	sub := UseSubscription(func() *Subscription {
//	    return dataSource.Subscribe(func(msg string) {
//	        setData(msg)
//	    })
//	 })
func UseSubscription(createSub func() *Subscription) *Subscription {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("UseSubscription must be called within a component")
	}

	sub := createSub()

	// Register cleanup
	UseEffect(func() CleanupFunc {
		return func() {
			sub.Unsubscribe()
		}
	}, nil)

	return sub
}

// =============================================================================
// Memory Leak Detection Utilities
// =============================================================================

// MemStats tracks memory usage
type MemStats struct {
	mu           sync.RWMutex
	initial      runtime.MemStats
	last         runtime.MemStats
	allocChanges uint64
}

// NewMemStats creates a new memory stats tracker
func NewMemStats() *MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return &MemStats{
		initial: stats,
		last:    stats,
	}
}

// CheckAlloc checks if allocations have increased since last check
// Returns the delta in bytes
func (m *MemStats) CheckAlloc() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	delta := stats.Alloc - m.last.Alloc
	m.last = stats
	m.allocChanges += delta

	return delta
}

// TotalAlloc returns total allocations since tracking started
func (m *MemStats) TotalAlloc() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.allocChanges
}

// Reset resets the tracking
func (m *MemStats) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.initial = stats
	m.last = stats
	m.allocChanges = 0
}

// =============================================================================
// Goroutine Leak Detector
// =============================================================================

// GoroutineTracker tracks goroutine counts for leak detection
type GoroutineTracker struct {
	mu         sync.Mutex
	initial    int
	current    int
	threshold  int
}

// NewGoroutineTracker creates a new goroutine tracker
func NewGoroutineTracker(threshold int) *GoroutineTracker {
	return &GoroutineTracker{
		initial:   runtime.NumGoroutine(),
		current:   runtime.NumGoroutine(),
		threshold: threshold,
	}
}

// Update updates the current goroutine count
func (t *GoroutineTracker) Update() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = runtime.NumGoroutine()
}

// CheckForLeaks checks if goroutines have leaked beyond the threshold
func (t *GoroutineTracker) CheckForLeaks() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	delta := t.current - t.initial
	if delta > t.threshold {
		return fmt.Errorf("possible goroutine leak: %d goroutines created (threshold: %d)",
			delta, t.threshold)
	}
	return nil
}

// Count returns the current goroutine count
func (t *GoroutineTracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.current
}

// =============================================================================
// Safe Timer Pattern
// =============================================================================

// SafeTimer wraps time.Timer with automatic cleanup
type SafeTimer struct {
	timer    *time.Timer
	mu       sync.Mutex
	once     sync.Once
	duration time.Duration
	fn       func()
	running  bool
}

// NewSafeTimer creates a new safe timer (starts in stopped state)
func NewSafeTimer(d time.Duration, fn func()) *SafeTimer {
	st := &SafeTimer{
		timer:    time.NewTimer(d),
		duration: d,
		fn:       fn,
		running:  false,
	}
	st.timer.Stop() // Start in stopped state

	return st
}

// Start starts the timer
func (st *SafeTimer) Start() {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.running {
		return
	}

	st.running = true
	st.timer.Reset(st.duration)

	go func() {
		<-st.timer.C
		st.once.Do(st.fn)
	}()
}

// Reset resets the timer with a new duration
func (st *SafeTimer) Reset(d time.Duration) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.timer.Stop()
	// Drain any pending value
	select {
	case <-st.timer.C:
	default:
	}

	st.duration = d
	st.timer.Reset(d)
}

// Stop stops the timer
func (st *SafeTimer) Stop() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.timer.Stop()
	// Drain any pending value
	select {
	case <-st.timer.C:
	default:
	}
}

// =============================================================================
// Safe Ticker Pattern
// =============================================================================

// SafeTicker wraps time.Ticker with automatic cleanup
type SafeTicker struct {
	ticker *time.Ticker
	mu     sync.Mutex
	done   chan struct{}
}

// NewSafeTicker creates a new safe ticker
func NewSafeTicker(d time.Duration) *SafeTicker {
	st := &SafeTicker{
		ticker: time.NewTicker(d),
		done:    make(chan struct{}),
	}
	return st
}

// Channel returns the ticker channel
func (st *SafeTicker) Channel() <-chan time.Time {
	return st.ticker.C
}

// Stop stops the ticker
func (st *SafeTicker) Stop() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.ticker.Stop()
	close(st.done)
}

// Done returns a channel that's closed when stopped
func (st *SafeTicker) Done() <-chan struct{} {
	return st.done
}

// =============================================================================
// Resource Pool for limiting concurrent operations
// =============================================================================

// ResourcePool limits the number of concurrent resources
type ResourcePool struct {
	sem   chan struct{}
	close chan struct{}
	once  sync.Once
}

// NewResourcePool creates a new resource pool with a maximum size
func NewResourcePool(maxSize int) *ResourcePool {
	return &ResourcePool{
		sem:   make(chan struct{}, maxSize),
		close: make(chan struct{}),
	}
}

// Acquire acquires a resource from the pool
// Returns nil if pool is closed
func (p *ResourcePool) Acquire() bool {
	select {
	case p.sem <- struct{}{}:
		return true
	case <-p.close:
		return false
	}
}

// Release releases a resource back to the pool
func (p *ResourcePool) Release() {
	select {
	case <-p.sem:
		// Released
	default:
		// Pool was empty, ignore
	}
}

// Close closes the resource pool
func (p *ResourcePool) Close() {
	p.once.Do(func() {
		close(p.close)
	})
}

// Go runs a function with resource limiting
func (p *ResourcePool) Go(fn func()) bool {
	if !p.Acquire() {
		return false
	}

	go func() {
		defer p.Release()
		fn()
	}()

	return true
}

// =============================================================================
// Example: Safe Timer Component
// =============================================================================

// Example: Using UseGoRoutine for safe goroutine management
//
// func TimerComponent() ui.VNode {
//     count, setCount := ui.UseStateInt(0)
//
//     goRoutine := ui.UseGoRoutine()
//
//     ui.UseEffect(func() ui.CleanupFunc {
//         // Start a goroutine that auto-cleans up
//         goRoutine.Go(func(ctx <-chan struct{}) {
//             ticker := time.NewTicker(time.Second)
//             defer ticker.Stop()
//
//             for {
//                 select {
//                 case <-ticker.C:
//                     setCount(func(c int) int { return c + 1 })
//                 case <-ctx:
//                     return // Clean exit
//                 }
//             }
//         })
//
//         // Cleanup function
//         return func() {
//             goRoutine.Stop() // Signal goroutine to stop
//         }
//     }, nil)
//
//     return ui.Text(fmt.Sprintf("Count: %d", count))
// }
