package filterbar

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	inputcomp "github.com/wwsheng009/mint/ui/components/input"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/text"
	wrapcomp "github.com/wwsheng009/mint/ui/components/wrap"
)

// Instance is the runtime entity for FilterBar.
type Instance struct {
	key        string
	title      string
	summary    string
	fields     []Field
	actions    []Action
	width      int
	gap        int
	rowGap     int
	rootGap    int
	wrap       bool
	labelWidth int
	rootStyle  style.Style
	dirty      bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a FilterBar instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:        getStringProp(props, propKey, ""),
		title:      getStringProp(props, propTitle, ""),
		summary:    getStringProp(props, propSummary, ""),
		fields:     normalizeFields(getFieldsProp(props, nil)),
		actions:    normalizeActions(getActionsProp(props, nil)),
		width:      getIntProp(props, propWidth, 0),
		gap:        getIntProp(props, propGap, 1),
		rowGap:     getIntProp(props, propRowGap, 1),
		rootGap:    getIntProp(props, propRootGap, 0),
		wrap:       getBoolProp(props, propWrap, false),
		labelWidth: getIntProp(props, propLabelWidth, 0),
		rootStyle:  getStyleProp(props, propStyle, style.Style{}),
		dirty:      true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()   {}
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() {}
func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := inst.snapshot()

	inst.key = getStringProp(props, propKey, inst.key)
	inst.title = getStringProp(props, propTitle, inst.title)
	inst.summary = getStringProp(props, propSummary, inst.summary)
	inst.fields = normalizeFields(getFieldsProp(props, inst.fields))
	inst.actions = normalizeActions(getActionsProp(props, inst.actions))
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.gap = getIntProp(props, propGap, inst.gap)
	inst.rowGap = getIntProp(props, propRowGap, inst.rowGap)
	inst.rootGap = getIntProp(props, propRootGap, inst.rootGap)
	inst.wrap = getBoolProp(props, propWrap, inst.wrap)
	inst.labelWidth = getIntProp(props, propLabelWidth, inst.labelWidth)
	inst.rootStyle = getStyleProp(props, propStyle, inst.rootStyle)
	inst.normalize()

	changed := !reflect.DeepEqual(old, inst.snapshot())
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propActions:    cloneActions(inst.actions),
		propFields:     cloneFields(inst.fields),
		propGap:        inst.gap,
		propKey:        inst.key,
		propLabelWidth: inst.labelWidth,
		propRowGap:     inst.rowGap,
		propRootGap:    inst.rootGap,
		propStyle:      inst.rootStyle,
		propSummary:    inst.summary,
		propTitle:      inst.title,
		propWidth:      inst.width,
		propWrap:       inst.wrap,
	}
}

// RuntimeChildren synthesizes the filter controls used by Fiber.
func (inst *Instance) RuntimeChildren() []rtui.VNode {
	controls := inst.buildControls()
	if len(controls) == 0 && strings.TrimSpace(inst.title) == "" && strings.TrimSpace(inst.summary) == "" {
		return nil
	}

	children := make([]rtui.VNode, 0, 4)
	if strings.TrimSpace(inst.title) != "" {
		children = append(children, text.NewBuilder(inst.title).
			Key(inst.childKey("title")).
			Style(style.NewStyle().Foreground(theme.Text()).Bold(true)).
			Build())
	}
	if strings.TrimSpace(inst.summary) != "" {
		children = append(children, text.NewBuilder(inst.summary).
			Key(inst.childKey("summary")).
			Style(style.NewStyle().Foreground(theme.Muted())).
			Build())
	}
	if len(controls) > 0 {
		children = append(children, inst.buildControlRoot(controls))
	}
	if reasons := inst.buildDisabledReasons(); reasons != nil {
		children = append(children, reasons)
	}

	root := rtui.VStackBuilder(children...).Gap(inst.effectiveRootGap()).AlignCross(rtui.AlignStart)
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}
	root.SetKey(inst.rootKey())
	return []rtui.VNode{root.Build()}
}

func (inst *Instance) buildControls() []rtui.VNode {
	controls := make([]rtui.VNode, 0, len(inst.fields)+len(inst.actions))
	for _, field := range inst.fields {
		if control := inst.buildField(field); control != nil {
			controls = append(controls, control)
		}
	}
	for _, action := range inst.actions {
		if control := inst.buildAction(action); control != nil {
			controls = append(controls, control)
		}
	}
	return controls
}

func (inst *Instance) buildControlRoot(controls []rtui.VNode) rtui.VNode {
	if inst.wrap {
		builder := wrapcomp.NewBuilder().
			Key(inst.childKey("wrap")).
			Gap(inst.gap).
			RowGap(inst.rowGap).
			Children(controls...)
		if inst.width > 0 {
			builder.Width(inst.width)
		}
		return builder.Build()
	}

	root := rtui.HStackBuilder(controls...).Gap(inst.gap).AlignCross(rtui.AlignCenter)
	if inst.width > 0 {
		root.Width(inst.width)
	}
	root.SetKey(inst.childKey("row"))
	return root.Build()
}

func (inst *Instance) buildField(field Field) rtui.VNode {
	control := inst.buildFieldControl(field)
	if control == nil {
		return nil
	}
	if strings.TrimSpace(field.Label) == "" {
		control.SetKey(inst.childKey("field-" + field.Key))
		return control
	}

	label := rtui.VNode(text.NewBuilder(field.Label).
		Key(inst.childKey("label-" + field.Key)).
		Style(style.NewStyle().Foreground(theme.Muted()).Bold(true)).
		Build())
	if width := inst.effectiveLabelWidth(field); width > 0 {
		box := rtui.Box().Width(width).Child(label).Build()
		box.SetKey(inst.childKey("label-box-" + field.Key))
		label = box
	}

	row := rtui.HStackBuilder(label, control).Gap(1).AlignCross(rtui.AlignCenter)
	row.SetKey(inst.childKey("field-" + field.Key))
	return row.Build()
}

func (inst *Instance) buildFieldControl(field Field) rtui.VNode {
	switch field.Kind {
	case FieldSearch:
		builder := inputcomp.NewBuilder().
			Key(inst.childKey("control-" + field.Key)).
			Search().
			NoBorder().
			Value(field.Value).
			Placeholder(inst.placeholder(field, "Search")).
			Width(inst.effectiveFieldWidth(field, 22)).
			Disabled(field.Disabled)
		inst.bindInput(builder, field)
		return builder.Build()
	case FieldSelect:
		builder := selectcomp.NewBuilder().
			Key(inst.childKey("control-" + field.Key)).
			Options(cloneOptions(field.Options)).
			Placeholder(inst.placeholder(field, "All")).
			Width(inst.effectiveFieldWidth(field, 18)).
			Disabled(field.Disabled)
		if field.HasSelectedIndex {
			builder.Selected(field.SelectedIndex)
		}
		inst.bindSelect(builder, field)
		return builder.Build()
	case FieldCustom:
		if field.Custom == nil {
			return nil
		}
		return field.Custom
	default:
		builder := inputcomp.NewBuilder().
			Key(inst.childKey("control-" + field.Key)).
			NoBorder().
			Value(field.Value).
			Placeholder(inst.placeholder(field, "Filter")).
			Width(inst.effectiveFieldWidth(field, 18)).
			Disabled(field.Disabled)
		inst.bindInput(builder, field)
		return builder.Build()
	}
}

func (inst *Instance) bindInput(builder *inputcomp.Builder, field Field) {
	if field.ChangeIntent != nil {
		builder.OnChange(field.ChangeIntent)
	} else if strings.TrimSpace(field.FieldName) != "" {
		builder.ForField(intent.BindField(field.FieldName))
	}
	if field.SubmitIntent != nil {
		builder.OnSubmit(field.SubmitIntent)
	}
}

func (inst *Instance) bindSelect(builder *selectcomp.Builder, field Field) {
	if field.ChangeIntent != nil {
		builder.OnChange(field.ChangeIntent)
	} else if strings.TrimSpace(field.FieldName) != "" {
		builder.ForField(intent.BindField(field.FieldName))
	}
}

func (inst *Instance) buildAction(action Action) rtui.VNode {
	if strings.TrimSpace(action.Label) == "" {
		return nil
	}
	builder := button.NewBuilder(action.Label).
		Key(inst.childKey("action-" + action.Key)).
		Variant(action.Variant).
		Disabled(action.Disabled)
	if action.PressIntent != nil {
		builder.OnPress(action.PressIntent)
	}
	node := builder.Build()
	if action.Width > 0 {
		box := rtui.Box().Width(action.Width).Child(node).Build()
		box.SetKey(inst.childKey("action-box-" + action.Key))
		return box
	}
	return node
}

func (inst *Instance) buildDisabledReasons() rtui.VNode {
	reasons := make([]string, 0, len(inst.actions))
	for _, action := range inst.actions {
		reason := strings.TrimSpace(action.DisabledReason)
		if action.Disabled && reason != "" {
			label := strings.TrimSpace(action.Label)
			if label != "" {
				reasons = append(reasons, label+": "+reason)
			} else {
				reasons = append(reasons, reason)
			}
		}
	}
	if len(reasons) == 0 {
		return nil
	}
	node := text.NewBuilder("Disabled: " + strings.Join(reasons, " | ")).
		Key(inst.childKey("disabled-reasons")).
		Style(style.NewStyle().Foreground(theme.Muted())).
		Build()
	return node
}

func (inst *Instance) effectiveLabelWidth(field Field) int {
	if field.LabelWidth > 0 {
		return field.LabelWidth
	}
	return inst.labelWidth
}

func (inst *Instance) effectiveFieldWidth(field Field, def int) int {
	if field.Width > 0 {
		return field.Width
	}
	return def
}

func (inst *Instance) placeholder(field Field, def string) string {
	if strings.TrimSpace(field.Placeholder) != "" {
		return field.Placeholder
	}
	return def
}

func (inst *Instance) effectiveRootGap() int {
	return inst.rootGap
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "filterbar-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) childKey(suffix string) string {
	if inst.key == "" {
		return "filterbar-" + suffix
	}
	return inst.key + "-" + suffix
}

func (inst *Instance) normalize() {
	inst.fields = normalizeFields(inst.fields)
	inst.actions = normalizeActions(inst.actions)
	inst.title = normalizeInlineText(inst.title)
	inst.summary = normalizeInlineText(inst.summary)
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.gap < 0 {
		inst.gap = 0
	}
	if inst.rowGap < 0 {
		inst.rowGap = 0
	}
	if inst.rootGap < 0 {
		inst.rootGap = 0
	}
	if inst.labelWidth < 0 {
		inst.labelWidth = 0
	}
}

type instanceSnapshot struct {
	key        string
	title      string
	summary    string
	fields     []Field
	actions    []Action
	width      int
	gap        int
	rowGap     int
	rootGap    int
	wrap       bool
	labelWidth int
	rootStyle  style.Style
}

func (inst *Instance) snapshot() instanceSnapshot {
	return instanceSnapshot{
		key:        inst.key,
		title:      inst.title,
		summary:    inst.summary,
		fields:     cloneFields(inst.fields),
		actions:    cloneActions(inst.actions),
		width:      inst.width,
		gap:        inst.gap,
		rowGap:     inst.rowGap,
		rootGap:    inst.rootGap,
		wrap:       inst.wrap,
		labelWidth: inst.labelWidth,
		rootStyle:  inst.rootStyle,
	}
}

func getFieldsProp(props rtui.Props, def []Field) []Field {
	if fields, ok := props[propFields].([]Field); ok {
		return cloneFields(fields)
	}
	return cloneFields(def)
}

func getActionsProp(props rtui.Props, def []Action) []Action {
	if actions, ok := props[propActions].([]Action); ok {
		return cloneActions(actions)
	}
	return cloneActions(def)
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}
