// Package snapshot provides state snapshot and restoration for DevTools.
package snapshot

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Snapshot Data Structures
// =============================================================================

// Snapshot represents a complete application state at a specific frame.
type Snapshot struct {
	ID        SnapshotID          `json:"id"`
	FrameID   devtools.FrameID    `json:"frame_id"`
	Timestamp time.Time           `json:"timestamp"`
	Metadata  SnapshotMetadata    `json:"metadata"`
	States    map[devtools.NodeID]*ComponentState `json:"states"`
	Global    GlobalState         `json:"global"`
}

// SnapshotID is a unique identifier for a snapshot.
type SnapshotID string

// SnapshotMetadata contains metadata about the snapshot.
type SnapshotMetadata struct {
	FramesSinceLast int           `json:"frames_since_last"`
	MutationsCount  int           `json:"mutations_count"`
	LayoutsCount    int           `json:"layouts_count"`
	RepaintsCount   int           `json:"repaints_count"`
	Tags            []string      `json:"tags,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// ComponentState represents the state of a single component.
type ComponentState struct {
	NodeID       devtools.NodeID `json:"node_id"`
	Type         string          `json:"type"`
	Props        map[string]interface{} `json:"props,omitempty"`
	State        map[string]interface{} `json:"state,omitempty"`
	Children     []devtools.NodeID `json:"children,omitempty"`
	Bounds       Rect            `json:"bounds"`
	Visible      bool            `json:"visible"`
	Focused      bool            `json:"focused"`
	Disabled     bool            `json:"disabled"`
	Style        StyleState      `json:"style,omitempty"`
}

// Rect represents a rectangular area.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// StyleState represents style information.
type StyleState struct {
	Foreground string `json:"foreground,omitempty"`
	Background string `json:"background,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Italic     bool   `json:"italic,omitempty"`
	Underline  bool   `json:"underline,omitempty"`
}

// GlobalState represents application-wide state.
type GlobalState struct {
	WindowSize   WindowSize     `json:"window_size"`
	Theme        string         `json:"theme,omitempty"`
	CustomData   map[string]interface{} `json:"custom_data,omitempty"`
}

// WindowSize represents the terminal window size.
type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// =============================================================================
// Snapshot Diff
// =============================================================================

// SnapshotDiff represents the difference between two snapshots.
type SnapshotDiff struct {
	FromID     SnapshotID       `json:"from_id"`
	ToID       SnapshotID       `json:"to_id"`
	Timestamp  time.Time        `json:"timestamp"`
	Changes    []StateChange    `json:"changes"`
	Summary    DiffSummary      `json:"summary"`
}

// StateChange represents a single state change.
type StateChange struct {
	NodeID     devtools.NodeID `json:"node_id"`
	ChangeType ChangeType      `json:"change_type"`
	Path       string          `json:"path,omitempty"`
	OldValue   interface{}      `json:"old_value,omitempty"`
	NewValue   interface{}      `json:"new_value,omitempty"`
}

// ChangeType represents the type of state change.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
	ChangeModified
	ChangeMoved
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeAdded:
		return "added"
	case ChangeRemoved:
		return "removed"
	case ChangeModified:
		return "modified"
	case ChangeMoved:
		return "moved"
	default:
		return "unknown"
	}
}

// DiffSummary provides a summary of changes.
type DiffSummary struct {
	ComponentsAdded    int `json:"components_added"`
	ComponentsRemoved  int `json:"components_removed"`
	ComponentsModified int `json:"components_modified"`
	PropsChanged       int `json:"props_changed"`
	StateChanged       int `json:"state_changed"`
	BoundsChanged      int `json:"bounds_changed"`
}

// =============================================================================
// Snapshot Pool for memory management
// =============================================================================

// SnapshotPool manages a pool of reusable snapshots.
type SnapshotPool struct {
	mu       sync.Mutex
	pool     []*Snapshot
	maxSize  int
	created  int
	reused   int
}

// NewSnapshotPool creates a new snapshot pool.
func NewSnapshotPool(maxSize int) *SnapshotPool {
	return &SnapshotPool{
		pool:    make([]*Snapshot, 0, maxSize),
		maxSize: maxSize,
	}
}

// Acquire gets a snapshot from the pool or creates a new one.
func (p *SnapshotPool) Acquire() *Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pool) > 0 {
		// Reuse from pool
		idx := len(p.pool) - 1
		snap := p.pool[idx]
		p.pool = p.pool[:idx]
		p.reused++
		return snap
	}

	// Create new
	p.created++
	return &Snapshot{
		States: make(map[devtools.NodeID]*ComponentState),
		Global: GlobalState{
			CustomData: make(map[string]interface{}),
		},
		Metadata: SnapshotMetadata{
			Labels: make(map[string]string),
		},
	}
}

// Release returns a snapshot to the pool for reuse.
func (p *SnapshotPool) Release(snap *Snapshot) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pool) >= p.maxSize {
		return // Pool full, discard
	}

	// Clear snapshot for reuse
	snap.States = make(map[devtools.NodeID]*ComponentState)
	snap.Global.CustomData = make(map[string]interface{})
	snap.Metadata.Tags = nil
	snap.Metadata.Labels = make(map[string]string)

	p.pool = append(p.pool, snap)
}

// Stats returns pool statistics.
func (p *SnapshotPool) Stats() (created, reused int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.created, p.reused
}

// =============================================================================
// Snapshot Serialization
// =============================================================================

// Serialize converts a snapshot to JSON bytes.
func (s *Snapshot) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize creates a snapshot from JSON bytes.
func DeserializeSnapshot(data []byte) (*Snapshot, error) {
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// =============================================================================
// Snapshot Builder
// =============================================================================

// Builder helps construct snapshots incrementally.
type Builder struct {
	snapshot *Snapshot
}

// NewBuilder creates a new snapshot builder.
func NewBuilder(id SnapshotID, frameID devtools.FrameID) *Builder {
	pool := NewSnapshotPool(100)
	snap := pool.Acquire()
	snap.ID = id
	snap.FrameID = frameID
	snap.Timestamp = time.Now()

	return &Builder{
		snapshot: snap,
	}
}

// SetWindowSize sets the window size in the snapshot.
func (b *Builder) SetWindowSize(width, height int) *Builder {
	b.snapshot.Global.WindowSize = WindowSize{
		Width:  width,
		Height: height,
	}
	return b
}

// SetTheme sets the theme in the snapshot.
func (b *Builder) SetTheme(theme string) *Builder {
	b.snapshot.Global.Theme = theme
	return b
}

// AddComponent adds a component state to the snapshot.
func (b *Builder) AddComponent(state *ComponentState) *Builder {
	b.snapshot.States[state.NodeID] = state
	return b
}

// SetMetadata sets the snapshot metadata.
func (b *Builder) SetMetadata(metadata SnapshotMetadata) *Builder {
	b.snapshot.Metadata = metadata
	return b
}

// AddTag adds a tag to the snapshot.
func (b *Builder) AddTag(tag string) *Builder {
	if b.snapshot.Metadata.Tags == nil {
		b.snapshot.Metadata.Tags = make([]string, 0)
	}
	b.snapshot.Metadata.Tags = append(b.snapshot.Metadata.Tags, tag)
	return b
}

// SetLabel sets a label in the snapshot metadata.
func (b *Builder) SetLabel(key, value string) *Builder {
	if b.snapshot.Metadata.Labels == nil {
		b.snapshot.Metadata.Labels = make(map[string]string)
	}
	b.snapshot.Metadata.Labels[key] = value
	return b
}

// Build returns the completed snapshot.
func (b *Builder) Build() *Snapshot {
	return b.snapshot
}
