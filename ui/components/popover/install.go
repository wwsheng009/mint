package popover

import (
	"reflect"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
)

type InstallerHost interface {
	AddMiddleware(action.ActionMiddleware)
}

var popoverInstallState sync.Map

func Install(host InstallerHost) {
	if host == nil {
		return
	}
	key := popoverInstallerHostKey(host)
	if key == 0 {
		host.AddMiddleware(NewMiddleware())
		return
	}
	if _, loaded := popoverInstallState.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	host.AddMiddleware(NewMiddleware())
}

func popoverInstallerHostKey(host InstallerHost) uintptr {
	value := reflect.ValueOf(host)
	if !value.IsValid() {
		return 0
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.UnsafePointer {
		return value.Pointer()
	}
	return 0
}
