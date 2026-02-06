// Package ui provides declarative UI components for terminal applications
package ui

// =============================================================================
// Layer API
// =============================================================================

// WithLayer sets the rendering layer for a VNode
// This allows components to be rendered in different visual layers
// (base, overlay, modal, tooltip) for proper z-ordering
func WithLayer(content VNode, layer Layer) VNode {
	if content == nil {
		return nil
	}
	return content.SetLayer(layer)
}

// =============================================================================
// Modal Component
// =============================================================================

// ModalBuilder builds a modal dialog component
type ModalBuilder struct {
	content         VNode
	onClose         func()
	closeOnESC      bool
	closeOnBackdrop bool
	centered        bool
}

// Modal creates a new modal builder
// Modals are centered dialogs that block interaction with background content
func Modal(content VNode) *ModalBuilder {
	return &ModalBuilder{
		content:         content,
		closeOnESC:      true,
		closeOnBackdrop: true,
		centered:        true,
	}
}

// OnClose sets the callback when modal is closed
func (b *ModalBuilder) OnClose(fn func()) *ModalBuilder {
	b.onClose = fn
	return b
}

// CloseOnESC enables/disables closing on ESC key
func (b *ModalBuilder) CloseOnESC(close bool) *ModalBuilder {
	b.closeOnESC = close
	return b
}

// CloseOnBackdropClick enables/disables closing when clicking outside
func (b *ModalBuilder) CloseOnBackdropClick(close bool) *ModalBuilder {
	b.closeOnBackdrop = close
	return b
}

// Centered sets whether the modal should be centered
func (b *ModalBuilder) Centered(centered bool) *ModalBuilder {
	b.centered = centered
	return b
}

// Build creates the modal VNode
func (b *ModalBuilder) Build() VNode {
	// Set modal props
	props := Props{
		"_layer":           LayerModal,
		"_closeOnESC":      b.closeOnESC,
		"_closeOnBackdrop": b.closeOnBackdrop,
		"_onClose":         b.onClose,
	}

	if b.content.Props() != nil {
		// Merge with existing props
		for k, v := range b.content.Props() {
			if k != "_layer" { // Don't override layer
				props[k] = v
			}
		}
	}

	b.content.SetProps(props)
	return b.content.SetLayer(LayerModal)
}

// =============================================================================
// Overlay Component
// =============================================================================

// OverlayBuilder builds an overlay component
// Overlays are for dropdowns, popovers, and similar content that appears above base content
type OverlayBuilder struct {
	content      VNode
	anchor       VNode // The element that anchors this overlay
	position     OverlayPosition
	offsetX      int
	offsetY      int
}

// OverlayPosition defines where the overlay appears relative to its anchor
type OverlayPosition int

const (
	// OverlayPositionBottom places overlay below the anchor
	OverlayPositionBottom OverlayPosition = iota
	// OverlayPositionTop places overlay above the anchor
	OverlayPositionTop
	// OverlayPositionLeft places overlay to the left of the anchor
	OverlayPositionLeft
	// OverlayPositionRight places overlay to the right of the anchor
	OverlayPositionRight
	// OverlayPositionCenter centers the overlay
	OverlayPositionCenter
)

// Overlay creates a new overlay builder
func Overlay(content VNode) *OverlayBuilder {
	return &OverlayBuilder{
		content:  content,
		position: OverlayPositionBottom,
		offsetX:  0,
		offsetY:  1, // Default: 1 cell below anchor
	}
}

// Anchor sets the anchor element for this overlay
func (b *OverlayBuilder) Anchor(anchor VNode) *OverlayBuilder {
	b.anchor = anchor
	return b
}

// Position sets the position relative to anchor
func (b *OverlayBuilder) Position(pos OverlayPosition) *OverlayBuilder {
	b.position = pos
	return b
}

// Offset sets the offset from the anchor
func (b *OverlayBuilder) Offset(x, y int) *OverlayBuilder {
	b.offsetX = x
	b.offsetY = y
	return b
}

// Build creates the overlay VNode
func (b *OverlayBuilder) Build() VNode {
	props := Props{
		"_layer":    LayerOverlay,
		"_position": b.position,
		"_offsetX":  b.offsetX,
		"_offsetY":  b.offsetY,
	}

	if b.content.Props() != nil {
		// Merge with existing props
		for k, v := range b.content.Props() {
			if k != "_layer" {
				props[k] = v
			}
		}
	}

	b.content.SetProps(props)
	return b.content.SetLayer(LayerOverlay)
}

// =============================================================================
// Tooltip Component
// =============================================================================

// TooltipBuilder builds a tooltip component
// Tooltips are small hints that appear on hover
type TooltipBuilder struct {
	text     string
	anchor   VNode
	delayMs  int // Delay before showing tooltip (not implemented yet)
}

// Tooltip creates a new tooltip builder
func Tooltip(text string) *TooltipBuilder {
	return &TooltipBuilder{
		text:    text,
		delayMs: 500, // Default 500ms delay
	}
}

// Anchor sets the anchor element for this tooltip
func (b *TooltipBuilder) Anchor(anchor VNode) *TooltipBuilder {
	b.anchor = anchor
	return b
}

// Delay sets the delay before showing the tooltip
func (b *TooltipBuilder) Delay(ms int) *TooltipBuilder {
	b.delayMs = ms
	return b
}

// Build creates the tooltip VNode
func (b *TooltipBuilder) Build() VNode {
	// Create tooltip content with styling using prop-based approach
	tooltipContent := Text(b.text)

	props := Props{
		"_layer":  LayerTooltip,
		"_delay":  b.delayMs,
		"_anchor": b.anchor,
		// Styling via props
		"fg":      "yellow",
		"bg":      "black",
	}

	tooltipContent.SetProps(props)
	return tooltipContent.SetLayer(LayerTooltip)
}

// =============================================================================
// Convenience Functions
// =============================================================================

// ModalContent is a convenience function to create a simple modal with content
func ModalContent(content VNode, onClose func()) VNode {
	return Modal(content).OnClose(onClose).Build()
}

// OverlayContent is a convenience function to create a simple overlay
func OverlayContent(content VNode) VNode {
	return Overlay(content).Build()
}

// TooltipText is a convenience function to create a simple tooltip
func TooltipText(text string) VNode {
	return Tooltip(text).Build()
}

// =============================================================================
// Layer Detection Helpers
// =============================================================================

// IsModal checks if a VNode is in the modal layer
func IsModal(vnode VNode) bool {
	return vnode != nil && vnode.GetLayer() == LayerModal
}

// IsOverlay checks if a VNode is in the overlay layer
func IsOverlay(vnode VNode) bool {
	return vnode != nil && vnode.GetLayer() == LayerOverlay
}

// IsTooltip checks if a VNode is in the tooltip layer
func IsTooltip(vnode VNode) bool {
	return vnode != nil && vnode.GetLayer() == LayerTooltip
}

// IsBaseLayer checks if a VNode is in the base layer
func IsBaseLayer(vnode VNode) bool {
	return vnode == nil || vnode.GetLayer() == LayerBase
}
