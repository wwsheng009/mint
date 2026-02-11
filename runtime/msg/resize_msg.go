package msg

import "time"

// ResizeMsg represents a window resize event.
type ResizeMsg struct {
	BaseMsg

	// OldWidth and OldHeight are the previous dimensions
	OldWidth int
	OldHeight int

	// NewWidth and NewHeight are the current dimensions
	NewWidth int
	NewHeight int
}

// NewResizeMsg creates a new resize message.
func NewResizeMsg(oldW, oldH, newW, newH int) *ResizeMsg {
	return &ResizeMsg{
		BaseMsg: BaseMsg{
			TypeValue:      MsgTypeResize,
			TimestampValue: time.Now(),
		},
		OldWidth:  oldW,
		OldHeight: oldH,
		NewWidth:  newW,
		NewHeight: newH,
	}
}

// String returns a string representation of the resize message.
func (r *ResizeMsg) String() string {
	return r.BaseMsg.String()
}
