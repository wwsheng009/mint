package selectcomp

import "github.com/wwsheng009/mint/runtime/action"

type InstallerHost interface {
	AddMiddleware(action.ActionMiddleware)
}

type installerStateHost interface {
	SetUserData(string, interface{})
	GetUserData(string) interface{}
}

const installStateUserDataKey = "mint.ui.components.select.installState"

func Install(host InstallerHost) {
	if host == nil {
		return
	}
	if stateHost, ok := host.(installerStateHost); ok {
		if installed, _ := stateHost.GetUserData(installStateUserDataKey).(bool); installed {
			return
		}
		stateHost.SetUserData(installStateUserDataKey, true)
		host.AddMiddleware(NewMiddleware())
		return
	}
	host.AddMiddleware(NewMiddleware())
}
