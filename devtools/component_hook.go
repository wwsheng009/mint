// Package devtools provides component state change hooks for DevTools.
//
// This file implements hooks for tracking component state changes
// and integrating them into the causal graph.
package devtools

import (
	"sync"
	"sync/atomic"
)

// ComponentHookManager manages hooks for component state changes.
type ComponentHookManager struct {
	enabled atomic.Uint32

	// Registered component hooks
	hooks   map[uint32]*ComponentHooks
	hooksMu sync.RWMutex

	// CausalBuilder for recording mutations
	builder *CausalBuilder
}

// ComponentHooks contains hooks for a specific component.
type ComponentHooks struct {
	ComponentID   uint32
	ComponentName string

	// BeforeStateChange is called before a state change.
	// Return false to cancel the change.
	BeforeStateChange func(field string, oldValue, newValue interface{}) bool

	// AfterStateChange is called after a state change.
	AfterStateChange func(field string, oldValue, newValue interface{})

	// BeforePropChange is called before a prop change.
	BeforePropChange func(prop string, oldValue, newValue interface{}) bool

	// AfterPropChange is called after a prop change.
	AfterPropChange func(prop string, oldValue, newValue interface{})

	// OnLayoutInvalidated is called when layout is invalidated.
	OnLayoutInvalidated func()
}

// HookContext provides context for hook execution.
type HookContext struct {
	ComponentID   uint32
	ComponentName string
	FieldID       uint16
	FieldName     string
	Kind          MutationKind
	OldValue      interface{}
	NewValue      interface{}
}

// NewComponentHookManager creates a new component hook manager.
func NewComponentHookManager(builder *CausalBuilder) *ComponentHookManager {
	chm := &ComponentHookManager{
		hooks:   make(map[uint32]*ComponentHooks),
		builder: builder,
	}
	chm.enabled.Store(0)
	return chm
}

// Enable enables the component hook manager.
func (chm *ComponentHookManager) Enable() {
	chm.enabled.Store(1)
}

// Disable disables the component hook manager.
func (chm *ComponentHookManager) Disable() {
	chm.enabled.Store(0)
}

// IsEnabled returns whether the hook manager is enabled.
func (chm *ComponentHookManager) IsEnabled() bool {
	return chm.enabled.Load() == 1
}

// RegisterHooks registers hooks for a component.
func (chm *ComponentHookManager) RegisterHooks(componentID uint32, componentName string, hooks *ComponentHooks) {
	chm.hooksMu.Lock()
	defer chm.hooksMu.Unlock()

	hooks.ComponentID = componentID
	hooks.ComponentName = componentName
	chm.hooks[componentID] = hooks
}

// UnregisterHooks unregisters hooks for a component.
func (chm *ComponentHookManager) UnregisterHooks(componentID uint32) {
	chm.hooksMu.Lock()
	defer chm.hooksMu.Unlock()

	delete(chm.hooks, componentID)
}

// GetHooks returns the hooks for a component.
func (chm *ComponentHookManager) GetHooks(componentID uint32) *ComponentHooks {
	chm.hooksMu.RLock()
	defer chm.hooksMu.RUnlock()

	return chm.hooks[componentID]
}

// BeforeStateChange calls the before state change hook for a component.
func (chm *ComponentHookManager) BeforeStateChange(componentID uint32, field string, oldValue, newValue interface{}) bool {
	if !chm.IsEnabled() {
		return true
	}

	hooks := chm.GetHooks(componentID)
	if hooks == nil || hooks.BeforeStateChange == nil {
		return true
	}

	return hooks.BeforeStateChange(field, oldValue, newValue)
}

// AfterStateChange calls the after state change hook for a component.
func (chm *ComponentHookManager) AfterStateChange(componentID uint32, field string, oldValue, newValue interface{}) {
	if !chm.IsEnabled() {
		return
	}

	hooks := chm.GetHooks(componentID)
	if hooks != nil && hooks.AfterStateChange != nil {
		hooks.AfterStateChange(field, oldValue, newValue)
	}

	// Record in causal builder
	if chm.builder != nil {
		componentName := ""
		if hooks != nil {
			componentName = hooks.ComponentName
		}
		chm.builder.RecordMutation(componentName, MutationState, field, oldValue, newValue)
	}
}

// BeforePropChange calls the before prop change hook for a component.
func (chm *ComponentHookManager) BeforePropChange(componentID uint32, prop string, oldValue, newValue interface{}) bool {
	if !chm.IsEnabled() {
		return true
	}

	hooks := chm.GetHooks(componentID)
	if hooks == nil || hooks.BeforePropChange == nil {
		return true
	}

	return hooks.BeforePropChange(prop, oldValue, newValue)
}

// AfterPropChange calls the after prop change hook for a component.
func (chm *ComponentHookManager) AfterPropChange(componentID uint32, prop string, oldValue, newValue interface{}) {
	if !chm.IsEnabled() {
		return
	}

	hooks := chm.GetHooks(componentID)
	if hooks != nil && hooks.AfterPropChange != nil {
		hooks.AfterPropChange(prop, oldValue, newValue)
	}

	// Record in causal builder
	if chm.builder != nil {
		componentName := ""
		if hooks != nil {
			componentName = hooks.ComponentName
		}
		chm.builder.RecordMutation(componentName, MutationProp, prop, oldValue, newValue)
	}
}

// OnLayoutInvalidated calls the layout invalidated hook for a component.
func (chm *ComponentHookManager) OnLayoutInvalidated(componentID uint32) {
	if !chm.IsEnabled() {
		return
	}

	hooks := chm.GetHooks(componentID)
	if hooks != nil && hooks.OnLayoutInvalidated != nil {
		hooks.OnLayoutInvalidated()
	}
}

// StateChangeHook is a simple function hook for state changes.
type StateChangeHook func(componentID uint32, field string, oldValue, newValue interface{})

// PropChangeHook is a simple function hook for prop changes.
type PropChangeHook func(componentID uint32, prop string, oldValue, newValue interface{})

// SimpleHookManager provides a simpler hook interface.
type SimpleHookManager struct {
	enabled    atomic.Uint32
	stateHooks []StateChangeHook
	propHooks  []PropChangeHook
	mu         sync.RWMutex
	builder    *CausalBuilder
}

// NewSimpleHookManager creates a new simple hook manager.
func NewSimpleHookManager(builder *CausalBuilder) *SimpleHookManager {
	shm := &SimpleHookManager{
		stateHooks: make([]StateChangeHook, 0, 8),
		propHooks:  make([]PropChangeHook, 0, 8),
		builder:    builder,
	}
	shm.enabled.Store(0)
	return shm
}

// Enable enables the simple hook manager.
func (shm *SimpleHookManager) Enable() {
	shm.enabled.Store(1)
}

// Disable disables the simple hook manager.
func (shm *SimpleHookManager) Disable() {
	shm.enabled.Store(0)
}

// IsEnabled returns whether the hook manager is enabled.
func (shm *SimpleHookManager) IsEnabled() bool {
	return shm.enabled.Load() == 1
}

// OnStateChange registers a hook for state changes.
func (shm *SimpleHookManager) OnStateChange(hook StateChangeHook) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	shm.stateHooks = append(shm.stateHooks, hook)
}

// OnPropChange registers a hook for prop changes.
func (shm *SimpleHookManager) OnPropChange(hook PropChangeHook) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	shm.propHooks = append(shm.propHooks, hook)
}

// RecordStateChange records and notifies state change hooks.
func (shm *SimpleHookManager) RecordStateChange(componentName string, field string, oldValue, newValue interface{}) {
	if !shm.IsEnabled() {
		return
	}

	shm.mu.RLock()
	hooks := make([]StateChangeHook, len(shm.stateHooks))
	copy(hooks, shm.stateHooks)
	shm.mu.RUnlock()

	// Call all hooks (without componentID in simple mode)
	for _, hook := range hooks {
		hook(0, field, oldValue, newValue)
	}

	// Record in causal builder
	if shm.builder != nil {
		shm.builder.RecordMutation(componentName, MutationState, field, oldValue, newValue)
	}
}

// RecordPropChange records and notifies prop change hooks.
func (shm *SimpleHookManager) RecordPropChange(componentName string, prop string, oldValue, newValue interface{}) {
	if !shm.IsEnabled() {
		return
	}

	shm.mu.RLock()
	hooks := make([]PropChangeHook, len(shm.propHooks))
	copy(hooks, shm.propHooks)
	shm.mu.RUnlock()

	// Call all hooks (without componentID in simple mode)
	for _, hook := range hooks {
		hook(0, prop, oldValue, newValue)
	}

	// Record in causal builder
	if shm.builder != nil {
		shm.builder.RecordMutation(componentName, MutationProp, prop, oldValue, newValue)
	}
}

// RecordStyleChange records a style change in the causal builder.
func (shm *SimpleHookManager) RecordStyleChange(componentName string, styleProp string, oldValue, newValue interface{}) {
	if !shm.IsEnabled() {
		return
	}

	if shm.builder != nil {
		shm.builder.RecordMutation(componentName, MutationStyle, styleProp, oldValue, newValue)
	}
}

// RecordFocusChange records a focus change in the causal builder.
func (shm *SimpleHookManager) RecordFocusChange(componentName string, oldFocus, newFocus interface{}) {
	if !shm.IsEnabled() {
		return
	}

	if shm.builder != nil {
		shm.builder.RecordMutation(componentName, MutationFocus, "focus", oldFocus, newFocus)
	}
}

// Global hook manager instance for convenience
var (
	globalHookManager     *ComponentHookManager
	globalHookManagerOnce sync.Once
	globalSimpleHookMgr   *SimpleHookManager
)

// GetGlobalHookManager returns the global component hook manager.
func GetGlobalHookManager(builder *CausalBuilder) *ComponentHookManager {
	globalHookManagerOnce.Do(func() {
		globalHookManager = NewComponentHookManager(builder)
	})
	return globalHookManager
}

// GetGlobalSimpleHookManager returns the global simple hook manager.
func GetGlobalSimpleHookManager(builder *CausalBuilder) *SimpleHookManager {
	if globalSimpleHookMgr == nil {
		globalSimpleHookMgr = NewSimpleHookManager(builder)
	}
	return globalSimpleHookMgr
}
