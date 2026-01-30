// Package snapshot provides state snapshot and restoration for DevTools.
package snapshot

import (
	"fmt"
	"reflect"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Snapshot Differ
// =============================================================================

// Differ computes differences between snapshots.
type Differ struct {
	ignoreProps    []string
	ignoreState    []string
	compareStyle   bool
	compareBounds  bool
}

// NewDiffer creates a new snapshot differ.
func NewDiffer() *Differ {
	return &Differ{
		compareStyle:  true,
		compareBounds: true,
	}
}

// SetIgnoreProps sets properties to ignore during comparison.
func (d *Differ) SetIgnoreProps(props []string) {
	d.ignoreProps = props
}

// SetIgnoreState sets state fields to ignore during comparison.
func (d *Differ) SetIgnoreState(fields []string) {
	d.ignoreState = fields
}

// SetCompareStyle sets whether to compare style.
func (d *Differ) SetCompareStyle(compare bool) {
	d.compareStyle = compare
}

// SetCompareBounds sets whether to compare bounds.
func (d *Differ) SetCompareBounds(compare bool) {
	d.compareBounds = compare
}

// Compare computes the difference between two snapshots.
func (d *Differ) Compare(from, to *Snapshot) *SnapshotDiff {
	if from == nil || to == nil {
		return &SnapshotDiff{
			Timestamp: time.Now(),
			Changes:   []StateChange{},
			Summary:   DiffSummary{},
		}
	}

	var fromID, toID SnapshotID
	if from != nil {
		fromID = from.ID
	}
	if to != nil {
		toID = to.ID
	}

	diff := &SnapshotDiff{
		FromID:    fromID,
		ToID:      toID,
		Timestamp: time.Now(),
		Changes:   make([]StateChange, 0),
		Summary:   DiffSummary{},
	}

	// Track all seen nodes
	seenNodes := make(map[devtools.NodeID]bool)

	// Check for additions and modifications in "to" snapshot
	for nodeID, toState := range to.States {
		seenNodes[nodeID] = true

		fromState, exists := from.States[nodeID]
		if !exists {
			// Component added
			diff.Summary.ComponentsAdded++
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeAdded,
				OldValue:   nil,
				NewValue:   toState,
			})
			continue
		}

		// Compare states
		d.compareStates(nodeID, fromState, toState, diff)
	}

	// Check for removals in "from" snapshot
	for nodeID, fromState := range from.States {
		if !seenNodes[nodeID] {
			// Component removed
			diff.Summary.ComponentsRemoved++
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeRemoved,
				OldValue:   fromState,
				NewValue:   nil,
			})
		}
	}

	return diff
}

// compareStates compares two component states.
func (d *Differ) compareStates(nodeID devtools.NodeID, from, to *ComponentState, diff *SnapshotDiff) {
	modified := false

	// Compare type
	if from.Type != to.Type {
		modified = true
		diff.Changes = append(diff.Changes, StateChange{
			NodeID:     nodeID,
			ChangeType: ChangeModified,
			Path:       "type",
			OldValue:   from.Type,
			NewValue:   to.Type,
		})
	}

	// Compare visibility
	if from.Visible != to.Visible {
		modified = true
		diff.Changes = append(diff.Changes, StateChange{
			NodeID:     nodeID,
			ChangeType: ChangeModified,
			Path:       "visible",
			OldValue:   from.Visible,
			NewValue:   to.Visible,
		})
	}

	// Compare focus
	if from.Focused != to.Focused {
		modified = true
		diff.Changes = append(diff.Changes, StateChange{
			NodeID:     nodeID,
			ChangeType: ChangeModified,
			Path:       "focused",
			OldValue:   from.Focused,
			NewValue:   to.Focused,
		})
	}

	// Compare bounds
	if d.compareBounds {
		if from.Bounds != to.Bounds {
			modified = true
			diff.Summary.BoundsChanged++
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeModified,
				Path:       "bounds",
				OldValue:   from.Bounds,
				NewValue:   to.Bounds,
			})
		}
	}

	// Compare style
	if d.compareStyle {
		if from.Style != to.Style {
			modified = true
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeModified,
				Path:       "style",
				OldValue:   from.Style,
				NewValue:   to.Style,
			})
		}
	}

	// Compare props
	d.compareMaps(nodeID, from.Props, to.Props, "props", d.ignoreProps, diff)

	// Compare state
	d.compareMaps(nodeID, from.State, to.State, "state", d.ignoreState, diff)

	if modified {
		diff.Summary.ComponentsModified++
	}
}

// compareMaps compares two maps (props or state).
func (d *Differ) compareMaps(nodeID devtools.NodeID, from, to map[string]interface{}, path string, ignore []string, diff *SnapshotDiff) {
	if from == nil && to == nil {
		return
	}

	// Initialize nil maps
	if from == nil {
		from = make(map[string]interface{})
	}
	if to == nil {
		to = make(map[string]interface{})
	}

	// Build ignore set
	ignoreSet := make(map[string]bool)
	for _, key := range ignore {
		ignoreSet[key] = true
	}

	// Check for additions and modifications
	for key, toValue := range to {
		if ignoreSet[key] {
			continue
		}

		fromValue, exists := from[key]
		if !exists {
			// Key added
			if path == "props" {
				diff.Summary.PropsChanged++
			} else {
				diff.Summary.StateChanged++
			}
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeModified,
				Path:       fmt.Sprintf("%s.%s", path, key),
				OldValue:   nil,
				NewValue:   toValue,
			})
			continue
		}

		// Compare values
		if !d.valuesEqual(fromValue, toValue) {
			if path == "props" {
				diff.Summary.PropsChanged++
			} else {
				diff.Summary.StateChanged++
			}
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeModified,
				Path:       fmt.Sprintf("%s.%s", path, key),
				OldValue:   fromValue,
				NewValue:   toValue,
			})
		}
	}

	// Check for removals
	for key := range from {
		if ignoreSet[key] {
			continue
		}

		if _, exists := to[key]; !exists {
			// Key removed
			if path == "props" {
				diff.Summary.PropsChanged++
			} else {
				diff.Summary.StateChanged++
			}
			diff.Changes = append(diff.Changes, StateChange{
				NodeID:     nodeID,
				ChangeType: ChangeModified,
				Path:       fmt.Sprintf("%s.%s", path, key),
				OldValue:   from[key],
				NewValue:   nil,
			})
		}
	}
}

// valuesEqual compares two values for equality.
func (d *Differ) valuesEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// =============================================================================
// Convenience Functions
// =============================================================================

// CompareSnapshots is a convenience function to compare two snapshots.
func CompareSnapshots(from, to *Snapshot) *SnapshotDiff {
	differ := NewDiffer()
	return differ.Compare(from, to)
}

// HasChanges returns true if the diff has any changes.
func (d *SnapshotDiff) HasChanges() bool {
	return len(d.Changes) > 0
}

// GetChangesByNode returns all changes for a specific node.
func (d *SnapshotDiff) GetChangesByNode(nodeID devtools.NodeID) []StateChange {
	changes := make([]StateChange, 0)
	for _, change := range d.Changes {
		if change.NodeID == nodeID {
			changes = append(changes, change)
		}
	}
	return changes
}

// GetChangesByType returns all changes of a specific type.
func (d *SnapshotDiff) GetChangesByType(changeType ChangeType) []StateChange {
	changes := make([]StateChange, 0)
	for _, change := range d.Changes {
		if change.ChangeType == changeType {
			changes = append(changes, change)
		}
	}
	return changes
}

// FormatChanges returns a human-readable format of the changes.
func (d *SnapshotDiff) FormatChanges() []string {
	lines := make([]string, 0, len(d.Changes))

	for _, change := range d.Changes {
		switch change.ChangeType {
		case ChangeAdded:
			lines = append(lines, fmt.Sprintf("[+] %s: component added", change.NodeID))
		case ChangeRemoved:
			lines = append(lines, fmt.Sprintf("[-] %s: component removed", change.NodeID))
		case ChangeModified:
			lines = append(lines, fmt.Sprintf("[~] %s.%s: %v -> %v",
				change.NodeID, change.Path, change.OldValue, change.NewValue))
		case ChangeMoved:
			lines = append(lines, fmt.Sprintf("[->] %s: moved", change.NodeID))
		}
	}

	return lines
}

// FormatSummary returns a human-readable summary of the diff.
func (d *SnapshotDiff) FormatSummary() string {
	return fmt.Sprintf(
		"Added: %d, Removed: %d, Modified: %d | Props: %d, State: %d, Bounds: %d",
		d.Summary.ComponentsAdded,
		d.Summary.ComponentsRemoved,
		d.Summary.ComponentsModified,
		d.Summary.PropsChanged,
		d.Summary.StateChanged,
		d.Summary.BoundsChanged,
	)
}

// =============================================================================
// Time Travel
// =============================================================================

// TimeTravelRange computes a sequence of diffs for a range of frames.
type TimeTravelRange struct {
	diffs    []*SnapshotDiff
	snapshots []*Snapshot
}

// NewTimeTravelRange creates a new time travel range.
func NewTimeTravelRange(snapshots []*Snapshot) *TimeTravelRange {
	return &TimeTravelRange{
		snapshots: snapshots,
		diffs:     make([]*SnapshotDiff, 0),
	}
}

// Compute computes all diffs between consecutive snapshots.
func (t *TimeTravelRange) Compute() *TimeTravelRange {
	differ := NewDiffer()

	for i := 1; i < len(t.snapshots); i++ {
		diff := differ.Compare(t.snapshots[i-1], t.snapshots[i])
		t.diffs = append(t.diffs, diff)
	}

	return t
}

// GetDiffAt returns the diff at a specific index.
func (t *TimeTravelRange) GetDiffAt(index int) (*SnapshotDiff, bool) {
	if index < 0 || index >= len(t.diffs) {
		return nil, false
	}
	return t.diffs[index], true
}

// GetAllDiffs returns all diffs.
func (t *TimeTravelRange) GetAllDiffs() []*SnapshotDiff {
	return t.diffs
}

// FindFrameWithChange finds the first frame where a specific node changed.
func (t *TimeTravelRange) FindFrameWithChange(nodeID devtools.NodeID) (int, bool) {
	for i, diff := range t.diffs {
		for _, change := range diff.Changes {
			if change.NodeID == nodeID {
				return i, true
			}
		}
	}
	return 0, false
}

// GetChangeHistory returns the change history for a specific node.
func (t *TimeTravelRange) GetChangeHistory(nodeID devtools.NodeID) []StateChange {
	history := make([]StateChange, 0)

	for _, diff := range t.diffs {
		for _, change := range diff.Changes {
			if change.NodeID == nodeID {
				history = append(history, change)
			}
		}
	}

	return history
}
