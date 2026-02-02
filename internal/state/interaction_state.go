package state

import (
	"fmt"
	"os"
	"sync"
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
	// Check debug environment
	return &KeyValidator{
		enableWarnings: os.Getenv("TUI_DEBUG_KEYS") == "true" || os.Getenv("TUI_DEBUG_UI") == "true",
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

	fmt.Fprintf(os.Stderr, "[Mint Warning] Missing keys in %s with %d children\n", parentName, len(children))
	fmt.Fprintf(os.Stderr, "  Positions without keys: %v\n", missing)
	fmt.Fprintf(os.Stderr, "  This may cause issues with:\n")
	fmt.Fprintf(os.Stderr, "    - Component state preservation\n")
	fmt.Fprintf(os.Stderr, "    - Hover/focus state in dynamic lists\n")
	fmt.Fprintf(os.Stderr, "    - Performance (unnecessary re-renders)\n")
	fmt.Fprintf(os.Stderr, "  Fix: Add unique keys to each child:\n")
	fmt.Fprintf(os.Stderr, "    ui.ComponentBuilder(\"Item\").Key(fmt.Sprintf(\"item-%%d\", id)).Build()\n")
}

// warnAboutDuplicateKeys prints a warning about duplicate keys
func (v *KeyValidator) warnAboutDuplicateKeys(parent VNode, duplicates map[string]int) {
	parentName := "Fragment"
	if parent != nil {
		parentName = parent.Type().String()
	}

	fmt.Fprintf(os.Stderr, "[Mint Warning] Duplicate keys in %s\n", parentName)
	for key, count := range duplicates {
		fmt.Fprintf(os.Stderr, "  Key %q appears %d times\n", key, count)
	}
	fmt.Fprintf(os.Stderr, "  Keys must be unique among siblings\n")
}
