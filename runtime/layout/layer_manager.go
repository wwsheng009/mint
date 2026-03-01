package layout

// LayerManager manages layer-specific layout transformations
// It handles layer collection, positioning, and normalization
type LayerManager struct {
	// Group boxes by layer
	layers map[Layer][]*LayoutBox

	// Offset to normalize each layer's coordinates to (0, 0)
	// For example, centered modals have their initial position adjusted before normalization
	layerOffsets map[Layer]Point

	// Container constraints for calculating transformations
	constraints Constraints
}

// NewLayerManager creates a new layer manager
func NewLayerManager() *LayerManager {
	return &LayerManager{
		layers:       make(map[Layer][]*LayoutBox),
		layerOffsets: make(map[Layer]Point),
	}
}

// CollectLayers traverses the layout tree and collects boxes by layer
func (lm *LayerManager) CollectLayers(root *LayoutBox) {
	if root == nil {
		return
	}

	var walk func(box *LayoutBox)
	walk = func(box *LayoutBox) {
		if box == nil {
			return
		}

		layer := box.Layer
		lm.layers[layer] = append(lm.layers[layer], box)

		for _, child := range box.Children {
			walk(child)
		}
	}
	walk(root)
}

// SetConstraints sets the container constraints for transformation calculations
func (lm *LayerManager) SetConstraints(constraints Constraints) {
	lm.constraints = constraints
}

// ApplyLayerTransforms normalizes all layer coordinates to (0, 0)
// This is called after layout but before painting
// Note: Centering/positioning is handled by the layout engine, not LayerManager
func (lm *LayerManager) ApplyLayerTransforms() {
	// For each layer, find min X/Y for normalization
	for layer := range lm.layers {
		boxes := lm.layers[layer]
		if len(boxes) == 0 {
			continue
		}

		// Find min X/Y
		minX, minY := boxes[0].X, boxes[0].Y
		for _, box := range boxes {
			if box.X < minX {
				minX = box.X
			}
			if box.Y < minY {
				minY = box.Y
			}
		}

		// Store offset for this layer
		if minX != 0 || minY != 0 {
			lm.layerOffsets[layer] = Point{X: minX, Y: minY}
		}
	}

	// Normalize all layer coordinates to (0, 0)
	lm.normalizeLayerCoordinates()
}

// normalizeLayerCoordinates normalizes all layer coordinates to start from (0, 0)
// This allows each layer to be rendered independently with consistent coordinate system
func (lm *LayerManager) normalizeLayerCoordinates() {
	for layer, offset := range lm.layerOffsets {
		if offset.X == 0 && offset.Y == 0 {
			continue
		}

		// Shift all boxes in this layer by subtracting the offset
		for _, box := range lm.layers[layer] {
			box.X -= offset.X
			box.Y -= offset.Y
		}

		// Reset offset after normalization
		lm.layerOffsets[layer] = Point{X: 0, Y: 0}
	}
}

// GetLayerBoxes returns all boxes for a given layer
func (lm *LayerManager) GetLayerBoxes(layer Layer) []*LayoutBox {
	return lm.layers[layer]
}

// GetLayers returns all layers that have boxes
func (lm *LayerManager) GetLayers() []Layer {
	layers := make([]Layer, 0, len(lm.layers))
	for layer := range lm.layers {
		layers = append(layers, layer)
	}
	return layers
}

// CountBoxes returns the total number of boxes across all layers
func (lm *LayerManager) CountBoxes() int {
	count := 0
	for _, boxes := range lm.layers {
		count += len(boxes)
	}
	return count
}
