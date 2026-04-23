package e2e

import (
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
)

func injectKeyMsg(app *App, key rune) error {
	if err := app.FrameworkApp().InjectMsg(runtimemsg.NewKeyMsg(key, platform.KeyUnknown, runtimemsg.Modifiers{})); err != nil {
		return err
	}
	return app.AwaitIdle()
}

func injectTextMsg(app *App, text string) error {
	for _, key := range text {
		if err := injectKeyMsg(app, key); err != nil {
			return err
		}
	}
	return nil
}
