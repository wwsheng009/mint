package selectcomp

import "sync"

type overlayRegistry struct {
	mu       sync.RWMutex
	triggers map[string]*Instance
	popups   map[string]*popupInstance
}

var selectOverlayRegistry = &overlayRegistry{
	triggers: make(map[string]*Instance),
	popups:   make(map[string]*popupInstance),
}

func (r *overlayRegistry) registerTrigger(ownerID string, inst *Instance) {
	if ownerID == "" || inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.triggers[ownerID] = inst
}

func (r *overlayRegistry) unregisterTrigger(ownerID string, inst *Instance) {
	if ownerID == "" || inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if current := r.triggers[ownerID]; current == inst {
		delete(r.triggers, ownerID)
	}
}

func (r *overlayRegistry) registerPopup(ownerID string, popup *popupInstance) {
	if ownerID == "" || popup == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.popups[ownerID] = popup
}

func (r *overlayRegistry) unregisterPopup(ownerID string, popup *popupInstance) {
	if ownerID == "" || popup == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if current := r.popups[ownerID]; current == popup {
		delete(r.popups, ownerID)
	}
}

func (r *overlayRegistry) trigger(ownerID string) *Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.triggers[ownerID]
}

func (r *overlayRegistry) popup(ownerID string) *popupInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.popups[ownerID]
}

func (r *overlayRegistry) markPopupDirty(ownerID string) {
	popup := r.popup(ownerID)
	if popup != nil {
		popup.MarkDirty()
	}
}

func (r *overlayRegistry) openTriggers() []*Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Instance, 0, len(r.triggers))
	for _, inst := range r.triggers {
		if inst != nil && inst.overlayPopup && inst.open {
			result = append(result, inst)
		}
	}
	return result
}
