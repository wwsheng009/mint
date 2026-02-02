// Package ui provides component instance management for persistent state.
package ui

import (
	state "github.com/wwsheng009/mint/internal/state"
)

// InstanceManager manages component instances across renders.
// This is a type alias to internal/state.InstanceManager.
// Use state.InstanceManager directly for new code.
type InstanceManager = state.InstanceManager

// NewInstanceManager creates a new instance manager.
// Deprecated: Use state.NewInstanceManager() instead.
func NewInstanceManager() *InstanceManager {
	return state.NewInstanceManager()
}
