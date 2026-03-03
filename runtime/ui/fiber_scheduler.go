// Package ui provides FiberScheduler for priority-based Fiber rendering.
package ui

import (
	"time"

	"github.com/wwsheng009/mint/runtime/scheduler"
)

// =============================================================================
// FiberScheduler - Priority-Based Fiber Rendering
// =============================================================================

// FiberScheduler integrates the Lane scheduling system with Fiber rendering.
// It enables:
//   - Priority-based update scheduling
//   - Interruptible rendering for low-priority work
//   - High-priority updates can interrupt ongoing work
type FiberScheduler struct {
	// Core scheduler
	scheduler *scheduler.Scheduler

	// Fiber tree references
	rootFiber      *Fiber
	workInProgress *Fiber

	// Callbacks
	onCommit      func()
	onBeginWork   func(fiber *Fiber)
	onCompleteWork func(fiber *Fiber)

	// State
	isRendering bool
}

// FiberSchedulerOption configures the scheduler.
type FiberSchedulerOption func(*FiberScheduler)

// WithOnCommit sets the commit callback.
func WithOnCommit(fn func()) FiberSchedulerOption {
	return func(fs *FiberScheduler) {
		fs.onCommit = fn
	}
}

// WithOnBeginWork sets the beginWork callback.
func WithOnBeginWork(fn func(fiber *Fiber)) FiberSchedulerOption {
	return func(fs *FiberScheduler) {
		fs.onBeginWork = fn
	}
}

// WithOnCompleteWork sets the completeWork callback.
func WithOnCompleteWork(fn func(fiber *Fiber)) FiberSchedulerOption {
	return func(fs *FiberScheduler) {
		fs.onCompleteWork = fn
	}
}

// NewFiberScheduler creates a new FiberScheduler.
func NewFiberScheduler(opts ...FiberSchedulerOption) *FiberScheduler {
	fs := &FiberScheduler{
		scheduler: scheduler.NewScheduler(),
	}

	for _, opt := range opts {
		opt(fs)
	}

	return fs
}

// Shutdown stops the scheduler.
func (fs *FiberScheduler) Shutdown() {
	fs.scheduler.Shutdown()
}

// =============================================================================
// Scheduling Methods
// =============================================================================

// SetRoot sets the root fiber for rendering.
func (fs *FiberScheduler) SetRoot(root *Fiber) {
	fs.rootFiber = root
}

// ScheduleUpdate schedules a Fiber update with a specific priority.
//
// Example:
//
//	fs.ScheduleUpdate(fiber, scheduler.InputLane)
func (fs *FiberScheduler) ScheduleUpdate(fiber *Fiber, lane scheduler.Lane) {
	fs.scheduler.ScheduleFunc(lane, func(shouldYield scheduler.ShouldYieldFunc) bool {
		return fs.performUnitOfWork(fiber, shouldYield)
	})
}

// ScheduleSyncUpdate schedules a synchronous (immediate) update.
func (fs *FiberScheduler) ScheduleSyncUpdate(fiber *Fiber) {
	fs.ScheduleUpdate(fiber, scheduler.SyncLane)
}

// ScheduleInputUpdate schedules a high-priority input update.
func (fs *FiberScheduler) ScheduleInputUpdate(fiber *Fiber) {
	fs.ScheduleUpdate(fiber, scheduler.InputLane)
}

// ScheduleTransitionUpdate schedules a low-priority transition update.
func (fs *FiberScheduler) ScheduleTransitionUpdate(fiber *Fiber) {
	fs.ScheduleUpdate(fiber, scheduler.TransitionLane)
}

// ScheduleIdleUpdate schedules a background idle update.
func (fs *FiberScheduler) ScheduleIdleUpdate(fiber *Fiber) {
	fs.ScheduleUpdate(fiber, scheduler.IdleLane)
}

// =============================================================================
// Work Execution
// =============================================================================

// Flush executes all pending work synchronously.
func (fs *FiberScheduler) Flush() {
	fs.scheduler.Flush()
	if fs.onCommit != nil {
		fs.onCommit()
	}
}

// PerformWork processes one unit of work.
func (fs *FiberScheduler) PerformWork() {
	fs.scheduler.PerformWork()
}

// HasPendingWork returns true if there's pending work.
func (fs *FiberScheduler) HasPendingWork() bool {
	return fs.scheduler.HasPendingWork()
}

// =============================================================================
// Work Loop
// =============================================================================

// performUnitOfWork executes work on a single Fiber node.
// Returns true if complete, false if interrupted.
func (fs *FiberScheduler) performUnitOfWork(fiber *Fiber, shouldYield scheduler.ShouldYieldFunc) bool {
	if fiber == nil {
		return true
	}

	// 1. Begin work on this fiber
	next := fs.beginWork(fiber)

	// 2. Check for interruption (only for interruptible lanes)
	if shouldYield() && fs.isInterruptible(fiber) {
		fs.workInProgress = fiber
		return false
	}

	// 3. If no child, complete this fiber
	if next == nil {
		fs.completeWork(fiber)

		// Try sibling
		if fiber.Sibling != nil {
			return fs.performUnitOfWork(fiber.Sibling, shouldYield)
		}

		// Return to parent
		if fiber.Return != nil {
			return fs.completeUnitOfWork(fiber.Return, shouldYield)
		}

		return true
	}

	// 4. Continue with child
	return fs.performUnitOfWork(next, shouldYield)
}

// completeUnitOfWork completes work on a parent fiber after children are done.
func (fs *FiberScheduler) completeUnitOfWork(fiber *Fiber, shouldYield scheduler.ShouldYieldFunc) bool {
	fs.completeWork(fiber)

	// Check for interruption
	if shouldYield() && fs.isInterruptible(fiber) {
		fs.workInProgress = fiber
		return false
	}

	// Try sibling
	if fiber.Sibling != nil {
		return fs.performUnitOfWork(fiber.Sibling, shouldYield)
	}

	// Return to parent
	if fiber.Return != nil {
		return fs.completeUnitOfWork(fiber.Return, shouldYield)
	}

	return true
}

// beginWork processes a Fiber node and returns the next Fiber to work on.
func (fs *FiberScheduler) beginWork(fiber *Fiber) *Fiber {
	if fs.onBeginWork != nil {
		fs.onBeginWork(fiber)
	}

	// The actual beginWork logic is in the reconciler
	// This is a simplified version for scheduling purposes

	// Return first child to process depth-first
	if fiber.Child != nil {
		return fiber.Child
	}

	return nil
}

// completeWork finalizes a Fiber node after its children are processed.
func (fs *FiberScheduler) completeWork(fiber *Fiber) {
	if fs.onCompleteWork != nil {
		fs.onCompleteWork(fiber)
	}

	// Mark as updated
	fiber.Flags |= EffectUpdate
}

// isInterruptible checks if a Fiber's work can be interrupted.
func (fs *FiberScheduler) isInterruptible(fiber *Fiber) bool {
	// SyncLane is never interruptible
	if fiber.Lanes == Lane(scheduler.SyncLane) {
		return false
	}
	return true
}

// =============================================================================
// Helper Methods
// =============================================================================

// GetPendingLanes returns the set of lanes with pending work.
func (fs *FiberScheduler) GetPendingLanes() scheduler.Lane {
	return fs.scheduler.GetPendingLanes()
}

// GetQueueLength returns the number of pending tasks for a lane.
func (fs *FiberScheduler) GetQueueLength(lane scheduler.Lane) int {
	return fs.scheduler.GetQueueLength(lane)
}

// GetCurrentWork returns the current work-in-progress fiber.
func (fs *FiberScheduler) GetCurrentWork() *Fiber {
	return fs.workInProgress
}

// IsRendering returns true if rendering is in progress.
func (fs *FiberScheduler) IsRendering() bool {
	return fs.isRendering
}

// =============================================================================
// Lane Utilities
// =============================================================================

// LaneToScheduler converts a Fiber Lane to scheduler.Lane.
func LaneToScheduler(lane Lane) scheduler.Lane {
	// Map Fiber Lane values to scheduler Lane values
	switch lane {
	case LaneSyncLane:
		return scheduler.SyncLane
	case LaneInputContinuousLane:
		return scheduler.InputLane
	case LaneDefaultLane:
		return scheduler.DefaultLane
	case LaneIdleLane:
		return scheduler.IdleLane
	default:
		return scheduler.DefaultLane
	}
}

// SchedulerToFiber converts a scheduler.Lane to Fiber Lane.
func SchedulerToFiber(lane scheduler.Lane) Lane {
	// Map scheduler Lane values to Fiber Lane values
	switch lane {
	case scheduler.SyncLane:
		return LaneSyncLane
	case scheduler.InputLane:
		return LaneInputContinuousLane
	case scheduler.DefaultLane:
		return LaneDefaultLane
	case scheduler.TransitionLane:
		return LaneDefaultLane // No direct equivalent, use Default
	case scheduler.IdleLane:
		return LaneIdleLane
	default:
		return LaneDefaultLane
	}
}

// =============================================================================
// Deadline Management
// =============================================================================

// shouldYieldForLane creates a should-yield function for a specific lane.
func (fs *FiberScheduler) shouldYieldForLane(lane scheduler.Lane) scheduler.ShouldYieldFunc {
	startTime := time.Now()
	deadline := scheduler.GetDeadline(lane)

	return func() bool {
		// Check time budget
		if time.Since(startTime) >= deadline {
			return true
		}

		// Check for higher priority work
		pending := fs.scheduler.GetPendingLanes()
		highestPending := scheduler.PickHighestPriorityLane(pending)

		return lane.IsLowerPriorityThan(highestPending)
	}
}
