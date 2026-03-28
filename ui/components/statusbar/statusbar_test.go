package statusbar

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type testIntent struct{}

func (testIntent) IntentType() string { return "statusbar.test" }

func TestBuilderBuildsThreeSlots(t *testing.T) {
	bar := NewBuilder().
		Left(Text("L")).
		Center(Text("C")).
		Right(Text("R")).
		Padding(0, 1, 0, 1).
		Build()

	if bar == nil {
		t.Fatal("Build() returned nil")
	}
	if len(bar.Children()) != 3 {
		t.Fatalf("children len = %d, want 3", len(bar.Children()))
	}
	padding, ok := bar.Props()["padding"].([4]int)
	if !ok {
		t.Fatal("padding prop missing or wrong type")
	}
	if padding != [4]int{0, 1, 0, 1} {
		t.Fatalf("padding = %#v, want %#v", padding, [4]int{0, 1, 0, 1})
	}
}

func TestBuildWithHelpAddsHelpLine(t *testing.T) {
	bar := NewBuilder().
		DefaultTheme().
		HelpFallback("Ready").
		Left(ActionText("Open", testIntent{}).WithHelp("Open current item")).
		BuildWithHelp()

	if bar == nil {
		t.Fatal("BuildWithHelp() returned nil")
	}
	if len(bar.Children()) != 2 {
		t.Fatalf("children len = %d, want 2", len(bar.Children()))
	}
}

func TestBuildWithHelpOverlayReturnsTooltipLayerNode(t *testing.T) {
	bar := NewBuilder().
		DefaultTheme().
		HelpFallback("Ready").
		HelpDisplayMode(HelpDisplayOverlay).
		Left(ActionText("Open", testIntent{}).WithHelp("Open current item")).
		BuildWithHelp()

	if bar == nil {
		t.Fatal("BuildWithHelp() returned nil")
	}
	children := bar.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	if children[1].GetLayer() != rtui.LayerTooltip {
		t.Fatalf("overlay layer = %v, want %v", children[1].GetLayer(), rtui.LayerTooltip)
	}
}

func TestBuildWithHelpBothOffsetsOverlay(t *testing.T) {
	bar := NewBuilder().
		DefaultTheme().
		HelpFallback("Ready").
		HelpDisplayMode(HelpDisplayBoth).
		Left(ActionText("Open", testIntent{}).WithHelp("Open current item")).
		BuildWithHelp()

	children := bar.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	overlayFactory, ok := children[1].(rtui.InstanceFactory)
	if !ok {
		t.Fatal("overlay child is not an instance factory")
	}
	overlayInst, ok := overlayFactory.CreateInstance().(*overlayHelpInstance)
	if !ok {
		t.Fatal("overlay instance type mismatch")
	}
	if overlayInst.bottomOffsetRows != 1 {
		t.Fatalf("overlay bottom offset = %d, want 1", overlayInst.bottomOffsetRows)
	}
}

func TestOverlayTooltipPlacementConfig(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 30, 10}, maxContentWidth: 20}
	anchor := [4]int{12, 4, 4, 1}

	inst.placement = TooltipPlacementTop
	topBox := inst.computeTooltipBox("Tooltip content", anchor)
	if topBox.y >= anchor[1] {
		t.Fatalf("top placement y = %d, want < %d", topBox.y, anchor[1])
	}

	inst.placement = TooltipPlacementBottom
	bottomBox := inst.computeTooltipBox("Tooltip content", anchor)
	if bottomBox.y < anchor[1]+anchor[3] {
		t.Fatalf("bottom placement y = %d, want > %d", bottomBox.y, anchor[1])
	}
}

func TestPreferredTooltipPlacementKeepsBottomBiasWhenBothSidesFit(t *testing.T) {
	got := preferredTooltipPlacement([4]int{12, 4, 4, 1}, TooltipPlacementAuto, 3, 1, 0, 1, 12)
	if got != TooltipPlacementBottom {
		t.Fatalf("preferredTooltipPlacement() = %v, want %v", got, TooltipPlacementBottom)
	}
}

func TestOverlayTooltipBoxFlipsAndClamps(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 30, 8}, placement: TooltipPlacementAuto, maxContentWidth: 20}
	box := inst.computeTooltipBox("Tooltip content", [4]int{26, 6, 3, 1})
	if box.x < 0 || box.x+box.width+1 > 30 {
		t.Fatalf("tooltip x overflow: x=%d width=%d", box.x, box.width)
	}
	if box.y >= 6 {
		t.Fatalf("tooltip should flip above anchor, got y=%d", box.y)
	}
}

func TestOverlayTooltipWrapsMultilineContent(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 40, 12}, placement: TooltipPlacementBottom, maxContentWidth: 12}
	box := inst.computeTooltipBox("This tooltip wraps into multiple display lines", [4]int{10, 2, 4, 1})
	if len(box.lines) <= 3 {
		t.Fatalf("wrapped tooltip lines = %d, want > 3", len(box.lines))
	}
	if box.height != len(box.lines) {
		t.Fatalf("box height = %d, want %d", box.height, len(box.lines))
	}
}

func TestOverlayTooltipUsesGapRows(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 40, 12}, placement: TooltipPlacementBottom, maxContentWidth: 12, gapRows: 2}
	box := inst.computeTooltipBox("Tooltip", [4]int{10, 2, 4, 1})
	if box.y != 5 {
		t.Fatalf("bottom gap y = %d, want 5", box.y)
	}

	inst.placement = TooltipPlacementTop
	box = inst.computeTooltipBox("Tooltip", [4]int{10, 8, 4, 1})
	if box.y >= 8-2 {
		t.Fatalf("top gap y = %d, want clearly above anchor", box.y)
	}
}

func TestOverlayTooltipOnlyShowsHoveredHelp(t *testing.T) {
	model := newHelpModel("", "? ")
	model.Update("demo", 0, "Overlay tooltip content", false, true, [4]int{5, 2, 4, 1})
	inst := &overlayHelpInstance{
		model:       model,
		fillStyle:   style.NewStyle().Foreground(style.White).Background(style.Blue),
		borderStyle: style.NewStyle().Foreground(style.Yellow).Background(style.Blue).Bold(true),
		shadowStyle: style.NewStyle().Foreground(style.BrightBlack).Background(style.Blue),
		bounds:      [4]int{0, 0, 40, 10},
	}
	if cmds := inst.Paint(0, 0); len(cmds) != 0 {
		t.Fatalf("focused-only overlay cmds = %d, want 0", len(cmds))
	}

	model.Update("demo", 0, "Overlay tooltip content", true, true, [4]int{5, 2, 4, 1})
	if cmds := inst.Paint(0, 0); len(cmds) == 0 {
		t.Fatal("hovered overlay should paint commands")
	}
}

func TestOverlayTooltipAddsArrowBubbleBelowAnchor(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 40, 12}, placement: TooltipPlacementBottom, maxContentWidth: 16, gapRows: 1, arrowStyle: TooltipArrowStyleSharp}
	box := inst.computeTooltipBox("Tooltip", [4]int{10, 2, 4, 1})
	if !box.hasArrow {
		t.Fatal("expected overlay arrow to be enabled")
	}
	if box.y != 4 {
		t.Fatalf("box y = %d, want 4", box.y)
	}
	if got := string([]rune(box.lines[0])[box.arrowX-box.x]); got != "▲" {
		t.Fatalf("top border arrow = %q, want %q", got, "▲")
	}
}
func TestOverlayTooltipAddsArrowBubbleAboveAnchor(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 40, 12}, placement: TooltipPlacementTop, maxContentWidth: 16, gapRows: 1, arrowStyle: TooltipArrowStyleSharp}
	box := inst.computeTooltipBox("Tooltip", [4]int{10, 8, 4, 1})
	if !box.hasArrow {
		t.Fatal("expected overlay arrow to be enabled")
	}
	if got := string([]rune(box.lines[len(box.lines)-1])[box.arrowX-box.x]); got != "▼" {
		t.Fatalf("bottom border arrow = %q, want %q", got, "▼")
	}
}
func TestOverlayTooltipRoundedArrowThemeUsesRoundedCorners(t *testing.T) {
	inst := &overlayHelpInstance{bounds: [4]int{0, 0, 40, 12}, placement: TooltipPlacementBottom, maxContentWidth: 16, gapRows: 1, arrowStyle: TooltipArrowStyleRounded}
	box := inst.computeTooltipBox("Tooltip", [4]int{10, 2, 4, 1})
	if got := string([]rune(box.lines[0])[0]); got != "╭" {
		t.Fatalf("top-left corner = %q, want %q", got, "╭")
	}
	if got := string([]rune(box.lines[len(box.lines)-1])[len([]rune(box.lines[len(box.lines)-1]))-1]); got != "╯" {
		t.Fatalf("bottom-right corner = %q, want %q", got, "╯")
	}
	if got := string([]rune(box.lines[0])[box.arrowX-box.x]); got != "△" {
		t.Fatalf("rounded top border arrow = %q, want %q", got, "△")
	}
}
func TestResolveThemeDefaultsPreservesTooltipArrowStyle(t *testing.T) {
	theme := resolveThemeDefaults(Theme{})
	if theme.TooltipArrowStyle != TooltipArrowStyleSharp {
		t.Fatalf("default arrow style = %v, want %v", theme.TooltipArrowStyle, TooltipArrowStyleSharp)
	}
	muted := resolveThemeDefaults(MutedTheme())
	if muted.TooltipArrowStyle != TooltipArrowStyleRounded {
		t.Fatalf("muted arrow style = %v, want %v", muted.TooltipArrowStyle, TooltipArrowStyleRounded)
	}
}

func TestResolveTooltipPositionUsesPopoverAnchoring(t *testing.T) {
	left := resolveTooltipPosition([4]int{12, 2, 4, 1}, TooltipPlacementBottom, 10, 3, 1, 0, 1, 40, 12)
	if left.X != 12 {
		t.Fatalf("left anchored x = %d, want 12", left.X)
	}

	right := resolveTooltipPosition([4]int{34, 2, 4, 1}, TooltipPlacementBottom, 10, 3, 1, 0, 1, 40, 12)
	if right.X != 28 {
		t.Fatalf("right anchored x = %d, want 28", right.X)
	}
}

func TestResolveTooltipPositionShiftsLeftWithinTopFamilyNearRightEdge(t *testing.T) {
	result := resolveTooltipPosition([4]int{34, 8, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 40, 16)
	if result.X != 22 {
		t.Fatalf("right-edge top fallback x = %d, want 22", result.X)
	}
	if result.Y >= 8 {
		t.Fatalf("right-edge top fallback y = %d, want above anchor row 8", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("right-edge top fallback placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestResolveTooltipPositionShiftsRightWithinTopFamilyNearLeftEdge(t *testing.T) {
	result := resolveTooltipPosition([4]int{2, 8, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 40, 16)
	if result.X != 2 {
		t.Fatalf("left-edge top fallback x = %d, want 2", result.X)
	}
	if result.Y >= 8 {
		t.Fatalf("left-edge top fallback y = %d, want above anchor row 8", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("left-edge top fallback placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestResolveTooltipPositionShiftsLeftWithinBottomFamilyNearRightEdge(t *testing.T) {
	result := resolveTooltipPosition([4]int{34, 4, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 40, 16)
	if result.X != 22 {
		t.Fatalf("right-edge bottom fallback x = %d, want 22", result.X)
	}
	if result.Y <= 4 {
		t.Fatalf("right-edge bottom fallback y = %d, want below anchor row 4", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("right-edge bottom fallback placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionShiftsRightWithinBottomFamilyNearLeftEdge(t *testing.T) {
	result := resolveTooltipPosition([4]int{2, 4, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 40, 16)
	if result.X != 2 {
		t.Fatalf("left-edge bottom fallback x = %d, want 2", result.X)
	}
	if result.Y <= 4 {
		t.Fatalf("left-edge bottom fallback y = %d, want below anchor row 4", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("left-edge bottom fallback placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionTopFallsBelowAndShiftsLeftNearTopRightCorner(t *testing.T) {
	result := resolveTooltipPosition([4]int{34, 1, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 40, 10)
	if result.X != 22 {
		t.Fatalf("top-right corner fallback x = %d, want 22", result.X)
	}
	if result.Y != 3 {
		t.Fatalf("top-right corner fallback y = %d, want 3", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("top-right corner fallback placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionTopClampsLeftAndStaysAboveInNarrowViewport(t *testing.T) {
	result := resolveTooltipPosition([4]int{9, 7, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 14, 14)
	if result.X != 0 {
		t.Fatalf("narrow top clamp x = %d, want 0", result.X)
	}
	if result.Y != 2 {
		t.Fatalf("narrow top clamp y = %d, want 2", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("narrow top clamp placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestResolveTooltipPositionTopClampsBothAxesAndPreservesTopFamilyWhenNothingFits(t *testing.T) {
	result := resolveTooltipPosition([4]int{9, 7, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 14, 5)
	if result.X != 0 {
		t.Fatalf("dual-axis top clamp x = %d, want 0", result.X)
	}
	if result.Y != 0 {
		t.Fatalf("dual-axis top clamp y = %d, want 0", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("dual-axis top clamp placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestResolveTooltipPositionTopFallsBelowAndShiftsRightNearTopLeftCorner(t *testing.T) {
	result := resolveTooltipPosition([4]int{2, 1, 4, 1}, TooltipPlacementTop, 16, 4, 1, 0, 1, 40, 10)
	if result.X != 2 {
		t.Fatalf("top-left corner fallback x = %d, want 2", result.X)
	}
	if result.Y != 3 {
		t.Fatalf("top-left corner fallback y = %d, want 3", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("top-left corner fallback placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionBottomFallsAboveAndShiftsLeftNearBottomRightCorner(t *testing.T) {
	result := resolveTooltipPosition([4]int{34, 8, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 40, 10)
	if result.X != 22 {
		t.Fatalf("bottom-right corner fallback x = %d, want 22", result.X)
	}
	if result.Y != 3 {
		t.Fatalf("bottom-right corner fallback y = %d, want 3", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("bottom-right corner fallback placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestResolveTooltipPositionBottomClampsLeftAndStaysBelowInNarrowViewport(t *testing.T) {
	result := resolveTooltipPosition([4]int{9, 7, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 14, 14)
	if result.X != 0 {
		t.Fatalf("narrow bottom clamp x = %d, want 0", result.X)
	}
	if result.Y != 9 {
		t.Fatalf("narrow bottom clamp y = %d, want 9", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("narrow bottom clamp placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionBottomClampsBothAxesAndPreservesBottomFamilyWhenNothingFits(t *testing.T) {
	result := resolveTooltipPosition([4]int{9, 7, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 14, 5)
	if result.X != 0 {
		t.Fatalf("dual-axis bottom clamp x = %d, want 0", result.X)
	}
	if result.Y != 0 {
		t.Fatalf("dual-axis bottom clamp y = %d, want 0", result.Y)
	}
	if result.Placement != TooltipPlacementBottom {
		t.Fatalf("dual-axis bottom clamp placement = %v, want %v", result.Placement, TooltipPlacementBottom)
	}
}

func TestResolveTooltipPositionBottomFallsAboveAndShiftsRightNearBottomLeftCorner(t *testing.T) {
	result := resolveTooltipPosition([4]int{2, 8, 4, 1}, TooltipPlacementBottom, 16, 4, 1, 0, 1, 40, 10)
	if result.X != 2 {
		t.Fatalf("bottom-left corner fallback x = %d, want 2", result.X)
	}
	if result.Y != 3 {
		t.Fatalf("bottom-left corner fallback y = %d, want 3", result.Y)
	}
	if result.Placement != TooltipPlacementTop {
		t.Fatalf("bottom-left corner fallback placement = %v, want %v", result.Placement, TooltipPlacementTop)
	}
}

func TestOverlayTooltipUsesViewportSizeForCornerFallback(t *testing.T) {
	inst := &overlayHelpInstance{placement: TooltipPlacementTop, maxContentWidth: 16}
	inst.SetViewportSize(40, 10)

	box := inst.computeTooltipBox("Tooltip", [4]int{34, 1, 4, 1})
	if box.y <= 1 {
		t.Fatalf("viewport-driven top corner fallback y = %d, want below anchor row 1", box.y)
	}
	if box.y < 0 {
		t.Fatalf("viewport-driven top corner fallback y = %d, want visible row", box.y)
	}
}

func TestOverlayTooltipPaintsBoxAndShadow(t *testing.T) {
	model := newHelpModel("", "? ")
	model.Update("demo", 0, "Overlay tooltip content", true, false, [4]int{5, 2, 4, 1})
	inst := &overlayHelpInstance{
		model:       model,
		fillStyle:   style.NewStyle().Foreground(style.White).Background(style.Blue),
		borderStyle: style.NewStyle().Foreground(style.Yellow).Background(style.Blue).Bold(true),
		shadowStyle: style.NewStyle().Foreground(style.BrightBlack).Background(style.Blue),
		bounds:      [4]int{0, 0, 40, 10},
	}
	cmds := inst.Paint(0, 0)
	if len(cmds) < 7 {
		t.Fatalf("overlay paint command count = %d, want at least 7", len(cmds))
	}
	if cmds[0].Style.FG != style.Yellow {
		t.Fatalf("border FG = %q, want %q", cmds[0].Style.FG, style.Yellow)
	}
	last := cmds[len(cmds)-1]
	if last.Style.FG != style.BrightBlack {
		t.Fatalf("shadow FG = %q, want %q", last.Style.FG, style.BrightBlack)
	}
}

func TestSectionsCopiesInput(t *testing.T) {
	source := []Section{Text("A"), Text("B")}
	sections := Sections(source...)
	if len(sections) != len(source) {
		t.Fatalf("len = %d, want %d", len(sections), len(source))
	}
	sections[0].Text = "changed"
	if source[0].Text != "A" {
		t.Fatalf("source section mutated: %q", source[0].Text)
	}
}

func TestFitTextPadsAndTruncatesASCII(t *testing.T) {
	if got := fitText("abc", 5, rtui.AlignStart, OverflowEllipsis); got != "abc  " {
		t.Fatalf("start align = %q, want %q", got, "abc  ")
	}
	if got := fitText("abc", 5, rtui.AlignEnd, OverflowEllipsis); got != "  abc" {
		t.Fatalf("end align = %q, want %q", got, "  abc")
	}
	if got := fitText("abc", 5, rtui.AlignCenter, OverflowEllipsis); got != " abc " {
		t.Fatalf("center align = %q, want %q", got, " abc ")
	}
	if got := fitText("abcdef", 4, rtui.AlignStart, OverflowEllipsis); got != "abc…" {
		t.Fatalf("ellipsis truncate = %q, want %q", got, "abc…")
	}
	if got := fitText("abcdef", 4, rtui.AlignStart, OverflowClip); got != "abcd" {
		t.Fatalf("clip truncate = %q, want %q", got, "abcd")
	}
}

func TestFitTextUsesDisplayWidth(t *testing.T) {
	if got := fitText("你好", 6, rtui.AlignStart, OverflowClip); got != "你好  " {
		t.Fatalf("wide-char padding = %q, want %q", got, "你好  ")
	}
	if width := paint.StringWidth(fitText("你好", 6, rtui.AlignStart, OverflowClip)); width != 6 {
		t.Fatalf("wide-char padded width = %d, want 6", width)
	}
	if got := fitText("你好世界", 5, rtui.AlignStart, OverflowEllipsis); got != "你好…" {
		t.Fatalf("wide-char ellipsis = %q, want %q", got, "你好…")
	}
	if got := fitText("你好世界", 4, rtui.AlignStart, OverflowClip); got != "你好" {
		t.Fatalf("wide-char clip = %q, want %q", got, "你好")
	}
}

func TestFitTextNormalizesSingleLineContent(t *testing.T) {
	if got := fitText("A\nB\tC", 6, rtui.AlignStart, OverflowClip); got != "A B C " {
		t.Fatalf("single-line normalize = %q, want %q", got, "A B C ")
	}
}

func TestResolveSectionUsesThemeDefaults(t *testing.T) {
	resolved := resolveSection(Section{Text: "demo", Width: 8}, rtui.AlignEnd, ContrastTheme(), true)
	if resolved.FgColor != "black" {
		t.Fatalf("FgColor = %q, want black", resolved.FgColor)
	}
	if resolved.BgColor != "yellow" {
		t.Fatalf("BgColor = %q, want yellow", resolved.BgColor)
	}
	if !resolved.Bold {
		t.Fatal("Bold should inherit from theme")
	}
	if resolved.Text != "    demo" {
		t.Fatalf("Text = %q, want %q", resolved.Text, "    demo")
	}
}

func TestSectionVNodeCarriesStyleInProps(t *testing.T) {
	section := resolveSection(Section{Text: "demo"}, rtui.AlignStart, DefaultTheme(), true)
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	styled, ok := node.Props()["style"].(style.Style)
	if !ok {
		t.Fatal("style prop missing from section vnode")
	}
	if styled.FG != style.White {
		t.Fatalf("style FG = %q, want %q", styled.FG, style.White)
	}
	if styled.BG != style.Blue {
		t.Fatalf("style BG = %q, want %q", styled.BG, style.Blue)
	}
}

func TestResolveHelpStyleUsesThemeHelpStyle(t *testing.T) {
	helpStyle := style.NewStyle().Foreground(style.Black).Background(style.Cyan).Bold(true)
	builder := NewBuilder().Theme(DefaultTheme().WithHelpStyle(helpStyle))
	resolved := builder.resolveHelpStyle()
	if resolved.FG != style.Black {
		t.Fatalf("help FG = %q, want %q", resolved.FG, style.Black)
	}
	if resolved.BG != style.Cyan {
		t.Fatalf("help BG = %q, want %q", resolved.BG, style.Cyan)
	}
	if !resolved.IsBold() {
		t.Fatal("help style should stay bold")
	}
}

func TestInteractiveSectionUsesThemeStateStyles(t *testing.T) {
	theme := DefaultTheme().
		WithHoverStyle(style.NewStyle().Background(style.Cyan)).
		WithFocusStyle(style.NewStyle().Foreground(style.Yellow).Bold(true)).
		WithPressedStyle(style.NewStyle().Reverse(true)).
		WithDisabledStyle(style.NewStyle().Foreground(style.Red))
	section := resolveSection(ActionText("Demo", testIntent{}), rtui.AlignStart, theme, true)
	node := renderSection(section, rtui.AlignStart, theme, true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseEnter)) {
		t.Fatal("mouse enter should be handled")
	}
	hoverCmd := inst.Paint(0, 0)
	if len(hoverCmd) != 1 || hoverCmd[0].Style.BG != style.Cyan {
		t.Fatalf("hover style BG = %q, want %q", hoverCmd[0].Style.BG, style.Cyan)
	}
	inst.SetFocus(true)
	focusCmd := inst.Paint(0, 0)
	if focusCmd[0].Style.FG != style.Yellow || !focusCmd[0].Style.IsBold() {
		t.Fatal("focus style overlay was not applied")
	}
	if !inst.HandleAction(action.NewAction(action.ActionClick)) {
		t.Fatal("click should be handled")
	}
	pressedCmd := inst.Paint(0, 0)
	if !pressedCmd[0].Style.IsReverse() {
		t.Fatal("pressed style overlay was not applied")
	}
	inst.state.Disabled = true
	disabledCmd := inst.Paint(0, 0)
	if disabledCmd[0].Style.FG != style.Red {
		t.Fatalf("disabled style FG = %q, want %q", disabledCmd[0].Style.FG, style.Red)
	}
}

func TestResolveSectionWithExplicitOverrides(t *testing.T) {
	resolved := resolveSection(
		Text("demo").WithWidth(8).WithAlign(rtui.AlignStart).WithBold(false),
		rtui.AlignEnd,
		ContrastTheme(),
		true,
	)
	if resolved.Bold {
		t.Fatal("Bold should remain false when explicitly disabled")
	}
	if resolved.Text != "demo    " {
		t.Fatalf("Text = %q, want %q", resolved.Text, "demo    ")
	}
}

func TestResolveSectionRespectsOverflowMode(t *testing.T) {
	ellipsis := resolveSection(Text("abcdef").WithWidth(4), rtui.AlignStart, Theme{}, false)
	if ellipsis.Text != "abc…" {
		t.Fatalf("default overflow = %q, want %q", ellipsis.Text, "abc…")
	}
	clip := resolveSection(Text("abcdef").WithWidth(4).WithClip(), rtui.AlignStart, Theme{}, false)
	if clip.Text != "abcd" {
		t.Fatalf("clip overflow = %q, want %q", clip.Text, "abcd")
	}
}

func TestBuilderClampsNegativeSpacing(t *testing.T) {
	bar := NewBuilder().Gap(-2).Padding(-1, 2, -3, 4).Build()
	padding, ok := bar.Props()["padding"].([4]int)
	if !ok {
		t.Fatal("padding prop missing or wrong type")
	}
	if padding != [4]int{0, 2, 0, 4} {
		t.Fatalf("padding = %#v, want %#v", padding, [4]int{0, 2, 0, 4})
	}
}

func TestHelpModelPrefersHoverThenFocus(t *testing.T) {
	model := newHelpModel("Ready", "? ")
	model.Update("focus", 1, "Focused", false, true, [4]int{1, 1, 4, 1})
	if got := model.Current(); got != "? Focused" {
		t.Fatalf("current help = %q, want %q", got, "? Focused")
	}
	model.Update("hover", 2, "Hovered", true, false, [4]int{2, 2, 4, 1})
	if got := model.Current(); got != "? Hovered" {
		t.Fatalf("current help = %q, want %q", got, "? Hovered")
	}
	model.Remove("hover")
	if got := model.Current(); got != "? Focused" {
		t.Fatalf("current help = %q, want %q", got, "? Focused")
	}
	model.Remove("focus")
	if got := model.Current(); got != "? Ready" {
		t.Fatalf("current help = %q, want %q", got, "? Ready")
	}
}

func TestInteractiveSectionEmitsIntentAndResetsOnRelease(t *testing.T) {
	section := resolveSection(ActionBadge(" GO ", "black", "green", testIntent{}).WithKey("go"), rtui.AlignStart, DefaultTheme(), true)
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}

	var emitted intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) { emitted = i })
	if !inst.HandleAction(action.NewAction(action.ActionClick)) {
		t.Fatal("click should be handled")
	}
	if emitted == nil || emitted.IntentType() != (testIntent{}).IntentType() {
		t.Fatal("press intent was not emitted")
	}
	if !inst.state.Pressed {
		t.Fatal("section should enter pressed state on click")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseRelease)) {
		t.Fatal("mouse release should be handled")
	}
	if inst.state.Pressed {
		t.Fatal("section should reset pressed state on release")
	}
}

func TestInteractiveSectionImplementsFocusable(t *testing.T) {
	section := resolveSection(ActionText("Details", testIntent{}), rtui.AlignStart, DefaultTheme(), true)
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}
	if _, ok := interface{}(inst).(rtui.FocusableInstance); !ok {
		t.Fatal("interactive section should implement FocusableInstance")
	}
	if !inst.IsFocusable() {
		t.Fatal("interactive section should be focusable")
	}
	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Fatal("section should have focus after SetFocus(true)")
	}
	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Fatal("section should lose focus after SetFocus(false)")
	}
}

func TestDisabledInteractiveSectionIsNotFocusable(t *testing.T) {
	section := resolveSection(ActionText("Disabled", testIntent{}).WithDisabled(true), rtui.AlignStart, DefaultTheme(), true)
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}
	if inst.IsFocusable() {
		t.Fatal("disabled section should not be focusable")
	}
}

func TestInteractiveSectionTracksHoverState(t *testing.T) {
	section := resolveSection(ActionText("Hover", testIntent{}), rtui.AlignStart, DefaultTheme(), true)
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseEnter)) {
		t.Fatal("mouse enter should be handled")
	}
	if !inst.state.Hovered {
		t.Fatal("section should become hovered on mouse enter")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseLeave)) {
		t.Fatal("mouse leave should be handled")
	}
	if inst.state.Hovered {
		t.Fatal("section should clear hovered state on mouse leave")
	}
}

func TestInteractiveSectionUpdatesHelpModel(t *testing.T) {
	model := newHelpModel("Ready", "? ")
	section := resolveSection(ActionText("Hover", testIntent{}).WithHelp("Open details"), rtui.AlignStart, DefaultTheme(), true)
	section.helpModel = model
	section.helpKey = "hover"
	section.helpOrder = 0
	node := renderSection(section, rtui.AlignStart, DefaultTheme(), true)
	factory, ok := node.(rtui.InstanceFactory)
	if !ok {
		t.Fatal("rendered section is not an instance factory")
	}
	inst, ok := factory.CreateInstance().(*sectionInstance)
	if !ok {
		t.Fatal("instance type mismatch")
	}
	if got := model.Current(); got != "? Ready" {
		t.Fatalf("initial help = %q, want %q", got, "? Ready")
	}
	inst.SetFocus(true)
	if got := model.Current(); got != "? Open details" {
		t.Fatalf("focused help = %q, want %q", got, "? Open details")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseEnter)) {
		t.Fatal("mouse enter should be handled")
	}
	if got := model.Current(); got != "? Open details" {
		t.Fatalf("hover help = %q, want %q", got, "? Open details")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseLeave)) {
		t.Fatal("mouse leave should be handled")
	}
	inst.SetFocus(false)
	if got := model.Current(); got != "? Ready" {
		t.Fatalf("fallback help = %q, want %q", got, "? Ready")
	}
}
