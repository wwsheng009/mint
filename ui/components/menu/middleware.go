package menu

import (
	"sort"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type menuRegistry struct {
	mu     sync.RWMutex
	nextID uint64
	menus  map[*popupInstance]uint64
	bars   map[*barInstance]struct{}
}

var menuRegistryGlobal = &menuRegistry{
	menus: make(map[*popupInstance]uint64),
	bars:  make(map[*barInstance]struct{}),
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
	r.bars = make(map[*barInstance]struct{})
	r.nextID = 0
}

func (r *menuRegistry) registerBar(inst *barInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bars[inst] = struct{}{}
}

func (r *menuRegistry) unregisterBar(inst *barInstance) {
	if inst == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.bars, inst)
}

func (r *menuRegistry) barsForMenuIDs(menuIDs map[string]struct{}) []*barInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(menuIDs) == 0 {
		return nil
	}
	bars := make([]*barInstance, 0, len(r.bars))
	for inst := range r.bars {
		if inst == nil {
			continue
		}
		if _, ok := menuIDs[inst.menuID()]; ok {
			bars = append(bars, inst)
		}
	}
	return bars
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
	if clickHitsOpenMenu(mouseMsg, menus) {
		return act
	}
	for _, inst := range menus {
		if inst != nil && inst.open && inst.model.CloseOnOutside {
			inst.close()
			return nil
		}
	}
	return act
}

func clickHitsOpenMenu(mouseMsg *runtimemsg.MouseMsg, menus []*popupInstance) bool {
	if mouseMsg == nil || len(menus) == 0 {
		return false
	}

	menuIDs := make(map[string]struct{}, len(menus))
	for _, inst := range menus {
		if inst == nil {
			continue
		}
		menuIDs[inst.menuID()] = struct{}{}
		if inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}

	if fiber := menuTargetFiber(mouseMsg); fiber != nil && fiberBelongsToMenu(fiber, menuIDs) {
		return true
	}

	for _, bar := range menuRegistryGlobal.barsForMenuIDs(menuIDs) {
		if bar.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}

	return false
}

func menuTargetFiber(mouseMsg *runtimemsg.MouseMsg) *rtui.Fiber {
	if mouseMsg == nil || mouseMsg.TargetFiber == nil {
		return nil
	}
	fiber, ok := mouseMsg.TargetFiber.(*rtui.Fiber)
	if !ok || fiber == nil {
		return nil
	}
	return fiber
}

func fiberBelongsToMenu(fiber *rtui.Fiber, menuIDs map[string]struct{}) bool {
	for node := fiber; node != nil; node = node.Return {
		switch inst := node.Instance.(type) {
		case *barInstance:
			if _, ok := menuIDs[inst.menuID()]; ok {
				return true
			}
		case *popupInstance:
			if _, ok := menuIDs[inst.menuID()]; ok {
				return true
			}
		}
	}
	return false
}
