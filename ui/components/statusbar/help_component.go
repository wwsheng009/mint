package statusbar

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type helpLineVNode struct {
	*rtui.ElementVNode
}

func newHelpLineVNode(model *helpModel, lineStyle style.Style) *helpLineVNode {
	node := &helpLineVNode{ElementVNode: rtui.NewElement("statusbar-help")}
	node.SetStyle(lineStyle)
	node.SetProp("style", lineStyle)
	node.SetProp("helpModel", model)
	return node
}

func (v *helpLineVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["style"] = v.Style()
	return newHelpLineInstance(props)
}

type overlayHelpVNode struct {
	*rtui.ElementVNode
}

func newOverlayHelpVNode(model *helpModel, fillStyle, borderStyle, shadowStyle style.Style, arrowStyle TooltipArrowStyle, placement TooltipPlacement, maxContentWidth, gapRows, bottomOffsetRows int) *overlayHelpVNode {
	node := &overlayHelpVNode{ElementVNode: rtui.NewElement("statusbar-help-overlay")}
	node.SetStyle(fillStyle)
	node.SetProp("style", fillStyle)
	node.SetProp("helpModel", model)
	node.SetProp("tooltipBorderStyle", borderStyle)
	node.SetProp("tooltipShadowStyle", shadowStyle)
	node.SetProp("tooltipArrowStyle", arrowStyle)
	node.SetProp("tooltipPlacement", placement)
	node.SetProp("tooltipMaxWidth", maxContentWidth)
	node.SetProp("tooltipGapRows", gapRows)
	node.SetProp("bottomOffsetRows", bottomOffsetRows)
	return node
}

func (v *overlayHelpVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["style"] = v.Style()
	return newOverlayHelpInstance(props)
}

func (v *overlayHelpVNode) GetLayer() rtui.Layer {
	return rtui.LayerTooltip
}

func (v *overlayHelpVNode) SetLayer(l rtui.Layer) rtui.VNode {
	return v
}

type helpLineInstance struct {
	model     *helpModel
	lineStyle style.Style
	bounds    [4]int
	dirty     bool
}

type overlayHelpInstance struct {
	model            *helpModel
	fillStyle        style.Style
	borderStyle      style.Style
	shadowStyle      style.Style
	arrowStyle       TooltipArrowStyle
	placement        TooltipPlacement
	maxContentWidth  int
	gapRows          int
	bottomOffsetRows int
	bounds           [4]int
	dirty            bool
}

type overlayTooltipBox struct {
	x        int
	y        int
	width    int
	height   int
	lines    []string
	shadowX  int
	shadowY  int
	shadowW  int
	shadowH  int
	arrowX   int
	arrowY   int
	arrow    string
	hasArrow bool
}

var (
	_ rtui.ComponentInstance = (*helpLineInstance)(nil)
	_ rtui.PaintableInstance = (*helpLineInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*helpLineInstance)(nil)
	_ rtui.ComponentInstance = (*overlayHelpInstance)(nil)
	_ rtui.PaintableInstance = (*overlayHelpInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*overlayHelpInstance)(nil)
)

func newHelpLineInstance(props rtui.Props) *helpLineInstance {
	return &helpLineInstance{
		model:     getHelpModelProp(props),
		lineStyle: getSectionStyleProp(props),
		dirty:     true,
	}
}

func newOverlayHelpInstance(props rtui.Props) *overlayHelpInstance {
	return &overlayHelpInstance{
		model:            getHelpModelProp(props),
		fillStyle:        getSectionStyleProp(props),
		borderStyle:      getSectionStylePropKey(props, "tooltipBorderStyle"),
		shadowStyle:      getSectionStylePropKey(props, "tooltipShadowStyle"),
		arrowStyle:       getTooltipArrowStyleProp(props, "tooltipArrowStyle", TooltipArrowStyleSharp),
		placement:        getTooltipPlacementProp(props, "tooltipPlacement", TooltipPlacementAuto),
		maxContentWidth:  getSectionIntProp(props, "tooltipMaxWidth", 48),
		gapRows:          getSectionIntProp(props, "tooltipGapRows", 1),
		bottomOffsetRows: getSectionIntProp(props, "bottomOffsetRows", 0),
		dirty:            true,
	}
}

func (inst *helpLineInstance) Key() string                        { return "" }
func (inst *helpLineInstance) SetKey(key string)                  {}
func (inst *helpLineInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *helpLineInstance) Destroy()                           {}
func (inst *helpLineInstance) OnMount()                           {}
func (inst *helpLineInstance) OnUnmount()                         {}
func (inst *helpLineInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *helpLineInstance) MarkDirty()                         { inst.dirty = true }
func (inst *helpLineInstance) IsDirty() bool                      { return inst.dirty }

func (inst *helpLineInstance) SetProps(props rtui.Props) bool {
	oldModel := inst.model
	oldStyle := inst.lineStyle
	inst.model = getHelpModelProp(props)
	inst.lineStyle = getSectionStyleProp(props)
	changed := oldModel != inst.model || oldStyle != inst.lineStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *helpLineInstance) GetProps() rtui.Props {
	return rtui.Props{"style": inst.lineStyle, "helpModel": inst.model}
}

func (inst *helpLineInstance) Paint(x, y int) []paint.DrawCmd {
	if inst.model == nil {
		return nil
	}
	text := inst.model.Current()
	if text == "" {
		return nil
	}
	if inst.bounds[2] > 0 {
		text = fitText(text, inst.bounds[2], rtui.AlignStart, OverflowClip)
	}
	return []paint.DrawCmd{{X: x, Y: y, Text: text, Style: inst.lineStyle}}
}

func (inst *helpLineInstance) Measure(constraints layout.Constraints) layout.Size {
	text := ""
	if inst.model != nil {
		text = inst.model.Current()
	}
	if text == "" {
		return layout.Size{Width: 0, Height: 1}
	}
	return layout.Size{Width: paint.StringWidth(text), Height: 1}
}

func (inst *helpLineInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *helpLineInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *overlayHelpInstance) Key() string                        { return "" }
func (inst *overlayHelpInstance) SetKey(key string)                  {}
func (inst *overlayHelpInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *overlayHelpInstance) Destroy()                           {}
func (inst *overlayHelpInstance) OnMount()                           {}
func (inst *overlayHelpInstance) OnUnmount()                         {}
func (inst *overlayHelpInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *overlayHelpInstance) MarkDirty()                         { inst.dirty = true }
func (inst *overlayHelpInstance) IsDirty() bool                      { return inst.dirty }

func (inst *overlayHelpInstance) SetProps(props rtui.Props) bool {
	oldModel := inst.model
	oldFill := inst.fillStyle
	oldBorder := inst.borderStyle
	oldShadow := inst.shadowStyle
	oldArrowStyle := inst.arrowStyle
	oldPlacement := inst.placement
	oldMaxWidth := inst.maxContentWidth
	oldGap := inst.gapRows
	oldOffset := inst.bottomOffsetRows
	inst.model = getHelpModelProp(props)
	inst.fillStyle = getSectionStyleProp(props)
	inst.borderStyle = getSectionStylePropKey(props, "tooltipBorderStyle")
	inst.shadowStyle = getSectionStylePropKey(props, "tooltipShadowStyle")
	inst.arrowStyle = getTooltipArrowStyleProp(props, "tooltipArrowStyle", inst.arrowStyle)
	inst.placement = getTooltipPlacementProp(props, "tooltipPlacement", inst.placement)
	inst.maxContentWidth = getSectionIntProp(props, "tooltipMaxWidth", inst.maxContentWidth)
	inst.gapRows = getSectionIntProp(props, "tooltipGapRows", inst.gapRows)
	inst.bottomOffsetRows = getSectionIntProp(props, "bottomOffsetRows", inst.bottomOffsetRows)
	changed := oldModel != inst.model || oldFill != inst.fillStyle || oldBorder != inst.borderStyle || oldShadow != inst.shadowStyle || oldArrowStyle != inst.arrowStyle || oldPlacement != inst.placement || oldMaxWidth != inst.maxContentWidth || oldGap != inst.gapRows || oldOffset != inst.bottomOffsetRows
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *overlayHelpInstance) GetProps() rtui.Props {
	return rtui.Props{
		"style":              inst.fillStyle,
		"helpModel":          inst.model,
		"tooltipBorderStyle": inst.borderStyle,
		"tooltipShadowStyle": inst.shadowStyle,
		"tooltipArrowStyle":  inst.arrowStyle,
		"tooltipPlacement":   inst.placement,
		"tooltipMaxWidth":    inst.maxContentWidth,
		"tooltipGapRows":     inst.gapRows,
		"bottomOffsetRows":   inst.bottomOffsetRows,
	}
}

func (inst *overlayHelpInstance) Paint(x, y int) []paint.DrawCmd {
	if inst.model == nil {
		return nil
	}
	text, anchor, ok := inst.model.HoveredActive()
	if !ok || text == "" {
		return nil
	}
	box := inst.computeTooltipBox(text, anchor)
	if box.width <= 0 || box.height <= 0 || len(box.lines) == 0 {
		return nil
	}

	fillStyle := inst.fillStyle
	borderStyle := inst.borderStyle
	if borderStyle.IsEmpty() {
		borderStyle = fillStyle
	}
	shadowStyle := inst.shadowStyle
	if shadowStyle.IsEmpty() {
		shadowStyle = style.NewStyle().Foreground(style.BrightBlack)
		if fillStyle.BG != "" {
			shadowStyle = shadowStyle.Background(fillStyle.BG)
		}
	}

	cmds := make([]paint.DrawCmd, 0, len(box.lines)*3+box.shadowH+box.height)
	for rowIndex, line := range box.lines {
		yPos := box.y + rowIndex
		runes := []rune(line)
		if len(runes) < 2 {
			cmds = append(cmds, paint.DrawCmd{X: box.x, Y: yPos, Text: line, Style: borderStyle})
			continue
		}
		left := string(runes[:1])
		right := string(runes[len(runes)-1:])
		middle := string(runes[1 : len(runes)-1])
		if rowIndex == 0 || rowIndex == len(box.lines)-1 {
			cmds = append(cmds, paint.DrawCmd{X: box.x, Y: yPos, Text: line, Style: borderStyle})
			continue
		}
		cmds = append(cmds,
			paint.DrawCmd{X: box.x, Y: yPos, Text: left, Style: borderStyle},
			paint.DrawCmd{X: box.x + 1, Y: yPos, Text: middle, Style: fillStyle},
			paint.DrawCmd{X: box.x + 1 + paint.StringWidth(middle), Y: yPos, Text: right, Style: borderStyle},
		)
	}

	shadowRune := "░"
	for row := 0; row < box.shadowH; row++ {
		cmds = append(cmds, paint.DrawCmd{
			X:     box.shadowX,
			Y:     box.shadowY + row,
			Text:  strings.Repeat(shadowRune, box.shadowW),
			Style: shadowStyle,
		})
	}
	for row := 0; row < box.height-1; row++ {
		cmds = append(cmds, paint.DrawCmd{
			X:     box.x + box.width,
			Y:     box.y + row + 1,
			Text:  shadowRune,
			Style: shadowStyle,
		})
	}

	return cmds
}

func (inst *overlayHelpInstance) computeTooltipBox(text string, anchor [4]int) overlayTooltipBox {
	viewportWidth := inst.bounds[2]
	viewportHeight := inst.bounds[3]
	contentWidth := inst.maxContentWidth
	if contentWidth <= 0 {
		contentWidth = 48
	}
	if viewportWidth > 0 {
		maxAllowed := viewportWidth - 8
		if maxAllowed < 12 {
			maxAllowed = 12
		}
		if contentWidth > maxAllowed {
			contentWidth = maxAllowed
		}
	}

	lines := wrapByDisplayWidth(normalizeStatusText(text), contentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}
	innerWidth := 0
	for _, line := range lines {
		if w := paint.StringWidth(line); w > innerWidth {
			innerWidth = w
		}
	}
	if innerWidth <= 0 {
		innerWidth = 1
	}

	boxWidth := innerWidth + 4
	boxHeight := len(lines) + 2
	shadowH := 1
	tooltipX := resolveTooltipX(anchor, boxWidth, viewportWidth)
	arrowX := resolveTooltipArrowX(anchor, tooltipX, boxWidth)
	tooltipY, placement := resolveTooltipY(anchor, inst.placement, boxHeight, inst.gapRows, inst.bottomOffsetRows, shadowH, viewportHeight)
	boxLines := buildOverlayTooltipLines(lines, innerWidth, placement, inst.arrowStyle, true, arrowX-tooltipX)

	return overlayTooltipBox{
		x:        tooltipX,
		y:        tooltipY,
		width:    boxWidth,
		height:   boxHeight,
		lines:    boxLines,
		shadowX:  tooltipX + 1,
		shadowY:  tooltipY + boxHeight,
		shadowW:  boxWidth,
		shadowH:  shadowH,
		arrowX:   arrowX,
		arrowY:   -1,
		arrow:    overlayTooltipArrowRune(placement, inst.arrowStyle),
		hasArrow: true,
	}
}

func buildOverlayTooltipLines(lines []string, innerWidth int, placement TooltipPlacement, arrowStyle TooltipArrowStyle, hasArrow bool, arrowOffset int) []string {
	topLeft, topRight, bottomLeft, bottomRight := tooltipCornerRunes(arrowStyle)
	top := topLeft + strings.Repeat("─", innerWidth+2) + topRight
	bottom := bottomLeft + strings.Repeat("─", innerWidth+2) + bottomRight
	if hasArrow {
		arrowRune := []rune(overlayTooltipArrowRune(placement, arrowStyle))[0]
		switch placement {
		case TooltipPlacementTop:
			bottom = replaceBorderRune(bottom, arrowOffset, arrowRune)
		default:
			top = replaceBorderRune(top, arrowOffset, arrowRune)
		}
	}

	boxLines := make([]string, 0, len(lines)+2)
	boxLines = append(boxLines, top)
	for _, line := range lines {
		boxLines = append(boxLines, "│ "+fitText(line, innerWidth, rtui.AlignStart, OverflowClip)+" │")
	}
	boxLines = append(boxLines, bottom)
	return boxLines
}

func replaceBorderRune(content string, index int, replacement rune) string {
	runes := []rune(content)
	if index <= 0 || index >= len(runes)-1 {
		return content
	}
	runes[index] = replacement
	return string(runes)
}

func resolveTooltipArrowX(anchor [4]int, boxX, boxWidth int) int {
	arrowX := anchor[0]
	if anchor[2] > 0 {
		arrowX = anchor[0] + anchor[2]/2
	}
	minX := boxX + 1
	maxX := boxX + boxWidth - 2
	if arrowX < minX {
		arrowX = minX
	}
	if arrowX > maxX {
		arrowX = maxX
	}
	return arrowX
}

func resolveTooltipY(anchor [4]int, placement TooltipPlacement, boxHeight, gapRows, bottomOffsetRows, shadowH, viewportHeight int) (int, TooltipPlacement) {
	if gapRows < 0 {
		gapRows = 0
	}
	if bottomOffsetRows < 0 {
		bottomOffsetRows = 0
	}

	aboveBoxY := anchor[1] - boxHeight - gapRows
	belowBoxY := anchor[1] + anchor[3] + gapRows + bottomOffsetRows

	resolved := placement
	if resolved == TooltipPlacementAuto {
		fitsBelow := viewportHeight <= 0 || belowBoxY+boxHeight+shadowH <= viewportHeight
		fitsAbove := aboveBoxY >= 0 && (viewportHeight <= 0 || aboveBoxY+boxHeight+shadowH <= viewportHeight)
		switch {
		case fitsAbove && !fitsBelow:
			resolved = TooltipPlacementTop
		case viewportHeight > 0 && fitsAbove && anchor[1] > viewportHeight/2:
			resolved = TooltipPlacementTop
		default:
			resolved = TooltipPlacementBottom
		}
		if viewportHeight <= 0 && anchor[1] > boxHeight+gapRows {
			resolved = TooltipPlacementTop
		}
	}

	boxY := belowBoxY
	if resolved == TooltipPlacementTop {
		boxY = aboveBoxY
	}

	if viewportHeight > 0 {
		maxBoxY := viewportHeight - boxHeight - shadowH
		if maxBoxY < 0 {
			maxBoxY = 0
		}
		if boxY < 0 {
			boxY = 0
		}
		if boxY > maxBoxY {
			boxY = maxBoxY
		}
	}

	return boxY, resolved
}

func overlayTooltipArrowRune(placement TooltipPlacement, arrowStyle TooltipArrowStyle) string {
	if arrowStyle == TooltipArrowStyleRounded {
		if placement == TooltipPlacementTop {
			return "▽"
		}
		return "△"
	}
	if placement == TooltipPlacementTop {
		return "▼"
	}
	return "▲"
}

func tooltipCornerRunes(arrowStyle TooltipArrowStyle) (string, string, string, string) {
	if arrowStyle == TooltipArrowStyleRounded {
		return "╭", "╮", "╰", "╯"
	}
	return "┌", "┐", "└", "┘"
}

func (inst *overlayHelpInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{}
}

func (inst *overlayHelpInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *overlayHelpInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func resolveTooltipX(anchor [4]int, boxWidth, viewportWidth int) int {
	anchorLeft := anchor[0]
	anchorRight := anchor[0] + anchor[2]
	tooltipX := anchorLeft

	if viewportWidth <= 0 {
		return tooltipX
	}

	maxX := viewportWidth - boxWidth - 1
	if maxX < 0 {
		maxX = 0
	}
	if tooltipX >= 0 && tooltipX <= maxX {
		return tooltipX
	}

	tooltipX = anchorRight - boxWidth
	if tooltipX >= 0 && tooltipX <= maxX {
		return tooltipX
	}

	tooltipX = anchor[0] + anchor[2]/2 - boxWidth/2
	if tooltipX < 0 {
		tooltipX = 0
	}
	if tooltipX > maxX {
		tooltipX = maxX
	}
	return tooltipX
}

func wrapByDisplayWidth(text string, maxWidth int) []string {
	text = normalizeStatusText(text)
	if text == "" {
		return []string{""}
	}
	if maxWidth <= 0 {
		return []string{text}
	}

	var lines []string
	var current strings.Builder
	currentWidth := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if currentWidth+rw > maxWidth && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(r)
		currentWidth += rw
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func getHelpModelProp(props rtui.Props) *helpModel {
	if v, ok := props["helpModel"].(*helpModel); ok {
		return v
	}
	return nil
}

func getTooltipArrowStyleProp(props rtui.Props, key string, fallback TooltipArrowStyle) TooltipArrowStyle {
	if v, ok := props[key].(TooltipArrowStyle); ok {
		return v
	}
	return fallback
}

func getTooltipPlacementProp(props rtui.Props, key string, fallback TooltipPlacement) TooltipPlacement {
	if v, ok := props[key].(TooltipPlacement); ok {
		return v
	}
	return fallback
}
