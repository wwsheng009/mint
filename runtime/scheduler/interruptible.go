// Package scheduler provides interruptible task execution.
package scheduler

import (
	"context"
	"sync"
	"time"
)

// TaskState represents the current state of a task.
type TaskState int

const (
	// TaskPending is the initial state before execution.
	TaskPending TaskState = iota
	// TaskRunning indicates the task is currently executing.
	TaskRunning
	// TaskPaused indicates the task was paused.
	TaskPaused
	// TaskCompleted indicates the task finished successfully.
	TaskCompleted
	// TaskCancelled indicates the task was cancelled.
	TaskCancelled
	// TaskErrored indicates the task encountered an error.
	TaskErrored
)

// String returns the string representation of the task state.
func (s TaskState) String() string {
	switch s {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskPaused:
		return "paused"
	case TaskCompleted:
		return "completed"
	case TaskCancelled:
		return "cancelled"
	case TaskErrored:
		return "errored"
	default:
		return "unknown"
	}
}

// TaskFunc is the function executed by the interruptible task.
// The done channel signals cancellation. Return true if completed,
// false if interrupted and should be resumed later.
type TaskFunc func(done <-chan struct{}) bool

// InterruptibleTask represents a task that can be paused, cancelled, and resumed.
type InterruptibleTask struct {
	mu       sync.RWMutex
	id       string
	fn       TaskFunc
	state    TaskState
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	result   interface{}
	err      error
	progress float64 // 0.0 to 1.0

	// Callbacks
	onStateChange func(TaskState, TaskState)
	onProgress   func(float64)
	onComplete   func(interface{}, error)
}

// NewInterruptibleTask creates a new interruptible task.
func NewInterruptibleTask(id string, fn TaskFunc) *InterruptibleTask {
	ctx, cancel := context.WithCancel(context.Background())

	return &InterruptibleTask{
		id:     id,
		fn:     fn,
		state:  TaskPending,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

// ID returns the task's unique identifier.
func (t *InterruptibleTask) ID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

// State returns the current task state.
func (t *InterruptibleTask) State() TaskState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// Progress returns the current progress (0.0 to 1.0).
func (t *InterruptibleTask) Progress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.progress
}

// Result returns the task result. Returns nil if not completed.
func (t *InterruptibleTask) Result() (interface{}, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result, t.err
}

// Execute runs the task to completion or interruption.
// Returns true if the task completed, false if it was interrupted.
func (t *InterruptibleTask) Execute() bool {
	t.mu.Lock()

	// Check if already in a terminal state
	if t.state == TaskCompleted || t.state == TaskCancelled || t.state == TaskErrored {
		t.mu.Unlock()
		return t.state == TaskCompleted
	}

	// Check if cancelled
	select {
	case <-t.ctx.Done():
		t.setState(TaskCancelled)
		t.mu.Unlock()
		return false
	default:
	}

	t.setState(TaskRunning)
	t.mu.Unlock()

	// Execute the task function
	completed := t.fn(t.done)

	t.mu.Lock()
	defer t.mu.Unlock()

	if completed {
		t.setState(TaskCompleted)
	} else {
		// Task was interrupted, will be resumed
		t.setState(TaskPaused)
	}

	return completed
}

// Resume continues a paused task.
func (t *InterruptibleTask) Resume() bool {
	t.mu.Lock()

	if t.state != TaskPaused {
		t.mu.Unlock()
		return false
	}

	// Create new done channel
	close(t.done)
	t.done = make(chan struct{})

	t.mu.Unlock()

	return t.Execute()
}

// Pause attempts to pause the running task.
// Returns true if the task was successfully paused.
func (t *InterruptibleTask) Pause() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != TaskRunning {
		return false
	}

	// Signal cancellation
	t.cancel()

	return true
}

// Cancel stops the task and marks it as cancelled.
func (t *InterruptibleTask) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == TaskCompleted || t.state == TaskCancelled || t.state == TaskErrored {
		return
	}

	t.cancel()
	close(t.done)
	t.setState(TaskCancelled)
}

// IsRunning returns true if the task is currently running.
func (t *InterruptibleTask) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == TaskRunning
}

// IsCompleted returns true if the task completed successfully.
func (t *InterruptibleTask) IsCompleted() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == TaskCompleted
}

// IsCancelled returns true if the task was cancelled.
func (t *InterruptibleTask) IsCancelled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == TaskCancelled
}

// SetProgress updates the task's progress.
func (t *InterruptibleTask) SetProgress(p float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}

	t.progress = p

	if t.onProgress != nil {
		t.onProgress(p)
	}
}

// SetResult sets the task's result.
func (t *InterruptibleTask) SetResult(result interface{}, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.result = result
	t.err = err

	if err != nil {
		t.setState(TaskErrored)
	}

	if t.onComplete != nil {
		t.onComplete(result, err)
	}
}

// OnStateChange sets a callback for state changes.
func (t *InterruptibleTask) OnStateChange(fn func(TaskState, TaskState)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onStateChange = fn
}

// OnProgress sets a callback for progress updates.
func (t *InterruptibleTask) OnProgress(fn func(float64)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onProgress = fn
}

// OnComplete sets a callback for task completion.
func (t *InterruptibleTask) OnComplete(fn func(interface{}, error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = fn
}

// setState updates the state and triggers callback.
func (t *InterruptibleTask) setState(newState TaskState) {
	oldState := t.state
	t.state = newState

	if t.onStateChange != nil && oldState != newState {
		// Call outside the lock to avoid deadlock
		go t.onStateChange(oldState, newState)
	}
}

// Done returns the done channel for cancellation signaling.
func (t *InterruptibleTask) Done() <-chan struct{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.done
}

// =============================================================================
// Task Manager
// =============================================================================

// TaskManager manages multiple interruptible tasks.
type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*InterruptibleTask
}

// NewTaskManager creates a new task manager.
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*InterruptibleTask),
	}
}

// Add adds a task to the manager.
func (tm *TaskManager) Add(task *InterruptibleTask) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks[task.ID()] = task
}

// Remove removes a task from the manager.
func (tm *TaskManager) Remove(id string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tasks, id)
}

// Get returns a task by ID.
func (tm *TaskManager) Get(id string) (*InterruptibleTask, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[id]
	return task, ok
}

// CancelAll cancels all managed tasks.
func (tm *TaskManager) CancelAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, task := range tm.tasks {
		task.Cancel()
	}
}

// PauseAll pauses all running tasks.
func (tm *TaskManager) PauseAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, task := range tm.tasks {
		if task.IsRunning() {
			task.Pause()
		}
	}
}

// RunningCount returns the number of running tasks.
func (tm *TaskManager) RunningCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	count := 0
	for _, task := range tm.tasks {
		if task.IsRunning() {
			count++
		}
	}
	return count
}

// CompletedCount returns the number of completed tasks.
func (tm *TaskManager) CompletedCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	count := 0
	for _, task := range tm.tasks {
		if task.IsCompleted() {
			count++
		}
	}
	return count
}

// Clear removes all tasks (cancelling any running tasks).
func (tm *TaskManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, task := range tm.tasks {
		task.Cancel()
	}

	tm.tasks = make(map[string]*InterruptibleTask)
}

// List returns all tasks.
func (tm *TaskManager) List() []*InterruptibleTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*InterruptibleTask, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		result = append(result, task)
	}

	return result
}

// =============================================================================
// Helper Functions
// =============================================================================

// RunWithTimeout runs a task with a timeout.
func RunWithTimeout(fn func() (interface{}, error), timeout time.Duration) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultCh := make(chan resultPair)

	go func() {
		result, err := fn()
		resultCh <- resultPair{result, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case rp := <-resultCh:
		return rp.result, rp.err
	}
}

type resultPair struct {
	result interface{}
	err    error
}

// RunWithCancellation runs a task that can be cancelled.
func RunWithCancellation(fn func(<-chan struct{}) (interface{}, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resultCh := make(chan resultPair)

		go func() {
			result, err := fn(ctx.Done())
			resultCh <- resultPair{result, err}
		}()

		select {
		case rp := <-resultCh:
			return rp.result, rp.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
