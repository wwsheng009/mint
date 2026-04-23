package ui

import (
	"io"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/internal/log"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

func installGraphicsBootstrap(app *framework.App, writer io.Writer) runtimeplatform.GraphicsCapabilities {
	if app == nil {
		return runtimeplatform.GraphicsCapabilities{}
	}

	caps, presenter := probeGraphicsBootstrap(writer)
	app.SetGraphicsPresenter(presenter)

	log.UILogger.IfEnabled().Debug("ui graphics bootstrap: %s presenter=%T", caps.Summary(), presenter)
	return caps
}

func probeGraphicsBootstrap(writer io.Writer) (runtimeplatform.GraphicsCapabilities, runtimeplatform.GraphicsPresenter) {
	caps := runtimeplatform.ProbeGraphicsCapabilities()

	switch caps.Mode {
	case runtimeplatform.GraphicsModeKitty:
		return caps, runtimeplatform.NewKittyGraphicsPresenter(writer, caps)
	case runtimeplatform.GraphicsModeInlineImage:
		return caps, runtimeplatform.NewInlineImageGraphicsPresenter(writer, caps)
	case runtimeplatform.GraphicsModeSixel:
		return caps, runtimeplatform.NewSixelGraphicsPresenter(writer, caps)
	default:
		return caps, nil
	}
}

func runtimeGraphicsWriter() io.Writer {
	return os.Stdout
}

func testGraphicsWriter() io.Writer {
	return nil
}
