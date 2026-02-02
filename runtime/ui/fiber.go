package ui

import "fmt"

// =============================================================================
// Fiber Architecture
// =============================================================================
// Fiber is the reconciliation algorithm, breaking rendering work into small units.
// Each Fiber node represents a component in the UI and forms a tree.
// =============================================================================

// EffectFlag represents the side effects of a fiber
type EffectFlag int

const (
	// EffectNoEffect indicates no side effects
	EffectNoEffect EffectFlag = 0
	// EffectPlacement indicates the fiber was newly inserted
	EffectPlacement EffectFlag = 1 << iota
	// EffectUpdate indicates the fiber was updated
	EffectUpdate
	// EffectDeletion indicates the fiber was deleted
	EffectDeletion
	// EffectRef indicates a ref changed
	EffectRef
)

// Lane represents the priority of work (lanes model)
type Lane uint64

const (
	// LaneNoLane indicates no lane
	LaneNoLane Lane = 0
	// LaneSyncLane is the default synchronous lane
	LaneSyncLane Lane = 1
	// LaneInputContinuousLane is for continuous input (dragging, typing)
	LaneInputContinuousLane Lane = 1 << 1
	// LaneDefaultLane is for default updates
	LaneDefaultLane Lane = 1 << 2
	// LaneIdleLane is for low-priority work
	LaneIdleLane Lane = 1 << 3
)

// LaneRoot represents all lanes combined
const LaneRoot Lane = LaneSyncLane | LaneInputContinuousLane | LaneDefaultLane | LaneIdleLane

// Fiber represents a unit of work in the reconciler
// Each Fiber corresponds to a VNode and contains work-in-progress state
type Fiber struct {
	// === VNode Reference ===
	// The virtual node this fiber represents
	VNode VNode

	// === Tree Structure ===
	// Pointer to parent fiber
	Return *Fiber
	// First child fiber
	Child *Fiber
	// Next sibling fiber
	Sibling *Fiber
	// For alternate tree (double buffering)
	Alternate *Fiber

	// === Props & State ===
	// Props for this fiber
	Props Props
	// Memoized props (previous props)
	MemoizedProps Props
	// Memoized state
	MemoizedState interface{}
	// Update queue
	UpdateQueue *UpdateQueue

	// === Effect Flags ===
	// Flags indicating what side effects this fiber has
	Flags EffectFlag
	// SubtreeFlags indicating effects in descendants
	SubtreeFlags EffectFlag

	// === Priority ===
	// Lanes representing work priority
	Lanes Lane
	// ChildLanes for pending work in children
	ChildLanes Lane

	// === Key for Diff ===
	// Key for reconciling lists
	Key string

	// === Type ===
	// The type of this fiber (cached from VNode)
	Type VNodeType

	// === Tag for debugging ===
	// Component or element tag
	Tag string

	// === Component Instance ===
	// Persistent component instance for state preservation
	ComponentInstance ComponentInstance
}

// Update represents a state update
type Update struct {
	// The new state or updater function
	Payload interface{}
	// The next update in the queue
	Next *Update
	// Lane for this update
	Lane Lane
	// Callback after update is committed
	Callback func()
}

// UpdateQueue holds pending updates
type UpdateQueue struct {
	// First pending update
	First *Update
	// Last pending update
	Last *Update
}

// =============================================================================
// Lane Operations
// =============================================================================

// MergeLanes combines two lane values
func MergeLanes(a, b Lane) Lane {
	return a | b
}

// RemoveLanes removes lanes from a
func RemoveLanes(a, b Lane) Lane {
	return a &^ b
}

// HasLanes checks if 'a' contains any lanes from 'b'
func HasLanes(a, b Lane) bool {
	return (a & b) != 0
}

// IsSubsetLanes checks if 'a' is a subset of 'b'
func IsSubsetLanes(a, b Lane) bool {
	return (a & b) == a
}

// GetHighestPriorityLane returns the highest priority lane set
func GetHighestPriorityLane(lanes Lane) Lane {
	// Find the least significant bit (highest priority)
	return lanes & (-lanes)
}

// =============================================================================
// Fiber String Representation
// =============================================================================

// String returns a string representation of the fiber
func (f *Fiber) String() string {
	if f == nil {
		return "nil"
	}

	return fmt.Sprintf("Fiber{Tag: %s, Key: %s, Flags: %d, Lanes: %d}",
		f.Tag, f.Key, f.Flags, f.Lanes)
}

// =============================================================================
// Fiber Helpers
// =============================================================================

// HasNoPendingWork returns true if the fiber has no pending work
func (f *Fiber) HasNoPendingWork() bool {
	return f.Lanes == LaneNoLane && f.ChildLanes == LaneNoLane
}

// HasEffect returns true if the fiber has effects
func (f *Fiber) HasEffect() bool {
	return f.Flags != EffectNoEffect
}

// HasSubtreeEffect returns true if any descendant has effects
func (f *Fiber) HasSubtreeEffect() bool {
	return f.SubtreeFlags != EffectNoEffect
}

// MarkUpdate marks the fiber as needing an update
func (f *Fiber) MarkUpdate(lane Lane) {
	f.Lanes = MergeLanes(f.Lanes, lane)
	f.Flags |= EffectUpdate

	// Propagate to parents
	for parent := f.Return; parent != nil; parent = parent.Return {
		parent.ChildLanes = MergeLanes(parent.ChildLanes, lane)
	}
}

// =============================================================================
// Update Queue Operations
// =============================================================================

// EnqueueUpdate adds an update to the fiber's queue
func (f *Fiber) EnqueueUpdate(update *Update) {
	if f.UpdateQueue == nil {
		f.UpdateQueue = &UpdateQueue{}
	}

	update.Next = nil
	if f.UpdateQueue.Last == nil {
		f.UpdateQueue.First = update
		f.UpdateQueue.Last = update
	} else {
		f.UpdateQueue.Last.Next = update
		f.UpdateQueue.Last = update
	}

	// Mark fiber as having work
	f.MarkUpdate(LaneSyncLane)
}
