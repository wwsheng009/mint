package scheduler

// =============================================================================
// Scheduler Adapter - 适配 runtime/scheduler 到声明式 UI
// =============================================================================

import (
	"sync"
	"time"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/priority"
	"github.com/wwsheng009/mint/runtime/scheduler"
)

// UIScheduler wraps the runtime scheduler for declarative UI use
type UIScheduler struct {
	scheduler *scheduler.Scheduler
	mu        sync.RWMutex
}

// NewUIScheduler creates a new UI scheduler
func NewUIScheduler() *UIScheduler {
	return &UIScheduler{
		scheduler: scheduler.New(),
	}
}

// NewUISchedulerWithBudget creates a scheduler with custom time budget
func NewUISchedulerWithBudget(budget time.Duration) *UIScheduler {
	return &UIScheduler{
		scheduler: scheduler.NewWithBudget(budget),
	}
}

// =============================================================================
// Fiber 更新调度
// =============================================================================

// ScheduleUpdate schedules a fiber for update
func (s *UIScheduler) ScheduleUpdate(fiber *reconciler.Fiber, lane reconciler.Lane) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert Lane to priority.DirtyLevel
	prio := laneToPriority(lane)

	// Get fiber ID for tracking
	id := getFiberID(fiber)

	// Mark dirty in runtime scheduler
	s.scheduler.MarkDirty(id, fiber, prio)
}

// ScheduleFiberTree schedules an entire fiber tree for update
func (s *UIScheduler) ScheduleFiberTree(root *reconciler.Fiber, lane reconciler.Lane) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Begin batch mode
	s.scheduler.BeginBatch()
	defer s.scheduler.EndBatch(true)

	// Walk the tree and mark all fibers as dirty
	reconciler.WalkFiberDepthFirst(root, func(fiber *reconciler.Fiber) bool {
		if fiber.HasEffect() || fiber.HasSubtreeEffect() {
			id := getFiberID(fiber)
			prio := laneToPriority(lane)
			s.scheduler.MarkDirty(id, fiber, prio)
		}
		return true
	})
}

// ProcessFrame processes one frame of updates
func (s *UIScheduler) ProcessFrame(renderer FiberRenderer) ProcessResult {
	s.mu.RLock()
	sched := s.scheduler
	s.mu.RUnlock()

	if sched == nil {
		return ProcessResult{}
	}

	// Create runtime renderer adapter
	adapter := &fiberRendererAdapter{renderer: renderer}

	// Process updates
	runtimeResult := sched.ProcessNext(adapter, scheduler.DefaultProcessOptions())

	// Convert runtime result to our result type
	return ProcessResult{
		Processed: runtimeResult.Processed,
		Remaining: runtimeResult.Remaining,
		OutOfTime: runtimeResult.OutOfTime,
		ByPriority: runtimeResult.ByPriority,
	}
}

// =============================================================================
// 批处理支持
// =============================================================================

// BeginBatch starts batch mode - updates are cached until EndBatch
func (s *UIScheduler) BeginBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler.BeginBatch()
}

// EndBatch ends batch mode and optionally flushes
func (s *UIScheduler) EndBatch(flush bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler.EndBatch(flush)
}

// FlushBatch flushes pending updates to dirty queue
func (s *UIScheduler) FlushBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler.FlushBatch()
}

// =============================================================================
// 状态查询
// =============================================================================

// DirtyCount returns the count of dirty nodes by priority level
func (s *UIScheduler) DirtyCount() map[priority.DirtyLevel]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scheduler.DirtyCount()
}

// TotalDirtyCount returns total dirty node count
func (s *UIScheduler) TotalDirtyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scheduler.TotalDirtyCount()
}

// IsBatching returns true if currently batching
func (s *UIScheduler) IsBatching() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scheduler.IsBatching()
}

// Clear clears all dirty nodes
func (s *UIScheduler) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler.Clear()
}

// =============================================================================
// 辅助函数
// =============================================================================

// laneToPriority converts a Lane to priority.DirtyLevel
func laneToPriority(lane reconciler.Lane) priority.DirtyLevel {
	// Map lanes to priority levels
	if lane&reconciler.LaneSyncLane != 0 {
		return priority.DirtyHigh
	}
	if lane&reconciler.LaneInputContinuousLane != 0 {
		return priority.DirtyHigh
	}
	if lane&reconciler.LaneDefaultLane != 0 {
		return priority.DirtyNormal
	}
	// LaneIdleLane or no lane
	return priority.DirtyLow
}

// getFiberID generates a unique ID for a fiber
func getFiberID(fiber *reconciler.Fiber) string {
	if fiber.Key != "" {
		return "fiber:" + fiber.Key
	}
	return "fiber:" + fiber.Tag + ":" + fiber.Type.String()
}

// =============================================================================
// 渲染器适配
// =============================================================================

// FiberRenderer is the interface for rendering fibers
type FiberRenderer interface {
	Layout(fiber *reconciler.Fiber)
	Paint(fiber *reconciler.Fiber)
}

// ProcessResult represents the result of processing a frame
type ProcessResult struct {
	Processed  int
	Remaining  int
	OutOfTime  bool
	ByPriority map[priority.DirtyLevel]int
}

// fiberRendererAdapter adapts FiberRenderer to runtime scheduler.Renderer
type fiberRendererAdapter struct {
	renderer FiberRenderer
}

// Layout implements scheduler.Renderer
func (a *fiberRendererAdapter) Layout(node interface{}) {
	if fiber, ok := node.(*reconciler.Fiber); ok {
		a.renderer.Layout(fiber)
	}
}

// Paint implements scheduler.Renderer
func (a *fiberRendererAdapter) Paint(node interface{}) {
	if fiber, ok := node.(*reconciler.Fiber); ok {
		a.renderer.Paint(fiber)
	}
}

// =============================================================================
// 全局调度器实例
// =============================================================================

var (
	globalScheduler *UIScheduler
	schedulerMu    sync.RWMutex
)

// SetGlobalScheduler sets the global scheduler instance
func SetGlobalScheduler(s *UIScheduler) {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	globalScheduler = s
}

// GetGlobalScheduler returns the global scheduler instance
func GetGlobalScheduler() *UIScheduler {
	schedulerMu.RLock()
	defer schedulerMu.RUnlock()
	if globalScheduler == nil {
		globalScheduler = NewUIScheduler()
	}
	return globalScheduler
}

// =============================================================================
// Fiber 工作循环
// =============================================================================

// WorkLoop represents the main work loop for fiber reconciliation
type WorkLoop struct {
	scheduler *UIScheduler
	renderer  FiberRenderer
	rootFiber *reconciler.Fiber
	running   bool
	mu        sync.RWMutex
}

// NewWorkLoop creates a new work loop
func NewWorkLoop(renderer FiberRenderer) *WorkLoop {
	return &WorkLoop{
		scheduler: NewUIScheduler(),
		renderer:  renderer,
		running:   false,
	}
}

// Start starts the work loop
func (w *WorkLoop) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = true
}

// Stop stops the work loop
func (w *WorkLoop) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false
}

// IsRunning returns true if the work loop is running
func (w *WorkLoop) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// Schedule schedules a fiber for update
func (w *WorkLoop) Schedule(fiber *reconciler.Fiber, lane reconciler.Lane) {
	w.scheduler.ScheduleUpdate(fiber, lane)
}

// ProcessFrame processes one frame of updates
func (w *WorkLoop) ProcessFrame() ProcessResult {
	w.mu.RLock()
	running := w.running
	w.mu.RUnlock()

	if !running {
		return ProcessResult{}
	}

	return w.scheduler.ProcessFrame(w.renderer)
}

// SetRoot sets the root fiber
func (w *WorkLoop) SetRoot(fiber *reconciler.Fiber) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.rootFiber = fiber
}

// GetRoot returns the root fiber
func (w *WorkLoop) GetRoot() *reconciler.Fiber {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.rootFiber
}

// Invalidate marks the entire tree for update
func (w *WorkLoop) Invalidate(lane reconciler.Lane) {
	w.mu.RLock()
	root := w.rootFiber
	w.mu.RUnlock()

	if root != nil {
		w.scheduler.ScheduleFiberTree(root, lane)
	}
}

// =============================================================================
// 渲染器实现（用于框架集成）
// =============================================================================

// DefaultFiberRenderer is a default renderer implementation
type DefaultFiberRenderer struct {
	onLayout func(*reconciler.Fiber)
	onPaint  func(*reconciler.Fiber)
}

// Layout implements FiberRenderer
func (r *DefaultFiberRenderer) Layout(fiber *reconciler.Fiber) {
	if r.onLayout != nil {
		r.onLayout(fiber)
	}
}

// Paint implements FiberRenderer
func (r *DefaultFiberRenderer) Paint(fiber *reconciler.Fiber) {
	if r.onPaint != nil {
		r.onPaint(fiber)
	}
}

// =============================================================================
// 时间切片支持
// =============================================================================

// TimeSlice provides time-sliced execution context
type TimeSlice struct {
	budget    time.Duration
	deadline  time.Time
	remaining time.Duration
}

// NewTimeSlice creates a new time slice
func NewTimeSlice(budget time.Duration) *TimeSlice {
	now := time.Now()
	return &TimeSlice{
		budget:    budget,
		deadline:  now.Add(budget),
		remaining: budget,
	}
}

// ShouldContinue returns true if there's time remaining
func (ts *TimeSlice) ShouldContinue() bool {
	ts.remaining = ts.deadline.Sub(time.Now())
	return ts.remaining > 0
}

// Elapsed returns elapsed time
func (ts *TimeSlice) Elapsed() time.Duration {
	return ts.budget - ts.remaining
}

// =============================================================================
// 渲染策略
// =============================================================================

// RenderStrategy defines when to render
type RenderStrategy int

const (
	// StrategyAlways always renders
	StrategyAlways RenderStrategy = iota
	// StrategyOnDirty only renders when dirty
	StrategyOnDirty
	// StrategyAuto adaptive rendering
	StrategyAuto
)

// RenderController controls rendering strategy
type RenderController struct {
	scheduler *UIScheduler
	strategy  RenderStrategy
	throttler *ThrottlerAdapter
	mu        sync.RWMutex
}

// NewRenderController creates a render controller
func NewRenderController() *RenderController {
	return &RenderController{
		scheduler: NewUIScheduler(),
		strategy:  StrategyOnDirty,
		throttler: NewThrottlerAdapter(60),
	}
}

// ShouldRender decides if rendering should happen
func (c *RenderController) ShouldRender() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check throttler first
	if !c.throttler.ShouldRender() {
		return false
	}

	// Check dirty state based on strategy
	switch c.strategy {
	case StrategyAlways:
		return true
	case StrategyOnDirty:
		return c.scheduler.TotalDirtyCount() > 0
	case StrategyAuto:
		// Adaptive: if many pending, allow higher FPS
		count := c.scheduler.TotalDirtyCount()
		if count > 10 {
			return true
		}
		return count > 0
	default:
		return true
	}
}

// SetStrategy sets the render strategy
func (c *RenderController) SetStrategy(s RenderStrategy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.strategy = s
}

// SetTargetFPS sets the target FPS
func (c *RenderController) SetTargetFPS(fps int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.throttler.SetFPS(fps)
}

// ProcessFrame processes a frame
func (c *RenderController) ProcessFrame(renderer FiberRenderer) ProcessResult {
	c.mu.RLock()
	sched := c.scheduler
	c.mu.RUnlock()

	if sched == nil {
		return ProcessResult{}
	}

	// Use UIScheduler.ProcessFrame which handles the adapter internally
	return sched.ProcessFrame(renderer)
}

// =============================================================================
// Throttler Adapter
// =============================================================================

// ThrottlerAdapter wraps runtime/render.Throttler
// Note: Since runtime/render uses the same package, we create a lightweight wrapper
type ThrottlerAdapter struct {
	targetFPS   int
	minInterval time.Duration
	lastRender  time.Time
	mu          sync.Mutex
}

// NewThrottlerAdapter creates a new throttler adapter
func NewThrottlerAdapter(fps int) *ThrottlerAdapter {
	if fps <= 0 {
		fps = 60
	}
	interval := time.Second / time.Duration(fps)
	return &ThrottlerAdapter{
		targetFPS:   fps,
		minInterval: interval,
		lastRender:  time.Time{}, // Zero time allows first render
	}
}

// ShouldRender checks if rendering should happen
func (t *ThrottlerAdapter) ShouldRender() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastRender)
	if elapsed < t.minInterval {
		return false
	}
	t.lastRender = now
	return true
}

// SetFPS sets the target FPS
func (t *ThrottlerAdapter) SetFPS(fps int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if fps < 1 {
		fps = 1
	}
	if fps > 120 {
		fps = 120
	}
	t.targetFPS = fps
	t.minInterval = time.Second / time.Duration(fps)
}

// ForceRender forces next render
func (t *ThrottlerAdapter) ForceRender() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRender = time.Time{} // Reset to allow immediate render
}

// FPS returns current target FPS
func (t *ThrottlerAdapter) FPS() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.targetFPS
}
