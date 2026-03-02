        package ui

import "github.com/wwsheng009/mint/runtime/types"

// =============================================================================
// Portal API - Type-safe helpers for Portal configuration
// =============================================================================
// This file provides both functional helper functions and chainable VNode methods
// for configuring Portal components.
//
// === Chainable Methods (Preferred) ===
// The simplest and most readable way to configure Portals is using the built-in
// VNode methods that are now available on all VNode types:
//
//   rtui.NewElement("portal").
//       SetPortalRoot("tooltip-root").
//       SetAnchorTo("button-1", types.AnchorBottomLeft).
//       SetPortalPosition(types.PositionFixed)
//
//   rtui.NewElement("div").
//       SetID("tooltip-root").
//       SetPortalRootId("tooltip-root")  // Mark as PortalRoot
//
// These methods are available on all VNode implementations:
//   - ElementVNode
//   - ComponentVNode
//   - FragmentVNode
//   - ErrorBoundaryVNode
//   - FiberVNode
//   - MemoVNode
//   - TextVNode (via ElementVNode embedding)
//
// === Functions (Alternative) ===
// The functions below provide a functional-style alternative. They are useful
// when you need to pass Portal configuration as a parameter to another function.
//
//   SetPortalRoot(rtui.NewElement("portal"), "tooltip-root")
//   SetAnchorTo(portalVNode, "button-1", types.AnchorBottomLeft)
//
// === Raw Props (Low-level) ===
// You can also use raw Props directly:
//
//   rtui.NewElement("portal").SetProps(rtui.Props{
//       "portalRoot": "tooltip-root",
//       "anchorId":   "button-1",
//       "anchor":     types.AnchorBottomLeft,
//       "position":   types.PositionFixed,
//   })
// =============================================================================

// SetPortalRoot is a functional helper that sets the portalRoot property
// Prefer using vnode.SetPortalRoot() instead for better readability
func SetPortalRoot(vnode VNode, portalRootID string) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["portalRoot"] = portalRootID
	return vnode.SetProps(props)
}

// SetAnchorTo is a functional helper that sets anchorId and anchor properties
// Prefer using vnode.SetAnchorTo() instead for better readability
func SetAnchorTo(vnode VNode, anchorID string, anchor types.Anchor) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["anchorId"] = anchorID
	props["anchor"] = anchor
	return vnode.SetProps(props)
}

// SetPosition is a functional helper that sets the position property
// Prefer using vnode.SetPortalPosition() instead for better readability
func SetPosition(vnode VNode, position types.PositionType) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["position"] = position
	return vnode.SetProps(props)
}

// SetPortalConfig is a convenience function that combines multiple Portal settings
// This allows setting all Portal properties in a single call
func SetPortalConfig(vnode VNode, portalRootID string, anchorID string, anchor types.Anchor, position types.PositionType, priority int) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["portalRoot"] = portalRootID
	props["anchorId"] = anchorID
	props["anchor"] = anchor
	props["position"] = position
	props["priority"] = priority
	return vnode.SetProps(props)
}

// SetAnchorId is a functional helper that sets only the anchorId property
func SetAnchorId(vnode VNode, anchorID string) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["anchorId"] = anchorID
	return vnode.SetProps(props)
}

// =============================================================================
// PortalRoot API - Type-safe helpers for PortalRoot configuration
// =============================================================================

// SetPortalRootId is a functional helper that sets the portalRootId property
// This marks a node as a PortalRoot (a mounting target for Portals)
// Prefer using vnode.SetPortalRootId() instead for better readability
func SetPortalRootId(vnode VNode, portalRootId string) VNode {
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["portalRootId"] = portalRootId
	return vnode.SetProps(props)
}

// AsPortalRoot is a convenience function to create a PortalRoot element
// Equivalent to: rtui.NewElement("div").SetPortalRootId("tooltip-root")
func AsPortalRoot(portalRootId string) VNode {
	return NewElement("div").SetProps(Props{
		"portalRootId": portalRootId,
	})
}
