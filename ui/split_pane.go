package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/splitpane"
)

type SplitPaneDirection = splitpane.Direction

const (
	SplitPaneHorizontal = splitpane.DirectionHorizontal
	SplitPaneVertical   = splitpane.DirectionVertical
)

type SplitPaneBuilder = splitpane.Builder
type SplitPaneVNode = splitpane.VNode

// NewSplitPaneBuilder creates a SplitPane builder for left/right or top/bottom layouts.
func NewSplitPaneBuilder() *splitpane.Builder {
	return splitpane.NewBuilder()
}

// SplitPane creates a horizontal SplitPane from two panes.
func SplitPane(primary, secondary rtui.VNode) rtui.VNode {
	return splitpane.Of(primary, secondary)
}

// HorizontalSplitPane creates a left/right SplitPane.
func HorizontalSplitPane(primary, secondary rtui.VNode) rtui.VNode {
	return splitpane.Horizontal(primary, secondary)
}

// VerticalSplitPane creates a top/bottom SplitPane.
func VerticalSplitPane(primary, secondary rtui.VNode) rtui.VNode {
	return splitpane.Vertical(primary, secondary)
}
