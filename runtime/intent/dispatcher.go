package intent

import (
	"context"
	"fmt"
	"sync"
	"time"

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

	// log enables debug logging
	log bool

	// logEntries stores dispatch history
	logEntries []DispatchLog

	// logMaxSize limits log entries
	logMaxSize int
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

// NewDispatcher creates a new intent dispatcher.
func NewDispatcher(registry *Registry) *Dispatcher {
	return &Dispatcher{
		registry:    registry,
		queue:       newIntentQueue(),
		logMaxSize:  1000,
		logEntries:  make([]DispatchLog, 0),
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

// EnableLog enables or disables debug logging.
func (d *Dispatcher) EnableLog(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log = enabled
}

// =============================================================================
// Dispatch Methods
// =============================================================================

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

	// Create action context
	ctx := NewActionContext(context.Background(), source, d.stateSetter)

	// Log dispatch start
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
		if d.log {
			logEntry.Duration = time.Since(start)
			logEntry.Handled = false
			logEntry.Error = result.Error
			d.addLog(logEntry)
		}
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
	if d.log {
		logEntry.Duration = time.Since(start)
		logEntry.Handled = result.Handled
		logEntry.Error = result.Error
		d.addLog(logEntry)
	}

	return result
}

// dispatchTransition handles async transition intents.
func (d *Dispatcher) dispatchTransition(handler Handler, ctx *ActionContext, intent Intent, lane priority.DirtyLevel, logEntry DispatchLog, start time.Time) IntentResult {
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

	if d.log {
		logEntry.Duration = time.Since(start)
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
