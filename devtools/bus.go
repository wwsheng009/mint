// Package devtools provides the asynchronous event bus for DevTools.
package devtools

import (
	"runtime"
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
	Type  DebugEventType
	Data  uintptr // Pointer to pre-allocated memory or inline data
	Frame int     // Frame number
	Time  int64   // Nanoseconds (optional)
}

// EventBus is a ring buffer based event bus for debug events.
// It allows the render thread to emit events without blocking,
// while background goroutines process them asynchronously.
type EventBus struct {
	enabled  uint32       // Atomic flag for quick enable/disable check
	writePos uint32       // Current write position in the ring buffer
	buffer   []DebugEvent // Ring buffer
	mask     uint32       // Mask for ring buffer indexing (size - 1)
	bufferMu sync.Mutex   // Protects buffer writes and dispatch reads

	// Subscribers
	subscribers []chan<- DebugEvent
	subMu       sync.RWMutex

	// Shutdown
	done chan struct{}
	once sync.Once

	// P1-1: 统计信息
	stats EventBusStats
}

// EventBusStats 统计信息
type EventBusStats struct {
	EventsSent        atomic.Uint64
	EventsDropped     atomic.Uint64
	BackpressureDrops atomic.Uint64
	CurrentBufferLen  atomic.Uint64
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
// Safe to call from the render thread.
func (b *EventBus) Emit(ev DebugEvent) {
	// Fast path: if disabled, return immediately
	// Branch prediction will handle this well
	if atomic.LoadUint32(&b.enabled) == 0 {
		return
	}

	b.bufferMu.Lock()
	pos := atomic.LoadUint32(&b.writePos)
	b.buffer[pos&b.mask] = ev
	atomic.StoreUint32(&b.writePos, pos+1)
	b.bufferMu.Unlock()
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
// P1-1: 优化为智能等待，降低CPU占用
func (b *EventBus) dispatchLoop(ch chan<- DebugEvent) {
	readPos := uint32(0)
	pollInterval := 10 * time.Millisecond
	maxPollInterval := 100 * time.Millisecond
	currentPollInterval := pollInterval
	maxBatch := 1000 // 单次最多处理1000个事件

	defer func() {
		// 清理资源
	}()

	for {
		// 检查退出信号
		select {
		case <-b.done:
			return
		default:
		}

		events := make([]DebugEvent, 0, maxBatch)

		b.bufferMu.Lock()
		writePos := atomic.LoadUint32(&b.writePos)
		if readPos < writePos {
			for readPos < writePos && len(events) < maxBatch {
				events = append(events, b.buffer[readPos&b.mask])
				readPos++
			}
		}
		b.bufferMu.Unlock()

		// 如果没有新事件，智能等待
		if len(events) == 0 {
			// 动态调整轮询间隔：没有事件时降低频率
			select {
			case <-b.done:
				return
			case <-time.After(currentPollInterval):
				// 下次轮询间隔加倍，直到最大值
				currentPollInterval *= 2
				if currentPollInterval > maxPollInterval {
					currentPollInterval = maxPollInterval
				}
				continue
			}
		}

		// 有新事件，重置轮询间隔
		currentPollInterval = pollInterval

		for _, ev := range events {
			// 发送到订阅者，带背压处理
			select {
			case ch <- ev:
				b.stats.EventsSent.Add(1)
			case <-b.done:
				return
			default:
				// 背压：跳过此事件
				b.stats.BackpressureDrops.Add(1)
			}
		}

		// 如果处理了大量事件，让出 CPU 时间片
		if len(events) > 50 {
			runtime.Gosched()
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

// P1-1: GetStats returns the current event bus statistics.
func (b *EventBus) GetStats() *EventBusStats {
	// Calculate current buffer usage
	writePos := atomic.LoadUint32(&b.writePos)
	b.stats.CurrentBufferLen.Store(uint64(writePos & b.mask))

	return &b.stats
}
