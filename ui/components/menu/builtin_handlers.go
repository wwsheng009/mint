package menu

import "github.com/wwsheng009/mint/runtime/intent"

func init() {
	registerBuiltinHandlers(intent.DefaultRegistry())
}

func registerBuiltinHandlers(registry *intent.Registry) {
	if registry == nil {
		return
	}
	intent.RegisterTypedWithOpts(registry, handleNavigateMenuIntent, intent.WithOverridable(true))
}

func handleNavigateMenuIntent(ctx *intent.ActionContext, i NavigateMenuIntent) intent.IntentResult {
	return intent.HandledResult()
}
