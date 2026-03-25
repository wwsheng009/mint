package transfer

import "github.com/wwsheng009/mint/runtime/intent"

// MoveDirection identifies a transfer operation direction.
type MoveDirection string

const (
	MoveDirectionToTarget MoveDirection = "toTarget"
	MoveDirectionToSource MoveDirection = "toSource"
)

// MoveIntent requests moving the current selection between the two lists.
type MoveIntent struct {
	ComponentID string
	Direction   MoveDirection
}

func (MoveIntent) IntentType() string              { return "transfer.MoveIntent" }
func (MoveIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (MoveIntent) IsTransition() bool              { return false }
func (MoveIntent) IsGlobal() bool                  { return false }
func (i MoveIntent) GetComponentID() string        { return i.ComponentID }

// ChangeIntent is emitted whenever target keys change.
type ChangeIntent struct {
	ComponentID string
	Direction   MoveDirection
	MovedKeys   []string
	SourceKeys  []string
	TargetKeys  []string
}

func (ChangeIntent) IntentType() string              { return "transfer.ChangeIntent" }
func (ChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ChangeIntent) IsTransition() bool              { return false }
func (ChangeIntent) IsGlobal() bool                  { return false }
func (i ChangeIntent) GetComponentID() string        { return i.ComponentID }

// MoveToTarget creates a move-to-target intent without routing metadata.
func MoveToTarget() MoveIntent {
	return MoveIntent{Direction: MoveDirectionToTarget}
}

// MoveToTargetWithID creates a move-to-target intent scoped to one transfer.
func MoveToTargetWithID(componentID string) MoveIntent {
	move := MoveToTarget()
	move.ComponentID = componentID
	return move
}

// MoveToSource creates a move-to-source intent without routing metadata.
func MoveToSource() MoveIntent {
	return MoveIntent{Direction: MoveDirectionToSource}
}

// MoveToSourceWithID creates a move-to-source intent scoped to one transfer.
func MoveToSourceWithID(componentID string) MoveIntent {
	move := MoveToSource()
	move.ComponentID = componentID
	return move
}

// Change creates a transfer change intent without routing metadata.
func Change(direction MoveDirection, movedKeys, sourceKeys, targetKeys []string) ChangeIntent {
	return ChangeIntent{
		Direction:  direction,
		MovedKeys:  append([]string(nil), movedKeys...),
		SourceKeys: append([]string(nil), sourceKeys...),
		TargetKeys: append([]string(nil), targetKeys...),
	}
}

// ChangeWithID creates a transfer change intent scoped to one transfer.
func ChangeWithID(componentID string, direction MoveDirection, movedKeys, sourceKeys, targetKeys []string) ChangeIntent {
	change := Change(direction, movedKeys, sourceKeys, targetKeys)
	change.ComponentID = componentID
	return change
}
