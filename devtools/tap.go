// Package devtools provides the lock-free mutation tap for DevTools.
package devtools

import (
	"sync/atomic"
)

// mutationTap is the global mutation recording tap.
// Uses a lock-free ring buffer for zero-allocation recording.
var mutationTap = struct {
	enabled   uint32
	writePos  uint32
	buffer    []MutationRecord
	mask      uint32
}{
	// 16K ring buffer = 16K * 32 bytes = 512KB
	buffer: make([]MutationRecord, 1<<14),
	mask:   (1 << 14) - 1,
}

// nextMutationID is the global mutation ID counter.
var nextMutationID uint64

// RecordMutation records a mutation in the tap.
// This is extremely fast: just an atomic add and array write.
// Safe to call from any goroutine.
//
// Parameters:
//   - compID: Pre-allocated component ID (use runtime.RegisterComponent)
//   - fieldID: Pre-allocated field ID (use runtime.RegisterField)
//   - kind: Type of mutation (MutationKind)
//   - oldValue: Old value (encoded as uint64)
//   - newValue: New value (encoded as uint64)
func RecordMutation(compID uint32, fieldID uint16, kind uint8, oldValue, newValue uint64) {
	// Fast path: if disabled, return immediately
	if atomic.LoadUint32(&mutationTap.enabled) == 0 {
		return
	}

	// Atomically increment write position and get the new position
	i := atomic.AddUint32(&mutationTap.writePos, 1)

	// Write mutation record to ring buffer
	mutationTap.buffer[(i-1)&mutationTap.mask] = MutationRecord{
		ComponentID: compID,
		FieldID:     fieldID,
		Kind:        kind,
		OldValue:    oldValue,
		NewValue:    newValue,
	}
}

// EnableMutationTap enables the mutation tap.
func EnableMutationTap() {
	atomic.StoreUint32(&mutationTap.enabled, 1)
}

// DisableMutationTap disables the mutation tap.
func DisableMutationTap() {
	atomic.StoreUint32(&mutationTap.enabled, 0)
}

// IsMutationTapEnabled returns true if the mutation tap is enabled.
func IsMutationTapEnabled() bool {
	return atomic.LoadUint32(&mutationTap.enabled) != 0
}

// PollMutations consumes mutation records from the tap.
// This should be called by the debug goroutine to process recorded mutations.
//
// Parameters:
//   - fromPos: The position to start reading from (updated to new position)
//
// Returns:
//   - Slice of mutation records (caller owns this memory)
func PollMutations(fromPos *uint32) []MutationRecord {
	currentPos := atomic.LoadUint32(&mutationTap.writePos)
	if *fromPos >= currentPos {
		return nil
	}

	// Allocate result slice
	count := int(currentPos - *fromPos)
	result := make([]MutationRecord, count)

	// Copy mutations
	idx := 0
	for *fromPos < currentPos {
		rec := mutationTap.buffer[*fromPos&mutationTap.mask]
		result[idx] = rec
		idx++
		*fromPos++
	}

	return result
}

// GetNextMutationID returns the next mutation ID.
func GetNextMutationID() MutationID {
	id := atomic.AddUint64(&nextMutationID, 1)
	return MutationID(id)
}

// GetCurrentMutationID returns the current mutation ID without incrementing.
func GetCurrentMutationID() MutationID {
	return MutationID(atomic.LoadUint64(&nextMutationID))
}

// ResetMutationTap resets the mutation tap (for testing).
func ResetMutationTap() {
	atomic.StoreUint32(&mutationTap.enabled, 0)
	atomic.StoreUint32(&mutationTap.writePos, 0)
	atomic.StoreUint64(&nextMutationID, 0)
}

// GetTapStats returns statistics about the mutation tap.
func GetTapStats() (enabled bool, writePos uint32, bufferSize int) {
	enabled = atomic.LoadUint32(&mutationTap.enabled) != 0
	writePos = atomic.LoadUint32(&mutationTap.writePos)
	bufferSize = len(mutationTap.buffer)
	return
}
