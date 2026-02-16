// Package layout provides layer types for multi-layer rendering support
package layout

// =============================================================================
// Layer Types
// =============================================================================

// Layer defines the rendering layer for a node
// Higher layers are rendered on top of lower layers
type Layer int

const (
	// LayerBase is the default layer for normal content
	LayerBase Layer = iota

	// LayerDropdown is for dropdown menus
	LayerDropdown

	// LayerSticky is for sticky positioned elements
	LayerSticky

	// LayerFixed is for fixed positioned elements
	LayerFixed

	// LayerModalBackdrop is for modal backdrop/overlay
	LayerModalBackdrop

	// LayerModal is for modal dialogs
	LayerModal

	// LayerPopover is for popover elements
	LayerPopover

	// LayerTooltip is for tooltips (highest layer)
	LayerTooltip
)

// String returns the string representation of Layer
func (l Layer) String() string {
	switch l {
	case LayerBase:
		return "base"
	case LayerDropdown:
		return "dropdown"
	case LayerSticky:
		return "sticky"
	case LayerFixed:
		return "fixed"
	case LayerModalBackdrop:
		return "modalBackdrop"
	case LayerModal:
		return "modal"
	case LayerPopover:
		return "popover"
	case LayerTooltip:
		return "tooltip"
	default:
		return "unknown"
	}
}

// ZIndex returns the default z-index for this layer
func (l Layer) ZIndex() int {
	// Each layer gets a 1000-unit range
	return int(l) * 1000
}

// IsHigher returns true if this layer is higher than the other
func (l Layer) IsHigher(other Layer) bool {
	return l > other
}

// IsLower returns true if this layer is lower than the other
func (l Layer) IsLower(other Layer) bool {
	return l < other
}

// =============================================================================
// Layered Interface
// =============================================================================

// Layered nodes can provide layer information
// This is an optional interface that nodes can implement
type Layered interface {
	Node

	// GetLayer returns the node's layer
	GetLayer() Layer

	// GetZIndex returns the node's z-index within its layer
	GetZIndex() int
}

// =============================================================================
// LayeredNode Wrapper
// =============================================================================

// LayeredNode wraps a node to add layer information
type LayeredNode struct {
	child  Node
	layer  Layer
	zIndex int
	id     string
}

// NewLayeredNode creates a new layered wrapper around a child node
func NewLayeredNode(id string, child Node, layer Layer, zIndex int) *LayeredNode {
	return &LayeredNode{
		id:     id,
		child:  child,
		layer:  layer,
		zIndex: zIndex,
	}
}

// ID returns the node identifier
func (n *LayeredNode) ID() string {
	return n.id
}

// Type returns the node type
func (n *LayeredNode) Type() string {
	return "layered"
}

// Children returns the child nodes
func (n *LayeredNode) Children() []Node {
	if n.child == nil {
		return nil
	}
	return []Node{n.child}
}

// GetPosition returns the current position
func (n *LayeredNode) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (n *LayeredNode) SetPosition(x, y int) {
	// Position is handled by parent layout
}

// GetSize returns the current size
func (n *LayeredNode) GetSize() (width, height int) {
	if n.child == nil {
		return 0, 0
	}
	return n.child.GetSize()
}

// SetSize sets the size
func (n *LayeredNode) SetSize(width, height int) {
	// Size is calculated during layout
}

// GetWidth returns the width
func (n *LayeredNode) GetWidth() int {
	w, _ := n.GetSize()
	return w
}

// GetHeight returns the height
func (n *LayeredNode) GetHeight() int {
	_, h := n.GetSize()
	return h
}

// GetLayer returns the layer
func (n *LayeredNode) GetLayer() Layer {
	return n.layer
}

// GetZIndex returns the z-index
func (n *LayeredNode) GetZIndex() int {
	return n.zIndex
}

// GetChild returns the wrapped child node
func (n *LayeredNode) GetChild() Node {
	return n.child
}

// EffectiveZIndex returns the effective z-index (layer + offset)
func (n *LayeredNode) EffectiveZIndex() int {
	return n.layer.ZIndex() + n.zIndex
}

// =============================================================================
// LayeredLayoutResult
// =============================================================================

// LayeredLayoutResult contains layout results organized by layer
type LayeredLayoutResult struct {
	// Root is the root layout box
	Root *LayoutBox

	// Layers maps layer to layout boxes
	Layers map[Layer][]*LayoutBox

	// AllBoxes is a flat list of all boxes for convenience
	AllBoxes []*LayoutBox
}

// NewLayeredLayoutResult creates a new layered layout result
func NewLayeredLayoutResult(root *LayoutBox) *LayeredLayoutResult {
	return &LayeredLayoutResult{
		Root:     root,
		Layers:   make(map[Layer][]*LayoutBox),
		AllBoxes: make([]*LayoutBox, 0),
	}
}

// AddBox adds a box to the appropriate layer
func (r *LayeredLayoutResult) AddBox(box *LayoutBox, layer Layer) {
	r.Layers[layer] = append(r.Layers[layer], box)
	r.AllBoxes = append(r.AllBoxes, box)
}

// GetLayer returns all boxes in a specific layer
func (r *LayeredLayoutResult) GetLayer(layer Layer) []*LayoutBox {
	return r.Layers[layer]
}

// GetLayers returns all layers that have boxes
func (r *LayeredLayoutResult) GetLayers() []Layer {
	layers := make([]Layer, 0, len(r.Layers))
	for layer := range r.Layers {
		layers = append(layers, layer)
	}
	return layers
}

// SortByZIndex sorts all boxes by effective z-index
func (r *LayeredLayoutResult) SortByZIndex() []*LayoutBox {
	// Create a flat list sorted by layer then z-index
	result := make([]*LayoutBox, 0)
	
	// Process layers in order
	for layer := LayerBase; layer <= LayerTooltip; layer++ {
		boxes := r.Layers[layer]
		// Sort boxes within layer by z-index (simple insertion sort for small lists)
		for i := 1; i < len(boxes); i++ {
			for j := i; j > 0; j-- {
				if boxes[j].ZIndex < boxes[j-1].ZIndex {
					boxes[j], boxes[j-1] = boxes[j-1], boxes[j]
				}
			}
		}
		result = append(result, boxes...)
	}
	
	return result
}

// =============================================================================
// Helper Functions
// =============================================================================

// isLayered checks if a node implements Layered interface
func isLayered(node Node) bool {
	_, ok := node.(Layered)
	return ok
}

// GetLayerFromNode safely gets layer from a node
func GetLayerFromNode(node Node) Layer {
	if node == nil {
		return LayerBase
	}
	if l, ok := node.(Layered); ok {
		return l.GetLayer()
	}
	return LayerBase
}

// GetZIndexFromNode safely gets z-index from a node
func GetZIndexFromNode(node Node) int {
	if node == nil {
		return 0
	}
	if l, ok := node.(Layered); ok {
		return l.GetZIndex()
	}
	return 0
}

// CompareZOrder compares two nodes by their z-order
// Returns: -1 if a is below b, 0 if equal, 1 if a is above b
func CompareZOrder(a, b Node) int {
	layerA := GetLayerFromNode(a)
	layerB := GetLayerFromNode(b)

	if layerA < layerB {
		return -1
	}
	if layerA > layerB {
		return 1
	}

	// Same layer, compare z-index
	zA := GetZIndexFromNode(a)
	zB := GetZIndexFromNode(b)

	if zA < zB {
		return -1
	}
	if zA > zB {
		return 1
	}

	return 0
}

// IsInHigherLayer checks if a node is in a higher layer than another
func IsInHigherLayer(a, b Node) bool {
	return GetLayerFromNode(a) > GetLayerFromNode(b)
}

// ParseLayer parses a string to Layer
func ParseLayer(s string) Layer {
	switch s {
	case "base", "":
		return LayerBase
	case "dropdown":
		return LayerDropdown
	case "sticky":
		return LayerSticky
	case "fixed":
		return LayerFixed
	case "modalBackdrop", "modal-backdrop":
		return LayerModalBackdrop
	case "modal":
		return LayerModal
	case "popover":
		return LayerPopover
	case "tooltip":
		return LayerTooltip
	default:
		return LayerBase
	}
}
