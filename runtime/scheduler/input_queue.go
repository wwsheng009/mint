// Package scheduler provides priority-based scheduling for input events.
package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// InputPriority represents the priority level of an input event.
type InputPriority int

const (
	// PriorityImmediate is for events that must be handled immediately (e.g., system events).
	PriorityImmediate InputPriority = iota
	// PriorityHigh is for user-initiated events (e.g., key press, click).
	PriorityHigh
	// PriorityContinuous is for continuous input (e.g., mouse move, drag).
	PriorityContinuous
	// PriorityLow is for background tasks.
	PriorityLow
)

// String returns the string representation of the priority.
func (p InputPriority) String() string {
	switch p {
	case PriorityImmediate:
		return "immediate"
	case PriorityHigh:
		return "high"
	case PriorityContinuous:
		return "continuous"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// ToDirtyLevel converts InputPriority to priority.DirtyLevel.
func (p InputPriority) ToDirtyLevel() priority.DirtyLevel {
	switch p {
	case PriorityImmediate:
		return priority.DirtyHigh
	case PriorityHigh:
		return priority.DirtyHigh
	case PriorityContinuous:
		return priority.DirtyNormal
	case PriorityLow:
		return priority.DirtyLow
	default:
		return priority.DirtyNormal
	}
}

// InputEvent represents a single input event to be processed.
type InputEvent struct {
	// ID is a unique identifier for this event.
	ID string

	// Type is the type of input event (e.g., "key", "mouse", "resize").
	Type string

	// Data contains the event-specific data.
	Data interface{}

	// Priority is the event's priority level.
	Priority InputPriority

	// Timestamp when the event was queued.
	Timestamp time.Time

	// Sequence number for ordering within the same priority.
	Sequence int64
}

// inputItem is used internally by the priority queue.
type inputItem struct {
	event     *InputEvent
	index     int // Index in the heap
}

// inputHeap implements a priority queue for input events.
type inputHeap []*inputItem

// Len implements heap.Interface.
func (h inputHeap) Len() int { return len(h) }

// Less implements heap.Interface.
// Lower priority value = higher priority (processed first).
func (h inputHeap) Less(i, j int) bool {
	// First compare by priority
	if h[i].event.Priority != h[j].event.Priority {
		return h[i].event.Priority < h[j].event.Priority
	}
	// Same priority: use sequence number (FIFO within priority)
	return h[i].event.Sequence < h[j].event.Sequence
}

// Swap implements heap.Interface.
func (h inputHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push implements heap.Interface.
func (h *inputHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*inputItem)
	item.index = n
	*h = append(*h, item)
}

// Pop implements heap.Interface.
func (h *inputHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// InputQueue manages input events with priority-based processing.
type InputQueue struct {
	mu       sync.RWMutex
	heap     inputHeap
	sequence int64
	closed   bool
	notEmpty chan struct{}
}

// NewInputQueue creates a new input queue.
func NewInputQueue() *InputQueue {
 iq := &InputQueue {
  heap:     make(inputHeap, 0),
  notEmpty: make(chan struct{}, 1),
 }
 heap.Init(&iq.heap)
 return iq
}

// Push adds an event to the queue with the given priority.
func (iq *InputQueue) Push(eventType string, data interface{}, prio InputPriority) *InputEvent {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 if iq.closed {
  return nil
 }

 event := &InputEvent{
  ID:        generateEventID(),
  Type:      eventType,
  Data:      data,
  Priority:  prio,
  Timestamp: time.Now(),
  Sequence:  iq.sequence,
 }
 iq.sequence++

 item := &inputItem{event: event}
 heap.Push(&iq.heap, item)

 // Signal non-empty
 select {
 case iq.notEmpty <- struct{}{}:
 default:
 }

 return event
}

// Pop removes and returns the highest priority event.
// Returns nil if the queue is empty.
func (iq *InputQueue) Pop() *InputEvent {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 if iq.heap.Len() == 0 {
  return nil
 }

 item := heap.Pop(&iq.heap).(*inputItem)
 return item.event
}

// PopBlocking waits for an event to be available and returns it.
// Returns nil if the queue is closed.
func (iq *InputQueue) PopBlocking(timeout time.Duration) *InputEvent {
 if timeout > 0 {
  select {
  case <-iq.notEmpty:
  case <-time.After(timeout):
   return nil
  }
 } else {
  <-iq.notEmpty
 }

 return iq.Pop()
}

// Peek returns the highest priority event without removing it.
func (iq *InputQueue) Peek() *InputEvent {
 iq.mu.RLock()
 defer iq.mu.RUnlock()

 if iq.heap.Len() == 0 {
  return nil
 }

 return iq.heap[0].event
}

// HasPending returns true if there are pending events.
func (iq *InputQueue) HasPending() bool {
 iq.mu.RLock()
 defer iq.mu.RUnlock()
 return iq.heap.Len() > 0
}

// Len returns the number of pending events.
func (iq *InputQueue) Len() int {
 iq.mu.RLock()
 defer iq.mu.RUnlock()
 return iq.heap.Len()
}

// Clear removes all pending events.
func (iq *InputQueue) Clear() {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 iq.heap = make(inputHeap, 0)
 heap.Init(&iq.heap)
}

// Close closes the queue. No more events can be pushed.
func (iq *InputQueue) Close() {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 iq.closed = true
 close(iq.notEmpty)
}

// IsClosed returns true if the queue is closed.
func (iq *InputQueue) IsClosed() bool {
 iq.mu.RLock()
 defer iq.mu.RUnlock()
 return iq.closed
}

// RemoveByID removes an event by its ID.
// Returns true if the event was found and removed.
func (iq *InputQueue) RemoveByID(id string) bool {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 for i, item := range iq.heap {
  if item.event.ID == id {
   heap.Remove(&iq.heap, i)
   return true
  }
 }

 return false
}

// Drain removes and returns all pending events in priority order.
func (iq *InputQueue) Drain() []*InputEvent {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 events := make([]*InputEvent, 0, iq.heap.Len())

 for iq.heap.Len() > 0 {
  item := heap.Pop(&iq.heap).(*inputItem)
  events = append(events, item.event)
 }

 return events
}

// GetByType returns all pending events of a specific type.
func (iq *InputQueue) GetByType(eventType string) []*InputEvent {
 iq.mu.RLock()
 defer iq.mu.RUnlock()

 var result []*InputEvent

 for _, item := range iq.heap {
  if item.event.Type == eventType {
   result = append(result, item.event)
  }
 }

 return result
}

// Stats returns statistics about the queue.
func (iq *InputQueue) Stats() InputQueueStats {
 iq.mu.RLock()
 defer iq.mu.RUnlock()

 stats := InputQueueStats{
  Total:   iq.heap.Len(),
  ByType:  make(map[string]int),
  ByPrio:  make(map[InputPriority]int),
 }

 for _, item := range iq.heap {
  stats.ByType[item.event.Type]++
  stats.ByPrio[item.event.Priority]++
 }

 return stats
}

// InputQueueStats contains statistics about the input queue.
type InputQueueStats struct {
 Total  int
 ByType map[string]int
 ByPrio map[InputPriority]int
}

// =============================================================================
// Event ID Generation
// =============================================================================

var (
	eventCounter int64
	eventIDMu    sync.Mutex
)

func generateEventID() string {
	eventIDMu.Lock()
	defer eventIDMu.Unlock()

	eventCounter++
	return time.Now().Format("20060102150405.000") + "-" + string(rune(eventCounter%1000))
}

// =============================================================================
// Priority Helpers
// =============================================================================

// EventPriority returns the appropriate priority for a given event type.
func EventPriority(eventType string) InputPriority {
	switch eventType {
	case "key", "click", "resize", "interrupt":
		return PriorityHigh
	case "mouse", "scroll":
		return PriorityContinuous
	default:
		return PriorityNormal
	}
}

// PriorityNormal is the default priority.
const PriorityNormal = PriorityContinuous

// =============================================================================
// Batch Operations
// =============================================================================

// PopMultiple pops up to n events in priority order.
func (iq *InputQueue) PopMultiple(n int) []*InputEvent {
 iq.mu.Lock()
 defer iq.mu.Unlock()

 if n <= 0 {
  return nil
 }

 count := min(n, iq.heap.Len())
 events := make([]*InputEvent, 0, count)

 for i := 0; i < count; i++ {
  item := heap.Pop(&iq.heap).(*inputItem)
  events = append(events, item.event)
 }

 // Signal if still not empty
 if iq.heap.Len() > 0 {
  select {
  case iq.notEmpty <- struct{}{}:
  default:
  }
 }

 return events
}

func min(a, b int) int {
 if a < b {
  return a
 }
 return b
}
