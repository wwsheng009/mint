// Package scheduler provides priority-based scheduling for UI rendering.
// It implements a Lane-based scheduling system inspired by React's priority system.
//
// Lanes allow the scheduler to:
//   - Prioritize user input over background updates
//   - Interrupt low-priority work for high-priority tasks
//   - Batch updates of the same priority
//   - Avoid blocking the main thread for large renders
package scheduler

import (
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Lane Definition
// =============================================================================

// Lane represents a priority level for rendering work.
// Lanes are implemented as bit flags to allow combining multiple lanes.
//
// Priority order (highest to lowest):
//
//	SyncLane > InputLane > AnimationLane > TransitionLane > IdleLane
//
// Higher priority lanes can interrupt lower priority lanes.
type Lane uint32

const (
	// NoLane indicates no work is scheduled.
	NoLane Lane = 0

	// SyncLane is for synchronous, immediate rendering.
	// Used for: blocking operations, critical updates
	// Cannot be interrupted.
	SyncLane Lane = 1 << iota

	// InputLane is for user input handling.
	// Used for: keyboard input, mouse clicks, form changes
	// High priority to maintain responsiveness.
	InputLane

	// AnimationLane is for animations and visual transitions.
	// Used for: CSS animations, smooth transitions
	// Medium-high priority for smooth UX.
	AnimationLane

	// DefaultLane is the default priority for most updates.
	// Used for: normal state updates, data fetching
	DefaultLane

	// TransitionLane is for non-urgent updates.
	// Used for: route transitions, large list updates
	// Can be interrupted by higher priority lanes.
	TransitionLane

	// IdleLane is for background work.
	// Used for: prefetching, analytics, non-critical updates
	// Lowest priority, runs only when idle.
	IdleLane
)

// AllLanes contains all lane bits.
const AllLanes Lane = SyncLane | InputLane | AnimationLane | DefaultLane | TransitionLane | IdleLane

// =============================================================================
// Lane Operations
// =============================================================================

// MergeLanes combines multiple lanes into a single lane bitmask.
//
// Example:
//
//	lanes := MergeLanes(InputLane, TransitionLane)
func MergeLanes(lanes ...Lane) Lane {
	var result Lane
	for _, l := range lanes {
		result |= l
	}
	return result
}

// Includes checks if the lane set includes a specific lane.
//
// Example:
//
//	if lanes.Includes(InputLane) {
//	    // Handle high-priority input
//	}
func (l Lane) Includes(lane Lane) bool {
	return l&lane != 0
}

// IsHigherPriorityThan returns true if l has higher priority than other.
//
// Example:
//
//	if InputLane.IsHigherPriorityThan(TransitionLane) { // true
//	    // Input is more important
//	}
func (l Lane) IsHigherPriorityThan(other Lane) bool {
	return l < other && l != NoLane
}

// IsLowerPriorityThan returns true if l has lower priority than other.
func (l Lane) IsLowerPriorityThan(other Lane) bool {
	return l > other && l != NoLane
}

// PickHighestPriorityLane returns the highest priority lane from a set.
//
// Example:
//
//	lanes := MergeLanes(TransitionLane, InputLane, IdleLane)
//	highest := PickHighestPriorityLane(lanes) // InputLane
func PickHighestPriorityLane(lanes Lane) Lane {
	if lanes == NoLane {
		return NoLane
	}

	// Check from highest to lowest priority
	for _, lane := range []Lane{SyncLane, InputLane, AnimationLane, DefaultLane, TransitionLane, IdleLane} {
		if lanes.Includes(lane) {
			return lane
		}
	}
	return NoLane
}

// PickLowestPriorityLane returns the lowest priority lane from a set.
//
// Example:
//
//	lanes := MergeLanes(TransitionLane, InputLane, IdleLane)
//	lowest := PickLowestPriorityLane(lanes) // IdleLane
func PickLowestPriorityLane(lanes Lane) Lane {
	if lanes == NoLane {
		return NoLane
	}

	// Check from lowest to highest priority
	for _, lane := range []Lane{IdleLane, TransitionLane, DefaultLane, AnimationLane, InputLane, SyncLane} {
		if lanes.Includes(lane) {
			return lane
		}
	}
	return NoLane
}

// String returns a human-readable representation of the lane.
func (l Lane) String() string {
	if l == NoLane {
		return "NoLane"
	}

	var parts []string
	if l.Includes(SyncLane) {
		parts = append(parts, "Sync")
	}
	if l.Includes(InputLane) {
		parts = append(parts, "Input")
	}
	if l.Includes(AnimationLane) {
		parts = append(parts, "Animation")
	}
	if l.Includes(DefaultLane) {
		parts = append(parts, "Default")
	}
	if l.Includes(TransitionLane) {
		parts = append(parts, "Transition")
	}
	if l.Includes(IdleLane) {
		parts = append(parts, "Idle")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Lane(%d)", l)
	}
	return strings.Join(parts, "|")
}

// Priority returns a numeric priority value (lower = higher priority).
func (l Lane) Priority() int {
	switch {
	case l.Includes(SyncLane):
		return 0
	case l.Includes(InputLane):
		return 1
	case l.Includes(AnimationLane):
		return 2
	case l.Includes(DefaultLane):
		return 3
	case l.Includes(TransitionLane):
		return 4
	case l.Includes(IdleLane):
		return 5
	default:
		return 6
	}
}

// =============================================================================
// Deadline Management
// =============================================================================

// DeadlineInterval defines the time budget for each lane type.
// Higher priority lanes get more time before yielding.
var DeadlineInterval = map[Lane]time.Duration{
	SyncLane:       0,             // No deadline, run to completion
	InputLane:      100 * time.Millisecond,
	AnimationLane:  33 * time.Millisecond,  // ~30fps
	DefaultLane:    50 * time.Millisecond,
	TransitionLane: 100 * time.Millisecond,
	IdleLane:       time.Second, // Very relaxed deadline
}

// GetDeadline returns the deadline duration for a lane.
func GetDeadline(lane Lane) time.Duration {
	if d, ok := DeadlineInterval[lane]; ok {
		return d
	}
	return 50 * time.Millisecond // Default deadline
}

// =============================================================================
// Lane Utilities
// =============================================================================

// IsInteractive returns true if the lane requires immediate user feedback.
func (l Lane) IsInteractive() bool {
	return l.Includes(SyncLane) || l.Includes(InputLane)
}

// IsInterruptible returns true if the lane can be interrupted.
func (l Lane) IsInterruptible() bool {
	return !l.Includes(SyncLane)
}

// ShouldYield checks if the current work should yield to higher priority work.
// This is typically called during render loops.
type ShouldYieldFunc func() bool

// DefaultShouldYield returns a simple should-yield function based on deadline.
func DefaultShouldYield(startTime time.Time, deadline time.Duration) ShouldYieldFunc {
	return func() bool {
		return time.Since(startTime) >= deadline
	}
}

// =============================================================================
// Lane Selector
// =============================================================================

// LaneSelector helps select appropriate lanes for different operations.
type LaneSelector struct {
	defaultLane Lane
}

// NewLaneSelector creates a new LaneSelector.
func NewLaneSelector() *LaneSelector {
	return &LaneSelector{
		defaultLane: DefaultLane,
	}
}

// SetDefault sets the default lane.
func (s *LaneSelector) SetDefault(lane Lane) {
	s.defaultLane = lane
}

// ForUserInput returns the appropriate lane for user input.
func (s *LaneSelector) ForUserInput() Lane {
	return InputLane
}

// ForDataFetch returns the appropriate lane for data fetching.
func (s *LaneSelector) ForDataFetch(isCritical bool) Lane {
	if isCritical {
		return DefaultLane
	}
	return TransitionLane
}

// ForAnimation returns the appropriate lane for animations.
func (s *LaneSelector) ForAnimation() Lane {
	return AnimationLane
}

// ForBackground returns the appropriate lane for background work.
func (s *LaneSelector) ForBackground() Lane {
	return IdleLane
}

// Default returns the default lane.
func (s *LaneSelector) Default() Lane {
	return s.defaultLane
}
