package inspector

import (
	"fmt"
	"os"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// IntegrationHelper provides helper methods for integrating the inspector
// with the Mint TUI framework event system
type IntegrationHelper struct {
	inspector *Inspector
	rootVNode rtui.VNode
}

// NewIntegrationHelper creates a new integration helper
func NewIntegrationHelper(inspector *Inspector) *IntegrationHelper {
	return &IntegrationHelper{
		inspector: inspector,
	}
}

// SetRootVNode sets the root VNode for the inspector to use
// This should be called whenever the VNode tree changes
func (ih *IntegrationHelper) SetRootVNode(root rtui.VNode) {
	ih.rootVNode = root
}

// CreateEventFilter returns an event filter function that can be used with framework.App
// The filter intercepts events before they reach the normal event routing
func (ih *IntegrationHelper) CreateEventFilter() func(frameworkevent.Event) bool {
	return func(ev frameworkevent.Event) bool {
		// Check if this is a keyboard event
		if kev, ok := ev.(*frameworkevent.KeyEvent); ok {
			// Convert framework.KeyEvent to inspector.KeyEvent
			inspectorEvent := KeyEvent{
				Key:   kev.Key.Name,
				Ctrl:  kev.Key.Ctrl,
				Alt:   kev.Key.Alt,
				Shift: kev.Modifiers&frameworkevent.ModShift != 0,
			}

			// Handle inspector shortcuts
			if ih.inspector.HandleKeyEvent(inspectorEvent) {
				// Event was handled by inspector, don't propagate
				if os.Getenv("TUI_INSPECTOR_DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "[Inspector] Handled key event: %s (ctrl=%v)\n",
						inspectorEvent.Key, inspectorEvent.Ctrl)
				}
				return false // Block event from normal routing
			}
		}

		// Let all other events pass through
		return true
	}
}

// CreateMouseHandler returns a mouse event handler function
// This can be used to hook into the framework's mouse event system
func (ih *IntegrationHelper) CreateMouseHandler() func(x, y int) bool {
	return func(x, y int) bool {
		if !ih.inspector.IsEnabled() {
			return false
		}

		// Handle mouse event
		handled := ih.inspector.HandleMouseEvent(x, y)

		// Update hovered node if we have a root
		if ih.rootVNode != nil {
			hovered := ih.inspector.FindVNodeAt(ih.rootVNode, x, y)
			ih.inspector.hoveredVNode = hovered

			if os.Getenv("TUI_INSPECTOR_DEBUG") == "true" && hovered != nil {
				info := ExtractElementInfo(hovered)
				fmt.Fprintf(os.Stderr, "[Inspector] Hovered: %s (%s)\n", info.Type, info.Label)
			}
		}

		return handled
	}
}

// RegisterWithApp provides a simple way to register the inspector with a framework App
// This is a convenience method that sets up all necessary hooks
//
// Usage:
//   inspector := inspector.NewInspector()
//   helper := inspector.NewIntegrationHelper(inspector)
//   helper.RegisterWithApp(app)
//
// Note: This is a simplified integration. For more control, use the individual
// CreateEventFilter() and CreateMouseHandler() methods.
func (ih *IntegrationHelper) RegisterWithApp(app interface{}) bool {
	// This is a placeholder for full framework integration
	// The actual implementation will depend on how framework.App exposes its event system
	// For now, this demonstrates the intended API

	// TODO: Implement actual registration once framework.App integration points are clear
	// Potential approaches:
	// 1. app.SetEventFilter(helper.CreateEventFilter())
	// 2. app.RegisterGlobalHandler(handler)
	// 3. app.AddInspector(inspector)

	if os.Getenv("TUI_INSPECTOR_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] RegisterWithApp called (integration pending)\n")
	}

	return false
}

// EnableFromEnvironment checks environment variables and enables the inspector if requested
// Returns true if the inspector was enabled
func (ih *IntegrationHelper) EnableFromEnvironment() bool {
	// Check for TUI_INSPECTOR environment variable
	if os.Getenv("TUI_INSPECTOR") == "true" {
		ih.inspector.Enable()
		fmt.Fprintf(os.Stderr, "[Inspector] Enabled via TUI_INSPECTOR=true\n")
		fmt.Fprintf(os.Stderr, "[Inspector] Press F12 or Ctrl+I to toggle, Tab to navigate, Esc to close\n")
		return true
	}

	// Check for TUI_INSPECTOR_AUTO environment variable (enable with auto-start)
	if os.Getenv("TUI_INSPECTOR_AUTO") == "true" {
		ih.inspector.Enable()
		fmt.Fprintf(os.Stderr, "[Inspector] Auto-enabled via TUI_INSPECTOR_AUTO=true\n")
		fmt.Fprintf(os.Stderr, "[Inspector] Press F12 or Ctrl+I to toggle, Tab to navigate, Esc to close\n")
		return true
	}

	return false
}

// GetInspector returns the underlying Inspector instance
func (ih *IntegrationHelper) GetInspector() *Inspector {
	return ih.inspector
}
