// Package framework provides Inspector integration with hook system
package framework

import (
	"os"
	"reflect"

	"github.com/wwsheng009/mint/internal/log"
)

// HookRegistrar is an interface that Inspector can implement to register its own hooks
// This avoids import cycles between framework and internal/inspector
type HookRegistrar interface {
	RegisterWithHookManager(interface{})
}

// registerInspectorHook registers Inspector with the rendering hook system
// This is called automatically by SetInspector()
func (a *App) registerInspectorHook(inspector interface{}) {
	if inspector == nil {
		return
	}

	// Get the root's renderer (which should be a PipelineRendererAdapter)
	if a.root == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] Cannot register Inspector hook: root not set")
		}
		return
	}

	// Try to get HookManager from the root node
	hookManager := a.getHookManager()
	if hookManager == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] Cannot register Inspector hook: no HookManager found")
		}
		return
	}

	// Register hook using interface to avoid import cycle
	if registrar, ok := inspector.(HookRegistrar); ok {
		registrar.RegisterWithHookManager(hookManager)
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] Inspector hook registered via HookRegistrar interface")
		}
	} else {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] Inspector does not implement HookRegistrar interface")
			log.UILogger.Debug("[APP] Inspector will not be automatically injected into render tree")
		}
	}
}

// getHookManager attempts to get the HookManager from the rendering pipeline
// Uses reflection to avoid import cycles between framework and internal/render
func (a *App) getHookManager() interface{} {
	if a.root == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] getHookManager: root is nil")
		}
		return nil
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("[APP] getHookManager: root type=%T", a.root)
	}

	// Use reflection to call GetHooks() method dynamically
	// This avoids import cycles - we don't need to import internal/render
	rootValue := reflect.ValueOf(a.root)
	getHooksMethod := rootValue.MethodByName("GetHooks")

	if !getHooksMethod.IsValid() {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] getHookManager: root does not have GetHooks() method")
		}
		return nil
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("[APP] getHookManager: found GetHooks() method via reflection")
	}

	// Call GetHooks() method (no parameters, returns interface{})
	results := getHooksMethod.Call(nil)
	if len(results) == 0 {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			log.UILogger.Debug("[APP] getHookManager: GetHooks() returned no value")
		}
		return nil
	}

	hooks := results[0].Interface()
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("[APP] getHookManager: got hooks type=%T", hooks)
	}

	return hooks
}
