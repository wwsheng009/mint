// Package scheduler provides the core scheduler implementation.
package scheduler

import (
	"context"
	"sync"
	"time"
)

// =============================================================================
// Work Definition
// =============================================================================

// Work represents a unit of work to be scheduled.
type Work interface {
	// Execute performs the work. Returns true if completed, false if interrupted.
	Execute(shouldYield ShouldYieldFunc) bool
}

// WorkFunc is an adapter to allow using functions as Work.
type WorkFunc func(shouldYield ShouldYieldFunc) bool

// Execute implements Work.
func (f WorkFunc) Execute(shouldYield ShouldYieldFunc) bool {
	return f(shouldYield)
}

// =============================================================================
// ScheduledTask
// =============================================================================

// ScheduledTask represents a task in the scheduler queue.
type ScheduledTask struct {
	ID        uint64
	Lane      Lane
	Work      Work
	CreatedAt time.Time
	Deadline  time.Time
	Canceled  bool
}

// Cancel marks the task as canceled.
func (t *ScheduledTask) Cancel() {
	t.Canceled = true
}

// IsExpired checks if the task has exceeded its deadline.
func (t *ScheduledTask) IsExpired() bool {
	return time.Now().After(t.Deadline)
}

// =============================================================================
// Scheduler
// =============================================================================

// Scheduler manages the execution of work based on priority.
type Scheduler struct {
	mu sync.Mutex

	// Task queues for each lane
	queues map[Lane][]*ScheduledTask

	// Task ID counter
	nextTaskID uint64

	// Current work in progress
	currentTask *ScheduledTask
	currentLane Lane

	// Scheduler state
	isPerformingWork bool
	pendingLanes     Lane

	// Callbacks
	onWorkStart   func(task *ScheduledTask)
	onWorkComplete func(task *ScheduledTask)
	onWorkYield   func(task *ScheduledTask)

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// SchedulerOption configures the scheduler.
type SchedulerOption func(*Scheduler)

// WithOnWorkStart sets the callback for when work starts.
func WithOnWorkStart(fn func(task *ScheduledTask)) SchedulerOption {
	return func(s *Scheduler) {
		s.onWorkStart = fn
	}
}

// WithOnWorkComplete sets the callback for when work completes.
func WithOnWorkComplete(fn func(task *ScheduledTask)) SchedulerOption {
	return func(s *Scheduler) {
		s.onWorkComplete = fn
	}
}

// WithOnWorkYield sets the callback for when work yields.
func WithOnWorkYield(fn func(task *ScheduledTask)) SchedulerOption {
	return func(s *Scheduler) {
		s.onWorkYield = fn
	}
}

// NewScheduler creates a new scheduler.
func NewScheduler(opts ...SchedulerOption) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		queues:     make(map[Lane][]*ScheduledTask),
		nextTaskID: 1,
		ctx:        ctx,
		cancel:     cancel,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Shutdown stops the scheduler.
func (s *Scheduler) Shutdown() {
	s.cancel()
}

// =============================================================================
// Scheduling Methods
// =============================================================================

// Schedule adds work to the scheduler queue with a given lane.
//
// Example:
//
//	task := scheduler.Schedule(InputLane, func(shouldYield ShouldYieldFunc) bool {
//	    // Process input...
//	    return true // completed
//	})
func (s *Scheduler) Schedule(lane Lane, work Work) *ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &ScheduledTask{
		ID:        s.nextTaskID,
		Lane:      lane,
		Work:      work,
		CreatedAt: time.Now(),
		Deadline:  time.Now().Add(GetDeadline(lane)),
	}
	s.nextTaskID++

	s.queues[lane] = append(s.queues[lane], task)
	s.pendingLanes |= lane

	return task
}

// ScheduleFunc adds a function to the scheduler queue.
func (s *Scheduler) ScheduleFunc(lane Lane, fn func(shouldYield ShouldYieldFunc) bool) *ScheduledTask {
	return s.Schedule(lane, WorkFunc(fn))
}

// Cancel removes a task from the queue.
func (s *Scheduler) Cancel(task *ScheduledTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.Cancel()
	s.removeFromQueue(task)
}

func (s *Scheduler) removeFromQueue(task *ScheduledTask) {
	queue := s.queues[task.Lane]
	for i, t := range queue {
		if t.ID == task.ID {
			s.queues[task.Lane] = append(queue[:i], queue[i+1:]...)
			break
		}
	}

	// Update pending lanes
	if len(s.queues[task.Lane]) == 0 {
		s.pendingLanes &^= task.Lane
	}
}

// =============================================================================
// Work Execution
// =============================================================================

// Flush performs all pending work synchronously.
// This is useful for testing or when immediate updates are needed.
func (s *Scheduler) Flush() {
	for s.HasPendingWork() {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		s.performWorkUntilDeadline()
	}
}

// PerformWork processes the highest priority work until yielding or completion.
func (s *Scheduler) PerformWork() {
	s.performWorkUntilDeadline()
}

// performWorkUntilDeadline processes work until there's no more work or we should yield.
func (s *Scheduler) performWorkUntilDeadline() {
	s.mu.Lock()
	s.isPerformingWork = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isPerformingWork = false
		s.mu.Unlock()
	}()

	for s.HasPendingWork() {
		// Check for context cancellation
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Get highest priority lane with work
		lane := PickHighestPriorityLane(s.pendingLanes)
		if lane == NoLane {
			return
		}

		// Get the next task
		task := s.getNextTask(lane)
		if task == nil {
			continue
		}

		s.currentTask = task
		s.currentLane = lane

		// Notify work start
		if s.onWorkStart != nil {
			s.onWorkStart(task)
		}

		// Execute the work
		startTime := time.Now()
		shouldYield := DefaultShouldYield(startTime, GetDeadline(lane))

		completed := task.Work.Execute(shouldYield)

		// Handle result
		s.mu.Lock()
		if completed || task.Canceled {
			s.removeFromQueue(task)
			if s.onWorkComplete != nil {
				s.onWorkComplete(task)
			}
		} else {
			// Work yielded, put it back in queue
			if s.onWorkYield != nil {
				s.onWorkYield(task)
			}
		}
		s.currentTask = nil
		s.currentLane = NoLane
		s.mu.Unlock()

		// Check if we should yield for higher priority work
		if !completed && !task.Canceled && shouldYield() {
			// Yield and let higher priority work run
			break
		}
	}
}

// getNextTask gets the next task from a lane's queue.
func (s *Scheduler) getNextTask(lane Lane) *ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	queue := s.queues[lane]
	if len(queue) == 0 {
		return nil
	}

	task := queue[0]
	if task.Canceled {
		s.removeFromQueue(task)
		return nil
	}

	return task
}

// =============================================================================
// Scheduler State
// =============================================================================

// HasPendingWork returns true if there's pending work.
func (s *Scheduler) HasPendingWork() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingLanes != NoLane
}

// GetPendingLanes returns the set of lanes with pending work.
func (s *Scheduler) GetPendingLanes() Lane {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingLanes
}

// IsPerformingWork returns true if work is being performed.
func (s *Scheduler) IsPerformingWork() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isPerformingWork
}

// GetQueueLength returns the number of tasks in a specific lane.
func (s *Scheduler) GetQueueLength(lane Lane) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queues[lane])
}

// GetTotalQueueLength returns the total number of pending tasks.
func (s *Scheduler) GetTotalQueueLength() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var total int
	for _, queue := range s.queues {
		total += len(queue)
	}
	return total
}

// CurrentTask returns the currently executing task.
func (s *Scheduler) CurrentTask() *ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentTask
}

// =============================================================================
// Lane Management
// =============================================================================

// MarkLaneReady marks a lane as having work to do.
func (s *Scheduler) MarkLaneReady(lane Lane) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingLanes |= lane
}

// MarkLaneComplete marks a lane as complete (no more work).
func (s *Scheduler) MarkLaneComplete(lane Lane) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingLanes &^= lane
}

// =============================================================================
// Batch Scheduling
// =============================================================================

// BatchScheduler helps schedule multiple tasks efficiently.
type BatchScheduler struct {
	scheduler *Scheduler
	lane      Lane
	tasks     []*ScheduledTask
}

// NewBatchScheduler creates a new batch scheduler.
func NewBatchScheduler(scheduler *Scheduler, lane Lane) *BatchScheduler {
	return &BatchScheduler{
		scheduler: scheduler,
		lane:      lane,
		tasks:     make([]*ScheduledTask, 0),
	}
}

// Add adds work to the batch.
func (b *BatchScheduler) Add(work Work) {
	task := b.scheduler.Schedule(b.lane, work)
	b.tasks = append(b.tasks, task)
}

// AddFunc adds a function to the batch.
func (b *BatchScheduler) AddFunc(fn func(shouldYield ShouldYieldFunc) bool) {
	b.Add(WorkFunc(fn))
}

// Flush executes all batch tasks.
func (b *BatchScheduler) Flush() {
	b.scheduler.Flush()
}

// Cancel cancels all tasks in the batch.
func (b *BatchScheduler) Cancel() {
	for _, task := range b.tasks {
		task.Cancel()
	}
	b.tasks = nil
}

// Count returns the number of tasks in the batch.
func (b *BatchScheduler) Count() int {
	return len(b.tasks)
}
