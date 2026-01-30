// Package memory provides memory optimization utilities for DevTools.
package memory

import (
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Ring Buffer for Circular Storage
// =============================================================================

// RingBuffer is a circular buffer for storing frame data.
// It automatically overwrites old data when full.
type RingBuffer struct {
	mu       sync.RWMutex
	buffer   []devtools.FrameID
	capacity int
	head     int  // Write position
	tail     int  // Read position
	size     int  // Current number of elements
	full     bool // Whether buffer is full
}

// NewRingBuffer creates a new ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}

	return &RingBuffer{
		buffer:   make([]devtools.FrameID, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
		full:     false,
	}
}

// Write adds a frame ID to the buffer.
func (rb *RingBuffer) Write(frameID devtools.FrameID) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = frameID
	rb.head = (rb.head + 1) % rb.capacity

	if rb.full {
		// Buffer was full, advance tail
		rb.tail = (rb.tail + 1) % rb.capacity
	} else if rb.head == rb.tail {
		// Buffer just became full
		rb.full = true
	} else {
		rb.size++
	}
}

// Read reads and removes the oldest frame ID from the buffer.
func (rb *RingBuffer) Read() (devtools.FrameID, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size == 0 && !rb.full {
		return 0, false
	}

	frameID := rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.full = false
	rb.size--

	return frameID, true
}

// Peek returns the oldest frame ID without removing it.
func (rb *RingBuffer) Peek() (devtools.FrameID, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 && !rb.full {
		return 0, false
	}

	return rb.buffer[rb.tail], true
}

// GetAll returns all frame IDs in order (oldest to newest).
func (rb *RingBuffer) GetAll() []devtools.FrameID {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]devtools.FrameID, 0, rb.size)

	if rb.full {
		// Buffer is full, read from tail to head
		for i := 0; i < rb.capacity; i++ {
			idx := (rb.tail + i) % rb.capacity
			result = append(result, rb.buffer[idx])
		}
	} else {
		// Buffer is not full, read from tail to head
		for i := rb.tail; i != rb.head; i = (i + 1) % rb.capacity {
			result = append(result, rb.buffer[i])
		}
	}

	return result
}

// GetRange returns frame IDs in the range [start, end].
func (rb *RingBuffer) GetRange(start, end int) []devtools.FrameID {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	all := rb.GetAll()

	if start < 0 {
		start = 0
	}
	if end >= len(all) {
		end = len(all) - 1
	}
	if start > end {
		return []devtools.FrameID{}
	}

	return all[start : end+1]
}

// Size returns the current number of elements.
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.full {
		return rb.capacity
	}
	return rb.size
}

// Capacity returns the buffer capacity.
func (rb *RingBuffer) Capacity() int {
	return rb.capacity
}

// IsFull returns true if the buffer is full.
func (rb *RingBuffer) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.full
}

// IsEmpty returns true if the buffer is empty.
func (rb *RingBuffer) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size == 0 && !rb.full
}

// Clear clears the buffer.
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.tail = 0
	rb.size = 0
	rb.full = false
}

// Resize changes the buffer capacity.
// If the new capacity is smaller, oldest data is lost.
func (rb *RingBuffer) Resize(newCapacity int) {
	if newCapacity <= 0 {
		return
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	if newCapacity == rb.capacity {
		return
	}

	// Get current data
	oldData := make([]devtools.FrameID, 0, rb.size)
	if rb.full {
		for i := 0; i < rb.capacity; i++ {
			oldData = append(oldData, rb.buffer[i])
		}
	} else {
		for i := rb.tail; i != rb.head; i = (i + 1) % rb.capacity {
			oldData = append(oldData, rb.buffer[i])
		}
	}

	// Create new buffer
	rb.buffer = make([]devtools.FrameID, newCapacity)
	rb.capacity = newCapacity
	rb.head = 0
	rb.tail = 0
	rb.full = false
	rb.size = 0

	// Copy data back (keep newest if shrinking)
	copyCount := len(oldData)
	if copyCount > newCapacity {
		copyCount = newCapacity
		// Keep newest data
		oldData = oldData[len(oldData)-copyCount:]
	}

	for _, frameID := range oldData {
		rb.buffer[rb.head] = frameID
		rb.head = (rb.head + 1) % newCapacity
		rb.size++
	}
}

// Stats returns buffer statistics.
type RingBufferStats struct {
	Capacity int           `json:"capacity"`
	Size     int           `json:"size"`
	Usage    float64       `json:"usage_percent"`
	Full     bool          `json:"full"`
	Empty    bool          `json:"empty"`
}

// GetStats returns buffer statistics.
func (rb *RingBuffer) GetStats() RingBufferStats {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	size := rb.size
	if rb.full {
		size = rb.capacity
	}

	usage := 0.0
	if rb.capacity > 0 {
		usage = float64(size) / float64(rb.capacity) * 100
	}

	return RingBufferStats{
		Capacity: rb.capacity,
		Size:     size,
		Usage:    usage,
		Full:     rb.full,
		Empty:    rb.size == 0 && !rb.full,
	}
}

// =============================================================================
// Generic Ring Buffer for any type
// =============================================================================

// GenericRingBuffer is a generic circular buffer.
type GenericRingBuffer[T any] struct {
	mu       sync.RWMutex
	buffer   []T
	capacity int
	head     int
	tail     int
	size     int
	full     bool
}

// NewGenericRingBuffer creates a new generic ring buffer.
func NewGenericRingBuffer[T any](capacity int) *GenericRingBuffer[T] {
	if capacity <= 0 {
		capacity = 1000
	}

	return &GenericRingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
	}
}

// Write adds an element to the buffer.
func (rb *GenericRingBuffer[T]) Write(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = item
	rb.head = (rb.head + 1) % rb.capacity

	if rb.full {
		rb.tail = (rb.tail + 1) % rb.capacity
	} else if rb.head == rb.tail {
		rb.full = true
	} else {
		rb.size++
	}
}

// Read reads and removes the oldest element.
func (rb *GenericRingBuffer[T]) Read() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	var zero T
	if rb.size == 0 && !rb.full {
		return zero, false
	}

	item := rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.full = false
	rb.size--

	return item, true
}

// GetAll returns all elements in order.
func (rb *GenericRingBuffer[T]) GetAll() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]T, 0, rb.size)

	if rb.full {
		for i := 0; i < rb.capacity; i++ {
			idx := (rb.tail + i) % rb.capacity
			result = append(result, rb.buffer[idx])
		}
	} else {
		for i := rb.tail; i != rb.head; i = (i + 1) % rb.capacity {
			result = append(result, rb.buffer[i])
		}
	}

	return result
}

// Size returns the current size.
func (rb *GenericRingBuffer[T]) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.full {
		return rb.capacity
	}
	return rb.size
}

// Clear clears the buffer.
func (rb *GenericRingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.tail = 0
	rb.size = 0
	rb.full = false
}

// =============================================================================
// Frame Window (Sliding Window for Recent Frames)
// =============================================================================

// FrameWindow maintains a sliding window of recent frames.
type FrameWindow struct {
	ring     *RingBuffer
	duration time.Duration // Time window to keep
	timestamps map[devtools.FrameID]time.Time
	mu       sync.RWMutex
}

// NewFrameWindow creates a new frame window.
func NewFrameWindow(capacity int, duration time.Duration) *FrameWindow {
	return &FrameWindow{
		ring:       NewRingBuffer(capacity),
		duration:   duration,
		timestamps: make(map[devtools.FrameID]time.Time),
	}
}

// Add adds a frame to the window.
func (fw *FrameWindow) Add(frameID devtools.FrameID, timestamp time.Time) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.ring.Write(frameID)
	fw.timestamps[frameID] = timestamp

	// Clean up old timestamps outside the window
	cutoff := timestamp.Add(-fw.duration)
	for fid, ts := range fw.timestamps {
		if ts.Before(cutoff) {
			delete(fw.timestamps, fid)
		}
	}
}

// GetRecent returns frames within the time window.
func (fw *FrameWindow) GetRecent() []devtools.FrameID {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	cutoff := time.Now().Add(-fw.duration)
	result := make([]devtools.FrameID, 0)

	all := fw.ring.GetAll()
	// Reverse to get newest first
	for i := len(all) - 1; i >= 0; i-- {
		if ts, ok := fw.timestamps[all[i]]; ok && ts.After(cutoff) {
			result = append(result, all[i])
		}
	}

	return result
}

// GetFrameCount returns the number of frames in the window.
func (fw *FrameWindow) GetFrameCount() int {
	return len(fw.timestamps)
}

// Clear clears the window.
func (fw *FrameWindow) Clear() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.ring.Clear()
	fw.timestamps = make(map[devtools.FrameID]time.Time)
}
