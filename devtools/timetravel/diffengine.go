// Package timetravel provides diff engine for time travel snapshots.
//
// This file implements the diff engine for comparing snapshots
// and computing efficient deltas.
package timetravel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wwsheng009/mint/devtools"
)

// DiffEngine computes and manages differences between snapshots.
type DiffEngine struct {
	// Cache for computed diffs
	diffCache map[string]*SnapshotDiff
}

// NewDiffEngine creates a new diff engine.
func NewDiffEngine() *DiffEngine {
	return &DiffEngine{
		diffCache: make(map[string]*SnapshotDiff),
	}
}

// Compute computes the diff between two snapshots.
func (de *DiffEngine) Compute(from, to *FrameSnapshot) *SnapshotDiff {
	if from == nil || to == nil {
		return nil
	}

	// Check cache
	cacheKey := de.cacheKey(from.FrameID, to.FrameID)
	if cached, exists := de.diffCache[cacheKey]; exists {
		return cached
	}

	diff := from.Diff(to)
	de.diffCache[cacheKey] = diff

	return diff
}

// cacheKey generates a cache key for two frame IDs.
func (de *DiffEngine) cacheKey(fromID, toID devtools.FrameID) string {
	return fmt.Sprintf("%d->%d", fromID, toID)
}

// ClearCache clears the diff cache.
func (de *DiffEngine) ClearCache() {
	de.diffCache = make(map[string]*SnapshotDiff)
}

// Delta represents a compact delta between states.
type Delta struct {
	Type     DeltaType
	Key      string
	OldValue interface{}
	NewValue interface{}
}

// DeltaType represents the type of delta.
type DeltaType int

const (
	// DeltaAdd represents an added value.
	DeltaAdd DeltaType = iota
	// DeltaRemove represents a removed value.
	DeltaRemove
	// DeltaChange represents a changed value.
	DeltaChange
	// DeltaMove represents a moved value.
	DeltaMove
)

// DeltaSet represents a set of deltas.
type DeltaSet struct {
	ComponentDeltas map[uint32][]ComponentDelta
	LayoutDeltas    []LayoutDeltaInfo
	RepaintDeltas   []RepaintDeltaInfo
}

// ComponentDelta represents a delta for a component.
type ComponentDelta struct {
	ComponentID   uint32
	ComponentName string
	Deltas        []FieldDelta
}

// FieldDelta represents a delta for a single field.
type FieldDelta struct {
	Field   string
	Type    DeltaType
	OldValue interface{}
	NewValue interface{}
}

// LayoutDeltaInfo represents layout delta information.
type LayoutDeltaInfo struct {
	NodeID     string
	ChangeMask devtools.ChangeMask
	OldRect    *devtools.Rect
	NewRect    *devtools.Rect
}

// RepaintDeltaInfo represents repaint delta information.
type RepaintDeltaInfo struct {
	DirtyRegions []devtools.Rect
	ChangedCells int
}

// ComputeDeltaSet computes a compact delta set from a snapshot diff.
func (de *DiffEngine) ComputeDeltaSet(diff *SnapshotDiff) *DeltaSet {
	if diff == nil {
		return nil
	}

	ds := &DeltaSet{
		ComponentDeltas: make(map[uint32][]ComponentDelta),
		LayoutDeltas:    make([]LayoutDeltaInfo, 0),
		RepaintDeltas:   make([]RepaintDeltaInfo, 0),
	}

	// Convert layout changes
	for _, lc := range diff.LayoutChanges {
		ds.LayoutDeltas = append(ds.LayoutDeltas, LayoutDeltaInfo{
			NodeID:     lc.NodeID,
			ChangeMask: lc.ChangeMask,
			OldRect:    lc.OldRect,
			NewRect:    lc.NewRect,
		})
	}

	// Convert component changes to field deltas
	for _, compDiff := range diff.ChangedComponents {
		for field, value := range compDiff.Changes.Added {
			ds.ComponentDeltas[compDiff.ComponentID] = append(
				ds.ComponentDeltas[compDiff.ComponentID],
				ComponentDelta{
					ComponentID: compDiff.ComponentID,
					Deltas: []FieldDelta{{
						Field:    field,
						Type:     DeltaAdd,
						NewValue: value,
					}},
				},
			)
		}

		for _, field := range compDiff.Changes.Removed {
			ds.ComponentDeltas[compDiff.ComponentID] = append(
				ds.ComponentDeltas[compDiff.ComponentID],
				ComponentDelta{
					ComponentID: compDiff.ComponentID,
					Deltas: []FieldDelta{{
						Field: field,
						Type:  DeltaRemove,
					}},
				},
			)
		}

		for field, valueChange := range compDiff.Changes.Modified {
			ds.ComponentDeltas[compDiff.ComponentID] = append(
				ds.ComponentDeltas[compDiff.ComponentID],
				ComponentDelta{
					ComponentID: compDiff.ComponentID,
					Deltas: []FieldDelta{{
						Field:    field,
						Type:     DeltaChange,
						OldValue: valueChange.OldValue,
						NewValue: valueChange.NewValue,
					}},
				},
			)
		}
	}

	return ds
}

// Apply applies a delta set to a snapshot.
func (ds *DeltaSet) Apply(snapshot *FrameSnapshot) {
	if snapshot == nil || ds == nil {
		return
	}

	// Apply component deltas
	for compID, deltas := range ds.ComponentDeltas {
		state := snapshot.ComponentStates[compID]
		if state == nil {
			continue
		}

		for _, delta := range deltas {
			for _, fieldDelta := range delta.Deltas {
				switch fieldDelta.Type {
				case DeltaAdd:
					if state.State == nil {
						state.State = make(map[string]interface{})
					}
					state.State[fieldDelta.Field] = fieldDelta.NewValue
				case DeltaRemove:
					if state.State != nil {
						delete(state.State, fieldDelta.Field)
					}
				case DeltaChange:
					if state.State != nil {
						state.State[fieldDelta.Field] = fieldDelta.NewValue
					}
				}
			}
		}
	}
}

// VisualDiff represents a visual diff for display.
type VisualDiff struct {
	Lines []DiffLine
}

// DiffLine represents a single line in a visual diff.
type DiffLine struct {
	Type    DiffLineType
	Content string
	LineNo  int
}

// DiffLineType represents the type of diff line.
type DiffLineType int

const (
	// DiffLineSame represents unchanged content.
	DiffLineSame DiffLineType = iota
	// DiffLineAdded represents added content.
	DiffLineAdded
	// DiffLineRemoved represents removed content.
	DiffLineRemoved
	// DiffLineChanged represents changed content.
	DiffLineChanged
)

// Visualize creates a visual diff from a delta set.
func (ds *DeltaSet) Visualize() *VisualDiff {
	vd := &VisualDiff{
		Lines: make([]DiffLine, 0),
	}

	// Add component deltas
	for _, deltas := range ds.ComponentDeltas {
		for _, delta := range deltas {
			for _, fieldDelta := range delta.Deltas {
				line := DiffLine{
					Content: fmt.Sprintf("%s.%s: %v -> %v",
						delta.ComponentName, fieldDelta.Field,
						fieldDelta.OldValue, fieldDelta.NewValue),
				}

				switch fieldDelta.Type {
				case DeltaAdd:
					line.Type = DiffLineAdded
					line.Content = "+ " + line.Content
				case DeltaRemove:
					line.Type = DiffLineRemoved
					line.Content = "- " + line.Content
				case DeltaChange:
					line.Type = DiffLineChanged
					line.Content = "~ " + line.Content
				default:
					line.Type = DiffLineSame
				}

				vd.Lines = append(vd.Lines, line)
			}
		}
	}

	return vd
}

// BufferDiff computes the diff between two buffer snapshots.
type BufferDiff struct {
	Added    []CellChange
	Removed  []CellChange
	Changed  []CellChange
}

// CellChange represents a change to a single cell.
type CellChange struct {
	Row     int
	Col     int
	OldChar rune
	NewChar rune
	OldStyle interface{}
	NewStyle interface{}
}

// ComputeBufferDiff computes the diff between two byte buffers.
func ComputeBufferDiff(oldBuf, newBuf []byte, width int) *BufferDiff {
	if oldBuf == nil {
		oldBuf = []byte{}
	}
	if newBuf == nil {
		newBuf = []byte{}
	}

	diff := &BufferDiff{
		Added:   make([]CellChange, 0),
		Removed: make([]CellChange, 0),
		Changed: make([]CellChange, 0),
	}

	oldLines := bytes.Split(oldBuf, []byte{'\n'})
	newLines := bytes.Split(newBuf, []byte{'\n'})

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for row := 0; row < maxLines; row++ {
		var oldLine, newLine []byte

		if row < len(oldLines) {
			oldLine = oldLines[row]
		}
		if row < len(newLines) {
			newLine = newLines[row]
		}

		// Compare cells
		maxCols := len(oldLine)
		if len(newLine) > maxCols {
			maxCols = len(newLine)
		}

		for col := 0; col < maxCols; col++ {
			var oldChar, newChar byte

			if col < len(oldLine) {
				oldChar = oldLine[col]
			}
			if col < len(newLine) {
				newChar = newLine[col]
			}

			if oldChar != newChar {
				change := CellChange{
					Row:     row,
					Col:     col,
					OldChar: rune(oldChar),
					NewChar: rune(newChar),
				}

				if oldChar == 0 {
					diff.Added = append(diff.Added, change)
				} else if newChar == 0 {
					diff.Removed = append(diff.Removed, change)
				} else {
					diff.Changed = append(diff.Changed, change)
				}
			}
		}
	}

	return diff
}

// JSONDiff computes a JSON diff between two values.
func JSONDiff(oldValue, newValue interface{}) ([]Delta, error) {
	deltas := make([]Delta, 0)

	oldBytes, err := json.Marshal(oldValue)
	if err != nil {
		return nil, err
	}

	newBytes, err := json.Marshal(newValue)
	if err != nil {
		return nil, err
	}

	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal(oldBytes, &oldMap); err != nil {
		oldMap = nil
	}
	if err := json.Unmarshal(newBytes, &newMap); err != nil {
		newMap = nil
	}

	if oldMap == nil && newMap == nil {
		// Compare as primitives
		if !reflect.DeepEqual(oldValue, newValue) {
			deltas = append(deltas, Delta{
				Type:     DeltaChange,
				OldValue: oldValue,
				NewValue: newValue,
			})
		}
		return deltas, nil
	}

	// Compare maps
	if oldMap != nil {
		for key, oldVal := range oldMap {
			if newVal, exists := newMap[key]; exists {
				if !reflect.DeepEqual(oldVal, newVal) {
					deltas = append(deltas, Delta{
						Type:     DeltaChange,
						Key:      key,
						OldValue: oldVal,
						NewValue: newVal,
					})
				}
			} else {
				deltas = append(deltas, Delta{
					Type:     DeltaRemove,
					Key:      key,
					OldValue: oldVal,
				})
			}
		}
	}

	if newMap != nil {
		for key, newVal := range newMap {
			if oldMap == nil {
				deltas = append(deltas, Delta{
					Type:     DeltaAdd,
					Key:      key,
					NewValue: newVal,
				})
			} else if _, exists := oldMap[key]; !exists {
				deltas = append(deltas, Delta{
					Type:     DeltaAdd,
					Key:      key,
					NewValue: newVal,
				})
			}
		}
	}

	return deltas, nil
}

// Patch applies a set of deltas to a value.
func Patch(value interface{}, deltas []Delta) (interface{}, error) {
	result := make(map[string]interface{})

	// Convert value to map if possible
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			result[key] = val
		}
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, err
		}
	}

	// Apply deltas
	for _, delta := range deltas {
		switch delta.Type {
		case DeltaAdd, DeltaChange:
			result[delta.Key] = delta.NewValue
		case DeltaRemove:
			delete(result, delta.Key)
		}
	}

	return result, nil
}

// FormatDelta formats a delta for display.
func FormatDelta(delta Delta) string {
	switch delta.Type {
	case DeltaAdd:
		return fmt.Sprintf("+ %s: %v", delta.Key, delta.NewValue)
	case DeltaRemove:
		return fmt.Sprintf("- %s: %v", delta.Key, delta.OldValue)
	case DeltaChange:
		return fmt.Sprintf("~ %s: %v -> %v", delta.Key, delta.OldValue, delta.NewValue)
	default:
		return fmt.Sprintf("? %s", delta.Key)
	}
}

// SizeHint returns an estimate of the diff size.
func (ds *DeltaSet) SizeHint() int {
	size := 0

	for _, deltas := range ds.ComponentDeltas {
		for _, delta := range deltas {
			size += len(delta.Deltas)
		}
	}

	size += len(ds.LayoutDeltas)
	size += len(ds.RepaintDeltas)

	return size
}

// IsEmpty returns whether the delta set is empty.
func (ds *DeltaSet) IsEmpty() bool {
	if ds == nil {
		return true
	}

	return len(ds.ComponentDeltas) == 0 &&
		len(ds.LayoutDeltas) == 0 &&
		len(ds.RepaintDeltas) == 0
}

// GetComponentDeltas returns deltas for a specific component.
func (ds *DeltaSet) GetComponentDeltas(componentID uint32) []ComponentDelta {
	if ds == nil {
		return nil
	}
	return ds.ComponentDeltas[componentID]
}

// HasComponentChanges returns whether a component has changes.
func (ds *DeltaSet) HasComponentChanges(componentID uint32) bool {
	if ds == nil {
		return false
	}
	deltas, exists := ds.ComponentDeltas[componentID]
	return exists && len(deltas) > 0
}

// Merge merges another delta set into this one.
func (ds *DeltaSet) Merge(other *DeltaSet) {
	if ds == nil || other == nil {
		return
	}

	if ds.ComponentDeltas == nil {
		ds.ComponentDeltas = make(map[uint32][]ComponentDelta)
	}

	for compID, deltas := range other.ComponentDeltas {
		ds.ComponentDeltas[compID] = append(ds.ComponentDeltas[compID], deltas...)
	}

	ds.LayoutDeltas = append(ds.LayoutDeltas, other.LayoutDeltas...)
	ds.RepaintDeltas = append(ds.RepaintDeltas, other.RepaintDeltas...)
}
