package popover

import (
	"sort"
	"sync"
)

type popoverRegistry struct {
	mu      sync.RWMutex
	nextSeq uint64
	entries map[*Instance]uint64
}

var popoverRegistryGlobal = &popoverRegistry{
	entries: make(map[*Instance]uint64),
}

func (r *popoverRegistry) register(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[inst]; !ok {
		r.entries[inst] = 0
	}
}

func (r *popoverRegistry) touch(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[inst]; !ok {
		return
	}
	r.nextSeq++
	r.entries[inst] = r.nextSeq
}

func (r *popoverRegistry) unregister(inst *Instance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, inst)
}

func (r *popoverRegistry) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = make(map[*Instance]uint64)
	r.nextSeq = 0
}

func (r *popoverRegistry) openInstances() []*Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type entry struct {
		inst *Instance
		seq  uint64
	}

	entries := make([]entry, 0, len(r.entries))
	for inst, seq := range r.entries {
		if inst != nil && inst.open {
			entries = append(entries, entry{inst: inst, seq: seq})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].seq > entries[j].seq
	})

	result := make([]*Instance, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.inst)
	}
	return result
}
