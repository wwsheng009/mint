package formdialog

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for FormDialog.
type Instance struct {
	key             string
	title           string
	description     string
	open            bool
	width           int
	height          int
	formID          string
	layout          form.FormLayout
	values          map[string]interface{}
	validateAll     bool
	children        []rtui.VNode
	submitText      string
	cancelText      string
	submitVariant   button.Variant
	submitDisabled  bool
	disabledReason  string
	submitIntent    intent.Intent
	cancelIntent    intent.Intent
	closeIntent     intent.Intent
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool
	rootStyle       style.Style
	dirty           bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a FormDialog instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:             getStringProp(props, propKey, ""),
		title:           getStringProp(props, propTitle, ""),
		description:     getStringProp(props, propDescription, ""),
		open:            getBoolProp(props, propOpen, false),
		width:           getIntProp(props, propWidth, 72),
		height:          getIntProp(props, propHeight, 18),
		formID:          getStringProp(props, propFormID, ""),
		layout:          getLayoutProp(props, propLayout, form.LayoutVertical),
		values:          getValuesProp(props, propValues, nil),
		validateAll:     getBoolProp(props, propValidateAll, true),
		children:        getChildrenProp(props, propChildren, nil),
		submitText:      getStringProp(props, propSubmitText, "Submit"),
		cancelText:      getStringProp(props, propCancelText, "Cancel"),
		submitVariant:   getButtonVariantProp(props, propSubmitVariant, button.VariantPrimary),
		submitDisabled:  getBoolProp(props, propSubmitDisabled, false),
		disabledReason:  getStringProp(props, propDisabledReason, ""),
		submitIntent:    getIntentProp(props, propSubmitIntent, nil),
		cancelIntent:    getIntentProp(props, propCancelIntent, nil),
		closeIntent:     getIntentProp(props, propCloseIntent, nil),
		closeable:       getBoolProp(props, propCloseable, true),
		closeOnEsc:      getBoolProp(props, propCloseOnEsc, true),
		closeOnBackdrop: getBoolProp(props, propCloseOnBackdrop, true),
		rootStyle:       getStyleProp(props, propStyle, style.Style{}),
		dirty:           true,
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
	inst.description = getStringProp(props, propDescription, inst.description)
	inst.open = getBoolProp(props, propOpen, inst.open)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.height = getIntProp(props, propHeight, inst.height)
	inst.formID = getStringProp(props, propFormID, inst.formID)
	inst.layout = getLayoutProp(props, propLayout, inst.layout)
	inst.values = getValuesProp(props, propValues, inst.values)
	inst.validateAll = getBoolProp(props, propValidateAll, inst.validateAll)
	inst.children = getChildrenProp(props, propChildren, inst.children)
	inst.submitText = getStringProp(props, propSubmitText, inst.submitText)
	inst.cancelText = getStringProp(props, propCancelText, inst.cancelText)
	inst.submitVariant = getButtonVariantProp(props, propSubmitVariant, inst.submitVariant)
	inst.submitDisabled = getBoolProp(props, propSubmitDisabled, inst.submitDisabled)
	inst.disabledReason = getStringProp(props, propDisabledReason, inst.disabledReason)
	inst.submitIntent = getIntentProp(props, propSubmitIntent, inst.submitIntent)
	inst.cancelIntent = getIntentProp(props, propCancelIntent, inst.cancelIntent)
	inst.closeIntent = getIntentProp(props, propCloseIntent, inst.closeIntent)
	inst.closeable = getBoolProp(props, propCloseable, inst.closeable)
	inst.closeOnEsc = getBoolProp(props, propCloseOnEsc, inst.closeOnEsc)
	inst.closeOnBackdrop = getBoolProp(props, propCloseOnBackdrop, inst.closeOnBackdrop)
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
		propCancelIntent:    inst.cancelIntent,
		propCancelText:      inst.cancelText,
		propChildren:        cloneChildren(inst.children),
		propCloseIntent:     inst.closeIntent,
		propCloseOnBackdrop: inst.closeOnBackdrop,
		propCloseOnEsc:      inst.closeOnEsc,
		propCloseable:       inst.closeable,
		propDescription:     inst.description,
		propDisabledReason:  inst.disabledReason,
		propFormID:          inst.formID,
		propHeight:          inst.height,
		propKey:             inst.key,
		propLayout:          inst.layout,
		propOpen:            inst.open,
		propStyle:           inst.rootStyle,
		propSubmitDisabled:  inst.submitDisabled,
		propSubmitIntent:    inst.submitIntent,
		propSubmitText:      inst.submitText,
		propSubmitVariant:   inst.submitVariant,
		propTitle:           inst.title,
		propValidateAll:     inst.validateAll,
		propValues:          cloneValues(inst.values),
		propWidth:           inst.width,
	}
}

// RuntimeChildren synthesizes a framed form surface for Fiber.
func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.open {
		return nil
	}
	return []rtui.VNode{
		panel.NewBuilder().
			Key(inst.childKey("root")).
			Title(inst.titleOrDefault()).
			Content(inst.buildContent()).
			Footer(inst.buildFooter()).
			Rounded().
			Padding(1).
			Width(inst.widthOrDefault()).
			Height(inst.heightOrDefault()).
			Style(inst.rootStyle).
			Build(),
	}
}

func (inst *Instance) buildContent() rtui.VNode {
	children := make([]rtui.VNode, 0, 3)
	if strings.TrimSpace(inst.description) != "" {
		children = append(children, text.NewBuilder(inst.description).
			Key(inst.childKey("description")).
			Style(style.NewStyle().Foreground(theme.Text())).
			Build())
	}
	children = append(children, inst.buildForm())
	if strings.TrimSpace(inst.disabledReason) != "" {
		children = append(children, text.NewBuilder(inst.disabledReason).
			Key(inst.childKey("disabled-reason")).
			Style(style.NewStyle().Foreground(theme.Muted())).
			Build())
	}
	root := rtui.VStackBuilder(children...).Gap(1).AlignCross(rtui.AlignStart).
		Width(inst.widthOrDefault()).
		Height(inst.contentHeight())
	root.SetKey(inst.childKey("content"))
	return root.Build()
}

func (inst *Instance) buildForm() rtui.VNode {
	formID := inst.formIDOrDefault()
	builder := form.NewForm(formID).
		Label("").
		Layout(inst.layout).
		SetValues(cloneValues(inst.values)).
		ValidateAll(inst.validateAll).
		AddChildren(inst.children...)
	if inst.submitIntent != nil {
		builder.OnSubmit(inst.submitIntent)
	}
	builder.SetKey(formID)
	return builder
}

func (inst *Instance) buildFooter() rtui.VNode {
	cancel := button.NewBuilder(inst.cancelTextOrDefault()).
		Key(inst.childKey("cancel")).
		SetID(inst.childKey("cancel")).
		Secondary()
	if inst.cancelIntent != nil {
		cancel.OnPress(inst.cancelIntent)
	}

	submit := button.NewBuilder(inst.submitTextOrDefault()).
		Key(inst.childKey("submit")).
		SetID(inst.childKey("submit")).
		Variant(inst.submitVariant).
		Disabled(inst.submitDisabled)
	if inst.submitIntent != nil {
		submit.OnPress(inst.submitIntent)
	}

	footer := rtui.HStackBuilder(cancel.Build(), submit.Build()).Gap(2).Align(rtui.AlignEnd).AlignCross(rtui.AlignCenter)
	footer.SetKey(inst.childKey("footer"))
	return footer.Build()
}

func (inst *Instance) titleOrDefault() string {
	if strings.TrimSpace(inst.title) == "" {
		return "Form"
	}
	return inst.title
}

func (inst *Instance) formIDOrDefault() string {
	if strings.TrimSpace(inst.formID) != "" {
		return inst.formID
	}
	if strings.TrimSpace(inst.key) != "" {
		return inst.key + "-form"
	}
	return "formdialog-form"
}

func (inst *Instance) submitTextOrDefault() string {
	if strings.TrimSpace(inst.submitText) == "" {
		return "Submit"
	}
	return inst.submitText
}

func (inst *Instance) cancelTextOrDefault() string {
	if strings.TrimSpace(inst.cancelText) == "" {
		return "Cancel"
	}
	return inst.cancelText
}

func (inst *Instance) effectiveCloseIntent() intent.Intent {
	if inst.closeIntent != nil {
		return inst.closeIntent
	}
	return inst.cancelIntent
}

func (inst *Instance) widthOrDefault() int {
	if inst.width > 0 {
		return inst.width
	}
	return 72
}

func (inst *Instance) heightOrDefault() int {
	if inst.height > 0 {
		return inst.height
	}
	return 18
}

func (inst *Instance) contentHeight() int {
	height := inst.heightOrDefault() - 6
	if height < 6 {
		return 6
	}
	return height
}

func (inst *Instance) childKey(suffix string) string {
	if inst.key == "" {
		return "formdialog-" + suffix
	}
	return inst.key + "-" + suffix
}

func (inst *Instance) normalize() {
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.height < 0 {
		inst.height = 0
	}
	inst.layout = normalizeLayout(inst.layout)
	if inst.values == nil {
		inst.values = map[string]interface{}{}
	}
}

type instanceSnapshot struct {
	key             string
	title           string
	description     string
	open            bool
	width           int
	height          int
	formID          string
	layout          form.FormLayout
	values          map[string]interface{}
	validateAll     bool
	children        []rtui.VNode
	submitText      string
	cancelText      string
	submitVariant   button.Variant
	submitDisabled  bool
	disabledReason  string
	submitIntent    intent.Intent
	cancelIntent    intent.Intent
	closeIntent     intent.Intent
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool
	rootStyle       style.Style
}

func (inst *Instance) snapshot() instanceSnapshot {
	return instanceSnapshot{
		key:             inst.key,
		title:           inst.title,
		description:     inst.description,
		open:            inst.open,
		width:           inst.width,
		height:          inst.height,
		formID:          inst.formID,
		layout:          inst.layout,
		values:          cloneValues(inst.values),
		validateAll:     inst.validateAll,
		children:        cloneChildren(inst.children),
		submitText:      inst.submitText,
		cancelText:      inst.cancelText,
		submitVariant:   inst.submitVariant,
		submitDisabled:  inst.submitDisabled,
		disabledReason:  inst.disabledReason,
		submitIntent:    inst.submitIntent,
		cancelIntent:    inst.cancelIntent,
		closeIntent:     inst.closeIntent,
		closeable:       inst.closeable,
		closeOnEsc:      inst.closeOnEsc,
		closeOnBackdrop: inst.closeOnBackdrop,
		rootStyle:       inst.rootStyle,
	}
}
