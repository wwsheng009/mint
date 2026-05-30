package ui

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/log"
)

// FiberFocusManager manages keyboard focus within a Fiber tree.
// This is the Fiber-first implementation that only accesses Fiber nodes.
//
// Design Principles:
// - VNode declares "what I want" (intent)
// - Fiber holds "what I am now" (runtime state)
// - Focus is runtime state, so it belongs to Fiber
// - Layer-aware: when a Modal is open, focus is trapped within the Modal layer
type FiberFocusManager struct {
	focusable      []*Fiber              // All focusable Fiber nodes in the tree
	focusableLayer []Layer               // Effective layer for each focusable Fiber (inherits modal/overlay scope)
	current        int                   // Index of currently focused Fiber, -1 if none
	onNavigate     func(from, to *Fiber) // Callback when focus changes
	activeLayer    Layer                 // Current active layer (highest layer with content)
}

type fiberFocusIdentity struct {
	nodeID uint64
	path   string
	tag    string
}

// NewFiberFocusManager creates a new Fiber-based focus manager.
func NewFiberFocusManager() *FiberFocusManager {
	return &FiberFocusManager{
		focusable:      []*Fiber{},
		focusableLayer: []Layer{},
		current:        -1,
		activeLayer:    LayerBase,
	}
}

// SetActiveLayer sets the active layer for focus trapping.
// When a Modal is open, only Modal layer fibers are focusable.
// This also automatically focuses the first item in the new active layer.
func (m *FiberFocusManager) SetActiveLayer(layer Layer) {
	oldLayer := m.activeLayer
	m.activeLayer = layer
	log.FocusLogger.IfEnabled().Debug("SetActiveLayer: %d -> %d", oldLayer, layer)

	// If switching to a higher layer, focus first item in that layer
	if layer > oldLayer && layer > LayerBase {
		m.FocusFirst()
	}
}

// GetActiveLayer returns the current active layer.
func (m *FiberFocusManager) GetActiveLayer() Layer {
	return m.activeLayer
}

// HasActiveLayer returns true if there's a modal or overlay layer active.
func (m *FiberFocusManager) HasActiveLayer() bool {
	return m.activeLayer > LayerBase
}

// CollectFromFiber collects all focusable Fiber nodes from a Fiber tree.
// This should be called during Commit phase after the Fiber tree is complete.
func (m *FiberFocusManager) CollectFromFiber(root *Fiber) {
	previousIdentity, previousIndex, previousFiber := m.snapshotCurrentFocus()
	m.focusable, m.focusableLayer = m.collectFocusableFibers(root, LayerBase)
	m.restoreCurrentFocus(previousIdentity, previousIndex, previousFiber)
	log.FocusLogger.IfEnabled().Debug("CollectFromFiber: collected %d focusable fibers", len(m.focusable))
	for i, f := range m.focusable {
		focusID := fmt.Sprintf("node-%d", f.NodeID)
		log.FocusLogger.IfEnabled().Debug("  [%d] FocusID=%s, Tag=%s, EffectiveLayer=%s", i, focusID, f.Tag, m.getFocusableLayer(i))
	}
}

// collectFocusableFibers recursively collects focusable Fiber nodes.
func (m *FiberFocusManager) collectFocusableFibers(fiber *Fiber, inheritedLayer Layer) ([]*Fiber, []Layer) {
	var result []*Fiber
	var layers []Layer

	if fiber == nil {
		return result, layers
	}

	// Skip the root ComponentVNode wrapper
	if fiber.Key == "root" && fiber.Type == VNodeComponent {
		return m.collectFocusableFibers(fiber.Child, inheritedLayer)
	}

	effectiveLayer := inheritedLayer
	if fiber.Layer != LayerBase && fiber.Layer.IsValid() {
		effectiveLayer = fiber.Layer
	}

	// Check if current Fiber is focusable via FocusableInstance
	// Fiber-first: use Instance instead of FocusableMeta
	if fiber.Instance != nil {
		if focusable, ok := fiber.Instance.(FocusableInstance); ok {
			// Only add to focusable list if not disabled
			if !focusable.IsDisabled() {
				result = append(result, fiber)
				layers = append(layers, effectiveLayer)
			}
		}
	}

	// Recursively check children
	if child := fiber.Child; child != nil {
		childFibers, childLayers := m.collectFocusableFibers(child, effectiveLayer)
		result = append(result, childFibers...)
		layers = append(layers, childLayers...)
	}

	// Recursively check siblings
	if sibling := fiber.Sibling; sibling != nil {
		siblingFibers, siblingLayers := m.collectFocusableFibers(sibling, inheritedLayer)
		result = append(result, siblingFibers...)
		layers = append(layers, siblingLayers...)
	}

	return result, layers
}

// SetFocusable sets the list of focusable Fiber nodes.
// It attempts to preserve focus by Fiber identity across re-renders.
func (m *FiberFocusManager) SetFocusable(fibers []*Fiber) {
	previousIdentity, previousIndex, previousFiber := m.snapshotCurrentFocus()

	m.focusable = fibers
	m.focusableLayer = make([]Layer, len(fibers))
	for i, fiber := range fibers {
		if fiber != nil {
			m.focusableLayer[i] = fiber.Layer
		}
	}

	m.restoreCurrentFocus(previousIdentity, previousIndex, previousFiber)

	log.FocusLogger.IfEnabled().Debug("SetFocusable: before=%d, after=%d, fibers=%d", previousIndex, m.current, len(fibers))

	// If no focus and there are focusable fibers, focus the first one
	if m.current < 0 && len(m.focusable) > 0 {
		m.FocusFirst()
	}
}

func (m *FiberFocusManager) snapshotCurrentFocus() (fiberFocusIdentity, int, *Fiber) {
	index := m.current
	if index < 0 || index >= len(m.focusable) {
		return fiberFocusIdentity{}, index, nil
	}
	fiber := m.focusable[index]
	return focusIdentityForFiber(fiber), index, fiber
}

func (m *FiberFocusManager) restoreCurrentFocus(identity fiberFocusIdentity, previousIndex int, previousFiber *Fiber) {
	newIndex := m.findFocusIdentity(identity)

	if newIndex < 0 && previousIndex >= 0 && len(m.focusable) > 0 {
		if previousIndex >= len(m.focusable) {
			newIndex = len(m.focusable) - 1
		} else {
			newIndex = previousIndex
		}
	}

	m.current = newIndex
	newFiber := m.getFiber(newIndex)
	if previousFiber != nil && previousFiber != newFiber {
		setFiberFocused(previousFiber, false)
	}
	setFiberFocused(newFiber, true)
}

func (m *FiberFocusManager) findFocusIdentity(identity fiberFocusIdentity) int {
	if !identity.valid() {
		return -1
	}
	for i, fiber := range m.focusable {
		if identity.matches(fiber) {
			return i
		}
	}
	return -1
}

func focusIdentityForFiber(fiber *Fiber) fiberFocusIdentity {
	if fiber == nil {
		return fiberFocusIdentity{}
	}
	return fiberFocusIdentity{
		nodeID: fiber.NodeID,
		path:   fiber.Path,
		tag:    fiber.Tag,
	}
}

func (id fiberFocusIdentity) valid() bool {
	return id.nodeID != 0 || id.path != ""
}

func (id fiberFocusIdentity) matches(fiber *Fiber) bool {
	if fiber == nil {
		return false
	}
	if id.nodeID != 0 && fiber.NodeID == id.nodeID {
		return true
	}
	return id.path != "" && fiber.Path == id.path && fiber.Tag == id.tag
}

// FocusNext moves focus to the next focusable Fiber.
// If activeLayer is set (e.g., Modal), only cycles within that layer.
func (m *FiberFocusManager) FocusNext() bool {
	log.FocusLogger.IfEnabled().Debug("FocusNext current=%d, len(focusable)=%d, activeLayer=%d", m.current, len(m.focusable), m.activeLayer)
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current

	// If we have an active layer (e.g., Modal), only navigate within that layer
	if m.activeLayer > LayerBase {
		newIndex := m.findNextInLayer(m.current, m.activeLayer)
		if newIndex == -1 {
			return false // No focusable items in active layer
		}
		m.current = newIndex
	} else {
		// Normal navigation: cycle through all focusable items
		m.current = (m.current + 1) % len(m.focusable)
	}

	log.FocusLogger.IfEnabled().Debug("FocusNext old=%d, new=%d", old, m.current)

	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(m.getFiber(old), m.focusable[m.current])
	}

	return true
}

// FocusPrev moves focus to the previous focusable Fiber.
// If activeLayer is set (e.g., Modal), only cycles within that layer.
func (m *FiberFocusManager) FocusPrev() bool {
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current

	// If we have an active layer (e.g., Modal), only navigate within that layer
	if m.activeLayer > LayerBase {
		newIndex := m.findPrevInLayer(m.current, m.activeLayer)
		if newIndex == -1 {
			return false // No focusable items in active layer
		}
		m.current = newIndex
	} else {
		// Normal navigation: cycle through all focusable items
		m.current = m.current - 1
		if m.current < 0 {
			m.current = len(m.focusable) - 1
		}
	}

	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(m.getFiber(old), m.focusable[m.current])
	}

	return true
}

// findNextInLayer finds the next focusable fiber in the specified layer.
func (m *FiberFocusManager) findNextInLayer(current int, layer Layer) int {
	if len(m.focusable) == 0 {
		return -1
	}

	// Start from the next item and wrap around
	for i := 1; i <= len(m.focusable); i++ {
		idx := (current + i) % len(m.focusable)
		if m.getFocusableLayer(idx) == layer {
			return idx
		}
	}
	return -1
}

// findPrevInLayer finds the previous focusable fiber in the specified layer.
func (m *FiberFocusManager) findPrevInLayer(current int, layer Layer) int {
	if len(m.focusable) == 0 {
		return -1
	}

	// Start from the previous item and wrap around
	for i := 1; i <= len(m.focusable); i++ {
		idx := current - i
		if idx < 0 {
			idx = len(m.focusable) + idx
		}
		if m.getFocusableLayer(idx) == layer {
			return idx
		}
	}
	return -1
}

// FocusFirst focuses the first focusable Fiber.
// If activeLayer is set, focuses the first fiber in that layer.
func (m *FiberFocusManager) FocusFirst() bool {
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current

	// If we have an active layer, find first in that layer
	if m.activeLayer > LayerBase {
		for i, fiber := range m.focusable {
			if m.getFocusableLayer(i) == m.activeLayer {
				m.current = i
				m.updateFocusState(old, m.current)
				if m.onNavigate != nil {
					m.onNavigate(nil, fiber)
				}
				return true
			}
		}
		return false // No items in active layer
	}

	// Normal: focus first
	m.current = 0
	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(nil, m.focusable[0])
	}

	return true
}

// FocusLast focuses the last focusable Fiber.
// If activeLayer is set, focuses the last fiber in that layer.
func (m *FiberFocusManager) FocusLast() bool {
	if len(m.focusable) == 0 {
		return false
	}

	old := m.current

	// If we have an active layer, find last in that layer
	if m.activeLayer > LayerBase {
		for i := len(m.focusable) - 1; i >= 0; i-- {
			if m.getFocusableLayer(i) == m.activeLayer {
				m.current = i
				m.updateFocusState(old, m.current)
				if m.onNavigate != nil {
					m.onNavigate(nil, m.focusable[i])
				}
				return true
			}
		}
		return false // No items in active layer
	}

	// Normal: focus last
	m.current = len(m.focusable) - 1
	m.updateFocusState(old, m.current)

	if m.onNavigate != nil {
		m.onNavigate(nil, m.focusable[m.current])
	}

	return true
}

// GetCurrent returns the currently focused Fiber, or nil if none.
func (m *FiberFocusManager) GetCurrent() *Fiber {
	if m.current < 0 || m.current >= len(m.focusable) {
		return nil
	}
	return m.focusable[m.current]
}

// HandleEvent handles keyboard events for focus navigation.
func (m *FiberFocusManager) HandleEvent(ev event.Event) (handled bool, shouldRender bool) {
	keyEvent, ok := ev.(*event.KeyEvent)
	if !ok {
		return false, false
	}

	// Tab - navigate to next
	if keyEvent.Special == event.KeyTab {
		if keyEvent.Modifiers == event.ModShift {
			return m.FocusPrev(), true
		}
		return m.FocusNext(), true
	}

	return false, false
}

// SetOnNavigate sets a callback for focus navigation changes.
func (m *FiberFocusManager) SetOnNavigate(fn func(from, to *Fiber)) {
	m.onNavigate = fn
}

// Count returns the number of focusable Fibers.
func (m *FiberFocusManager) Count() int {
	return len(m.focusable)
}

// UpdateFocusableList directly updates the focusable list without changing focus state.
// This is used internally when the list needs to be refreshed but the focus index should be preserved.
func (m *FiberFocusManager) UpdateFocusableList(fibers []*Fiber) {
	m.focusable = fibers
	m.focusableLayer = make([]Layer, len(fibers))
	for i, fiber := range fibers {
		if fiber != nil {
			m.focusableLayer[i] = fiber.Layer
		}
	}
}

// GetFocusable returns the list of focusable Fibers.
func (m *FiberFocusManager) GetFocusable() []*Fiber {
	return m.focusable
}

// CurrentIndex returns the index of the currently focused Fiber.
func (m *FiberFocusManager) CurrentIndex() int {
	return m.current
}

// SetFocusByIndex sets focus to a Fiber at the given index.
func (m *FiberFocusManager) SetFocusByIndex(index int) bool {
	if index < 0 || index >= len(m.focusable) {
		return false
	}
	old := m.current
	m.current = index
	m.updateFocusState(old, m.current)
	if m.onNavigate != nil {
		m.onNavigate(m.getFiber(old), m.focusable[index])
	}
	return true
}

// SetFocusByID sets focus to a Fiber with the given FocusID.
// NodeID-based focus ID format: "node-{NodeID}"
func (m *FiberFocusManager) SetFocusByID(id string) bool {
	for i, fiber := range m.focusable {
		fiberID := fmt.Sprintf("%d", fiber.NodeID)
		// Check NodeID-based FocusID (Fiber-first)
		if fiberID == id {
			old := m.current
			m.current = i
			m.updateFocusState(old, m.current)
			if m.onNavigate != nil {
				m.onNavigate(m.getFiber(old), fiber)
			}
			return true
		}
	}
	return false
}

// HasFocus returns whether the given Fiber currently has focus.
func (m *FiberFocusManager) HasFocus(fiber *Fiber) bool {
	if m.current < 0 || m.current >= len(m.focusable) {
		return false
	}
	return m.focusable[m.current] == fiber
}

// GetFocusedFiber returns the currently focused Fiber, or nil if none.
func (m *FiberFocusManager) GetFocusedFiber() *Fiber {
	if m.current < 0 || m.current >= len(m.focusable) {
		return nil
	}
	return m.focusable[m.current]
}

// DebugString returns a debug string representation.
func (m *FiberFocusManager) DebugString() string {
	if len(m.focusable) == 0 {
		return "FiberFocusManager{no focusable fibers}"
	}

	currentID := "none"
	if m.current >= 0 && m.current < len(m.focusable) {
		// Fiber-first: use NodeID to generate FocusID
		currentID = fmt.Sprintf("node-%d", m.focusable[m.current].NodeID)
	}

	return fmt.Sprintf("FiberFocusManager{count=%d, current=%d, currentID=%s}",
		len(m.focusable), m.current, currentID)
}

// updateFocusState updates the focus state when focus changes.
func (m *FiberFocusManager) updateFocusState(oldIndex, newIndex int) {
	// If focus index didn't change, avoid toggling false->true.
	// Toggling would reset caret blink phase every render.
	if oldIndex == newIndex {
		m.applyFocusState(newIndex, true)
		return
	}

	// Remove focus from old fiber
	m.applyFocusState(oldIndex, false)
	// Add focus to new fiber
	m.applyFocusState(newIndex, true)
}

// applyFocusState applies focus state to a fiber at the given index.
func (m *FiberFocusManager) applyFocusState(index int, focused bool) {
	if index < 0 || index >= len(m.focusable) {
		return
	}

	setFiberFocused(m.focusable[index], focused)
}

func setFiberFocused(fiber *Fiber, focused bool) {
	if fiber == nil || fiber.Instance == nil {
		return
	}

	// IMPORTANT: Use Instance.(FocusableInstance).SetFocus() instead of FocusableVNode.SetFocus()
	// FocusableVNode is DEPRECATED and its hasFocus field is not used during rendering.
	// Instance.state.Focused is the actual runtime state used by resolveStyle().
	if focusable, ok := fiber.Instance.(interface{ SetFocus(bool) }); ok {
		focusable.SetFocus(focused)
	}
}

// getFiber safely gets a fiber at index, returns nil if out of bounds.
func (m *FiberFocusManager) getFiber(index int) *Fiber {
	if index < 0 || index >= len(m.focusable) {
		return nil
	}
	return m.focusable[index]
}

func (m *FiberFocusManager) getFocusableLayer(index int) Layer {
	if index < 0 || index >= len(m.focusable) {
		return LayerBase
	}
	if index < len(m.focusableLayer) {
		return m.focusableLayer[index]
	}
	if fiber := m.focusable[index]; fiber != nil {
		return fiber.Layer
	}
	return LayerBase
}
