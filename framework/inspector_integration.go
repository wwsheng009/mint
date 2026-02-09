// Package framework provides Inspector integration with hook system
package framework

import (
	"fmt"
	"os"
	"reflect"
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
			fmt.Fprintf(os.Stderr, "[APP] Cannot register Inspector hook: root not set\n")
		}
		return
	}

	// Try to get HookManager from the root node
	hookManager := a.getHookManager()
	if hookManager == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] Cannot register Inspector hook: no HookManager found\n")
		}
		return
	}

	// Register hook using interface to avoid import cycle
	if registrar, ok := inspector.(HookRegistrar); ok {
		registrar.RegisterWithHookManager(hookManager)
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] ✅ Inspector hook registered via HookRegistrar interface\n")
		}
	} else {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] ⚠️  Inspector does not implement HookRegistrar interface\n")
			fmt.Fprintf(os.Stderr, "[APP] Inspector will not be automatically injected into render tree\n")
		}
	}
}

// getHookManager attempts to get the HookManager from the rendering pipeline
// Uses reflection to avoid import cycles between framework and internal/render
func (a *App) getHookManager() interface{} {
	if a.root == nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] getHookManager: root is nil\n")
		}
		return nil
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] getHookManager: root type=%T\n", a.root)
	}

	// Use reflection to call GetHooks() method dynamically
	// This avoids import cycles - we don't need to import internal/render
	rootValue := reflect.ValueOf(a.root)
	getHooksMethod := rootValue.MethodByName("GetHooks")

	if !getHooksMethod.IsValid() {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] getHookManager: root does not have GetHooks() method\n")
		}
		return nil
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] getHookManager: found GetHooks() method via reflection\n")
	}

	// Call GetHooks() method (no parameters, returns interface{})
	results := getHooksMethod.Call(nil)
	if len(results) == 0 {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[APP] getHookManager: GetHooks() returned no value\n")
		}
		return nil
	}

	hooks := results[0].Interface()
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[APP] getHookManager: got hooks type=%T\n", hooks)
	}

	return hooks
}
