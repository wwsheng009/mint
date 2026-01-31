package ui

// PatchType represents the type of patch operation
type PatchType int

const (
	// PatchCreate creates a new node
	PatchCreate PatchType = iota
	// PatchUpdate updates an existing node
	PatchUpdate
	// PatchDelete deletes a node
	PatchDelete
	// PatchReplace replaces a node with a different type
	PatchReplace
	// PatchNone no change
	PatchNone
)

// Patch represents a change to be applied
type Patch struct {
	Type     PatchType
	Old      VNode
	New      VNode
	PropsDiff PropsDiff
}

// PropsDiff represents property changes
type PropsDiff struct {
	Added   Props
	Updated Props
	Removed []string
}

// Diff computes the difference between two VNode trees
func Diff(old, new VNode) Patch {
	// Case 1: Creation
	if old == nil && new != nil {
		return Patch{Type: PatchCreate, New: new}
	}

	// Case 2: Deletion
	if old != nil && new == nil {
		return Patch{Type: PatchDelete, Old: old}
	}

	// Case 3: Type change - replace
	if old.Type() != new.Type() {
		return Patch{Type: PatchReplace, Old: old, New: new}
	}

	// Case 4: Key change - replace
	if old.Key() != "" && new.Key() != "" && old.Key() != new.Key() {
		return Patch{Type: PatchReplace, Old: old, New: new}
	}

	// Case 5: Same type - check for updates
	switch old.Type() {
	case VNodeText:
		return diffText(old.(*TextVNode), new.(*TextVNode))
	case VNodeElement:
		return diffElement(old, new)
	case VNodeComponent:
		return diffComponent(old.(*ComponentVNode), new.(*ComponentVNode))
	case VNodeFragment:
		return diffFragment(old.(*FragmentVNode), new.(*FragmentVNode))
	default:
		return Patch{Type: PatchNone}
	}
}

// diffText diffs two text nodes
func diffText(old, new *TextVNode) Patch {
	if old.Content() == new.Content() {
		return Patch{Type: PatchNone}
	}
	return Patch{
		Type: PatchUpdate,
		Old:  old,
		New:  new,
	}
}

// diffElement diffs two element nodes
func diffElement(old, new VNode) Patch {
	// Check for tag change (shouldn't happen with Type check, but just in case)
	if oldEl, ok := old.(*ElementVNode); ok {
		if newEl, ok := new.(*ElementVNode); ok {
			if oldEl.Tag() != newEl.Tag() {
				return Patch{Type: PatchReplace, Old: old, New: new}
			}
		}
	}

	// Check props
	propsDiff := diffProps(old.Props(), new.Props())

	// Check style
	styleChanged := !stylesEqual(old.Style(), new.Style())

	if propsDiff.IsEmpty() && !styleChanged {
		// Check children
		return diffChildren(old, new)
	}

	return Patch{
		Type:     PatchUpdate,
		Old:      old,
		New:      new,
		PropsDiff: propsDiff,
	}
}

// diffComponent diffs two component nodes
func diffComponent(old, new *ComponentVNode) Patch {
	// Components should always re-render
	return Patch{
		Type: PatchUpdate,
		Old:  old,
		New:  new,
	}
}

// diffFragment diffs two fragment nodes
func diffFragment(old, new *FragmentVNode) Patch {
	// Fragments just diff their children
	return diffChildren(old, new)
}

// diffProps diffs two Props maps
func diffProps(old, new Props) PropsDiff {
	added := make(Props)
	updated := make(Props)
	removed := make([]string, 0)

	// Check for additions and updates
	for key, newVal := range new {
		oldVal, exists := old[key]
		if !exists {
			added[key] = newVal
		} else if !valuesEqual(oldVal, newVal) {
			updated[key] = newVal
		}
	}

	// Check for removals
	for key := range old {
		if _, exists := new[key]; !exists {
			removed = append(removed, key)
		}
	}

	return PropsDiff{
		Added:   added,
		Updated: updated,
		Removed: removed,
	}
}

// diffChildren diffs children arrays
func diffChildren(old, new VNode) Patch {
	oldChildren := old.Children()
	newChildren := new.Children()

	// Simple length check
	if len(oldChildren) != len(newChildren) {
		return Patch{
			Type: PatchUpdate,
			Old:  old,
			New:  new,
		}
	}

	// Check each child
	for i := 0; i < len(oldChildren); i++ {
		patch := Diff(oldChildren[i], newChildren[i])
		if patch.Type != PatchNone {
			return Patch{
				Type: PatchUpdate,
				Old:  old,
				New:  new,
			}
		}
	}

	return Patch{Type: PatchNone}
}

// IsEmpty returns true if the PropsDiff has no changes
func (d PropsDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Updated) == 0 && len(d.Removed) == 0
}

// valuesEqual compares two values for equality
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a == b
}

// stylesEqual compares two Style structs for equality
func stylesEqual(a, b interface{}) bool {
	// For MVP, assume styles are always different
	// A proper implementation would compare each field
	return false
}
