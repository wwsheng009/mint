package menu

import (
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func popupPlacementActive(model Model) bool {
	return model.AnchorID != "" && model.Placement != ""
}

func resolvePopupPlacement(model Model) Placement {
	if !popupPlacementActive(model) {
		return ""
	}
	if model.Placement == PlacementAuto {
		switch model.Anchor {
		case rttypes.AnchorBottomRight:
			return PlacementBottomEnd
		case rttypes.AnchorTopLeft, rttypes.AnchorTop:
			return PlacementTopStart
		case rttypes.AnchorTopRight:
			return PlacementTopEnd
		case rttypes.AnchorRight:
			return PlacementRightStart
		case rttypes.AnchorLeft:
			return PlacementLeftStart
		default:
			return PlacementBottomStart
		}
	}
	return model.Placement
}

func popupPortalPlacement(model ThemeableModel) (anchor rttypes.Anchor, offsetX, offsetY, width, height int, ok bool) {
	if !popupPlacementActive(model.Model) {
		return 0, 0, 0, 0, 0, false
	}

	metrics := popupMetricsForModel(model.Model, model.Theme, model.Items)
	width = metrics.surfaceWidth
	height = metrics.surfaceHeight

	switch resolvePopupPlacement(model.Model) {
	case PlacementBottomEnd:
		return rttypes.AnchorBottomRight, 0, height, width, height, true
	case PlacementTopStart:
		return rttypes.AnchorTopLeft, 0, -height, width, height, true
	case PlacementTopEnd:
		return rttypes.AnchorTopRight, 0, -height, width, height, true
	case PlacementRightStart:
		return rttypes.AnchorTopRight, width, 0, width, height, true
	case PlacementLeftStart:
		return rttypes.AnchorTopLeft, -width, 0, width, height, true
	default:
		return rttypes.AnchorBottomLeft, 0, height, width, height, true
	}
}

func applyPopupPortalProps(node *rtui.ElementVNode, model ThemeableModel) {
	if node == nil {
		return
	}

	portalModel := model.Model
	if anchor, offsetX, offsetY, width, height, ok := popupPortalPlacement(model); ok {
		portalModel.Anchor = anchor
		portalModel.PortalOffsetX += offsetX
		portalModel.PortalOffsetY += offsetY
		node.SetProp("width", width)
		node.SetProp("height", height)
		node.SetProp("positioningWidth", width)
		node.SetProp("positioningHeight", height)
	}

	applyPortalProps(node, portalModel)
}
