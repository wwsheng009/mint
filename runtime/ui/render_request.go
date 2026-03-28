package ui

import "sync"

var (
	globalRenderSchedulerMu sync.RWMutex
	globalRenderScheduler   func()
)

// SetGlobalRenderScheduler registers a process-wide callback used by async component
// callbacks to request a UI update on the next render tick.
func SetGlobalRenderScheduler(fn func()) {
	globalRenderSchedulerMu.Lock()
	defer globalRenderSchedulerMu.Unlock()
	globalRenderScheduler = fn
}

// RequestGlobalRender triggers the registered async render callback when available.
func RequestGlobalRender() {
	globalRenderSchedulerMu.RLock()
	fn := globalRenderScheduler
	globalRenderSchedulerMu.RUnlock()
	if fn != nil {
		fn()
	}
}
