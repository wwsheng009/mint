package animation

import "time"

// TweenDriverConfig configures a value tween driven by external ticks.
type TweenDriverConfig struct {
	From      float64
	To        float64
	Duration  time.Duration
	Delay     time.Duration
	Easing    EasingFunction
	AutoStart bool
}

// TweenDriver interpolates a float value without owning a ticker.
// Call Tick(now) from the framework's existing render loop.
type TweenDriver struct {
	from           float64
	to             float64
	value          float64
	duration       time.Duration
	delay          time.Duration
	delayRemaining time.Duration
	easing         EasingFunction
	running        bool
	done           bool
	started        bool
	lastTick       time.Time
	elapsed        time.Duration
}

// NewTweenDriver creates a new externally-driven tween.
func NewTweenDriver(cfg TweenDriverConfig) *TweenDriver {
	if cfg.Duration <= 0 {
		cfg.Duration = time.Millisecond
	}
	if cfg.Easing == nil {
		cfg.Easing = Linear
	}

	driver := &TweenDriver{
		from:     cfg.From,
		to:       cfg.To,
		value:    cfg.From,
		duration: cfg.Duration,
		delay:    cfg.Delay,
		easing:   cfg.Easing,
	}
	driver.Reset()
	if cfg.AutoStart {
		driver.Start()
	}
	return driver
}

// Start restarts the tween from the initial state.
func (d *TweenDriver) Start() {
	d.running = true
	d.done = false
	d.started = d.delay <= 0
	d.delayRemaining = d.delay
	d.lastTick = time.Time{}
	d.elapsed = 0
	d.value = d.from
}

// Prime anchors the tween's time origin without advancing progress.
func (d *TweenDriver) Prime(now time.Time) {
	d.lastTick = now
}

// Primed reports whether the tween has a time anchor.
func (d *TweenDriver) Primed() bool {
	return !d.lastTick.IsZero()
}

// Stop stops the tween without resetting the current value.
func (d *TweenDriver) Stop() {
	d.running = false
}

// Reset returns the tween to its initial value and clears progress.
func (d *TweenDriver) Reset() {
	d.running = false
	d.done = false
	d.started = d.delay <= 0
	d.delayRemaining = d.delay
	d.lastTick = time.Time{}
	d.elapsed = 0
	d.value = d.from
}

// WantsTick reports whether the tween still needs time updates.
func (d *TweenDriver) WantsTick() bool {
	return d.running && !d.done
}

// Tick advances the tween to the supplied time.
// It returns true when the visible tween state changed.
func (d *TweenDriver) Tick(now time.Time) bool {
	if !d.WantsTick() {
		return false
	}

	if d.lastTick.IsZero() {
		d.lastTick = now
		return false
	}

	delta := now.Sub(d.lastTick)
	if delta < 0 {
		delta = 0
	}
	d.lastTick = now

	changed := false
	if d.delayRemaining > 0 {
		if delta < d.delayRemaining {
			d.delayRemaining -= delta
			return false
		}
		delta -= d.delayRemaining
		d.delayRemaining = 0
		d.started = true
		changed = true
	}

	if delta == 0 {
		return changed
	}

	prevValue := d.value
	d.elapsed += delta
	if d.elapsed >= d.duration {
		d.elapsed = d.duration
		d.value = d.to
		d.running = false
		d.done = true
		return changed || d.value != prevValue
	}

	progress := float64(d.elapsed) / float64(d.duration)
	d.value = d.from + (d.to-d.from)*d.easing(progress)
	return changed || d.value != prevValue
}

// Value returns the current interpolated value.
func (d *TweenDriver) Value() float64 {
	return d.value
}

// Progress returns normalized tween progress in the range [0, 1].
func (d *TweenDriver) Progress() float64 {
	if d.duration <= 0 {
		return 1
	}
	progress := float64(d.elapsed) / float64(d.duration)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// Started reports whether the delay window has elapsed.
func (d *TweenDriver) Started() bool {
	return d.started
}

// Done reports whether the tween reached the final value.
func (d *TweenDriver) Done() bool {
	return d.done
}

// LoopDriverConfig configures a repeating time loop driven by external ticks.
type LoopDriverConfig struct {
	Duration  time.Duration
	Delay     time.Duration
	Cycles    int // 0 means infinite loop
	AutoStart bool
}

// LoopDriver tracks loop progress without owning a ticker.
type LoopDriver struct {
	duration       time.Duration
	delay          time.Duration
	delayRemaining time.Duration
	cycles         int
	running        bool
	done           bool
	started        bool
	lastTick       time.Time
	elapsed        time.Duration
	completed      int
}

// NewLoopDriver creates a new externally-driven loop.
func NewLoopDriver(cfg LoopDriverConfig) *LoopDriver {
	if cfg.Duration <= 0 {
		cfg.Duration = time.Millisecond
	}

	driver := &LoopDriver{
		duration: cfg.Duration,
		delay:    cfg.Delay,
		cycles:   cfg.Cycles,
	}
	driver.Reset()
	if cfg.AutoStart {
		driver.Start()
	}
	return driver
}

// Start restarts the loop from the beginning.
func (l *LoopDriver) Start() {
	l.running = true
	l.done = false
	l.started = l.delay <= 0
	l.delayRemaining = l.delay
	l.lastTick = time.Time{}
	l.elapsed = 0
	l.completed = 0
}

// Prime anchors the loop's time origin without advancing progress.
func (l *LoopDriver) Prime(now time.Time) {
	l.lastTick = now
}

// Primed reports whether the loop has a time anchor.
func (l *LoopDriver) Primed() bool {
	return !l.lastTick.IsZero()
}

// Stop stops the loop without clearing current progress.
func (l *LoopDriver) Stop() {
	l.running = false
}

// Reset clears loop progress and stops it.
func (l *LoopDriver) Reset() {
	l.running = false
	l.done = false
	l.started = l.delay <= 0
	l.delayRemaining = l.delay
	l.lastTick = time.Time{}
	l.elapsed = 0
	l.completed = 0
}

// WantsTick reports whether the loop still needs time updates.
func (l *LoopDriver) WantsTick() bool {
	return l.running && !l.done
}

// Tick advances the loop to the supplied time.
// It returns true when loop progress or visibility changed.
func (l *LoopDriver) Tick(now time.Time) bool {
	if !l.WantsTick() {
		return false
	}

	if l.lastTick.IsZero() {
		l.lastTick = now
		return false
	}

	delta := now.Sub(l.lastTick)
	if delta < 0 {
		delta = 0
	}
	l.lastTick = now

	changed := false
	if l.delayRemaining > 0 {
		if delta < l.delayRemaining {
			l.delayRemaining -= delta
			return false
		}
		delta -= l.delayRemaining
		l.delayRemaining = 0
		l.started = true
		changed = true
	}

	if delta == 0 {
		return changed
	}

	prevElapsed := l.elapsed
	prevCompleted := l.completed

	for delta > 0 && l.running {
		remaining := l.duration - l.elapsed
		if remaining <= 0 {
			remaining = l.duration
		}

		if delta < remaining {
			l.elapsed += delta
			delta = 0
			break
		}

		delta -= remaining
		l.completed++
		if l.cycles > 0 && l.completed >= l.cycles {
			l.elapsed = l.duration
			l.running = false
			l.done = true
			break
		}
		l.elapsed = 0
	}

	return changed || l.elapsed != prevElapsed || l.completed != prevCompleted
}

// Progress returns current in-cycle progress in the range [0, 1].
func (l *LoopDriver) Progress() float64 {
	if l.duration <= 0 {
		return 1
	}
	if l.done {
		return 1
	}
	progress := float64(l.elapsed) / float64(l.duration)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// StepIndex converts loop progress into a discrete frame index.
func (l *LoopDriver) StepIndex(steps int) int {
	if steps <= 1 {
		return 0
	}
	index := int(l.Progress() * float64(steps))
	if index >= steps {
		index = steps - 1
	}
	if index < 0 {
		index = 0
	}
	return index
}

// Cycle returns the number of completed loop cycles.
func (l *LoopDriver) Cycle() int {
	return l.completed
}

// Started reports whether the loop delay has elapsed.
func (l *LoopDriver) Started() bool {
	return l.started
}

// Done reports whether the loop exhausted its configured cycles.
func (l *LoopDriver) Done() bool {
	return l.done
}
