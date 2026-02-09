// Package render provides hook system for VNode transformation
// Hooks allow plugins (like Inspector) to modify VNode trees without
// requiring application code to handle Layer management
package render

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// VNodeHook is a function that transforms a VNode tree
// It can be used to inject overlays (Inspector, Modal, Tooltip) automatically
type VNodeHook func(rtui.VNode) rtui.VNode

// HookManager manages all registered hooks
// Hooks are applied in the order they were registered
type HookManager struct {
	vnodeHooks []VNodeHook
}

// NewHookManager creates a new HookManager
func NewHookManager() *HookManager {
	return &HookManager{
		vnodeHooks: make([]VNodeHook, 0),
	}
}

// RegisterVNodeHook registers a VNode transformation hook
// Hooks are called in LIFO order (last registered runs first)
// This allows later hooks to wrap the output of earlier hooks
func (hm *HookManager) RegisterVNodeHook(hook VNodeHook) {
	if hook == nil {
		return
	}
	hm.vnodeHooks = append(hm.vnodeHooks, hook)
}

// ApplyVNodeHooks applies all registered hooks to a VNode tree
// Hooks are applied in reverse order (last registered runs first)
func (hm *HookManager) ApplyVNodeHooks(vnode rtui.VNode) rtui.VNode {
	if hm == nil || len(hm.vnodeHooks) == 0 {
		return vnode
	}

	// Apply hooks in reverse order (LIFO)
	// This allows Inspector (registered later) to wrap everything
	result := vnode
	for i := len(hm.vnodeHooks) - 1; i >= 0; i-- {
		hook := hm.vnodeHooks[i]
		result = hook(result)
	}

	return result
}

// ClearVNodeHooks removes all registered VNode hooks
func (hm *HookManager) ClearVNodeHooks() {
	if hm == nil {
		return
	}
	hm.vnodeHooks = make([]VNodeHook, 0)
}

// VNodeHookCount returns the number of registered VNode hooks
func (hm *HookManager) VNodeHookCount() int {
	if hm == nil {
		return 0
	}
	return len(hm.vnodeHooks)
}
