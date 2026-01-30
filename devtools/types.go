// Package devtools provides the TUI DevTools implementation.
//
// This package implements an observability and debugging system for the Mint TUI Runtime.
// It uses an incremental delta model to track changes with minimal performance impact.
package devtools

// NodeID is a unique identifier for a layout node.
type NodeID string

// MutationID is a unique identifier for a mutation event.
type MutationID uint64

// EventID is a unique identifier for an event.
type EventID uint64

// FrameID is a unique identifier for a frame.
type FrameID int

// LayoutID is a unique identifier for a layout change in the causal graph.
type LayoutID uint64

// RepaintID is a unique identifier for a repaint operation in the causal graph.
type RepaintID uint64

// ChangeMask is a bitmask indicating which fields of a node have changed.
type ChangeMask uint8

const (
	// ChangeRect indicates the node's position or size changed.
	ChangeRect ChangeMask = 1 << iota
	// ChangeZ indicates the node's Z-Index changed.
	ChangeZ
	// ChangeVisibility indicates the node's visibility changed.
	ChangeVisibility
	// ChangeFlex indicates the node's flex properties changed.
	ChangeFlex
	// ChangeProps indicates the node's props changed.
	ChangeProps
	// ChangeStyle indicates the node's style changed.
	ChangeStyle
	// ChangeContent indicates the node's content changed.
	ChangeContent
)

// Rect represents a rectangular area.
type Rect struct {
	X, Y, Width, Height int
}

// NodeDelta represents changes to a single node.
type NodeDelta struct {
	ID   NodeID
	Mask ChangeMask

	// Only changed fields are populated
	Rect       *Rect
	ZIndex     *int
	Visible    *bool
	FlexGrow   *float32
	FlexShrink *float32
	Props      map[string]interface{}
}

// LayoutDelta represents incremental layout changes for a frame.
type LayoutDelta struct {
	FrameID FrameID

	// Added nodes that didn't exist in the previous frame
	Added []NodeID

	// Removed nodes that no longer exist
	Removed []NodeID

	// Changed nodes with their specific changes
	Changed []NodeDelta

	// Metrics (optional, computed lazily)
	Metrics *LayoutMetrics
}

// LayoutMetrics contains metrics about the layout.
type LayoutMetrics struct {
	TotalNodes    int
	ChangedNodes  int
	AddedNodes    int
	RemovedNodes  int
	DirtyNodes    int
	CacheHits     int
	CacheMisses   int
	LayoutTimeMS  float64
}

// RepaintDelta represents incremental repaint changes for a frame.
type RepaintDelta struct {
	FrameID FrameID

	// Dirty regions that need repainting
	DirtyRegions []Rect

	// Statistics
	ChangedCells int
	TotalCells   int
}

// EventEntry represents a single event in the event trace.
type EventEntry struct {
	Type     string    // Event type (e.g., "keypress", "click")
	Target   NodeID    // Target node ID
	Phase    string    // Event phase (e.g., "capture", "bubble")
	Time     int64     // Event timestamp (nanoseconds)
	Data     map[string]interface{} // Additional event data
}

// EventDelta represents events that occurred in a frame.
type EventDelta struct {
	FrameID FrameID

	// Events that occurred in this frame
	Events []EventEntry

	// Causal link to mutations caused by these events
	CausedMutations []MutationID
}

// MutationKind represents the type of mutation.
type MutationKind uint8

const (
	// MutationState indicates a state change.
	MutationState MutationKind = iota
	// MutationProp indicates a prop change.
	MutationProp
	// MutationStyle indicates a style change.
	MutationStyle
	// MutationFocus indicates a focus change.
	MutationFocus
	// MutationSelection indicates a selection change.
	MutationSelection
)

// MutationRecord represents a state change (lightweight, no allocations in hot path).
// Designed for lock-free ring buffer storage.
type MutationRecord struct {
	ComponentID uint32 // Pre-allocated component ID
	FieldID     uint16 // Pre-allocated field ID
	Kind        uint8  // MutationKind
	_           uint8  // Padding for alignment (makes struct exactly 24 bytes)
	OldValue    uint64 // Old value (encoded)
	NewValue    uint64 // New value (encoded)
}

// MutationNode represents a mutation in the causal graph.
type MutationNode struct {
	ID        MutationID
	Component string
	Kind      MutationKind
	Field     string
	OldValue  interface{}
	NewValue  interface{}
}

// EventNode represents an event in the causal graph.
type EventNode struct {
	ID       EventID
	Type     string
	TargetID NodeID
	Phase    string
}

// EdgeType represents the type of causal relationship.
type EdgeType uint8

const (
	// EdgeEventToMutation indicates an event caused a mutation.
	EdgeEventToMutation EdgeType = iota
	// EdgeMutationToLayout indicates a mutation caused a layout change.
	EdgeMutationToLayout
	// EdgeLayoutToRepaint indicates a layout caused a repaint.
	EdgeLayoutToRepaint
)

// Edge represents a causal relationship between nodes.
type Edge struct {
	From uint64
	To   uint64
	Type EdgeType
}

// FrameRecord represents a complete record of a frame including all causal relationships.
type FrameRecord struct {
	FrameID      FrameID
	Time         int64

	// Input
	Events       []*EventNode

	// Intermediate state changes
	Mutations    []*MutationNode

	// Output
	LayoutDelta  *LayoutDelta
	RepaintDelta *RepaintDelta

	// Causal edges
	Edges        []Edge
}

// DebugMessage is a message sent from runtime to DevTools client.
type DebugMessage struct {
	Type    MessageType
	Payload interface{}
}

// MessageType represents the type of debug message.
type MessageType uint8

const (
	// MsgLayoutDelta indicates a layout delta message.
	MsgLayoutDelta MessageType = iota
	// MsgRepaintDelta indicates a repaint delta message.
	MsgRepaintDelta
	// MsgEventDelta indicates an event delta message.
	MsgEventDelta
	// MsgFrameTimeline indicates a frame timeline message.
	MsgFrameTimeline
	// MsgMutation indicates a mutation message.
	MsgMutation
	// MsgError indicates an error message.
	MsgError
)

// LayoutDebugInfo represents debug information about a layout node.
type LayoutDebugInfo struct {
	ID            string
	Type          string
	X, Y          int
	Width, Height int
	Visible       bool
	ZIndex        int
	Children      []LayoutDebugInfo
}

// LayoutDebugView is an interface for accessing layout debug information.
// This interface will be implemented by runtime.LayoutResult when debug views are added.
type LayoutDebugView interface {
	// ForEachBox iterates over all layout boxes and calls the provided function.
	ForEachBox(fn func(LayoutDebugInfo))
	// GetBoxInfo returns debug information for a specific box.
	GetBoxInfo(nodeID string) *LayoutDebugInfo
}

// Config contains DevTools configuration options.
type Config struct {
	// BufferSize is the size of the event bus buffer.
	BufferSize int
	// EnableOverlay enables the debug overlay.
	EnableOverlay bool
	// EnableMutationTap enables the mutation tap.
	EnableMutationTap bool
	// EnableTimeline enables frame timeline tracking.
	EnableTimeline bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BufferSize:        4096,
		EnableOverlay:     true,
		EnableMutationTap: true,
		EnableTimeline:    true,
	}
}
