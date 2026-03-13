package menu

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propKey = "key"
	propStyle = "style"
)

type barVNode struct{ *rtui.ElementVNode }
type popupVNode struct{ *rtui.ElementVNode }

func newBarVNode(model ThemeableModel) *barVNode {
	node := &barVNode{ElementVNode: rtui.NewElement("menu-bar")}
	if model.ID != "" {
		node.SetID(model.ID)
		node.SetKey(model.ID)
	}
	applyPortalProps(node.ElementVNode, model.Model)
	node.SetProp("model", model.Model)
	node.SetProp("theme", model.Theme)
	node.SetProp("style", model.Style)
	node.SetProp("layer", model.Layer)
	return node
}

func newPopupVNode(model ThemeableModel) *popupVNode {
	tag := "menu-popup"
	if model.Variant == VariantContext {
		tag = "context-menu"
	}
	node := &popupVNode{ElementVNode: rtui.NewElement(tag)}
	if model.ID != "" {
		node.SetID(model.ID)
		node.SetKey(model.ID)
	}
	node.SetProp("model", model.Model)
	node.SetProp("theme", model.Theme)
	node.SetProp("style", model.Style)
	node.SetProp("layer", model.Layer)
	return node
}

func (v *barVNode) SetProps(p rtui.Props) rtui.VNode {
	v.ElementVNode.SetProps(v.ElementVNode.Props().Merge(p))
	return v
}

func (v *popupVNode) SetProps(p rtui.Props) rtui.VNode {
	v.ElementVNode.SetProps(v.ElementVNode.Props().Merge(p))
	return v
}

func (v *barVNode) GetLayer() rtui.Layer {
	if layer, ok := v.Props()["layer"].(rtui.Layer); ok {
		return layer
	}
	return rtui.LayerBase
}

func (v *barVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("layer", l)
	return v
}

func (v *popupVNode) GetLayer() rtui.Layer {
	if layer, ok := v.Props()["layer"].(rtui.Layer); ok {
		return layer
	}
	return rtui.LayerOverlay
}

func (v *popupVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("layer", l)
	return v
}

func (v *barVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props[propKey] = v.Key()
	props[propStyle] = getNodeStyle(v)
	return newBarInstance(props)
}

func (v *popupVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props[propKey] = v.Key()
	props[propStyle] = getNodeStyle(v)
	return newPopupInstance(props)
}

