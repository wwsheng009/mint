package alert

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func getAlertTypeProp(props rtui.Props, def AlertType) AlertType {
	if v, ok := props[propAlertType]; ok {
		if t, ok := v.(AlertType); ok {
			return t
		}
	}
	return def
}
