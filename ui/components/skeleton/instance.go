package skeleton

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

const (
	defaultContentWidth = 24
	minPlaceholderWidth = 4
)

// Instance is the runtime entity for Skeleton components.
type Instance struct {
	key              string
	content          rtui.VNode
	loading          bool
	active           bool
	showAvatar       bool
	avatarShape      Shape
	avatarSize       int
	showTitle        bool
	titleWidth       int
	showParagraph    bool
	paragraphRows    int
	paragraphWidths  []int
	width            int
	gap              int
	rootStyle        style.Style
	placeholderStyle style.Style
	dirty            bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Skeleton instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:              proputil.GetString(props, propKey, ""),
		content:          getVNodeProp(props, propContent),
		loading:          proputil.GetBool(props, propLoading, true),
		active:           proputil.GetBool(props, propActive, false),
		showAvatar:       proputil.GetBool(props, propAvatar, false),
		avatarShape:      getShapeProp(props, ShapeSquare),
		avatarSize:       normalizedAvatarSize(proputil.GetInt(props, propAvatarSize, 4)),
		showTitle:        proputil.GetBool(props, propTitle, true),
		titleWidth:       proputil.GetInt(props, propTitleWidth, 0),
		showParagraph:    proputil.GetBool(props, propParagraph, true),
		paragraphRows:    normalizedRows(proputil.GetInt(props, propParagraphRows, 3)),
		paragraphWidths:  cloneInts(getIntsProp(props, propParagraphWidths)),
		width:            proputil.GetInt(props, propWidth, 0),
		gap:              maxInt(0, proputil.GetInt(props, propGap, 1)),
		rootStyle:        proputil.GetStyle(props, propStyle, style.Style{}),
		placeholderStyle: proputil.GetStyle(props, propPlaceholderStyle, style.Style{}),
		dirty:            true,
	}
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}
func (inst *Instance) MarkDirty()            { inst.dirty = true }
func (inst *Instance) IsDirty() bool         { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	old.paragraphWidths = cloneInts(inst.paragraphWidths)

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.content = getVNodePropWithDefault(props, propContent, inst.content)
	inst.loading = proputil.GetBool(props, propLoading, inst.loading)
	inst.active = proputil.GetBool(props, propActive, inst.active)
	inst.showAvatar = proputil.GetBool(props, propAvatar, inst.showAvatar)
	inst.avatarShape = getShapeProp(props, inst.avatarShape)
	inst.avatarSize = normalizedAvatarSize(proputil.GetInt(props, propAvatarSize, inst.avatarSize))
	inst.showTitle = proputil.GetBool(props, propTitle, inst.showTitle)
	inst.titleWidth = proputil.GetInt(props, propTitleWidth, inst.titleWidth)
	inst.showParagraph = proputil.GetBool(props, propParagraph, inst.showParagraph)
	inst.paragraphRows = normalizedRows(proputil.GetInt(props, propParagraphRows, inst.paragraphRows))
	inst.paragraphWidths = cloneInts(getIntsPropWithDefault(props, propParagraphWidths, inst.paragraphWidths))
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.gap = maxInt(0, proputil.GetInt(props, propGap, inst.gap))
	inst.rootStyle = proputil.GetStyle(props, propStyle, inst.rootStyle)
	inst.placeholderStyle = proputil.GetStyle(props, propPlaceholderStyle, inst.placeholderStyle)

	changed := old.key != inst.key ||
		!reflect.DeepEqual(old.content, inst.content) ||
		old.loading != inst.loading ||
		old.active != inst.active ||
		old.showAvatar != inst.showAvatar ||
		old.avatarShape != inst.avatarShape ||
		old.avatarSize != inst.avatarSize ||
		old.showTitle != inst.showTitle ||
		old.titleWidth != inst.titleWidth ||
		old.showParagraph != inst.showParagraph ||
		old.paragraphRows != inst.paragraphRows ||
		!reflect.DeepEqual(old.paragraphWidths, inst.paragraphWidths) ||
		old.width != inst.width ||
		old.gap != inst.gap ||
		old.rootStyle != inst.rootStyle ||
		old.placeholderStyle != inst.placeholderStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propActive:           inst.active,
		propAvatar:           inst.showAvatar,
		propAvatarShape:      inst.avatarShape,
		propAvatarSize:       inst.avatarSize,
		propContent:          inst.content,
		propGap:              inst.gap,
		propKey:              inst.key,
		propLoading:          inst.loading,
		propParagraph:        inst.showParagraph,
		propParagraphRows:    inst.paragraphRows,
		propParagraphWidths:  cloneInts(inst.paragraphWidths),
		propPlaceholderStyle: inst.placeholderStyle,
		propStyle:            inst.rootStyle,
		propTitle:            inst.showTitle,
		propTitleWidth:       inst.titleWidth,
		propWidth:            inst.width,
	}
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.loading {
		if inst.content == nil {
			return nil
		}
		return []rtui.VNode{inst.content}
	}

	contentWidth := inst.contentWidth()
	lines := make([]rtui.VNode, 0, 1+inst.paragraphRows)

	if inst.showTitle {
		lines = append(lines, inst.newPlaceholderLine(inst.resolvedTitleWidth(contentWidth), "title"))
	}
	if inst.showParagraph {
		for index, width := range inst.resolvedParagraphWidths(contentWidth) {
			lines = append(lines, inst.newPlaceholderLine(width, "paragraph-"+strconv.Itoa(index)))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, inst.newPlaceholderLine(contentWidth, "fallback"))
	}

	content := rtui.VStackBuilder(lines...).Gap(1).AlignCross(rtui.AlignStart)
	if !inst.showAvatar && !inst.rootStyle.IsEmpty() {
		content.SetStyleProps(inst.rootStyle)
	}

	if inst.showAvatar {
		rootChildren := []rtui.VNode{
			inst.avatarNode(),
			rtui.Flex(content.Build(), 1),
		}
		root := rtui.HStackBuilder(rootChildren...).Gap(inst.gap).AlignCross(rtui.AlignStart)
		if inst.width > 0 {
			root.Width(inst.width)
		}
		if !inst.rootStyle.IsEmpty() {
			root.SetStyleProps(inst.rootStyle)
		}
		node := root.Build()
		node.SetKey(inst.key + "-skeleton-root")
		return []rtui.VNode{node}
	}

	if inst.width > 0 {
		content.Width(inst.width)
	}
	node := content.Build()
	node.SetKey(inst.key + "-skeleton-root")
	return []rtui.VNode{node}
}

func (inst *Instance) avatarNode() rtui.VNode {
	lines := inst.avatarLines()
	children := make([]rtui.VNode, 0, len(lines))
	for index, line := range lines {
		node := textcomp.New(line).SetStyleProps(inst.resolvedPlaceholderStyle())
		node.SetKey(inst.key + "-avatar-" + strconv.Itoa(index))
		children = append(children, node)
	}
	block := rtui.VStackBuilder(children...).Gap(0).AlignCross(rtui.AlignStart).Build()
	block.SetKey(inst.key + "-avatar")
	return block
}

func (inst *Instance) avatarLines() []string {
	size := normalizedAvatarSize(inst.avatarSize)
	lines := make([]string, 0, size)
	fill := inst.placeholderChar()

	if inst.avatarShape == ShapeRound && size >= 3 {
		topBottomWidth := maxInt(1, size-2)
		leftPad := (size - topBottomWidth) / 2
		rightPad := size - topBottomWidth - leftPad
		topBottom := strings.Repeat(" ", leftPad) + strings.Repeat(fill, topBottomWidth) + strings.Repeat(" ", rightPad)
		for row := 0; row < size; row++ {
			switch row {
			case 0, size - 1:
				lines = append(lines, topBottom)
			default:
				lines = append(lines, strings.Repeat(fill, size))
			}
		}
		return lines
	}

	full := strings.Repeat(fill, size)
	for row := 0; row < size; row++ {
		lines = append(lines, full)
	}
	return lines
}

func (inst *Instance) contentWidth() int {
	if inst.width > 0 {
		available := inst.width
		if inst.showAvatar {
			available -= normalizedAvatarSize(inst.avatarSize) + inst.gap
		}
		return maxInt(minPlaceholderWidth, available)
	}

	width := defaultContentWidth
	if inst.titleWidth > 0 {
		width = maxInt(width, inst.titleWidth)
	}
	for _, paragraphWidth := range inst.paragraphWidths {
		width = maxInt(width, paragraphWidth)
	}
	return maxInt(minPlaceholderWidth, width)
}

func (inst *Instance) resolvedTitleWidth(contentWidth int) int {
	if !inst.showTitle {
		return 0
	}
	if inst.titleWidth > 0 {
		return clampWidth(inst.titleWidth, contentWidth)
	}
	defaultWidth := contentWidth * 3 / 5
	if defaultWidth < 10 {
		defaultWidth = minInt(contentWidth, 10)
	}
	return clampWidth(defaultWidth, contentWidth)
}

func (inst *Instance) resolvedParagraphWidths(contentWidth int) []int {
	if !inst.showParagraph {
		return nil
	}
	rows := normalizedRows(inst.paragraphRows)
	widths := make([]int, rows)
	for index := 0; index < rows; index++ {
		if index < len(inst.paragraphWidths) && inst.paragraphWidths[index] > 0 {
			widths[index] = clampWidth(inst.paragraphWidths[index], contentWidth)
			continue
		}
		switch {
		case index == rows-1:
			widths[index] = clampWidth(maxInt(minPlaceholderWidth*2, contentWidth*2/3), contentWidth)
		case index%2 == 1:
			widths[index] = clampWidth(contentWidth-2, contentWidth)
		default:
			widths[index] = clampWidth(contentWidth, contentWidth)
		}
	}
	return widths
}

func (inst *Instance) newPlaceholderLine(width int, suffix string) rtui.VNode {
	if width <= 0 {
		width = minPlaceholderWidth
	}
	node := textcomp.New(strings.Repeat(inst.placeholderChar(), width)).
		SetStyleProps(inst.resolvedPlaceholderStyle())
	node.SetKey(inst.key + "-" + suffix)
	return node
}

func (inst *Instance) resolvedPlaceholderStyle() style.Style {
	base := style.NewStyle().
		Foreground(theme.Muted()).
		Background(theme.Surface())
	if inst.active {
		base = base.Bold(true)
	}
	if !inst.placeholderStyle.IsEmpty() {
		base = base.Merge(inst.placeholderStyle)
	}
	return base
}

func (inst *Instance) placeholderChar() string {
	if inst.active {
		return "▒"
	}
	return "▓"
}

func getShapeProp(props rtui.Props, def Shape) Shape {
	if value, ok := props[propAvatarShape]; ok {
		if shape, ok := value.(Shape); ok {
			return shape
		}
	}
	return def
}

func getVNodeProp(props rtui.Props, key string) rtui.VNode {
	if value, ok := props[key]; ok {
		if node, ok := value.(rtui.VNode); ok {
			return node
		}
	}
	return nil
}

func getVNodePropWithDefault(props rtui.Props, key string, def rtui.VNode) rtui.VNode {
	if node := getVNodeProp(props, key); node != nil {
		return node
	}
	return def
}

func getIntsProp(props rtui.Props, key string) []int {
	if value, ok := props[key]; ok {
		if values, ok := value.([]int); ok {
			return values
		}
	}
	return nil
}

func getIntsPropWithDefault(props rtui.Props, key string, def []int) []int {
	if values := getIntsProp(props, key); values != nil {
		return values
	}
	return def
}

func normalizedAvatarSize(size int) int {
	if size < 2 {
		return 2
	}
	return size
}

func normalizedRows(rows int) int {
	if rows < 1 {
		return 1
	}
	return rows
}

func clampWidth(width, contentWidth int) int {
	if width < minPlaceholderWidth {
		width = minPlaceholderWidth
	}
	if contentWidth >= minPlaceholderWidth && width > contentWidth {
		return contentWidth
	}
	return width
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
