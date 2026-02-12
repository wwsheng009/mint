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
		if child.VNode != nil {
			childTypeID := pg.getTypeIdentifier(child.VNode)
			if childTypeID == typeID {
				count++
			}
		}
		child = child.Sibling
	}

	return count
}

// getTypeIndexFromVNodes 获取组件类型在VNode数组中的索引
// 统计在siblingIndex之前有多少同类型的VNode
// 当Fiber节点还未链接时使用此方法
func (pg *PathGenerator) getTypeIndexFromVNodes(
	children []rtui.VNode,
	typeID string,
	siblingIndex int,
) int {
	if siblingIndex <= 0 {
		return 0
	}

	// 统计在当前索引之前有多少同类型的VNode
	count := 0
	for i := 0; i < siblingIndex && i < len(children); i++ {
		childTypeID := pg.getTypeIdentifier(children[i])
		if childTypeID == typeID {
			count++
		}
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

// GeneratePathWithIndex generates a path using a pre-calculated type index
// This is used by createAllNewChildren where type indices are tracked externally
func (pg *PathGenerator) GeneratePathWithIndex(
	parent *Fiber,
	vnode rtui.VNode,
	siblingIndex int,
	typeIndex int,
) string {
	// 1. Check if it's root node
	if parent == nil {
		return pg.generateRootPath(vnode)
	}

	// 2. Get component type identifier
	typeID := pg.getTypeIdentifier(vnode)

	// 3. Use the provided type index
	segment := fmt.Sprintf("%s[%d]", typeID, typeIndex)

	// 4. Concatenate full path
	return parent.Path + "/" + segment
}

// GeneratePathFromVNodes generates a path using VNode array for type counting
// This is used when fibers are not yet linked to parent
func (pg *PathGenerator) GeneratePathFromVNodes(
	parent *Fiber,
	vnode rtui.VNode,
	siblingIndex int,
	children []rtui.VNode,
) string {
	// 1. Check if it's root node
	if parent == nil {
		return pg.generateRootPath(vnode)
	}

	// 2. Get component type identifier
	typeID := pg.getTypeIdentifier(vnode)

	// 3. Get type index from VNode array
	typeIndex := pg.getTypeIndexFromVNodes(children, typeID, siblingIndex)

	// 4. Generate path segment
	segment := fmt.Sprintf("%s[%d]", typeID, typeIndex)

	// 5. Concatenate full path
	return parent.Path + "/" + segment
}
