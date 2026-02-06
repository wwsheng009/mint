// Package ui provides a declarative UI framework for terminal applications.
// This package re-exports core types from runtime/ui for backward compatibility.
package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Core Types (re-exported from runtime/ui)
// =============================================================================

// VNode is the virtual node interface - the core of the declarative UI system.
type VNode = rtui.VNode

// VNodeType represents the type of VNode
type VNodeType = rtui.VNodeType

const (
	// VNodeElement is a standard element node (div, span, etc.)
	VNodeElement = rtui.VNodeElement

	// VNodeText is a text node with content
	VNodeText = rtui.VNodeText

	// VNodeComponent is a function component
	VNodeComponent = rtui.VNodeComponent

	// VNodeFragment is a fragment that doesn't add extra DOM nodes
	VNodeFragment = rtui.VNodeFragment
)

// Props represents a map of properties for a VNode
type Props = rtui.Props

// ComponentFunc represents a function component that returns a VNode
type ComponentFunc = rtui.ComponentFunc

// ComponentFuncWithProps represents a component that accepts props
type ComponentFuncWithProps = rtui.ComponentFuncWithProps

// Layer represents a visual rendering layer for overlay components
type Layer = rtui.Layer

const (
	// LayerBase is the default layer for normal UI content
	LayerBase = rtui.LayerBase

	// LayerOverlay is for dropdown menus, popovers, and similar components
	LayerOverlay = rtui.LayerOverlay

	// LayerModal is for modal dialogs that require user attention
	LayerModal = rtui.LayerModal

	// LayerTooltip is for tooltips and hints
	LayerTooltip = rtui.LayerTooltip
)
