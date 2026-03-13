package notification

import (
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
)

// =============================================================================
// Manager - Multiple Notification Support
// =============================================================================

// Manager manages a stack of active notifications.
type Manager struct {
	notifications []*Instance
}

// NewManager creates a new notification Manager.
func NewManager() *Manager {
	return &Manager{
		notifications: make([]*Instance, 0),
	}
}

// Add adds a notification instance to be displayed.
func (m *Manager) Add(inst *Instance) {
	if inst != nil {
		inst.Show()
		m.notifications = append(m.notifications, inst)
	}
}

// AddFromVNode adds a notification from a VNode.
func (m *Manager) AddFromVNode(v *VNode) {
	if v != nil {
		inst := v.CreateInstance().(*Instance)
		m.Add(inst)
	}
}

// Info shows a persistent info notification.
func (m *Manager) Info(title, message string) {
	inst := NewBuilder(message).Title(title).Info().BuildInstance()
	m.Add(inst)
}

// Success shows a persistent success notification.
func (m *Manager) Success(title, message string) {
	inst := NewBuilder(message).Title(title).Success().BuildInstance()
	m.Add(inst)
}

// Warning shows a persistent warning notification.
func (m *Manager) Warning(title, message string) {
	inst := NewBuilder(message).Title(title).Warning().BuildInstance()
	m.Add(inst)
}

// Error shows a persistent error notification.
func (m *Manager) Error(title, message string) {
	inst := NewBuilder(message).Title(title).Error().BuildInstance()
	m.Add(inst)
}

// InfoTimed shows an info notification that auto-dismisses after d.
func (m *Manager) InfoTimed(title, message string, d time.Duration) {
	inst := NewBuilder(message).Title(title).Info().Duration(d).BuildInstance()
	m.Add(inst)
}

// SuccessTimed shows a success notification that auto-dismisses after d.
func (m *Manager) SuccessTimed(title, message string, d time.Duration) {
	inst := NewBuilder(message).Title(title).Success().Duration(d).BuildInstance()
	m.Add(inst)
}

// WarningTimed shows a warning notification that auto-dismisses after d.
func (m *Manager) WarningTimed(title, message string, d time.Duration) {
	inst := NewBuilder(message).Title(title).Warning().Duration(d).BuildInstance()
	m.Add(inst)
}

// ErrorTimed shows an error notification that auto-dismisses after d.
func (m *Manager) ErrorTimed(title, message string, d time.Duration) {
	inst := NewBuilder(message).Title(title).Error().Duration(d).BuildInstance()
	m.Add(inst)
}

// Dismiss hides and removes the notification at the given index.
func (m *Manager) Dismiss(index int) {
	if index < 0 || index >= len(m.notifications) {
		return
	}
	m.notifications[index].Hide()
	m.notifications = append(m.notifications[:index], m.notifications[index+1:]...)
}

// DismissAll hides and removes all notifications.
func (m *Manager) DismissAll() {
	for _, inst := range m.notifications {
		inst.Hide()
	}
	m.notifications = m.notifications[:0]
}

// Tick removes any expired (timed-out) notifications. Call this on each
// render frame to enable auto-dismiss behaviour.
func (m *Manager) Tick() {
	active := m.notifications[:0]
	for _, inst := range m.notifications {
		if inst.IsExpired() {
			inst.Hide()
		} else {
			active = append(active, inst)
		}
	}
	m.notifications = active
}

// Active returns the current list of visible notifications.
func (m *Manager) Active() []*Instance {
	return m.notifications
}

// Count returns the number of active notifications.
func (m *Manager) Count() int {
	return len(m.notifications)
}

// IsEmpty returns true when there are no active notifications.
func (m *Manager) IsEmpty() bool {
	return len(m.notifications) == 0
}

// Render returns the draw commands for all visible notifications, stacked
// vertically starting at (x, y) with the given width.
func (m *Manager) Render(x, y, width int) []paint.DrawCmd {
	var cmds []paint.DrawCmd
	curY := y
	for _, inst := range m.notifications {
		if !inst.IsVisible() {
			continue
		}
		size := inst.Measure(layout.Constraints{MaxWidth: width})
		inst.SetBounds(x, curY, size.Width, size.Height)
		cmds = append(cmds, inst.Paint(x, curY)...)
		curY += size.Height + 1 // 1-row gap between notifications
	}
	return cmds
}
