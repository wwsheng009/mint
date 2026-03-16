package modal

import (
	"github.com/wwsheng009/mint/framework/theme"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

type staticAction string

const (
	staticActionAcknowledge staticAction = "acknowledge"
	staticActionConfirm     staticAction = "confirm"
	staticActionCancel      staticAction = "cancel"
)

type staticActionIntent struct {
	Action staticAction
	Next   runtimeintent.Intent
}

func (staticActionIntent) IntentType() string {
	return "Modal:StaticAction"
}

func (staticActionIntent) Priority() runtimeintent.ActionPriority {
	return runtimeintent.PriorityUserBlocking
}

type StaticOption func(*staticDialogConfig)

type staticDialogConfig struct {
	title          string
	message        string
	width          int
	height         int
	showCancel     bool
	confirmText    string
	cancelText     string
	confirmVariant button.Variant
	cancelVariant  button.Variant
	confirmIntent  runtimeintent.Intent
	cancelIntent   runtimeintent.Intent
	modalStyle     style.Style
}

func init() {
	runtimeintent.RegisterTypedGlobally(handleStaticAction)
}

func WithConfirmIntent(i runtimeintent.Intent) StaticOption {
	return func(cfg *staticDialogConfig) {
		cfg.confirmIntent = i
	}
}

func WithCancelIntent(i runtimeintent.Intent) StaticOption {
	return func(cfg *staticDialogConfig) {
		cfg.cancelIntent = i
	}
}

func WithConfirmText(text string) StaticOption {
	return func(cfg *staticDialogConfig) {
		if text != "" {
			cfg.confirmText = text
		}
	}
}

func WithCancelText(text string) StaticOption {
	return func(cfg *staticDialogConfig) {
		if text != "" {
			cfg.cancelText = text
		}
	}
}

func WithHelperStyle(s style.Style) StaticOption {
	return func(cfg *staticDialogConfig) {
		cfg.modalStyle = s
	}
}

func Alert(title, message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          title,
		message:        message,
		width:          44,
		height:         10,
		showCancel:     false,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantPrimary,
		cancelVariant:  button.VariantSecondary,
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func Confirm(title, message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          title,
		message:        message,
		width:          44,
		height:         10,
		showCancel:     true,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantPrimary,
		cancelVariant:  button.VariantSecondary,
		modalStyle:     style.NewStyle().Foreground(theme.Primary()),
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func Info(message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          "Info",
		message:        message,
		width:          44,
		height:         9,
		showCancel:     false,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantPrimary,
		cancelVariant:  button.VariantSecondary,
		modalStyle:     style.NewStyle().Foreground(theme.Primary()),
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func Success(message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          "Success",
		message:        message,
		width:          44,
		height:         9,
		showCancel:     false,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantSuccess,
		cancelVariant:  button.VariantSecondary,
		modalStyle:     style.NewStyle().Foreground(theme.Success()),
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func Error(message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          "Error",
		message:        message,
		width:          44,
		height:         9,
		showCancel:     false,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantDanger,
		cancelVariant:  button.VariantSecondary,
		modalStyle:     style.NewStyle().Foreground(theme.Error()),
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func Warning(message string, opts ...StaticOption) *Builder {
	cfg := staticDialogConfig{
		title:          "Warning",
		message:        message,
		width:          44,
		height:         9,
		showCancel:     false,
		confirmText:    "OK",
		cancelText:     "Cancel",
		confirmVariant: button.VariantSecondary,
		cancelVariant:  button.VariantSecondary,
		modalStyle:     style.NewStyle().Foreground(theme.Warning()),
	}
	applyStaticOptions(&cfg, opts...)
	return newStaticDialogBuilder(cfg)
}

func applyStaticOptions(cfg *staticDialogConfig, opts ...StaticOption) {
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
}

func newStaticDialogBuilder(cfg staticDialogConfig) *Builder {
	builder := NewBuilder().
		Title(cfg.title).
		Content(newtext.New(cfg.message)).
		Footer(staticFooter(cfg)).
		Width(cfg.width).
		Height(cfg.height).
		Padding(1).
		Open(true).
		Rounded()

	if !cfg.modalStyle.IsEmpty() {
		builder.Style(cfg.modalStyle)
	}

	return builder
}

func staticFooter(cfg staticDialogConfig) rtui.VNode {
	children := make([]rtui.VNode, 0, 2)
	if cfg.showCancel {
		children = append(children, newStaticFooterButton(cfg.cancelText, cfg.cancelVariant, staticActionCancel, cfg.cancelIntent))
	}
	children = append(children, newStaticFooterButton(cfg.confirmText, cfg.confirmVariant, confirmStaticAction(cfg), cfg.confirmIntent))
	return rtui.HStackBuilder(children...).Align(rtui.AlignEnd).Gap(1).Build()
}

func confirmStaticAction(cfg staticDialogConfig) staticAction {
	if cfg.showCancel {
		return staticActionConfirm
	}
	return staticActionAcknowledge
}

func newStaticFooterButton(label string, variant button.Variant, action staticAction, next runtimeintent.Intent) rtui.VNode {
	builder := button.NewBuilder(label).Variant(variant).OnPress(staticActionIntent{
		Action: action,
		Next:   next,
	})
	return builder.Build()
}

func handleStaticAction(_ *runtimeintent.ActionContext, i staticActionIntent) runtimeintent.IntentResult {
	target := globalRegistry.getTopmostOpenModal()
	if target != nil {
		target.requestClose()
	}
	emitOptionalGlobalIntent(i.Next)
	return runtimeintent.HandledResult()
}

func emitOptionalGlobalIntent(i runtimeintent.Intent) {
	if i == nil {
		return
	}
	rt := rtui.GetGlobalIntentRuntime()
	if rt == nil || rt.Registry == nil {
		return
	}
	intentType := i.IntentType()
	if !rt.Registry.HasHandler(intentType) && !rt.Registry.HasFallback() {
		return
	}
	rtui.EmitIntentGlobal(i)
}
