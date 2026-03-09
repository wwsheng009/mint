package ui

import "testing"

func TestElementVNode_SetLayerUsesProps(t *testing.T) {
	vnode := NewElement("box")
	vnode.SetLayer(LayerOverlay)
	if got := vnode.GetLayer(); got != LayerOverlay {
		t.Fatalf("GetLayer() = %v, want LayerOverlay", got)
	}
}

func TestFragmentVNode_SetLayerUsesProps(t *testing.T) {
	vnode := NewFragment()
	vnode.SetLayer(LayerTooltip)
	if got := vnode.GetLayer(); got != LayerTooltip {
		t.Fatalf("GetLayer() = %v, want LayerTooltip", got)
	}
}
