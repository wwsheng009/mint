package intent

import (
	"context"
	"fmt"
	"sync"
	"time"

	mintlog "github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/priority"
)

// =============================================================================
// Dispatcher
// =============================================================================

// Dispatcher handles intent dispatching with priority-based scheduling.
type Dispatcher struct {
	mu sync.RWMutex

	// registry is the intent registry
	registry *Registry

	// queue holds pending intents
	queue *intentQueue

	// scheduler is the callback for scheduling fiber updates
	scheduler ScheduleFunc

	// stateSetter is used by ActionContext
	stateSetter StateSetter

	// logger is the structured logger for intent dispatch events
	logger *mintlog.Logger

	// log enables debug logging (deprecated - use logger instead)
	log bool

	// logEntries stores dispatch history
	logEntries []DispatchLog

	// logMaxSize limits log entries
	logMaxSize int

	// errorStrategy defines how to handle intent errors
	errorStrategy ErrorHandlingStrategy

	// errorHandler is a custom error handler (when errorStrategy == ErrorCustomCallback)
	errorHandler func(intent Intent, err error)

	// maxRetry is the maximum number of retries for ErrorLogRetry strategy
	maxRetry int

	// useCustomLogger indicates if a custom logger was set by the user
	useCustomLogger bool
}

// ScheduleFunc is called to schedule a fiber update with the given lane.
type ScheduleFunc func(lane priority.DirtyLevel)

// DispatchLog records a dispatch event.
type DispatchLog struct {
	Intent    Intent
	Type      string
	Priority  ActionPriority
	Lane      priority.DirtyLevel
	Timestamp time.Time
	Duration  time.Duration
	Handled   bool
	Error     error
}

// ErrorHandlingStrategy defines how to handle intent errors.
type ErrorHandlingStrategy int

const (
	// ErrorLogIgnore logs the error and ignores it
	ErrorLogIgnore ErrorHandlingStrategy = iota
	// ErrorLogPanic logs the error and panics
	ErrorLogPanic
	// ErrorLogRetry logs the error and retries (not implemented)
	ErrorLogRetry
	// ErrorCustomCallback calls the custom error handler
	ErrorCustomCallback
)

// NewDispatcher creates a new intent dispatcher.
// By default, it uses the IntentLogger for logging.
func NewDispatcher(registry *Registry) *Dispatcher {
	return &Dispatcher{
		registry:    registry,
		queue:       newIntentQueue(),
		logMaxSize:  1000,
		logEntries:  make([]DispatchLog, 0),
		logger:      mintlog.IntentLogger, // Use dedicated IntentLogger by default
	}
}

// SetScheduler sets the scheduler callback.
func (d *Dispatcher) SetScheduler(fn ScheduleFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.scheduler = fn
}

// SetStateSetter sets the state setter for ActionContext.
func (d *Dispatcher) SetStateSetter(setter StateSetter) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stateSetter = setter
}

// SetLogger sets a custom structured logger for intent dispatch events.
// If not set, the dispatcher uses the default IntentLogger.
func (d *Dispatcher) SetLogger(logger *mintlog.Logger) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logger = logger
	d.useCustomLogger = true
}

// GetLogger returns the current logger being used by the dispatcher.
// Returns the IntentLogger by default, or a custom logger if SetLogger was called.
func (d *Dispatcher) GetLogger() *mintlog.Logger {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.logger
}

// EnableLog enables or disables debug logging.
func (d *Dispatcher) EnableLog(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log = enabled
}

// SetErrorStrategy sets the error handling strategy for intent dispatch failures.
func (d *Dispatcher) SetErrorStrategy(strategy ErrorHandlingStrategy) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.errorStrategy = strategy
}

// SetErrorHandler sets a custom error handler for when errorStrategy is ErrorCustomCallback.
func (d *Dispatcher) SetErrorHandler(handler func(intent Intent, err error)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.errorHandler = handler
}

// SetMaxRetry sets the maximum number of retries for ErrorLogRetry strategy.
func (d *Dispatcher) SetMaxRetry(maxRetry int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxRetry = maxRetry
}

// =============================================================================
// Dispatch Methods
// =============================================================================

// applyErrorStrategy applies the configured error handling strategy to a failed intent result.
func (d *Dispatcher) applyErrorStrategy(intent Intent, result IntentResult) {
	d.mu.RLock()
	strategy := d.errorStrategy
	handler := d.errorHandler
	d.mu.RUnlock()

	switch strategy {
	case ErrorLogIgnore:
		// Already logged, do nothing

	case ErrorLogPanic:
		panic(fmt.Sprintf("Intent dispatch failed: type=%s, error=%v", intent.IntentType(), result.Error))

	case ErrorCustomCallback:
		if handler != nil {
			handler(intent, result.Error)
		}

	case ErrorLogRetry:
		d.retryDispatch(intent, result)
	}
}

// retryDispatch implements retry logic for failed intents.
// It will retry the intent dispatch up to maxRetry times.
// This method directly invokes the handler to avoid infinite recursion through applyErrorStrategy.
func (d *Dispatcher) retryDispatch(intent Intent, lastResult IntentResult) {
	d.mu.RLock()
	maxRetries := d.maxRetry
	if maxRetries <= 0 {
		maxRetries = 3 // Default retry count
	}
	stateSetter := d.stateSetter
	scheduler := d.scheduler
	d.mu.RUnlock()

	intentType := intent.IntentType()

	// Get handler from registry
	handler, ok := d.registry.GetHandler(intentType)
	if !ok {
		// No handler to retry with
		if d.logger != nil && d.logger.Enabled() {
			d.logger.Error("Cannot retry intent type=%s: no handler registered", intentType)
		}
		return
	}

	if d.logger != nil && d.logger.Enabled() {
		d.logger.Warn("Starting retry for intent type=%s, max attempts=%d",
			intentType, maxRetries+1)
	}

	// Resolve priority for scheduling (only if needed)
	var lane priority.DirtyLevel
	if scheduler != nil {
		priority := d.registry.GetPriority(intent)
		lane = priority.ToLane()
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Small delay between retries (linear backoff: 50ms * attempt)
		delay := time.Duration(attempt*50) * time.Millisecond
		time.Sleep(delay)

		// Create context for this retry attempt
		ctx := NewActionContext(context.Background(), "retry", stateSetter)

		// Directly invoke handler to avoid infinite recursion through applyErrorStrategy
		result := handler.Handle(ctx, intent)

		if result.Error == nil {
			// Schedule fiber update if needed
			if result.Handled && scheduler != nil {
				scheduler(lane)
			}

			if d.logger != nil && d.logger.Enabled() {
				d.logger.Debug("Intent succeeded after retry: type=%s, attempt=%d/%d",
					intentType, attempt+1, maxRetries+1)
			}
			return // Success, stop retrying
		}

		if d.logger != nil && d.logger.Enabled() {
			d.logger.Warn("Retry attempt %d failed for intent type=%s: %v",
				attempt, intentType, result.Error)
		}
	}

	// All retries exhausted
	if d.logger != nil && d.logger.Enabled() {
		d.logger.Error("All retry attempts exhausted for intent type=%s, original error: %v",
			intentType, lastResult.Error)
	}
}

// Dispatch dispatches an intent with automatic priority resolution.
func (d *Dispatcher) Dispatch(intent Intent) IntentResult {
	return d.DispatchWithSource(intent, "")
}

// DispatchWithSource dispatches an intent with a source identifier.
func (d *Dispatcher) DispatchWithSource(intent Intent, source string) IntentResult {
	priority := d.registry.GetPriority(intent)
	return d.DispatchWithPriority(intent, source, priority)
}

// DispatchWithPriority dispatches an intent with explicit priority.
func (d *Dispatcher) DispatchWithPriority(intent Intent, source string, p ActionPriority) IntentResult {
	start := time.Now()
	intentType := intent.IntentType()
	lane := p.ToLane()

	// Log dispatch start
	if d.logger != nil && d.logger.Enabled() {
		d.logger.Debug("Dispatching intent: type=%s, source=%s, priority=%s, lane=%s",
			intentType, source, p, lane)
	}

	// Create action context
	ctx := NewActionContext(context.Background(), source, d.stateSetter)

	// Log dispatch start (backward compatibility)
	var logEntry DispatchLog
	if d.log {
		logEntry = DispatchLog{
			Intent:    intent,
			Type:      intentType,
			Priority:  p,
			Lane:      lane,
			Timestamp: start,
		}
	}

	// Get handler
	handler, ok := d.registry.GetHandler(intentType)
	if !ok {
		result := ErrorResult(fmt.Errorf("no handler registered for intent type: %s", intentType))

		// Log error using structured logger
		if d.logger != nil {
			if result.Error != nil {
				d.logger.Error("No handler for intent type=%s: %v", intentType, result.Error)
			} else {
				d.logger.Warn("No handler registered for intent type=%s", intentType)
			}
		}

		// Backward compatible logging
		if d.log {
			logEntry.Duration = time.Since(start)
			logEntry.Handled = false
			logEntry.Error = result.Error
			d.addLog(logEntry)
		}

		// Apply error handling strategy
		d.applyErrorStrategy(intent, result)
		return result
	}

	// Check if this is a transition intent
	if d.registry.IsTransition(intent) {
		// Queue for async processing
		return d.dispatchTransition(handler, ctx, intent, lane, logEntry, start)
	}

	// Immediate dispatch for non-transition intents
	result := handler.Handle(ctx, intent)

	// Schedule fiber update if needed
	if result.Handled && d.scheduler != nil {
		d.scheduler(lane)
	}

	// Log result
	duration := time.Since(start)
	if d.logger != nil && d.logger.Enabled() {
		if result.Error != nil {
			d.logger.Error("Intent failed: type=%s, duration=%v, error=%v",
				intentType, duration, result.Error)
		} else if result.Handled {
			d.logger.Debug("Intent handled: type=%s, duration=%v", intentType, duration)
		} else {
			d.logger.Debug("Intent not handled: type=%s", intentType)
		}
	}

	// Backward compatible logging
	if d.log {
		logEntry.Duration = duration
		logEntry.Handled = result.Handled
		logEntry.Error = result.Error
		d.addLog(logEntry)
	}

	// Apply error handling strategy for failed intents
	if result.Error != nil {
		d.applyErrorStrategy(intent, result)
	}

	return result
}

// dispatchTransition handles async transition intents.
func (d *Dispatcher) dispatchTransition(handler Handler, ctx *ActionContext, intent Intent, lane priority.DirtyLevel, logEntry DispatchLog, start time.Time) IntentResult {
	// Log transition intent
	intentType := intent.IntentType()
	if d.logger != nil && d.logger.Enabled() {
		d.logger.Debug("Queueing transition intent: type=%s", intentType)
	}

	// Add to queue for async processing
	item := &queueItem{
		intent:  intent,
		handler: handler,
		ctx:     ctx,
		lane:    lane,
	}
	d.queue.Push(item)

	// For transitions, return immediately with async indicator
	// The actual processing happens in ProcessQueue
	done := make(chan struct{})

	duration := time.Since(start)
	if d.logger != nil && d.logger.Enabled() {
		d.logger.Debug("Transition intent queued: type=%s, duration=%v", intentType, duration)
	}

	if d.log {
		logEntry.Duration = duration
		logEntry.Handled = true
		d.addLog(logEntry)
	}

	return AsyncResult(done)
}

// ProcessQueue processes pending transition intents.
// This should be called by the scheduler during idle time.
func (d *Dispatcher) ProcessQueue(maxDuration time.Duration) int {
	deadline := time.Now().Add(maxDuration)
	processed := 0

	for time.Now().Before(deadline) {
		item := d.queue.Pop()
		if item == nil {
			break
		}

		// Execute handler
		result := item.handler.Handle(item.ctx, item.intent)

		// Schedule fiber update if needed
		if result.Handled && d.scheduler != nil {
			d.scheduler(item.lane)
		}

		processed++
	}

	return processed
}

// QueueSize returns the number of pending intents in the queue.
func (d *Dispatcher) QueueSize() int {
	return d.queue.Size()
}

// =============================================================================
// Logging
// =============================================================================

func (d *Dispatcher) addLog(entry DispatchLog) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logEntries = append(d.logEntries, entry)
	if len(d.logEntries) > d.logMaxSize {
		d.logEntries = d.logEntries[1:]
	}
}

// GetLogs returns recent dispatch logs.
func (d *Dispatcher) GetLogs() []DispatchLog {
	d.mu.RLock()
	defer d.mu.RUnlock()

	logs := make([]DispatchLog, len(d.logEntries))
	copy(logs, d.logEntries)
	return logs
}

// ClearLogs clears all dispatch logs.
func (d *Dispatcher) ClearLogs() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logEntries = make([]DispatchLog, 0)
}

// =============================================================================
// Intent Queue
// =============================================================================

type queueItem struct {
	intent  Intent
	handler Handler
	ctx     *ActionContext
	lane    priority.DirtyLevel
}

type intentQueue struct {
	mu    sync.Mutex
	items []*queueItem
}

func newIntentQueue() *intentQueue {
	return &intentQueue{
		items: make([]*queueItem, 0),
	}
}

func (q *intentQueue) Push(item *queueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *intentQueue) Pop() *queueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *intentQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
