package result

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for Result components.
type Instance struct {
	key           string
	status        Status
	icon          string
	title         string
	subtitle      string
	extra         rtui.VNode
	bordered      bool
	width         int
	rootStyle     style.Style
	iconStyle     style.Style
	titleStyle    style.Style
	subtitleStyle style.Style
	dirty         bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Result instance.
func NewInstance(props rtui.Props) *Instance {
	return &Instance{
		key:           proputil.GetString(props, propKey, ""),
		status:        getStatusProp(props, StatusInfo),
		icon:          proputil.GetString(props, propIcon, ""),
		title:         proputil.GetString(props, propTitle, ""),
		subtitle:      proputil.GetString(props, propSubtitle, ""),
		extra:         getVNodeProp(props, propExtra),
		bordered:      proputil.GetBool(props, propBordered, false),
		width:         proputil.GetInt(props, propWidth, 0),
		rootStyle:     proputil.GetStyle(props, propResultStyle, style.Style{}),
		iconStyle:     proputil.GetStyle(props, propIconStyle, style.Style{}),
		titleStyle:    proputil.GetStyle(props, propTitleStyle, style.Style{}),
		subtitleStyle: proputil.GetStyle(props, propSubtitleStyle, style.Style{}),
		dirty:         true,
	}
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.status = getStatusProp(props, inst.status)
	inst.icon = proputil.GetString(props, propIcon, inst.icon)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.subtitle = proputil.GetString(props, propSubtitle, inst.subtitle)
	inst.extra = getVNodePropWithDefault(props, propExtra, inst.extra)
	inst.bordered = proputil.GetBool(props, propBordered, inst.bordered)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.rootStyle = proputil.GetStyle(props, propResultStyle, inst.rootStyle)
	inst.iconStyle = proputil.GetStyle(props, propIconStyle, inst.iconStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.subtitleStyle = proputil.GetStyle(props, propSubtitleStyle, inst.subtitleStyle)

	changed := old.key != inst.key ||
		old.status != inst.status ||
		old.icon != inst.icon ||
		old.title != inst.title ||
		old.subtitle != inst.subtitle ||
		!reflect.DeepEqual(old.extra, inst.extra) ||
		old.bordered != inst.bordered ||
		old.width != inst.width ||
		old.rootStyle != inst.rootStyle ||
		old.iconStyle != inst.iconStyle ||
		old.titleStyle != inst.titleStyle ||
		old.subtitleStyle != inst.subtitleStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBordered:      inst.bordered,
		propExtra:         inst.extra,
		propIcon:          inst.icon,
		propIconStyle:     inst.iconStyle,
		propKey:           inst.key,
		propResultStyle:   inst.rootStyle,
		propStatus:        inst.status,
		propSubtitle:      inst.subtitle,
		propSubtitleStyle: inst.subtitleStyle,
		propTitle:         inst.title,
		propTitleStyle:    inst.titleStyle,
		propWidth:         inst.width,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	children := make([]rtui.VNode, 0, 4)
	icon := inst.iconText()
	if icon != "" {
		iconNode := textcomp.New(icon).
			SetStyleProps(inst.iconTextStyle()).
			SetTextAlignProps(rtui.AlignCenter)
		children = append(children, iconNode)
	}

	title := inst.titleText()
	if title != "" {
		titleNode := textcomp.New(title).
			SetStyleProps(inst.titleTextStyle()).
			SetTextAlignProps(rtui.AlignCenter)
		children = append(children, titleNode)
	}

	if subtitle := strings.TrimSpace(inst.subtitle); subtitle != "" {
		subtitleNode := textcomp.New(subtitle).
			SetWrap(true).
			SetMaxWidth(inst.maxContentWidth()).
			SetStyleProps(inst.subtitleTextStyle()).
			SetTextAlignProps(rtui.AlignCenter)
		children = append(children, subtitleNode)
	}

	if inst.extra != nil {
		children = append(children, inst.extra)
	}

	if len(children) == 0 {
		return nil
	}

	builder := rtui.VStackBuilder(children...).Gap(1).AlignCross(rtui.AlignCenter)
	if inst.width > 0 {
		builder.Width(inst.width)
	}
	if inst.bordered {
		builder.SingleBorder()
		builder.SetBorderColor(inst.statusColor())
	}
	if !inst.rootStyle.IsEmpty() {
		builder.SetStyleProps(inst.rootStyle)
	}
	node := builder.Build()
	node.SetKey(inst.key + "-result")
	return []rtui.VNode{node}
}

func (inst *Instance) iconText() string {
	if strings.TrimSpace(inst.icon) != "" {
		return inst.icon
	}
	switch inst.status {
	case StatusSuccess:
		return "✔"
	case StatusWarning:
		return "!"
	case StatusError:
		return "✖"
	case Status403:
		return "403"
	case Status404:
		return "404"
	case Status500:
		return "500"
	default:
		return "i"
	}
}

func (inst *Instance) titleText() string {
	if strings.TrimSpace(inst.title) != "" {
		return inst.title
	}
	switch inst.status {
	case StatusSuccess:
		return "Operation completed"
	case StatusWarning:
		return "Attention required"
	case StatusError:
		return "Operation failed"
	case Status403:
		return "403 Forbidden"
	case Status404:
		return "404 Not Found"
	case Status500:
		return "500 Server Error"
	default:
		return "Information"
	}
}

func (inst *Instance) statusColor() style.Color {
	switch inst.status {
	case StatusSuccess:
		return theme.Success()
	case StatusWarning:
		return theme.Warning()
	case StatusError, Status500:
		return theme.Error()
	case Status403, Status404:
		return theme.Primary()
	default:
		return theme.Primary()
	}
}

func (inst *Instance) iconTextStyle() style.Style {
	base := style.NewStyle().Foreground(inst.statusColor()).Bold(true)
	if !inst.iconStyle.IsEmpty() {
		base = base.Merge(inst.iconStyle)
	}
	return base
}

func (inst *Instance) titleTextStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Text()).Bold(true)
	if !inst.titleStyle.IsEmpty() {
		base = base.Merge(inst.titleStyle)
	}
	return base
}

func (inst *Instance) subtitleTextStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Muted())
	if !inst.subtitleStyle.IsEmpty() {
		base = base.Merge(inst.subtitleStyle)
	}
	return base
}

func (inst *Instance) maxContentWidth() int {
	if inst.width > 4 {
		return inst.width - 4
	}
	return 48
}

func getStatusProp(props rtui.Props, def Status) Status {
	if value, ok := props[propStatus]; ok {
		if status, ok := value.(Status); ok {
			return status
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
