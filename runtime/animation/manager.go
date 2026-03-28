package animation

import (
	"sync"
	"time"
)

// Manager manages running animations.
type Manager struct {
	animations map[string]*Animation
	mu         sync.RWMutex
	running    bool
	ticker     *time.Ticker
	stopChan   chan struct{}
	frameTime  time.Duration
	lastTick   time.Time
}

// NewManager creates a new animation manager.
func NewManager() *Manager {
	return &Manager{
		animations: make(map[string]*Animation),
		running:    false,
		stopChan:   make(chan struct{}),
		frameTime:  time.Second / 60,
	}
}

// Start starts the animation manager.
func (m *Manager) Start(fps int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	if fps <= 0 {
		fps = 60
	}
	interval := time.Second / time.Duration(fps)
	m.frameTime = interval
	m.lastTick = time.Time{}
	m.ticker = time.NewTicker(interval)
	ticker := m.ticker
	stopChan := m.stopChan

	go m.updateLoop(ticker, stopChan)
}

// Stop stops the animation manager.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	m.ticker.Stop()
	m.ticker = nil
	m.lastTick = time.Time{}
	close(m.stopChan)
	m.stopChan = make(chan struct{})
}

// updateLoop runs the animation update loop.
func (m *Manager) updateLoop(ticker *time.Ticker, stopChan <-chan struct{}) {
	for {
		select {
		case now := <-ticker.C:
			m.Tick(now)
		case <-stopChan:
			return
		}
	}
}

// Update updates all running animations.
func (m *Manager) Update() {
	m.UpdateWithDelta(m.getFrameTime())
}

// UpdateWithDelta updates all running animations using an explicit frame delta.
func (m *Manager) UpdateWithDelta(delta time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateWithDeltaLocked(delta)
}

// Tick advances running animations using wall-clock timestamps.
// The first tick primes the manager and does not advance progress.
func (m *Manager) Tick(now time.Time) {
	if now.IsZero() {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastTick.IsZero() {
		m.lastTick = now
		return
	}

	delta := now.Sub(m.lastTick)
	if delta < 0 {
		delta = 0
	}
	m.lastTick = now
	m.updateWithDeltaLocked(delta)
}

func (m *Manager) updateWithDeltaLocked(delta time.Duration) {
	if delta < 0 {
		delta = 0
	}
	for _, anim := range m.animations {
		if anim.State == AnimationStateRunning {
			m.updateAnimation(anim, delta)
		}
	}
	m.cleanupCompletedAnimations()
}

// getFrameTime returns the time per frame based on ticker interval.
func (m *Manager) getFrameTime() time.Duration {
	if m.frameTime <= 0 {
		return time.Second / 60
	}
	return m.frameTime
}

// updateAnimation updates a single animation.
func (m *Manager) updateAnimation(anim *Animation, delta time.Duration) {
	// Handle delay
	if anim.Delay > 0 && anim.Elapsed == 0 {
		anim.Delay -= delta
		if anim.Delay < 0 {
			anim.Delay = 0
		}
		return
	}

	// Update elapsed time
	anim.Elapsed += delta

	// Clamp to duration
	if anim.Elapsed > anim.Duration {
		anim.Elapsed = anim.Duration
	}

	// Calculate progress and apply easing
	progress := anim.GetEasedProgress()
	anim.Current = interpolate(anim.From, anim.To, progress)

	// Call progress callback
	if anim.OnProgress != nil {
		anim.OnProgress(progress)
	}

	// Check if animation is complete
	if anim.IsFinished() {
		if anim.Repeat == 0 || (anim.Repeat > 0 && anim.Repeat == 1) {
			anim.State = AnimationStateCompleted
			if anim.OnComplete != nil {
				anim.OnComplete()
			}
		} else {
			// Handle repeat
			if anim.Repeat > 0 {
				anim.Repeat--
			}

			// Reset for next iteration
			if anim.Alternate {
				// Swap from/to for alternate
				anim.From, anim.To = anim.To, anim.From
			}

			anim.Elapsed = 0
			if anim.RepeatDelay > 0 {
				anim.Delay = anim.RepeatDelay
			}
		}
	}
}

// interpolate interpolates between two values based on progress.
func interpolate(from, to interface{}, progress float64) interface{} {
	// Handle different types
	switch from := from.(type) {
	case float64:
		toVal, ok := to.(float64)
		if ok {
			return from + (toVal-from)*progress
		}
	case int:
		toVal, ok := to.(int)
		if ok {
			return from + int(float64(toVal-from)*progress)
		}
	case string:
		// String interpolation could be implemented for typewriter effect
		toVal, ok := to.(string)
		if ok {
			// Simple character-by-character interpolation
			fromLen := len(from)
			toLen := len(toVal)
			if fromLen == toLen {
				// Character-by-character (for color transitions, etc.)
				// This is a simplified version
				if progress < 0.5 {
					return from
				}
				return toVal
			}
			// Length-based interpolation (typewriter effect)
			targetLen := int(float64(toLen) * progress)
			if targetLen > toLen {
				targetLen = toLen
			}
			return toVal[:targetLen]
		}
	}

	return to
}

// cleanupCompletedAnimations removes non-repeating completed animations.
func (m *Manager) cleanupCompletedAnimations() {
	for id, anim := range m.animations {
		if anim.State == AnimationStateCompleted {
			delete(m.animations, id)
		}
	}
}

// Add adds an animation to the manager.
func (m *Manager) Add(anim *Animation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.animations[anim.ID] = anim
}

// Remove removes an animation from the manager.
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.animations, id)
}

// Get returns an animation by ID.
func (m *Manager) Get(id string) (*Animation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	anim, ok := m.animations[id]
	return anim, ok
}

// StartAnimation starts an animation.
func (m *Manager) StartAnimation(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	anim, ok := m.animations[id]
	if !ok {
		return false
	}

	anim.State = AnimationStateRunning
	return true
}

// PauseAnimation pauses an animation.
func (m *Manager) PauseAnimation(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	anim, ok := m.animations[id]
	if !ok {
		return false
	}

	anim.State = AnimationStatePaused
	return true
}

// StopAnimation stops and resets an animation.
func (m *Manager) StopAnimation(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	anim, ok := m.animations[id]
	if !ok {
		return false
	}

	anim.State = AnimationStateIdle
	anim.Reset()
	return true
}

// CancelAnimation cancels an animation.
func (m *Manager) CancelAnimation(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	anim, ok := m.animations[id]
	if !ok {
		return false
	}

	anim.State = AnimationStateCancelled
	delete(m.animations, id)
	return true
}

// Clear removes all animations.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.animations = make(map[string]*Animation)
}

// Count returns the number of animations.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.animations)
}

// GetRunningCount returns the number of running animations.
func (m *Manager) GetRunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, anim := range m.animations {
		if anim.State == AnimationStateRunning {
			count++
		}
	}
	return count
}

// HasRunning returns true if there are running animations.
func (m *Manager) HasRunning() bool {
	return m.GetRunningCount() > 0
}

// GetAll returns all animations.
func (m *Manager) GetAll() map[string]*Animation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*Animation, len(m.animations))
	for id, anim := range m.animations {
		result[id] = anim
	}
	return result
}

// IsRunning returns true if the manager is running.
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running
}
