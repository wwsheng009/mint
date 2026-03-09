package menu

import (
	"sort"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

type menuRegistry struct {
	mu     sync.RWMutex
	nextID uint64
	menus  map[*popupInstance]uint64
}

var menuRegistryGlobal = &menuRegistry{
	menus: make(map[*popupInstance]uint64),
}

func (r *menuRegistry) register(inst *popupInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	r.menus[inst] = r.nextID
}

func (r *menuRegistry) unregister(inst *popupInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.menus, inst)
}

func (r *menuRegistry) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.menus = make(map[*popupInstance]uint64)
	r.nextID = 0
}

func (r *menuRegistry) openMenus() []*popupInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	type entry struct {
		inst *popupInstance
		seq  uint64
	}
	entries := make([]entry, 0, len(r.menus))
	for inst, seq := range r.menus {
		if inst != nil && inst.open {
			entries = append(entries, entry{inst: inst, seq: seq})
		}
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

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Name() string {
	return "MenuMiddleware"
}

func (m *Middleware) Before(act *action.Action) *action.Action {
	if act == nil {
		return nil
	}
	switch act.Type {
	case action.ActionCancel, action.ActionQuit:
		return m.handleEscape(act)
	case action.ActionClick:
		return m.handleClickOutside(act)
	}
	return act
}

func (m *Middleware) After(act *action.Action, result *action.RouterResult) {}

func (m *Middleware) handleEscape(act *action.Action) *action.Action {
	for _, inst := range menuRegistryGlobal.openMenus() {
		if inst != nil && inst.open && inst.model.CloseOnEscape {
			inst.close()
			return nil
		}
	}
	return act
}

func (m *Middleware) handleClickOutside(act *action.Action) *action.Action {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return act
	}
	if mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}
	menus := menuRegistryGlobal.openMenus()
	if len(menus) == 0 {
		return act
	}
	for _, inst := range menus {
		if inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return act
		}
	}
	for _, inst := range menus {
		if inst != nil && inst.open && inst.model.CloseOnOutside {
			inst.close()
			return nil
		}
	}
	return act
}
