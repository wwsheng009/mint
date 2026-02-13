// Package inspector provides Inspector integration with render hook system
package inspector

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// CreateInspectorHook creates a VNode hook that automatically injects Inspector overlay
// This hook:
// 1. Checks if Inspector is visible
// 2. Wraps base VNode with Inspector overlay in Fragment
// 3. Sets LayerInspector on the overlay
// 4. Positions Inspector at configured coordinates
//
// The hook is called by PipelineRenderer during rendering, so application code
// doesn't need to manually handle Fragment wrapping or SetLayer() calls.
func CreateInspectorHook(inspector *StandaloneInspector) render.VNodeHook {
	return func(vnode rtui.VNode) rtui.VNode {
		// If Inspector is not visible, return original VNode unchanged
		if !inspector.IsVisible() {
			log.InspectorLogger.Debug("[InspectorHook] Inspector not visible, skipping injection")
			return vnode
		}

		log.InspectorLogger.Debug("[InspectorHook] Injecting Inspector overlay")

		// Get Inspector content (UI only, no Layer set)
		inspectorContent := inspector.RenderContent()
		if inspectorContent == nil {
			log.InspectorLogger.Debug("[InspectorHook] RenderContent() returned nil")
			return vnode
		}

		// Get Inspector position and size
		inspector.mu.RLock()
		x := inspector.floatX
		y := inspector.floatY
		width := inspector.overlayWidth
		height := inspector.overlayHeight
		inspector.mu.RUnlock()

		// Set positioning props
		inspectorContent.SetProps(rtui.Props{
			"x":      x,
			"y":      y,
			"width":  width,
			"height": height,
		})

		// Set the layer - this is the ONLY place where SetLayer is called
		// Application code and Inspector itself don't need to know about Layer
		inspectorContent.SetLayer(rtui.LayerInspector)

		log.InspectorLogger.Debug("[InspectorHook] Inspector overlay: layer=%d, pos=(%d,%d), size=%dx%d",
			rtui.LayerInspector, x, y, width, height)

		// Wrap base VNode and Inspector in Fragment
		// PipelineRenderer will detect the LayerInspector and use multi-layer rendering
		return rtui.Fragment(vnode, inspectorContent)
	}
}

// RegisterInspector registers Inspector with a HookManager
// This is typically called by framework.App.SetInspector()
//
// Example:
//
//	app.SetInspector(inspector) // Internally calls RegisterInspector
func RegisterInspector(inspector *StandaloneInspector, hookManager interface{}) {
	hook := CreateInspectorHook(inspector)

	// Type assertion to RegisterVNodeHook method
	type hookManagerWrapper interface {
		RegisterVNodeHook(render.VNodeHook)
	}

	if hm, ok := hookManager.(hookManagerWrapper); ok {
		hm.RegisterVNodeHook(hook)

		log.InspectorLogger.Debug("[Inspector] Registered with render hook system")
	}
}

// RegisterWithHookManager implements framework.HookRegistrar interface
// This allows framework to register Inspector hook without direct import
func (si *StandaloneInspector) RegisterWithHookManager(hookManager interface{}) {
	RegisterInspector(si, hookManager)
}
