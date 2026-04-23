package notification

import (
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func getNotificationTypeProp(props rtui.Props, def NotificationType) NotificationType {
	if v, ok := props[propNotificationType]; ok {
		if t, ok := v.(NotificationType); ok {
			return t
		}
	}
	return def
}

func getDurationProp(props rtui.Props, def time.Duration) time.Duration {
	if v, ok := props[propDuration]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return def
}

func getPlacementProp(props rtui.Props, def Placement) Placement {
	if v, ok := props[propPlacement]; ok {
		if p, ok := v.(Placement); ok {
			return p
		}
	}
	return def
}
