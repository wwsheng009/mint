package paint

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// Layer Type - 使用 types.Layer 统一类型
// =============================================================================

// RenderLayer 是 types.Layer 的类型别名，保持向后兼容
// Deprecated: 请直接使用 types.Layer
type RenderLayer = types.Layer

// 层级常量 - 引用 types 包的统一常量
const (
	RenderLayerBase      = types.LayerBase
	RenderLayerOverlay   = types.LayerOverlay
	RenderLayerModal     = types.LayerModal
	RenderLayerTooltip   = types.LayerTooltip
	RenderLayerInspector = types.LayerInspector
)

// =============================================================================
// PaintablePlanes - 多层渲染平面管理 (解耦版本)
// =============================================================================
// PaintablePlanes 管理多个渲染平面，每层包含该层节点的 PaintableBox 集合
// 这是 RenderPlanes 的解耦版本，不依赖 compute.ComputedBox
//
// 架构优势：
// 1. 完全解耦：不依赖 compute 或 layout 包
// 2. 分离关注点：每层独立管理，易于理解和维护
// 3. 层级化渲染：按 Z-order 从低到高渲染
// 4. 层级化事件处理：HitTest 从高到低检查，先命中上层元素

// PaintablePlanes 管理多个渲染平面
type PaintablePlanes struct {
	// planes 存储每层的 PaintableBox 集合
	// LayerBase(0) < LayerOverlay(1) < LayerModal(2) < LayerTooltip(3) < LayerInspector(4)
	planes map[RenderLayer][]*PaintableBox

	// renderOrder 存储渲染顺序（从低层到高层）
	renderOrder []RenderLayer
}

// NewPaintablePlanes 创建新的 PaintablePlanes
func NewPaintablePlanes() *PaintablePlanes {
	return &PaintablePlanes{
		planes: make(map[RenderLayer][]*PaintableBox),
		renderOrder: []RenderLayer{
			RenderLayerBase,
			RenderLayerOverlay,
			RenderLayerModal,
			RenderLayerTooltip,
			RenderLayerInspector,
		},
	}
}

// =============================================================================
// PaintablePlanes API
// =============================================================================

// AddToLayer 添加一个 PaintableBox 到指定层
func (pp *PaintablePlanes) AddToLayer(layer RenderLayer, box *PaintableBox) {
	if box == nil {
		return
	}

	_, ok := pp.planes[layer]
	if !ok {
		pp.planes[layer] = make([]*PaintableBox, 0)
	}

	pp.planes[layer] = append(pp.planes[layer], box)
}

// GetLayer 获取指定层的所有 PaintableBox
func (pp *PaintablePlanes) GetLayer(layer RenderLayer) []*PaintableBox {
	return pp.planes[layer]
}

// GetAllLayers 获取所有层的键
func (pp *PaintablePlanes) GetAllLayers() []RenderLayer {
	keys := make([]RenderLayer, 0, len(pp.planes))
	for layer := range pp.planes {
		keys = append(keys, layer)
	}
	return keys
}

// GetRenderOrder 获取渲染顺序（从低层到高层）
func (pp *PaintablePlanes) GetRenderOrder() []RenderLayer {
	return pp.renderOrder
}

// IsLayerEmpty 检查指定层是否为空
func (pp *PaintablePlanes) IsLayerEmpty(layer RenderLayer) bool {
	boxes, ok := pp.planes[layer]
	return !ok || len(boxes) == 0
}

// CountBoxes 返回所有层的 PaintableBox 总数
func (pp *PaintablePlanes) CountBoxes() int {
	total := 0
	for _, boxes := range pp.planes {
		total += len(boxes)
	}
	return total
}

// HasLayer 检查是否有指定层级的节点
func (pp *PaintablePlanes) HasLayer(layer RenderLayer) bool {
	return !pp.IsLayerEmpty(layer)
}

// GetHighestLayer 返回有节点的最高层级
func (pp *PaintablePlanes) GetHighestLayer() RenderLayer {
	for i := len(pp.renderOrder) - 1; i >= 0; i-- {
		layer := pp.renderOrder[i]
		if pp.HasLayer(layer) {
			return layer
		}
	}
	return RenderLayerBase
}

// Clear 清空所有层
func (pp *PaintablePlanes) Clear() {
	pp.planes = make(map[RenderLayer][]*PaintableBox)
}

// =============================================================================
// 遍历方法
// =============================================================================

// Iterate 从低层到高层遍历所有 PaintableBox
func (pp *PaintablePlanes) Iterate(callback func(layer RenderLayer, box *PaintableBox) bool) {
	for _, layer := range pp.renderOrder {
		for _, box := range pp.planes[layer] {
			if !callback(layer, box) {
				return
			}
		}
	}
}

// IterateReverse 从高层到低层遍历所有 PaintableBox（用于 HitTest）
func (pp *PaintablePlanes) IterateReverse(callback func(layer RenderLayer, box *PaintableBox) bool) {
	for i := len(pp.renderOrder) - 1; i >= 0; i-- {
		layer := pp.renderOrder[i]
		for j := len(pp.planes[layer]) - 1; j >= 0; j-- {
			box := pp.planes[layer][j]
			if !callback(layer, box) {
				return
			}
		}
	}
}

// =============================================================================
// 从 PaintableBox 树构建
// =============================================================================

// BuildFromPaintableBox 从 PaintableBox 树构建 PaintablePlanes
func BuildFromPaintableBox(root *PaintableBox) *PaintablePlanes {
	if root == nil {
		return NewPaintablePlanes()
	}

	pp := NewPaintablePlanes()

	var walk func(box *PaintableBox)
	walk = func(box *PaintableBox) {
		if box == nil {
			return
		}

		// 添加到对应层
		pp.AddToLayer(RenderLayer(box.Layer), box)

		// 递归处理子节点
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return pp
}

// =============================================================================
// 从 PaintableLayouts 构建
// =============================================================================

// PaintableLayouts 映射每个 RenderLayer 到其 PaintableLayout
type PaintableLayouts map[RenderLayer]*PaintableLayout

// BuildPaintablePlanesFromLayouts 从 PaintableLayouts 构建 PaintablePlanes
func BuildPaintablePlanesFromLayouts(layouts PaintableLayouts) *PaintablePlanes {
	pp := NewPaintablePlanes()

	renderOrder := []RenderLayer{
		RenderLayerBase,
		RenderLayerOverlay,
		RenderLayerModal,
		RenderLayerTooltip,
		RenderLayerInspector,
	}

	for _, layer := range renderOrder {
		layout, ok := layouts[layer]
		if !ok || layout.Root == nil {
			continue
		}

		var walk func(box *PaintableBox)
		walk = func(box *PaintableBox) {
			if box == nil {
				return
			}

			pp.AddToLayer(layer, box)

			for _, child := range box.Children {
				walk(child)
			}
		}

		walk(layout.Root)
	}

	return pp
}

// =============================================================================
// Debug Support
// =============================================================================

// DebugInfo 返回 PaintablePlanes 的调试信息
func (pp *PaintablePlanes) DebugInfo() string {
	var sb strings.Builder

	sb.WriteString("=== PaintablePlanes DebugInfo ===\n")
	sb.WriteString(fmt.Sprintf("Total Boxes: %d\n", pp.CountBoxes()))

	for _, layer := range pp.renderOrder {
		boxes := pp.planes[layer]
		sb.WriteString(fmt.Sprintf("\nLayer %s (%d boxes):\n", layer.String(), len(boxes)))

		for _, box := range boxes {
			tag := ""
			if box.Node != nil {
				tag = box.Node.Tag()
			}
			sb.WriteString(fmt.Sprintf("  NodeID=%d Tag=%s Bounds=(%d,%d,%dx%d) Layer=%d\n",
				box.NodeID, tag, box.X, box.Y, box.Width, box.Height, box.Layer))
		}
	}

	return sb.String()
}
