package reconciler

// =============================================================================
// Path Generator - 自动路径Key生成器
// =============================================================================
// 为静态UI组件自动生成基于路径的唯一Key
// 格式: /root/{layer}[{index}]/{type}[{index}]/...
// =============================================================================

import (
	"fmt"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PathGenerator 生成组件的路径Key
type PathGenerator struct {
	// 预留：未来可以添加缓存
	// segmentCache map[string]int
}

// NewPathGenerator 创建路径生成器
func NewPathGenerator() *PathGenerator {
	return &PathGenerator{
		// segmentCache: make(map[string]int),
	}
}

// GeneratePath 生成组件的Key（基于路径）
// 返回: 完整的路径字符串，如 /root/base[0]/vstack[0]/panel[0]
//
// 参数:
//   parent - 父Fiber节点
//   vnode - 当前VNode
//   siblingIndex - 当前节点在兄弟节点中的索引
//
// 示例:
//   GeneratePath(nil, vnode, 0) → "/root/base[0]"
//   GeneratePath(parent, vnode, 1) → "/root/base[0]/vstack[0]/panel[1]"
func (pg *PathGenerator) GeneratePath(
	parent *Fiber,
	vnode rtui.VNode,
	siblingIndex int,
) string {
	// 1. 检查是否是根节点
	if parent == nil {
		return pg.generateRootPath(vnode)
	}

	// 2. 获取组件类型标识
	typeID := pg.getTypeIdentifier(vnode)

	// 3. 获取该类型的索引
	index := pg.getTypeIndex(parent, typeID, siblingIndex)

	// 4. 生成路径段
	segment := fmt.Sprintf("%s[%d]", typeID, index)

	// 5. 拼接完整路径
	return parent.Path + "/" + segment
}

// generateRootPath 生成根节点路径
func (pg *PathGenerator) generateRootPath(vnode rtui.VNode) string {
	layer := vnode.GetLayer()
	layerName := getLayerName(layer)
	return fmt.Sprintf("/root/%s[0]", layerName)
}

// getTypeIdentifier 获取组件的类型标识
// 用于路径段，如 "button", "panel", "vstack"
func (pg *PathGenerator) getTypeIdentifier(vnode rtui.VNode) string {
	switch v := vnode.(type) {
	case *rtui.ComponentVNode:
		return v.Name()
	case *rtui.ElementVNode:
		return v.Tag()
	case *rtui.TextVNode:
		return "text"
	case *rtui.FragmentVNode:
		return "fragment"
	default:
		return "unknown"
	}
}

// getTypeIndex 获取组件类型在父节点中的索引
// 统计在siblingIndex之前有多少同类型的兄弟节点
func (pg *PathGenerator) getTypeIndex(
	parent *Fiber,
	typeID string,
	siblingIndex int,
) int {
	if parent == nil {
		return 0
	}

	// 统计在当前索引之前有多少同类型的兄弟节点
	count := 0
	child := parent.Child

	for i := 0; i < siblingIndex && child != nil; i++ {
		childTypeID := pg.getTypeIdentifier(child.VNode)
		if childTypeID == typeID {
			count++
		}
		child = child.Sibling
	}

	return count
}

// getLayerName 获取Layer的名称
func getLayerName(layer rtui.Layer) string {
	switch layer {
	case rtui.LayerBase:
		return "base"
	case rtui.LayerOverlay:
		return "overlay"
	case rtui.LayerModal:
		return "modal"
	case rtui.LayerTooltip:
		return "tooltip"
	case rtui.LayerInspector:
		return "inspector"
	default:
		return "unknown"
	}
}
