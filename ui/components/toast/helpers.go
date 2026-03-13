package toast

import (
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{}
}

func getToastDurationProp(props rtui.Props, def time.Duration) time.Duration {
	if v, ok := props["duration"]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return def
}

func getToastTypeProp(props rtui.Props, def ToastType) ToastType {
	if v, ok := props["toastType"]; ok {
		if tt, ok := v.(ToastType); ok {
			return tt
		}
	}
	return def
}
