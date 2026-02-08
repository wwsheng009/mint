// Package layer provides layer collection and management for TUI components
package layer

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// LayerNode
// =============================================================================

// LayerNode represents a node in a specific layer
type LayerNode struct {
	// Layer is the visual layer this node belongs to
	Layer rtui.Layer

	// ID is a unique identifier for this node
	ID string

	// Content is the VNode to render in this layer
	Content rtui.VNode

	// Visible indicates whether this node should be rendered
	Visible bool

	// FocusID is the ID of the component that should have focus when this layer is active
	FocusID string
}

// NewLayerNode creates a new layer node
func NewLayerNode(layer rtui.Layer, id string, content rtui.VNode) *LayerNode {
	return &LayerNode{
		Layer:   layer,
		ID:      id,
		Content: content,
		Visible: true,
	}
}

// SetVisible sets the visibility of the layer node
func (n *LayerNode) SetVisible(visible bool) {
	n.Visible = visible
}

// =============================================================================
// LayerMap
// =============================================================================

// LayerMap is a map of layer to nodes in that layer
type LayerMap map[rtui.Layer][]*LayerNode

// NewLayerMap creates a new layer map
func NewLayerMap() LayerMap {
	return make(LayerMap)
}

// Add adds a node to a layer
func (m LayerMap) Add(layer rtui.Layer, node *LayerNode) {
	m[layer] = append(m[layer], node)
}

// Get returns all nodes in a layer
func (m LayerMap) Get(layer rtui.Layer) []*LayerNode {
	return m[layer]
}

// HasModal checks if any modal layer nodes exist
func (m LayerMap) HasModal() bool {
	return len(m[rtui.LayerModal]) > 0
}

// HasOverlay checks if any overlay layer nodes exist
func (m LayerMap) HasOverlay() bool {
	for l := rtui.LayerOverlay; l <= rtui.LayerInspector; l++ {
		if len(m[l]) > 0 {
			return true
		}
	}
	return false
}

// HasInspector checks if any inspector layer nodes exist
func (m LayerMap) HasInspector() bool {
	return len(m[rtui.LayerInspector]) > 0
}

// GetHighestLayer returns the highest layer that has visible nodes
func (m LayerMap) GetHighestLayer() rtui.Layer {
	for l := rtui.LayerInspector; l >= rtui.LayerBase; l-- {
		if len(m[l]) > 0 {
			for _, node := range m[l] {
				if node.Visible {
					return l
				}
			}
		}
	}
	return rtui.LayerBase
}

// Clear removes all nodes from all layers
func (m LayerMap) Clear() {
	for l := range m {
		delete(m, l)
	}
}

// =============================================================================
// Collector
// =============================================================================

// Collector collects layer nodes from a VNode tree
// It identifies nodes that should be rendered in different visual layers
type Collector struct {
	layers LayerMap
}

// NewCollector creates a new layer collector
func NewCollector() *Collector {
	return &Collector{
		layers: NewLayerMap(),
	}
}

// Collect walks the VNode tree and collects nodes with layer properties
// Nodes with a layer property are extracted and not included in the base tree
func (c *Collector) Collect(vnode rtui.VNode) {
	c.layers.Clear()
	c.walk(vnode)
}

// walk recursively walks the VNode tree, collecting layer nodes
func (c *Collector) walk(vnode rtui.VNode) {
	if vnode == nil {
		return
	}

	// Check if this node has a non-base layer
	if layer := vnode.GetLayer(); layer != rtui.LayerBase && layer.IsValid() {

		// Create a layer node for this content
		node := &LayerNode{
			Layer:   layer,
			ID:      c.getNodeID(vnode),
			Content: vnode,
			Visible: c.isVisible(vnode),
		}
		c.layers.Add(layer, node)

		// Don't traverse children of layer nodes
		// They will be rendered separately in their own layout pass
		return
	}

	// Recursively walk children
	for _, child := range vnode.Children() {
		c.walk(child)
	}
}

// getNodeID returns a unique identifier for a node
func (c *Collector) getNodeID(vnode rtui.VNode) string {
	if key := vnode.Key(); key != "" {
		return key
	}
	if tag := vnode.Tag(); tag != "" {
		return tag
	}
	return vnode.Type().String()
}

// isVisible checks if a node should be visible
func (c *Collector) isVisible(vnode rtui.VNode) bool {
	props := vnode.Props()
	if props == nil {
		return true
	}

	// Check for hidden property
	if hidden, ok := props["hidden"].(bool); ok {
		return !hidden
	}

	// Check for visible property
	if visible, ok := props["visible"].(bool); ok {
		return visible
	}

	return true
}

// GetLayers returns the collected layer map
func (c *Collector) GetLayers() LayerMap {
	return c.layers
}

// HasModal checks if any modal nodes were collected
func (c *Collector) HasModal() bool {
	return c.layers.HasModal()
}

// HasOverlay checks if any overlay nodes were collected
func (c *Collector) HasOverlay() bool {
	return c.layers.HasOverlay()
}

// GetHighestLayer returns the highest layer with visible nodes
func (c *Collector) GetHighestLayer() rtui.Layer {
	return c.layers.GetHighestLayer()
}

// StripLayers returns a new VNode tree with layer nodes removed
// This creates a "clean" base tree for normal rendering
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// If this node itself is a layer node, return nil
	// (it will be rendered separately)
	if vnode.GetLayer() != rtui.LayerBase {
		return nil
	}

	// Clone the node and filter out layer children
	cloned := c.cloneWithoutLayers(vnode)
	return cloned
}

// cloneWithoutLayers creates a copy of a VNode without layer children
func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// Get non-layer children
	var nonLayerChildren []rtui.VNode
	for _, child := range vnode.Children() {
		if child.GetLayer() == rtui.LayerBase {
			// Recursively filter this child's children
			nonLayerChildren = append(nonLayerChildren, c.cloneWithoutLayers(child))
		}
		// Layer children are simply omitted
	}

	// If no children changed, return original
	if len(nonLayerChildren) == len(vnode.Children()) {
		return vnode
	}

	// Create a new node with filtered children
	switch n := vnode.(type) {
	case *rtui.ElementVNode:
		cloned := rtui.NewElement(n.Tag())
		cloned.SetProps(n.Props().Clone())
		cloned.SetStyle(n.Style())
		cloned.SetKey(n.Key())
		cloned.SetChildren(nonLayerChildren)
		return cloned
	case *rtui.LayoutNode:
		// LayoutNode embeds ElementVNode, handle specially to preserve layout properties
		cloned := rtui.NewElement(n.Tag())
		cloned.SetProps(n.Props().Clone())
		cloned.SetStyle(n.Style())
		cloned.SetKey(n.Key())
		cloned.SetChildren(nonLayerChildren)
		return cloned
	case *rtui.BorderedNode:
		// BorderedNode embeds ElementVNode, handle specially
		cloned := rtui.NewElement(n.Tag())
		cloned.SetProps(n.Props().Clone())
		cloned.SetStyle(n.Style())
		cloned.SetKey(n.Key())
		cloned.SetChildren(nonLayerChildren)
		return cloned
	case *rtui.ComponentVNode:
		cloned := rtui.NewComponent(n.Name(), nil)
		cloned.SetProps(n.Props().Clone())
		cloned.SetKey(n.Key())
		cloned.SetChildren(nonLayerChildren)
		return cloned
	case *rtui.FragmentVNode:
		cloned := rtui.NewFragment()
		cloned.SetChildren(nonLayerChildren)
		return cloned
	case *rtui.TextVNode:
		// Text nodes don't have children
		return vnode
	default:
		// For unknown types, try to set children if the interface supports it
		if len(nonLayerChildren) != len(vnode.Children()) {
			vnode.SetChildren(nonLayerChildren)
		}
		return vnode
	}
}

// GetLayerNodes returns all nodes in a specific layer
func (c *Collector) GetLayerNodes(layer rtui.Layer) []*LayerNode {
	return c.layers.Get(layer)
}

// GetModalNodes returns all modal layer nodes
func (c *Collector) GetModalNodes() []*LayerNode {
	return c.layers.Get(rtui.LayerModal)
}

// GetOverlayNodes returns all overlay layer nodes
func (c *Collector) GetOverlayNodes() []*LayerNode {
	return c.layers.Get(rtui.LayerOverlay)
}

// GetTooltipNodes returns all tooltip layer nodes
func (c *Collector) GetTooltipNodes() []*LayerNode {
	return c.layers.Get(rtui.LayerTooltip)
}

// GetInspectorNodes returns all inspector layer nodes
func (c *Collector) GetInspectorNodes() []*LayerNode {
	return c.layers.Get(rtui.LayerInspector)
}

// HasInspector checks if any inspector nodes were collected
func (c *Collector) HasInspector() bool {
	return c.layers.HasInspector()
}
