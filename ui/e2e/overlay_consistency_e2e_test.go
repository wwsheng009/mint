package e2e

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/ui"
	absolutecomp "github.com/wwsheng009/mint/ui/components/absolute"
	popconfirmcomp "github.com/wwsheng009/mint/ui/components/popconfirm"
	popovercomp "github.com/wwsheng009/mint/ui/components/popover"
	statusbarcomp "github.com/wwsheng009/mint/ui/components/statusbar"
)

func newOverlayAutoPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Auto Placement Fixture").Build(),
				ui.NewTooltipBuilder(
					ui.NewButtonBuilder("Hover auto tooltip").
						SetID("overlay-auto-tooltip-anchor").
						Build(),
					"Auto Tooltip",
				).
					Auto().
					Delay(0).
					Build(),
				ui.NewPopoverBuilder(
					ui.NewButtonBuilder("Open auto popover").
						SetID("overlay-auto-popover-anchor").
						OnPress(popovercomp.ToggleWithID("fixture.overlay.auto.popover")).
						Build(),
				).
					SetID("overlay-auto-popover").
					ComponentID("fixture.overlay.auto.popover").
					Title("Auto Popover").
					Body("Auto popover body").
					Placement(ui.PopoverPlacementAuto).
					Trigger(ui.PopoverTriggerClick).
					Build(),
				ui.NewPopconfirmBuilder(
					ui.NewButtonBuilder("Open auto popconfirm").
						SetID("overlay-auto-popconfirm-anchor").
						Build(),
				).
					SetID("overlay-auto-popconfirm").
					ComponentID("fixture.overlay.auto.popconfirm").
					Title("Auto Popconfirm").
					Description("Auto popconfirm body").
					Placement(ui.PopconfirmPlacementAuto).
					Build(),
				ui.NewTextBuilder("Neutral row").Build(),
				ui.NewTextBuilder("Tail row").Build(),
			})
	}
}

func newOverlayTopPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Top Placement Fixture").Build(),
				ui.NewTooltipBuilder(
					ui.NewButtonBuilder("Hover top tooltip").
						SetID("overlay-top-tooltip-anchor").
						Build(),
					"Top Tooltip",
				).
					Top().
					Delay(0).
					Build(),
				ui.NewPopoverBuilder(
					ui.NewButtonBuilder("Open top popover").
						SetID("overlay-top-popover-anchor").
						OnPress(popovercomp.ToggleWithID("fixture.overlay.top.popover")).
						Build(),
				).
					SetID("overlay-top-popover").
					ComponentID("fixture.overlay.top.popover").
					Title("Top Popover").
					Body("Top popover body that should fall below").
					MaxWidth(24).
					Placement(ui.PopoverPlacementTop).
					Trigger(ui.PopoverTriggerClick).
					Build(),
				ui.NewPopconfirmBuilder(
					ui.NewButtonBuilder("Open top popconfirm").
						SetID("overlay-top-popconfirm-anchor").
						Build(),
				).
					SetID("overlay-top-popconfirm").
					ComponentID("fixture.overlay.top.popconfirm").
					Title("Top Popconfirm").
					Description("Top popconfirm body that should fall below").
					MaxWidth(24).
					Placement(ui.PopconfirmPlacementTop).
					Build(),
				ui.NewTextBuilder("Neutral row").Build(),
				ui.NewTextBuilder("Tail row").Build(),
			})
	}
}

func newOverlayRightEdgeTopPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Right Edge Top Placement Fixture").Build(),
				ui.NewTextBuilder("Workspace: gamma").Build(),
				ui.NewTextBuilder("Branch: overlay").Build(),
				ui.NewTextBuilder("Runtime: stable").Build(),
				ui.NewTextBuilder("Queue: healthy").Build(),
				ui.NewTextBuilder("Focus: anchored").Build(),
				ui.NewTextBuilder("Viewport: expanded").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Top Tip").
									SetID("overlay-right-top-tooltip-anchor").
									Build(),
								"Right Edge Top Tooltip",
							).
								Top().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("Open Top").
									SetID("overlay-right-top-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.right.popover")).
									Build(),
							).
								SetID("overlay-right-top-popover").
								ComponentID("fixture.overlay.right.popover").
								Title("Right Edge Top Popover").
								Body("Top-family fallback should stay above and shift left near the viewport edge.").
								MaxWidth(28).
								Placement(ui.PopoverPlacementTop).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(16).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("Confirm").
									SetID("overlay-right-top-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-right-top-popconfirm").
								ComponentID("fixture.overlay.right.popconfirm").
								Title("Right Edge Top Popconfirm").
								Description("Top-family fallback should stay above and shift left near the viewport edge.").
								MaxWidth(28).
								Placement(ui.PopconfirmPlacementTop).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Details: right-edge top-family fallback").Build(),
				statusbarcomp.NewBuilder().
					DefaultTheme().
					HelpFallback("Ready").
					HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
					TooltipPlacement(statusbarcomp.TooltipPlacementTop).
					TooltipMaxWidth(28).
					Left(statusbarcomp.Text("Mode: stable").WithWidth(16)).
					Center(statusbarcomp.Text("Overlay: right edge").WithWidth(24).WithAlign(ui.AlignCenter)).
					Right(statusbarcomp.Text("Help").WithKey("overlay-right-top-statusbar-anchor").WithHelp("Right Edge Statusbar Help").WithWidth(16).WithAlign(ui.AlignEnd)).
					BuildWithHelp(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-right-neutral-target").Build(),
			})
	}
}

func newOverlayLeftEdgeTopPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Left Edge Top Placement Fixture").Build(),
				ui.NewTextBuilder("Workspace: delta").Build(),
				ui.NewTextBuilder("Branch: overlay").Build(),
				ui.NewTextBuilder("Runtime: stable").Build(),
				ui.NewTextBuilder("Queue: healthy").Build(),
				ui.NewTextBuilder("Focus: anchored").Build(),
				ui.NewTextBuilder("Viewport: expanded").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Tip").
									SetID("overlay-left-top-tooltip-anchor").
									Build(),
								"Left Edge Top Tooltip",
							).
								Top().
								Delay(0).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(10).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("Open").
									SetID("overlay-left-top-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.left.popover")).
									Build(),
							).
								SetID("overlay-left-top-popover").
								ComponentID("fixture.overlay.left.popover").
								Title("Left Edge Top Popover").
								Body("Top-family fallback should stay above and shift right.").
								MaxWidth(28).
								Placement(ui.PopoverPlacementTop).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(12).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("Ask").
									SetID("overlay-left-top-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-left-top-popconfirm").
								ComponentID("fixture.overlay.left.popconfirm").
								Title("Left Edge Top Popconfirm").
								Description("Top-family fallback should stay above and shift right.").
								MaxWidth(28).
								Placement(ui.PopconfirmPlacementTop).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(10).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Details: left-edge top-family fallback").Build(),
				statusbarcomp.NewBuilder().
					DefaultTheme().
					HelpFallback("Ready").
					HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
					TooltipPlacement(statusbarcomp.TooltipPlacementTop).
					TooltipMaxWidth(28).
					Left(statusbarcomp.Text("Help").WithKey("overlay-left-top-statusbar-anchor").WithHelp("Left Edge Statusbar Help")).
					Center(statusbarcomp.Text("Overlay: left edge").WithWidth(24).WithAlign(ui.AlignCenter)).
					Right(statusbarcomp.Text("Mode: stable").WithWidth(16).WithAlign(ui.AlignEnd)).
					BuildWithHelp(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-left-neutral-target").Build(),
			})
	}
}

func newOverlayRightEdgeBottomPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Right Edge Bottom Placement Fixture").Build(),
				ui.NewTextBuilder("Workspace: epsilon").Build(),
				ui.NewTextBuilder("Branch: overlay").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Bottom Tip").
									SetID("overlay-right-bottom-tooltip-anchor").
									Build(),
								"Right Edge Bottom Tooltip",
							).
								Bottom().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(16).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("Open Bottom").
									SetID("overlay-right-bottom-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.right.bottom.popover")).
									Build(),
							).
								SetID("overlay-right-bottom-popover").
								ComponentID("fixture.overlay.right.bottom.popover").
								Title("Right Edge Bottom Popover").
								Body("Bottom-family fallback should stay below and shift left near the viewport edge.").
								MaxWidth(28).
								Placement(ui.PopoverPlacementBottom).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(16).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("Confirm Bottom").
									SetID("overlay-right-bottom-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-right-bottom-popconfirm").
								ComponentID("fixture.overlay.right.bottom.popconfirm").
								Title("Right Edge Bottom Popconfirm").
								Description("Bottom-family fallback should stay below and shift left near the viewport edge.").
								MaxWidth(28).
								Placement(ui.PopconfirmPlacementBottom).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(20).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Details: right-edge bottom-family fallback").Build(),
				statusbarcomp.NewBuilder().
					DefaultTheme().
					HelpFallback("Ready").
					HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
					TooltipPlacement(statusbarcomp.TooltipPlacementBottom).
					TooltipMaxWidth(28).
					Left(statusbarcomp.Text("Mode: stable").WithWidth(16)).
					Center(statusbarcomp.Text("Overlay: right edge bottom").WithWidth(28).WithAlign(ui.AlignCenter)).
					Right(statusbarcomp.Text("Help").WithKey("overlay-right-bottom-statusbar-anchor").WithHelp("Right Bottom Help").WithWidth(16).WithAlign(ui.AlignEnd)).
					BuildWithHelp(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-right-bottom-neutral-target").Build(),
			})
	}
}

func newOverlayLeftEdgeBottomPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Left Edge Bottom Placement Fixture").Build(),
				ui.NewTextBuilder("Workspace: zeta").Build(),
				ui.NewTextBuilder("Branch: overlay").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("Bottom Tip").
									SetID("overlay-left-bottom-tooltip-anchor").
									Build(),
								"Left Edge Bottom Tooltip",
							).
								Bottom().
								Delay(0).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("Open Bottom").
									SetID("overlay-left-bottom-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.left.bottom.popover")).
									Build(),
							).
								SetID("overlay-left-bottom-popover").
								ComponentID("fixture.overlay.left.bottom.popover").
								Title("Left Edge Bottom Popover").
								Body("Bottom-family fallback should stay below and shift right.").
								MaxWidth(28).
								Placement(ui.PopoverPlacementBottom).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("Ask Bottom").
									SetID("overlay-left-bottom-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-left-bottom-popconfirm").
								ComponentID("fixture.overlay.left.bottom.popconfirm").
								Title("Left Edge Bottom Popconfirm").
								Description("Bottom-family fallback should stay below and shift right.").
								MaxWidth(28).
								Placement(ui.PopconfirmPlacementBottom).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewTextBuilder("Details: left-edge bottom-family fallback").Build(),
				statusbarcomp.NewBuilder().
					DefaultTheme().
					HelpFallback("Ready").
					HelpDisplayMode(statusbarcomp.HelpDisplayOverlay).
					TooltipPlacement(statusbarcomp.TooltipPlacementBottom).
					TooltipMaxWidth(28).
					Left(statusbarcomp.Text("Help").WithKey("overlay-left-bottom-statusbar-anchor").WithHelp("Left Bottom Help")).
					Center(statusbarcomp.Text("Overlay: left edge bottom").WithWidth(28).WithAlign(ui.AlignCenter)).
					Right(statusbarcomp.Text("Mode: stable").WithWidth(16).WithAlign(ui.AlignEnd)).
					BuildWithHelp(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-left-bottom-neutral-target").Build(),
			})
	}
}

func newOverlayTopRightCornerPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Top Right Corner Fixture").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("TR Tip").
									SetID("overlay-top-right-corner-tooltip-anchor").
									Build(),
								"Top Right Corner Tooltip",
							).
								TopRight().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(10).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("TR Popover").
									SetID("overlay-top-right-corner-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.corner.topright.popover")).
									Build(),
							).
								SetID("overlay-top-right-corner-popover").
								ComponentID("fixture.overlay.corner.topright.popover").
								Title("TR Corner Popover").
								Body("TR popover below right family.").
								MaxWidth(24).
								Placement(ui.PopoverPlacementTopRight).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-top-right-corner-neutral-target").Build(),
			})
	}
}

func newOverlayTopLeftCornerPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Top Left Corner Fixture").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("TL Tip").
									SetID("overlay-top-left-corner-tooltip-anchor").
									Build(),
								"Top Left Corner Tooltip",
							).
								TopLeft().
								Delay(0).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(10).
							Height(1).
							Build(),
					}),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("TL Popover").
									SetID("overlay-top-left-corner-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.corner.topleft.popover")).
									Build(),
							).
								SetID("overlay-top-left-corner-popover").
								ComponentID("fixture.overlay.corner.topleft.popover").
								Title("TL Corner Popover").
								Body("TL popover below left family.").
								MaxWidth(24).
								Placement(ui.PopoverPlacementTopLeft).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-top-left-corner-neutral-target").Build(),
			})
	}
}

func newOverlayTopRightCornerPopconfirmFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Top Right Corner Popconfirm Fixture").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("TR Confirm").
									SetID("overlay-top-right-corner-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-top-right-corner-popconfirm").
								ComponentID("fixture.overlay.corner.topright.popconfirm").
								Title("TR Corner Popconfirm").
								Description("TR popconfirm below right family.").
								MaxWidth(24).
								Placement(ui.PopconfirmPlacementTopRight).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-top-right-corner-popconfirm-neutral-target").Build(),
			})
	}
}

func newOverlayTopLeftCornerPopconfirmFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Top Left Corner Popconfirm Fixture").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(1).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("TL Confirm").
									SetID("overlay-top-left-corner-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-top-left-corner-popconfirm").
								ComponentID("fixture.overlay.corner.topleft.popconfirm").
								Title("TL Corner Popconfirm").
								Description("TL popconfirm below left family.").
								MaxWidth(24).
								Placement(ui.PopconfirmPlacementTopLeft).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Top(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-top-left-corner-popconfirm-neutral-target").Build(),
			})
	}
}

func newOverlayBottomRightCornerPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Bottom Right Corner Fixture").Build(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-bottom-right-corner-neutral-target").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(22).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("BR Tip").
									SetID("overlay-bottom-right-corner-tooltip-anchor").
									Build(),
								"Bottom Right Corner Tooltip",
							).
								BottomRight().
								Delay(0).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Bottom(absolutecomp.AbsolutePos(1)).
							Width(10).
							Height(1).
							Build(),
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("BR Popover").
									SetID("overlay-bottom-right-corner-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.corner.bottomright.popover")).
									Build(),
							).
								SetID("overlay-bottom-right-corner-popover").
								ComponentID("fixture.overlay.corner.bottomright.popover").
								Title("BR Corner Popover").
								Body("BR popover above right").
								MaxWidth(24).
								Placement(ui.PopoverPlacementBottomRight).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Bottom(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
			})
	}
}

func newOverlayBottomLeftCornerPlacementFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Bottom Left Corner Fixture").Build(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-bottom-left-corner-neutral-target").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(22).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewTooltipBuilder(
								ui.NewButtonBuilder("BL Tip").
									SetID("overlay-bottom-left-corner-tooltip-anchor").
									Build(),
								"Bottom Left Corner Tooltip",
							).
								BottomLeft().
								Delay(0).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Bottom(absolutecomp.AbsolutePos(1)).
							Width(10).
							Height(1).
							Build(),
						absolutecomp.NewBuilder(
							ui.NewPopoverBuilder(
								ui.NewButtonBuilder("BL Popover").
									SetID("overlay-bottom-left-corner-popover-anchor").
									OnPress(popovercomp.ToggleWithID("fixture.overlay.corner.bottomleft.popover")).
									Build(),
							).
								SetID("overlay-bottom-left-corner-popover").
								ComponentID("fixture.overlay.corner.bottomleft.popover").
								Title("BL Corner Popover").
								Body("BL popover above left").
								MaxWidth(24).
								Placement(ui.PopoverPlacementBottomLeft).
								Trigger(ui.PopoverTriggerClick).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Bottom(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
			})
	}
}

func newOverlayBottomRightCornerPopconfirmFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Bottom Right Corner Popconfirm Fixture").Build(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-bottom-right-corner-popconfirm-neutral-target").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(22).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("BR Confirm").
									SetID("overlay-bottom-right-corner-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-bottom-right-corner-popconfirm").
								ComponentID("fixture.overlay.corner.bottomright.popconfirm").
								Title("BR Corner Popconfirm").
								Description("BR popconfirm above").
								MaxWidth(24).
								Placement(ui.PopconfirmPlacementBottomRight).
								Build(),
						).
							Right(absolutecomp.AbsolutePos(0)).
							Bottom(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
			})
	}
}

func newOverlayBottomLeftCornerPopconfirmFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Overlay Bottom Left Corner Popconfirm Fixture").Build(),
				ui.NewButtonBuilder("Neutral overlay target").SetID("overlay-bottom-left-corner-popconfirm-neutral-target").Build(),
				ui.NewVStack().
					SetWidth(72).
					SetHeight(22).
					SetChildrenList([]ui.VNode{
						absolutecomp.NewBuilder(
							ui.NewPopconfirmBuilder(
								ui.NewButtonBuilder("BL Confirm").
									SetID("overlay-bottom-left-corner-popconfirm-anchor").
									Build(),
							).
								SetID("overlay-bottom-left-corner-popconfirm").
								ComponentID("fixture.overlay.corner.bottomleft.popconfirm").
								Title("BL Corner Popconfirm").
								Description("BL popconfirm above").
								MaxWidth(24).
								Placement(ui.PopconfirmPlacementBottomLeft).
								Build(),
						).
							Left(absolutecomp.AbsolutePos(2)).
							Bottom(absolutecomp.AbsolutePos(0)).
							Width(14).
							Height(1).
							Build(),
					}),
			})
	}
}

func assertOverlayAboveAndShiftedLeft(t *testing.T, name string, overlayX, overlayY, anchorX, anchorY, anchorWidth, viewportWidth int) {
	t.Helper()
	if overlayY >= anchorY {
		t.Fatalf("%s row = %d, want above anchor row %d near right edge", name, overlayY, anchorY)
	}
	anchorMidX := anchorX + anchorWidth/2
	if overlayX >= anchorMidX {
		t.Fatalf("%s column = %d, want left of anchor midpoint column %d near right edge", name, overlayX, anchorMidX)
	}
	if overlayX < 0 || overlayX >= viewportWidth {
		t.Fatalf("%s column = %d, want within viewport width %d", name, overlayX, viewportWidth)
	}
}

func assertOverlayAboveAndShiftedRight(t *testing.T, name string, overlayX, overlayY, anchorX, anchorY, viewportWidth int) {
	t.Helper()
	if overlayY >= anchorY {
		t.Fatalf("%s row = %d, want above anchor row %d near left edge", name, overlayY, anchorY)
	}
	if overlayX <= anchorX {
		t.Fatalf("%s column = %d, want right of anchor column %d near left edge", name, overlayX, anchorX)
	}
	if overlayX < 0 || overlayX >= viewportWidth {
		t.Fatalf("%s column = %d, want within viewport width %d", name, overlayX, viewportWidth)
	}
}

func assertOverlayBelowAndShiftedLeft(t *testing.T, name string, overlayX, overlayY, anchorX, anchorY, anchorWidth, viewportWidth int) {
	t.Helper()
	if overlayY <= anchorY {
		t.Fatalf("%s row = %d, want below anchor row %d near right edge", name, overlayY, anchorY)
	}
	anchorMidX := anchorX + anchorWidth/2
	if overlayX >= anchorMidX {
		t.Fatalf("%s column = %d, want left of anchor midpoint column %d near right edge", name, overlayX, anchorMidX)
	}
	if overlayX < 0 || overlayX >= viewportWidth {
		t.Fatalf("%s column = %d, want within viewport width %d", name, overlayX, viewportWidth)
	}
}

func assertOverlayBelowAndShiftedRight(t *testing.T, name string, overlayX, overlayY, anchorX, anchorY, viewportWidth int) {
	t.Helper()
	if overlayY <= anchorY {
		t.Fatalf("%s row = %d, want below anchor row %d near left edge", name, overlayY, anchorY)
	}
	if overlayX <= anchorX {
		t.Fatalf("%s column = %d, want right of anchor column %d near left edge", name, overlayX, anchorX)
	}
	if overlayX < 0 || overlayX >= viewportWidth {
		t.Fatalf("%s column = %d, want within viewport width %d", name, overlayX, viewportWidth)
	}
}

func TestE2EOverlayAutoPlacementConsistentNearTopEdge(t *testing.T) {
	app, err := Run(
		newOverlayAutoPlacementFixture(),
		ui.WithSize(96, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popoverAnchor, err := app.BoundsOf(ByID("overlay-auto-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-auto-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	tooltipAnchor, err := app.BoundsOf(ByID("overlay-auto-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-auto-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Auto Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Auto Popover"))
	if err != nil {
		t.Fatal(err)
	}
	if popoverPoint.Y <= popoverAnchor.Y {
		t.Fatalf("auto popover row = %d, want below anchor row %d near top edge", popoverPoint.Y, popoverAnchor.Y)
	}

	if err := app.Driver().Click(ByID("overlay-auto-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Auto Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Auto Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	if popconfirmPoint.Y <= popconfirmAnchor.Y {
		t.Fatalf("auto popconfirm row = %d, want below anchor row %d near top edge", popconfirmPoint.Y, popconfirmAnchor.Y)
	}

	if err := app.Driver().Move(ByID("overlay-auto-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Auto Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Auto Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	if tooltipPoint.Y <= tooltipAnchor.Y {
		t.Fatalf("auto tooltip row = %d, want below anchor row %d near top edge", tooltipPoint.Y, tooltipAnchor.Y)
	}
}

func TestE2EOverlayExplicitTopPlacementFallsBackBelowTopEdge(t *testing.T) {
	app, err := Run(
		newOverlayTopPlacementFixture(),
		ui.WithSize(72, 14),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popoverAnchor, err := app.BoundsOf(ByID("overlay-top-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-top-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	tooltipAnchor, err := app.BoundsOf(ByID("overlay-top-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-top-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Top Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Top Popover"))
	if err != nil {
		t.Fatal(err)
	}
	if popoverPoint.Y <= popoverAnchor.Y {
		t.Fatalf("top popover row = %d, want below anchor row %d after fallback", popoverPoint.Y, popoverAnchor.Y)
	}

	if err := app.Driver().Click(ByID("overlay-top-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Top Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Top Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	if popconfirmPoint.Y <= popconfirmAnchor.Y {
		t.Fatalf("top popconfirm row = %d, want below anchor row %d after fallback", popconfirmPoint.Y, popconfirmAnchor.Y)
	}

	if err := app.Driver().Move(ByID("overlay-top-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Top Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Top Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	if tooltipPoint.Y <= tooltipAnchor.Y {
		t.Fatalf("top tooltip row = %d, want below anchor row %d after fallback", tooltipPoint.Y, tooltipAnchor.Y)
	}
}

func TestE2EOverlayExplicitTopPlacementStaysAboveAndShiftsLeftNearRightEdge(t *testing.T) {
	app, err := Run(
		newOverlayRightEdgeTopPlacementFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-right-top-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-right-top-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-right-top-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	statusbarAnchor, err := app.BoundsOf(ByKey("overlay-right-top-statusbar-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-right-top-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Top Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Right Edge Top Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "right-edge top tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, tooltipAnchor.Width, 72)

	if err := app.Driver().Click(ByID("overlay-right-top-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Top Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Right Edge Top Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "right-edge top popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, popoverAnchor.Width, 72)

	if err := app.Driver().Click(ByID("overlay-right-top-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Top Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Right Edge Top Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "right-edge top popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, popconfirmAnchor.Width, 72)

	if err := app.Driver().Move(ByKey("overlay-right-top-statusbar-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Statusbar Help"))
	}); err != nil {
		t.Fatal(err)
	}

	statusbarPoint, err := app.ResolvePoint(ByText("Right Edge Statusbar Help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "right-edge top statusbar help", statusbarPoint.X, statusbarPoint.Y, statusbarAnchor.X, statusbarAnchor.Y, statusbarAnchor.Width, 72)
}

func TestE2EOverlayExplicitTopPlacementStaysAboveAndShiftsRightNearLeftEdge(t *testing.T) {
	app, err := Run(
		newOverlayLeftEdgeTopPlacementFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-left-top-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-left-top-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-left-top-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	statusbarAnchor, err := app.BoundsOf(ByKey("overlay-left-top-statusbar-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-left-top-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Top Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Left Edge Top Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "left-edge top tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, 72)

	if err := app.Driver().Click(ByID("overlay-left-top-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Top Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Left Edge Top Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "left-edge top popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, 72)

	if err := app.Driver().Click(ByID("overlay-left-top-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Top Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Left Edge Top Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "left-edge top popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, 72)

	if err := app.Driver().Move(ByKey("overlay-left-top-statusbar-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Statusbar Help"))
	}); err != nil {
		t.Fatal(err)
	}

	statusbarPoint, err := app.ResolvePoint(ByText("Left Edge Statusbar Help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "left-edge top statusbar help", statusbarPoint.X, statusbarPoint.Y, statusbarAnchor.X, statusbarAnchor.Y, 72)
}

func TestE2EOverlayExplicitBottomPlacementStaysBelowAndShiftsLeftNearRightEdge(t *testing.T) {
	app, err := Run(
		newOverlayRightEdgeBottomPlacementFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-right-bottom-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-right-bottom-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-right-bottom-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	statusbarAnchor, err := app.BoundsOf(ByKey("overlay-right-bottom-statusbar-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-right-bottom-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Bottom Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Right Edge Bottom Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "right-edge bottom tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, tooltipAnchor.Width, 72)

	if err := app.Driver().Click(ByID("overlay-right-bottom-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Bottom Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Right Edge Bottom Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "right-edge bottom popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, popoverAnchor.Width, 72)

	if err := app.Driver().Move(ByKey("overlay-right-bottom-statusbar-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Bottom Help"))
	}); err != nil {
		t.Fatal(err)
	}

	statusbarPoint, err := app.ResolvePoint(ByText("Right Bottom Help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "right-edge bottom statusbar help", statusbarPoint.X, statusbarPoint.Y, statusbarAnchor.X, statusbarAnchor.Y, statusbarAnchor.Width, 72)

	if err := app.Driver().Click(ByID("overlay-right-bottom-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Right Edge Bottom Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Right Edge Bottom Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "right-edge bottom popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, popconfirmAnchor.Width, 72)
}

func TestE2EOverlayExplicitBottomPlacementStaysBelowAndShiftsRightNearLeftEdge(t *testing.T) {
	app, err := Run(
		newOverlayLeftEdgeBottomPlacementFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-left-bottom-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-left-bottom-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-left-bottom-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	statusbarAnchor, err := app.BoundsOf(ByKey("overlay-left-bottom-statusbar-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-left-bottom-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Bottom Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Left Edge Bottom Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "left-edge bottom tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, 72)

	if err := app.Driver().Click(ByID("overlay-left-bottom-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Bottom Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("Left Edge Bottom Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "left-edge bottom popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, 72)

	if err := app.Driver().Move(ByKey("overlay-left-bottom-statusbar-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Bottom Help"))
	}); err != nil {
		t.Fatal(err)
	}

	statusbarPoint, err := app.ResolvePoint(ByText("Left Bottom Help"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "left-edge bottom statusbar help", statusbarPoint.X, statusbarPoint.Y, statusbarAnchor.X, statusbarAnchor.Y, 72)

	if err := app.Driver().Click(ByID("overlay-left-bottom-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Left Edge Bottom Popconfirm"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("Left Edge Bottom Popconfirm"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "left-edge bottom popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, 72)
}

func TestE2EOverlayExplicitTopRightCornerFallsBelowWithinRightFamily(t *testing.T) {
	app, err := Run(
		newOverlayTopRightCornerPlacementFixture(),
		ui.WithSize(72, 14),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-top-right-corner-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-top-right-corner-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-top-right-corner-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Top Right Corner Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Top Right Corner Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "top-right corner tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, tooltipAnchor.Width, 72)

	if err := app.Driver().Move(ByID("overlay-top-right-corner-neutral-target")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-top-right-corner-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TR popover below right"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("TR popover below right"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "top-right corner popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, popoverAnchor.Width, 72)
}

func TestE2EOverlayExplicitTopLeftCornerFallsBelowWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newOverlayTopLeftCornerPlacementFixture(),
		ui.WithSize(72, 14),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-top-left-corner-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-top-left-corner-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-top-left-corner-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Top Left Corner Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Top Left Corner Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "top-left corner tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, 72)

	if err := app.Driver().Move(ByID("overlay-top-left-corner-neutral-target")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-top-left-corner-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TL popover below left"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("TL popover below left"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "top-left corner popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, 72)
}

func TestE2EOverlayExplicitTopRightCornerPopconfirmFallsBelowWithinRightFamily(t *testing.T) {
	app, err := Run(
		newOverlayTopRightCornerPopconfirmFixture(),
		ui.WithSize(72, 14),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-top-right-corner-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-top-right-corner-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TR popconfirm below"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("TR popconfirm below"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedLeft(t, "top-right corner popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, popconfirmAnchor.Width, 72)
}

func TestE2EOverlayExplicitTopLeftCornerPopconfirmFallsBelowWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newOverlayTopLeftCornerPopconfirmFixture(),
		ui.WithSize(72, 14),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-top-left-corner-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-top-left-corner-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("TL popconfirm below"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("TL popconfirm below"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayBelowAndShiftedRight(t, "top-left corner popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, 72)
}

func TestE2EOverlayExplicitBottomRightCornerFallsAboveWithinRightFamily(t *testing.T) {
	app, err := Run(
		newOverlayBottomRightCornerPlacementFixture(),
		ui.WithSize(72, 24),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-bottom-right-corner-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-bottom-right-corner-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-bottom-right-corner-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Bottom Right Corner Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Bottom Right Corner Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "bottom-right corner tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, tooltipAnchor.Width, 72)

	if err := app.Driver().Move(ByID("overlay-bottom-right-corner-neutral-target")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-bottom-right-corner-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("BR Corner Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("BR Corner Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "bottom-right corner popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, popoverAnchor.Width, 72)
}

func TestE2EOverlayExplicitBottomLeftCornerFallsAboveWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newOverlayBottomLeftCornerPlacementFixture(),
		ui.WithSize(72, 24),
		ui.WithPluginSetup(func(app *framework.App) {
			popovercomp.Install(app)
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	tooltipAnchor, err := app.BoundsOf(ByID("overlay-bottom-left-corner-tooltip-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popoverAnchor, err := app.BoundsOf(ByID("overlay-bottom-left-corner-popover-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Move(ByID("overlay-bottom-left-corner-tooltip-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Bottom Left Corner Tooltip"))
	}); err != nil {
		t.Fatal(err)
	}

	tooltipPoint, err := app.ResolvePoint(ByText("Bottom Left Corner Tooltip"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "bottom-left corner tooltip", tooltipPoint.X, tooltipPoint.Y, tooltipAnchor.X, tooltipAnchor.Y, 72)

	if err := app.Driver().Move(ByID("overlay-bottom-left-corner-neutral-target")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-bottom-left-corner-popover-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("BL Corner Popover"))
	}); err != nil {
		t.Fatal(err)
	}

	popoverPoint, err := app.ResolvePoint(ByText("BL Corner Popover"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "bottom-left corner popover", popoverPoint.X, popoverPoint.Y, popoverAnchor.X, popoverAnchor.Y, 72)
}

func TestE2EOverlayExplicitBottomRightCornerPopconfirmFallsAboveWithinRightFamily(t *testing.T) {
	app, err := Run(
		newOverlayBottomRightCornerPopconfirmFixture(),
		ui.WithSize(72, 24),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-bottom-right-corner-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-bottom-right-corner-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("BR popconfirm above"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("BR popconfirm above"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedLeft(t, "bottom-right corner popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, popconfirmAnchor.Width, 72)
}

func TestE2EOverlayExplicitBottomLeftCornerPopconfirmFallsAboveWithinLeftFamily(t *testing.T) {
	app, err := Run(
		newOverlayBottomLeftCornerPopconfirmFixture(),
		ui.WithSize(72, 24),
		ui.WithPluginSetup(func(app *framework.App) {
			popconfirmcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	popconfirmAnchor, err := app.BoundsOf(ByID("overlay-bottom-left-corner-popconfirm-anchor"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("overlay-bottom-left-corner-popconfirm-anchor")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("BL popconfirm above"))
	}); err != nil {
		t.Fatal(err)
	}

	popconfirmPoint, err := app.ResolvePoint(ByText("BL popconfirm above"))
	if err != nil {
		t.Fatal(err)
	}
	assertOverlayAboveAndShiftedRight(t, "bottom-left corner popconfirm", popconfirmPoint.X, popconfirmPoint.Y, popconfirmAnchor.X, popconfirmAnchor.Y, 72)
}
