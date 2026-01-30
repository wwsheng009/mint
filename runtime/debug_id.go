// Package runtime provides debug ID registration for DevTools.
//
// This file implements a debug ID system that pre-allocates numeric IDs
// for components and fields to avoid string operations in the hot path.
package runtime

import (
	"sync"
)

// debugIDRegistry manages the allocation of debug IDs.
// These IDs are used by DevTools to track components and fields without
// using strings in the hot path.
var debugIDRegistry struct {
	mu                sync.RWMutex
	componentIDNext   uint32
	fieldIDNext       uint16
	componentNames    map[uint32]string
	fieldNames        map[uint16]string
	componentNameToID map[string]uint32
	fieldNameToID     map[string]uint16
}

func init() {
	debugIDRegistry.componentNames = make(map[uint32]string)
	debugIDRegistry.fieldNames = make(map[uint16]string)
	debugIDRegistry.componentNameToID = make(map[string]uint32)
	debugIDRegistry.fieldNameToID = make(map[string]uint16)
}

// RegisterComponent registers a component type and returns its debug ID.
// This should be called during component initialization.
func RegisterComponent(name string) uint32 {
	debugIDRegistry.mu.Lock()
	defer debugIDRegistry.mu.Unlock()

	if id, ok := debugIDRegistry.componentNameToID[name]; ok {
		return id
	}

	id := debugIDRegistry.componentIDNext
	debugIDRegistry.componentIDNext++
	debugIDRegistry.componentNames[id] = name
	debugIDRegistry.componentNameToID[name] = id
	return id
}

// RegisterField registers a field name and returns its debug ID.
// This should be called during component initialization.
func RegisterField(name string) uint16 {
	debugIDRegistry.mu.Lock()
	defer debugIDRegistry.mu.Unlock()

	if id, ok := debugIDRegistry.fieldNameToID[name]; ok {
		return id
	}

	id := debugIDRegistry.fieldIDNext
	debugIDRegistry.fieldIDNext++
	debugIDRegistry.fieldNames[id] = name
	debugIDRegistry.fieldNameToID[name] = id
	return id
}

// GetComponentName returns the component name for a given debug ID.
func GetComponentName(id uint32) string {
	debugIDRegistry.mu.RLock()
	defer debugIDRegistry.mu.RUnlock()
	return debugIDRegistry.componentNames[id]
}

// GetFieldName returns the field name for a given debug ID.
func GetFieldName(id uint16) string {
	debugIDRegistry.mu.RLock()
	defer debugIDRegistry.mu.RUnlock()
	return debugIDRegistry.fieldNames[id]
}

// GetComponentID returns the debug ID for a component name.
func GetComponentID(name string) uint32 {
	debugIDRegistry.mu.RLock()
	defer debugIDRegistry.mu.RUnlock()
	return debugIDRegistry.componentNameToID[name]
}

// GetFieldID returns the debug ID for a field name.
func GetFieldID(name string) uint16 {
	debugIDRegistry.mu.RLock()
	defer debugIDRegistry.mu.RUnlock()
	return debugIDRegistry.fieldNameToID[name]
}

// ComponentID is a type alias for component debug IDs.
type ComponentID uint32

// FieldID is a type alias for field debug IDs.
type FieldID uint16

// Common field IDs for frequently used fields.
const (
	FieldIDState      FieldID = 0
	FieldIDProps      FieldID = 1
	FieldIDStyle      FieldID = 2
	FieldIDVisible    FieldID = 3
	FieldIDDisabled   FieldID = 4
	FieldIDText       FieldID = 5
	FieldIDValue      FieldID = 6
	FieldIDFocusable  FieldID = 7
	FieldIDFocused    FieldID = 8
	FieldIDSelected   FieldID = 9
	FieldIDExpanded   FieldID = 10
	FieldIDChecked    FieldID = 11
	FieldIDProgress   FieldID = 12
)

// InitCommonFieldIDs initializes common field IDs.
func InitCommonFieldIDs() {
	RegisterField("State")
	RegisterField("Props")
	RegisterField("Style")
	RegisterField("Visible")
	RegisterField("Disabled")
	RegisterField("Text")
	RegisterField("Value")
	RegisterField("Focusable")
	RegisterField("Focused")
	RegisterField("Selected")
	RegisterField("Expanded")
	RegisterField("Checked")
	RegisterField("Progress")
}

func init() {
	InitCommonFieldIDs()
}
