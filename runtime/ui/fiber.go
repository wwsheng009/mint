package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// Fiber Architecture
// =============================================================================
// Fiber is a reconciliation algorithm, breaking rendering work into small units.
// Each Fiber node represents a component in the UI and forms a tree.
//
// Fiber-first Architecture:
//   - VNode is used only during Fiber creation, then discarded
//   - Fiber.Instance holds the runtime entity (persists across renders)
//   - All runtime state is in Instance, not in Fiber
//   - Paint phase uses Instance.Paint(), never accesses VNode
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

	// === Layout and Paint Dirty Flags (Fiber-first Architecture) ===
	// These flags support incremental layout and paint optimization

	// FlagLayoutDirty indicates the fiber needs layout recalculation
	// Used by layout.Dirtyable interface for incremental layout
	FlagLayoutDirty EffectFlag = 1 << 10

	// FlagPaintDirty indicates the fiber needs repaint
	// Used for incremental paint optimization
	FlagPaintDirty EffectFlag = 1 << 11
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
// Each Fiber corresponds to a VNode during creation, then holds Instance for runtime.
//
// Fiber-first Architecture (from FIBER_PAINT_ARCHITECTURE.md):
//   - VNode = Description (created every render, then discarded)
//   - Fiber = Tree structure, scheduling (runtime persistent)
//   - Instance = Behavior + State (runtime persistent)
type Fiber struct {
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
	Props Props
	// Memoized props (previous props for diffing)
	MemoizedProps Props
	// Memoized state serves different purposes based on VNode type:
	// - TextVNode: stores text content (set by completeWorkText)
	// - ComponentVNode with UpdateQueue: stores state for functional updates
	// - ComponentVNode with hooks: NOT used (state is in ComponentContext.Hooks)
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
	// Key is the DiffKey used for reconciling lists - this is the primary key for diffing
	// DiffKey is copied from VNode.Key() and should NOT be generated from Path
	DiffKey string

	// Key is an alias for DiffKey for backward compatibility
	Key string

	// === ID for business reference/positioning ===
	// ID is the business identifier for element reference/positioning
	// This is separate from Key which is used for diffing
	// ID is copied from VNode.ID()
	ID string

	// === Root Fiber Marker ===
	// IsRoot indicates if this fiber is the root component
	// Set during prepareFreshStack for the root component wrapper
	// Used by beginWorkComponent to provide global state context
	IsRoot bool

	// ✨ NodeID for stable runtime identity
	NodeID uint64

	// ✨ Layer specifies rendering layer (Z-order) for this fiber
	// Layers: Base(0) < Overlay(1) < Modal(2) < Tooltip(3) < Inspector(4)
	Layer Layer

	// ✨ Path for automatic key generation (Mixed Strategy)
	Path string

	// ✨ PathSegment is the last segment of the path
	PathSegment string

	// ✨ SiblingIndex is the position in the parent's children list
	SiblingIndex int

	// === Type ===
	// The type of this fiber (cached from VNode)
	Type VNodeType

	// === Tag for debugging ===
	// Component or element tag
	Tag string
	
	// === Layout Style (Phase 1) ===
	// These fields are populated in completeWork from VNode props
	LayoutDirection  Direction
	LayoutAlign      Align
	LayoutCrossAlign Align
	LayoutGap        int
	LayoutPadding    [4]int
	LayoutFlex       int

	// ✨ Border Style (方案 A - 边框作为容器属性)
	// Populated in completeWork from VNode props
	// All containers can now support borders natively
	BorderStyle  string  // "none", "single", "double", "rounded", "dashed"
	BorderLabel  string  // Optional label displayed on top border

	// ✨ Modal Centering (Phase 1.4)
	// Populated in completeWork from VNode props
	// Controls whether Modal should be centered in viewport
	Centered bool

	// ✨ Position Fixed (Phase 2.1)
	// Populated in completeWork from VNode props
	// Controls positioning scheme: Relative/Absolute/Fixed
	Position types.PositionType

	// ✨ Anchor (Phase 2.1)
	// Populated in completeWork from VNode props
	// Controls alignment for fixed/absolute positioning
	Anchor types.Anchor

	// ✨ Portal Root (Phase 3.1)
	// Specifies the target fiber where this node should be mounted during layout/render
	// Used by Portal components to mount children to a different location in the tree
	// nil means normal mounting (follow parent tree structure)
	PortalRoot *Fiber

	// === Visual Style (Fiber-first) ===
	// Copied from VNode.Style() during Fiber creation
	Style style.Style

	// === Component Instance (Fiber-first Architecture) ===
	// Instance is the runtime entity for this Fiber.
	// It persists across renders and holds all state (focus, hover, disabled, etc.)
	// VNode is used only during creation, then discarded.
	// During commit/paint phase, Instance.Paint() is called.
	//
	// Key Design Points:
	// 1. Fiber.Instance holds the persistent runtime entity
	// 2. Instance is created ONCE during Fiber creation via InstanceFactory
	// 3. CloneFiber REUSES Instance (Instance is NEVER cloned)
	// 4. Paint phase uses Instance.Paint() - NO VNode access
	// 5. Instance holds all runtime state (focus, hover, pressed, etc.)
	//
	// See: docs/fiber/fiber_first/FIBER_PAINT_ARCHITECTURE.md
	Instance ComponentInstance

	// === Special VNode Types Support ===
	// These fields store data from special VNode types
	ComponentFunc          ComponentFunc
	ComponentFuncWithProps ComponentFuncWithProps
	ComponentName          string
	ErrorBoundaryFunc      ComponentFunc
	ErrorBoundaryFallbackFiber *Fiber
	MemoCompare            PropsEqual
	MemoShouldUpdate       bool

	// === FocusableVNode (DEPRECATED - Use Instance.FocusableInstance instead) ===
	// ⚠️ DEPRECATED: This field will be removed in future versions.
	// In Fiber-first architecture, all state (focus, hover, disabled, etc.) is in ComponentInstance.
	// This field is kept ONLY for backward compatibility during migration.
	// New components should NOT use this field.
	// FocusableVNode FocusableVNode

	// === ActionTargetID (Fiber-first Action Architecture) ===
	// Fiber only stores ActionTargetID for routing to component.
	ActionTargetID string

	// === FocusableMeta (DEPRECATED - Use Instance.FocusableInstance instead) ===
	// ⚠️ DEPRECATED: This field will be removed in future versions.
	// Focusable info (TabIndex, Disabled, FocusID) is now managed by ComponentInstance.
	// This field is kept ONLY for backward compatibility during migration.
	// New components should check Instance.(FocusableInstance) instead.
	// FocusableMeta *FocusableMeta

	// === ComponentInstance (Legacy - Use Instance field instead) ===
	// DEPRECATED: This field is kept for backward compatibility during migration.
	// Use Instance field instead.
	// ComponentInstance ComponentInstance
}

// =============================================================================
// Fiber Methods
// =============================================================================

// GetActionTargetID returns the ActionTargetID for ActionBridge routing.
func (f *Fiber) GetActionTargetID() string {
	return f.ActionTargetID
}

// GetProps returns the props from the Fiber.
func (f *Fiber) GetProps() Props {
	return f.Props
}

// GetMemoizedProps returns the memoized props for comparison during reconciliation.
func (f *Fiber) GetMemoizedProps() Props {
	return f.MemoizedProps
}

// GetInstance returns the ComponentInstance for this Fiber.
// Returns nil if no instance is attached.
func (f *Fiber) GetInstance() ComponentInstance {
	// Priority: Instance field (new) > ComponentInstance field (legacy)
	return f.Instance
}

// GetPaintableInstance returns the PaintableInstance if available.
// Returns nil if the instance doesn't implement PaintableInstance.
func (f *Fiber) GetPaintableInstance() PaintableInstance {
	inst := f.GetInstance()
	if inst == nil {
		return nil
	}
	if pi, ok := inst.(PaintableInstance); ok {
		return pi
	}
	return nil
}

// GetFocusableInstance returns the FocusableInstance if available.
// Returns nil if the instance doesn't implement FocusableInstance.
func (f *Fiber) GetFocusableInstance() FocusableInstance {
	inst := f.GetInstance()
	if inst == nil {
		return nil
	}
	if fi, ok := inst.(FocusableInstance); ok {
		return fi
	}
	return nil
}

// HasInstance returns true if fiber has a runtime instance.
// This is useful for checking if the fiber has been fully initialized.
func (f *Fiber) HasInstance() bool {
	return f.Instance != nil
}

// HasStyle returns true if fiber has explicit width/height style defined.
// This is used by the layout engine to determine if explicit sizing is set.
func (f *Fiber) HasStyle() bool {
	// Check if explicit dimensions are set in style
	return f.Style.Width > 0 || f.Style.Height > 0
}

// IsLayoutDirty returns true if the fiber needs layout recalculation.
// This implements part of the layout.Dirtyable interface contract.
func (f *Fiber) IsLayoutDirty() bool {
	return f.Flags&FlagLayoutDirty != 0
}

// MarkLayoutDirty marks the fiber as needing layout recalculation.
func (f *Fiber) MarkLayoutDirty() {
	f.Flags |= FlagLayoutDirty
}

// ClearLayoutDirty clears the layout dirty flag.
func (f *Fiber) ClearLayoutDirty() {
	f.Flags &^= FlagLayoutDirty
}

// IsPaintDirty returns true if the fiber needs repaint.
func (f *Fiber) IsPaintDirty() bool {
	return f.Flags&FlagPaintDirty != 0
}

// MarkPaintDirty marks the fiber as needing repaint.
func (f *Fiber) MarkPaintDirty() {
	f.Flags |= FlagPaintDirty
}

// ClearPaintDirty clears the paint dirty flag.
func (f *Fiber) ClearPaintDirty() {
	f.Flags &^= FlagPaintDirty
}

// Update represents a state update
type Update struct {
	Payload interface{}
	Next    *Update
	Lane    Lane
	Callback func()
}

// UpdateQueue holds pending updates
type UpdateQueue struct {
	First *Update
	Last  *Update
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

	return fmt.Sprintf("Fiber{Tag: %s, Key: %s, NodeID: %d, Flags: %d, Lanes: %d, HasInstance: %v}",
		f.Tag, f.Key, f.NodeID, f.Flags, f.Lanes, f.Instance != nil)
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

	lane := update.Lane
	if lane == LaneNoLane {
		lane = LaneSyncLane
	}
	f.MarkUpdate(lane)
}

// =============================================================================
// Reconciler Interface
// =============================================================================

// Reconciler is the interface for Fiber reconciliation
type Reconciler interface {
	Render(ctx interface{}, buffer interface{}, renderFunc func() VNode)
	SetApp(app interface{})
	GetRenderedRoot() VNode
}
