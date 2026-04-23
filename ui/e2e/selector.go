package e2e

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
)

type locatorKind string

const (
	locatorComponentID locatorKind = "component_id"
	locatorAt          locatorKind = "at"
	locatorText        locatorKind = "text"
	locatorID          locatorKind = "id"
	locatorKey         locatorKind = "key"
	locatorTargetID    locatorKind = "target_id"
	locatorFocused     locatorKind = "focused"
	locatorTag         locatorKind = "tag"
)

// Locator identifies a target in the rendered app.
type Locator struct {
	kind  locatorKind
	value string
	x     int
	y     int
}

func At(x, y int) Locator {
	return Locator{kind: locatorAt, x: x, y: y}
}

func ByText(text string) Locator {
	return Locator{kind: locatorText, value: text}
}

func ByComponentID(componentID string) Locator {
	return Locator{kind: locatorComponentID, value: componentID}
}

func ByID(id string) Locator {
	return Locator{kind: locatorID, value: id}
}

func ByKey(key string) Locator {
	return Locator{kind: locatorKey, value: key}
}

func ByTargetID(targetID string) Locator {
	return Locator{kind: locatorTargetID, value: targetID}
}

func ByTag(tag string) Locator {
	return Locator{kind: locatorTag, value: tag}
}

func Focused() Locator {
	return Locator{kind: locatorFocused}
}

// FocusSnapshot captures the focused fiber's observable state.
type FocusSnapshot struct {
	Focused            bool
	Index              int
	Type               int
	NodeID             uint64
	ComponentID        string
	TargetID           string
	Key                string
	ID                 string
	Tag                string
	ComponentContextID string
	Bounds             layout.Rect
	FocusCount         int
}

// Equal reports whether two focus snapshots refer to the same focused element.
func (s FocusSnapshot) Equal(other FocusSnapshot) bool {
	return s.NodeID == other.NodeID &&
		s.ComponentID == other.ComponentID &&
		s.TargetID == other.TargetID &&
		s.Key == other.Key &&
		s.ID == other.ID &&
		s.Tag == other.Tag &&
		s.Index == other.Index
}

func (l Locator) matchFocus(snapshot FocusSnapshot) error {
	switch l.kind {
	case locatorFocused:
		return nil
	case locatorComponentID:
		if snapshot.ComponentID != l.value {
			return fmt.Errorf("focused componentID = %q, want %q", snapshot.ComponentID, l.value)
		}
	case locatorID:
		if snapshot.ID != l.value {
			return fmt.Errorf("focused id = %q, want %q", snapshot.ID, l.value)
		}
	case locatorKey:
		if snapshot.Key != l.value {
			return fmt.Errorf("focused key = %q, want %q", snapshot.Key, l.value)
		}
	case locatorTargetID:
		if snapshot.TargetID != l.value {
			return fmt.Errorf("focused targetID = %q, want %q", snapshot.TargetID, l.value)
		}
	case locatorTag:
		if snapshot.Tag != l.value {
			return fmt.Errorf("focused tag = %q, want %q", snapshot.Tag, l.value)
		}
	case locatorAt:
		if !snapshot.Bounds.Contains(l.x, l.y) {
			return fmt.Errorf("focused bounds %v do not contain point (%d,%d)", snapshot.Bounds, l.x, l.y)
		}
	default:
		return fmt.Errorf("locator %q is not supported for focus assertions", l.kind)
	}
	return nil
}
