// Package devtools provides the asynchronous event bus for DevTools.
package devtools

import (
	"sync"
	"sync/atomic"
	"time"
)

// DebugEventType represents the type of debug event.
type DebugEventType uint8

const (
	// EventLayout indicates a layout event.
	EventLayout DebugEventType = iota
	// EventRepaint indicates a repaint event.
	EventRepaint
	// EventInput indicates an input event.
	EventInput
	// EventMutation indicates a mutation event.
	EventMutation
	// EventFocus indicates a focus event.
	EventFocus
)

// DebugEvent represents a lightweight debug event.
// This structure is designed to be allocated on stack and copied without heap allocation.
type DebugEvent struct {
	Type   DebugEventType
	Data   uintptr  // Pointer to pre-allocated memory or inline data
	Frame  int      // Frame number
	Time   int64    // Nanoseconds (optional)
}

// EventBus is a lock-free ring buffer based event bus for debug events.
// It allows the render thread to emit events without blocking,
// while background goroutines process them asynchronously.
type EventBus struct {
	enabled    uint32        // Atomic flag for quick enable/disable check
	writePos   uint32        // Current write position in the ring buffer
	buffer     []DebugEvent  // Ring buffer
	mask       uint32        // Mask for ring buffer indexing (size - 1)

	// Subscribers
	subscribers []chan<- DebugEvent
	subMu       sync.RWMutex

	// Shutdown
	done chan struct{}
	once sync.Once
}

// NewEventBus creates a new event bus with the specified buffer size.
// The buffer size must be a power of 2 for efficient ring buffer operation.
func NewEventBus(size int) *EventBus {
	// Ensure size is a power of 2
	if size&(size-1) != 0 {
		// Round up to next power of 2
		size = 1 << (32 - nlz32(uint32(size)))
	}

	return &EventBus{
		buffer:      make([]DebugEvent, size),
		mask:        uint32(size - 1),
		subscribers: make([]chan<- DebugEvent, 0),
		done:        make(chan struct{}),
	}
}

// nlz32 counts leading zeros in a 32-bit integer.
func nlz32(x uint32) uint32 {
	if x == 0 {
		return 32
}
	var n uint32 = 0
	if x>>16 == 0 {
		n += 16
		x <<= 16
	}
	if x>>24 == 0 {
		n += 8
		x <<= 8
	}
	if x>>28 == 0 {
		n += 4
		x <<= 4
	}
	if x>>30 == 0 {
		n += 2
		x <<= 2
	}
	if x>>31 == 0 {
		n++
	}
	return n
}

// Emit sends an event to the bus.
// This is lock-free and extremely fast (just an atomic add and array write).
// Safe to call from the render thread.
func (b *EventBus) Emit(ev DebugEvent) {
	// Fast path: if disabled, return immediately
	// Branch prediction will handle this well
	if atomic.LoadUint32(&b.enabled) == 0 {
		return
	}

	// Atomically increment write position and get the new position
	pos := atomic.AddUint32(&b.writePos, 1)

	// Write event to ring buffer (masked to wrap around)
	b.buffer[(pos-1)&b.mask] = ev
}

// Subscribe subscribes a channel to receive events from the bus.
// Returns a function that can be called to unsubscribe.
// Each subscriber gets its own dispatch goroutine.
func (b *EventBus) Subscribe(ch chan<- DebugEvent) func() {
	b.subMu.Lock()
	defer b.subMu.Unlock()

	b.subscribers = append(b.subscribers, ch)

	// Start dispatch loop for this subscriber
	go b.dispatchLoop(ch)

	// Return unsubscribe function
	return func() {
		b.subMu.Lock()
		defer b.subMu.Unlock()
		for i, sub := range b.subscribers {
			if sub == ch {
				b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
				break
			}
		}
	}
}

// dispatchLoop runs in a separate goroutine for each subscriber.
// It reads events from the ring buffer and sends them to the subscriber's channel.
func (b *EventBus) dispatchLoop(ch chan<- DebugEvent) {
	readPos := uint32(0)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			// Get current write position
			writePos := atomic.LoadUint32(&b.writePos)

			// Process all new events
			for readPos < writePos {
				ev := b.buffer[readPos&b.mask]
				readPos++

				// Send to subscriber with backpressure handling
				select {
				case ch <- ev:
				default:
					// Backpressure: skip this event
					// This prevents blocking the dispatch goroutine
				}
			}
		}
	}
}

// Enable enables the event bus.
func (b *EventBus) Enable() {
	atomic.StoreUint32(&b.enabled, 1)
}

// Disable disables the event bus.
func (b *EventBus) Disable() {
	atomic.StoreUint32(&b.enabled, 0)
}

// IsEnabled returns true if the event bus is enabled.
func (b *EventBus) IsEnabled() bool {
	return atomic.LoadUint32(&b.enabled) != 0
}

// Close closes the event bus and stops all dispatch goroutines.
func (b *EventBus) Close() {
	b.once.Do(func() {
		close(b.done)
	})
}

// WritePosition returns the current write position.
// This is useful for debugging and testing.
func (b *EventBus) WritePosition() uint32 {
	return atomic.LoadUint32(&b.writePos)
}
