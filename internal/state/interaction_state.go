package state

import (
	"sync"

	"github.com/wwsheng009/mint/internal/log"
)

// InteractionType represents the type of interaction state
type InteractionType int

const (
	InteractionHovered InteractionType = iota
	InteractionFocused
	InteractionPressed
	InteractionSelected
)

// InteractionStateManager manages persistent interaction states across renders.
// This decouples interaction state from VNode instances, allowing state
// to survive re-renders when VNodes are recreated.
type InteractionStateManager struct {
	mu    sync.RWMutex
	states map[string]map[InteractionType]bool // key -> type -> state
}

// NewInteractionStateManager creates a new interaction state manager
func NewInteractionStateManager() *InteractionStateManager {
	return &InteractionStateManager{
		states: make(map[string]map[InteractionType]bool),
	}
}

// Get retrieves an interaction state for a key
func (m *InteractionStateManager) Get(key string, itype InteractionType) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if key == "" {
		return false
	}

	if states, ok := m.states[key]; ok {
		return states[itype]
	}
	return false
}

// Set sets an interaction state for a key
func (m *InteractionStateManager) Set(key string, itype InteractionType, value bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key == "" {
		return
	}

	if m.states[key] == nil {
		m.states[key] = make(map[InteractionType]bool)
	}
	m.states[key][itype] = value
}

// Clear removes all interaction states for a key
func (m *InteractionStateManager) Clear(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key == "" {
		return
	}

	delete(m.states, key)
}

// ClearType removes a specific interaction type for a key
func (m *InteractionStateManager) ClearType(key string, itype InteractionType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key == "" {
		return
	}

	if states, ok := m.states[key]; ok {
		delete(states, itype)
		if len(states) == 0 {
			delete(m.states, key)
		}
	}
}

// Cleanup removes states for keys not in the active set
func (m *InteractionStateManager) Cleanup(activeKeys []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	activeSet := make(map[string]bool)
	for _, key := range activeKeys {
		activeSet[key] = true
	}

	for key := range m.states {
		if !activeSet[key] {
			delete(m.states, key)
		}
	}
}

// IsHovered checks if a component is hovered
func (m *InteractionStateManager) IsHovered(key string) bool {
	return m.Get(key, InteractionHovered)
}

// SetHovered sets the hovered state
func (m *InteractionStateManager) SetHovered(key string, hovered bool) {
	m.Set(key, InteractionHovered, hovered)
}

// IsFocused checks if a component is focused
func (m *InteractionStateManager) IsFocused(key string) bool {
	return m.Get(key, InteractionFocused)
}

// SetFocused sets the focused state
func (m *InteractionStateManager) SetFocused(key string, focused bool) {
	m.Set(key, InteractionFocused, focused)
}

// IsPressed checks if a component is pressed
func (m *InteractionStateManager) IsPressed(key string) bool {
	return m.Get(key, InteractionPressed)
}

// SetPressed sets the pressed state
func (m *InteractionStateManager) SetPressed(key string, pressed bool) {
	m.Set(key, InteractionPressed, pressed)
}

// IsSelected checks if a component is selected
func (m *InteractionStateManager) IsSelected(key string) bool {
	return m.Get(key, InteractionSelected)
}

// SetSelected sets the selected state
func (m *InteractionStateManager) SetSelected(key string, selected bool) {
	m.Set(key, InteractionSelected, selected)
}

// =============================================================================
// Key Validation
// =============================================================================

// KeyValidator validates keys in component trees
type KeyValidator struct {
	enableWarnings bool
}

// NewKeyValidator creates a new key validator
func NewKeyValidator() *KeyValidator {
	// Check debug environment via Logger enabled state
	return &KeyValidator{
		enableWarnings: log.KeyLogger.Enabled() || log.UILogger.Enabled(),
	}
}

// ValidateChildren checks if children in a list have proper keys
// Returns true if valid, false if keys are missing in a potential list scenario
func (v *KeyValidator) ValidateChildren(parent VNode, children []VNode) bool {
	if !v.enableWarnings {
		return true
	}

	// Only validate for lists (multiple children)
	if len(children) <= 1 {
		return true
	}

	// Check if all children have keys
	missingKeys := make([]int, 0)
	for i, child := range children {
		if child.Key() == "" {
			missingKeys = append(missingKeys, i)
		}
	}

	if len(missingKeys) > 0 {
		v.warnAboutMissingKeys(parent, children, missingKeys)
		return false
	}

	// Check for duplicate keys
	keyCount := make(map[string]int)
	duplicates := make(map[string]int)
	for _, child := range children {
		key := child.Key()
		if key != "" {
			keyCount[key]++
			if keyCount[key] > 1 {
				duplicates[key] = keyCount[key]
			}
		}
	}

	if len(duplicates) > 0 {
		v.warnAboutDuplicateKeys(parent, duplicates)
		return false
	}

	return true
}

// warnAboutMissingKeys prints a warning about missing keys
func (v *KeyValidator) warnAboutMissingKeys(parent VNode, children []VNode, missing []int) {
	parentName := "Fragment"
	if parent != nil {
		parentName = parent.Type().String()
	}

	log.EngineLogger.IfEnabled().Debug("[Mint Warning] Missing keys in %s with %d children", parentName, len(children))
	log.EngineLogger.IfEnabled().Debug("  Positions without keys: %v", missing)
	log.EngineLogger.IfEnabled().Debug("  This may cause issues with:")
	log.EngineLogger.IfEnabled().Debug("    - Component state preservation")
	log.EngineLogger.IfEnabled().Debug("    - Hover/focus state in dynamic lists")
	log.EngineLogger.IfEnabled().Debug("    - Performance (unnecessary re-renders)")
	log.EngineLogger.IfEnabled().Debug("  Fix: Add unique keys to each child:")
	log.EngineLogger.IfEnabled().Debug("    ui.ComponentBuilder(\"Item\").Key(fmt.Sprintf(\"item-%%d\", id)).Build()")
}

// warnAboutDuplicateKeys prints a warning about duplicate keys
func (v *KeyValidator) warnAboutDuplicateKeys(parent VNode, duplicates map[string]int) {
	parentName := "Fragment"
	if parent != nil {
		parentName = parent.Type().String()
	}

	log.EngineLogger.IfEnabled().Debug("[Mint Warning] Duplicate keys in %s", parentName)
	for key, count := range duplicates {
		log.EngineLogger.IfEnabled().Debug("  Key %q appears %d times", key, count)
	}
	log.EngineLogger.IfEnabled().Debug("  Keys must be unique among siblings")
}
