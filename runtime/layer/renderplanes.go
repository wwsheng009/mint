package layer

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// RenderPlanes - 多层渲染平面管理
// =============================================================================
// RenderPlanes 管理多个渲染平面，每层包含该层节点的 ComputedBox 集合
// 这是 "Render Plane Projection" 模式的核心实现
//
// 架构优势：
// 1. 分离关注点：每层独立管理，易于理解和维护
// 2. 层级化渲染：按 Z-order 从低到高渲染
// 3. 层级化事件处理：HitTest 从高到低检查，先命中上层元素
//
// 参考文档：
// - docs/render/fiber/diff_layer.md
// - docs/plan/fiber/TODO_LIST.md Phase 3

// RenderPlanes 管理多个渲染平面
type RenderPlanes struct {
	// planes 存储每层的 ComputedBox 集合
	// LayerBase(0) < LayerOverlay(1) < LayerModal(2) < LayerTooltip(3) < LayerInspector(4)
	planes map[rtui.Layer][]*compute.ComputedBox

	// renderOrder 存储渲染顺序（从低层到高层）
	// [LayerBase, LayerOverlay, LayerModal, LayerTooltip, LayerInspector]
	renderOrder []rtui.Layer
}

// NewRenderPlanes 创建新的 RenderPlanes
func NewRenderPlanes() *RenderPlanes {
	return &RenderPlanes{
		planes:      make(map[rtui.Layer][]*compute.ComputedBox),
		renderOrder: []rtui.Layer{
			rtui.LayerBase,
			rtui.LayerOverlay,
			rtui.LayerModal,
			rtui.LayerTooltip,
			rtui.LayerInspector,
		},
	}
}

// =============================================================================
// RenderPlanes API
// =============================================================================

// AddToLayer 添加一个 ComputedBox 到指定层
// 如果层不存在则自动创建
func (rp *RenderPlanes) AddToLayer(layer rtui.Layer, box *compute.ComputedBox) {
	if box == nil {
		return
	}

	_, ok := rp.planes[layer]
	if !ok {
		rp.planes[layer] = make([]*compute.ComputedBox, 0)
	}

	rp.planes[layer] = append(rp.planes[layer], box)
}

// GetLayer 获取指定层的所有 ComputedBox
// 如果层不存在则返回 nil
func (rp *RenderPlanes) GetLayer(layer rtui.Layer) []*compute.ComputedBox {
	return rp.planes[layer]
}

// GetAllLayers 获取所有层的键
func (rp *RenderPlanes) GetAllLayers() []rtui.Layer {
	keys := make([]rtui.Layer, 0, len(rp.planes))
	for layer := range rp.planes {
		keys = append(keys, layer)
	}
	return keys
}

// GetRenderOrder 获取渲染顺序（从低层到高层）
// 返回的顺序确保低层先渲染，高层后渲染（覆盖效果）
func (rp *RenderPlanes) GetRenderOrder() []rtui.Layer {
	return rp.renderOrder
}

// IsLayerEmpty 检查指定层是否为空
func (rp *RenderPlanes) IsLayerEmpty(layer rtui.Layer) bool {
	boxes, ok := rp.planes[layer]
	return !ok || len(boxes) == 0
}

// CountBoxes 返回所有层的 ComputedBox 总数
func (rp *RenderPlanes) CountBoxes() int {
	total := 0
	for _, boxes := range rp.planes {
		total += len(boxes)
	}
	return total
}

// HasLayer 检查是否有指定层级的节点
func (rp *RenderPlanes) HasLayer(layer rtui.Layer) bool {
	return !rp.IsLayerEmpty(layer)
}

// GetHighestLayer 返回有节点的最高层级
func (rp *RenderPlanes) GetHighestLayer() rtui.Layer {
	// 从高到低检查
	for i := len(rp.renderOrder) - 1; i >= 0; i-- {
		layer := rp.renderOrder[i]
		if rp.HasLayer(layer) {
			return layer
		}
	}
	return rtui.LayerBase
}

// Clear 清空所有层
func (rp *RenderPlanes) Clear() {
	rp.planes = make(map[rtui.Layer][]*compute.ComputedBox)
}

// BuildFromFiber 从 Fiber 树构建 RenderPlanes
// 遍历整个 Fiber 树，提取每个 Fiber 的 ComputedBox 并按 Layer 分组
//
// 参数：
//   root - Fiber 树的根节点
//
// 返回：
//   *RenderPlanes - 包含所有分层 ComputedBox 的 RenderPlanes
//
// 注意：此方法假设 Fiber 节点已经完成了布局（即 fiber.ComputedBox 已设置）
func BuildFromFiber(root *rtui.Fiber) *RenderPlanes {
	if root == nil {
		return NewRenderPlanes()
	}

	rp := NewRenderPlanes()

	// 遍历 Fiber 树并收集 ComputedBox
	var walk func(fiber *rtui.Fiber)
	walk = func(fiber *rtui.Fiber) {
		if fiber == nil {
			return
		}

		// 获取 ComputedBox
		if fiber.ComputedBox != nil {
			box, ok := fiber.ComputedBox.(*compute.ComputedBox)
			if ok && box != nil {
				// 添加到对应层
				rp.AddToLayer(fiber.Layer, box)
			}
		}

		// 递归处理子节点（Fiber 树：Child -> Sibling）
		for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
			walk(childFiber)
		}
	}

	walk(root)
	return rp
}

// =============================================================================
// Helper Functions
// =============================================================================

// BuildRenderPlanes 从 ComputedBox 树构建 RenderPlanes
// 遍历整个 ComputeBox 树，按 Layer 分组收集所有 ComputedBox
//
// 参数：
//   root - ComputedBox 树的根节点
//
// 返回：
//   *RenderPlanes - 包含所有分层 ComputedBox 的 RenderPlanes
func BuildRenderPlanes(root *compute.ComputedBox) *RenderPlanes {
	if root == nil {
		return NewRenderPlanes()
	}

	rp := NewRenderPlanes()

	// 遍历 ComputedBox 树并按 Layer 分组
	var walk func(box *compute.ComputedBox)
	walk = func(box *compute.ComputedBox) {
		if box == nil {
			return
		}

		// 添加到对应层
		rp.AddToLayer(box.Layer, box)

		// 递归处理子节点
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return rp
}

// =============================================================================
// Rendering Support
// =============================================================================

// Iterate 从低层到高层遍历所有 ComputedBox
// 回调函数接收：layer, box, 返回 false 时停止遍历
func (rp *RenderPlanes) Iterate(callback func(layer rtui.Layer, box *compute.ComputedBox) bool) {
	for _, layer := range rp.renderOrder {
		for _, box := range rp.planes[layer] {
			if !callback(layer, box) {
				return
			}
		}
	}
}

// IterateReverse 从高层到低层遍历所有 ComputedBox
// 用于 HitTest：先检测上层元素，命中则停止
func (rp *RenderPlanes) IterateReverse(callback func(layer rtui.Layer, box *compute.ComputedBox) bool) {
	// 从渲染顺序的末尾遍历（高层到低层）
	for i := len(rp.renderOrder) - 1; i >= 0; i-- {
		layer := rp.renderOrder[i]
		for j := len(rp.planes[layer]) - 1; j >= 0; j-- {
			box := rp.planes[layer][j]
			if !callback(layer, box) {
				return
			}
		}
	}
}

// =============================================================================
// Debug Support
// =============================================================================

// DebugInfo 返回 RenderPlanes 的调试信息
func (rp *RenderPlanes) DebugInfo() string {
	var sb strings.Builder

	sb.WriteString("=== RenderPlanes DebugInfo ===\n")
	sb.WriteString(fmt.Sprintf("Total Boxes: %d\n", rp.CountBoxes()))

	for _, layer := range rp.renderOrder {
		boxes := rp.planes[layer]
		sb.WriteString(fmt.Sprintf("\nLayer %s (%d boxes):\n", layer.String(), len(boxes)))

		for _, box := range boxes {
			sb.WriteString(fmt.Sprintf("  NodeID=%d Bounds=(%d,%d,%dx%d) Layer=%d\n",
				box.NodeID, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, box.Layer))
		}
	}

	return sb.String()
}

// Validate 验证 RenderPlanes 的完整性
// 检查：所有 Box 的 Layer 和其所在层是否一致
func (rp *RenderPlanes) Validate() []error {
	var errors []error

	for layer, boxes := range rp.planes {
		for _, box := range boxes {
			if box.Layer != layer {
				errors = append(errors, fmt.Errorf("box.NodeID=%d Layer mismatch: expected %d, got %d",
					box.NodeID, layer, box.Layer))
			}
		}
	}

	return errors
}
