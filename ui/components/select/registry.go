package selectcomp

import (
	"sort"
	"sync"
)

type popupRegistry struct {
	mu     sync.RWMutex
	nextID uint64
	popups map[*popupInstance]uint64
}

var selectPopupRegistry = &popupRegistry{
	popups: make(map[*popupInstance]uint64),
}

func (r *popupRegistry) register(inst *popupInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.popups[inst]; exists {
		return
	}
	r.nextID++
	r.popups[inst] = r.nextID
}

func (r *popupRegistry) unregister(inst *popupInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.popups, inst)
}

func (r *popupRegistry) openPopups() []*popupInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type entry struct {
		inst *popupInstance
		seq  uint64
	}
	entries := make([]entry, 0, len(r.popups))
	for inst, seq := range r.popups {
		if inst == nil {
			continue
		}
		entries = append(entries, entry{inst: inst, seq: seq})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].seq > entries[j].seq
	})

	result := make([]*popupInstance, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.inst)
	}
	return result
}
