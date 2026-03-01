package reconciler

import (
	"fmt"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type PathGenerator struct{}

func NewPathGenerator() *PathGenerator {
	return &PathGenerator{}
}

func (pg *PathGenerator) GeneratePath(
	parent *Fiber,
	vnode rtui.VNode,
	siblingIndex int,
) string {
	if parent == nil {
		return pg.generateRootPath(vnode)
	}
	typeID := pg.getTypeIdentifier(vnode)
	index := pg.getTypeIndex(parent, typeID, siblingIndex)
	segment := fmt.Sprintf("%s[%d]", typeID, index)
	return parent.Path + "/" + segment
}

func (pg *PathGenerator) GeneratePathFromFiber(
	parent *Fiber,
	fiber *rtui.Fiber,
	siblingIndex int,
) string {
	if parent == nil {
		return pg.generateRootPathFromFiber(fiber)
	}
	typeID := pg.getTypeIdentifierFromFiber(fiber)
	index := pg.getTypeIndexFromFiber(parent, typeID, siblingIndex)
	segment := fmt.Sprintf("%s[%d]", typeID, index)
	return parent.Path + "/" + segment
}

func (pg *PathGenerator) GeneratePathWithIndex(
	parent *Fiber,
	vnode rtui.VNode,
	siblingIndex int,
	typeIndex int,
) string {
	if parent == nil {
		return pg.generateRootPath(vnode)
	}
	typeID := pg.getTypeIdentifier(vnode)
	segment := fmt.Sprintf("%s[%d]", typeID, typeIndex)
	return parent.Path + "/" + segment
}

func (pg *PathGenerator) GeneratePathWithIndexFromFiber(
	parent *Fiber,
	fiber *rtui.Fiber,
	siblingIndex int,
	typeIndex int,
) string {
	if parent == nil {
		return pg.generateRootPathFromFiber(fiber)
	}
	typeID := pg.getTypeIdentifierFromFiber(fiber)
	segment := fmt.Sprintf("%s[%d]", typeID, typeIndex)
	return parent.Path + "/" + segment
}

func (pg *PathGenerator) generateRootPath(vnode rtui.VNode) string {
	layer := vnode.GetLayer()
	layerName := getLayerName(layer)
	return fmt.Sprintf("/root/%s[0]", layerName)
}

func (pg *PathGenerator) generateRootPathFromFiber(fiber *rtui.Fiber) string {
	layerName := getLayerName(fiber.Layer)
	return fmt.Sprintf("/root/%s[0]", layerName)
}

func (pg *PathGenerator) getTypeIdentifier(vnode rtui.VNode) string {
	if vnode == nil {
		return "nil"
	}
	switch vnode.Type() {
	case rtui.VNodeComponent:
		if namer, ok := vnode.(interface{ Name() string }); ok {
			return namer.Name()
		}
		return "component"
	case rtui.VNodeElement:
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			return tagger.Tag()
		}
		return "element"
	case rtui.VNodeText:
		return "text"
	case rtui.VNodeFragment:
		return "fragment"
	default:
		return "unknown"
	}
}

func (pg *PathGenerator) getTypeIdentifierFromFiber(fiber *rtui.Fiber) string {
	switch fiber.Type {
	case rtui.VNodeComponent:
		if fiber.ComponentName != "" {
			return fiber.ComponentName
		}
		if fiber.Tag != "" {
			return fiber.Tag
		}
		return "component"
	case rtui.VNodeElement:
		if fiber.Tag != "" {
			return fiber.Tag
		}
		return "element"
	case rtui.VNodeText:
		return "text"
	case rtui.VNodeFragment:
		return "fragment"
	default:
		return "unknown"
	}
}

func (pg *PathGenerator) getTypeIndex(
	parent *Fiber,
	typeID string,
	siblingIndex int,
) int {
	if parent == nil {
		return 0
	}
	count := 0
	child := parent.Child
	for i := 0; i < siblingIndex && child != nil; i++ {
		childTypeID := pg.getTypeIdentifierFromFiber(child)
		if childTypeID == typeID {
			count++
		}
		child = child.Sibling
	}
	return count
}

func (pg *PathGenerator) getTypeIndexFromFiber(
	parent *Fiber,
	typeID string,
	siblingIndex int,
) int {
	return pg.getTypeIndex(parent, typeID, siblingIndex)
}

func (pg *PathGenerator) getTypeIndexFromVNodes(
	children []rtui.VNode,
	typeID string,
	siblingIndex int,
) int {
	if siblingIndex <= 0 {
		return 0
	}
	count := 0
	for i := 0; i < siblingIndex && i < len(children); i++ {
		childTypeID := pg.getTypeIdentifier(children[i])
		if childTypeID == typeID {
			count++
		}
	}
	return count
}

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
