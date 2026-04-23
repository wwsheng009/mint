package popconfirm

import (
	"reflect"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
)

type InstallerHost interface {
	AddMiddleware(action.ActionMiddleware)
}

var popconfirmInstallState sync.Map

func Install(host InstallerHost) {
	if host == nil {
		return
	}
	key := popconfirmInstallerHostKey(host)
	if key == 0 {
		host.AddMiddleware(NewMiddleware())
		return
	}
	if _, loaded := popconfirmInstallState.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	host.AddMiddleware(NewMiddleware())
}

func popconfirmInstallerHostKey(host InstallerHost) uintptr {
	value := reflect.ValueOf(host)
	if !value.IsValid() {
		return 0
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.UnsafePointer {
		return value.Pointer()
	}
	return 0
}
