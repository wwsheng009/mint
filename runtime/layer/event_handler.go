// Package layer provides event handling support for layered UI components
package layer

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Layer Event Handler
// =============================================================================

// EventHandler handles layer-specific events like ESC to close modal,
// click-outside to close, and focus trap management
type EventHandler struct {
	manager *Manager
}

// NewEventHandler creates a new layer event handler
func NewEventHandler(manager *Manager) *EventHandler {
	return &EventHandler{
		manager: manager,
	}
}

// HandleKeyEvent processes a key event and returns true if handled
func (h *EventHandler) HandleKeyEvent(keyName string, keyRune rune) bool {
	if os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[LayerEventHandler] HandleKeyEvent: keyName=%q keyRune=%c\n", keyName, keyRune)
	}

	// Check if there's an active modal
	if !h.manager.HasModal() {
		return false
	}

	// Get modal nodes
	modalNodes := h.manager.GetModalNodes()
	if len(modalNodes) == 0 {
		return false
	}

	modalNode := modalNodes[0]

	// Handle ESC key to close modal
	if keyName == "esc" {
		if h.shouldCloseOnESC(modalNode) {
			if os.Getenv("TUI_LAYER_DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "[LayerEventHandler] ESC pressed, closing modal\n")
			}
			h.triggerOnClose(modalNode)
			return true
		}
	}

	return false
}

// HandleMouseEvent processes a mouse event and returns true if handled
func (h *EventHandler) HandleMouseEvent(x, y int) bool {
	if os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[LayerEventHandler] HandleMouseEvent: x=%d y=%d\n", x, y)
	}

	// Check if there's an active modal
	if !h.manager.HasModal() {
		return false
	}

	// Get modal layout
	modalLayout, ok := h.manager.GetLayout(rtui.LayerModal)
	if !ok || modalLayout == nil || modalLayout.Root == nil {
		return false
	}

	// Check if click is outside modal bounds
	if h.isClickOutsideModal(x, y, modalLayout.Root.Box) {
		modalNodes := h.manager.GetModalNodes()
		if len(modalNodes) > 0 {
			modalNode := modalNodes[0]
			if h.shouldCloseOnBackdrop(modalNode) {
				if os.Getenv("TUI_LAYER_DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "[LayerEventHandler] Click outside modal, closing\n")
				}
				h.triggerOnClose(modalNode)
				return true
			}
		}
	}

	return false
}

// ShouldBlockEvent returns true if events should be blocked (modal is open)
func (h *EventHandler) ShouldBlockEvent(x, y int) bool {
	return h.manager.ShouldBlockEvent(x, y)
}

// GetFocusableIDs returns focusable component IDs within the modal
func (h *EventHandler) GetFocusableIDs() []string {
	modalLayout, ok := h.manager.GetLayout(rtui.LayerModal)
	if !ok || modalLayout == nil {
		return []string{}
	}

	return h.collectFocusableIDs(modalLayout.Root)
}

// =============================================================================
// Helper Methods
// =============================================================================

// shouldCloseOnESC checks if the modal should close on ESC
func (h *EventHandler) shouldCloseOnESC(node *LayerNode) bool {
	if node == nil || node.Content == nil {
		return true // Default: close on ESC
	}

	props := node.Content.Props()
	if props == nil {
		return true
	}

	if closeOnESC, ok := props["_closeOnESC"].(bool); ok {
		return closeOnESC
	}

	return true // Default: close on ESC
}

// shouldCloseOnBackdrop checks if the modal should close on backdrop click
func (h *EventHandler) shouldCloseOnBackdrop(node *LayerNode) bool {
	if node == nil || node.Content == nil {
		return true // Default: close on backdrop
	}

	props := node.Content.Props()
	if props == nil {
		return true
	}

	if closeOnBackdrop, ok := props["_closeOnBackdrop"].(bool); ok {
		return closeOnBackdrop
	}

	return true // Default: close on backdrop
}

// triggerOnClose calls the modal's OnClose callback
func (h *EventHandler) triggerOnClose(node *LayerNode) {
	if node == nil || node.Content == nil {
		return
	}

	props := node.Content.Props()
	if props == nil {
		return
	}

	if onClose, ok := props["_onClose"].(func()); ok {
		// Call the OnClose callback in a goroutine to avoid blocking
		go onClose()
	}
}

// isClickOutsideModal checks if a click position is outside the modal bounds
func (h *EventHandler) isClickOutsideModal(x, y int, modalBox runtime.Box) bool {
	return x < modalBox.X || x >= modalBox.X+modalBox.Width ||
		y < modalBox.Y || y >= modalBox.Y+modalBox.Height
}

// collectFocusableIDs recursively collects focusable component IDs
func (h *EventHandler) collectFocusableIDs(node *compute.ComputedBox) []string {
	// This is a placeholder - the actual implementation would need
	// to interface with the focus system
	// For now, return empty to avoid blocking focus
	return []string{}
}
