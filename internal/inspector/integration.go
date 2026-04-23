package inspector

import (
	"os"
	"strings"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/log"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// IntegrationHelper provides helper methods for integrating the inspector
// with the Mint TUI framework event system
type IntegrationHelper struct {
	inspector *Inspector
	rootVNode rtui.VNode
}

type inspectorAppBridge struct {
	helper *IntegrationHelper
}

func (b *inspectorAppBridge) ToggleVisibility() {
	if b == nil || b.helper == nil || b.helper.inspector == nil {
		return
	}

	if b.helper.inspector.IsEnabled() {
		b.helper.inspector.Disable()
		return
	}
	b.helper.inspector.Enable()
}

func (b *inspectorAppBridge) IsVisible() bool {
	if b == nil || b.helper == nil || b.helper.inspector == nil {
		return false
	}
	return b.helper.inspector.IsEnabled()
}

func (b *inspectorAppBridge) HandleKeyEvent(key string, alt, ctrl, shift bool) bool {
	if b == nil || b.helper == nil || b.helper.inspector == nil {
		return false
	}

	return b.helper.inspector.HandleKeyEvent(KeyEvent{
		Key:   key,
		Alt:   alt,
		Ctrl:  ctrl,
		Shift: shift,
	})
}

func (b *inspectorAppBridge) HandleMouseEvent(_ frameworkevent.EventType, ev *frameworkevent.MouseEvent) bool {
	if b == nil || b.helper == nil || ev == nil {
		return false
	}

	return b.helper.CreateMouseHandler()(ev.X, ev.Y)
}

func (b *inspectorAppBridge) AttachToApp(root rtui.VNode) {
	if b == nil || b.helper == nil {
		return
	}
	b.helper.SetRootVNode(root)
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
				log.InspectorLogger.Debug("Handled key event: %s (ctrl=%v)",
					inspectorEvent.Key, inspectorEvent.Ctrl)
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

			if hovered != nil {
				info := ExtractElementInfo(hovered)
				log.InspectorLogger.IfEnabled().Debug("Hovered: %s (%s)", info.Type, info.Label)
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
	if ih == nil || ih.inspector == nil || app == nil {
		return false
	}

	if appWithInspector, ok := app.(interface{ SetInspector(interface{}) }); ok {
		appWithInspector.SetInspector(&inspectorAppBridge{helper: ih})
		if shortcutHost, ok := app.(interface{ SetupInspectorShortcut() }); ok {
			shortcutHost.SetupInspectorShortcut()
		}
		log.InspectorLogger.IfEnabled().Debug("RegisterWithApp: registered inspector bridge")
		return true
	}

	if appWithFilter, ok := app.(interface {
		SetEventFilter(func(frameworkevent.Event) bool)
	}); ok {
		appWithFilter.SetEventFilter(ih.CreateEventFilter())
		log.InspectorLogger.IfEnabled().Debug("RegisterWithApp: registered event filter fallback")
		return true
	}

	log.InspectorLogger.IfEnabled().Debug("RegisterWithApp: unsupported app type %T", app)
	return false
}

// EnableFromEnvironment checks environment variables and enables the inspector if requested
// Returns true if the inspector was enabled
func (ih *IntegrationHelper) EnableFromEnvironment() bool {
	if ih == nil || ih.inspector == nil {
		return false
	}

	if envEnabled("TUI_INSPECTOR") {
		ih.inspector.Enable()
		log.InspectorLogger.IfEnabled().Debug("Enabled via TUI_INSPECTOR=true")
		log.InspectorLogger.IfEnabled().Debug("Press F12 or Ctrl+I to toggle, Tab to navigate, Esc to close")
		return true
	}

	if envEnabled("TUI_INSPECTOR_AUTO") {
		ih.inspector.Enable()
		log.InspectorLogger.IfEnabled().Debug("Auto-enabled via TUI_INSPECTOR_AUTO=true")
		log.InspectorLogger.IfEnabled().Debug("Press F12 or Ctrl+I to toggle, Tab to navigate, Esc to close")
		return true
	}

	return false
}

// GetInspector returns the underlying Inspector instance
func (ih *IntegrationHelper) GetInspector() *Inspector {
	return ih.inspector
}

func envEnabled(name string) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
