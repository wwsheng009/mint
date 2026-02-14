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
//
// NOTE: The current implementation operates in synchronous mode, meaning all
// updates are processed immediately regardless of lane priority. The lane system
// is maintained for:
// 1. Future extensibility - async scheduling can be added
// 2. Work categorization - tracking what type of update each fiber represents
// 3. Scheduler integration - the scheduler uses lanes for priority decisions
//
// To fully implement priority scheduling, the work loop would need to:
// - Process higher priority lanes first
// - Time-slice work and resume for lower priority lanes
// - Defer low-priority work when high-priority work is pending
type Lane uint64

const (
	// LaneNoLane indicates no lane
	LaneNoLane Lane = 0
	// LaneSyncLane is the default synchronous lane (highest priority)
	LaneSyncLane Lane = 1
	// LaneInputContinuousLane is for continuous input (dragging, typing)
	// Higher priority than default, lower than sync
	LaneInputContinuousLane Lane = 1 << 1
	// LaneDefaultLane is for default updates (normal priority)
	LaneDefaultLane Lane = 1 << 2
	// LaneIdleLane is for low-priority work (background tasks)
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
	// Props for this fiber (snapshot taken when Fiber was created).
	// NOTE: This field may become stale if VNode props are updated.
	// Use GetProps() method to get current props from VNode.
	Props Props
	// Memoized props (previous props for diffing)
	MemoizedProps Props
	// Memoized state serves different purposes based on VNode type:
	// - TextVNode: stores text content (set by completeWorkText)
	// - ComponentVNode with UpdateQueue: stores state for functional updates (beginWork)
	// - ComponentVNode with hooks: NOT used (state is in ComponentContext.Hooks)
	MemoizedState interface{}
	// Update queue
	UpdateQueue *UpdateQueue

	// === Effect Flags ===
	// Flags indicating what side effects this fiber has
	Flags EffectFlag
	// SubtreeFlags indicating effects in descendants
	//
	// This field aggregates all effect flags from the entire subtree below this fiber.
	// It is computed during the render phase by collectChildEffects() which:
	// 1. ORs each child's Flags into the parent's SubtreeFlags
	// 2. ORs each child's SubtreeFlags into the parent's SubtreeFlags
	//
	// This allows ancestors to efficiently know if any descendant has effects
	// without traversing the entire subtree during commit.
	//
	// Propagation is bottom-up: when a child's flags change, all ancestors
	// are updated during the next render cycle.
	//
	// Example:
	//   Parent (Flags: 0, SubtreeFlags: EffectUpdate | EffectDeletion)
	//     ├── Child1 (Flags: EffectUpdate, SubtreeFlags: 0)
	//     └── Child2 (Flags: 0, SubtreeFlags: EffectDeletion)
	SubtreeFlags EffectFlag

	// === Priority ===
	// Lanes representing work priority
	Lanes Lane
	// ChildLanes for pending work in children
	ChildLanes Lane

	// === Key for Diff ===
	// Key is the DiffKey used for reconciling lists - this is the primary key for diffing
	// DiffKey is copied from VNode.Key() and should NOT be generated from Path
	DiffKey string

	// Key is an alias for DiffKey for backward compatibility
	// TODO: Gradually migrate all code to use DiffKey
	Key string

	// ✨ NodeID for stable runtime identity
	// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
	NodeID uint64

	// ✨ Path for automatic key generation (Mixed Strategy)
	// Full path from root: /root/base[0]/vstack[0]/panel[0]
	// Used for generating automatic keys for static UI components
	Path string

	// ✨ PathSegment is the last segment of the path: panel[0]
	// Used for debugging and quick type identification
	PathSegment string

	// ✨ SiblingIndex is the position in the parent's children list
	// Used for calculating the path index
	SiblingIndex int

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

// =============================================================================
// Fiber Methods
// =============================================================================

// GetProps returns the current props from the VNode.
// This method should be used instead of accessing the Props field directly,
// as the Props field is only a snapshot taken when the Fiber was created.
//
// The Props field may become stale if the VNode's props are updated after
// the Fiber is created. Use GetProps() to always get the current props.
func (f *Fiber) GetProps() Props {
	if f.VNode == nil {
		return nil
	}
	return f.VNode.Props()
}

// GetMemoizedProps returns the memoized props for comparison during reconciliation.
// These are the props from the previous render.
func (f *Fiber) GetMemoizedProps() Props {
	return f.MemoizedProps
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

	return fmt.Sprintf("Fiber{Tag: %s, Key: %s, NodeID: %d, Flags: %d, Lanes: %d}",
		f.Tag, f.Key, f.NodeID, f.Flags, f.Lanes)
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

	// Mark fiber as having work - use update's lane if specified, otherwise default to SyncLane
	lane := update.Lane
	if lane == LaneNoLane {
		lane = LaneSyncLane
	}
	f.MarkUpdate(lane)
}

// =============================================================================
// Reconciler Interface
// =============================================================================
// This interface allows internal/render to use Fiber reconciler without
// importing internal/reconciler directly (which would cause a cycle).

// Reconciler is the interface for Fiber reconciliation
type Reconciler interface {
	// Render executes the rendering process
	// ctx and buffer are passed as interface{} to avoid import cycles
	Render(ctx interface{}, buffer interface{}, renderFunc func() VNode)
	// SetApp sets the framework app for scheduling
	SetApp(app interface{})
	// GetRenderedRoot returns the rendered VNode tree for focus management
	GetRenderedRoot() VNode
}
