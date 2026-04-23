package form

import "sync"

// Compatibility Form Registry
//
// Runtime form items and GetFormContext resolve through the instance tree.
// The registry remains only as an explicit compatibility path for ownerless /
// cross-tree lookups via the deprecated exported helpers below.
type formRegistryState struct {
	mu    sync.RWMutex
	forms map[string]*Instance
}

var compatRegistry = formRegistryState{
	forms: make(map[string]*Instance),
}

func registerCompatibleForm(formID string, form *Instance) {
	if formID == "" || form == nil {
		return
	}
	compatRegistry.mu.Lock()
	defer compatRegistry.mu.Unlock()
	compatRegistry.forms[formID] = form
}

func unregisterCompatibleForm(formID string) {
	if formID == "" {
		return
	}
	compatRegistry.mu.Lock()
	defer compatRegistry.mu.Unlock()
	delete(compatRegistry.forms, formID)
}

func lookupCompatibleForm(formID string) *Instance {
	if formID == "" {
		return nil
	}
	compatRegistry.mu.RLock()
	defer compatRegistry.mu.RUnlock()
	return compatRegistry.forms[formID]
}

func resetCompatibleRegistry() {
	compatRegistry.mu.Lock()
	defer compatRegistry.mu.Unlock()
	compatRegistry.forms = make(map[string]*Instance)
}

// RegisterForm registers a form instance with the compatibility registry.
// Deprecated: Form instances register themselves on mount; avoid calling this
// directly outside compatibility shims and tests.
func RegisterForm(formID string, form *Instance) {
	registerCompatibleForm(formID, form)
}

// UnregisterForm removes a form instance from the compatibility registry.
// Deprecated: Form instances unregister themselves on unmount; avoid calling
// this directly outside compatibility shims and tests.
func UnregisterForm(formID string) {
	unregisterCompatibleForm(formID)
}

// GetRegisteredForm returns the registered form instance for the given formID.
// Deprecated: prefer GetFormContext for owner-bound instance-tree resolution or
// GetRegisteredFormContext only for explicit cross-tree compatibility.
func GetRegisteredForm(formID string) *Instance {
	return lookupCompatibleForm(formID)
}

// GetForm returns the registered form instance for the given formID.
// Deprecated: use GetFormContext for owner-bound instance-tree resolution,
// GetRegisteredFormContext for registry-backed context access, or
// GetRegisteredForm for explicit registry instance lookup.
func GetForm(formID string) *Instance {
	return GetRegisteredForm(formID)
}

// GetRegisteredFormContext returns a FormContext backed by the compatibility
// registry only.
// Deprecated: prefer GetFormContext for owner-bound instance-tree resolution.
func GetRegisteredFormContext(formID string) FormContext {
	return newFormContext(lookupCompatibleForm(formID))
}

// ResetRegistry clears all registered form instances.
// Intended for use in tests to ensure isolation between test cases.
func ResetRegistry() {
	resetCompatibleRegistry()
}
