package tooltip

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Toast Manager - Multiple Toast Support
// =============================================================================

// Manager manages multiple toast notifications.
type Manager struct {
	toasts []*ToastInstance
}

// NewManager creates a new toast manager.
func NewManager() *Manager {
	return &Manager{
		toasts: make([]*ToastInstance, 0),
	}
}

// Add adds a toast to be displayed.
func (tm *Manager) Add(toast *ToastInstance) {
	if toast != nil {
		toast.Show()
		tm.toasts = append(tm.toasts, toast)
	}
}

// AddFromVNode adds a toast from a Toast VNode.
func (tm *Manager) AddFromVNode(toastVNode *ToastVNode) {
	if toastVNode != nil {
		inst := toastVNode.CreateInstance().(*ToastInstance)
		tm.Add(inst)
	}
}

// Info shows an info toast with the given message.
func (tm *Manager) Info(message string) {
	toast := NewToastBuilder(message).
		Info().
		BuildInstance()
	tm.Add(toast)
}

// Success shows a success toast with the given message.
func (tm *Manager) Success(message string) {
	toast := NewToastBuilder(message).
		Success().
		BuildInstance()
	tm.Add(toast)
}

// Warning shows a warning toast with the given message.
func (tm *Manager) Warning(message string) {
	toast := NewToastBuilder(message).
		Warning().
		BuildInstance()
	tm.Add(toast)
}

// Error shows an error toast with the given message.
func (tm *Manager) Error(message string) {
	toast := NewToastBuilder(message).
		Error().
		BuildInstance()
	tm.Add(toast)
}

// Remove removes a toast from the manager.
// It doesn't hide the toast - the caller should call Hide() first.
func (tm *Manager) Remove(toast *ToastInstance) {
	for i, t := range tm.toasts {
		if t == toast {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			break
		}
	}
}

// HideAndRemove hides and removes a toast.
func (tm *Manager) HideAndRemove(toast *ToastInstance) {
	toast.Hide()
	tm.Remove(toast)
}

// Clear removes all toasts from the manager.
func (tm *Manager) Clear() {
	// Hide all toasts before clearing
	for _, toast := range tm.toasts {
		toast.Hide()
	}
	tm.toasts = tm.toasts[:0]
}

// GetToasts returns all active toast instances.
func (tm *Manager) GetToasts() []*ToastInstance {
	return tm.toasts
}

// GetVNodes returns all active toasts as VNodes.
func (tm *Manager) GetVNodes() []rtui.VNode {
	if len(tm.toasts) == 0 {
		return nil
	}

	vnodes := make([]rtui.VNode, len(tm.toasts))
	for i, inst := range tm.toasts {
		// Create a new VNode from instance props
		vnodes[i] = NewToastBuilder(inst.Message()).
			Title(inst.Title()).
			Type(inst.ToastType()).
			Duration(inst.Duration()).
			Build()
	}
	return vnodes
}

// CleanExpired removes all expired toasts.
func (tm *Manager) CleanExpired() {
	activeToasts := make([]*ToastInstance, 0, len(tm.toasts))
	for _, toast := range tm.toasts {
		if !toast.IsExpired() && toast.visible {
			activeToasts = append(activeToasts, toast)
		} else if toast.visible {
			// Toast is expired but still visible, hide it
			toast.Hide()
		}
	}
	tm.toasts = activeToasts
}

// Count returns the number of active toasts.
func (tm *Manager) Count() int {
	return len(tm.toasts)
}

// IsEmpty returns true if there are no active toasts.
func (tm *Manager) IsEmpty() bool {
	return len(tm.toasts) == 0
}
