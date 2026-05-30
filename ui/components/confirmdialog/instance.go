package confirmdialog

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/descriptions"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/modal"
	"github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for ConfirmDialog.
type Instance struct {
	key                      string
	title                    string
	message                  string
	warning                  string
	open                     bool
	width                    int
	height                   int
	targetItems              []TargetItem
	reasonLabel              string
	reasonValue              string
	reasonField              string
	reasonPlaceholder        string
	reasonRequired           bool
	confirmPhrase            string
	confirmPhraseValue       string
	confirmPhraseField       string
	confirmPhraseLabel       string
	confirmPhrasePlaceholder string
	confirmText              string
	cancelText               string
	confirmVariant           button.Variant
	disableConfirm           bool
	disabledReason           string
	confirmIntent            intent.Intent
	cancelIntent             intent.Intent
	rootStyle                style.Style
	dirty                    bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a ConfirmDialog instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                      getStringProp(props, propKey, ""),
		title:                    getStringProp(props, propTitle, ""),
		message:                  getStringProp(props, propMessage, ""),
		warning:                  getStringProp(props, propWarning, ""),
		open:                     getBoolProp(props, propOpen, false),
		width:                    getIntProp(props, propWidth, 68),
		height:                   getIntProp(props, propHeight, 18),
		targetItems:              normalizeTargetItems(getTargetItemsProp(props, propTargetItems, nil)),
		reasonLabel:              getStringProp(props, propReasonLabel, "Reason"),
		reasonValue:              getStringProp(props, propReasonValue, ""),
		reasonField:              getStringProp(props, propReasonField, ""),
		reasonPlaceholder:        getStringProp(props, propReasonPlaceholder, "Audit reason"),
		reasonRequired:           getBoolProp(props, propReasonRequired, false),
		confirmPhrase:            getStringProp(props, propConfirmPhrase, ""),
		confirmPhraseValue:       getStringProp(props, propConfirmPhraseValue, ""),
		confirmPhraseField:       getStringProp(props, propConfirmPhraseField, ""),
		confirmPhraseLabel:       getStringProp(props, propConfirmPhraseLabel, ""),
		confirmPhrasePlaceholder: getStringProp(props, propConfirmPhrasePlaceholder, ""),
		confirmText:              getStringProp(props, propConfirmText, "Confirm"),
		cancelText:               getStringProp(props, propCancelText, "Cancel"),
		confirmVariant:           getButtonVariantProp(props, propConfirmVariant, button.VariantDanger),
		disableConfirm:           getBoolProp(props, propDisableConfirm, false),
		disabledReason:           getStringProp(props, propDisabledReason, ""),
		confirmIntent:            getIntentProp(props, propConfirmIntent, nil),
		cancelIntent:             getIntentProp(props, propCancelIntent, nil),
		rootStyle:                getStyleProp(props, propStyle, style.Style{}),
		dirty:                    true,
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
	inst.message = getStringProp(props, propMessage, inst.message)
	inst.warning = getStringProp(props, propWarning, inst.warning)
	inst.open = getBoolProp(props, propOpen, inst.open)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.height = getIntProp(props, propHeight, inst.height)
	inst.targetItems = normalizeTargetItems(getTargetItemsProp(props, propTargetItems, inst.targetItems))
	inst.reasonLabel = getStringProp(props, propReasonLabel, inst.reasonLabel)
	inst.reasonValue = getStringProp(props, propReasonValue, inst.reasonValue)
	inst.reasonField = getStringProp(props, propReasonField, inst.reasonField)
	inst.reasonPlaceholder = getStringProp(props, propReasonPlaceholder, inst.reasonPlaceholder)
	inst.reasonRequired = getBoolProp(props, propReasonRequired, inst.reasonRequired)
	inst.confirmPhrase = getStringProp(props, propConfirmPhrase, inst.confirmPhrase)
	inst.confirmPhraseValue = getStringProp(props, propConfirmPhraseValue, inst.confirmPhraseValue)
	inst.confirmPhraseField = getStringProp(props, propConfirmPhraseField, inst.confirmPhraseField)
	inst.confirmPhraseLabel = getStringProp(props, propConfirmPhraseLabel, inst.confirmPhraseLabel)
	inst.confirmPhrasePlaceholder = getStringProp(props, propConfirmPhrasePlaceholder, inst.confirmPhrasePlaceholder)
	inst.confirmText = getStringProp(props, propConfirmText, inst.confirmText)
	inst.cancelText = getStringProp(props, propCancelText, inst.cancelText)
	inst.confirmVariant = getButtonVariantProp(props, propConfirmVariant, inst.confirmVariant)
	inst.disableConfirm = getBoolProp(props, propDisableConfirm, inst.disableConfirm)
	inst.disabledReason = getStringProp(props, propDisabledReason, inst.disabledReason)
	inst.confirmIntent = getIntentProp(props, propConfirmIntent, inst.confirmIntent)
	inst.cancelIntent = getIntentProp(props, propCancelIntent, inst.cancelIntent)
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
		propCancelIntent:             inst.cancelIntent,
		propCancelText:               inst.cancelText,
		propConfirmIntent:            inst.confirmIntent,
		propConfirmPhrase:            inst.confirmPhrase,
		propConfirmPhraseField:       inst.confirmPhraseField,
		propConfirmPhraseLabel:       inst.confirmPhraseLabel,
		propConfirmPhrasePlaceholder: inst.confirmPhrasePlaceholder,
		propConfirmPhraseValue:       inst.confirmPhraseValue,
		propConfirmText:              inst.confirmText,
		propConfirmVariant:           inst.confirmVariant,
		propDisableConfirm:           inst.disableConfirm,
		propDisabledReason:           inst.disabledReason,
		propHeight:                   inst.height,
		propKey:                      inst.key,
		propMessage:                  inst.message,
		propOpen:                     inst.open,
		propReasonField:              inst.reasonField,
		propReasonLabel:              inst.reasonLabel,
		propReasonPlaceholder:        inst.reasonPlaceholder,
		propReasonRequired:           inst.reasonRequired,
		propReasonValue:              inst.reasonValue,
		propStyle:                    inst.rootStyle,
		propTargetItems:              cloneTargetItems(inst.targetItems),
		propTitle:                    inst.title,
		propWarning:                  inst.warning,
		propWidth:                    inst.width,
	}
}

// RuntimeChildren synthesizes the modal, body, footer, and controls used by Fiber.
func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.open {
		return nil
	}
	builder := modal.NewBuilder().
		Key(inst.rootKey()).
		Title(inst.titleOrDefault()).
		Content(inst.buildBody()).
		Footer(inst.buildFooter()).
		Open(true).
		Centered(true).
		Closeable(false).
		CloseOnEsc(false).
		CloseOnBackdrop(false).
		Padding(1).
		Width(inst.widthOrDefault()).
		Height(inst.heightOrDefault()).
		BorderStyle("double")
	if !inst.rootStyle.IsEmpty() {
		builder.Style(inst.rootStyle)
	}
	return []rtui.VNode{builder.Build()}
}

func (inst *Instance) buildContent() rtui.VNode {
	root := rtui.VStackBuilder(inst.buildBody(), inst.buildFooter()).Gap(1).AlignCross(rtui.AlignStart)
	root.SetKey(inst.childKey("content"))
	return root.Build()
}

func (inst *Instance) buildBody() rtui.VNode {
	children := make([]rtui.VNode, 0, 6)
	if strings.TrimSpace(inst.message) != "" {
		children = append(children, text.NewBuilder(inst.message).
			Key(inst.childKey("message")).
			Style(style.NewStyle().Foreground(theme.Text())).
			Build())
	}
	if strings.TrimSpace(inst.warning) != "" {
		children = append(children, text.NewBuilder(inst.warning).
			Key(inst.childKey("warning")).
			Style(style.NewStyle().Foreground(style.Yellow).Bold(true)).
			Build())
	}
	if len(inst.targetItems) > 0 {
		children = append(children, inst.buildTargetSummary())
	}
	if inst.reasonField != "" || inst.reasonRequired || inst.reasonValue != "" {
		children = append(children, inst.buildReasonInput())
	}
	if strings.TrimSpace(inst.confirmPhrase) != "" {
		children = append(children, inst.buildConfirmPhraseInput())
	}
	if disabledReason := inst.confirmDisabledReason(); disabledReason != "" {
		children = append(children, text.NewBuilder(disabledReason).
			Key(inst.childKey("disabled-reason")).
			Style(style.NewStyle().Foreground(style.BrightBlack)).
			Build())
	}
	if len(children) == 0 {
		children = append(children, text.NewBuilder("Confirm this operation.").Key(inst.childKey("default-message")).Build())
	}
	root := rtui.VStackBuilder(children...).Gap(1).AlignCross(rtui.AlignStart)
	root.SetKey(inst.childKey("body"))
	return root.Build()
}

func (inst *Instance) buildTargetSummary() rtui.VNode {
	columns := inst.targetSummaryColumns()
	builder := descriptions.NewBuilder().
		Key(inst.childKey("targets")).
		Column(columns).
		LabelWidth(14).
		ContentWidth(inst.targetContentWidth()).
		EmptyText("-").
		MaskText("masked")
	for _, item := range inst.targetItems {
		if item.Sensitive {
			builder.Item(descriptions.SensitiveField(item.Label, item.Value))
		} else {
			builder.Item(descriptions.Value(item.Label, item.Value))
		}
	}
	return builder.Build()
}

func (inst *Instance) buildReasonInput() rtui.VNode {
	label := inst.reasonLabel
	if strings.TrimSpace(label) == "" {
		label = "Reason"
	}
	if inst.reasonRequired {
		label += " *"
	}
	inputBuilder := input.NewBuilder().
		Key(inst.childKey("reason-input")).
		Value(inst.reasonValue).
		Placeholder(inst.reasonPlaceholder).
		Width(inst.reasonInputWidth())
	if strings.TrimSpace(inst.reasonField) != "" {
		inputBuilder.ForField(intent.BindField(inst.reasonField))
	}
	return rtui.VStackBuilder(
		text.NewBuilder(label).
			Key(inst.childKey("reason-label")).
			Style(style.NewStyle().Foreground(theme.Muted()).Bold(true)).
			Build(),
		inputBuilder.Build(),
	).Gap(0).AlignCross(rtui.AlignStart).Build()
}

func (inst *Instance) buildConfirmPhraseInput() rtui.VNode {
	expected := strings.TrimSpace(inst.confirmPhrase)
	label := inst.confirmPhraseLabel
	if strings.TrimSpace(label) == "" {
		label = "Confirmation"
	}
	if expected != "" {
		label += " *"
	}
	placeholder := inst.confirmPhrasePlaceholder
	if strings.TrimSpace(placeholder) == "" {
		placeholder = confirmPhrasePlaceholder(expected)
	}
	inputBuilder := input.NewBuilder().
		Key(inst.childKey("confirm-phrase-input")).
		Value(inst.confirmPhraseValue).
		Placeholder(placeholder).
		Width(inst.reasonInputWidth())
	if strings.TrimSpace(inst.confirmPhraseField) != "" {
		inputBuilder.ForField(intent.BindField(inst.confirmPhraseField))
	}
	return rtui.VStackBuilder(
		text.NewBuilder(label).
			Key(inst.childKey("confirm-phrase-label")).
			Style(style.NewStyle().Foreground(theme.Muted()).Bold(true)).
			Build(),
		inputBuilder.Build(),
	).Gap(0).AlignCross(rtui.AlignStart).Build()
}

func (inst *Instance) buildFooter() rtui.VNode {
	cancel := button.NewBuilder(inst.cancelTextOrDefault()).
		Key(inst.childKey("cancel")).
		Secondary()
	if inst.cancelIntent != nil {
		cancel.OnPress(inst.cancelIntent)
	}

	confirmDisabled := inst.disableConfirm ||
		(inst.reasonRequired && strings.TrimSpace(inst.reasonValue) == "") ||
		!inst.confirmPhraseSatisfied()
	confirm := button.NewBuilder(inst.confirmTextOrDefault()).
		Key(inst.childKey("confirm")).
		Variant(inst.confirmVariant).
		Disabled(confirmDisabled)
	if inst.confirmIntent != nil {
		confirm.OnPress(inst.confirmIntent)
	}

	footer := rtui.HStackBuilder(cancel.Build(), confirm.Build()).Gap(2).Align(rtui.AlignEnd).AlignCross(rtui.AlignCenter)
	footer.SetKey(inst.childKey("footer"))
	return footer.Build()
}

func (inst *Instance) confirmDisabledReason() string {
	reasons := make([]string, 0, 3)
	if reason := strings.TrimSpace(inst.disabledReason); reason != "" {
		reasons = append(reasons, reason)
	}
	if inst.reasonRequired && strings.TrimSpace(inst.reasonValue) == "" {
		label := strings.TrimSpace(strings.TrimSuffix(inst.reasonLabel, "*"))
		if label == "" {
			label = "reason"
		}
		reasons = append(reasons, "Enter "+strings.ToLower(label)+" before confirming.")
	}
	expected := strings.TrimSpace(inst.confirmPhrase)
	if expected != "" && !inst.confirmPhraseSatisfied() {
		reasons = append(reasons, confirmPhrasePlaceholder(expected))
	}
	if inst.disableConfirm && len(reasons) == 0 {
		reasons = append(reasons, "Confirmation is disabled.")
	}
	return strings.Join(reasons, " ")
}

func (inst *Instance) confirmPhraseSatisfied() bool {
	expected := strings.TrimSpace(inst.confirmPhrase)
	if expected == "" {
		return true
	}
	return strings.TrimSpace(inst.confirmPhraseValue) == expected
}

func (inst *Instance) titleOrDefault() string {
	if strings.TrimSpace(inst.title) == "" {
		return "Confirm Operation"
	}
	return inst.title
}

func (inst *Instance) confirmTextOrDefault() string {
	if strings.TrimSpace(inst.confirmText) == "" {
		return "Confirm"
	}
	return inst.confirmText
}

func (inst *Instance) cancelTextOrDefault() string {
	if strings.TrimSpace(inst.cancelText) == "" {
		return "Cancel"
	}
	return inst.cancelText
}

func (inst *Instance) widthOrDefault() int {
	if inst.width > 0 {
		return inst.width
	}
	return 68
}

func (inst *Instance) heightOrDefault() int {
	if inst.height > 0 {
		return inst.height
	}
	if strings.TrimSpace(inst.confirmPhrase) != "" {
		return minConfirmPhraseHeight
	}
	return 18
}

func (inst *Instance) targetContentWidth() int {
	if inst.targetSummaryColumns() > 1 {
		innerWidth := inst.widthOrDefault() - 6
		if innerWidth < 1 {
			innerWidth = 1
		}
		columnGap := 3
		cellWidth := (innerWidth - columnGap) / 2
		width := cellWidth - 15
		if width < 16 {
			return 16
		}
		return width
	}
	width := inst.widthOrDefault() - 26
	if width < 20 {
		return 20
	}
	return width
}

func (inst *Instance) targetSummaryColumns() int {
	if len(inst.targetItems) >= 6 && inst.widthOrDefault() >= 80 {
		return 2
	}
	return 1
}

func (inst *Instance) reasonInputWidth() int {
	width := inst.widthOrDefault() - 10
	if width < 20 {
		return 20
	}
	return width
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "confirmdialog-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) childKey(suffix string) string {
	if inst.key == "" {
		return "confirmdialog-" + suffix
	}
	return inst.key + "-" + suffix
}

func (inst *Instance) normalize() {
	inst.targetItems = normalizeTargetItems(inst.targetItems)
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.height < 0 {
		inst.height = 0
	}
	if strings.TrimSpace(inst.confirmPhrase) != "" && inst.height > 0 && inst.height < minConfirmPhraseHeight {
		inst.height = minConfirmPhraseHeight
	}
}

type instanceSnapshot struct {
	key                      string
	title                    string
	message                  string
	warning                  string
	open                     bool
	width                    int
	height                   int
	targetItems              []TargetItem
	reasonLabel              string
	reasonValue              string
	reasonField              string
	reasonPlaceholder        string
	reasonRequired           bool
	confirmPhrase            string
	confirmPhraseValue       string
	confirmPhraseField       string
	confirmPhraseLabel       string
	confirmPhrasePlaceholder string
	confirmText              string
	cancelText               string
	confirmVariant           button.Variant
	disableConfirm           bool
	disabledReason           string
	confirmIntent            intent.Intent
	cancelIntent             intent.Intent
	rootStyle                style.Style
}

func (inst *Instance) snapshot() instanceSnapshot {
	return instanceSnapshot{
		key:                      inst.key,
		title:                    inst.title,
		message:                  inst.message,
		warning:                  inst.warning,
		open:                     inst.open,
		width:                    inst.width,
		height:                   inst.height,
		targetItems:              cloneTargetItems(inst.targetItems),
		reasonLabel:              inst.reasonLabel,
		reasonValue:              inst.reasonValue,
		reasonField:              inst.reasonField,
		reasonPlaceholder:        inst.reasonPlaceholder,
		reasonRequired:           inst.reasonRequired,
		confirmPhrase:            inst.confirmPhrase,
		confirmPhraseValue:       inst.confirmPhraseValue,
		confirmPhraseField:       inst.confirmPhraseField,
		confirmPhraseLabel:       inst.confirmPhraseLabel,
		confirmPhrasePlaceholder: inst.confirmPhrasePlaceholder,
		confirmText:              inst.confirmText,
		cancelText:               inst.cancelText,
		confirmVariant:           inst.confirmVariant,
		disableConfirm:           inst.disableConfirm,
		disabledReason:           inst.disabledReason,
		confirmIntent:            inst.confirmIntent,
		cancelIntent:             inst.cancelIntent,
		rootStyle:                inst.rootStyle,
	}
}

func getTargetItemsProp(props rtui.Props, key string, def []TargetItem) []TargetItem {
	if items, ok := props[key].([]TargetItem); ok {
		return cloneTargetItems(items)
	}
	return cloneTargetItems(def)
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

func getIntentProp(props rtui.Props, key string, def intent.Intent) intent.Intent {
	if value, ok := props[key]; ok {
		if i, ok := value.(intent.Intent); ok {
			return i
		}
	}
	return def
}

func getButtonVariantProp(props rtui.Props, key string, def button.Variant) button.Variant {
	if value, ok := props[key]; ok {
		if variant, ok := value.(button.Variant); ok {
			return variant
		}
	}
	return def
}
