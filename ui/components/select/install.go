package selectcomp

import (
	"reflect"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
)

type InstallerHost interface {
	AddMiddleware(action.ActionMiddleware)
}

var selectInstallState sync.Map

func Install(host InstallerHost) {
	if host == nil {
		return
	}
	key := installerHostKey(host)
	if key == 0 {
		host.AddMiddleware(NewMiddleware())
		return
	}
	if _, loaded := selectInstallState.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	host.AddMiddleware(NewMiddleware())
}

func installerHostKey(host InstallerHost) uintptr {
	value := reflect.ValueOf(host)
	if !value.IsValid() {
		return 0
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.UnsafePointer {
		return value.Pointer()
	}
	return 0
}
